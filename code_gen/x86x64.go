package codegen

import (
	"fmt"
	"io"

	"github.com/your-moon/mon_lang/base"
	"github.com/your-moon/mon_lang/code_gen/asmtype"
	"github.com/your-moon/mon_lang/stringpool"
	"github.com/your-moon/mon_lang/util"
)

type OsType string

const (
	Linux   OsType = "linux"
	Aarch64 OsType = "arch64"
)

type Comment struct {
	Text string
}

func (a Comment) Ir() string {
	return fmt.Sprintf("# %s", a.Text)
}

type AsmGen struct {
	writer          io.Writer
	ostype          util.OsType
	currentInstrIdx int
	currentFn       string
}

func NewGenASM(writer io.Writer, osType util.OsType) AsmGen {
	return AsmGen{
		writer:          writer,
		ostype:          osType,
		currentInstrIdx: 0,
		currentFn:       "",
	}
}

func (a *AsmGen) AddString(value string) string {
	if base.Debug {
		fmt.Printf("[DEBUG] AddString: '%s'\n", value)
		fmt.Printf("[DEBUG] AddString Raw runes: ")
		for _, r := range value {
			fmt.Printf("%d ", r)
		}
		fmt.Println()
	}

	label := stringpool.GetLabel(value)

	return label
}

func (a *AsmGen) GenStringData() {
	strings := stringpool.GetAllStrings()

	if len(strings) > 0 {
		a.Write(".data")
		for str, id := range strings {
			label := fmt.Sprintf(".LC%d", id)
			a.Write(fmt.Sprintf("%s:", label))
			a.Write(fmt.Sprintf(".string \"%s\"", str))
		}
		a.Write("")
	}
}

func (a *AsmGen) GenAsm(program AsmProgram) {
	for _, fn := range program.AsmFnDef {
		for _, instr := range fn.Irs {
			switch ast := instr.(type) {
			case StringLiteral:
				a.AddString(ast.Value)
			case AsmMov:
				if strLit, isStrLit := ast.Src.(StringLiteral); isStrLit {
					a.AddString(strLit.Value)
				}
			}
		}
	}

	a.GenStringData()

	a.Write(".text")

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
		a.Write(".globl _main")
		a.Write("_main:")
	}

	if a.ostype == util.Linux {
		a.Write("    call wndsen")
		a.Write("    movq %rax, %rdi")
		a.Write("    movq $60, %rax")
		a.Write("    syscall")
	} else if a.ostype == util.Darwin {
		a.Write("    call _wndsen")
		a.Write("    ret")
	}

	for _, fn := range program.AsmFnDef {
		a.currentFn = fn.Ident
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
		a.currentInstrIdx += 1
	}
}

func (a *AsmGen) GenInstr(instr AsmInstruction) {
	switch ast := instr.(type) {
	case StringLiteral:
		label := a.AddString(ast.Value)
		a.Write(fmt.Sprintf("    leaq %s(%%rip), %%rax", label))
		return
	case Push:
		a.Write(fmt.Sprintf("    pushq %s", a.GenOperand(ast.Op, &asmtype.QuadWord{})))
	case Call:
		if a.ostype == util.Linux {
			a.Write(fmt.Sprintf("    call %s", ast.Ident))
		} else if a.ostype == util.Darwin {
			a.Write(fmt.Sprintf("    call _%s", ast.Ident))
		}
	case Label:
		a.Write(fmt.Sprintf(".L%s:", ast.Ident))
	case SetCC:
		a.Write(fmt.Sprintf("    set%s %s", ast.CC, a.GenOperand(ast.Op, nil)))
	case JmpCC:
		a.Write(fmt.Sprintf("    j%s .L%s", ast.CC, ast.Ident))
	case Jmp:
		a.Write(fmt.Sprintf("    jmp .L%s", ast.Ident))
	case Cmp:
		a.Write(fmt.Sprintf("    cmp%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type), a.GenOperand(ast.Dst, ast.Type)))
	case AsmBinary:
		if ast.Op == Mult {
			a.Write(fmt.Sprintf("    imul%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type), a.GenOperand(ast.Dst, ast.Type)))
		} else if ast.Op == Add {
			a.Write(fmt.Sprintf("    add%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type), a.GenOperand(ast.Dst, ast.Type)))
		} else if ast.Op == Sub {
			a.Write(fmt.Sprintf("    sub%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type), a.GenOperand(ast.Dst, ast.Type)))
		}
	case Cdq:
		switch ast.Type.(type) {
		case *asmtype.QuadWord:
			a.Write("    cqo")
		case *asmtype.LongWord:
			a.Write("    cdq")

		}
	case Idiv:
		a.Write(fmt.Sprintf("    idiv%s %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type)))
	case AsmMov:
		if strLit, isStrLit := ast.Src.(StringLiteral); isStrLit {
			label := a.AddString(strLit.Value)
			a.Write(fmt.Sprintf("    leaq %s(%%rip), %%rax", label))
			if reg, isReg := ast.Dst.(Register); !isReg || reg.Reg != AX {
				a.Write(fmt.Sprintf("    mov%s %%rax, %s", a.GenType(ast.Type), a.GenOperand(ast.Dst, ast.Type)))
			}
		} else {
			a.Write(fmt.Sprintf("    mov%s %s, %s", a.GenType(ast.Type), a.GenOperand(ast.Src, ast.Type), a.GenOperand(ast.Dst, ast.Type)))
		}
	case Return:
		a.Write("    movq %rbp, %rsp")
		a.Write("    popq %rbp")
		a.Write("    ret")
	case Unary:
		a.Write(fmt.Sprintf("    %s%s %s", string(ast.Op), a.GenType(ast.Type), a.GenOperand(ast.Dst, ast.Type)))
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
	case *asmtype.StringType:
		return "q"
	default:
		panic(fmt.Sprintf("unimplemented gentype: %v, on idx: %d, fn:%s", ksmtype, a.currentInstrIdx, a.currentFn))
	}
}

func (a *AsmGen) GenOperand(op AsmOperand, asmType asmtype.AsmType) string {
	if asmType == nil {
		asmType = &asmtype.QuadWord{}
	}
	switch ast := op.(type) {
	case Register:
		return a.RegisterShow(ast, asmType)
	case Imm:
		if _, ok := asmType.(*asmtype.StringType); ok {
			return fmt.Sprintf("$_str_%d", ast.Value)
		}
		return fmt.Sprintf("$%d", ast.Value)
	case Stack:
		return fmt.Sprintf("%d(%%rbp)", ast.Value)
	case StringLiteral:
		label := a.AddString(ast.Value)
		return fmt.Sprintf("%s(%%rip)", label)
	default:
		panic("unimplemented operand")
	}
}

func (a *AsmGen) RegisterShow(reg Register, asmType asmtype.AsmType) string {
	switch asmType.(type) {
	case *asmtype.QuadWord, *asmtype.StringType:
		if reg.Reg == R10 || reg.Reg == R11 {
			return "%" + string(reg.Reg)
		}
		if reg.Reg == SP {
			return "%rsp"
		}
		return "%r" + string(reg.Reg)
	case *asmtype.LongWord:
		if reg.Reg == R10 || reg.Reg == R11 {
			return "%" + string(reg.Reg) + "d"
		}
		if reg.Reg == SP {
			return "%esp"
		}
		return "%e" + string(reg.Reg)
	default:
		panic(fmt.Sprintf("unimplemented register type: %v", asmType))
	}
}

func (a *AsmGen) Write(line string) {
	fmt.Fprintln(a.writer, line)
}
