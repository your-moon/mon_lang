/*
 * mon_lang - TACKY optimizations
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// TACKY-level optimizations (Sandler ch19): constant folding, local copy
// propagation, dead-store elimination and unreachable-code removal, run to
// a fixpoint per function. They are on by default - the differential test
// suite compares optimized native output against the unoptimized cc path,
// so a miscompiling optimization cannot land silently.
package tackygen

import (
	"github.com/your-moon/mon_lang/mconstant"
)

// Optimize rewrites each function's instruction list. Four rounds reach the
// fixpoint for realistic programs (fold enables propagation enables more
// folding); instruction equality can't be tested directly because FnCall
// holds a slice, so the loop is bounded rather than convergence-checked.
func Optimize(p TackyProgram) TackyProgram {
	for i := range p.FnDefs {
		irs := p.FnDefs[i].Instructions
		for range [4]int{} {
			irs = foldConstants(irs)
			irs = propagateCopies(irs)
			irs = dropUnreachable(irs)
			irs = dropDeadStores(irs)
		}
		p.FnDefs[i].Instructions = irs
	}
	return p
}

func constOf(v TackyVal) (int64, bool) {
	c, ok := v.(Constant)
	if !ok {
		return 0, false
	}
	// float constants carry bit patterns; integer folding rules never apply
	if _, isFloat := c.Value.(*mconstant.Float64); isFloat {
		return 0, false
	}
	return c.Value.GetValue(), true
}

func constIs64(v TackyVal) bool {
	c, ok := v.(Constant)
	if !ok {
		return false
	}
	_, is64 := c.Value.(*mconstant.Int64)
	return is64
}

// makeConst preserves the operand width: the emitter derives move sizes
// from operand types, and widening a folded constant to 64 bits would turn
// a 4-byte store into an 8-byte one, clobbering the neighbouring slot.
func makeConst(v int64, is64 bool) TackyVal {
	if is64 {
		return Constant{Value: &mconstant.Int64{Value: v}}
	}
	return Constant{Value: &mconstant.Int32{Value: int32(v)}}
}

func foldBinary(op TackyBinaryOp, l, r int64) (int64, bool) {
	switch op {
	case Add:
		return l + r, true
	case Sub:
		return l - r, true
	case Mul:
		return l * r, true
	case Div:
		if r == 0 {
			return 0, false // keep the runtime trap
		}
		return l / r, true
	case Modulo:
		if r == 0 {
			return 0, false
		}
		return l % r, true
	case UDiv:
		if r == 0 {
			return 0, false
		}
		return int64(uint64(l) / uint64(r)), true
	case UModulo:
		if r == 0 {
			return 0, false
		}
		return int64(uint64(l) % uint64(r)), true
	case Equal:
		return b2i(l == r), true
	case NotEqual:
		return b2i(l != r), true
	case LessThan:
		return b2i(l < r), true
	case LessThanEqual:
		return b2i(l <= r), true
	case GreaterThan:
		return b2i(l > r), true
	case GreaterThanEqual:
		return b2i(l >= r), true
	case ULessThan:
		return b2i(uint64(l) < uint64(r)), true
	case ULessThanEqual:
		return b2i(uint64(l) <= uint64(r)), true
	case UGreaterThan:
		return b2i(uint64(l) > uint64(r)), true
	case UGreaterThanEqual:
		return b2i(uint64(l) >= uint64(r)), true
	}
	return 0, false
}

func isRelational(op TackyBinaryOp) bool {
	switch op {
	case Equal, NotEqual, LessThan, LessThanEqual, GreaterThan, GreaterThanEqual,
		ULessThan, ULessThanEqual, UGreaterThan, UGreaterThanEqual:
		return true
	}
	return false
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func foldConstants(irs []Instruction) []Instruction {
	out := make([]Instruction, 0, len(irs))
	for _, in := range irs {
		switch instr := in.(type) {
		case Binary:
			l, lok := constOf(instr.Src1)
			r, rok := constOf(instr.Src2)
			if lok && rok {
				if v, ok := foldBinary(instr.Op, l, r); ok {
					is64 := constIs64(instr.Src1) || constIs64(instr.Src2)
					if isRelational(instr.Op) {
						is64 = false // comparison results are 32-bit
					}
					out = append(out, Copy{Src: makeConst(v, is64), Dst: instr.Dst})
					continue
				}
			}
		case Unary:
			if v, ok := constOf(instr.Src); ok {
				switch instr.Op {
				case Negate:
					out = append(out, Copy{Src: makeConst(-v, constIs64(instr.Src)), Dst: instr.Dst})
					continue
				case Complement:
					out = append(out, Copy{Src: makeConst(^v, constIs64(instr.Src)), Dst: instr.Dst})
					continue
				}
			}
		case JumpIfZero:
			if v, ok := constOf(instr.Val); ok {
				if v == 0 {
					out = append(out, Jump{Target: instr.Ident})
				}
				continue // nonzero: the branch never fires
			}
		case JumpIfNotZero:
			if v, ok := constOf(instr.Val); ok {
				if v != 0 {
					out = append(out, Jump{Target: instr.Ident})
				}
				continue
			}
		case SignExtend:
			if v, ok := constOf(instr.Src); ok {
				out = append(out, Copy{Src: makeConst(v, true), Dst: instr.Dst})
				continue
			}
		case ZeroExtend:
			if v, ok := constOf(instr.Src); ok {
				out = append(out, Copy{Src: makeConst(int64(uint32(v)), true), Dst: instr.Dst})
				continue
			}
		}
		out = append(out, in)
	}
	return out
}

// propagateCopies replaces uses of temps that hold a known constant or the
// value of another variable, within basic-block boundaries. Knowledge dies
// at labels/jumps (control may merge) and calls invalidate globals - since
// TACKY names don't distinguish globals from locals here, calls kill all
// non-constant knowledge, and stores kill nothing (they write through
// pointers, never named vars).
func propagateCopies(irs []Instruction) []Instruction {
	// address-taken vars can change through any Store: never track them
	aliased := map[string]bool{}
	for _, in := range irs {
		if ga, ok := in.(GetAddress); ok {
			if v, ok := ga.Src.(Var); ok {
				aliased[v.Name] = true
			}
		}
	}

	known := map[string]TackyVal{}
	kill := func() { known = map[string]TackyVal{} }
	// drop any mapping that reads a var being overwritten
	killVar := func(name string) {
		delete(known, name)
		for k, v := range known {
			if vv, ok := v.(Var); ok && vv.Name == name {
				delete(known, k)
			}
		}
	}
	subst := func(v TackyVal) TackyVal {
		if vv, ok := v.(Var); ok {
			if rep, ok := known[vv.Name]; ok {
				return rep
			}
		}
		return v
	}

	out := make([]Instruction, 0, len(irs))
	for _, in := range irs {
		switch instr := in.(type) {
		case Copy:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
				if !aliased[dst.Name] {
					switch s := instr.Src.(type) {
					case Constant:
						known[dst.Name] = s
					case Var:
						if s.Name != dst.Name && !aliased[s.Name] {
							known[dst.Name] = s
						}
					}
				}
			}
			out = append(out, instr)
		case Binary:
			instr.Src1 = subst(instr.Src1)
			instr.Src2 = subst(instr.Src2)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case Unary:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case SignExtend:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case ZeroExtend:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case IntToDouble:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case DoubleToInt:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case JumpIfZero:
			instr.Val = subst(instr.Val)
			out = append(out, instr)
		case JumpIfNotZero:
			instr.Val = subst(instr.Val)
			out = append(out, instr)
		case Return:
			if instr.Value != nil {
				instr.Value = subst(instr.Value)
			}
			out = append(out, instr)
		case FnCall:
			for i := range instr.Args {
				instr.Args[i] = subst(instr.Args[i])
			}
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			kill() // the callee may mutate globals
			out = append(out, instr)
		case Load:
			instr.Src = subst(instr.Src)
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			out = append(out, instr)
		case Store:
			instr.Src = subst(instr.Src)
			instr.Dst = subst(instr.Dst)
			out = append(out, instr)
		case GetAddress:
			if dst, ok := instr.Dst.(Var); ok {
				killVar(dst.Name)
			}
			// taking a var's address makes it aliased: stop reasoning
			// about it entirely within this function
			if src, ok := instr.Src.(Var); ok {
				killVar(src.Name)
			}
			out = append(out, instr)
		case Label, Jump:
			kill()
			out = append(out, in)
		default:
			kill()
			out = append(out, in)
		}
	}
	return out
}

// dropUnreachable removes instructions between an unconditional transfer
// and the next label.
func dropUnreachable(irs []Instruction) []Instruction {
	out := make([]Instruction, 0, len(irs))
	dead := false
	for _, in := range irs {
		switch in.(type) {
		case Label:
			dead = false
		case Jump, Return:
			if dead {
				continue
			}
			out = append(out, in)
			dead = true
			continue
		}
		if dead {
			continue
		}
		out = append(out, in)
	}
	return out
}

// dropDeadStores removes pure writes to temps that are never read. Address-
// taken vars are exempt (reads may happen through pointers).
func dropDeadStores(irs []Instruction) []Instruction {
	read := map[string]bool{}
	aliased := map[string]bool{}
	note := func(v TackyVal) {
		if vv, ok := v.(Var); ok {
			read[vv.Name] = true
		}
	}
	for _, in := range irs {
		switch instr := in.(type) {
		case Copy:
			note(instr.Src)
		case Binary:
			note(instr.Src1)
			note(instr.Src2)
		case Unary:
			note(instr.Src)
		case SignExtend:
			note(instr.Src)
		case ZeroExtend:
			note(instr.Src)
		case IntToDouble:
			note(instr.Src)
		case DoubleToInt:
			note(instr.Src)
		case JumpIfZero:
			note(instr.Val)
		case JumpIfNotZero:
			note(instr.Val)
		case Return:
			if instr.Value != nil {
				note(instr.Value)
			}
		case FnCall:
			for _, a := range instr.Args {
				note(a)
			}
		case Load:
			note(instr.Src)
		case Store:
			note(instr.Src)
			note(instr.Dst)
		case GetAddress:
			if v, ok := instr.Src.(Var); ok {
				aliased[v.Name] = true
			}
		}
	}
	deadDst := func(dst TackyVal) bool {
		v, ok := dst.(Var)
		return ok && !read[v.Name] && !aliased[v.Name]
	}
	out := make([]Instruction, 0, len(irs))
	for _, in := range irs {
		switch instr := in.(type) {
		case Copy:
			if deadDst(instr.Dst) {
				continue
			}
		case Unary:
			if deadDst(instr.Dst) {
				continue
			}
		case Binary:
			// division traps stay even when unused
			if instr.Op != Div && instr.Op != Modulo && instr.Op != UDiv &&
				instr.Op != UModulo && deadDst(instr.Dst) {
				continue
			}
		case SignExtend:
			if deadDst(instr.Dst) {
				continue
			}
		case ZeroExtend:
			if deadDst(instr.Dst) {
				continue
			}
		case IntToDouble:
			if deadDst(instr.Dst) {
				continue
			}
		case DoubleToInt:
			if deadDst(instr.Dst) {
				continue
			}
		case Load:
			if deadDst(instr.Dst) {
				continue
			}
		case GetAddress:
			if deadDst(instr.Dst) {
				continue
			}
		}
		out = append(out, in)
	}
	return out
}
