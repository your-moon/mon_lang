.globl тообод
тообод:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.globl үндсэн
үндсэн:
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp

    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    subq $8, %rsp

    movl -4(%rbp), %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
    movl $1, %edi
    call тообод
    movl -8(%rbp), %eax
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
