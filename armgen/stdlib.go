/*
 * mon_lang - arm64 built-in stdlib (raw syscalls)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package armgen

// syscall numbers (macOS BSD, x16)
const (
	sysWrite = 4
)

// emitStdlib appends the arm64 built-in functions. Currently: khevle (print a
// signed decimal integer, no newline) — enough to make integer programs
// produce visible output. Digits are built backwards in a 32-byte stack buffer
// and written with a single write(2) syscall.
func (g *gen) emitStdlib() {
	if g.b.HasLabel(fnLabel("хэвлэ")) {
		return
	}
	b := g.b
	b.Label(fnLabel("хэвлэ"))
	b.Prologue(32) // buffer at [sp,#0..32)

	b.MovImm(9, 0) // x9 = sign flag
	b.CmpImm(x0, 0)
	b.BCond(condGE, "stdlib.khevle.pos")
	b.MovImm(9, 1)
	b.Neg(x0, x0)
	b.Label("stdlib.khevle.pos")

	b.AddImm(2, sp, 32) // x2 = end-of-buffer pointer (digits grow down)
	b.MovImm(3, 10)
	b.Label("stdlib.khevle.loop")
	b.UDiv(4, x0, 3)    // x4 = n / 10
	b.MSub(5, 4, 3, x0) // x5 = n - (n/10)*10  = digit
	b.AddImm(5, 5, '0')
	b.SubImm(2, 2, 1)
	b.StrbReg(5, 2)
	b.MovReg(x0, 4)
	b.Cbnz(x0, "stdlib.khevle.loop")

	b.CmpImm(9, 0)
	b.BCond(condEQ, "stdlib.khevle.nosign")
	b.SubImm(2, 2, 1)
	b.MovImm(5, '-')
	b.StrbReg(5, 2)
	b.Label("stdlib.khevle.nosign")

	b.MovReg(1, 2)      // x1 = buf
	b.AddImm(4, sp, 32) // x4 = end
	b.Sub(2, 4, 2)      // x2 = len = end - buf
	b.MovImm(x0, 1)     // fd = stdout
	b.MovImm(x16, sysWrite)
	b.Svc()

	b.Epilogue(32)
}
