.global _main
.data
    # Test messages
    test_header: .asciz "=== Testing Library Functions ===\n"
    test_footer: .asciz "=== All Tests Completed ===\n"
    newline: .asciz "\n"
    
    # khevle test messages
    khevle_header: .asciz "\n--- Testing khevle (print number) ---\n"
    khevle_test1: .asciz "Testing positive number: "
    khevle_test2: .asciz "Testing zero: "
    khevle_test3: .asciz "Testing negative number: "
    khevle_test4: .asciz "Testing large number: "
    khevle_test5: .asciz "Testing very large number: "
    
    # unsh test messages
    unsh_header: .asciz "\n--- Testing unsh (read number) ---\n"
    unsh_test1: .asciz "Testing positive number (12345): "
    unsh_test2: .asciz "Testing negative number (-42): "
    unsh_test3: .asciz "Testing large number (999999999): "
    
    # khevle_mqr test messages
    mqr_header: .asciz "\n--- Testing khevle_mqr (print string) ---\n"
    mqr_test1: .asciz "Testing normal string: "
    mqr_test2: .asciz "Testing empty string: "
    mqr_test3: .asciz "Testing string with special chars: "
    mqr_test4: .asciz "Testing long string: "
    
    # sanamsargwyToo test messages
    rand_header: .asciz "\n--- Testing sanamsargwyToo (random number) ---\n"
    rand_test1: .asciz "Generating 10 random numbers (should be between 1-100):\n"
    rand_result: .asciz "Random number: "
    
    # unsh32 test messages
    unsh32_header: .asciz "\n--- Testing unsh32 (read 32-bit number) ---\n"
    unsh32_test1: .asciz "Testing normal 32-bit number (12345): "
    unsh32_test2: .asciz "Testing large 32-bit number (2147483647): "
    unsh32_test3: .asciz "Testing negative 32-bit number (-2147483648): "
    
    # odoo test messages
    odoo_header: .asciz "\n--- Testing odoo (get current time) ---\n"
    odoo_result: .asciz "Current Unix timestamp: "

    # Test strings
    test_str1: .asciz "Hello, World!"
    test_str2: .asciz ""
    test_str3: .asciz "!@#$%^&*()"
    test_str4: .asciz "This is a very long string that should be printed correctly without any issues"

.text
_main:
    # Print test header
    leaq test_header(%rip), %rdi
    call _khevle_mqr

    # Test khevle
    leaq khevle_header(%rip), %rdi
    call _khevle_mqr

    # Test positive number
    leaq khevle_test1(%rip), %rdi
    call _khevle_mqr
    movq $12345, %rdi
    call _khevle

    # Test zero
    leaq khevle_test2(%rip), %rdi
    call _khevle_mqr
    movq $0, %rdi
    call _khevle

    # Test negative number
    leaq khevle_test3(%rip), %rdi
    call _khevle_mqr
    movq $-42, %rdi
    call _khevle

    # Test large number
    leaq khevle_test4(%rip), %rdi
    call _khevle_mqr
    movq $999999999, %rdi
    call _khevle

    # Test very large number
    leaq khevle_test5(%rip), %rdi
    call _khevle_mqr
    movq $9223372036854775807, %rdi  # Max 64-bit signed integer
    call _khevle

    # Test unsh with predefined values
    leaq unsh_header(%rip), %rdi
    call _khevle_mqr

    # Test positive number
    leaq unsh_test1(%rip), %rdi
    call _khevle_mqr
    movq $12345, %rdi
    call _khevle

    # Test negative number
    leaq unsh_test2(%rip), %rdi
    call _khevle_mqr
    movq $-42, %rdi
    call _khevle

    # Test large number
    leaq unsh_test3(%rip), %rdi
    call _khevle_mqr
    movq $999999999, %rdi
    call _khevle

    # Test khevle_mqr
    leaq mqr_header(%rip), %rdi
    call _khevle_mqr

    # Test normal string
    leaq mqr_test1(%rip), %rdi
    call _khevle_mqr
    leaq test_str1(%rip), %rdi
    call _khevle_mqr

    # Test empty string
    leaq mqr_test2(%rip), %rdi
    call _khevle_mqr
    leaq test_str2(%rip), %rdi
    call _khevle_mqr

    # Test string with special chars
    leaq mqr_test3(%rip), %rdi
    call _khevle_mqr
    leaq test_str3(%rip), %rdi
    call _khevle_mqr

    # Test long string
    leaq mqr_test4(%rip), %rdi
    call _khevle_mqr
    leaq test_str4(%rip), %rdi
    call _khevle_mqr

    # Test sanamsargwyToo
    leaq rand_header(%rip), %rdi
    call _khevle_mqr
    leaq rand_test1(%rip), %rdi
    call _khevle_mqr

    # Generate 10 random numbers
    movq $10, %rcx
rand_loop:
    pushq %rcx
    leaq rand_result(%rip), %rdi
    call _khevle_mqr
    call _sanamsargwyToo
    movq %rax, %rdi
    call _khevle
    popq %rcx
    loop rand_loop

    # Test unsh32 with predefined values
    leaq unsh32_header(%rip), %rdi
    call _khevle_mqr

    # Test normal 32-bit number
    leaq unsh32_test1(%rip), %rdi
    call _khevle_mqr
    movq $12345, %rdi
    call _khevle

    # Test large 32-bit number
    leaq unsh32_test2(%rip), %rdi
    call _khevle_mqr
    movq $2147483647, %rdi  # Max 32-bit signed integer
    call _khevle

    # Test negative 32-bit number
    leaq unsh32_test3(%rip), %rdi
    call _khevle_mqr
    movq $-2147483648, %rdi  # Min 32-bit signed integer
    call _khevle

    # Test odoo
    leaq odoo_header(%rip), %rdi
    call _khevle_mqr
    leaq odoo_result(%rip), %rdi
    call _khevle_mqr
    call _odoo
    movq %rax, %rdi
    call _khevle

    # Print test footer
    leaq test_footer(%rip), %rdi
    call _khevle_mqr

    # Exit program
    movq $0x2000001, %rax    # syscall: exit
    xorq %rdi, %rdi          # exit code 0
    syscall

.data
empty_string: .asciz "" 