package tackygen

import (
	"fmt"

	"github.com/your-moon/mn_compiler/mconstant"
)

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
	Modulo           TackyBinaryOp = "%"
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
	Value mconstant.Const
}

func (c Constant) val() string {
	return fmt.Sprintf("%d", c.Value)
}

type StringConstant struct {
	Value string
}

func (s StringConstant) val() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type Var struct {
	Name string
}

func (v Var) val() string {
	return v.Name
}

type Truncate struct {
	Src TackyVal
	Dst TackyVal
}

func (s Truncate) Ir() {
	fmt.Printf("%s truncate:= %s\n", s.Dst.val(), s.Src.val())
}

type SignExtend struct {
	Src TackyVal
	Dst TackyVal
}

func (s SignExtend) Ir() {
	fmt.Printf("%s signextend:= %s\n", s.Dst.val(), s.Src.val())
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

// type TackyExternFn struct {
// 	Name   string
// 	Params []TackyVal
// }
//
// func (f TackyExternFn) Ir() {
// 	fmt.Printf("extern %s(", f.Name)
// 	for i, param := range f.Params {
// 		if i > 0 {
// 			fmt.Printf(", ")
// 		}
// 		fmt.Printf("%s", param.val())
// 	}
// 	fmt.Printf(")\n")
// }

type TackyFn struct {
	Name         string
	Params       []TackyVal
	IsExtern     bool
	Global       bool
	Instructions []Instruction
}

func (f TackyFn) Ir() {
	if f.IsExtern {
		fmt.Printf("extern fn %s \n", f.Name)
		return
	} else if f.Global {
		fmt.Printf("globl fn %s {\n", f.Name)
	} else {
		fmt.Printf("fn %s {\n", f.Name)
	}
	for _, instr := range f.Instructions {
		instr.Ir()
	}
	fmt.Printf("}\n")
}

type TackyProgram struct {
	FnDefs     []TackyFn
	ExternDefs []TackyFn
}

func (p TackyProgram) Ir() {
	for _, fn := range p.FnDefs {
		fn.Ir()
	}
}

type FnCall struct {
	Name string
	Args []TackyVal
	Dst  TackyVal
}

func (f FnCall) Ir() {
	args := ""
	for i, arg := range f.Args {
		if i > 0 {
			args += ", "
		}
		args += arg.val()
	}
	fmt.Printf("%s := call %s(%s)\n", f.Dst.val(), f.Name, args)
}
