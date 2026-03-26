.data
.LC2:
.string "\n"
.LC0:
.string "█"
.LC1:
.string " "

.text
.extern _khevle
.extern _mqr_khevlekh
.extern _unsh
.extern _unsh32
.extern _sanamsargwyToo
.extern _odoo
.extern _malloc
.globl _main
_main:
    call _wndsen
    ret
_dwrem110:
    pushq %rbp
    movq %rsp, %rbp
    subq $48, %rsp
    movl %edi, -4(%rbp)
    movl %esi, -8(%rbp)
    movl %edx, -12(%rbp)
    cmpl $1, -4(%rbp)
    movl $0, -16(%rbp)
    sete -16(%rbp)
    cmpl $0, -16(%rbp)
    je .Lif_end.0
    cmpl $1, -8(%rbp)
    movl $0, -20(%rbp)
    sete -20(%rbp)
    cmpl $0, -20(%rbp)
    je .Lif_end.1
    cmpl $1, -12(%rbp)
    movl $0, -24(%rbp)
    sete -24(%rbp)
    cmpl $0, -24(%rbp)
    je .Lif_end.2
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.2:
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.1:
    cmpl $1, -12(%rbp)
    movl $0, -28(%rbp)
    sete -28(%rbp)
    cmpl $0, -28(%rbp)
    je .Lif_end.3
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.3:
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.0:
    cmpl $1, -8(%rbp)
    movl $0, -32(%rbp)
    sete -32(%rbp)
    cmpl $0, -32(%rbp)
    je .Lif_end.4
    cmpl $1, -12(%rbp)
    movl $0, -36(%rbp)
    sete -36(%rbp)
    cmpl $0, -36(%rbp)
    je .Lif_end.5
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.5:
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.4:
    cmpl $1, -12(%rbp)
    movl $0, -40(%rbp)
    sete -40(%rbp)
    cmpl $0, -40(%rbp)
    je .Lif_end.6
    movl $1, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
.Lif_end.6:
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
_tqlqvKhevlekh:
    pushq %rbp
    movq %rsp, %rbp
    subq $64, %rsp
    movq %rdi, -8(%rbp)
    movl $0, -12(%rbp)
.Lwhile.7:
    cmpl $100, -12(%rbp)
    movl $0, -16(%rbp)
    setl -16(%rbp)
    cmpl $0, -16(%rbp)
    je .Lwhile_break.9
    movslq -12(%rbp), %r11
    movq %r11, -24(%rbp)
    movq -24(%rbp), %r10
    movq %r10, -32(%rbp)
    movq -32(%rbp), %r11
    imulq $4, %r11
    movq %r11, -32(%rbp)
    movq -8(%rbp), %r10
    movq %r10, -40(%rbp)
    movq -32(%rbp), %r10
    addq %r10, -40(%rbp)
    movq -40(%rbp), %r10
    movl (%r10), %r11d
    movl %r11d, -44(%rbp)
    cmpl $1, -44(%rbp)
    movl $0, -48(%rbp)
    sete -48(%rbp)
    cmpl $0, -48(%rbp)
    je .Lelse.10
    leaq .LC0(%rip), %rax
    movq %rax, %rdi
    call _mqr_khevlekh
    movl %eax, -52(%rbp)
    jmp .L.11
.Lelse.10:
    leaq .LC1(%rip), %rax
    movq %rax, %rdi
    call _mqr_khevlekh
    movl %eax, -56(%rbp)
.L.11:
    movl -12(%rbp), %r10d
    movl %r10d, -60(%rbp)
    addl $1, -60(%rbp)
    movl -60(%rbp), %r10d
    movl %r10d, -12(%rbp)
.Lwhile_continue.8:
    jmp .Lwhile.7
.Lwhile_break.9:
    leaq .LC2(%rip), %rax
    movq %rax, %rdi
    call _mqr_khevlekh
    movl %eax, -64(%rbp)
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
_ankhnyiTqlqv:
    pushq %rbp
    movq %rsp, %rbp
    subq $96, %rsp
    movl $100, %r10d
    movslq %r10d, %r11
    movq %r11, -8(%rbp)
    movq -8(%rbp), %r10
    movq %r10, -16(%rbp)
    movq -16(%rbp), %r11
    imulq $4, %r11
    movq %r11, -16(%rbp)
    movq -16(%rbp), %rdi
    call _malloc
    movq %rax, -24(%rbp)
    movq -24(%rbp), %r10
    movq %r10, -32(%rbp)
    movl $0, -36(%rbp)
