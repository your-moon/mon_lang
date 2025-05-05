_toobod:
    pushq %rbp
    movq %rsp, %rbp
    subq $-16, %rsp

    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $-16, %rsp

    movl $1, %edi
    call _toobod
    movq %rbp, %rsp
    popq %rbp
    ret
