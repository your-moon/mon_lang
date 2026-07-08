/*
 * mon_lang - x86_64 instruction encoder
 *
 * Copyright (c) 2026 Munkherdene
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

// Package encoder assembles the compiler's closed x86_64 subset into machine
// code. Labels use TCC-style backpatching: forward references emit a rel32
// placeholder recorded in fixups, patched once the label address is known.
// The whole program lives in one flat buffer, so every call/jump is internal
// rel32 and the output needs no relocations or symbol table.
package encoder

import (
	"encoding/binary"
	"fmt"
)

// Reg is an x86_64 register number as used in ModRM/REX encoding.
type Reg int

const (
	RAX Reg = 0
	RCX Reg = 1
	RDX Reg = 2
	RBX Reg = 3
	RSP Reg = 4
	RBP Reg = 5
	RSI Reg = 6
	RDI Reg = 7
	R8  Reg = 8
	R9  Reg = 9
	R10 Reg = 10
	R11 Reg = 11
)

// Cond is a condition code: the low nibble of the 0F 8x / 0F 9x opcodes.
type Cond int

const (
	CondE  Cond = 0x4
	CondNE Cond = 0x5
	CondL  Cond = 0xC
	CondGE Cond = 0xD
	CondLE Cond = 0xE
	CondG  Cond = 0xF
)

type fixup struct {
	pos   int    // buffer offset of the rel32/disp32 field
	label string // target label
	// rel32 is pc-relative to the end of the field; disp32 fixups (dataRef)
	// are patched to an absolute delta by the layout stage instead.
	dataRef bool
	// rip-relative disp32 is measured from the END of the instruction; when
	// an immediate follows the disp field, extra is its size in bytes.
	extra int
}

// Buf accumulates machine code and resolves labels.
type Buf struct {
	code   []byte
	labels map[string]int
	fixups []fixup
}

func New() *Buf {
	return &Buf{labels: map[string]int{}}
}

func (b *Buf) Len() int     { return len(b.code) }
func (b *Buf) Code() []byte { return b.code }

func (b *Buf) byte(v ...byte) { b.code = append(b.code, v...) }
func (b *Buf) u32(v uint32)   { b.code = binary.LittleEndian.AppendUint32(b.code, v) }
func (b *Buf) u64(v uint64)   { b.code = binary.LittleEndian.AppendUint64(b.code, v) }

// Label defines label name at the current position. Defining a name twice is
// a compiler bug, so it panics.
func (b *Buf) Label(name string) {
	if _, dup := b.labels[name]; dup {
		panic("encoder: duplicate label " + name)
	}
	b.labels[name] = len(b.code)
}

func (b *Buf) LabelPos(name string) (int, bool) {
	p, ok := b.labels[name]
	return p, ok
}

// rel32 emits a 4-byte pc-relative reference to label.
func (b *Buf) rel32(label string) {
	b.fixups = append(b.fixups, fixup{pos: len(b.code), label: label})
	b.u32(0)
}

// DataRel32 emits a 4-byte rip-relative reference to a data label whose
// address is only known at layout time (strings, globals).
func (b *Buf) DataRel32(label string) {
	b.fixups = append(b.fixups, fixup{pos: len(b.code), label: label, dataRef: true})
	b.u32(0)
}

// dataRel32Imm is DataRel32 for instructions where an immediate follows the
// disp field: the CPU adds disp32 to the address of the NEXT instruction, so
// the immediate's size must be part of the fixup math.
func (b *Buf) dataRel32Imm(label string, immSize int) {
	b.fixups = append(b.fixups, fixup{pos: len(b.code), label: label, dataRef: true, extra: immSize})
	b.u32(0)
}

// Resolve patches all code-internal fixups. dataAddr gives the final address
// of each data label relative to the code start (may be past the code end);
// callers compute it during layout.
func (b *Buf) Resolve(dataAddr map[string]int) error {
	for _, f := range b.fixups {
		var target int
		if f.dataRef {
			t, ok := dataAddr[f.label]
			if !ok {
				return fmt.Errorf("encoder: undefined data label %s", f.label)
			}
			target = t
		} else {
			t, ok := b.labels[f.label]
			if !ok {
				return fmt.Errorf("encoder: undefined label %s", f.label)
			}
			target = t
		}
		rel := int32(target - (f.pos + 4 + f.extra))
		binary.LittleEndian.PutUint32(b.code[f.pos:], uint32(rel))
	}
	return nil
}

/* --- instruction encoding helpers --- */

// rex builds a REX prefix; w=1 for 64-bit operand size. reg/rm are the
// extension bits for the ModRM reg and rm fields.
func rexByte(w bool, reg, rm Reg) byte {
	r := byte(0x40)
	if w {
		r |= 8
	}
	if reg >= R8 {
		r |= 4
	}
	if rm >= R8 {
		r |= 1
	}
	return r
}

