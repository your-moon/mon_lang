/*
 * mon_lang - arm64 (AArch64) instruction encoder
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// Package armgen assembles a closed AArch64 subset into machine code and lowers
// the architecture-neutral Tacky IR straight onto it. Like the x86 encoder it
// keeps the whole program in one flat buffer with TCC-style label backpatching,
// so the finished code needs no relocations. Values live in stack frame slots
// (a template / stack-machine codegen, correct by construction); x8/x9 are the
// only scratch registers and never span a call.
package armgen

import "encoding/binary"

// register numbers (X0..X30, plus SP/XZR = 31 depending on context)
const (
	x0  = 0
	x1  = 1
	x8  = 8 // scratch A / result
	x9  = 9 // scratch B
	x16 = 16
	x29 = 29
	x30 = 30
	sp  = 31
	xzr = 31
)

// condition codes
const (
	condEQ = 0
	condNE = 1
	condHS = 2
	condLO = 3
	condHI = 8
	condLS = 9
	condGE = 10
	condLT = 11
	condGT = 12
	condLE = 13
)

type fixupKind int

const (
	fixB   fixupKind = iota // b / bl  (imm26)
	fixCond                 // b.cond / cbz / cbnz (imm19)
)

type armFixup struct {
	pos   int // instruction offset in buf
	label string
	kind  fixupKind
}

// Buf accumulates AArch64 machine code and resolves label references.
type Buf struct {
	code   []byte
	labels map[string]int
	fixups []armFixup
	// data references (adrp/add pairs) patched at layout time
	dataFixups []dataFixup
}

type dataFixup struct {
	adrpPos int    // offset of the adrp instruction
	addPos  int    // offset of the add instruction
	label   string // ro/data label
}

func NewBuf() *Buf { return &Buf{labels: map[string]int{}} }

func (b *Buf) Len() int     { return len(b.code) }
func (b *Buf) Code() []byte { return b.code }

func (b *Buf) w(instr uint32) { b.code = binary.LittleEndian.AppendUint32(b.code, instr) }

func (b *Buf) Label(name string) {
	if _, dup := b.labels[name]; dup {
		panic("armgen: duplicate label " + name)
	}
	b.labels[name] = len(b.code)
}

func (b *Buf) HasLabel(name string) bool { _, ok := b.labels[name]; return ok }

/* --- immediate load --- */

// MovImm loads a 64-bit constant into reg via movz + up to three movk.
func (b *Buf) MovImm(reg int, v int64) {
	u := uint64(v)
	// movz reg, #(u & 0xffff)
	b.w(0xD2800000 | (uint32(u&0xffff) << 5) | uint32(reg))
	for shift := uint(1); shift < 4; shift++ {
		part := (u >> (16 * shift)) & 0xffff
		if part != 0 {
			// movk reg, #part, LSL #(16*shift)
			b.w(0xF2800000 | (uint32(shift) << 21) | (uint32(part) << 5) | uint32(reg))
		}
	}
}

/* --- frame loads/stores (scaled unsigned offset from a base) --- */

// LdrFrame: ldr Xt, [Xn, #off]  (off must be a multiple of 8, >= 0).
func (b *Buf) LdrFrame(rt, rn, off int) {
	b.w(0xF9400000 | (uint32(off/8) << 10) | (uint32(rn) << 5) | uint32(rt))
}

// StrFrame: str Xt, [Xn, #off].
func (b *Buf) StrFrame(rt, rn, off int) {
	b.w(0xF9000000 | (uint32(off/8) << 10) | (uint32(rn) << 5) | uint32(rt))
}

// LdrReg / StrReg: ldr/str Xt, [Xn] (zero offset, for pointer load/store).
func (b *Buf) LdrReg(rt, rn int) { b.w(0xF9400000 | (uint32(rn) << 5) | uint32(rt)) }
func (b *Buf) StrReg(rt, rn int) { b.w(0xF9000000 | (uint32(rn) << 5) | uint32(rt)) }

// LdrbReg / StrbReg: byte load (zero-extended) / store, [Xn] zero offset.
func (b *Buf) LdrbReg(rt, rn int) { b.w(0x39400000 | (uint32(rn) << 5) | uint32(rt)) }
func (b *Buf) StrbReg(rt, rn int) { b.w(0x39000000 | (uint32(rn) << 5) | uint32(rt)) }

