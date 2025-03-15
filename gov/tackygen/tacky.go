package tackygen

type UnaryOperator string

const (
	Complement UnaryOperator = "Complement"
	Negate     UnaryOperator = "Negate"
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

// instr implements Instruction.
func (u Return) instr() {}

type Unary struct {
	Op  UnaryOperator
	Src TackyVal
	Dst TackyVal
}

// instr implements Instruction.
func (u Unary) instr() {}

type Instruction interface {
	instr()
}

type TackyFn struct {
	Name         string
	Instructions []Instruction
}

type TackyProgram struct {
	FnDef TackyFn
}
