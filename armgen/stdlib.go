/*
 * mon_lang - arm64 built-in stdlib (raw macOS syscalls)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package armgen

// macOS BSD syscall numbers (arm64 passes the raw number in x16, without the
// x86 0x2000000 class prefix).
const (
	sysExit  = 1
	sysRead  = 3
	sysWrite = 4
	sysOpen  = 5
	sysClose = 6
	sysMmap  = 197
)

const heapChunk = 1 << 20 // mmap granularity for the bump allocator

// heap state lives in reserved __DATA slots (see Compile).
const (
	heapPtrLabel = "std.heap_ptr"
	heapEndLabel = "std.heap_end"
	argcLabel    = "std.argc"
	argvLabel    = "std.argv"
)

const (
	sysLseek = 199
)

// emitStdlib appends the arm64 built-in functions used by mon programs. All I/O
// is via raw `svc #0x80` syscalls. Heap-backed builtins (monAlloc etc.) live in
// heap.go; this file is the syscall + string/format core.
func (g *gen) emitStdlib() {
	g.stdKhevle()   // хэвлэ  — signed decimal, no newline
	g.stdEkhevle()  // эхэвлэ — unsigned decimal
	g.stdMqrUrt()   // мөр_урт — strlen
	g.stdBayt()     // байт / байт_тавих — indexed byte load/store
	g.stdMqrPrint() // мөр_хэвлэх — write a NUL-terminated string
	g.stdTemdegt()  // тэмдэгт_хэвлэх — UTF-8 encode one codepoint
	g.stdClear()    // дэлгэцЦэвэрлэх — ANSI home+clear
	g.stdMonAlloc() // monAlloc — bump allocator over mmap'd chunks
	g.stdMqrShine() // мөр_шинэ — allocate a zeroed (NUL-terminated) buffer
	g.stdFaylBichikh() // файл_бичих — write a string to a file
	g.stdFaylUnshikh() // файл_унших_бүтэн — read a whole file into a string
	g.stdUnsh()        // унш — parse a signed decimal from stdin
	g.stdArgs()        // аргумент_тоо / аргумент — argc / argv[i]
	g.stdButarkhay()   // бутархай_хэвлэх — print a double, 6 fixed decimals
	g.stdNoops()       // чөлөөлөх (free), хүлээх (sleep) — no-ops
}

// stdButarkhay: бутархай_хэвлэх(d) prints a double with 6 truncated fractional
// digits (printf %f). The bit pattern arrives in x0. The integer part goes
// through хэвлэ, '.' and each digit through тэмдэгт_хэвлэх; d and the loop state
// live on the stack across those calls (which clobber the FP scratch).
func (g *gen) stdButarkhay() {
	b := g.b
	b.Label(fnLabel("бутархай_хэвлэх"))
	b.Prologue(32)
	b.StrFrame(x0, sp, 0) // save bit pattern

	// sign: if d < 0, print '-' and negate
	b.FmovXtoD(0, x0)
	b.MovImm(9, 0)
	b.FmovXtoD(1, 9) // d1 = 0.0
	b.Fcmp(0, 1)
	b.BCond(condGE, "std.dbl.pos")
	b.MovImm(x0, '-')
	b.BL(fnLabel("тэмдэгт_хэвлэх"))
	b.LdrFrame(x0, sp, 0)
	b.FmovXtoD(0, x0)
	b.Fneg(0, 0)
	b.FmovDtoX(x0, 0)
	b.StrFrame(x0, sp, 0)
	b.Label("std.dbl.pos")

	// integer part -> хэвлэ
	b.LdrFrame(x0, sp, 0)
	b.FmovXtoD(0, x0)
	b.Fcvtzs(x0, 0)
	b.StrFrame(x0, sp, 8) // ip
	b.BL(fnLabel("хэвлэ"))

	// '.'
	b.MovImm(x0, '.')
	b.BL(fnLabel("тэмдэгт_хэвлэх"))

	// frac = d - (double)ip
	b.LdrFrame(x0, sp, 0)
	b.FmovXtoD(0, x0)
	b.LdrFrame(x0, sp, 8)
	b.Scvtf(1, x0)
	b.Fsub(0, 0, 1)
	b.FmovDtoX(x0, 0)
	b.StrFrame(x0, sp, 0) // frac
	b.MovImm(9, 6)
	b.StrFrame(9, sp, 16) // counter

	b.Label("std.dbl.loop")
	b.LdrFrame(x0, sp, 0)
	b.FmovXtoD(0, x0)
	b.MovImm(9, 10)
	b.Scvtf(1, 9)
	b.Fmul(0, 0, 1) // frac *= 10
	b.Fcvtzs(9, 0)  // dg = (int)frac
	b.StrFrame(9, sp, 24)
	b.Scvtf(1, 9)
	b.Fsub(0, 0, 1) // frac -= dg
	b.FmovDtoX(x0, 0)
	b.StrFrame(x0, sp, 0)
	b.LdrFrame(x0, sp, 24)
	b.AddImm(x0, x0, '0')
	b.BL(fnLabel("тэмдэгт_хэвлэх"))
	b.LdrFrame(9, sp, 16)
	b.SubImm(9, 9, 1)
	b.StrFrame(9, sp, 16)
	b.Cbnz(9, "std.dbl.loop")

	b.Epilogue(32)
}

// stdFaylUnshikh: файл_унших_бүтэн(зам) returns the whole file as a
// NUL-terminated string (open / lseek-end / lseek-0 / monAlloc / read / close).
func (g *gen) stdFaylUnshikh() {
	b := g.b
	b.Label(fnLabel("файл_унших_бүтэн"))
	b.Prologue(32)
	b.MovImm(1, 0) // O_RDONLY
	b.MovImm(2, 0)
	b.MovImm(x16, sysOpen)
	b.Svc() // x0 = fd
	b.StrFrame(x0, sp, 0)
	b.MovImm(1, 0)
	b.MovImm(2, 2) // SEEK_END
	b.MovImm(x16, sysLseek)
	b.Svc() // x0 = size
	b.StrFrame(x0, sp, 8)
	b.LdrFrame(x0, sp, 0) // fd
	b.MovImm(1, 0)
	b.MovImm(2, 0) // SEEK_SET
	b.MovImm(x16, sysLseek)
	b.Svc()
	b.LdrFrame(x0, sp, 8) // size
	b.AddImm(x0, x0, 1)
	b.BL(fnLabel("monAlloc"))
	b.StrFrame(x0, sp, 16) // buf
	b.LdrFrame(x0, sp, 0)  // fd
	b.LdrFrame(1, sp, 16)  // buf
	b.LdrFrame(2, sp, 8)   // size
	b.MovImm(x16, sysRead)
	b.Svc() // x0 = bytes read
	b.LdrFrame(9, sp, 16)
	b.Add(9, 9, x0) // buf + read
	b.MovImm(10, 0)
	b.StrbReg(10, 9) // NUL-terminate
	b.LdrFrame(x0, sp, 0)
	b.MovImm(x16, sysClose)
	b.Svc()
	b.LdrFrame(x0, sp, 16) // return buf
	b.Epilogue(32)
}

// stdArgs: аргумент_тоо() returns argc; аргумент(и) returns argv[и] (a C
// string pointer). Both read the slots captured by the entry stub.
func (g *gen) stdArgs() {
	b := g.b
	b.Label(fnLabel("аргумент_тоо"))
	b.Prologue(0)
	b.AdrpAdd(9, argcLabel)
	b.LdrReg(x0, 9)
	b.Epilogue(0)

	b.Label(fnLabel("аргумент"))
	b.Prologue(0)
	b.AdrpAdd(9, argvLabel)
	b.LdrReg(9, 9)  // argv base
	b.MovImm(10, 8) // 8-byte pointers
	b.Mul(0, x0, 10)
	b.Add(9, 9, x0) // &argv[и]
	b.LdrReg(x0, 9) // argv[и]
	b.Epilogue(0)
}

// stdFaylBichikh: файл_бичих(зам, агуулга) writes агуулга to file зам
// (O_WRONLY|O_CREAT|O_TRUNC, 0644), returning 0.
func (g *gen) stdFaylBichikh() {
	b := g.b
	b.Label(fnLabel("файл_бичих"))
	b.Prologue(16)
	b.StrFrame(1, sp, 0) // save агуулга (content)
	b.MovImm(1, 0x601)   // O_WRONLY|O_CREAT|O_TRUNC
	b.MovImm(2, 0o644)
	b.MovImm(x16, sysOpen)
	b.Svc() // x0 = fd
	b.StrFrame(x0, sp, 8)
	b.LdrFrame(x0, sp, 0) // content
	b.BL(fnLabel("мөр_урт"))
	b.MovReg(2, x0)       // len
	b.LdrFrame(x0, sp, 8) // fd
	b.LdrFrame(1, sp, 0)  // content
	b.MovImm(x16, sysWrite)
	b.Svc()
	b.LdrFrame(x0, sp, 8) // fd
	b.MovImm(x16, sysClose)
	b.Svc()
	b.MovImm(x0, 0)
	b.Epilogue(16)
}

// stdUnsh: унш() reads bytes from stdin, skips leading junk, then parses an
// optional '-' and decimal digits until a non-digit — scanf-style.
func (g *gen) stdUnsh() {
	b := g.b
	b.Label(fnLabel("унш"))
	b.Prologue(16)
	b.MovImm(8, 0)  // accumulator
	b.MovImm(9, 0)  // sign
	b.MovImm(10, 0) // started
	b.Label("std.unsh.read")
	b.MovImm(x0, 0) // fd = stdin
	b.AddImm(1, sp, 0)
	b.MovImm(2, 1)
	b.MovImm(x16, sysRead)
	b.Svc()
	b.CmpImm(x0, 0)
	b.BCond(condLE, "std.unsh.done") // EOF/error
	b.LdrbReg(11, sp)
	b.CmpImm(10, 0)
	b.BCond(condNE, "std.unsh.digit")
	b.CmpImm(11, '-')
	b.BCond(condNE, "std.unsh.digit0")
	b.MovImm(9, 1)
	b.MovImm(10, 1)
	b.B("std.unsh.read")
	b.Label("std.unsh.digit0")
	b.CmpImm(11, '0')
	b.BCond(condLT, "std.unsh.read") // skip junk
	b.CmpImm(11, '9')
	b.BCond(condGT, "std.unsh.read")
	b.MovImm(10, 1)
	b.B("std.unsh.acc")
	b.Label("std.unsh.digit")
	b.CmpImm(11, '0')
	b.BCond(condLT, "std.unsh.done")
	b.CmpImm(11, '9')
	b.BCond(condGT, "std.unsh.done")
	b.Label("std.unsh.acc")
	b.MovImm(12, 10)
	b.Mul(8, 8, 12)
	b.SubImm(11, 11, '0')
	b.Add(8, 8, 11)
	b.B("std.unsh.read")
	b.Label("std.unsh.done")
	b.MovReg(x0, 8)
	b.CmpImm(9, 0)
	b.BCond(condEQ, "std.unsh.ret")
	b.Neg(x0, x0)
	b.Label("std.unsh.ret")
	b.Epilogue(16)
}

// stdMqrShine: мөр_шинэ(урт) allocates a mutable string buffer of урт+1 bytes.
// mmap pages arrive zeroed, so the buffer is born NUL-terminated everywhere.
func (g *gen) stdMqrShine() {
	b := g.b
	b.Label(fnLabel("мөр_шинэ"))
	b.Prologue(0)
	b.AddImm(x0, x0, 1) // room for the trailing NUL
	b.BL(fnLabel("monAlloc"))
	b.Epilogue(0)
}

// stdMonAlloc: bump allocator. Rounds the request to 16 bytes, serves it from
// the current chunk, and mmaps a fresh MAX(n, 1 MiB) chunk when it runs out.
// free is a no-op (memory is reclaimed at exit) — the standard teaching-compiler
// simplification. Heap state persists in two __DATA slots. n is kept in x9 and
// the bump target in x12 across both paths so the tail is shared; both are
// spilled around the mmap syscall (the kernel may clobber x9-x15).
func (g *gen) stdMonAlloc() {
	b := g.b
	b.Label(fnLabel("monAlloc"))
	b.Prologue(16)
	b.AddImm(x0, x0, 15)
	b.MovImm(2, -16)
	b.And(9, x0, 2) // x9 = n rounded up to 16

	b.AdrpAdd(10, heapPtrLabel)
	b.LdrReg(11, 10) // cur = heap_ptr
	b.Add(12, 11, 9) // newp = cur + n
	b.AdrpAdd(13, heapEndLabel)
	b.LdrReg(14, 13) // end = heap_end
	b.Cmp(12, 14)
	b.BCond(condLS, "std.malloc.fit") // newp <= end → fits in current chunk

	// grow: chunk = max(n, heapChunk)
	b.MovImm(2, heapChunk)
	b.Cmp(9, 2)
	b.BCond(condLS, "std.malloc.chunk")
	b.MovReg(2, 9)
	b.Label("std.malloc.chunk")
	b.StrFrame(9, sp, 0) // spill n and chunk across the syscall
	b.StrFrame(2, sp, 8)
	b.MovReg(1, 2)      // len = chunk
	b.MovImm(x0, 0)     // addr = 0
	b.MovImm(2, 3)      // PROT_READ|PROT_WRITE
	b.MovImm(3, 0x1002) // MAP_ANON|MAP_PRIVATE
	b.MovImm(4, -1)     // fd
	b.MovImm(5, 0)      // offset
	b.MovImm(x16, sysMmap)
	b.Svc() // x0 = mapped base
	b.LdrFrame(9, sp, 0)
	b.LdrFrame(1, sp, 8)
	b.Add(14, x0, 1) // end = base + chunk
	b.AdrpAdd(13, heapEndLabel)
	b.StrReg(14, 13)
	b.Add(12, x0, 9) // newp = base + n

	b.Label("std.malloc.fit")
	b.Sub(x0, 12, 9) // result = newp - n
	b.AdrpAdd(10, heapPtrLabel)
	b.StrReg(12, 10) // heap_ptr = newp
	b.Epilogue(16)
}

// itoa emits the shared digit loop for хэвлэ/эхэвлэ. On entry x0 = value (must
// be non-negative); writes into a 32-byte frame buffer and calls write(2).
// signed controls whether a leading '-' is emitted for a negative input.
func (g *gen) stdKhevle() {
	b := g.b
	b.Label(fnLabel("хэвлэ"))
	b.Prologue(32)
	b.MovImm(9, 0) // sign flag
	b.CmpImm(x0, 0)
	b.BCond(condGE, "std.khevle.pos")
	b.MovImm(9, 1)
	b.Neg(x0, x0)
	b.Label("std.khevle.pos")
	b.AddImm(2, sp, 32)
	b.MovImm(3, 10)
	b.Label("std.khevle.loop")
	b.UDiv(4, x0, 3)
	b.MSub(5, 4, 3, x0)
	b.AddImm(5, 5, '0')
	b.SubImm(2, 2, 1)
	b.StrbReg(5, 2)
	b.MovReg(x0, 4)
	b.Cbnz(x0, "std.khevle.loop")
	b.CmpImm(9, 0)
	b.BCond(condEQ, "std.khevle.nosign")
	b.SubImm(2, 2, 1)
	b.MovImm(5, '-')
	b.StrbReg(5, 2)
	b.Label("std.khevle.nosign")
	b.writeBuf(32)
	b.Epilogue(32)
}

func (g *gen) stdEkhevle() {
	b := g.b
	b.Label(fnLabel("эхэвлэ"))
	b.Prologue(32)
	b.AddImm(2, sp, 32)
	b.MovImm(3, 10)
	b.Label("std.ekhevle.loop")
	b.UDiv(4, x0, 3)
	b.MSub(5, 4, 3, x0)
	b.AddImm(5, 5, '0')
	b.SubImm(2, 2, 1)
	b.StrbReg(5, 2)
	b.MovReg(x0, 4)
	b.Cbnz(x0, "std.ekhevle.loop")
	b.writeBuf(32)
	b.Epilogue(32)
}

// writeBuf writes [x2, sp+bufEnd) to stdout: x2 = start pointer, buffer top at
// sp+bufEnd. Clobbers x0/x1/x2/x4/x16.
func (b *Buf) writeBuf(bufEnd int) {
	b.MovReg(1, 2)         // buf = start
	b.AddImm(4, sp, bufEnd) // top
	b.Sub(2, 4, 2)          // len = top - start
	b.MovImm(x0, 1)         // fd = stdout
	b.MovImm(x16, sysWrite)
	b.Svc()
}

func (g *gen) stdMqrUrt() {
	b := g.b
	b.Label(fnLabel("мөр_урт"))
	b.Prologue(0)
	b.MovReg(9, x0) // cursor
	b.MovImm(x0, 0) // length
	b.Label("std.strlen.loop")
	b.LdrbReg(10, 9)
	b.Cbz(10, "std.strlen.done")
	b.AddImm(9, 9, 1)
	b.AddImm(x0, x0, 1)
	b.B("std.strlen.loop")
	b.Label("std.strlen.done")
	b.Epilogue(0)
}

func (g *gen) stdBayt() {
	b := g.b
	b.Label(fnLabel("байт"))
	b.Prologue(0)
	b.Add(x0, x0, 1) // s + i
	b.LdrbReg(x0, x0)
	b.Epilogue(0)

	b.Label(fnLabel("байт_тавих"))
	b.Prologue(0)
	b.Add(x0, x0, 1) // s + i
	b.StrbReg(2, x0) // store byte (x2)
	b.Epilogue(0)
}

func (g *gen) stdMqrPrint() {
	b := g.b
	b.Label(fnLabel("мөр_хэвлэх"))
	b.Prologue(0)
	b.MovReg(9, x0) // cursor for strlen
	b.MovImm(2, 0)  // len
	b.Label("std.puts.len")
	b.LdrbReg(10, 9)
	b.Cbz(10, "std.puts.write")
	b.AddImm(9, 9, 1)
	b.AddImm(2, 2, 1)
	b.B("std.puts.len")
	b.Label("std.puts.write")
	b.MovReg(1, x0) // buf = s
	b.MovImm(x0, 1) // fd
	b.MovImm(x16, sysWrite)
	b.Svc()
	b.Epilogue(0)
}

func (g *gen) stdTemdegt() {
	b := g.b
	b.Label(fnLabel("тэмдэгт_хэвлэх"))
	b.Prologue(16)
	b.AddImm(2, sp, 0) // write cursor at buffer start
	b.CmpImm(x0, 0x80)
	b.BCond(condLO, "std.utf8.b1")
	b.CmpImm(x0, 0x800)
	b.BCond(condLO, "std.utf8.b2")
	b.MovImm(11, 0x10000)
	b.Cmp(x0, 11)
	b.BCond(condLO, "std.utf8.b3")
	// 4-byte
	b.LsrImm(3, x0, 18)
	b.MovImm(4, 0xF0)
	b.Orr(3, 3, 4)
	b.StrbReg(3, 2)
	b.AddImm(2, 2, 1)
	b.emitCont(x0, 12)
	b.B("std.utf8.tail2")
	b.Label("std.utf8.b3")
	b.LsrImm(3, x0, 12)
	b.MovImm(4, 0xE0)
	b.Orr(3, 3, 4)
	b.StrbReg(3, 2)
	b.AddImm(2, 2, 1)
	b.B("std.utf8.tail2")
	b.Label("std.utf8.b2")
	b.LsrImm(3, x0, 6)
	b.MovImm(4, 0xC0)
	b.Orr(3, 3, 4)
	b.StrbReg(3, 2)
	b.AddImm(2, 2, 1)
	b.B("std.utf8.tail1")
	b.Label("std.utf8.tail2")
	b.emitCont(x0, 6)
	b.Label("std.utf8.tail1")
	b.emitCont(x0, 0)
	b.B("std.utf8.write")
	b.Label("std.utf8.b1")
	b.StrbReg(x0, 2)
	b.AddImm(2, 2, 1)
	b.Label("std.utf8.write")
	// bytes were written forward from sp; x2 = end cursor, buf = sp.
	b.AddImm(1, sp, 0) // buf
	b.Sub(2, 2, 1)     // len = end - buf
	b.MovImm(x0, 1)
	b.MovImm(x16, sysWrite)
	b.Svc()
	b.Epilogue(16)
}

// emitCont writes a UTF-8 continuation byte 0x80 | ((cp >> shift) & 0x3F).
func (b *Buf) emitCont(cp, shift int) {
	if shift > 0 {
		b.LsrImm(3, cp, shift)
	} else {
		b.MovReg(3, cp)
	}
	b.MovImm(4, 0x3F)
	b.And(3, 3, 4)
	b.MovImm(4, 0x80)
	b.Orr(3, 3, 4)
	b.StrbReg(3, 2)
	b.AddImm(2, 2, 1)
}

func (g *gen) stdClear() {
	b := g.b
	b.Label(fnLabel("дэлгэцЦэвэрлэх"))
	b.Prologue(0)
	b.AdrpAdd(1, g.internString("\x1b[H\x1b[2J"))
	b.MovImm(2, 7)
	b.MovImm(x0, 1)
	b.MovImm(x16, sysWrite)
	b.Svc()
	b.Epilogue(0)
}

func (g *gen) stdNoops() {
	b := g.b
	for _, name := range []string{"чөлөөлөх", "хүлээх"} {
		b.Label(fnLabel(name))
		b.Prologue(0)
		b.Epilogue(0)
	}
}
