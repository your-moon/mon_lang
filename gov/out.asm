.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl $1, %r11d
    cmpl $1, %r11d
    movl $0, -4(%rbp)
    sete -4(%rbp)
    cmpl $0, -4(%rbp)
    je .Lelse.0
    movl $3, -8(%rbp)
    movl -8(%rbp), %r11d
    imull $5, %r11d
    movl %r11d, -8(%rbp)
    movl -8(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lelse.0:
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.L.0:
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
