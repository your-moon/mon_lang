.globl тообод
тообод:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl -4(%rbp), %edi
    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.globl үндсэн
үндсэн:
    pushq %rbp
    movq %rsp, %rbp
    subq $16, %rsp

    movl $1, %edi
    call тообод
    movl -4(%rbp), %eax
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
