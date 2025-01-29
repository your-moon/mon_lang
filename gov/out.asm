    .globl _main
_main:
    # prologue start
    pushq %rbp
    movq %rsp, %rbp
    subq $8, %rsp
    # prologue end

    # push instruction
    movl $2, -4(%rbp)
    pushq -4(%rbp)
    # print instruction 
    # not implemented 
    movq $1, %rax 
    movq $1, %rdi 
    leaq message(%rip), %rsi 
    movq $1, %rdx
    
    # exit
    movq $60, %rax # 60 is exit code 
    movq $0, %rdi # exit value 
    syscall

    # epilogue start
    movq %rbp, %rsp
    popq %rbp
    ret
    # epilogue end