func (b *Buf) rex(w bool, reg, rm Reg) {
	r := rexByte(w, reg, rm)
	// REX is optional for 32-bit ops on low registers, but SPL/BPL/SIL/DIL
	// byte ops need it; emitting it unconditionally when != 0x40 keeps the
	// encoder simple and only costs a byte.
	if r != 0x40 || w {
		b.byte(r)
	}
}

func modrm(mod byte, reg, rm Reg) byte {
	return mod<<6 | byte(reg&7)<<3 | byte(rm&7)
}

// memRBP encodes ModRM+disp for disp(%rbp).
func (b *Buf) memRBP(reg Reg, disp int) {
	if disp >= -128 && disp <= 127 {
		b.byte(modrm(1, reg, RBP), byte(int8(disp)))
	} else {
		b.byte(modrm(2, reg, RBP))
		b.u32(uint32(int32(disp)))
	}
}

// memBase encodes (base) with no displacement; base must not be RSP/RBP/R12/R13
// except R10/R11 which this compiler uses.
func (b *Buf) memBase(reg, base Reg) {
	b.byte(modrm(0, reg, base))
}

/* --- register-register / register-memory forms ---
   Each instruction the code_gen AST needs gets one exported method.
   w selects 64-bit (true) vs 32-bit (false) operand size. */

func (b *Buf) MovRR(w bool, src, dst Reg) { // mov src, dst
	b.rex(w, src, dst)
	b.byte(0x89, modrm(3, src, dst))
}

func (b *Buf) MovRM(w bool, src Reg, disp int) { // mov src, disp(%rbp)
	b.rex(w, src, RBP)
	b.byte(0x89)
	b.memRBP(src, disp)
}

func (b *Buf) MovMR(w bool, disp int, dst Reg) { // mov disp(%rbp), dst
	b.rex(w, dst, RBP)
	b.byte(0x8B)
	b.memRBP(dst, disp)
}

// MovIR loads an immediate. 32-bit ops use the C7 /0 imm32 form; 64-bit uses
// movabs when the value doesn't fit in a sign-extended imm32.
func (b *Buf) MovIR(w bool, imm int64, dst Reg) {
	if w && (imm > 0x7fffffff || imm < -0x80000000) {
		b.byte(rexByte(true, 0, dst), 0xB8+byte(dst&7))
		b.u64(uint64(imm))
		return
	}
	b.rex(w, 0, dst)
	b.byte(0xC7, modrm(3, 0, dst))
	b.u32(uint32(int32(imm)))
}

func (b *Buf) MovIM(w bool, imm int64, disp int) { // mov $imm, disp(%rbp)
	if imm > 0x7fffffff || imm < -0x80000000 {
		panic("encoder: imm64 to memory needs a scratch register")
	}
	b.rex(w, 0, RBP)
	b.byte(0xC7)
	b.memRBP(0, disp)
	b.u32(uint32(int32(imm)))
}

func (b *Buf) MovsxdRR(src, dst Reg) { // movslq src32, dst64
	b.rex(true, dst, src)
	b.byte(0x63, modrm(3, dst, src))
}

func (b *Buf) MovsxdMR(disp int, dst Reg) { // movslq disp(%rbp), dst64
	b.rex(true, dst, RBP)
	b.byte(0x63)
	b.memRBP(dst, disp)
}

func (b *Buf) MovsxdRipR(label string, dst Reg) { // movslq label(%rip), dst64
	b.rex(true, dst, RBP)
	b.byte(0x63, modrm(0, dst, RBP))
	b.DataRel32(label)
}

// LeaRip emits lea label(%rip), dst against a data label.
func (b *Buf) LeaRip(dst Reg, label string) {
	b.rex(true, dst, RBP)
	b.byte(0x8D, modrm(0, dst, RBP)) // mod=00 rm=101 = rip-relative
	b.DataRel32(label)
}

// MovRipR / MovRRip access a global: mov label(%rip), dst and mov src, label(%rip).
func (b *Buf) MovRipR(w bool, label string, dst Reg) {
	b.rex(w, dst, RBP)
	b.byte(0x8B, modrm(0, dst, RBP))
	b.DataRel32(label)
}

func (b *Buf) MovRRip(w bool, src Reg, label string) {
	b.rex(w, src, RBP)
	b.byte(0x89, modrm(0, src, RBP))
	b.DataRel32(label)
}

func (b *Buf) MovIRip(w bool, imm int64, label string) { // mov $imm, label(%rip)
	if imm > 0x7fffffff || imm < -0x80000000 {
		panic("encoder: imm64 to memory needs a scratch register")
	}
	b.rex(w, 0, RBP)
	b.byte(0xC7, modrm(0, 0, RBP))
	b.dataRel32Imm(label, 4)
	b.u32(uint32(int32(imm)))
}

