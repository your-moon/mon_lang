.globl _start
_start:
    # Call khevle with number 1
    movl $15, %edi
    call _khevle

    # Exit program
    movq $60, %rax        # System call number for exit (Linux)
    movq $0, %rdi         # Exit code 0
    syscall


.globl _khevle
_khevle:
    pushq %rbp
    movq %rsp, %rbp
    subq $32, %rsp        # Allocate space for local variables

    # Store number in buffer
    movl %edi, -4(%rbp)   # Store input number
    leaq -16(%rbp), %rdi  # Buffer position
    movq $0, -24(%rbp)    # Initialize digit count

    # Handle zero case
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

    # Reverse the digits
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
    # Add newline
    movb $'\n', (%rdi)    # Store newline
    incq -24(%rbp)        # Include newline in length

    # Write system call
    movq $1, %rax         # System call number for write (Linux)
    movq $1, %rdi         # File descriptor (stdout)
    leaq -16(%rbp), %rsi  # Buffer address
    movq -24(%rbp), %rdx  # Length (digits + newline)
    syscall

    # Exit system call
    movq $60, %rax        # System call number for exit
    xorq %rdi, %rdi       # Exit code 0
    syscall

    movq %rbp, %rsp
    popq %rbp
    ret 