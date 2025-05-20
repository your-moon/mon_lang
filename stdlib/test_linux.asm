.section .data
    test_msg: .string "=== Testing Library Functions ===\n\n"
    khevle_msg: .string "--- Testing khevle (print number) ---\n\n"
    unsh_msg: .string "--- Testing unsh (read number) ---\n\n"
    khevle_mqr_msg: .string "--- Testing khevle_mqr (print string) ---\n\n"
    sanamsargwyToo_msg: .string "--- Testing sanamsargwyToo (random number) ---\n\n"
    unsh32_msg: .string "--- Testing unsh32 (read 32-bit number) ---\n\n"
    odoo_msg: .string "--- Testing odoo (get current time) ---\n\n"
    done_msg: .string "=== All Tests Completed ===\n\n"

    # Test strings for khevle_mqr
    test_str1: .string "Hello, World!"
    test_str2: .string ""
    test_str3: .string "!@#$%^&*()"
    test_str4: .string "This is a very long string that should be printed correctly without any issues"

    # Test prompts
    prompt1: .string "Testing positive number (12345): \n"
    prompt2: .string "Testing negative number (-42): \n"
    prompt3: .string "Testing large number (999999999): \n"
    prompt4: .string "Testing normal 32-bit number (12345): \n"
    prompt5: .string "Testing large 32-bit number (2147483647): \n"
    prompt6: .string "Testing negative 32-bit number (-2147483648): \n"

.section .text
.globl _start
_start:
    # Print test header
    movq $1, %rax
    movq $1, %rdi
    leaq test_msg(%rip), %rsi
    movq $32, %rdx
    syscall

    # Test khevle
    movq $1, %rax
    movq $1, %rdi
    leaq khevle_msg(%rip), %rsi
    movq $35, %rdx
    syscall

    # Test positive number
    movq $12345, %rdi
    call khevle

    # Test zero
    movq $0, %rdi
    call khevle

    # Test negative number
    movq $-42, %rdi
    call khevle

    # Test large number
    movq $999999999, %rdi
    call khevle

    # Test very large number
    movq $9223372036854775807, %rdi
    call khevle

    # Test unsh
    movq $1, %rax
    movq $1, %rdi
    leaq unsh_msg(%rip), %rsi
    movq $31, %rdx
    syscall

    # Test positive number input
    movq $1, %rax
    movq $1, %rdi
    leaq prompt1(%rip), %rsi
    movq $35, %rdx
    syscall
    call unsh

    # Test negative number input
    movq $1, %rax
    movq $1, %rdi
    leaq prompt2(%rip), %rsi
    movq $35, %rdx
    syscall
    call unsh

    # Test large number input
    movq $1, %rax
    movq $1, %rdi
    leaq prompt3(%rip), %rsi
    movq $35, %rdx
    syscall
    call unsh

    # Test khevle_mqr
    movq $1, %rax
    movq $1, %rdi
    leaq khevle_mqr_msg(%rip), %rsi
    movq $38, %rdx
    syscall

    # Test normal string
    leaq test_str1(%rip), %rdi
    call khevle_mqr

    # Test empty string
    leaq test_str2(%rip), %rdi
    call khevle_mqr

    # Test string with special chars
    leaq test_str3(%rip), %rdi
    call khevle_mqr

    # Test long string
    leaq test_str4(%rip), %rdi
    call khevle_mqr

    # Test sanamsargwyToo
    movq $1, %rax
    movq $1, %rdi
    leaq sanamsargwyToo_msg(%rip), %rsi
    movq $45, %rdx
    syscall

    # Generate 10 random numbers
    movq $10, %rcx
random_loop:
    pushq %rcx
    call sanamsargwyToo
    movq %rax, %rdi
    call khevle
    popq %rcx
    loop random_loop

    # Test unsh32
    movq $1, %rax
    movq $1, %rdi
    leaq unsh32_msg(%rip), %rsi
    movq $38, %rdx
    syscall

    # Test normal 32-bit number
    movq $1, %rax
    movq $1, %rdi
    leaq prompt4(%rip), %rsi
    movq $42, %rdx
    syscall
    call unsh32

    # Test large 32-bit number
    movq $1, %rax
    movq $1, %rdi
    leaq prompt5(%rip), %rsi
    movq $45, %rdx
    syscall
    call unsh32

    # Test negative 32-bit number
    movq $1, %rax
    movq $1, %rdi
    leaq prompt6(%rip), %rsi
    movq $48, %rdx
    syscall
    call unsh32

    # Test odoo
    movq $1, %rax
    movq $1, %rdi
    leaq odoo_msg(%rip), %rsi
    movq $35, %rdx
    syscall

    call odoo
    movq %rax, %rdi
    call khevle

    # Print completion message
    movq $1, %rax
    movq $1, %rdi
    leaq done_msg(%rip), %rsi
    movq $28, %rdx
    syscall

    # Exit
    movq $60, %rax
    xorq %rdi, %rdi
    syscall 