.globl үндсэн
үндсэн:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl $10, -4(%rbp)
    addl $3, -4(%rbp)
    movl -4(%rbp), %r10d
    movl %r10d, -8(%rbp)
    movl -8(%rbp), %r10d
    movl %r10d, -12(%rbp)
    movl -12(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -12(%rbp)
    movl -12(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -16(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
