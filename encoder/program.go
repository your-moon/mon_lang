/*
 * mon_lang - asm AST to machine code mapping
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// This file lowers the post-fixup asm AST (code_gen) onto the encoder.
// Symbols share one label space, namespaced by prefix: "f." functions,
// "l." branch labels, "s." string constants, "d." globals. All calls and
// jumps are internal, so the finished code needs no relocations.
package encoder

import (
	"fmt"
	"maps"

	codegen "github.com/your-moon/mon_lang/code_gen"
	"github.com/your-moon/mon_lang/code_gen/asmtype"
)

func fnLabel(s string) string   { return "f." + s }
func jmpLabel(s string) string  { return "l." + s }
func strLabel(s string) string  { return "s." + s }
func dataLabel(s string) string { return "d." + s }

// Program is the encoded result plus the data the layout stage must place.
type Program struct {
	Code    []byte
	ROData  []byte         // string constants, appended to __TEXT
	ROAddr  map[string]int // label -> offset within ROData
	Globals []codegen.GlobalVarAsm
	// DataLabels: label -> (offset, size) within the __DATA blob built by
	// BuildDataBlob; kept here so macho and Resolve agree on placement.
	buf *Buf
}

var regMap = map[codegen.AsmRegister]Reg{
	codegen.AX: RAX, codegen.AL: RAX, codegen.CX: RCX, codegen.DX: RDX,
	codegen.DI: RDI, codegen.SI: RSI, codegen.R8: R8, codegen.R9: R9,
	codegen.R10: R10, codegen.R11: R11, codegen.SP: RSP,
}

func mustReg(r codegen.AsmRegister) Reg {
	e, ok := regMap[r]
	if !ok {
		panic(fmt.Sprintf("encoder: unmapped register %q", r))
	}
	return e
}

var condMap = map[codegen.CondCode]Cond{
	codegen.E: CondE, codegen.NE: CondNE, codegen.G: CondG,
	codegen.GE: CondGE, codegen.L: CondL, codegen.LE: CondLE,
}

func isQuad(t asmtype.AsmType) bool {
	switch t.(type) {
	case *asmtype.QuadWord, *asmtype.StringType:
		return true
	}
	return false
}

// stringPool interns string literals, mirroring x86x64.go AddString.
type stringPool struct {
	labels map[string]string
	order  []string // interned values in first-seen order
}

func (p *stringPool) intern(v string) string {
	if l, ok := p.labels[v]; ok {
		return l
	}
	l := strLabel(fmt.Sprintf("LC%d", len(p.order)))
	p.labels[v] = l
	p.order = append(p.order, v)
	return l
}

// EncodeProgram assembles the whole program: entry stub, user functions,
// stdlib. Returned Program still has unresolved data references; the macho
// layout stage calls Finish with final addresses.
func EncodeProgram(prog codegen.AsmProgram) (*Program, error) {
	b := New()
	pool := &stringPool{labels: map[string]string{}}

	stdlib := map[string]bool{}
	for _, n := range StdlibFns() {
		stdlib[n] = true
	}
	defined := map[string]bool{}
	for _, fn := range prog.AsmFnDef {
		defined[fn.Ident] = true
	}
	// Every extern must be satisfied by the built-in stdlib now that no
	// external linker resolves symbols.
	for _, ext := range prog.AsmExternFn {
		if !stdlib[ext.Name] && !defined[ext.Name] {
			return nil, fmt.Errorf("encoder: тодорхойгүй гадаад функц: %s", ext.Name)
		}
	}

	b.EmitEntry(fnLabel("wndsen"))

	for _, fn := range prog.AsmFnDef {
		b.Label(fnLabel(fn.Ident))
		b.Prologue()
		for _, instr := range fn.Irs {
			if err := encodeInstr(b, pool, instr); err != nil {
				return nil, fmt.Errorf("%s: %w", fn.Ident, err)
			}
		}
	}

	b.EmitStdlib(fnLabel, dataLabel, strLabel)

	// Build the read-only blob: interned program strings + stdlib strings.
	ro := []byte{}
	roAddr := map[string]int{}
	for _, v := range pool.order {
		roAddr[pool.labels[v]] = len(ro)
		ro = append(ro, v...)
		ro = append(ro, 0)
	}
	for name, v := range StdlibStrings() {
		roAddr[strLabel(name)] = len(ro)
		ro = append(ro, v...)
		ro = append(ro, 0)
	}

	return &Program{Code: b.Code(), ROData: ro, ROAddr: roAddr,
		Globals: prog.GlobalVars, buf: b}, nil
}

// Finish patches all label references. roBase and dataBase are the final
// addresses of the two blobs expressed relative to the code start (the
// encoder emits rel32, so only deltas matter).
func (p *Program) Finish(roBase int, dataBase map[string]int) error {
	addr := map[string]int{}
	for l, off := range p.ROAddr {
		addr[l] = roBase + off
	}
	maps.Copy(addr, dataBase)
	return p.buf.Resolve(addr)
}

/* --- operand classification --- */