// AluIRip applies an ALU op with immediate to a global: op $imm, label(%rip).
func (b *Buf) AluIRip(op aluOp, w bool, imm int64, label string) {
	b.rex(w, op.slash, RBP)
	b.byte(0x81, modrm(0, op.slash, RBP))
	b.dataRel32Imm(label, 4)
	b.u32(uint32(int32(imm)))
}

/* --- ALU: add/sub/imul/cmp share operand shapes --- */

type aluOp struct {
	rmR   byte // opcode: r/m, reg form (add=01, sub=29, cmp=39)
	rRM   byte // opcode: reg, r/m form (add=03, sub=2B, cmp=3B)
	slash Reg  // /digit for the 81 imm32 form
}

var (
	OpAdd = aluOp{0x01, 0x03, 0}
	OpSub = aluOp{0x29, 0x2B, 5}
	OpCmp = aluOp{0x39, 0x3B, 7}
)

func (b *Buf) AluRR(op aluOp, w bool, src, dst Reg) {
	b.rex(w, src, dst)
	b.byte(op.rmR, modrm(3, src, dst))
}

func (b *Buf) AluRM(op aluOp, w bool, src Reg, disp int) {
	b.rex(w, src, RBP)
	b.byte(op.rmR)
	b.memRBP(src, disp)
}

func (b *Buf) AluMR(op aluOp, w bool, disp int, dst Reg) {
	b.rex(w, dst, RBP)
	b.byte(op.rRM)
	b.memRBP(dst, disp)
}

func (b *Buf) AluIR(op aluOp, w bool, imm int64, dst Reg) {
	b.rex(w, 0, dst)
	b.byte(0x81, modrm(3, op.slash, dst))
	b.u32(uint32(int32(imm)))
}

func (b *Buf) AluIM(op aluOp, w bool, imm int64, disp int) {
	b.rex(w, 0, RBP)
	b.byte(0x81)
	b.memRBP(op.slash, disp)
	b.u32(uint32(int32(imm)))
}

// AluRipR compares/adds a global against a register (reg, r/m form).
func (b *Buf) AluRipR(op aluOp, w bool, label string, dst Reg) {
	b.rex(w, dst, RBP)
	b.byte(op.rRM, modrm(0, dst, RBP))
	b.DataRel32(label)
}

func (b *Buf) AluRRip(op aluOp, w bool, src Reg, label string) {
	b.rex(w, src, RBP)
	b.byte(op.rmR, modrm(0, src, RBP))
	b.DataRel32(label)
}

/* imul has its own encodings: 0F AF reg,r/m and 69 reg,r/m,imm32 */

func (b *Buf) ImulRR(w bool, src, dst Reg) { // imul src, dst
	b.rex(w, dst, src)
	b.byte(0x0F, 0xAF, modrm(3, dst, src))
}

func (b *Buf) ImulMR(w bool, disp int, dst Reg) { // imul disp(%rbp), dst
	b.rex(w, dst, RBP)
	b.byte(0x0F, 0xAF)
	b.memRBP(dst, disp)
}

func (b *Buf) ImulIR(w bool, imm int64, dst Reg) { // imul $imm, dst, dst
	b.rex(w, dst, dst)
	b.byte(0x69, modrm(3, dst, dst))
	b.u32(uint32(int32(imm)))
}

/* unary group F7: /2 not, /3 neg, /7 idiv */

func (b *Buf) unaryF7(w bool, slash Reg, rm Reg) {
	b.rex(w, 0, rm)
	b.byte(0xF7, modrm(3, slash, rm))
}

func (b *Buf) NotR(w bool, r Reg)  { b.unaryF7(w, 2, r) }
func (b *Buf) NegR(w bool, r Reg)  { b.unaryF7(w, 3, r) }
func (b *Buf) IdivR(w bool, r Reg) { b.unaryF7(w, 7, r) }

func (b *Buf) unaryF7M(w bool, slash Reg, disp int) {
	b.rex(w, 0, RBP)
	b.byte(0xF7)
	b.memRBP(slash, disp)
}

func (b *Buf) NotM(w bool, disp int)  { b.unaryF7M(w, 2, disp) }
func (b *Buf) NegM(w bool, disp int)  { b.unaryF7M(w, 3, disp) }
func (b *Buf) IdivM(w bool, disp int) { b.unaryF7M(w, 7, disp) }

func (b *Buf) Cdq() { b.byte(0x99) }
func (b *Buf) Cqo() { b.byte(0x48, 0x99) }

/* setcc: 0F 90+cc r/m8; memory form writes the low byte of the slot */

