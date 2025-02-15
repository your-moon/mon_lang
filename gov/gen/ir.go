package gen

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

type IRType string

type IR interface {
	Ir(depth int) string
}

type IRVar struct {
	Value string
}

func (i *IRVar) Ir(depth int) string {
	return fmt.Sprintf("%sIR_VAR", indent(depth))
}

type IRFn struct {
	Name  string
	Stmts []IR
}

func (i *IRFn) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sIR_FN %s[\n", indent(depth), i.Name))

	for _, stmt := range i.Stmts {
		out.WriteString(stmt.Ir(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s]", indent(depth)))
	return out.String()
}

type IRUnary struct {
	Op  IrUnOp
	Src IR
	Dst IR
}

func (i *IRUnary) Ir(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sIR_UNARY %s[\n", indent(depth), i.Op))

	if i.Src != nil {
		out.WriteString(i.Src.Ir(depth+1) + "\n")
	}

	if i.Dst != nil {
		out.WriteString(i.Dst.Ir(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s]", indent(depth)))
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
	Value int64
}

func (i *IRMov) Ir(depth int) string {
	return fmt.Sprintf("%sOP_PUSH %d", indent(depth), i.Value)
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
