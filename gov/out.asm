    .globl _main
_main:
    # prologue start
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp
    # prologue end

    # push instruction
    movl $2, -4(%rbp)
    push %rax
    # return instruction
    pop %rax
    ret
    # epilogue start
    movq %rbp, %rsp
    popq %rbp
    ret
    # epilogue end
