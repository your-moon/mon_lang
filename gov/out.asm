.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $84, %rsp

    movl $10, -4(%rbp)
    movl $5, -8(%rbp)
    movl $2, -12(%rbp)
    movl -8(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -16(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -16(%rbp)
    movl -4(%rbp), %r10d
    movl %r10d, -20(%rbp)
    movl -16(%rbp), %r10d
    addl %r10d, -20(%rbp)
    movl -12(%rbp), %r10d
    movl %r10d, -24(%rbp)
    movl -24(%rbp), %r11d
    imull $10, %r11d
    movl %r11d, -24(%rbp)
    movl -24(%rbp), %r10d
    cmpl %r10d, -20(%rbp)
    movl $0, -28(%rbp)
    setg -28(%rbp)
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    cmpl $0, -28(%rbp)
    je .Lif_end.0
.Lif_end.0:
    movl -8(%rbp), %r10d
    movl %r10d, -32(%rbp)
    movl -12(%rbp), %r10d
    addl %r10d, -32(%rbp)
    movl -32(%rbp), %r10d
    cmpl %r10d, -4(%rbp)
    movl $0, -36(%rbp)
    setg -36(%rbp)
    cmpl $0, -36(%rbp)
    je .Land_false.4
    movl -8(%rbp), %r10d
    movl %r10d, -40(%rbp)
    movl -40(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -40(%rbp)
    cmpl $10, -40(%rbp)
    movl $0, -44(%rbp)
    sete -44(%rbp)
    cmpl $0, -44(%rbp)
    je .Land_false.4
    movl $1, -48(%rbp)
    jmp .Land_end.5
.Land_false.4:
    movl $0, -48(%rbp)
.Land_end.5:
    cmpl $0, -48(%rbp)
    jne .Lor_true.2
    movl -12(%rbp), %r10d
    movl %r10d, -52(%rbp)
    movl -52(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -52(%rbp)
    movl -8(%rbp), %r10d
    cmpl %r10d, -52(%rbp)
    movl $0, -56(%rbp)
    setl -56(%rbp)
    cmpl $0, -56(%rbp)
    jne .Lor_true.2
    movl $0, -60(%rbp)
    jmp .Lor_end.3
.Lor_true.2:
    movl $1, -60(%rbp)
.Lor_end.3:
    movl $2, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    cmpl $0, -60(%rbp)
    je .Lif_end.1
.Lif_end.1:
    movl -4(%rbp), %r10d
    movl %r10d, -64(%rbp)
    movl -8(%rbp), %r10d
    addl %r10d, -64(%rbp)
    movl -12(%rbp), %r10d
    movl %r10d, -68(%rbp)
    addl $3, -68(%rbp)
    movl -64(%rbp), %r10d
    movl %r10d, -72(%rbp)
    movl -72(%rbp), %r11d
    imull -68(%rbp), %r11d
    movl %r11d, -72(%rbp)
    movl $2, -76(%rbp)
    addl $1, -76(%rbp)
    movl -72(%rbp), %eax
    cdq
    movl -76(%rbp), %r10d
    idivl %r10d
    movl %eax, -80(%rbp)
    movl -80(%rbp), %r10d
    movl %r10d, -84(%rbp)
    movl -84(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
