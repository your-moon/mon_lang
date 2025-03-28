.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $4, %rsp

    movl $0, -4(%rbp)
