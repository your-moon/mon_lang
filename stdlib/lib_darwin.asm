.globl _khevle
_khevle:
    pushq %rbp
    movq %rsp, %rbp
    subq $48, %rsp        # Allocate local space

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq %rdi, -8(%rbp)   # Store input number (64-bit)
    leaq -32(%rbp), %r12  # Buffer pointer (use r12 instead of rdi)
    movq $0, -16(%rbp)    # Initialize digit count

    movq -8(%rbp), %rax
    cmpq $0, %rax
    jne convert_loop
    movb $'0', (%r12)
    incq -16(%rbp)
    jmp write_number

convert_loop:
    movq -8(%rbp), %rax   # Load input number
    movq $10, %rcx        # Divisor
    xorq %rdx, %rdx       # Clear remainder
    idivq %rcx            # Divide RDX:RAX by RCX -> Quotient in RAX, remainder in RDX
    movq %rax, -8(%rbp)   # Store quotient
    addb $'0', %dl        # Convert remainder to ASCII
    movb %dl, (%r12)      # Store digit
    incq %r12             # Move buffer pointer forward
    incq -16(%rbp)        # Increment digit count
    cmpq $0, -8(%rbp)     # Check if quotient is zero
    jne convert_loop

    # Now reverse digits
    leaq -32(%rbp), %r13  # Start of buffer
    movq %r12, %r14       # End of buffer pointer
    decq %r14             # Point to last digit

reverse_loop:
    cmpq %r13, %r14
    jle write_number
    movb (%r13), %al
    movb (%r14), %bl
    movb %bl, (%r13)
    movb %al, (%r14)
    incq %r13
    decq %r14
    jmp reverse_loop

write_number:
    movb $'\n', (%r12)    # Add newline
    incq -16(%rbp)        # Increment digit count

    movq $0x2000004, %rax # sys_write
    movq $1, %rdi         # stdout
    leaq -32(%rbp), %rsi  # buffer
    movq -16(%rbp), %rdx  # length
    syscall

    # Return original input value (lower 32 bits)
    movq -8(%rbp), %rax

    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    movq %rbp, %rsp
    popq %rbp
    ret
