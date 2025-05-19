package codegen

import (
	"fmt"
	"io"

	"github.com/your-moon/mn_compiler_go_version/code_gen/asmtype"
	"github.com/your-moon/mn_compiler_go_version/util"
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
	writer io.Writer
	ostype util.OsType
}

func NewGenASM(writer io.Writer, osType util.OsType) AsmGen {
	return AsmGen{
		writer: writer,
		ostype: osType,
	}
}

func (a *AsmGen) GenAsm(program AsmProgram) {
	for _, fn := range program.AsmExternFn {
		if a.ostype == util.Linux {
			a.Write(fmt.Sprintf(".extern %s", fn.Name))
		} else if a.ostype == util.Darwin {
			a.Write(fmt.Sprintf(".extern _%s", fn.Name))
		}
	}

	if a.ostype == util.Linux {
		a.Write(".globl main")
		a.Write("main:")

	} else if a.ostype == util.Darwin {
		a.Write(".globl _start")
		a.Write("_start:")
	}

	if a.ostype == util.Linux {
		a.Write("    call wndsen")
		a.Write("    movq %rax, %rdi") // Move return value to %rdi for syscall
		a.Write("    movq $60, %rax")  // Exit syscall number
		a.Write("    syscall")         // Make the syscall
	} else if a.ostype == util.Darwin {
		a.Write("    call _wndsen")
		a.Write("    ret") // Return value is already in %rax
	}

	for _, fn := range program.AsmFnDef {
		a.GenFn(fn)
	}
	if a.ostype == util.Linux {
		a.Write(".section note.GNU-stack,\"\",@progbits")
	}
}

func (a *AsmGen) GenFn(fn AsmFnDef) {
	if a.ostype == util.Linux {
		a.Write(fmt.Sprintf("%s:", fn.Ident))
	} else if a.ostype == util.Darwin {
		a.Write(fmt.Sprintf("_%s:", fn.Ident))
	}
	a.Write("    pushq %rbp")
	a.Write("    movq %rsp, %rbp")
	for _, instr := range fn.Irs {
		a.GenInstr(instr)
	}
}

func (a *AsmGen) GenInstr(instr AsmInstruction) {
	switch ast := instr.(type) {
	case Push:
		a.Write(fmt.Sprintf("    pushq %s", a.GenOperand(ast.Op)))
	case DeallocateStack:
		a.Write(fmt.Sprintf("    addq $%d, %%rsp", ast.Value))
	case Call:
		if a.ostype == util.Linux {
			a.Write(fmt.Sprintf("    call %s", ast.Ident))
		} else if a.ostype == util.Darwin {
			a.Write(fmt.Sprintf("    call _%s", ast.Ident))
		}
	case Label:
		a.Write(fmt.Sprintf(".L%s:", ast.Ident))
	case SetCC:
		a.Write(fmt.Sprintf("    set%s %s", ast.CC, a.GenOperand(ast.Op)))
	case JmpCC:
		a.Write(fmt.Sprintf("    j%s .L%s", ast.CC, ast.Ident))
	case Jmp:
		a.Write(fmt.Sprintf("    jmp .L%s", ast.Ident))
	case Cmp:
		fmt.Printf("[debug] Cmp type: %v\n", ast.Type)
		a.Write(fmt.Sprintf("    cmp%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
	case AsmBinary:
		fmt.Printf("[debug] AsmBinary type: %v\n", ast.Type)
		if ast.Op == Mult {
			a.Write(fmt.Sprintf("    imul%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		} else if ast.Op == Add {
			a.Write(fmt.Sprintf("    add%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		} else if ast.Op == Sub {
			a.Write(fmt.Sprintf("    sub%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
		}
	case Cdq:
		fmt.Printf("[debug] Cdq type: %v\n", ast.Type)
		a.Write("    cdq")
	case Idiv:
		fmt.Printf("[debug] Idiv type: %v\n", ast.Type)
		a.Write(fmt.Sprintf("    idiv%s %s", a.GenType(ast.Type), a.GenOperand(ast.Src)))
	case AsmMov:
		fmt.Printf("[debug] AsmMov type: %v\n", ast.Type)
		a.Write(fmt.Sprintf("    mov%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src), a.GenOperand(ast.Dst)))
	case Return:
		a.Write("    movq %rbp, %rsp")
		a.Write("    popq %rbp")
		a.Write("    ret")
	case Unary:
		fmt.Printf("[debug] Unary type: %v\n", ast.Type)
		a.Write(fmt.Sprintf("    %s%s %s", string(ast.Op), a.GenType(ast.Type), a.GenOperand(ast.Dst)))
	case AllocateStack:
		a.Write(fmt.Sprintf("    subq $%d, %%rsp", ast.Value))
		a.Write("")
	}
}

func (a *AsmGen) GenType(ksmtype asmtype.AsmType) string {
	switch ksmtype.(type) {
	case *asmtype.QuadWord:
		return "q"
	case *asmtype.LongWord:
		return "l"
	default:
		panic(fmt.Sprintf("unimplemented gentype: %v", ksmtype))
	}
}

func (a *AsmGen) GenOperand(op AsmOperand) string {
	switch ast := op.(type) {
	case Register:
		return ast.Op()
	case Imm:
		return fmt.Sprintf("$%d", ast.Value)
	case Stack:
		return fmt.Sprintf("%d(%%rbp)", ast.Value)
	default:
		panic("unimplemented operand")
	}
}

func (a *AsmGen) Write(line string) {
	fmt.Fprintln(a.writer, line)
}
