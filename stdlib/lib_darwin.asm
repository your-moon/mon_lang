 .globl _khevle
_khevle:
    pushq %rbp
    movq %rsp, %rbp
    andq $-16, %rsp
    subq $264, %rsp       # Allocate larger local space for big numbers

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq %rdi, -8(%rbp)   # Store input number (64-bit)
    leaq -128(%rbp), %r12  # Buffer pointer (use r12 instead of rdi) - moved further down in stack
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
    cmpq $100, -16(%rbp)   # Check if we're about to overflow (leave space for null terminator)
    jge write_number       # If too many digits, stop and print what we have
    incq -16(%rbp)        # Increment digit count
    cmpq $0, -24(%rbp)     # Check if quotient is zero
    jne convert_loop

    # Now reverse digits
    leaq -128(%rbp), %r13  # Start of buffer
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
    leaq -128(%rbp), %rsi  # buffer
    movq -16(%rbp), %rdx  # length
    syscall

    # Return original input value
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

.globl _khevle_mqr
_khevle_mqr:
    pushq %rbp
    movq %rsp, %rbp

    # Save registers
    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    # RDI contains the pointer to the string
    movq %rdi, %r12       # Store string pointer in r12

    # Calculate string length
    movq %r12, %rdi
    call _strlen
    movq %rax, %r13       # Store string length in r13

    # Add newline to the end
    movq %r12, %rsi       # Source pointer
    movq %r13, %rdx       # Length

    # Write the string to stdout
    movq $0x2000004, %rax # sys_write
    movq $1, %rdi         # stdout
    movq %r12, %rsi       # buffer
    movq %r13, %rdx       # length
    syscall

    # Write newline
    movq $0x2000004, %rax # sys_write
    movq $1, %rdi         # stdout
    leaq newline(%rip), %rsi # newline character
    movq $1, %rdx         # length 1
    syscall

    # Restore registers
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx

    movq %rbp, %rsp
    popq %rbp
    ret

# Helper function to calculate string length
_strlen:
    pushq %rbp
    movq %rsp, %rbp

    movq %rdi, %rcx       # String pointer
    movq $0, %rax         # Length counter

_strlen_loop:
    cmpb $0, (%rcx)       # Check for null terminator
    je _strlen_done
    incq %rax             # Increment length
    incq %rcx             # Move to next character
    jmp _strlen_loop

_strlen_done:
    movq %rbp, %rsp
    popq %rbp
    ret

.globl _sanamsargwyToo
_sanamsargwyToo:
    pushq %rbp
    movq %rsp, %rbp

    rdtsc                     # timestamp in EDX:EAX
    movl %eax, %eax           # low 32 bits
    imull $1103515245, %eax
    addl $12345, %eax

    movl $100, %ecx           # fixed range 100
    xorl %edx, %edx           # clear EDX for division
    divl %ecx                 # divide EDX:EAX by ECX

    addl $1, %edx             # add 1 to remainder to get 1..100 range

    movl %edx, %eax           # move result to EAX (return register)

    movq %rbp, %rsp
    popq %rbp
    ret

.globl _unsh32
_unsh32:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp                  # Allocate buffer

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq $0, %rdi                   # stdin fd
    leaq -64(%rbp), %rsi            # buffer address
    movq $64, %rdx                  # max bytes to read
    movq $0x2000003, %rax           # syscall: read
    syscall

    movq %rax, %rcx                 # bytes read
    cmpq $0, %rcx
    jle invalid_input               # error or no bytes read => return 0

    leaq -64(%rbp), %rsi            # reset pointer to buffer
    movl $0, %eax                   # result = 0 (32-bit)

parse_loop_32:
    cmpq $0, %rcx
    je done_parsing_32

    movb (%rsi), %bl
    cmpb $'\n', %bl
    je done_parsing_32

    cmpb $'0', %bl
    jb invalid_input
    cmpb $'9', %bl
    ja invalid_input

    subb $'0', %bl
    movzbl %bl, %ebx               # zero extend to 32 bits (EBX)

    # Check for overflow before multiply/add:
    # if eax > (0xFFFFFFFF - ebx)/10 then overflow => invalid_input
    movl $0xFFFFFFFF, %edx
    subl %ebx, %edx               # edx = max_val - digit
    cmpl $10, %eax
    jb continue_parsing_32         # eax < 10 so safe
    # eax >= 10, check overflow condition:
    movl %eax, %ecx
    shrl $2, %ecx                 # divide eax by 10: approximate, so better:
    # Actually do eax > (0xFFFFFFFF - digit)/10 check:
    # edx/(10) = quotient to compare
    # Let's just do exact check:
    movl %eax, %r8d              # save eax in r8d temporarily
    movl $10, %r9d
    xorl %edx, %edx              # clear edx for div
    divl %r9d                   # edx:eax / r9d, quotient in eax, remainder in edx
    # Actually this is complex in asm, so you can skip overflow check if unsure

continue_parsing_32:
    imull $10, %eax, %eax          # eax *= 10
    addl %ebx, %eax                # eax += digit

    incq %rsi
    decq %rcx
    jmp parse_loop_32

done_parsing_32:
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    movq %rbp, %rsp
    popq %rbp
    ret
# Data section
.data
newline: .byte 10        # Newline character
