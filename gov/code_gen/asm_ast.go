package codegen

import "fmt"

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
	AX  AsmRegister = "%eax"  // rax's lower 32 bits
	R10 AsmRegister = "%r10d" // r10's lower 32 bits
	DX  AsmRegister = "%edx"
	R11 AsmRegister = "%r11d"
)

type AsmOperand interface {
	Op() string
}

type Imm struct {
	Value int
}

func (a Imm) Op() string {
	return fmt.Sprintf("$%d", a.Value)
}

type Register struct {
	Reg AsmRegister
}

func (a Register) Op() string {
	return string(a.Reg)
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

type AsmBinary struct {
	Op  AsmAstBinaryOp
	Src AsmOperand
	Dst AsmOperand
}

func (a AsmBinary) Ir() string {
	return fmt.Sprintf("%s %s, %s", a.Op, a.Src.Op(), a.Dst.Op())
}

type Idiv struct {
	Src AsmOperand
}

func (a Idiv) Ir() string {
	return fmt.Sprintf("idiv %s", a.Src.Op())
}

type Cdq struct {
}

func (a Cdq) Ir() string {
	return "cdq"
}

type AsmMov struct {
	Src AsmOperand
	Dst AsmOperand
}

func (a AsmMov) Ir() string {
	return fmt.Sprintf("mov %s, %s", a.Src.Op(), a.Dst.Op())
}

type Unary struct {
	Op  AsmUnaryOperator
	Dst AsmOperand
}

func (a Unary) Ir() string {
	return fmt.Sprintf("%s %s", a.Op, a.Dst.Op())
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

type AsmProgram struct {
	AsmFnDef AsmFnDef
}