type opKind int

const (
	oReg opKind = iota
	oImm
	oStack
	oRip
	oStr
)

func classify(op codegen.AsmOperand) opKind {
	switch op.(type) {
	case codegen.Register:
		return oReg
	case codegen.Imm:
		return oImm
	case codegen.Stack:
		return oStack
	case codegen.RipRelative:
		return oRip
	case codegen.StringLiteral:
		return oStr
	case codegen.Pseudo:
		panic("encoder: pseudo operand survived replacement pass")
	}
	panic(fmt.Sprintf("encoder: unknown operand %T", op))
}

func opReg(op codegen.AsmOperand) Reg    { return mustReg(op.(codegen.Register).Reg) }
func opImm(op codegen.AsmOperand) int64  { return op.(codegen.Imm).Value }
func opDisp(op codegen.AsmOperand) int   { return op.(codegen.Stack).Value }
func opRip(op codegen.AsmOperand) string { return dataLabel(op.(codegen.RipRelative).Label) }

func encodeInstr(b *Buf, pool *stringPool, instr codegen.AsmInstruction) error {
	switch ast := instr.(type) {

	case codegen.Comment:
		return nil

	case codegen.StringLiteral:
		// Bare literal: address materialized in rax (mirrors text emitter).
		b.LeaRip(RAX, pool.intern(ast.Value))
		return nil

	case codegen.AsmMov:
		return encodeMov(b, pool, ast)

	case codegen.AsmMovSx:
		switch classify(ast.Src) {
		case oReg:
			b.MovsxdRR(opReg(ast.Src), opReg(ast.Dst))
		case oStack:
			b.MovsxdMR(opDisp(ast.Src), opReg(ast.Dst))
		case oRip:
			b.MovsxdRipR(opRip(ast.Src), opReg(ast.Dst))
		default:
			return fmt.Errorf("movsx: unsupported src %T", ast.Src)
		}
		return nil

	case codegen.AsmLea:
		dst := opReg(ast.Dst) // fixup guarantees a register destination
		switch classify(ast.Src) {
		case oStack:
			b.LeaRBP(opDisp(ast.Src), dst)
		case oRip:
			b.LeaRip(dst, opRip(ast.Src))
		default:
			return fmt.Errorf("lea: unsupported src %T", ast.Src)
		}
		return nil

	case codegen.AsmBinary:
		return encodeBinary(b, ast)

	case codegen.Cmp:
		return encodeAlu(b, OpCmp, isQuad(ast.Type), ast.Src, ast.Dst)

	case codegen.Unary:
		w := isQuad(ast.Type)
		switch classify(ast.Dst) {
		case oReg:
			if ast.Op == codegen.Not {
				b.NotR(w, opReg(ast.Dst))
			} else {
				b.NegR(w, opReg(ast.Dst))
			}
		case oStack:
			if ast.Op == codegen.Not {
				b.NotM(w, opDisp(ast.Dst))
			} else {
				b.NegM(w, opDisp(ast.Dst))
			}
		default:
			return fmt.Errorf("unary: unsupported dst %T", ast.Dst)
		}
		return nil

	case codegen.Idiv:
		w := isQuad(ast.Type)
		switch classify(ast.Src) {
		case oReg:
			b.IdivR(w, opReg(ast.Src))
		case oStack:
			b.IdivM(w, opDisp(ast.Src))
		default:
			return fmt.Errorf("idiv: unsupported src %T", ast.Src)
		}
		return nil

	case codegen.Cdq:
		if isQuad(ast.Type) {
			b.Cqo()
		} else {
			b.Cdq()
		}
		return nil

	case codegen.SetCC:
		cc := condMap[ast.CC]
		switch classify(ast.Op) {
		case oReg:
			b.SetccR(cc, opReg(ast.Op))
		case oStack:
			b.SetccM(cc, opDisp(ast.Op))
		case oRip:
			b.SetccRip(cc, opRip(ast.Op))
		default:
			return fmt.Errorf("setcc: unsupported operand %T", ast.Op)
		}
		return nil

	case codegen.Jmp:
		b.Jmp(jmpLabel(ast.Ident))
		return nil

	case codegen.JmpCC:
		b.Jcc(condMap[ast.CC], jmpLabel(ast.Ident))
		return nil

	case codegen.Label:
		b.Label(jmpLabel(ast.Ident))
		return nil

	case codegen.Call:
		b.Call(fnLabel(ast.Ident))
		return nil

	case codegen.Push:
		switch classify(ast.Op) {
		case oReg:
			b.PushR(opReg(ast.Op))
		case oImm:
			b.PushI(opImm(ast.Op))
		default:
			return fmt.Errorf("push: unsupported operand %T", ast.Op)
		}
		return nil

	case codegen.AsmLoadFromMem:
		b.LoadBaseR(isQuad(ast.Type), mustReg(ast.Base), opReg(ast.Dst))
		return nil

	case codegen.AsmStoreToMem:
		b.StoreRBase(isQuad(ast.Type), opReg(ast.Src), mustReg(ast.Base))
		return nil

	case codegen.Return:
		b.Epilogue()
		return nil

	case codegen.AllocateStack:
		b.AluIR(OpSub, true, int64(ast.Value), RSP)
		return nil

	case codegen.AsmExternFn:
		return nil // satisfied by the built-in stdlib, checked upfront

	default:
		return fmt.Errorf("encoder: unimplemented instruction %T", instr)
	}
}

