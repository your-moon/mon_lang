    section .text
    global _start
_start:
    ; push instruction
    mov rbp, 2
    push rbp
    ; return instruction ; not implemented
    ; exit
    mov rax, 60 ; 60 is exit code 
    mov rdi, 0 ; exit value 
    syscall

