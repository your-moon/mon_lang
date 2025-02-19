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

type Unary struct {
	Op  UnaryOperator
	Src TackyVal
	Dst TackyVal
}

type Instruction interface{}

type TackyFn struct {
	Name         string
	Instructions []Instruction
}

type TackyProgram struct {
	FnDefs []TackyFn
}
