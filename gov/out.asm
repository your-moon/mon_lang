.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $136, %rsp

    movl $3, -4(%rbp)
    movl -4(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -4(%rbp)
    movl $5, -8(%rbp)
    movl -4(%rbp), %r10d
    addl %r10d, -8(%rbp)
    movl $8, %eax
    cdq
    movl $4, %r10d
    idivl %r10d
    movl %eax, -12(%rbp)
    movl -8(%rbp), %r10d
    movl %r10d, -16(%rbp)
    movl -12(%rbp), %r10d
    addl %r10d, -16(%rbp)
    movl -16(%rbp), %r10d
    movl %r10d, -20(%rbp)
    movl $5, -24(%rbp)
    addl $3, -24(%rbp)
    movl $2, -28(%rbp)
    addl $8, -28(%rbp)
    movl -24(%rbp), %r10d
    movl %r10d, -32(%rbp)
    movl -32(%rbp), %r11d
    imull -28(%rbp), %r11d
    movl %r11d, -32(%rbp)
    movl -32(%rbp), %eax
    cdq
    movl $4, %r10d
    idivl %r10d
    movl %eax, -36(%rbp)
    movl -36(%rbp), %r10d
    movl %r10d, -40(%rbp)
    movl $2, -44(%rbp)
    movl -44(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -44(%rbp)
    movl -44(%rbp), %r10d
    movl %r10d, -48(%rbp)
    movl -48(%rbp), %r11d
    imull $4, %r11d
    movl %r11d, -48(%rbp)
    movl $6, %eax
    cdq
    movl $2, %r10d
    idivl %r10d
    movl %eax, -52(%rbp)
    movl -48(%rbp), %r10d
    movl %r10d, -56(%rbp)
    movl -52(%rbp), %r10d
    addl %r10d, -56(%rbp)
    movl -56(%rbp), %r10d
    movl %r10d, -60(%rbp)
    movl -40(%rbp), %r10d
    cmpl %r10d, -20(%rbp)
    movl $0, -64(%rbp)
    setg -64(%rbp)
    cmpl $0, -64(%rbp)
    je .Land_false.3
    movl -60(%rbp), %r10d
    cmpl %r10d, -40(%rbp)
    movl $0, -68(%rbp)
    setl -68(%rbp)
    cmpl $0, -68(%rbp)
    je .Land_false.3
    movl $1, -72(%rbp)
    jmp .Land_end.4
.Land_false.3:
    movl $0, -72(%rbp)
.Land_end.4:
    cmpl $0, -72(%rbp)
    jne .Lor_true.1
    cmpl $13, -20(%rbp)
    movl $0, -76(%rbp)
    sete -76(%rbp)
    cmpl $0, -76(%rbp)
    jne .Lor_true.1
    movl $0, -80(%rbp)
    jmp .Lor_end.2
.Lor_true.1:
    movl $1, -80(%rbp)
.Lor_end.2:
    cmpl $0, -80(%rbp)
    je .Lif_end.0
.Lif_end.0:
    movl -20(%rbp), %r10d
    movl %r10d, -84(%rbp)
    movl -40(%rbp), %r10d
    addl %r10d, -84(%rbp)
    movl -60(%rbp), %r10d
    movl %r10d, -88(%rbp)
    subl $5, -88(%rbp)
    movl -84(%rbp), %r10d
    movl %r10d, -92(%rbp)
    movl -92(%rbp), %r11d
    imull -88(%rbp), %r11d
    movl %r11d, -92(%rbp)
    movl $2, -96(%rbp)
    addl $1, -96(%rbp)
    movl -92(%rbp), %eax
    cdq
    movl -96(%rbp), %r10d
    idivl %r10d
    movl %eax, -100(%rbp)
    movl -100(%rbp), %r10d
    movl %r10d, -104(%rbp)
    movl -40(%rbp), %r10d
    movl %r10d, -108(%rbp)
    movl -108(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -108(%rbp)
    movl -20(%rbp), %r10d
    movl %r10d, -112(%rbp)
    movl -108(%rbp), %r10d
    addl %r10d, -112(%rbp)
    movl -60(%rbp), %r10d
    movl %r10d, -116(%rbp)
    movl -116(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -116(%rbp)
    movl -116(%rbp), %r10d
    cmpl %r10d, -112(%rbp)
    movl $0, -120(%rbp)
    setg -120(%rbp)
    cmpl $0, -120(%rbp)
    je .Land_false.8
    cmpl $0, -40(%rbp)
    movl $0, -124(%rbp)
    setg -124(%rbp)
    cmpl $0, -124(%rbp)
    je .Land_false.8
    movl $1, -128(%rbp)
    jmp .Land_end.9
.Land_false.8:
    movl $0, -128(%rbp)
.Land_end.9:
    cmpl $0, -128(%rbp)
    jne .Lor_true.6
    cmpl $13, -20(%rbp)
    movl $0, -132(%rbp)
    sete -132(%rbp)
    cmpl $0, -132(%rbp)
    jne .Lor_true.6
    movl $0, -136(%rbp)
    jmp .Lor_end.7
.Lor_true.6:
    movl $1, -136(%rbp)
.Lor_end.7:
    cmpl $0, -136(%rbp)
    je .Lif_end.5
.Lif_end.5:
    movl -104(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
