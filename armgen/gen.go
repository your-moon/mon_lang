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
	"github.com/your-moon/mon_lang/mconstant"
	"github.com/your-moon/mon_lang/tackygen"
)

const scratchC = 10 // third scratch (x10), for modulo

func fnLabel(s string) string  { return "f." + s }
func jmpLabel(s string) string { return "l." + s }

type gen struct {
	b       *Buf
	slots   map[string]int    // Var name -> frame offset (from sp)
	frame   int
	strPool map[string]string // string value -> ro label
	strOrd  []string
	globals map[string]string // global var name -> data label
	sizeOf  func(string) int  // temp/var name -> byte width (4 or 8)
	isDbl   func(string) bool // temp/var name -> is a double
}

func dataLabel(s string) string { return "g." + s }

// Compile lowers a whole Tacky program to a native, signed arm64 executable.
// sizeOf reports the byte width (4 or 8) of a Tacky temp/var; it drives the
// width of pointer loads/stores so sub-word (Int32) struct fields aren't read
// or written 8 bytes wide (which would corrupt neighbours). Pass nil to treat
// everything as 64-bit.
func Compile(prog tackygen.TackyProgram, sizeOf func(string) int, isDbl func(string) bool, outPath string) error {
	if sizeOf == nil {
		sizeOf = func(string) int { return 8 }
	}
	if isDbl == nil {
		isDbl = func(string) bool { return false }
	}
	g := &gen{b: NewBuf(), strPool: map[string]string{}, globals: map[string]string{}, sizeOf: sizeOf, isDbl: isDbl}

	// __DATA blob: every global gets an 8-byte little-endian slot seeded with
	// its initial value (all scalars/pointers are 64-bit in this backend).
	data := []byte{}
	dataOff := map[string]int{}
	for _, gv := range prog.GlobalVars {
		lbl := dataLabel(gv.Name)
		g.globals[gv.Name] = lbl
		dataOff[lbl] = len(data)
		v := uint64(gv.InitValue)
		for i := 0; i < 8; i++ {
			data = append(data, byte(v>>(8*i)))
		}
	}
	// reserved heap + argv state (all start at 0). argc/argv are captured by
	// the entry stub, which dyld calls as main(argc, argv, ...).
	for _, lbl := range []string{heapPtrLabel, heapEndLabel, argcLabel, argvLabel} {
		dataOff[lbl] = len(data)
		data = append(data, make([]byte, 8)...)
	}

	// entry stub at offset 0: call the user entry (үндсэн), exit(its result).
	// Labels are internal map keys in this byte encoder, so the Cyrillic
	// identifiers from Tacky are used verbatim — no transliteration needed
	// (unlike the AT&T text path, which mangles to Latin: wndsen/khevle).
	g.b.AdrpAdd(9, argcLabel) // dyld calls entry as main(argc,argv,...)
	g.b.StrReg(x0, 9)
	g.b.AdrpAdd(9, argvLabel)
	g.b.StrReg(x1, 9)
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

	lay := macho.PlanARM64(g.b.Len(), len(ro), len(data))
	dataAddr := map[string]int{}
	for l, off := range roAddr {
		dataAddr[l] = lay.ROAddr + off
	}
	for l, off := range dataOff {
		dataAddr[l] = lay.DataAddr + off
	}
	if err := g.b.Resolve(lay.CodeAddr, dataAddr); err != nil {
		return err
	}
	return macho.WriteExecutableARM64(outPath, g.b.Code(), ro, data, 0)
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
		if _, isGlobal := g.globals[name]; isGlobal {
			return // globals live in __DATA, not the frame
		}
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
		if lbl, isGlobal := g.globals[a.Name]; isGlobal {
			g.b.AdrpAdd(reg, lbl) // reg = &global
			g.b.LdrReg(reg, reg)  // reg = *reg
			return nil
		}
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
	if lbl, isGlobal := g.globals[vv.Name]; isGlobal {
		// address scratch must differ from the value reg; x8 values use x9.
		addr := x9
		if reg == x9 {
			addr = x8
		}
		g.b.AdrpAdd(addr, lbl)
		g.b.StrReg(reg, addr)
		return nil
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

	case tackygen.SignExtend:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.Sxtw(x8, x8) // sign-extend low 32 bits (fixes wide reads of Int32)
		return g.storeVar(x8, a.Dst)
	case tackygen.ZeroExtend:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.Uxtw(x8, x8)
		return g.storeVar(x8, a.Dst)
	case tackygen.Truncate:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.Uxtw(x8, x8) // keep the low 32 bits
		return g.storeVar(x8, a.Dst)

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
		// FP when the result is a double (arithmetic) or the operands are
		// doubles (comparison, whose result is an int).
		if g.isDblVal(a.Dst) || g.isDblVal(a.Src1) {
			if err := g.fpBinary(a.Op); err != nil {
				return err
			}
			return g.storeVar(x8, a.Dst)
		}
		if err := g.binary(a.Op); err != nil {
			return err
		}
		return g.storeVar(x8, a.Dst)

	case tackygen.IntToDouble:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.Scvtf(0, x8)     // d0 = (double)x8
		g.b.FmovDtoX(x8, 0)  // x8 = bits(d0)
		return g.storeVar(x8, a.Dst)
	case tackygen.DoubleToInt:
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.FmovXtoD(0, x8)  // d0 = bits
		g.b.Fcvtzs(x8, 0)    // x8 = (int64)d0
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
		// Width from the loaded value's type: a pointer element stays 8 bytes
		// (must not be truncated), a signed Int32 field loads sign-extended.
		g.b.LdrW(x8, x8, g.valWidth(a.Dst))
		return g.storeVar(x8, a.Dst)
	case tackygen.Store:
		if err := g.loadVal(x9, a.Dst); err != nil {
			return err
		}
		if err := g.loadVal(x8, a.Src); err != nil {
			return err
		}
		g.b.StrW(x8, x9, g.valWidth(a.Src)) // width = stored value's type
		return nil
	case tackygen.GetAddress:
		vv, ok := a.Src.(tackygen.Var)
		if !ok {
			return fmt.Errorf("getaddress of non-var")
		}
		if lbl, isGlobal := g.globals[vv.Name]; isGlobal {
			g.b.AdrpAdd(x8, lbl)
		} else {
			g.b.AddImm(x8, sp, g.slots[vv.Name])
		}
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

// isDblVal reports whether a Tacky value is a double.
func (g *gen) isDblVal(v tackygen.TackyVal) bool {
	switch a := v.(type) {
	case tackygen.Var:
		return g.isDbl(a.Name)
	case tackygen.Constant:
		_, ok := a.Value.(mconstant.Float64)
		return ok
	}
	return false
}

// fpBinary applies a double op. Operand bit patterns are in x8/x9; the result
// (or 0/1 for a comparison) lands back in x8.
func (g *gen) fpBinary(op tackygen.TackyBinaryOp) error {
	g.b.FmovXtoD(0, x8)
	g.b.FmovXtoD(1, x9)
	switch op {
	case tackygen.Add:
		g.b.Fadd(0, 0, 1)
	case tackygen.Sub:
		g.b.Fsub(0, 0, 1)
	case tackygen.Mul:
		g.b.Fmul(0, 0, 1)
	case tackygen.Div:
		g.b.Fdiv(0, 0, 1)
	case tackygen.Equal:
		return g.fpCmp(condEQ)
	case tackygen.NotEqual:
		return g.fpCmp(condNE)
	case tackygen.GreaterThan:
		return g.fpCmp(condGT)
	case tackygen.GreaterThanEqual:
		return g.fpCmp(condGE)
	case tackygen.LessThan:
		return g.fpCmp(condLO) // ordered <: carry clear after fcmp
	case tackygen.LessThanEqual:
		return g.fpCmp(condLS) // ordered <=: C clear or Z set
	default:
		return fmt.Errorf("fp binary op %s", op)
	}
	g.b.FmovDtoX(x8, 0)
	return nil
}

func (g *gen) fpCmp(cond int) error {
	g.b.Fcmp(0, 1)
	g.b.Cset(x8, cond)
	return nil
}

// valWidth reports the byte width (4 or 8) of a Tacky value, used to size
// pointer loads/stores. Vars/temps consult the symbol sizes; a constant's
// width comes from its mconstant type (Int32 vs Int64); strings are pointers.
func (g *gen) valWidth(v tackygen.TackyVal) int {
	switch a := v.(type) {
	case tackygen.Var:
		w := g.sizeOf(a.Name)
		if w == 4 {
			return 4
		}
		return 8
	case tackygen.Constant:
		if _, is32 := a.Value.(mconstant.Int32); is32 {
			return 4
		}
		return 8 // Int64 and Float64 (bit pattern) are 8 bytes
	default:
		return 8
	}
}
