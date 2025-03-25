package tackygen

import "fmt"

type TackyBinaryOp string

const (
	Add              TackyBinaryOp = "add"
	Sub              TackyBinaryOp = "sub"
	Mul              TackyBinaryOp = "mul"
	Div              TackyBinaryOp = "div"
	Equal            TackyBinaryOp = "=="
	NotEqual         TackyBinaryOp = "!="
	LessThan         TackyBinaryOp = "<"
	LessThanEqual    TackyBinaryOp = "<="
	GreaterThan      TackyBinaryOp = ">"
	GreaterThanEqual TackyBinaryOp = ">="
	Remainder        TackyBinaryOp = "remainder"
)

type UnaryOperator string

const (
	Complement UnaryOperator = "Complement"
	Negate     UnaryOperator = "Negate"
	Not        UnaryOperator = "Not"
	Unknown    UnaryOperator = "Unknown"
)

type TackyVal interface {
	val() string
}

type Constant struct {
	Value int
}

func (c Constant) val() string {
	return fmt.Sprintf("Constant(%d)", c.Value)
}

type Var struct {
	Name string
}

func (v Var) val() string {
	return fmt.Sprintf("Var(%s)", v.Name)
}

type Copy struct {
	Src TackyVal
	Dst TackyVal
}

func (j Copy) Ir() {
	fmt.Println(fmt.Sprintf("COPY %s, %s", j.Src.val(), j.Dst.val()))
}

type Jump struct {
	Target string
}

func (j Jump) Ir() {
	fmt.Println(fmt.Sprintf("JUMP %s", j.Target))
}

type JumpIfZero struct {
	Val   TackyVal
	Ident string
}

// Ir implements Instruction.
func (j JumpIfZero) Ir() {
	fmt.Println(fmt.Sprintf("JUMPIFZERO %s %s", j.Val.val(), j.Ident))
}

type JumpIfNotZero struct {
	Val   TackyVal
	Ident string
}

func (j JumpIfNotZero) Ir() {
	fmt.Println(fmt.Sprintf("JUMPIFNOTZERO %s %s", j.Val.val(), j.Ident))
}

type Label struct {
	Ident string
}

func (u Label) Ir() {
	fmt.Println(fmt.Sprintf("LABEL %s", u.Ident))
}

type Return struct {
	Value TackyVal
}

func (u Return) Ir() {
	fmt.Println(fmt.Sprintf("RETURN %s", u.Value.val()))
}

type Binary struct {
	Op   TackyBinaryOp
	Src1 TackyVal
	Src2 TackyVal
	Dst  TackyVal
}

func (u Binary) Ir() {
	fmt.Println(fmt.Sprintf("BINARY %s %s %s %s", u.Op, u.Src1.val(), u.Src2.val(), u.Dst.val()))
}

type Unary struct {
	Op  UnaryOperator
	Src TackyVal
	Dst TackyVal
}

// instr implements Instruction.
func (u Unary) Ir() {
	fmt.Println(fmt.Sprintf("UNARY %s %s %s", u.Op, u.Src.val(), u.Dst.val()))
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