func encodeMov(b *Buf, pool *stringPool, ast codegen.AsmMov) error {
	// String source: materialize address in rax, then store (text emitter
	// contract: leaq .LCn(%rip), %rax [+ mov %rax, dst]).
	if lit, ok := ast.Src.(codegen.StringLiteral); ok {
		b.LeaRip(RAX, pool.intern(lit.Value))
		if r, isReg := ast.Dst.(codegen.Register); isReg && mustReg(r.Reg) == RAX {
			return nil
		}
		return encodeMov(b, pool, codegen.AsmMov{Type: ast.Type,
			Src: codegen.Register{Reg: codegen.AX}, Dst: ast.Dst})
	}

	w := isQuad(ast.Type)
	switch classify(ast.Src) {
	case oReg:
		switch classify(ast.Dst) {
		case oReg:
			b.MovRR(w, opReg(ast.Src), opReg(ast.Dst))
		case oStack:
			b.MovRM(w, opReg(ast.Src), opDisp(ast.Dst))
		case oRip:
			b.MovRRip(w, opReg(ast.Src), opRip(ast.Dst))
		default:
			return fmt.Errorf("mov: reg to %T", ast.Dst)
		}
	case oImm:
		v := opImm(ast.Src)
		switch classify(ast.Dst) {
		case oReg:
			b.MovIR(w, v, opReg(ast.Dst))
		case oStack:
			b.MovIM(w, v, opDisp(ast.Dst))
		case oRip:
			b.MovIRip(w, v, opRip(ast.Dst))
		default:
			return fmt.Errorf("mov: imm to %T", ast.Dst)
		}
	case oStack:
		switch classify(ast.Dst) {
		case oReg:
			b.MovMR(w, opDisp(ast.Src), opReg(ast.Dst))
		default:
			return fmt.Errorf("mov: mem to %T survived fixup", ast.Dst)
		}
	case oRip:
		switch classify(ast.Dst) {
		case oReg:
			b.MovRipR(w, opRip(ast.Src), opReg(ast.Dst))
		default:
			return fmt.Errorf("mov: rip to %T survived fixup", ast.Dst)
		}
	default:
		return fmt.Errorf("mov: unsupported src %T", ast.Src)
	}
	return nil
}

func encodeBinary(b *Buf, ast codegen.AsmBinary) error {
	w := isQuad(ast.Type)
	switch ast.Op {
	case codegen.Add:
		return encodeAlu(b, OpAdd, w, ast.Src, ast.Dst)
	case codegen.Sub:
		return encodeAlu(b, OpSub, w, ast.Src, ast.Dst)
	case codegen.Mult:
		// fixup guarantees a register destination
		dst := opReg(ast.Dst)
		switch classify(ast.Src) {
		case oReg:
			b.ImulRR(w, opReg(ast.Src), dst)
		case oImm:
			b.ImulIR(w, opImm(ast.Src), dst)
		case oStack:
			b.ImulMR(w, opDisp(ast.Src), dst)
		default:
			return fmt.Errorf("imul: unsupported src %T", ast.Src)
		}
		return nil
	}
	return fmt.Errorf("binary: unknown op %s", ast.Op)
}

func encodeAlu(b *Buf, op aluOp, w bool, src, dst codegen.AsmOperand) error {
	switch classify(src) {
	case oReg:
		switch classify(dst) {
		case oReg:
			b.AluRR(op, w, opReg(src), opReg(dst))
		case oStack:
			b.AluRM(op, w, opReg(src), opDisp(dst))
		case oRip:
			b.AluRRip(op, w, opReg(src), opRip(dst))
		default:
			return fmt.Errorf("alu: reg to %T", dst)
		}
	case oImm:
		v := opImm(src)
		switch classify(dst) {
		case oReg:
			b.AluIR(op, w, v, opReg(dst))
		case oStack:
			b.AluIM(op, w, v, opDisp(dst))
		case oRip:
			b.AluIRip(op, w, v, opRip(dst))
		default:
			return fmt.Errorf("alu: imm to %T", dst)
		}
	case oStack:
		switch classify(dst) {
		case oReg:
			b.AluMR(op, w, opDisp(src), opReg(dst))
		default:
			return fmt.Errorf("alu: mem/mem survived fixup (%T)", dst)
		}
	case oRip:
		switch classify(dst) {
		case oReg:
			b.AluRipR(op, w, opRip(src), opReg(dst))
		default:
			return fmt.Errorf("alu: rip/%T survived fixup", dst)
		}
	default:
		return fmt.Errorf("alu: unsupported src %T", src)
	}
	return nil
}
