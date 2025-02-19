package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type IrUnOp string

const (
	IrNegateOp  IrUnOp = "neg"
	IrBitwiseOp IrUnOp = "bitwise"
)

func ConvertToUnOp(tok lexer.TokenType) IrUnOp {
	if tok == lexer.MINUS {
		return IrNegateOp
	}
	if tok == lexer.MINUS {
		return IrBitwiseOp
	}

	panic("can't covert to unop: UNKNOWN TOKEN")
}

type IRValue interface {
	String() string
}

type IRConstant struct {
	Value int64
}

func (i *IRConstant) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *IRConstant) Ir(depth int) string {
	return fmt.Sprintf("%s %d", indent(depth), i.Value)
}

type IRIdent struct {
	Name string
}

func (i *IRIdent) String() string {
	return i.Name
}
func (i *IRIdent) Ir(depth int) string {
	return fmt.Sprintf("%s %s", indent(depth), i.Name)
}

type IR interface {
	Ir(depth int) string
}

type IRVar struct {
	Value IRValue
}

func (i *IRVar) Ir(depth int) string {
	return fmt.Sprintf("%sIR_VAR %s", indent(depth), i.Value.String())
}

type TackyFn struct {
	Name         string
	Instructions []IR
}

func (i *TackyFn) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(
		fmt.Sprintf("%s %s:  \n", indent(depth), i.Name),
	)
	// out.WriteString(fmt.Sprintf("%sIR_FN %s[\n", indent(depth), i.Name))
	//
	// for _, stmt := range i.Instructions {
	// 	out.WriteString(stmt.Ir(depth+1) + "\n")
	// }
	//
	// out.WriteString(fmt.Sprintf("%s]", indent(depth)))
	return out.String()
}

type IRUnary struct {
	Op  IrUnOp
	Src IR
	Dst IR
}

func (i *IRUnary) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(
		fmt.Sprintf("%s %s %s %s \n", indent(depth), i.Op, i.Dst.Ir(depth), i.Src.Ir(depth)),
	)

	return out.String()
}

type IRPrint struct{}

func (i *IRPrint) Ir(depth int) string {
	return fmt.Sprintf("%sIR_PRINT", indent(depth))
}

type IRReturn struct {
	Inner IR
}

func (i *IRReturn) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sIR_RET[\n", indent(depth)))

	if i.Inner != nil {
		out.WriteString(i.Inner.Ir(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s]", indent(depth)))
	return out.String()
}

type IRMov struct {
	Src IR
	Dst IR
}

func (i *IRMov) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sIR_MOV [\n", indent(depth)))

	if i.Src != nil {
		out.WriteString(i.Src.Ir(depth+1) + "\n")
	}
	if i.Dst != nil {
		out.WriteString(i.Dst.Ir(depth+1) + "\n")
	}
	return out.String()
}

type IRBinary struct {
	Value  int64
	Binary int64
}

func (i *IRBinary) Ir(depth int) string {
	return fmt.Sprintf("%sOP_BIN %d", indent(depth), i.Value)
}

func indent(depth int) string {
	return fmt.Sprintf("%s", strings.Repeat("  ", depth))
}
