    # fn stmt construct
    .globl mahin
mahin:
    pushq %rbp
    movq %rsp, %rbp
    # return instruction
    movq %rbp, %rsp
    popq %rbp
    ret
