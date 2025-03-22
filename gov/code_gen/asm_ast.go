package codegen

type AsmUnaryOperator string

const (
	Not AsmUnaryOperator = "neg"
	Neg AsmUnaryOperator = "not"
)

type AsmRegister string

const (
	AX  AsmRegister = "eax"  // rax's lower 32 bits
	R10 AsmRegister = "r10d" // r10's lower 32 bits
)

type AsmOperand interface {
	op()
}

type Imm struct {
	Value int
}

func (a Imm) op() {}

type Register struct {
	Reg AsmRegister
}

func (a Register) op() {}

type Pseudo struct {
	Ident string
}

func (a Pseudo) op() {}

type Stack struct {
	Value int
}

func (a Stack) op() {}

type AsmInstruction interface {
	ir()
}

type Mov struct {
	Src AsmOperand
	Dst AsmOperand
}

func (a Mov) ir() {}

type Unary struct {
	Op  AsmUnaryOperator
	Dst AsmOperand
}

func (a Unary) ir() {}

// for prologue subq $n, %rsp that how much stack we need to allocate
type AllocateStack struct {
	Value int
}

func (a AllocateStack) ir() {}

type Return struct{}

func (a Return) ir() {}

type AsmFnDef struct {
	Ident string
	Irs   []AsmInstruction
}

type AsmProgram struct {
	AsmFnDef AsmFnDef
}
