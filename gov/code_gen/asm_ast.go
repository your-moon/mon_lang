package codegen

import "fmt"

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
	AX  AsmRegister = "%eax" // rax's lower 32 bits
	CX  AsmRegister = "%ecx" // rax's lower 32 bits
	DX  AsmRegister = "%edx"
	DI  AsmRegister = "%edi"
	SI  AsmRegister = "%esi"
	R8  AsmRegister = "%r8d"
	R9  AsmRegister = "%r9d"
	R10 AsmRegister = "%r10d" // r10's lower 32 bits
	R11 AsmRegister = "%r11d"
)

func (a AsmRegister) String() string {
	return string(a)
}

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
	Src AsmOperand
	Dst AsmOperand
}

func (a Cmp) Ir() string {
	return fmt.Sprintf("cmp %s, %s", a.Src.Op(), a.Dst.Op())
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
	AsmFnDef []AsmFnDef
}
