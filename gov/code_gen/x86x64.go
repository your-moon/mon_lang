package codegen

import (
	"fmt"
	"os"
)

type OsType string

const (
	Linux   OsType = "linux"
	Aarch64 OsType = "arch64"
)

// Comment represents an assembly comment
type Comment struct {
	Text string
}

// Ir returns the IR representation of a comment
func (a Comment) Ir() string {
	return fmt.Sprintf("# %s", a.Text)
}

type AsmGen struct {
	file   *os.File
	ostype OsType
}

func NewGenASM(file *os.File, ostype OsType) AsmGen {
	return AsmGen{
		file:   file,
		ostype: ostype,
	}
}

func (a *AsmGen) GenAsm(program AsmProgram) {
	//if its linux
	if a.ostype == Linux {
		a.Write("    .section ,note.GNU-stack,\"\",@progbits")
	}

	a.GenFn(program.AsmFnDef)
}

func (a *AsmGen) GenFn(fn AsmFnDef) {
	a.Write(fmt.Sprintf(".globl %s", fn.Ident))
	a.Write(fmt.Sprintf("%s:", fn.Ident))
	a.Write("    pushq %rbp")
	a.Write("    movq %rsp, %rbp")
	for _, instr := range fn.Irs {
		a.GenInstr(instr)
	}
}

func (a *AsmGen) GenInstr(instr AsmInstruction) {
	switch ast := instr.(type) {
	case Label:
		a.Write(fmt.Sprintf(".L%s:", ast.Ident))
	case SetCC:
		a.Write(fmt.Sprintf("    set%s %s", ast.CC, a.GenOperand(ast.Op)))
	case JmpCC:
		a.Write(fmt.Sprintf("    j%s .L%s", ast.CC, ast.Ident))
	case Jmp:
		a.Write(fmt.Sprintf("    jmp .L%s", ast.Ident))
	case Cmp:
		a.Write(fmt.Sprintf("    cmpl %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
	case AsmBinary:
		if ast.Op == Mult {
			a.Write(fmt.Sprintf("    imull %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		} else if ast.Op == Add {
			a.Write(fmt.Sprintf("    addl %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		} else if ast.Op == Sub {
			a.Write(fmt.Sprintf("    subl %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		}
	case Cdq:
		a.Write("    cdq")
	case Idiv:
		a.Write(fmt.Sprintf("    idivl %s", a.GenOperand(ast.Src)))
	case AsmMov:
		a.Write(fmt.Sprintf("    movl %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
	case Return:
		a.Write("    movq %rbp, %rsp")
		a.Write("    popq %rbp")
		a.Write("    movq %rax, %rdi") // Move return value to %rdi for syscall
		a.Write("    movq $60, %rax")  // Exit syscall number
		a.Write("    syscall")         // Make the syscall
	case Unary:
		a.Write(fmt.Sprintf("    %s %s", string(ast.Op), a.GenOperand(ast.Dst)))
	case AllocateStack:
		a.Write(fmt.Sprintf("    subq $%d, %%rsp", ast.Value))
		a.Write("")
	}
}

func (a *AsmGen) GenOperand(op AsmOperand) string {
	switch ast := op.(type) {
	case Register:
		return string(ast.Reg)
	case Imm:
		return fmt.Sprintf("$%d", ast.Value)
	case Stack:
		return fmt.Sprintf("%d(%%rbp)", ast.Value)
	default:
		panic("unimplemented operand")
	}
}

func (a *AsmGen) Write(line string) {
	a.file.WriteString(line + "\n")
}
