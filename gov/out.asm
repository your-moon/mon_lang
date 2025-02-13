    section .text
    global _start
_start:
    ; exit
    mov rax, 60 ; 60 is exit code 
    mov rdi, 0 ; exit value 
    syscall