.Lwhile.12:
    cmpl $100, -36(%rbp)
    movl $0, -40(%rbp)
    setl -40(%rbp)
    cmpl $0, -40(%rbp)
    je .Lwhile_break.14
    movslq -36(%rbp), %r11
    movq %r11, -48(%rbp)
    movq -48(%rbp), %r10
    movq %r10, -56(%rbp)
    movq -56(%rbp), %r11
    imulq $4, %r11
    movq %r11, -56(%rbp)
    movq -32(%rbp), %r10
    movq %r10, -64(%rbp)
    movq -56(%rbp), %r10
    addq %r10, -64(%rbp)
    movq -64(%rbp), %r10
    movl $0, %r11d
    movl %r11d, (%r10)
    movl -36(%rbp), %r10d
    movl %r10d, -68(%rbp)
    addl $1, -68(%rbp)
    movl -68(%rbp), %r10d
    movl %r10d, -36(%rbp)
.Lwhile_continue.13:
    jmp .Lwhile.12
.Lwhile_break.14:
    movl $100, %eax
    cdq
    movl $2, %r10d
    idivl %r10d
    movl %eax, -72(%rbp)
    movslq -72(%rbp), %r11
    movq %r11, -80(%rbp)
    movq -80(%rbp), %r10
    movq %r10, -88(%rbp)
    movq -88(%rbp), %r11
    imulq $4, %r11
    movq %r11, -88(%rbp)
    movq -32(%rbp), %r10
    movq %r10, -96(%rbp)
    movq -88(%rbp), %r10
    addq %r10, -96(%rbp)
    movq -96(%rbp), %r10
    movl $1, %r11d
    movl %r11d, (%r10)
    movq -32(%rbp), %rax
    movq %rbp, %rsp
    popq %rbp
    ret
_wndsen:
    pushq %rbp
    movq %rsp, %rbp
    subq $304, %rsp
    call _ankhnyiTqlqv
    movq %rax, -8(%rbp)
    movq -8(%rbp), %r10
    movq %r10, -16(%rbp)
    movl $100, %r10d
    movslq %r10d, %r11
    movq %r11, -24(%rbp)
    movq -24(%rbp), %r10
    movq %r10, -32(%rbp)
    movq -32(%rbp), %r11
    imulq $4, %r11
    movq %r11, -32(%rbp)
    movq -32(%rbp), %rdi
    call _malloc
    movq %rax, -40(%rbp)
    movq -40(%rbp), %r10
    movq %r10, -48(%rbp)
    movl $0, -52(%rbp)
.Lwhile.15:
    cmpl $50, -52(%rbp)
    movl $0, -56(%rbp)
    setl -56(%rbp)
    cmpl $0, -56(%rbp)
    je .Lwhile_break.17
    movq -16(%rbp), %rdi
    call _tqlqvKhevlekh
    movl %eax, -60(%rbp)
    movl $0, -64(%rbp)
.Lwhile.18:
    cmpl $100, -64(%rbp)
    movl $0, -68(%rbp)
    setl -68(%rbp)
    cmpl $0, -68(%rbp)
    je .Lwhile_break.20
    cmpl $0, -64(%rbp)
    movl $0, -72(%rbp)
    sete -72(%rbp)
    cmpl $0, -72(%rbp)
    je .Lconditional_else.22
    movl $0, -76(%rbp)
    jmp .Lconditional_end.21
.Lconditional_else.22:
    movl -64(%rbp), %r10d
    movl %r10d, -80(%rbp)
    subl $1, -80(%rbp)
    movslq -80(%rbp), %r11
    movq %r11, -88(%rbp)
    movq -88(%rbp), %r10
    movq %r10, -96(%rbp)
    movq -96(%rbp), %r11
    imulq $4, %r11
    movq %r11, -96(%rbp)
    movq -16(%rbp), %r10
    movq %r10, -104(%rbp)
    movq -96(%rbp), %r10
    addq %r10, -104(%rbp)
    movq -104(%rbp), %r10
    movl (%r10), %r11d
    movl %r11d, -108(%rbp)
    movl -108(%rbp), %r10d
    movl %r10d, -76(%rbp)
