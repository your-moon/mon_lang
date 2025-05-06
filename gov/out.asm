.globl main
main:
    call wndsen
    movq %rax, %rdi
    movq $60, %rax
    syscall
faktorial:
    pushq %rbp
    movq %rsp, %rbp
    subq $32, %rsp

    movl %edi, -4(%rbp)
    cmpl $1, -4(%rbp)
    movl $0, -8(%rbp)
    setle -8(%rbp)
    cmpl $0, -8(%rbp)
    je .Lif_end.0
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    movq %rax, %rdi
    movq $60, %rax
    syscall
.Lif_end.0:
    movl -4(%rbp), %r10d
    movl %r10d, -12(%rbp)
    subl $1, -12(%rbp)
    movl -12(%rbp), %edi
    call faktorial
    movl %eax, -16(%rbp)
    movl -4(%rbp), %r10d
    movl %r10d, -20(%rbp)
    movl -20(%rbp), %r11d
    imull -16(%rbp), %r11d
    movl %r11d, -20(%rbp)
    movl -20(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    movq %rax, %rdi
    movq $60, %rax
    syscall
wndsen:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl $8, %edi
    call faktorial
    movl %eax, -4(%rbp)
    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    movq %rax, %rdi
    movq $60, %rax
    syscall
.section note.GNU-stack,"",@progbits
