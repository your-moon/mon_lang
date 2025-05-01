.globl main
main:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl $0, -4(%rbp)
    movl $1, -8(%rbp)
.Lrange_start.0:
    cmpl $5, -8(%rbp)
    movl $0, -12(%rbp)
    setle -12(%rbp)
    cmpl $0, -12(%rbp)
    je .Lrange_break.2
    movl -4(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -8(%rbp), %r10d
    addl %r10d, -16(%rbp)
    movl -16(%rbp), %r10d
    movl %r10d, -4(%rbp)
.Lrange_cont.1:
    movl -8(%rbp), %r10d
    movl %r10d, -8(%rbp)
    addl $1, -8(%rbp)
    jmp .Lrange_start.0
.Lrange_break.2:
    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
