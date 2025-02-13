    .globl _main
_main:
    # prologue start
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp
    # prologue end

    # push instruction
    movl $3, -4(%rbp)
    # return instruction # not implemented
    movq %rbp, %rsp
    popq %rbp
    ret

