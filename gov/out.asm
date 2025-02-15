    # fn stmt construct
    .globl _mahin
_mahin:
    pushq %rbp
    movq %rsp, %rbp
    # push instruction
    movl $3, -4(%rbp)
    pushq rbp
    # return instruction
    movq %rbp, %rsp
    popq %rbp
    ret
