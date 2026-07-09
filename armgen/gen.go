/*
 * mon_lang - Tacky IR -> arm64 lowering + native executable driver
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package armgen

import (
	"fmt"

	"github.com/your-moon/mon_lang/macho"
	"github.com/your-moon/mon_lang/tackygen"
)

const scratchC = 10 // third scratch (x10), for modulo

func fnLabel(s string) string  { return "f." + s }
func jmpLabel(s string) string { return "l." + s }

type gen struct {
	b       *Buf
	slots   map[string]int // Var name -> frame offset (from sp)
	frame   int
	strPool map[string]string // string value -> ro label
	strOrd  []string
}

// Compile lowers a whole Tacky program to a native, signed arm64 executable.
func Compile(prog tackygen.TackyProgram, outPath string) error {
	g := &gen{b: NewBuf(), strPool: map[string]string{}}

	// entry stub at offset 0: call the user entry (үндсэн), exit(its result).
	// Labels are internal map keys in this byte encoder, so the Cyrillic
	// identifiers from Tacky are used verbatim — no transliteration needed
	// (unlike the AT&T text path, which mangles to Latin: wndsen/khevle).
	g.b.BL(fnLabel("үндсэн"))
	g.b.MovImm(x16, 1) // SYS_exit
	g.b.Svc()

	for _, fn := range prog.FnDefs {
		if fn.IsExtern {
			continue
		}
		if err := g.function(fn); err != nil {
			return fmt.Errorf("%s: %w", fn.Name, err)
		}
	}

	g.emitStdlib()

	// build ro blob (strings), assign offsets.
	ro := []byte{}
	roAddr := map[string]int{} // label -> offset within ro
	for _, v := range g.strOrd {
		roAddr[g.strPool[v]] = len(ro)
		ro = append(ro, v...)
		ro = append(ro, 0)
	}

	lay := macho.PlanARM64(g.b.Len(), len(ro), 0)
	dataAddr := map[string]int{}
	for l, off := range roAddr {
		dataAddr[l] = lay.ROAddr + off
	}
	if err := g.b.Resolve(lay.CodeAddr, dataAddr); err != nil {
		return err
	}
	return macho.WriteExecutableARM64(outPath, g.b.Code(), ro, nil, 0)
}

func (g *gen) internString(v string) string {
	if l, ok := g.strPool[v]; ok {
		return l
	}
	l := fmt.Sprintf("s.%d", len(g.strOrd))
	g.strPool[v] = l
	g.strOrd = append(g.strOrd, v)
	return l
}

// assignSlots gives every Var (params + temps) an 8-byte frame slot.
func (g *gen) assignSlots(fn tackygen.TackyFn) {
	g.slots = map[string]int{}
	add := func(name string) {
		if _, ok := g.slots[name]; !ok {
			g.slots[name] = len(g.slots) * 8
		}
	}
	for _, p := range fn.Params {
		if v, ok := p.(tackygen.Var); ok {
			add(v.Name)
		}
	}
	var touch func(v tackygen.TackyVal)
	touch = func(v tackygen.TackyVal) {
		if vv, ok := v.(tackygen.Var); ok {
			add(vv.Name)
		}
	}
	for _, ins := range fn.Instructions {
		switch a := ins.(type) {
		case tackygen.Copy:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.Unary:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.Binary:
			touch(a.Src1)
			touch(a.Src2)
			touch(a.Dst)
		case tackygen.Return:
			touch(a.Value)
		case tackygen.JumpIfZero:
			touch(a.Val)
		case tackygen.JumpIfNotZero:
			touch(a.Val)
		case tackygen.FnCall:
			for _, ar := range a.Args {
				touch(ar)
			}
			touch(a.Dst)
		case tackygen.Load:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.Store:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.GetAddress:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.SignExtend:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.ZeroExtend:
			touch(a.Src)
			touch(a.Dst)
		case tackygen.Truncate:
			touch(a.Src)
			touch(a.Dst)
		}
	}
	g.frame = (len(g.slots)*8 + 15) &^ 15
}

func (g *gen) function(fn tackygen.TackyFn) error {
	g.assignSlots(fn)
	g.b.Label(fnLabel(fn.Name))
	g.b.Prologue(g.frame)
	// store incoming params (x0..x7) into their slots.
	for i, p := range fn.Params {
		if i >= 8 {
			return fmt.Errorf("more than 8 params unsupported")
		}
		if v, ok := p.(tackygen.Var); ok {
			g.b.StrFrame(i, sp, g.slots[v.Name])
		}
	}
	for _, ins := range fn.Instructions {
		if err := g.instr(fn, ins); err != nil {
			return err
		}
	}
	// fall-through safety: return 0.
	g.b.MovImm(x0, 0)
	g.b.Epilogue(g.frame)
	return nil
}

// loadVal materializes a Tacky value into reg.
func (g *gen) loadVal(reg int, v tackygen.TackyVal) error {
	switch a := v.(type) {
	case tackygen.Constant:
		g.b.MovImm(reg, a.Value.GetValue())
	case tackygen.Var:
		off, ok := g.slots[a.Name]
		if !ok {
			return fmt.Errorf("unknown var %s", a.Name)
		}
		g.b.LdrFrame(reg, sp, off)
	case tackygen.StringConstant:
		g.b.AdrpAdd(reg, g.internString(a.Value))
	default:
		return fmt.Errorf("unsupported value %T", v)
	}
	return nil
}

func (g *gen) storeVar(reg int, v tackygen.TackyVal) error {
	vv, ok := v.(tackygen.Var)
	if !ok {
		return fmt.Errorf("store to non-var %T", v)
	}
	off, ok := g.slots[vv.Name]
	if !ok {
		return fmt.Errorf("unknown dst var %s", vv.Name)
	}
	g.b.StrFrame(reg, sp, off)
	return nil
}

func (g *gen) instr(fn tackygen.TackyFn, ins tackygen.Instruction) error {
	switch a := ins.(type) {
	case tackygen.Copy:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		return g.storeVar(x8, a.Dst)

	case tackygen.SignExtend, tackygen.ZeroExtend, tackygen.Truncate:
		// all values are 64-bit in this backend; treat as a copy.
		var src, dst tackygen.TackyVal
		switch t := a.(type) {
		case tackygen.SignExtend:
			src, dst = t.Src, t.Dst
		case tackygen.ZeroExtend:
			src, dst = t.Src, t.Dst
		case tackygen.Truncate:
			src, dst = t.Src, t.Dst
		}
		if err := g.loadVal(x8, src); err != nil {
			return err
		}
		return g.storeVar(x8, dst)

	case tackygen.Unary:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		switch a.Op {
		case tackygen.Negate:
			g.b.Neg(x8, x8)
		case tackygen.Complement:
			g.b.Mvn(x8, x8)
		case tackygen.Not:
			g.b.CmpImm(x8, 0)
			g.b.Cset(x8, condEQ)
		default:
			return fmt.Errorf("unary op %s", a.Op)
		}
		return g.storeVar(x8, a.Dst)

	case tackygen.Binary:
		if err := g.loadVal(x8, a.Src1); err != nil {
			return err
		}
		if err := g.loadVal(x9, a.Src2); err != nil {
			return err
		}
		if err := g.binary(a.Op); err != nil {
			return err
		}
		return g.storeVar(x8, a.Dst)

	case tackygen.Return:
		if err := g.loadVal(x0, a.Value); err != nil {
			return err
		}
		g.b.Epilogue(g.frame)
		return nil

	case tackygen.Jump:
		g.b.B(jmpLabel(a.Target))
		return nil
	case tackygen.Label:
		g.b.Label(jmpLabel(a.Ident))
		return nil
	case tackygen.JumpIfZero:
		if err := g.loadVal(x8, a.Val); err != nil {
			return err
		}
		g.b.Cbz(x8, jmpLabel(a.Ident))
		return nil
	case tackygen.JumpIfNotZero:
		if err := g.loadVal(x8, a.Val); err != nil {
			return err
		}
		g.b.Cbnz(x8, jmpLabel(a.Ident))
		return nil

	case tackygen.FnCall:
		if len(a.Args) > 8 {
			return fmt.Errorf("more than 8 call args unsupported")
		}
		for i, ar := range a.Args {
			if err := g.loadVal(i, ar); err != nil {
				return err
			}
		}
		g.b.BL(fnLabel(a.Name))
		return g.storeVar(x0, a.Dst)

	case tackygen.Load:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.LdrReg(x8, x8)
		return g.storeVar(x8, a.Dst)
	case tackygen.Store:
		if err := g.loadVal(x9, a.Dst); err != nil {
			return err
		}
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.StrReg(x8, x9)
		return nil
	case tackygen.GetAddress:
		vv, ok := a.Src.(tackygen.Var)
		if !ok {
			return fmt.Errorf("getaddress of non-var")
		}
		g.b.AddImm(x8, sp, g.slots[vv.Name])
		return g.storeVar(x8, a.Dst)

	default:
		return fmt.Errorf("unimplemented instruction %T", ins)
	}
}

func (g *gen) binary(op tackygen.TackyBinaryOp) error {
	switch op {
	case tackygen.Add:
		g.b.Add(x8, x8, x9)
	case tackygen.Sub:
		g.b.Sub(x8, x8, x9)
	case tackygen.Mul:
		g.b.Mul(x8, x8, x9)
	case tackygen.Div:
		g.b.SDiv(x8, x8, x9)
	case tackygen.UDiv:
		g.b.UDiv(x8, x8, x9)
	case tackygen.Modulo:
		g.b.SDiv(scratchC, x8, x9)
		g.b.MSub(x8, scratchC, x9, x8)
	case tackygen.UModulo:
		g.b.UDiv(scratchC, x8, x9)
		g.b.MSub(x8, scratchC, x9, x8)
	case tackygen.Equal:
		g.cmpSet(condEQ)
	case tackygen.NotEqual:
		g.cmpSet(condNE)
	case tackygen.LessThan:
		g.cmpSet(condLT)
	case tackygen.LessThanEqual:
		g.cmpSet(condLE)
	case tackygen.GreaterThan:
		g.cmpSet(condGT)
	case tackygen.GreaterThanEqual:
		g.cmpSet(condGE)
	case tackygen.ULessThan:
		g.cmpSet(condLO)
	case tackygen.ULessThanEqual:
		g.cmpSet(condLS)
	case tackygen.UGreaterThan:
		g.cmpSet(condHI)
	case tackygen.UGreaterThanEqual:
		g.cmpSet(condHS)
	default:
		return fmt.Errorf("binary op %s", op)
	}
	return nil
}

func (g *gen) cmpSet(cond int) {
	g.b.Cmp(x8, x9)
	g.b.Cset(x8, cond)
}
