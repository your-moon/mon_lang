.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $12, %rsp

    movl $1, %r11d
    cmpl $1, %r11d
    movl $0, -4(%rbp)
    sete -4(%rbp)
    cmpl $0, -4(%rbp)
    je .Land_false.0
    movl $2, %r11d
    cmpl $2, %r11d
    movl $0, -8(%rbp)
    sete -8(%rbp)
    cmpl $0, -8(%rbp)
    je .Land_false.0
    movl $1, -12(%rbp)
    jmp .Land_end.0
.Land_false.0:
    movl $0, -12(%rbp)
.Land_end.0:
    movl -12(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
