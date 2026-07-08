/*
 * mon_lang - built-in standard library (raw darwin syscalls)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// The stdlib is hand-assembled with the encoder and appended to every
// program, replacing stdlib/lib.c and its cc/libc dependency. Functions
// follow the SysV ABI so compiler-generated call sites need no changes.
// Semantics mirror lib.c exactly (khevle prints without newline, unsh
// skips leading junk like scanf, sanamsargwyToo returns 1..n).
package encoder

// darwin BSD syscall numbers (class 2 << 24).
const (
	sysExit         = 0x2000001
	sysRead         = 0x2000003
	sysWrite        = 0x2000004
	sysSelect       = 0x200005D
	sysGettimeofday = 0x2000074
	sysMmap         = 0x20000C5
)

const heapChunk = 1 << 20 // mmap granularity for the bump allocator

// StdlibDataLabels lists data labels the stdlib expects the layout stage to
// provide: name -> (size, initial value).
type StdlibData struct {
	Label string
	Size  int
	Init  int64
}

func StdlibDataDefs() []StdlibData {
	return []StdlibData{
		{Label: "rand_state", Size: 8, Init: 0},
		{Label: "heap_ptr", Size: 8, Init: 0},
		{Label: "heap_end", Size: 8, Init: 0},
	}
}

// StdlibStrings returns string constants the stdlib references.
func StdlibStrings() map[string]string { // label -> bytes
	return map[string]string{"clear_seq": "\x1b[H\x1b[2J"}
}

// EmitEntry writes the process entry stub: call user main (wndsen), pass its
// return value to exit(2). Must be emitted at code offset 0.
func (b *Buf) EmitEntry(mainLabel string) {
	b.Call(mainLabel)
	b.MovRR(false, RAX, RDI) // exit code = wndsen() result (low 32 bits)
	b.MovIR(false, sysExit, RAX)
	b.Syscall()
}

// EmitStdlib appends every stdlib function. fnLabel/lblLabel namespace the
// symbols consistently with user code.
func (b *Buf) EmitStdlib(fnLabel func(string) string, dataLabel func(string) string, strLabel func(string) string) {
	/* khevle(n): print n in decimal, no newline.
	   Digits are built backwards in a 32-byte stack buffer. neg on
	   INT64_MIN wraps to the same bit pattern, which unsigned div then
	   prints correctly, so no special case. */
	b.Label(fnLabel("khevle"))
	b.Prologue()
	b.AluIR(OpSub, true, 48, RSP)
	b.MovRR(true, RDI, RAX)
	b.XorRR(true, R9, R9) // sign flag
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondNS, "stdlib.khevle.pos")
	b.MovIR(false, 1, R9)
	b.NegR(true, RAX)
	b.Label("stdlib.khevle.pos")
	b.LeaRBP(-8, RSI) // end of buffer; digits grow downwards
	b.MovIR(false, 10, R8)
	b.Label("stdlib.khevle.loop")
	b.XorRR(true, RDX, RDX)
	b.DivR(true, R8) // rax = quot, rdx = digit
	b.AluIR(OpAdd, false, '0', RDX)
	b.DecR(true, RSI)
	b.StoreByte(RDX, RSI)
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondNE, "stdlib.khevle.loop")
	b.TestRR(true, R9, R9)
	b.Jcc(CondE, "stdlib.khevle.nosign")
	b.MovIR(false, '-', R8)
	b.DecR(true, RSI)
	b.StoreByte(R8, RSI)
	b.Label("stdlib.khevle.nosign")
	b.LeaRBP(-8, RDX)
	b.AluRR(OpSub, true, RSI, RDX) // len = bufend - ptr
	b.MovIR(false, 1, RDI)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.Epilogue()

	/* mqr_khevlekh(s): write(1, s, strlen(s)) */
	b.Label(fnLabel("mqr_khevlekh"))
	b.Prologue()
	b.MovRR(true, RDI, R10)
	b.XorRR(true, RDX, RDX)
	b.Label("stdlib.puts.len")
	b.LoadByte(R10, RAX)
	b.TestRR(false, RAX, RAX)
	b.Jcc(CondE, "stdlib.puts.write")
	b.IncR(true, R10)
	b.IncR(true, RDX)
	b.Jmp("stdlib.puts.len")
	b.Label("stdlib.puts.write")
	b.MovRR(true, RDI, RSI)
	b.MovIR(false, 1, RDI)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.Epilogue()

	/* unsh(): parse a signed decimal from stdin, scanf-style: skip
	   non-digit junk, then optional '-', then digits until non-digit.
	   Reads one byte per syscall - fine for interactive input. */
	b.Label(fnLabel("unsh"))
	b.Prologue()
	b.AluIR(OpSub, true, 16, RSP)
	b.XorRR(true, R8, R8)   // accumulator
	b.XorRR(true, R9, R9)   // sign
	b.XorRR(true, R11, R11) // started flag
	b.Label("stdlib.unsh.read")
	b.XorRR(true, RDI, RDI)
	b.LeaRBP(-16, RSI)
	b.MovIR(false, 1, RDX)
	b.MovIR(false, sysRead, RAX)
	b.Syscall()
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondLE, "stdlib.unsh.done") // EOF or error
	b.LeaRBP(-16, R10)
	b.LoadByte(R10, RCX)
	b.TestRR(true, R11, R11)
	b.Jcc(CondNE, "stdlib.unsh.digit") // already in number: any non-digit ends
	b.AluIR(OpCmp, false, '-', RCX)
	b.Jcc(CondNE, "stdlib.unsh.digit0")
	b.MovIR(false, 1, R9)
	b.MovIR(false, 1, R11)
	b.Jmp("stdlib.unsh.read")
	b.Label("stdlib.unsh.digit0")
	b.AluIR(OpCmp, false, '0', RCX)
	b.Jcc(CondL, "stdlib.unsh.read") // leading junk: skip
	b.AluIR(OpCmp, false, '9', RCX)
	b.Jcc(CondG, "stdlib.unsh.read")
	b.MovIR(false, 1, R11)
	b.Jmp("stdlib.unsh.acc")
	b.Label("stdlib.unsh.digit")
	b.AluIR(OpCmp, false, '0', RCX)
	b.Jcc(CondL, "stdlib.unsh.done")
	b.AluIR(OpCmp, false, '9', RCX)
	b.Jcc(CondG, "stdlib.unsh.done")
	b.Label("stdlib.unsh.acc")
	b.ImulIR(true, 10, R8)
	b.AluIR(OpSub, false, '0', RCX)
	b.AluRR(OpAdd, true, RCX, R8)
	b.Jmp("stdlib.unsh.read")
	b.Label("stdlib.unsh.done")
	b.MovRR(true, R8, RAX)
	b.TestRR(true, R9, R9)
	b.Jcc(CondE, "stdlib.unsh.ret")
	b.NegR(true, RAX)
	b.Label("stdlib.unsh.ret")
	b.Epilogue()

	/* unsh32(): same parse, caller reads eax */
	b.Label(fnLabel("unsh32"))
	b.Prologue()
	b.Call(fnLabel("unsh"))
	b.Epilogue()

	/* sanamsargwyToo(n): xorshift64, seeded lazily from rdtsc; (x % n) + 1 */
	b.Label(fnLabel("sanamsargwyToo"))
	b.Prologue()
	b.MovRipR(true, dataLabel("rand_state"), RAX)
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondNE, "stdlib.rand.have")
	b.Rdtsc() // edx:eax = cycle counter
	b.ShlIR(true, 32, RDX)
	b.AluRR(OpAdd, true, RDX, RAX)
	b.OrIR(true, 1, RAX) // never zero
	b.Label("stdlib.rand.have")
	b.MovRR(true, RAX, RCX) // x ^= x << 13
	b.ShlIR(true, 13, RCX)
	b.XorRR(true, RCX, RAX)
	b.MovRR(true, RAX, RCX) // x ^= x >> 7
	b.ShrIR(true, 7, RCX)
	b.XorRR(true, RCX, RAX)
	b.MovRR(true, RAX, RCX) // x ^= x << 17
	b.ShlIR(true, 17, RCX)
	b.XorRR(true, RCX, RAX)
	b.MovRRip(true, RAX, dataLabel("rand_state"))
	b.ShrIR(true, 1, RAX) // keep quotient in range: dividend < 2^63
	b.MovsxdRR(RDI, RCX)  // n
	b.XorRR(true, RDX, RDX)
	b.DivR(true, RCX) // rdx = x % n  (n == 0 traps, same as C)
	b.MovRR(true, RDX, RAX)
	b.IncR(false, RAX)
	b.Epilogue()

	/* odoo(): seconds since epoch via gettimeofday(&tv, NULL) */
	b.Label(fnLabel("odoo"))
	b.Prologue()
	b.AluIR(OpSub, true, 16, RSP)
	b.MovIM(true, 0, -16) // tv.tv_sec = 0 in case the kernel returns in regs
	b.LeaRBP(-16, RDI)
	b.XorRR(true, RSI, RSI)
	b.MovIR(false, sysGettimeofday, RAX)
	b.Syscall()
	b.MovMR(true, -16, RCX) // xnu stores to the buffer; older ABIs return in rax
	b.TestRR(true, RCX, RCX)
	b.Jcc(CondE, "stdlib.odoo.ret")
	b.MovRR(true, RCX, RAX)
	b.Label("stdlib.odoo.ret")
	b.Epilogue()

	/* malloc(n): bump allocator over anonymous mmap chunks. free is a
	   no-op, so memory is reclaimed only at exit - the right tradeoff for
	   a teaching language (TCC uses the same simplification for -run). */
	b.Label(fnLabel("malloc"))
	b.Prologue()
	b.AluIR(OpAdd, true, 15, RDI) // round request to 16
	b.AndIR(true, ^int64(15), RDI)
	b.MovRipR(true, dataLabel("heap_ptr"), RAX)
	b.MovRR(true, RAX, RCX)
	b.AluRR(OpAdd, true, RDI, RCX) // rcx = new ptr
	b.AluRipR(OpCmp, true, dataLabel("heap_end"), RCX)
	b.Jcc(CondLE, "stdlib.malloc.fit") // new ptr <= heap_end: chunk has room
	// grow: mmap(0, max(heapChunk, n), RW, ANON|PRIVATE, -1, 0)
	b.MovRR(true, RDI, RSI)
	b.AluIR(OpCmp, true, heapChunk, RSI)
	b.Jcc(CondGE, "stdlib.malloc.big")
	b.MovIR(true, heapChunk, RSI)
	b.Label("stdlib.malloc.big")
	b.PushR(RDI) // syscall clobbers rcx/r11; n and chunk len live on the stack
	b.PushR(RSI) // keeps rsp 16-aligned and preserves chunk len
	b.XorRR(true, RDI, RDI)
	b.MovIR(true, 3, RDX)      // PROT_READ|PROT_WRITE
	b.MovIR(true, 0x1002, R10) // MAP_ANON|MAP_PRIVATE
	b.MovIR(true, -1, R8)
	b.XorRR(true, R9, R9)
	b.MovIR(false, sysMmap, RAX)
	b.Syscall()
	b.PopR(RSI)
	b.PopR(RDI)
	b.Jcc(CondB, "stdlib.malloc.fail") // carry set on darwin syscall error
	b.MovRR(true, RAX, RCX)
	b.AluRR(OpAdd, true, RSI, RCX)
	b.MovRRip(true, RCX, dataLabel("heap_end"))
	b.MovRR(true, RAX, RCX)
	b.AluRR(OpAdd, true, RDI, RCX) // rcx = new ptr within fresh chunk
	b.Label("stdlib.malloc.fit")
	// both paths: rcx = bump target, result = rcx - n
	b.MovRR(true, RCX, RDX)
	b.AluRR(OpSub, true, RDI, RDX) // result = new ptr - n
	b.MovRR(true, RDX, RAX)
	b.MovRRip(true, RCX, dataLabel("heap_ptr"))
	b.Epilogue()
	b.Label("stdlib.malloc.fail")
	b.MovIR(false, 71, RDI) // out of memory: exit(71)
	b.MovIR(false, sysExit, RAX)
	b.Syscall()

	/* chqlqqlqkh(p): free - bump allocator reclaims at exit, no-op */
	b.Label(fnLabel("chqlqqlqkh"))
	b.Prologue()
	b.Epilogue()

	/* khwleekh(ms): sleep via select(0,0,0,0,&tv) */
	b.Label(fnLabel("khwleekh"))
	b.Prologue()
	b.AluIR(OpSub, true, 16, RSP)
	b.MovsxdRR(RDI, RAX)
	b.MovIR(false, 1000, RCX)
	b.XorRR(true, RDX, RDX)
	b.DivR(true, RCX) // rax = sec, rdx = ms remainder
	b.MovRM(true, RAX, -16)
	b.ImulIR(true, 1000, RDX) // usec
	b.MovRM(true, RDX, -8)
	b.XorRR(true, RDI, RDI)
	b.XorRR(true, RSI, RSI)
	b.XorRR(true, RDX, RDX)
	b.XorRR(true, R10, R10)
	b.LeaRBP(-16, R8)
	b.MovIR(false, sysSelect, RAX)
	b.Syscall()
	b.Epilogue()

	/* delgetsTseverlekh(): write ANSI home+clear */
	b.Label(fnLabel("delgetsTseverlekh"))
	b.Prologue()
	b.LeaRip(RSI, strLabel("clear_seq"))
	b.MovIR(false, 1, RDI)
	b.MovIR(false, 7, RDX)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.Epilogue()
}

// StdlibFns lists the function names EmitStdlib defines; the program mapper
// uses this to know which extern declarations are satisfied internally.
func StdlibFns() []string {
	return []string{"khevle", "mqr_khevlekh", "unsh", "unsh32",
		"sanamsargwyToo", "odoo", "malloc", "chqlqqlqkh", "khwleekh",
		"delgetsTseverlekh"}
}
