.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $4, %rsp

    movl $1, %r11d
    cmpl $0, %r11d
    movl $0, -4(%rbp)
    sete -4(%rbp)
