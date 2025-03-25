.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl $1, -4(%rbp)
    movl -4(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -4(%rbp)
    movl -4(%rbp), %eax
    cdq
    movl $2, %r10d
    idivl %r10d
    movl %eax, -8(%rbp)
    movl -8(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
