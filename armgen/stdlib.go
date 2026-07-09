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
	sysMmap  = 197
)

const heapChunk = 1 << 20 // mmap granularity for the bump allocator

// heap state lives in reserved __DATA slots (see Compile).
const (
	heapPtrLabel = "std.heap_ptr"
	heapEndLabel = "std.heap_end"
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
	g.stdNoops()    // чөлөөлөх (free), хүлээх (sleep) — no-ops
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
