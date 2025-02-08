    .globl _main
_main:
    # prologue start
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp
    # prologue end

    # push instruction
    movq $247, -4(%rbp)
    pushq -4(%rbp)
    # return instruction ; not implemented
    # exit
    movq $60, %rax # 60 is exit code 
    movq $0, %rdi # exit value 
    syscall

    # epilogue start
    movq %rbp, %rsp
    popq %rbp
    ret
    # epilogue end