func (b *Buf) SetccR(c Cond, r Reg) {
	// byte-register access to SPL..DIL/R8B.. needs REX
	if r >= RSP {
		b.byte(rexByte(false, 0, r))
	}
	b.byte(0x0F, 0x90+byte(c), modrm(3, 0, r))
}

func (b *Buf) SetccM(c Cond, disp int) {
	b.byte(0x0F, 0x90+byte(c))
	b.memRBP(0, disp)
}

func (b *Buf) SetccRip(c Cond, label string) {
	b.byte(0x0F, 0x90+byte(c), modrm(0, 0, RBP))
	b.DataRel32(label)
}

/* control flow */

func (b *Buf) Jmp(label string) {
	b.byte(0xE9)
	b.rel32(label)
}

func (b *Buf) Jcc(c Cond, label string) {
	b.byte(0x0F, 0x80+byte(c))
	b.rel32(label)
}

func (b *Buf) Call(label string) {
	b.byte(0xE8)
	b.rel32(label)
}

func (b *Buf) Ret() { b.byte(0xC3) }

func (b *Buf) PushR(r Reg) {
	if r >= R8 {
		b.byte(0x41)
	}
	b.byte(0x50 + byte(r&7))
}

func (b *Buf) PopR(r Reg) {
	if r >= R8 {
		b.byte(0x41)
	}
	b.byte(0x58 + byte(r&7))
}

func (b *Buf) PushI(imm int64) {
	b.byte(0x68)
	b.u32(uint32(int32(imm)))
}

/* loads/stores through a base register: mov (%base), dst / mov src, (%base) */

func (b *Buf) LoadBaseR(w bool, base, dst Reg) {
	b.rex(w, dst, base)
	b.byte(0x8B)
	b.memBase(dst, base)
}

func (b *Buf) StoreRBase(w bool, src, base Reg) {
	b.rex(w, src, base)
	b.byte(0x89)
	b.memBase(src, base)
}

// Byte-granularity forms used by the syscall stdlib (itoa/parse loops).
func (b *Buf) LoadByte(base, dst Reg) { // movzbl (%base), dst
	b.rex(false, dst, base)
	b.byte(0x0F, 0xB6)
	b.memBase(dst, base)
}

func (b *Buf) StoreByte(src, base Reg) { // mov %srcb, (%base)
	b.byte(rexByte(false, src, base), 0x88, modrm(0, src, base))
}

func (b *Buf) IncR(w bool, r Reg) { b.rex(w, 0, r); b.byte(0xFF, modrm(3, 0, r)) }
func (b *Buf) DecR(w bool, r Reg) { b.rex(w, 0, r); b.byte(0xFF, modrm(3, 1, r)) }

func (b *Buf) XorRR(w bool, src, dst Reg) {
	b.rex(w, src, dst)
	b.byte(0x31, modrm(3, src, dst))
}

func (b *Buf) TestRR(w bool, a, r Reg) {
	b.rex(w, a, r)
	b.byte(0x85, modrm(3, a, r))
}

func (b *Buf) Syscall() { b.byte(0x0F, 0x05) }

/* --- ops used only by the built-in stdlib --- */

func (b *Buf) DivR(w bool, r Reg) { b.unaryF7(w, 6, r) } // unsigned div

func (b *Buf) ShlIR(w bool, n byte, r Reg) { // shl $n, r
	b.rex(w, 0, r)
	b.byte(0xC1, modrm(3, 4, r), n)
}

func (b *Buf) ShrIR(w bool, n byte, r Reg) { // shr $n, r
	b.rex(w, 0, r)
	b.byte(0xC1, modrm(3, 5, r), n)
}

func (b *Buf) OrIR(w bool, imm int64, r Reg) { // or $imm, r
	b.rex(w, 0, r)
	b.byte(0x81, modrm(3, 1, r))
	b.u32(uint32(int32(imm)))
}

func (b *Buf) AndIR(w bool, imm int64, r Reg) { // and $imm, r
	b.rex(w, 0, r)
	b.byte(0x81, modrm(3, 4, r))
	b.u32(uint32(int32(imm)))
}

// LeaRBP computes lea disp(%rbp), dst.
func (b *Buf) LeaRBP(disp int, dst Reg) {
	b.rex(true, dst, RBP)
	b.byte(0x8D)
	b.memRBP(dst, disp)
}

func (b *Buf) Rdtsc() { b.byte(0x0F, 0x31) }

const (
	CondS  Cond = 0x8
	CondNS Cond = 0x9
	CondB  Cond = 0x2 // unsigned <
	CondA  Cond = 0x7 // unsigned >
)

/* function frame */

func (b *Buf) Prologue() {
	b.PushR(RBP)
	b.MovRR(true, RSP, RBP)
}

func (b *Buf) Epilogue() {
	b.MovRR(true, RBP, RSP)
	b.PopR(RBP)
	b.Ret()
}
