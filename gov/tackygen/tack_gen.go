package tackygen

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/unique"
)

type TackyGen struct {
	Irs        []Instruction
	TempCount  uint64
	LabelCount uint64
	UniqueGen  unique.UniqueGen
}

func NewTackyGen(uniquegen unique.UniqueGen) TackyGen {
	return TackyGen{
		TempCount:  0,
		LabelCount: 0,
		UniqueGen:  uniquegen,
	}
}

func (c *TackyGen) makeTemp() Var {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) continueLabel(id string) Var {
	temp := fmt.Sprintf("continue.%s", id)
	c.LabelCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) breakLabel(id string) Var {
	temp := fmt.Sprintf("break.%s", id)
	c.LabelCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) makeLabel(prefix string) Var {
	temp := fmt.Sprintf("%s.%d", prefix, c.LabelCount)
	c.LabelCount += 1
	return Var{Name: temp}

}

func (c *TackyGen) EmitTacky(node *parser.ASTProgram) TackyProgram {
	program := TackyProgram{}

	for _, stmt := range node.Decls[0].(*parser.FnDecl).Block.BlockItems {
		switch ast := stmt.(type) {
		case *parser.Decl:
			if ast.Expr != nil {
				c.EmitExpr(&parser.ASTAssignment{
					Left:  &parser.ASTVar{Ident: ast.Ident},
					Right: ast.Expr,
				})
			}
		case parser.ASTStmt:
			c.EmitTackyStmt(ast)
		}
	}

	program.FnDef = TackyFn{
		Name:         node.Decls[0].TokenLiteral(),
		Instructions: c.Irs,
	}

	return program
}

func (c *TackyGen) EmitTackyBlock(node parser.ASTBlock) {
	for _, stmt := range node.BlockItems {
		switch ast := stmt.(type) {
		case *parser.Decl:
			if ast.Expr != nil {
				c.EmitExpr(&parser.ASTAssignment{
					Left:  &parser.ASTVar{Ident: ast.Ident},
					Right: ast.Expr,
				})
			}
		case parser.ASTStmt:
			c.EmitTackyStmt(ast)
		}
	}
}

