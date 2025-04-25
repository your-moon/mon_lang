.globl _main
_main:
    pushq %rbp
    movq %rsp, %rbp
    subq $88, %rsp

    movl $10, -4(%rbp)
    movl $5, -8(%rbp)
    movl $2, -12(%rbp)
    movl -4(%rbp), %r10d
    cmpl %r10d, -8(%rbp)
    movl $0, -16(%rbp)
    setg -16(%rbp)
    cmpl $0, -16(%rbp)
    je .Lconditional_else.1
    movl $1, -20(%rbp)
    jmp .Lconditional_end.0
.Lconditional_else.1:
    movl $0, -20(%rbp)
.Lconditional_end.0:
    movl -20(%rbp), %r10d
    movl %r10d, -24(%rbp)
    movl -24(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl -4(%rbp), %r10d
    cmpl %r10d, -8(%rbp)
    movl $0, -28(%rbp)
    setg -28(%rbp)
    cmpl $0, -28(%rbp)
    je .Lconditional_else.5
    movl -8(%rbp), %r10d
    cmpl %r10d, -12(%rbp)
    movl $0, -32(%rbp)
    setg -32(%rbp)
    cmpl $0, -32(%rbp)
    je .Lconditional_else.7
    movl $2, -36(%rbp)
    jmp .Lconditional_end.6
.Lconditional_else.7:
    movl $3, -36(%rbp)
.Lconditional_end.6:
    movl -36(%rbp), %r10d
    movl %r10d, -40(%rbp)
    jmp .Lconditional_end.4
.Lconditional_else.5:
    movl -12(%rbp), %r10d
    cmpl %r10d, -8(%rbp)
    movl $0, -44(%rbp)
    setg -44(%rbp)
    movl -44(%rbp), %r10d
    movl %r10d, -40(%rbp)
.Lconditional_end.4:
    cmpl $0, -40(%rbp)
    je .Lconditional_else.3
    movl $4, -48(%rbp)
    jmp .Lconditional_end.2
.Lconditional_else.3:
    movl $5, -48(%rbp)
.Lconditional_end.2:
    movl -48(%rbp), %r10d
    movl %r10d, -52(%rbp)
    movl -52(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl -4(%rbp), %r10d
    movl %r10d, -56(%rbp)
    movl -8(%rbp), %r10d
    addl %r10d, -56(%rbp)
    movl -12(%rbp), %r10d
    movl %r10d, -60(%rbp)
    movl -60(%rbp), %r11d
    imull $5, %r11d
    movl %r11d, -60(%rbp)
    movl -56(%rbp), %r10d
    cmpl %r10d, -60(%rbp)
    movl $0, -64(%rbp)
    setg -64(%rbp)
    cmpl $0, -64(%rbp)
    je .Lconditional_else.9
    movl -4(%rbp), %r10d
    movl %r10d, -68(%rbp)
    movl -68(%rbp), %r11d
    imull $2, %r11d
    movl %r11d, -68(%rbp)
    movl -68(%rbp), %r10d
    movl %r10d, -72(%rbp)
    jmp .Lconditional_end.8
.Lconditional_else.9:
    movl -8(%rbp), %r10d
    movl %r10d, -76(%rbp)
    movl -76(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -76(%rbp)
    movl -76(%rbp), %r10d
    movl %r10d, -72(%rbp)
.Lconditional_end.8:
    movl -72(%rbp), %r10d
    movl %r10d, -80(%rbp)
    movl -80(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl -24(%rbp), %r10d
    movl %r10d, -84(%rbp)
    movl -52(%rbp), %r10d
    addl %r10d, -84(%rbp)
    movl -84(%rbp), %r10d
    movl %r10d, -88(%rbp)
    movl -80(%rbp), %r10d
    addl %r10d, -88(%rbp)
    movl -88(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
