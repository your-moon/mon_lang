.globl үндсэн
үндсэн:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl $0, -4(%rbp)
    movl $1, -8(%rbp)
.Lloop_start.0:
    cmpl $5, -8(%rbp)
    movl $0, -12(%rbp)
    setle -12(%rbp)
    cmpl $0, -12(%rbp)
    je .Lloop_end.1
    movl -4(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -8(%rbp), %r10d
    addl %r10d, -16(%rbp)
    movl -16(%rbp), %r10d
    movl %r10d, -4(%rbp)
    movl -8(%rbp), %r10d
    movl %r10d, -8(%rbp)
    addl $1, -8(%rbp)
    jmp .Lloop_start.0
.Lloop_end.1:
    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
