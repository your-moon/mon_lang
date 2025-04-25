.globl main
main:
    pushq %rbp
    movq %rsp, %rbp
    subq $36, %rsp

    movl $1, %r11d
    cmpl $1, %r11d
    movl $0, -4(%rbp)
    sete -4(%rbp)
    cmpl $0, -4(%rbp)
    je .Lconditional_else.1
    movl $3, %r11d
    cmpl $4, %r11d
    movl $0, -8(%rbp)
    sete -8(%rbp)
    cmpl $0, -8(%rbp)
    je .Lconditional_else.3
    movl $4, -12(%rbp)
    movl -12(%rbp), %r11d
    imull $5, %r11d
    movl %r11d, -12(%rbp)
    movl -12(%rbp), %r10d
    movl %r10d, -16(%rbp)
    addl $3, -16(%rbp)
    movl -16(%rbp), %r10d
    movl %r10d, -20(%rbp)
    jmp .Lconditional_end.2
.Lconditional_else.3:
    movl $4, -24(%rbp)
    movl -24(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -24(%rbp)
    movl -24(%rbp), %r10d
    movl %r10d, -20(%rbp)
.Lconditional_end.2:
    movl -20(%rbp), %r10d
    movl %r10d, -28(%rbp)
    jmp .Lconditional_end.0
.Lconditional_else.1:
    movl $3, -32(%rbp)
    movl -32(%rbp), %r11d
    imull $3, %r11d
    movl %r11d, -32(%rbp)
    movl -32(%rbp), %r10d
    movl %r10d, -28(%rbp)
.Lconditional_end.0:
    movl -28(%rbp), %r10d
    movl %r10d, -36(%rbp)
    movl -36(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    movq %rax, %rdi
    movq $60, %rax
    syscall
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    movq %rax, %rdi
    movq $60, %rax
    syscall
