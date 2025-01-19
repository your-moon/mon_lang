	.section	__TEXT,__text,regular,pure_instructions
	.build_version macos, 14, 0	sdk_version 14, 5
	.globl	_main                           ## -- Begin function main
	.p2align	4, 0x90
_main:                                  ## @main
	.cfi_startproc
## %bb.0:
	pushq	%rbp
	.cfi_def_cfa_offset 16
	.cfi_offset %rbp, -16
	movq	%rsp, %rbp
	.cfi_def_cfa_register %rbp
	subq	$272, %rsp                      ## imm = 0x110
	movq	___stack_chk_guard@GOTPCREL(%rip), %rax
	movq	(%rax), %rax
	movq	%rax, -8(%rbp)
	movl	$0, -228(%rbp)
	xorl	%edi, %edi
	leaq	L_.str(%rip), %rsi
	callq	_setlocale
	leaq	-224(%rbp), %rdi
	leaq	l___const.main.source(%rip), %rsi
	movl	$216, %edx
	callq	_memcpy
	movl	-224(%rbp), %esi
	leaq	l_.str.1(%rip), %rdi
	movb	$0, %al
	callq	_wprintf
	leaq	-224(%rbp), %rdi
	callq	_initScanner
LBB0_1:                                 ## =>This Inner Loop Header: Depth=1
	leaq	-256(%rbp), %rdi
	movb	$0, %al
	callq	_scanToken
	movl	-256(%rbp), %esi
	leaq	l_.str.2(%rip), %rdi
	movb	$0, %al
	callq	_wprintf
	movl	-240(%rbp), %esi
	leaq	l_.str.3(%rip), %rdi
	movb	$0, %al
	callq	_wprintf
	cmpl	$23, -256(%rbp)
	jne	LBB0_3
## %bb.2:
	movl	$-1, -228(%rbp)
	jmp	LBB0_6
LBB0_3:                                 ##   in Loop: Header=BB0_1 Depth=1
	cmpl	$24, -256(%rbp)
	jne	LBB0_5
## %bb.4:
	movl	$0, -228(%rbp)
	jmp	LBB0_6
LBB0_5:                                 ##   in Loop: Header=BB0_1 Depth=1
	jmp	LBB0_1
LBB0_6:
	movl	-228(%rbp), %eax
	movl	%eax, -260(%rbp)                ## 4-byte Spill
	movq	___stack_chk_guard@GOTPCREL(%rip), %rax
	movq	(%rax), %rax
	movq	-8(%rbp), %rcx
	cmpq	%rcx, %rax
	jne	LBB0_8
## %bb.7:
	movl	-260(%rbp), %eax                ## 4-byte Reload
	addq	$272, %rsp                      ## imm = 0x110
	popq	%rbp
	retq
LBB0_8:
	callq	___stack_chk_fail
	ud2
	.cfi_endproc
                                        ## -- End function
	.section	__TEXT,__cstring,cstring_literals
L_.str:                                 ## @.str
	.space	1

	.section	__TEXT,__const
	.p2align	4, 0x0                          ## @__const.main.source
l___const.main.source:
	.long	1090                            ## 0x442
	.long	1072                            ## 0x430
	.long	1085                            ## 0x43d
	.long	1080                            ## 0x438
	.long	1075                            ## 0x433
	.long	1095                            ## 0x447
	.long	32                              ## 0x20
	.long	1079                            ## 0x437
	.long	1072                            ## 0x430
	.long	1088                            ## 0x440
	.long	1083                            ## 0x43b
	.long	32                              ## 0x20
	.long	40                              ## 0x28
	.long	49                              ## 0x31
	.long	32                              ## 0x20
	.long	43                              ## 0x2b
	.long	32                              ## 0x20
	.long	49                              ## 0x31
	.long	41                              ## 0x29
	.long	42                              ## 0x2a
	.long	50                              ## 0x32
	.long	47                              ## 0x2f
	.long	51                              ## 0x33
	.long	32                              ## 0x20
	.long	49                              ## 0x31
	.long	50                              ## 0x32
	.long	51                              ## 0x33
	.long	32                              ## 0x20
	.long	10                              ## 0xa
	.long	32                              ## 0x20
	.long	1092                            ## 0x444
	.long	1085                            ## 0x43d
	.long	32                              ## 0x20
	.long	1086                            ## 0x43e
	.long	1088                            ## 0x440
	.long	1091                            ## 0x443
	.long	1091                            ## 0x443
	.long	1083                            ## 0x43b
	.long	1072                            ## 0x430
	.long	1093                            ## 0x445
	.long	32                              ## 0x20
	.long	1093                            ## 0x445
	.long	1086                            ## 0x43e
	.long	1086                            ## 0x43e
	.long	1089                            ## 0x441
	.long	1086                            ## 0x43e
	.long	1085                            ## 0x43d
	.long	32                              ## 0x20
	.long	1093                            ## 0x445
	.long	1101                            ## 0x44d
	.long	1088                            ## 0x440
	.long	1074                            ## 0x432
	.long	0                               ## 0x0
	.long	0                               ## 0x0

	.p2align	2, 0x0                          ## @.str.1
l_.str.1:
	.long	67                              ## 0x43
	.long	85                              ## 0x55
	.long	82                              ## 0x52
	.long	82                              ## 0x52
	.long	69                              ## 0x45
	.long	78                              ## 0x4e
	.long	84                              ## 0x54
	.long	32                              ## 0x20
	.long	37                              ## 0x25
	.long	108                             ## 0x6c
	.long	99                              ## 0x63
	.long	10                              ## 0xa
	.long	0                               ## 0x0

	.p2align	2, 0x0                          ## @.str.2
l_.str.2:
	.long	84                              ## 0x54
	.long	72                              ## 0x48
	.long	73                              ## 0x49
	.long	83                              ## 0x53
	.long	32                              ## 0x20
	.long	73                              ## 0x49
	.long	83                              ## 0x53
	.long	32                              ## 0x20
	.long	84                              ## 0x54
	.long	79                              ## 0x4f
	.long	75                              ## 0x4b
	.long	69                              ## 0x45
	.long	78                              ## 0x4e
	.long	32                              ## 0x20
	.long	37                              ## 0x25
	.long	100                             ## 0x64
	.long	10                              ## 0xa
	.long	0                               ## 0x0

	.p2align	2, 0x0                          ## @.str.3
l_.str.3:
	.long	84                              ## 0x54
	.long	72                              ## 0x48
	.long	73                              ## 0x49
	.long	83                              ## 0x53
	.long	32                              ## 0x20
	.long	73                              ## 0x49
	.long	83                              ## 0x53
	.long	32                              ## 0x20
	.long	76                              ## 0x4c
	.long	101                             ## 0x65
	.long	110                             ## 0x6e
	.long	103                             ## 0x67
	.long	116                             ## 0x74
	.long	104                             ## 0x68
	.long	32                              ## 0x20
	.long	37                              ## 0x25
	.long	100                             ## 0x64
	.long	10                              ## 0xa
	.long	10                              ## 0xa
	.long	0                               ## 0x0

.subsections_via_symbols
