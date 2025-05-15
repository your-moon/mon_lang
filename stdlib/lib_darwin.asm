.globl _khevle
_khevle:
    pushq %rbp
    movq %rsp, %rbp
    subq $48, %rsp        # Align stack to 16 bytes and allocate space for locals

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movl %edi, -28(%rbp)

    movl %edi, -4(%rbp)   # Store input number
    leaq -16(%rbp), %rdi  # Buffer position
    movq $0, -24(%rbp)    # Initialize digit count

    cmpl $0, -4(%rbp)
    jne convert_loop
    movb $'0', (%rdi)
    incq -24(%rbp)
    jmp write_number

convert_loop:
    movl -4(%rbp), %eax   # Load number
    movl $10, %ecx        # Divisor for decimal
    xorl %edx, %edx       # Clear remainder
    idivl %ecx            # Divide by 10
    movl %eax, -4(%rbp)   # Store quotient
    addb $'0', %dl        # Convert remainder to ASCII
    movb %dl, (%rdi)      # Store digit
    incq %rdi             # Move buffer pointer
    incq -24(%rbp)        # Increment digit count
    cmpl $0, -4(%rbp)     # Check if quotient is 0
    jne convert_loop      # If not zero, continue loop

    leaq -16(%rbp), %rax  # Start of buffer
    movq %rdi, %rcx       # End of buffer
    decq %rcx             # Point to last digit
reverse_loop:
    cmpq %rax, %rcx       # Compare pointers
    jle write_number      # If start >= end, we're done
    movb (%rax), %dl      # Load from start
    movb (%rcx), %bl      # Load from end
    movb %bl, (%rax)      # Store end at start
    movb %dl, (%rcx)      # Store start at end
    incq %rax             # Move start pointer
    decq %rcx             # Move end pointer
    jmp reverse_loop      # Continue until pointers meet

write_number:
    movb $'\n', (%rdi)    # Store newline
    incq -24(%rbp)        # Include newline in length

    movq $0x2000004, %rax # System call number for write
    movq $1, %rdi         # File descriptor (stdout)
    leaq -16(%rbp), %rsi  # Buffer address
    movq -24(%rbp), %rdx  # Length (digits + newline)
    syscall

    # Restore original input value to return
    movl -28(%rbp), %eax  # Load saved input number into return register

    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx

    movq %rbp, %rsp
    popq %rbp
    ret
