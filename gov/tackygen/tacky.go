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
	return fmt.Sprintf("%d", c.Value)
}

type Var struct {
	Name string
}

func (v Var) val() string {
	return v.Name
}

type Copy struct {
	Src TackyVal
	Dst TackyVal
}

func (j Copy) Ir() {
	fmt.Printf("%s := %s\n", j.Dst.val(), j.Src.val())
}

type Jump struct {
	Target string
}

func (j Jump) Ir() {
	fmt.Printf("goto %s\n", j.Target)
}

type JumpIfZero struct {
	Val   TackyVal
	Ident string
}

func (j JumpIfZero) Ir() {
	fmt.Printf("if %s == 0 goto %s\n", j.Val.val(), j.Ident)
}

type JumpIfNotZero struct {
	Val   TackyVal
	Ident string
}

func (j JumpIfNotZero) Ir() {
	fmt.Printf("if %s != 0 goto %s\n", j.Val.val(), j.Ident)
}

type Label struct {
	Ident string
}

func (u Label) Ir() {
	fmt.Printf("%s:\n", u.Ident)
}

type Return struct {
	Value TackyVal
}

func (u Return) Ir() {
	fmt.Printf("return %s\n", u.Value.val())
}

type Binary struct {
	Op   TackyBinaryOp
	Src1 TackyVal
	Src2 TackyVal
	Dst  TackyVal
}

func (u Binary) Ir() {
	fmt.Printf("%s := %s %s %s\n",
		u.Dst.val(),
		u.Src1.val(),
		u.Op,
		u.Src2.val())
}

type Unary struct {
	Op  UnaryOperator
	Src TackyVal
	Dst TackyVal
}

func (u Unary) Ir() {
	fmt.Printf("%s := %s %s\n",
		u.Dst.val(),
		u.Op,
		u.Src.val())
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
