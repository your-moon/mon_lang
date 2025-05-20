.globl khevle
khevle:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp           # Allocate space for buffer + vars

    movq %rdi, -8(%rbp)      # Save input number (factorial result)
    leaq -48(%rbp), %rsi     # Pointer to buffer start (write from here)
    movq %rsi, -16(%rbp)     # Save buffer base for reversal
    movq $0, -24(%rbp)       # Digit count

    cmpq $0, -8(%rbp)
    jne .convert_loop

    # Special case for 0
    movb $'0', (%rsi)
    incq -24(%rbp)
    jmp .write_number

.convert_loop:
    movq -8(%rbp), %rax
    movq $10, %rcx
    xorq %rdx, %rdx
    divq %rcx                 # Divide rax by 10, remainder in rdx
    movq %rax, -8(%rbp)       # Store quotient

    addb $'0', %dl
    movb %dl, (%rsi)          # Store digit
    incq %rsi                 # Advance buffer pointer
    incq -24(%rbp)            # Increment digit count

    cmpq $0, -8(%rbp)
    jne .convert_loop

    # Prepare for reverse
    movq -16(%rbp), %rax      # Start = buffer base
    movq %rsi, %rcx           # End = current pointer
    decq %rcx                 # Last digit

.reverse_loop:
    cmpq %rax, %rcx
    jle .write_number         # Done if start >= end

    movb (%rax), %dl          # Swap *rax and *rcx
    movb (%rcx), %bl
    movb %bl, (%rax)
    movb %dl, (%rcx)

    incq %rax
    decq %rcx
    jmp .reverse_loop

.write_number:
    movb $'\n', (%rsi)        # Add newline at end
    incq -24(%rbp)            # Include it in length

    movq $1, %rax             # syscall: write
    movq $1, %rdi             # stdout
    movq -16(%rbp), %rsi      # Buffer address
    movq -24(%rbp), %rdx      # Length
    syscall

    movq %rbp, %rsp
    popq %rbp
    ret

.globl unsh
unsh:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp                  # Allocate buffer

    movq $0, %rdi                   # stdin
    leaq -64(%rbp), %rsi            # buffer address
    movq $64, %rdx                  # max bytes
    movq $0, %rax                   # syscall: read
    syscall

    movq %rax, %rcx                 # store number of bytes read
    cmpq $0, %rcx
    jle .invalid_input              # if nothing read or error, return 0

    leaq -64(%rbp), %rsi            # reset pointer to buffer
    movl $0, %eax                   # result = 0

.parse_loop:
    cmpq $0, %rcx
    je .done_parsing

    movb (%rsi), %bl
    cmpb $'\n', %bl
    je .done_parsing

    cmpb $'0', %bl
    jb .invalid_input
    cmpb $'9', %bl
    ja .invalid_input

    subb $'0', %bl
    movzbl %bl, %ebx
    imull $10, %eax, %eax
    addl %ebx, %eax

    incq %rsi
    decq %rcx
    jmp .parse_loop

.done_parsing:
    movq %rbp, %rsp
    popq %rbp
    ret

.invalid_input:
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret

.globl khevle_mqr
khevle_mqr:
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
    call strlen
    movq %rax, %r13       # Store string length in r13

    # Write the string to stdout
    movq $1, %rax         # syscall: write
    movq $1, %rdi         # stdout
    movq %r12, %rsi       # buffer
    movq %r13, %rdx       # length
    syscall

    # Write newline
    movq $1, %rax         # syscall: write
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
strlen:
    pushq %rbp
    movq %rsp, %rbp

    movq %rdi, %rcx       # String pointer
    movq $0, %rax         # Length counter

strlen_loop:
    cmpb $0, (%rcx)       # Check for null terminator
    je strlen_done
    incq %rax             # Increment length
    incq %rcx             # Move to next character
    jmp strlen_loop

strlen_done:
    movq %rbp, %rsp
    popq %rbp
    ret

.globl sanamsargwyToo
sanamsargwyToo:
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

.globl unsh32
unsh32:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp                  # Allocate buffer

    pushq %rbx
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    movq $0, %rdi                   # stdin (fd 0)
    leaq -64(%rbp), %rsi            # buffer
    movq $64, %rdx                  # max bytes
    movq $0, %rax                   # syscall: read
    syscall

    movq %rax, %rcx                 # rcx = bytes read
    cmpq $0, %rcx
    jle invalid_input_32            # nothing read or error

    leaq -64(%rbp), %rsi            # reset pointer to buffer
    movl $0, %eax                   # result in eax = 0

parse_loop_32:
    cmpq $0, %rcx
    je done_parsing_32

    movb (%rsi), %bl
    cmpb $'\n', %bl
    je done_parsing_32

    cmpb $'0', %bl
    jb invalid_input_32
    cmpb $'9', %bl
    ja invalid_input_32

    subb $'0', %bl
    movzbl %bl, %r8d                # r8d = digit (0–9)

    imull $10, %eax, %eax           # result *= 10
    addl %r8d, %eax                 # result += digit

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

invalid_input_32:
    movl $0, %eax                   # return 0 for invalid input
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    popq %rbx
    movq %rbp, %rsp
    popq %rbp
    ret

.globl odoo
odoo:
    pushq %rbp
    movq %rsp, %rbp
    subq $32, %rsp            # Allocate space for timeval struct

    leaq -16(%rbp), %rdi      # Pointer to timeval struct (tv_sec:8, tv_usec:8)
    movq $0, %rsi             # NULL for timezone
    movq $96, %rax            # gettimeofday syscall number for Linux
    syscall

    movq -16(%rbp), %rax      # Load tv_sec (seconds) into RAX

    movq %rbp, %rsp
    popq %rbp
    ret

.data
newline: .byte 10        # Newline character