// Ldr32Reg / Str32Reg: 32-bit load (zero-extends into Xt) / store, [Xn].
func (b *Buf) Ldr32Reg(rt, rn int) { b.w(0xB9400000 | (uint32(rn) << 5) | uint32(rt)) }
func (b *Buf) Str32Reg(rt, rn int) { b.w(0xB9000000 | (uint32(rn) << 5) | uint32(rt)) }

// LdrW / StrW: width-parameterized pointer load/store (4 or 8 bytes).
func (b *Buf) LdrW(rt, rn, width int) {
	if width == 4 {
		b.Ldr32Reg(rt, rn)
	} else {
		b.LdrReg(rt, rn)
	}
}
func (b *Buf) StrW(rt, rn, width int) {
	if width == 4 {
		b.Str32Reg(rt, rn)
	} else {
		b.StrReg(rt, rn)
	}
}

/* --- register moves & ALU (X-form) --- */

func (b *Buf) MovReg(rd, rm int) { b.w(0xAA0003E0 | (uint32(rm) << 16) | uint32(rd)) } // orr rd,xzr,rm

func (b *Buf) Add(rd, rn, rm int) { b.w(0x8B000000 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd)) }
func (b *Buf) Sub(rd, rn, rm int) { b.w(0xCB000000 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd)) }
func (b *Buf) Mul(rd, rn, rm int) { b.w(0x9B007C00 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd)) }
func (b *Buf) SDiv(rd, rn, rm int) {
	b.w(0x9AC00C00 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd))
}
func (b *Buf) UDiv(rd, rn, rm int) {
	b.w(0x9AC00800 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd))
}

// MSub: rd = ra - rn*rm  (used for modulo: a - (a/b)*b).
func (b *Buf) MSub(rd, rn, rm, ra int) {
	b.w(0x9B008000 | (uint32(rm) << 16) | (uint32(ra) << 10) | (uint32(rn) << 5) | uint32(rd))
}

func (b *Buf) Neg(rd, rm int) { b.w(0xCB0003E0 | (uint32(rm) << 16) | uint32(rd)) }  // sub rd,xzr,rm
func (b *Buf) Mvn(rd, rm int) { b.w(0xAA2003E0 | (uint32(rm) << 16) | uint32(rd)) }  // orn rd,xzr,rm
func (b *Buf) And(rd, rn, rm int) { b.w(0x8A000000 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd)) }
func (b *Buf) Orr(rd, rn, rm int) { b.w(0xAA000000 | (uint32(rm) << 16) | (uint32(rn) << 5) | uint32(rd)) }

// LsrImm: rd = rn >> shift (unsigned) via UBFM rd,rn,#shift,#63.
func (b *Buf) LsrImm(rd, rn, shift int) {
	b.w(0xD3400000 | (uint32(shift) << 16) | (63 << 10) | (uint32(rn) << 5) | uint32(rd))
}

// Sxtw: rd(64) = sign-extend rn(32) — SBFM rd,rn,#0,#31.
func (b *Buf) Sxtw(rd, rn int) { b.w(0x93407C00 | (uint32(rn) << 5) | uint32(rd)) }

// Uxtw: rd(64) = zero-extend rn(32) — mov Wd,Wn (32-bit ops clear the top half).
func (b *Buf) Uxtw(rd, rn int) { b.w(0x2A0003E0 | (uint32(rn) << 16) | uint32(rd)) }

// AddImm / SubImm: rd = rn +/- imm12 (imm 0..4095).
func (b *Buf) AddImm(rd, rn, imm int) {
	b.w(0x91000000 | (uint32(imm) << 10) | (uint32(rn) << 5) | uint32(rd))
}
func (b *Buf) SubImm(rd, rn, imm int) {
	b.w(0xD1000000 | (uint32(imm) << 10) | (uint32(rn) << 5) | uint32(rd))
}

// Cmp: subs xzr, rn, rm.
func (b *Buf) Cmp(rn, rm int) { b.w(0xEB00001F | (uint32(rm) << 16) | (uint32(rn) << 5)) }

// CmpImm: subs xzr, rn, #imm12.
func (b *Buf) CmpImm(rn, imm int) { b.w(0xF100001F | (uint32(imm) << 10) | (uint32(rn) << 5)) }

// Cset: rd = (cond) ? 1 : 0  (csinc rd,xzr,xzr,invert(cond)).
func (b *Buf) Cset(rd, cond int) {
	inv := uint32(cond) ^ 1
	b.w(0x9A9F07E0 | (inv << 12) | uint32(rd))
}

