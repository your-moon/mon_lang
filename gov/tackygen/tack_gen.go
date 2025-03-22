package tackygen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

type TackyGen struct {
	Irs       []Instruction
	TempCount uint64
}

func NewTackyGen() TackyGen {
	return TackyGen{
		TempCount: 0,
	}
}

func (c *TackyGen) makeTemp() Var {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) EmitTacky(node *parser.ASTProgram) TackyProgram {
	program := TackyProgram{}

	switch ast := node.FnDef.Stmt.(type) {
	case *parser.ASTReturnStmt:
		if ast.ReturnValue != nil {
			val := c.EmitExpr(ast.ReturnValue)
			c.Irs = append(c.Irs, Return{Value: val})
		}

	}
	program.FnDef = TackyFn{
		Name:         node.FnDef.TokenLiteral(),
		Instructions: c.Irs,
	}

	return program
}

func ToTackyOp(op lexer.TokenType) (UnaryOperator, error) {
	if op == lexer.MINUS {
		return Negate, nil
	}
	if op == lexer.TILDE {
		return Complement, nil
	}

	return Unknown, fmt.Errorf("Cannot convert token to tackyop")
}

func (c *TackyGen) EmitExpr(node parser.ASTExpression) TackyVal {
	switch expr := node.(type) {
	case *parser.ASTIntLitExpression:
		return Constant{Value: int(expr.Value)}
	case *parser.ASTUnary:
		src := c.EmitExpr(expr.Inner)
		dst := c.makeTemp()

		op, err := ToTackyOp(expr.Op)
		if err != nil {
			panic(err)
		}

		instr := Unary{
			Op:  op,
			Src: src,
			Dst: dst,
		}
		c.Irs = append(c.Irs, instr)
		return dst
	default:
		panic("unimplemented")
	}
}

func (c *TackyGen) Emit(op Instruction) {
	c.Irs = append(c.Irs, op)
}
