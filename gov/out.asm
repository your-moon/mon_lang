.globl үндсэн
үндсэн:
    pushq %rbp
    movq %rsp, %rbp
    subq $20, %rsp

    movl $0, -4(%rbp)
    movl $1, -8(%rbp)
.Lloop.0:
    cmpl $5, -8(%rbp)
    movl $0, -12(%rbp)
    setle -12(%rbp)
    cmpl $0, -12(%rbp)
    je .Lbreak.loop.0
    movl -4(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -8(%rbp), %r10d
    addl %r10d, -16(%rbp)
    movl -16(%rbp), %r10d
    movl %r10d, -4(%rbp)
    cmpl $4, -8(%rbp)
    movl $0, -20(%rbp)
    sete -20(%rbp)
    cmpl $0, -20(%rbp)
    je .Lif_end.3
    jmp .Lbreak.loop.0
.Lif_end.3:
.Lcontinue.loop.0:
    movl -8(%rbp), %r10d
    movl %r10d, -8(%rbp)
    addl $1, -8(%rbp)
    jmp .Lloop.0
.Lbreak.loop.0:
    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
