.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl $1, -4(%rbp)
    movl -4(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -4(%rbp)
    movl -4(%rbp), %r10d
    movl %r10d, -8(%rbp)
    addl $12, -8(%rbp)
    movl -8(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