/* --- control flow --- */

func (b *Buf) B(label string) {
	b.fixups = append(b.fixups, armFixup{pos: len(b.code), label: label, kind: fixB})
	b.w(0x14000000)
}
func (b *Buf) BL(label string) {
	b.fixups = append(b.fixups, armFixup{pos: len(b.code), label: label, kind: fixB})
	b.w(0x94000000)
}
func (b *Buf) Cbz(rt int, label string) {
	b.fixups = append(b.fixups, armFixup{pos: len(b.code), label: label, kind: fixCond})
	b.w(0xB4000000 | uint32(rt))
}
func (b *Buf) Cbnz(rt int, label string) {
	b.fixups = append(b.fixups, armFixup{pos: len(b.code), label: label, kind: fixCond})
	b.w(0xB5000000 | uint32(rt))
}
func (b *Buf) BCond(cond int, label string) {
	b.fixups = append(b.fixups, armFixup{pos: len(b.code), label: label, kind: fixCond})
	b.w(0x54000000 | uint32(cond))
}

func (b *Buf) Ret() { b.w(0xD65F03C0) }
func (b *Buf) Svc() { b.w(0xD4001001) } // svc #0x80

/* --- prologue / epilogue --- */

// Prologue: stp x29,x30,[sp,#-16]! ; mov x29,sp ; sub sp,sp,#frame.
func (b *Buf) Prologue(frame int) {
	b.w(0xA9BF7BFD)     // stp x29,x30,[sp,#-16]!
	b.w(0x910003FD)     // mov x29,sp
	if frame > 0 {
		b.SubImm(sp, sp, frame)
	}
}

// Epilogue: add sp,sp,#frame ; ldp x29,x30,[sp],#16 ; ret.
func (b *Buf) Epilogue(frame int) {
	if frame > 0 {
		b.AddImm(sp, sp, frame)
	}
	b.w(0xA8C17BFD) // ldp x29,x30,[sp],#16
	b.Ret()
}

/* --- PC-relative data address: adrp Xd,label ; add Xd,Xd,#:lo12: --- */

func (b *Buf) AdrpAdd(rd int, label string) {
	adrpPos := len(b.code)
	b.w(0x90000000 | uint32(rd)) // adrp rd, 0 (patched)
	addPos := len(b.code)
	b.w(0x91000000 | (uint32(rd) << 5) | uint32(rd)) // add rd,rd,#0 (patched)
	b.dataFixups = append(b.dataFixups, dataFixup{adrpPos: adrpPos, addPos: addPos, label: label})
}

/* --- resolution --- */

// Resolve patches branch fixups (internal) and data fixups. dataAddr maps a
// data label to its absolute vm address; codeVM is the vm address of code[0].
func (b *Buf) Resolve(codeVM int, dataAddr map[string]int) error {
	for _, f := range b.fixups {
		target, ok := b.labels[f.label]
		if !ok {
			return errUndef(f.label)
		}
		delta := target - f.pos
		instr := binary.LittleEndian.Uint32(b.code[f.pos:])
		switch f.kind {
		case fixB:
			imm := uint32((delta / 4)) & 0x03FFFFFF
			instr |= imm
		case fixCond:
			imm := uint32((delta / 4)) & 0x7FFFF
			instr |= imm << 5
		}
		binary.LittleEndian.PutUint32(b.code[f.pos:], instr)
	}
	for _, d := range b.dataFixups {
		addr, ok := dataAddr[d.label]
		if !ok {
			return errUndef(d.label)
		}
		pcPage := (codeVM + d.adrpPos) &^ 0xfff
		targetPage := addr &^ 0xfff
		pageDelta := (targetPage - pcPage) >> 12
		adrp := binary.LittleEndian.Uint32(b.code[d.adrpPos:])
		immlo := uint32(pageDelta) & 0x3
		immhi := (uint32(pageDelta) >> 2) & 0x7FFFF
		adrp |= (immlo << 29) | (immhi << 5)
		binary.LittleEndian.PutUint32(b.code[d.adrpPos:], adrp)
		add := binary.LittleEndian.Uint32(b.code[d.addPos:])
		add |= (uint32(addr&0xfff) << 10)
		binary.LittleEndian.PutUint32(b.code[d.addPos:], add)
	}
	return nil
}

type undefErr string

func (e undefErr) Error() string { return "armgen: undefined label " + string(e) }
func errUndef(l string) error    { return undefErr(l) }
