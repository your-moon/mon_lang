.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $4, %rsp

    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
