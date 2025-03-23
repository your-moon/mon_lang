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
	case Mov:
		a.Write(fmt.Sprintf("    movl %s, %s", a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
	case Return:
		a.Write("    movq %rbp, %rsp")
		a.Write("    popq %rbp")
		a.Write("    ret")
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
		var regval string
		if ast.Reg == AX {
			regval = string(AX)
		} else if ast.Reg == R10 {
			regval = string(R10)
		}
		return regval
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