func (c *TackyGen) EmitTackyStmt(node parser.ASTStmt) {
	switch ast := node.(type) {
	case *parser.ASTLoop:
		startLabel := c.makeLabel("loop")
		continueLabel := c.continueLabel(ast.Id)
		breakLabel := c.breakLabel(ast.Id)
		ast.Id = breakLabel.Name

		rangeExpr, ok := ast.Expr.(*parser.ASTRangeExpr)
		if ok {
			rangeStart := c.EmitExpr(rangeExpr.Start)
			loopVar, ok := ast.Var.(*parser.ASTVar)
			if !ok {
				panic("Expected ASTVar for loop variable")
			}
			c.Irs = append(c.Irs, Copy{
				Src: rangeStart,
				Dst: Var{Name: loopVar.Ident},
			})

			c.Irs = append(c.Irs, Label{Ident: startLabel.Name})

			endVal := c.EmitExpr(rangeExpr.End)
			loopVarVal := c.EmitExpr(ast.Var)
			temp := c.makeTemp()
			c.Irs = append(c.Irs, Binary{
				Op:   LessThanEqual,
				Src1: loopVarVal,
				Src2: endVal,
				Dst:  temp,
			})
			c.Irs = append(c.Irs, JumpIfZero{
				Val:   temp,
				Ident: breakLabel.Name,
			})

			c.EmitTackyBlock(ast.Body)

			c.Irs = append(c.Irs, Label{Ident: continueLabel.Name})

			c.Irs = append(c.Irs, Binary{
				Op:   Add,
				Src1: Var{Name: loopVar.Ident},
				Src2: Constant{Value: 1},
				Dst:  Var{Name: loopVar.Ident},
			})

			c.Irs = append(c.Irs, Jump{Target: startLabel.Name})
		} else {
			c.Irs = append(c.Irs, Label{Ident: startLabel.Name})

			condVal := c.EmitExpr(ast.Expr)
			c.Irs = append(c.Irs, JumpIfZero{
				Val:   condVal,
				Ident: breakLabel.Name,
			})

			c.EmitTackyBlock(ast.Body)

			c.Irs = append(c.Irs, Label{Ident: continueLabel.Name})
			c.Irs = append(c.Irs, Jump{Target: startLabel.Name})
		}

		// End label
		c.Irs = append(c.Irs, Label{Ident: breakLabel.Name})
	case *parser.ASTBreakStmt:
		c.Irs = append(c.Irs, Jump{Target: c.breakLabel(ast.Id).Name})
	case *parser.ASTContinueStmt:
		c.Irs = append(c.Irs, Jump{Target: c.continueLabel(ast.Id).Name})
	case *parser.ASTCompoundStmt:
		c.EmitTackyBlock(ast.Block)
	case *parser.ASTIfStmt:
		// no else clause
		if ast.Else == nil {
			endLabel := c.makeLabel("if_end")
			//instruction of cond
			// c = result of cond
			evalCond := c.EmitExpr(ast.Cond)
			// jumpifzero(c, end)
			jmpifzero := JumpIfZero{Val: evalCond, Ident: endLabel.Name}
			c.Irs = append(c.Irs, jmpifzero)

			// instructions of body
			c.EmitTackyStmt(ast.Then)
			// label(end)
			c.Irs = append(c.Irs, Label{Ident: endLabel.Name})

		} else {
			elseLabel := c.makeLabel("else")
			endLabel := c.makeLabel("")
			evalCond := c.EmitExpr(ast.Cond)
			jmpifzero := JumpIfZero{Val: evalCond, Ident: elseLabel.Name}
			c.Irs = append(c.Irs, []Instruction{jmpifzero}...)
			c.EmitTackyStmt(ast.Then)
			c.Irs = append(c.Irs, []Instruction{Label{Ident: elseLabel.Name}}...)
			c.EmitTackyStmt(ast.Else)
			c.Irs = append(c.Irs, []Instruction{Label{Ident: endLabel.Name}}...)
		}

	case *parser.ASTReturnStmt:
		if ast.ReturnValue != nil {
			val := c.EmitExpr(ast.ReturnValue)
			c.Irs = append(c.Irs, Return{Value: val})
		}
	case *parser.ExpressionStmt:
		c.EmitExpr(ast.Expression)
	}
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

	return Unknown, fmt.Errorf("annot convert token to tackyop")
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

	return Add, fmt.Errorf("cannot convert token to tackyop")
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

	left := c.EmitExpr(expr.Left)
	c.Irs = append(c.Irs, JumpIfNotZero{
		Val:   left,
		Ident: trueLabel.Name,
	})
	right := c.EmitExpr(expr.Right)

	c.Irs = append(c.Irs, []Instruction{
		JumpIfNotZero{
			Val:   right,
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
	case *parser.ASTRangeExpr:
		start := c.EmitExpr(expr.Start)
		end := c.EmitExpr(expr.End)
		dst := c.makeTemp()
		c.Irs = append(c.Irs, Binary{
			Op:   Sub,
			Src1: end,
			Src2: start,
			Dst:  dst,
		})
		c.Irs = append(c.Irs, Binary{
			Op:   Add,
			Src1: dst,
			Src2: Constant{Value: 1},
			Dst:  dst,
		})
		return dst

	case *parser.ASTConditional:
		endLabel := c.makeLabel("conditional_end")
		elseLabel := c.makeLabel("conditional_else")
		dst := c.makeTemp()

		condEval := c.EmitExpr(expr.Cond)
		c.Irs = append(c.Irs, JumpIfZero{Val: condEval, Ident: elseLabel.Name})

		evalThen := c.EmitExpr(expr.Then)
		c.Irs = append(c.Irs, Copy{Src: evalThen, Dst: dst})
		c.Irs = append(c.Irs, Jump{Target: endLabel.Name})
		c.Irs = append(c.Irs, Label{Ident: elseLabel.Name})

		evalElse := c.EmitExpr(expr.Else)
		c.Irs = append(c.Irs, Copy{Src: evalElse, Dst: dst})
		c.Irs = append(c.Irs, Label{Ident: endLabel.Name})
		return dst
	case *parser.ASTVar:
		return Var{Name: expr.Ident}
	case *parser.ASTAssignment:
		astVar, ok := expr.Left.(*parser.ASTVar)
		if !ok {
			panic("assignment left side must be var")
		}

		rhsResult := c.EmitExpr(expr.Right)

		c.Irs = append(c.Irs, Copy{Src: rhsResult, Dst: Var{Name: astVar.Ident}})
		return Var{Name: astVar.Ident}
	case *parser.ASTConstant:
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
		panic(fmt.Sprintf("unimplemented expr: %s", node.TokenLiteral()))
	}
}

func (c *TackyGen) Emit(op Instruction) {
	c.Irs = append(c.Irs, op)
}

func (c *TackyGen) PrettyPrint(program TackyProgram) {
	fmt.Println("\n// Function:", program.FnDef.Name)
	fmt.Println("// Three-address code:")
	fmt.Println()
	for _, instr := range program.FnDef.Instructions {
		instr.Ir()
	}
	fmt.Println()
}
