package tackygen

import "fmt"

type BinaryOp string

const (
	Add       BinaryOp = "add"
	Sub       BinaryOp = "sub"
	Mul       BinaryOp = "mul"
	Div       BinaryOp = "div"
	Remainder BinaryOp = "remainder"
)

type UnaryOperator string

const (
	Complement UnaryOperator = "Complement"
	Negate     UnaryOperator = "Negate"
	Unknown    UnaryOperator = "Unknown"
)

type TackyVal interface{}

type Constant struct {
	Value int
}

type Var struct {
	Name string
}

type Return struct {
	Value TackyVal
}

type Binary struct {
	Op   BinaryOp
	Src1 TackyVal
	Src2 TackyVal
	Dst  TackyVal
}

func (u Binary) Ir() {
	fmt.Println("BINARY")
}

// instr implements Instruction.
func (u Return) Ir() {
	fmt.Println("RETURN")
}

type Unary struct {
	Op  UnaryOperator
	Src TackyVal
	Dst TackyVal
}

// instr implements Instruction.
func (u Unary) Ir() {
	fmt.Println("UNARY")
}

type Instruction interface {
	Ir()
}

type TackyFn struct {
	Name         string
	Instructions []Instruction
}

type TackyProgram struct {
	FnDef TackyFn
}
