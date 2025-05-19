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
    movq -8(%rbp), %rax   # Copy input to rax for processing
    movq %rax, -24(%rbp)  # Store working copy

    cmpq $0, %rax
    jne convert_loop
    movb $'0', (%r12)
    incq -16(%rbp)
    jmp write_number

convert_loop:
    movq -24(%rbp), %rax   # Load working copy
    movq $10, %rcx        # Divisor
    xorq %rdx, %rdx       # Clear remainder
    idivq %rcx            # Divide RDX:RAX by RCX -> Quotient in RAX, remainder in RDX
    movq %rax, -24(%rbp)   # Store quotient
    addb $'0', %dl        # Convert remainder to ASCII
    movb %dl, (%r12)      # Store digit
    incq %r12             # Move buffer pointer forward
    incq -16(%rbp)        # Increment digit count
    cmpq $0, -24(%rbp)     # Check if quotient is zero
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

.globl _unsh
_unsh:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp                  # Allocate buffer

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq $0, %rdi                   # stdin
    leaq -64(%rbp), %rsi            # buffer address
    movq $64, %rdx                  # max bytes
    movq $0x2000003, %rax           # syscall: read
    syscall

    movq %rax, %rcx                 # store number of bytes read
    cmpq $0, %rcx
    jle invalid_input               # if nothing read or error, return 0

    leaq -64(%rbp), %rsi            # reset pointer to buffer
    movq $0, %rax                   # result = 0

parse_loop:
    cmpq $0, %rcx
    je done_parsing

    movb (%rsi), %bl
    cmpb $'\n', %bl
    je done_parsing

    cmpb $'0', %bl
    jb invalid_input
    cmpb $'9', %bl
    ja invalid_input

    subb $'0', %bl
    movzbq %bl, %rbx               # zero extend to 64 bits using movzbq
    imulq $10, %rax, %rax          # multiply by 10
    addq %rbx, %rax                # add digit

    incq %rsi
    decq %rcx
    jmp parse_loop

done_parsing:
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    movq %rbp, %rsp
    popq %rbp
    ret

invalid_input:
    movq $0, %rax                  # return 0 for invalid input
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    movq %rbp, %rsp
    popq %rbp
    ret
