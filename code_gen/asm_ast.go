package codegen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/code_gen/asmtype"
)

// cond_code = E | NE | G | GE | L | LE
type CondCode string

const (
	E  CondCode = "e"
	NE CondCode = "ne"
	G  CondCode = "g"
	GE CondCode = "ge"
	L  CondCode = "l"
	LE CondCode = "le"
)

type AsmAstBinaryOp string

const (
	Add  AsmAstBinaryOp = "addl"
	Sub  AsmAstBinaryOp = "subl"
	Mult AsmAstBinaryOp = "imull"
)

type AsmUnaryOperator string

const (
	Not AsmUnaryOperator = "notl"
	Neg AsmUnaryOperator = "negl"
)

type AsmRegister string

const (
	AL AsmRegister = "al"
	AX AsmRegister = "ax"

	CX  AsmRegister = "cx"
	DX  AsmRegister = "dx"
	DI  AsmRegister = "di"
	SI  AsmRegister = "si"
	R8  AsmRegister = "r8"
	R9  AsmRegister = "r9"
	R10 AsmRegister = "r10"
	R11 AsmRegister = "r11"
	SP  AsmRegister = "sp"
)

func (a AsmRegister) String() string {
	return string(a)
}

type AsmOperand interface {
	Op() string
}

type Imm struct {
	Value int64
}

func (a Imm) Op() string {
	return fmt.Sprintf("$%d", a.Value)
}

type Register struct {
	Reg AsmRegister
}

func (a Register) Op() string {
	return a.Reg.Asm32()
}

func (a AsmRegister) Asm32() string {
	switch a {
	case AX:
		return "ax"
	case CX:
		return "cx"
	case DX:
		return "dx"
	case DI:
		return "di"
	case SI:
		return "si"
	case R8:
		return "r8"
	case R9:
		return "r9"
	case R10:
		return "r10"
	case R11:
		return "r11"
	case SP:
		return "%rsp" //rsp must be 64bit
	default:
		panic(fmt.Sprintf("unimplemented register %s", a))
	}
}

type Pseudo struct {
	Ident string
}

func (a Pseudo) Op() string {
	return a.Ident
}

type Stack struct {
	Value int
}

func (a Stack) Op() string {
	return fmt.Sprintf("(%d)", a.Value)
}

type AsmInstruction interface {
	Ir() string
}

type AsmExternFn struct {
	Name string
}

func (a AsmExternFn) Ir() string {
	return fmt.Sprintf("extern %s", a.Name)
}

type DeallocateStack struct {
	Value int
}

func (a DeallocateStack) Ir() string {
	return fmt.Sprintf("addq $%d, %s", a.Value, R10)
}

type Push struct {
	Op AsmOperand
}

func (a Push) Ir() string {
	return fmt.Sprintf("push %s", a.Op.Op())
}

type Call struct {
	Ident string
}

func (a Call) Ir() string {
	return fmt.Sprintf("call %s", a.Ident)
}

type Cmp struct {
	Type asmtype.AsmType
	Src  AsmOperand
	Dst  AsmOperand
}

func (a Cmp) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("cmp %s, %s, %s", a.Src.Op(), a.Dst.Op(), typeStr)
}

type Jmp struct {
	Ident string
}

func (a Jmp) Ir() string {
	return fmt.Sprintf("jmp %s", a.Ident)
}

type JmpCC struct {
	CC    CondCode
	Ident string
}

func (a JmpCC) Ir() string {
	return fmt.Sprintf("jmpcc%s %s", a.CC, a.Ident)
}

type SetCC struct {
	CC CondCode
	Op AsmOperand
}

func (a SetCC) Ir() string {
	return fmt.Sprintf("setcc%s %s", a.CC, a.Op.Op())
}

type Label struct {
	Ident string
}

func (a Label) Ir() string {
	return fmt.Sprintf("label %s", a.Ident)
}

type AsmBinary struct {
	Op   AsmAstBinaryOp
	Type asmtype.AsmType
	Src  AsmOperand
	Dst  AsmOperand
}

func (a AsmBinary) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("%s %s, %s, %s", a.Op, a.Src.Op(), a.Dst.Op(), typeStr)
}

type Idiv struct {
	Type asmtype.AsmType
	Src  AsmOperand
}

func (a Idiv) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("idiv %s, %s", a.Src.Op(), typeStr)
}

type Cdq struct {
	Type asmtype.AsmType
}

func (a Cdq) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("cdq %s", typeStr)
}

type AsmMovSx struct {
	Src AsmOperand
	Dst AsmOperand
}

func (a AsmMovSx) Ir() string {
	return fmt.Sprintf("movsx %s, %s", a.Src.Op(), a.Dst.Op())
}

type AsmMovZb struct {
	Type asmtype.AsmType
	Src  AsmOperand
	Dst  AsmOperand
}

func (a AsmMovZb) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("movzb %s, %s, %s", a.Src.Op(), a.Dst.Op(), typeStr)
}

type AsmMov struct {
	Type asmtype.AsmType
	Src  AsmOperand
	Dst  AsmOperand
}

func (a AsmMov) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("mov %s, %s, %s", a.Src.Op(), a.Dst.Op(), typeStr)
}

type Unary struct {
	Type asmtype.AsmType
	Op   AsmUnaryOperator
	Dst  AsmOperand
}

func (a Unary) Ir() string {
	typeStr := "LongType"
	if _, ok := a.Type.(*asmtype.QuadWord); ok {
		typeStr = "QuadType"
	}
	return fmt.Sprintf("%s %s, %s", a.Op, a.Dst.Op(), typeStr)
}

// for prologue subq $n, %rsp that how much stack we need to allocate
type AllocateStack struct {
	Value int
}

func (a AllocateStack) Ir() string {
	return fmt.Sprintf("subq $%d, %s", a.Value, R10)
}

type Return struct{}

func (a Return) Ir() string {
	return "ret"
}

type AsmFnDef struct {
	Ident string
	Irs   []AsmInstruction
}

type StringLiteral struct {
	Value string
}

func (s StringLiteral) Ir() string {
	return fmt.Sprintf("# string literal: %s", s.Value)
}

func (s StringLiteral) Op() string {
	return fmt.Sprintf("$%s", s.Value)
}

var _ AsmInstruction = StringLiteral{} // Ensure StringLiteral implements AsmInstruction

type AsmProgram struct {
	AsmFnDef    []AsmFnDef
	AsmExternFn []AsmExternFn
}
