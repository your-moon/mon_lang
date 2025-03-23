.globl майн
майн:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl $3, -4(%rbp)
    negl -4(%rbp)
    movl -4(%rbp), %r10d
    movl %r10d, -8(%rbp)
    notl -8(%rbp)
    movl -8(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
