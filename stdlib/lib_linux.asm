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
