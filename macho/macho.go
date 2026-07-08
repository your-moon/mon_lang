/*
 * mon_lang - Mach-O executable writer
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// Package macho writes static x86_64 Mach-O executables directly, replacing
// the external as/cc toolchain (TCC's tccmacho.c model). The output uses an
// LC_UNIXTHREAD entry point and raw syscalls: no dyld, no imports, and no
// code signature - unsigned x86_64 binaries run under Rosetta on current
// macOS, and Linux/arm64 writers can slot in beside this one later.
//
// Layout:
//
//	file 0x0000  __TEXT   header + load commands (page 0)
//	file 0x1000           machine code, then read-only strings
//	file  ...    __DATA   globals (page-aligned)
package macho

import (
	"encoding/binary"
	"os"
)

const (
	page         = 0x1000
	base         = 0x100000000 // image base; __PAGEZERO spans [0, base)
	magic64      = 0xfeedfacf
	cpuAMD64     = 0x01000007
	cpuSubLE     = 0x80000003 // CPU_SUBTYPE_X86_64_ALL | LIB64
	mhExec       = 2
	mhNoUndefs   = 1
	lcSegment64  = 0x19
	lcUnixthread = 0x5
	protRX       = 5
	protRW       = 3
)

type buf struct{ b []byte }

func (w *buf) u32(v uint32)   { w.b = binary.LittleEndian.AppendUint32(w.b, v) }
func (w *buf) u64(v uint64)   { w.b = binary.LittleEndian.AppendUint64(w.b, v) }
func (w *buf) name(s string)  { n := make([]byte, 16); copy(n, s); w.b = append(w.b, n...) }
func (w *buf) bytes(p []byte) { w.b = append(w.b, p...) }
func (w *buf) pad(align int) {
	for len(w.b)%align != 0 {
		w.b = append(w.b, 0)
	}
}

func alignUp(v, a int) int { return (v + a - 1) &^ (a - 1) }

// Layout describes where the pieces land; the encoder needs the deltas to
// patch rip-relative references before WriteExecutable is called.
type Layout struct {
	TextSize int // code+rodata bytes at file offset page
	DataOff  int // __DATA file offset (page-aligned)
	DataVM   int // __DATA address delta from code start
}

// Plan computes the layout for given blob sizes.
func Plan(codeLen, roLen, dataLen int) Layout {
	text := codeLen + roLen
	dataOff := alignUp(page+text, page)
	return Layout{
		TextSize: text,
		DataOff:  dataOff,
		// code starts at vm base+page and file page; vm and file deltas
		// coincide because both are page-aligned from the same origin.
		DataVM: dataOff - page,
	}
}

// WriteExecutable writes the finished binary. code+ro map r-x, data maps rw.
// entry is the offset of the entry stub within code (normally 0).
func WriteExecutable(path string, code, ro, data []byte, entry int) error {
	lay := Plan(len(code), len(ro), len(data))

	seg := func(w *buf, name string, vmaddr, vmsize, off, fsize uint64, prot uint32, sects int, sb []byte) {
		w.u32(lcSegment64)
		w.u32(uint32(72 + len(sb)))
		w.name(name)
		w.u64(vmaddr)
		w.u64(vmsize)
		w.u64(off)
		w.u64(fsize)
		w.u32(prot) // maxprot
		w.u32(prot) // initprot
		w.u32(uint32(sects))
		w.u32(0)
		w.bytes(sb)
	}
	sect := func(sname, segname string, addr, size, off uint64, flags uint32) []byte {
		s := &buf{}
		s.name(sname)
		s.name(segname)
		s.u64(addr)
		s.u64(size)
		s.u32(uint32(off))
		s.u32(4) // 2^4 alignment
		s.u32(0)
		s.u32(0)
		s.u32(flags)
		s.u32(0)
		s.u32(0)
		s.u32(0)
		return s.b
	}

	codeAddr := uint64(base + page)
	textVM := uint64(alignUp(page+lay.TextSize, page))
	dataVM := uint64(base) + uint64(lay.DataOff)

	cmds := &buf{}
	seg(cmds, "__PAGEZERO", 0, base, 0, 0, 0, 0, nil)

	textSects := sect("__text", "__TEXT", codeAddr, uint64(len(code)), page, 0x80000400)
	if len(ro) > 0 {
		textSects = append(textSects, sect("__cstring", "__TEXT",
			codeAddr+uint64(len(code)), uint64(len(ro)), uint64(page+len(code)), 0x2)...)
	}
	nText := 1
	if len(ro) > 0 {
		nText = 2
	}
	seg(cmds, "__TEXT", base, textVM, 0, uint64(page+lay.TextSize), protRX, nText, textSects)

	if len(data) > 0 {
		dSects := sect("__data", "__DATA", dataVM, uint64(len(data)), uint64(lay.DataOff), 0)
		seg(cmds, "__DATA", dataVM, uint64(alignUp(len(data), page)),
			uint64(lay.DataOff), uint64(len(data)), protRW, 1, dSects)
	}

	// LC_UNIXTHREAD: x86_THREAD_STATE64 (flavor 4, 21 quads), rip at index 16.
	th := &buf{}
	th.u32(lcUnixthread)
	th.u32(uint32(16 + 8*21))
	th.u32(4)
	th.u32(42)
	regs := make([]uint64, 21)
	regs[16] = codeAddr + uint64(entry)
	for _, r := range regs {
		th.u64(r)
	}
	cmds.bytes(th.b)

	ncmds := 3
	if len(data) > 0 {
		ncmds = 4
	}

	out := &buf{}
	out.u32(magic64)
	out.u32(cpuAMD64)
	out.u32(cpuSubLE)
	out.u32(mhExec)
	out.u32(uint32(ncmds))
	out.u32(uint32(len(cmds.b)))
	out.u32(mhNoUndefs)
	out.u32(0)
	out.bytes(cmds.b)
	if len(out.b) > page {
		panic("macho: load commands exceed header page")
	}
	out.pad(page)
	out.bytes(code)
	out.bytes(ro)
	if len(data) > 0 {
		out.pad(page)
		out.bytes(data)
	}

	if err := os.WriteFile(path, out.b, 0755); err != nil {
		return err
	}
	return nil
}
