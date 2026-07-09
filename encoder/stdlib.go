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
// skips leading junk like scanf, sanamsargwy_too returns 1..n).
package encoder

// darwin BSD syscall numbers (class 2 << 24).
const (
	sysExit         = 0x2000001
	sysRead         = 0x2000003
	sysWrite        = 0x2000004
	sysSelect       = 0x200005D
	sysGettimeofday = 0x2000074
	sysMmap         = 0x20000C5
	sysOpen         = 0x2000005
	sysClose        = 0x2000006
	sysLseek        = 0x20000C7
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
		{Label: "argc", Size: 8, Init: 0},
		{Label: "argv", Size: 8, Init: 0},
	}
}

// StdlibStrings returns string constants the stdlib references.
func StdlibStrings() map[string]string { // label -> bytes
	return map[string]string{
		"clear_seq": "\x1b[H\x1b[2J",
		"empty_str": "",
	}
}

// EmitEntry writes the process entry stub: capture argc/argv from the
// kernel-provided stack layout ([rsp]=argc, rsp+8=argv), call user main
// (wndsen), pass its return value to exit(2). Must be at code offset 0.
func (b *Buf) EmitEntry(mainLabel string, dataLabel func(string) string) {
	b.LoadBaseR(true, RSP, RAX) // argc
	b.MovRRip(true, RAX, dataLabel("argc"))
	b.MovRR(true, RSP, RAX)
	b.AluIR(OpAdd, true, 8, RAX) // &argv[0]
	b.MovRRip(true, RAX, dataLabel("argv"))
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

	/* ekhevle(n): print n as an unsigned decimal (эхэвлэ) */
	b.Label(fnLabel("ekhevle"))
	b.Prologue()
	b.AluIR(OpSub, true, 48, RSP)
	b.MovRR(true, RDI, RAX)
	b.LeaRBP(-8, RSI)
	b.MovIR(false, 10, R8)
	b.Label("stdlib.ekhevle.loop")
	b.XorRR(true, RDX, RDX)
	b.DivR(true, R8)
	b.AluIR(OpAdd, false, '0', RDX)
	b.DecR(true, RSI)
	b.StoreByte(RDX, RSI)
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondNE, "stdlib.ekhevle.loop")
	b.LeaRBP(-8, RDX)
	b.AluRR(OpSub, true, RSI, RDX)
	b.MovIR(false, 1, RDI)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.Epilogue()

	/* temdegt_khevlekh(cp): UTF-8 encode one codepoint and write it */
	b.Label(fnLabel("temdegt_khevlekh"))
	b.Prologue()
	b.AluIR(OpSub, true, 16, RSP)
	b.MovRR(false, RDI, RAX) // codepoint, zero-extended
	b.LeaRBP(-16, RSI)       // write cursor
	b.AluIR(OpCmp, false, 0x80, RAX)
	b.Jcc(CondB, "stdlib.utf8.b1")
	b.AluIR(OpCmp, false, 0x800, RAX)
	b.Jcc(CondB, "stdlib.utf8.b2")
	b.AluIR(OpCmp, false, 0x10000, RAX)
	b.Jcc(CondB, "stdlib.utf8.b3")
	// 4 bytes: F0|cp>>18, 80|cp>>12&3F, 80|cp>>6&3F, 80|cp&3F
	b.MovRR(false, RAX, RCX)
	b.ShrIR(false, 18, RCX)
	b.OrIR(false, 0xF0, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.MovRR(false, RAX, RCX)
	b.ShrIR(false, 12, RCX)
	b.AndIR(false, 0x3F, RCX)
	b.OrIR(false, 0x80, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.Jmp("stdlib.utf8.tail2")
	b.Label("stdlib.utf8.b3") // E0|cp>>12, then shared 2-byte tail
	b.MovRR(false, RAX, RCX)
	b.ShrIR(false, 12, RCX)
	b.OrIR(false, 0xE0, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.Jmp("stdlib.utf8.tail2")
	b.Label("stdlib.utf8.b2") // C0|cp>>6, then final continuation byte
	b.MovRR(false, RAX, RCX)
	b.ShrIR(false, 6, RCX)
	b.OrIR(false, 0xC0, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.Jmp("stdlib.utf8.tail1")
	b.Label("stdlib.utf8.tail2") // 80|cp>>6&3F then fall into tail1
	b.MovRR(false, RAX, RCX)
	b.ShrIR(false, 6, RCX)
	b.AndIR(false, 0x3F, RCX)
	b.OrIR(false, 0x80, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.Label("stdlib.utf8.tail1") // 80|cp&3F
	b.MovRR(false, RAX, RCX)
	b.AndIR(false, 0x3F, RCX)
	b.OrIR(false, 0x80, RCX)
	b.StoreByte(RCX, RSI)
	b.IncR(true, RSI)
	b.Jmp("stdlib.utf8.write")
	b.Label("stdlib.utf8.b1") // ASCII: the codepoint itself
	b.StoreByte(RAX, RSI)
	b.IncR(true, RSI)
	b.Label("stdlib.utf8.write")
	b.MovRR(true, RSI, RDX) // len = cursor - buf
	b.LeaRBP(-16, RSI)
	b.AluRR(OpSub, true, RSI, RDX)
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

	/* sanamsargwy_too(n): xorshift64, seeded lazily from rdtsc; (x % n) + 1 */
	b.Label(fnLabel("sanamsargwy_too"))
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

	/* monAlloc(n): bump allocator over anonymous mmap chunks. free is a
	   no-op, so memory is reclaimed only at exit - the right tradeoff for
	   a teaching language (TCC uses the same simplification for -run). */
	b.Label(fnLabel("monAlloc"))
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

	/* delgets_tseverlekh(): write ANSI home+clear */
	b.Label(fnLabel("delgets_tseverlekh"))
	b.Prologue()
	b.LeaRip(RSI, strLabel("clear_seq"))
	b.MovIR(false, 1, RDI)
	b.MovIR(false, 7, RDX)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.Epilogue()

	/* mqr_urt(s): strlen in bytes */
	b.Label(fnLabel("mqr_urt"))
	b.Prologue()
	b.MovRR(true, RDI, R10)
	b.XorRR(true, RAX, RAX)
	b.Label("stdlib.strlen.loop")
	b.LoadByte(R10, RCX)
	b.TestRR(false, RCX, RCX)
	b.Jcc(CondE, "stdlib.strlen.done")
	b.IncR(true, R10)
	b.IncR(true, RAX)
	b.Jmp("stdlib.strlen.loop")
	b.Label("stdlib.strlen.done")
	b.Epilogue()

	/* bayt(s, i): unsigned byte at index */
	b.Label(fnLabel("bayt"))
	b.Prologue()
	b.AluRR(OpAdd, true, RSI, RDI)
	b.LoadByte(RDI, RAX)
	b.Epilogue()

	/* bayt_tavikh(s, i, b): store byte at index */
	b.Label(fnLabel("bayt_tavikh"))
	b.Prologue()
	b.AluRR(OpAdd, true, RSI, RDI)
	b.StoreByte(RDX, RDI)
	b.Epilogue()

	/* fayl_unshikh_bwten(path): whole file as a NUL-terminated string.
	   open / lseek-end / lseek-0 / malloc(size+1) / read / close.
	   Errors return "" (the empty string constant). */
	b.Label(fnLabel("fayl_unshikh_bwten"))
	b.Prologue()
	b.AluIR(OpSub, true, 32, RSP)
	b.XorRR(true, RSI, RSI) // O_RDONLY
	b.XorRR(true, RDX, RDX)
	b.MovIR(false, sysOpen, RAX)
	b.Syscall()
	b.Jcc(CondB, "stdlib.fread.fail")
	b.MovRM(true, RAX, -8) // fd
	b.MovRR(true, RAX, RDI)
	b.XorRR(true, RSI, RSI)
	b.MovIR(false, 2, RDX) // SEEK_END
	b.MovIR(false, sysLseek, RAX)
	b.Syscall()
	b.MovRM(true, RAX, -16) // size
	b.MovMR(true, -8, RDI)
	b.XorRR(true, RSI, RSI)
	b.XorRR(true, RDX, RDX) // SEEK_SET
	b.MovIR(false, sysLseek, RAX)
	b.Syscall()
	b.MovMR(true, -16, RDI)
	b.IncR(true, RDI) // +1 for NUL
	b.Call(fnLabel("monAlloc"))
	b.MovRM(true, RAX, -24) // buf
	b.MovMR(true, -8, RDI)
	b.MovRR(true, RAX, RSI)
	b.MovMR(true, -16, RDX)
	b.MovIR(false, sysRead, RAX)
	b.Syscall()
	// NUL-terminate at the actually-read length
	b.MovMR(true, -24, RCX)
	b.AluRR(OpAdd, true, RAX, RCX)
	b.XorRR(true, RDX, RDX)
	b.StoreByte(RDX, RCX)
	b.MovMR(true, -8, RDI)
	b.MovIR(false, sysClose, RAX)
	b.Syscall()
	b.MovMR(true, -24, RAX)
	b.Epilogue()
	b.Label("stdlib.fread.fail")
	b.LeaRip(RAX, strLabel("empty_str"))
	b.Epilogue()

	/* fayl_bichikh(path, content): write string to file, 0 on success */
	b.Label(fnLabel("fayl_bichikh"))
	b.Prologue()
	b.AluIR(OpSub, true, 32, RSP)
	b.MovRM(true, RSI, -16) // content
	b.MovIR(false, 0x601, RSI) // O_WRONLY|O_CREAT|O_TRUNC
	b.MovIR(false, 0o644, RDX)
	b.MovIR(false, sysOpen, RAX)
	b.Syscall()
	b.Jcc(CondB, "stdlib.fwrite.fail")
	b.MovRM(true, RAX, -8) // fd
	b.MovMR(true, -16, RDI)
	b.Call(fnLabel("mqr_urt"))
	b.MovRR(true, RAX, RDX) // len
	b.MovMR(true, -8, RDI)
	b.MovMR(true, -16, RSI)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.MovMR(true, -8, RDI)
	b.MovIR(false, sysClose, RAX)
	b.Syscall()
	b.XorRR(true, RAX, RAX)
	b.Epilogue()
	b.Label("stdlib.fwrite.fail")
	b.MovIR(false, 1, RAX)
	b.Epilogue()

	/* fayl_bichikh_bayt(path, buf, len): write len raw bytes (NULs included) */
	b.Label(fnLabel("fayl_bichikh_bayt"))
	b.Prologue()
	b.AluIR(OpSub, true, 32, RSP)
	b.MovRM(true, RSI, -16) // buf
	b.MovRM(true, RDX, -24) // len
	b.MovIR(false, 0x601, RSI)
	b.MovIR(false, 0o644, RDX)
	b.MovIR(false, sysOpen, RAX)
	b.Syscall()
	b.Jcc(CondB, "stdlib.fwriteb.fail")
	b.MovRM(true, RAX, -8) // fd
	b.MovMR(true, -8, RDI)
	b.MovMR(true, -16, RSI)
	b.MovMR(true, -24, RDX)
	b.MovIR(false, sysWrite, RAX)
	b.Syscall()
	b.MovMR(true, -8, RDI)
	b.MovIR(false, sysClose, RAX)
	b.Syscall()
	b.XorRR(true, RAX, RAX)
	b.Epilogue()
	b.Label("stdlib.fwriteb.fail")
	b.MovIR(false, 1, RAX)
	b.Epilogue()

	/* butarkhay_khevlekh(d in xmm0): fixed decimal, 6 truncated fractional
	   digits, byte-identical to the C shim. Uses khevle for the integer
	   part and temdegt_khevlekh for '.' and each digit; d and the loop
	   state live on the stack across those calls (they clobber xmm/regs).
	   Constants (0.0/-1.0/10.0/(double)n) are made with cvtsi2sd, so no
	   double literal pool is needed here. */
	b.Label(fnLabel("butarkhay_khevlekh"))
	b.Prologue()
	b.AluIR(OpSub, true, 32, RSP)
	b.MovsdRM(0, -8) // save d at [rbp-8]

	// sign: if d < 0, print '-' and negate
	b.MovsdMR(-8, 0)
	b.XorRR(true, R10, R10)
	b.Cvtsi2sdRR(1, R10) // xmm1 = 0.0
	b.ComisdRR(0, 1)     // d - 0.0
	b.Jcc(CondAE, "stdlib.dbl.pos")
	b.MovIR(false, '-', RDI)
	b.Call(fnLabel("temdegt_khevlekh"))
	b.MovsdMR(-8, 0)
	b.MovIR(true, -1, R10)
	b.Cvtsi2sdRR(1, R10)      // xmm1 = -1.0
	b.sseArithRR(sseMul, 0, 1) // d *= -1
	b.MovsdRM(0, -8)
	b.Label("stdlib.dbl.pos")

	// integer part
	b.MovsdMR(-8, 0)
	b.Cvttsd2siRR(RDI, 0) // ip = (long)d
	b.MovRM(true, RDI, -16)
	b.Call(fnLabel("khevle"))

	// '.'
	b.MovIR(false, '.', RDI)
	b.Call(fnLabel("temdegt_khevlekh"))

	// frac = d - (double)ip
	b.MovsdMR(-8, 0)
	b.MovMR(true, -16, R10)
	b.Cvtsi2sdRR(1, R10)
	b.sseArithRR(sseSub, 0, 1)
	b.MovsdRM(0, -8) // frac at [rbp-8]

	// six fractional digits
	b.MovIM(true, 6, -24) // counter
	b.Label("stdlib.dbl.loop")
	b.MovsdMR(-8, 0)
	b.MovIR(true, 10, R10)
	b.Cvtsi2sdRR(1, R10)
	b.sseArithRR(sseMul, 0, 1) // frac *= 10
	b.Cvttsd2siRR(R11, 0)      // dg = (int)frac
	b.MovRM(true, R11, -32)
	b.MovMR(true, -32, R10)
	b.Cvtsi2sdRR(1, R10)
	b.sseArithRR(sseSub, 0, 1) // frac -= dg
	b.MovsdRM(0, -8)
	b.MovMR(true, -32, RDI)
	b.AluIR(OpAdd, false, '0', RDI)
	b.Call(fnLabel("temdegt_khevlekh"))
	// counter--
	b.MovMR(true, -24, RAX)
	b.DecR(true, RAX)
	b.MovRM(true, RAX, -24)
	b.TestRR(true, RAX, RAX)
	b.Jcc(CondNE, "stdlib.dbl.loop")
	b.Epilogue()

	/* mqr_shine(len): mutable string buffer; bump-allocated mmap pages
	   arrive zeroed, so the buffer is born NUL-terminated everywhere */
	b.Label(fnLabel("mqr_shine"))
	b.Prologue()
	b.IncR(true, RDI) // room for NUL
	b.Call(fnLabel("monAlloc"))
	b.Epilogue()

	/* argumyent_too(): argc captured by the entry stub */
	b.Label(fnLabel("argumyent_too"))
	b.Prologue()
	b.MovRipR(true, dataLabel("argc"), RAX)
	b.Epilogue()

	/* argumyent(i): argv[i], "" when out of range */
	b.Label(fnLabel("argumyent"))
	b.Prologue()
	b.MovsxdRR(RDI, RCX)
	b.TestRR(true, RCX, RCX)
	b.Jcc(CondS, "stdlib.argv.bad")
	b.MovRipR(true, dataLabel("argc"), RAX)
	b.AluRR(OpCmp, true, RAX, RCX) // cmp rax, rcx: sets flags for rcx-rax
	b.Jcc(CondGE, "stdlib.argv.bad")
	b.MovRipR(true, dataLabel("argv"), RAX)
	b.ShlIR(true, 3, RCX)
	b.AluRR(OpAdd, true, RCX, RAX)
	b.LoadBaseR(true, RAX, RAX)
	b.Epilogue()
	b.Label("stdlib.argv.bad")
	b.LeaRip(RAX, strLabel("empty_str"))
	b.Epilogue()
}

// StdlibFns lists the function names EmitStdlib defines; the program mapper
// uses this to know which extern declarations are satisfied internally.
func StdlibFns() []string {
	return []string{"khevle", "ekhevle", "temdegt_khevlekh", "mqr_khevlekh", "unsh", "unsh32",
		"sanamsargwy_too", "odoo", "monAlloc", "chqlqqlqkh", "khwleekh",
		"delgets_tseverlekh", "mqr_urt", "bayt", "bayt_tavikh",
		"fayl_unshikh_bwten", "fayl_bichikh", "fayl_bichikh_bayt", "argumyent_too", "argumyent",
		"mqr_shine", "butarkhay_khevlekh"}
}