.Lconditional_end.21:
    movl -76(%rbp), %r10d
    movl %r10d, -112(%rbp)
    movslq -64(%rbp), %r11
    movq %r11, -120(%rbp)
    movq -120(%rbp), %r10
    movq %r10, -128(%rbp)
    movq -128(%rbp), %r11
    imulq $4, %r11
    movq %r11, -128(%rbp)
    movq -16(%rbp), %r10
    movq %r10, -136(%rbp)
    movq -128(%rbp), %r10
    addq %r10, -136(%rbp)
    movq -136(%rbp), %r10
    movl (%r10), %r11d
    movl %r11d, -140(%rbp)
    movl -140(%rbp), %r10d
    movl %r10d, -144(%rbp)
    movl $100, -148(%rbp)
    subl $1, -148(%rbp)
    movl -148(%rbp), %r10d
    cmpl %r10d, -64(%rbp)
    movl $0, -152(%rbp)
    sete -152(%rbp)
    cmpl $0, -152(%rbp)
    je .Lconditional_else.24
    movl $0, -156(%rbp)
    jmp .Lconditional_end.23
.Lconditional_else.24:
    movl -64(%rbp), %r10d
    movl %r10d, -160(%rbp)
    addl $1, -160(%rbp)
    movslq -160(%rbp), %r11
    movq %r11, -168(%rbp)
    movq -168(%rbp), %r10
    movq %r10, -176(%rbp)
    movq -176(%rbp), %r11
    imulq $4, %r11
    movq %r11, -176(%rbp)
    movq -16(%rbp), %r10
    movq %r10, -184(%rbp)
    movq -176(%rbp), %r10
    addq %r10, -184(%rbp)
    movq -184(%rbp), %r10
    movl (%r10), %r11d
    movl %r11d, -188(%rbp)
    movl -188(%rbp), %r10d
    movl %r10d, -156(%rbp)
.Lconditional_end.23:
    movl -156(%rbp), %r10d
    movl %r10d, -192(%rbp)
    movslq -64(%rbp), %r11
    movq %r11, -200(%rbp)
    movq -200(%rbp), %r10
    movq %r10, -208(%rbp)
    movq -208(%rbp), %r11
    imulq $4, %r11
    movq %r11, -208(%rbp)
    movq -48(%rbp), %r10
    movq %r10, -216(%rbp)
    movq -208(%rbp), %r10
    addq %r10, -216(%rbp)
    movl -112(%rbp), %edi
    movl -144(%rbp), %esi
    movl -192(%rbp), %edx
    call _dwrem110
    movl %eax, -220(%rbp)
    movq -216(%rbp), %r10
    movl -220(%rbp), %r11d
    movl %r11d, (%r10)
    movl -64(%rbp), %r10d
    movl %r10d, -224(%rbp)
    addl $1, -224(%rbp)
    movl -224(%rbp), %r10d
    movl %r10d, -64(%rbp)
.Lwhile_continue.19:
    jmp .Lwhile.18
.Lwhile_break.20:
    movl $0, -64(%rbp)
.Lwhile.25:
    cmpl $100, -64(%rbp)
    movl $0, -228(%rbp)
    setl -228(%rbp)
    cmpl $0, -228(%rbp)
    je .Lwhile_break.27
    movslq -64(%rbp), %r11
    movq %r11, -240(%rbp)
    movq -240(%rbp), %r10
    movq %r10, -248(%rbp)
    movq -248(%rbp), %r11
    imulq $4, %r11
    movq %r11, -248(%rbp)
    movq -16(%rbp), %r10
    movq %r10, -256(%rbp)
    movq -248(%rbp), %r10
    addq %r10, -256(%rbp)
    movslq -64(%rbp), %r11
    movq %r11, -264(%rbp)
    movq -264(%rbp), %r10
    movq %r10, -272(%rbp)
    movq -272(%rbp), %r11
    imulq $4, %r11
    movq %r11, -272(%rbp)
    movq -48(%rbp), %r10
    movq %r10, -280(%rbp)
    movq -272(%rbp), %r10
    addq %r10, -280(%rbp)
    movq -280(%rbp), %r10
    movl (%r10), %r11d
    movl %r11d, -284(%rbp)
    movq -256(%rbp), %r10
    movl -284(%rbp), %r11d
    movl %r11d, (%r10)
    movl -64(%rbp), %r10d
    movl %r10d, -288(%rbp)
    addl $1, -288(%rbp)
    movl -288(%rbp), %r10d
    movl %r10d, -64(%rbp)
.Lwhile_continue.26:
    jmp .Lwhile.25
.Lwhile_break.27:
    movl -52(%rbp), %r10d
    movl %r10d, -292(%rbp)
    addl $1, -292(%rbp)
    movl -292(%rbp), %r10d
    movl %r10d, -52(%rbp)
.Lwhile_continue.16:
    jmp .Lwhile.15
.Lwhile_break.17:
    movl $0, %eax
    movq %rbp, %rsp
    popq %rbp
    ret
