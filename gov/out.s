.data
.balign 8
_str:
	.ascii "hello world"
	.byte 0
/* end data */

.text
.balign 4
.globl _main
_main:
	hint	#34
	stp	x29, x30, [sp, -16]!
	mov	x29, sp
	adrp	x0, _str@page
	add	x0, x0, _str@pageoff
	bl	_puts
	mov	w0, #0
	ldp	x29, x30, [sp], 16
	ret
/* end function main */

