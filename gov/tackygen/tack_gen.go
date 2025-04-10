package tackygen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

type TackyGen struct {
	Irs        []Instruction
	TempCount  uint64
	LabelCount uint64
}

func NewTackyGen() TackyGen {
	return TackyGen{
		TempCount:  0,
		LabelCount: 0,
	}
}

func (c *TackyGen) makeTemp() Var {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) makeLabel(prefix string) Var {
	temp := fmt.Sprintf("%s.%d", prefix, c.LabelCount)
	c.TempCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) EmitTacky(node *parser.ASTProgram) TackyProgram {
	program := TackyProgram{}

	for _, stmt := range node.FnDef.BlockItems {
		switch ast := stmt.(type) {
		case *parser.ASTReturnStmt:
			if ast.ReturnValue != nil {
				val := c.EmitExpr(ast.ReturnValue)
				c.Irs = append(c.Irs, Return{Value: val})
			}

		}
	}

	program.FnDef = TackyFn{
		Name:         node.FnDef.TokenLiteral(),
		Instructions: c.Irs,
	}

	return program
}

func ToUnaryTackyOp(op lexer.TokenType) (UnaryOperator, error) {
	if op == lexer.MINUS {
		return Negate, nil
	}
	if op == lexer.TILDE {
		return Complement, nil
	}
	if op == lexer.NOT {
		return Not, nil
	}

	return Unknown, fmt.Errorf("Cannot convert token to tackyop")
}

func ToTackyOp(op parser.ASTBinOp) (TackyBinaryOp, error) {
	if op == parser.ASTBinOp(parser.A_PLUS) {
		return Add, nil
	}
	if op == parser.ASTBinOp(parser.A_MINUS) {
		return Sub, nil
	}
	if op == parser.ASTBinOp(parser.A_DIV) {
		return Div, nil
	}
	if op == parser.ASTBinOp(parser.A_MUL) {
		return Mul, nil
	}
	if op == parser.ASTBinOp(parser.A_EQUALTO) {
		return Equal, nil
	}

	if op == parser.ASTBinOp(parser.A_OR) {
		return Add, fmt.Errorf("or is cannot converted to tacky")
	}
	if op == parser.ASTBinOp(parser.A_AND) {
		return Add, fmt.Errorf("and is cannot converted to tacky")
	}

	if op == parser.ASTBinOp(parser.A_NOTEQUAL) {
		return NotEqual, nil
	}
	if op == parser.ASTBinOp(parser.A_LESSTHAN) {
		return LessThan, nil
	}
	if op == parser.ASTBinOp(parser.A_LESSTHANEQUAL) {
		return LessThanEqual, nil
	}
	if op == parser.ASTBinOp(parser.A_GREATERTHAN) {
		return GreaterThan, nil
	}
	if op == parser.ASTBinOp(parser.A_GREATERTHANEQUAL) {
		return GreaterThanEqual, nil
	}

	return Add, fmt.Errorf("Cannot convert token to tackyop")
}

func (c *TackyGen) EmitAndExpr(expr *parser.ASTBinary) TackyVal {
	falseLabel := c.makeLabel("and_false")
	endLabel := c.makeLabel("and_end")
	dst := c.makeTemp()

	v1 := c.EmitExpr(expr.Left)
	c.Irs = append(c.Irs, JumpIfZero{
		Val:   v1,
		Ident: falseLabel.Name,
	})
	v2 := c.EmitExpr(expr.Right)

	c.Irs = append(c.Irs, []Instruction{
		JumpIfZero{
			Val:   v2,
			Ident: falseLabel.Name,
		},
		Copy{
			Src: Constant{Value: 1},
			Dst: dst,
		},
		Jump{
			Target: endLabel.Name,
		},
		Label{
			Ident: falseLabel.Name,
		},
		Copy{
			Src: Constant{Value: 0},
			Dst: dst,
		},
		Label{
			Ident: endLabel.Name,
		},
	}...)

	return dst
}

func (c *TackyGen) EmitOrExpr(expr *parser.ASTBinary) TackyVal {
	trueLabel := c.makeLabel("or_true")
	endLabel := c.makeLabel("or_end")
	dst := c.makeTemp()

	v1 := c.EmitExpr(expr.Left)
	c.Irs = append(c.Irs, JumpIfNotZero{
		Val:   v1,
		Ident: trueLabel.Name,
	})
	v2 := c.EmitExpr(expr.Right)

	c.Irs = append(c.Irs, []Instruction{
		JumpIfNotZero{
			Val:   v2,
			Ident: trueLabel.Name,
		},
		Copy{
			Src: Constant{Value: 0},
			Dst: dst,
		},
		Jump{
			Target: endLabel.Name,
		},
		Label{
			Ident: trueLabel.Name,
		},
		Copy{
			Src: Constant{Value: 1},
			Dst: dst,
		},
		Label{
			Ident: endLabel.Name,
		},
	}...)

	return dst
}

func (c *TackyGen) EmitExpr(node parser.ASTExpression) TackyVal {
	switch expr := node.(type) {
	case *parser.ASTIntLitExpression:
		return Constant{Value: int(expr.Value)}
	case *parser.ASTUnary:
		src := c.EmitExpr(expr.Inner)
		dst := c.makeTemp()

		op, err := ToUnaryTackyOp(expr.Op)
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
	case *parser.ASTBinary:
		if expr.Op == parser.ASTBinOp(parser.A_AND) {
			return c.EmitAndExpr(expr)
		}
		if expr.Op == parser.ASTBinOp(parser.A_OR) {
			return c.EmitOrExpr(expr)
		}

		op, err := ToTackyOp(expr.Op)
		if err != nil {
			panic(err)
		}

		v1 := c.EmitExpr(expr.Left)
		v2 := c.EmitExpr(expr.Right)
		dst := c.makeTemp()
		instr := Binary{
			Op:   op,
			Src1: v1,
			Src2: v2,
			Dst:  dst,
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
