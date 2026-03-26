package tackygen

import (
	"fmt"

	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mconstant"
	"github.com/your-moon/mon_lang/mtypes"
	"github.com/your-moon/mon_lang/parser"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

type TackyGen struct {
	TempCount       uint64
	LabelCount      uint64
	UniqueGen       unique.UniqueGen
	SymbolTable     *symbols.SymbolTable
	GlobalConstants map[string]mconstant.Const
	MutableGlobals  map[string]bool
}

func NewTackyGen(uniquegen unique.UniqueGen, table *symbols.SymbolTable) TackyGen {
	return TackyGen{
		TempCount:       0,
		LabelCount:      0,
		UniqueGen:       uniquegen,
		SymbolTable:     table,
		GlobalConstants: make(map[string]mconstant.Const),
		MutableGlobals:  make(map[string]bool),
	}
}

func (c *TackyGen) EmitTacky(node *parser.ASTProgram) TackyProgram {
	program := TackyProgram{}

	// Register implicit stdlib externs
	implicitExterns := []string{"хэвлэ", "мөр_хэвлэх", "унш", "унш32", "санамсаргүйТоо", "одоо", "malloc", "чөлөөлөх"}
	for _, name := range implicitExterns {
		program.ExternDefs = append(program.ExternDefs, TackyFn{Name: name, IsExtern: true})
	}

	for _, stmt := range node.Decls {
		switch stmttype := stmt.(type) {
		case *parser.FnDecl:
			if !stmttype.IsExtern {
				program.FnDefs = append(program.FnDefs, c.EmitTackyFn(stmttype))
			} else {
				program.ExternDefs = append(program.ExternDefs, c.EmitTackyFn(stmttype))
			}
		case *parser.VarDecl:
			// Top-level variable declarations with constant initializers become global constants
			if stmttype.Expr != nil {
				switch constExpr := stmttype.Expr.(type) {
				case *parser.ASTConstInt:
					c.GlobalConstants[stmttype.Ident] = &mconstant.Int32{Value: int32(constExpr.Value)}
				case *parser.ASTConstLong:
					c.GlobalConstants[stmttype.Ident] = &mconstant.Int64{Value: constExpr.Value}
				}
			}
		}
	}

	return program
}

func (c *TackyGen) EmitTackyFn(node *parser.FnDecl) TackyFn {
	irs := []Instruction{}
	if node.Body != nil {
		irs = append(irs, c.EmitTackyBlock(*node.Body)...)
	}

	if !c.isReturnExistsIn(node) {
		irs = append(irs, Return{Value: Constant{Value: &mconstant.IntZero}})
	}

	params := []TackyVal{}
	for _, param := range node.Params {
		params = append(params, c.EmitTackyParam(&param))
	}
	return TackyFn{Name: node.Ident, Instructions: irs, Params: params, Global: node.IsPublic, IsExtern: node.IsExtern}
}

func (c *TackyGen) EmitTackyBlock(node parser.ASTBlock) []Instruction {
	irs := []Instruction{}
	for _, stmt := range node.BlockItems {
		irs = append(irs, c.EmitTackyBlockItem(stmt)...)
	}
	return irs
}

func (c *TackyGen) EmitTackyBlockItem(node parser.BlockItem) []Instruction {
	switch ast := node.(type) {
	case parser.ASTStmt:
		return c.EmitTackyStmt(ast)
	case parser.ASTDecl:
		return c.EmitTackyLocalDecl(ast)
	}
	return []Instruction{}
}

func (c *TackyGen) EmitTackyLocalDecl(node parser.ASTDecl) []Instruction {
	switch ast := node.(type) {
	case *parser.FnDecl:
		panic("can't decl the fn in local")
	case *parser.VarDecl:
		return c.EmitVarDecl(ast)
	}
	return []Instruction{}
}

func (c *TackyGen) EmitVarDecl(node *parser.VarDecl) []Instruction {
	irs := []Instruction{}
	haveInit := node.Expr != nil
	if haveInit {
		rhsResult, rhsValIrs := c.EmitExpr(node.Expr)
		irs = append(irs, rhsValIrs...)
		// Sign-extend if assigning тоо to тоо64 variable
		if node.VarType != nil && node.Expr.GetType() != nil {
			extended, extIrs := c.maybeSignExtend(rhsResult, node.Expr.GetType(), node.VarType)
			irs = append(irs, extIrs...)
			rhsResult = extended
		}
		irs = append(irs, Copy{Src: rhsResult, Dst: Var{Name: node.Ident}})
	}
	return irs
}

func (c *TackyGen) EmitTackyStmt(node parser.ASTStmt) []Instruction {
	switch ast := node.(type) {
	case *parser.ASTWhile:
		irs := []Instruction{}
		startLabel := c.makeLabel("while_start")
		continueLabel := c.continueLabel(ast.Id)
		breakLabel := c.breakLabel(ast.Id)

		irs = append(irs, Label{Ident: startLabel.Name})

		condVal, condValIrs := c.EmitExpr(ast.Cond)
		irs = append(irs, condValIrs...)
		irs = append(irs, JumpIfZero{
			Val:   condVal,
			Ident: breakLabel.Name,
		})

		blockIrs := c.EmitTackyBlock(ast.Body)
		irs = append(irs, blockIrs...)

		irs = append(irs, Label{Ident: continueLabel.Name})
		irs = append(irs, Jump{Target: startLabel.Name})

		irs = append(irs, Label{Ident: breakLabel.Name})
		return irs
	case *parser.ASTLoop:
		irs := []Instruction{}
		startLabel := c.makeLabel("loop")
		continueLabel := c.continueLabel(ast.Id)
		breakLabel := c.breakLabel(ast.Id)
		ast.Id = breakLabel.Name

		rangeExpr, ok := ast.Expr.(*parser.ASTRangeExpr)
		if ok {
			rangeStart, rangeStartIrs := c.EmitExpr(rangeExpr.Start)
			loopVar, ok := ast.Var.(*parser.ASTVar)
			if !ok {
				panic("Expected ASTVar for loop variable")
			}
			irs = append(irs, Copy{
				Src: rangeStart,
				Dst: Var{Name: loopVar.Ident},
			})
			irs = append(irs, rangeStartIrs...)

			irs = append(irs, Label{Ident: startLabel.Name})

			endVal, endValIrs := c.EmitExpr(rangeExpr.End)
			loopVarVal, loopVarValIrs := c.EmitExpr(ast.Var)
			temp := c.makeTemp(ast.Var.GetType())
			irs = append(irs, Binary{
				Op:   LessThanEqual,
				Src1: loopVarVal,
				Src2: endVal,
				Dst:  temp,
			})
			irs = append(irs, loopVarValIrs...)
			irs = append(irs, endValIrs...)
			irs = append(irs, JumpIfZero{
				Val:   temp,
				Ident: breakLabel.Name,
			})

			blockIrs := c.EmitTackyBlock(ast.Body)
			irs = append(irs, blockIrs...)

			irs = append(irs, Label{Ident: continueLabel.Name})

			irs = append(irs, Binary{
				Op:   Add,
				Src1: Var{Name: loopVar.Ident},
				Src2: Constant{Value: &mconstant.IntOne},
				Dst:  Var{Name: loopVar.Ident},
			})

			irs = append(irs, Jump{Target: startLabel.Name})
		} else {
			irs = append(irs, Label{Ident: startLabel.Name})

			condVal, condValIrs := c.EmitExpr(ast.Expr)
			irs = append(irs, condValIrs...)
			irs = append(irs, JumpIfZero{
				Val:   condVal,
				Ident: breakLabel.Name,
			})

			c.EmitTackyBlock(ast.Body)

			irs = append(irs, Label{Ident: continueLabel.Name})
			irs = append(irs, Jump{Target: startLabel.Name})
		}

		// End label
		irs = append(irs, Label{Ident: breakLabel.Name})
		return irs
	case *parser.ASTBreakStmt:
		irs := []Instruction{}
		irs = append(irs, Jump{Target: c.breakLabel(ast.Id).Name})
		return irs
	case *parser.ASTContinueStmt:
		irs := []Instruction{}
		irs = append(irs, Jump{Target: c.continueLabel(ast.Id).Name})
		return irs
	case *parser.ASTCompoundStmt:
		irs := []Instruction{}
		irs = append(irs, c.EmitTackyBlock(ast.Block)...)
		return irs
	case *parser.ASTIfStmt:
		irs := []Instruction{}
		// no else clause
		if ast.Else == nil {
			endLabel := c.makeLabel("if_end")
			//instruction of cond
			// c = result of cond
			evalCond, condIrs := c.EmitExpr(ast.Cond)
			irs = append(irs, condIrs...)
			// jumpifzero(c, end)
			jmpifzero := JumpIfZero{Val: evalCond, Ident: endLabel.Name}
			irs = append(irs, jmpifzero)

			// instructions of body
			irs = append(irs, c.EmitTackyStmt(ast.Then)...)
			// label(end)
			irs = append(irs, Label{Ident: endLabel.Name})
			return irs
		} else {
			elseLabel := c.makeLabel("else")
			endLabel := c.makeLabel("")
			evalCond, condIrs := c.EmitExpr(ast.Cond)
			irs = append(irs, condIrs...)
			jmpifzero := JumpIfZero{Val: evalCond, Ident: elseLabel.Name}
			irs = append(irs, jmpifzero)
			irs = append(irs, c.EmitTackyStmt(ast.Then)...)
			irs = append(irs, Jump{Target: endLabel.Name})
			irs = append(irs, Label{Ident: elseLabel.Name})
			irs = append(irs, c.EmitTackyStmt(ast.Else)...)
			irs = append(irs, Label{Ident: endLabel.Name})
			return irs
		}
	case *parser.ASTReturnStmt:
		irs := []Instruction{}
		if ast.ReturnValue != nil {
			val, valIrs := c.EmitExpr(ast.ReturnValue)
			irs = append(irs, valIrs...)
			irs = append(irs, Return{Value: val})
		} else {
			irs = append(irs, Return{Value: Constant{Value: &mconstant.IntZero}})
		}
		return irs
	case *parser.ExpressionStmt:
		irs := []Instruction{}
		_, valIrs := c.EmitExpr(ast.Expression)
		irs = append(irs, valIrs...)
		return irs
	}

	return []Instruction{}
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
	if op == parser.ASTBinOp(parser.A_MOD) {
		return Modulo, nil
	}
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

func (c *TackyGen) EmitAndExpr(expr *parser.ASTBinary) (TackyVal, []Instruction) {
	irs := []Instruction{}
	falseLabel := c.makeLabel("and_false")
	endLabel := c.makeLabel("and_end")
	dst := c.makeTemp(&mtypes.Int32Type{})

	v1, v1Irs := c.EmitExpr(expr.Left)
	irs = append(irs, v1Irs...)
	irs = append(irs, JumpIfZero{
		Val:   v1,
		Ident: falseLabel.Name,
	})
	v2, v2Irs := c.EmitExpr(expr.Right)
	irs = append(irs, v2Irs...)
	irs = append(irs, []Instruction{
		JumpIfZero{
			Val:   v2,
			Ident: falseLabel.Name,
		},
		Copy{
			Src: Constant{Value: &mconstant.IntOne},
			Dst: dst,
		},
		Jump{
			Target: endLabel.Name,
		},
		Label{
			Ident: falseLabel.Name,
		},
		Copy{
			Src: Constant{Value: &mconstant.IntZero},
			Dst: dst,
		},
		Label{
			Ident: endLabel.Name,
		},
	}...)

	return dst, irs
}

func (c *TackyGen) EmitOrExpr(expr *parser.ASTBinary) (TackyVal, []Instruction) {
	irs := []Instruction{}
	trueLabel := c.makeLabel("or_true")
	endLabel := c.makeLabel("or_end")
	dst := c.makeTemp(&mtypes.Int32Type{})

	left, leftIrs := c.EmitExpr(expr.Left)
	irs = append(irs, leftIrs...)
	irs = append(irs, JumpIfNotZero{
		Val:   left,
		Ident: trueLabel.Name,
	})
	right, rightIrs := c.EmitExpr(expr.Right)
	irs = append(irs, rightIrs...)
	irs = append(irs, []Instruction{
		JumpIfNotZero{
			Val:   right,
			Ident: trueLabel.Name,
		},
		Copy{
			Src: Constant{Value: &mconstant.IntZero},
			Dst: dst,
		},
		Jump{
			Target: endLabel.Name,
		},
		Label{
			Ident: trueLabel.Name,
		},
		Copy{
			Src: Constant{Value: &mconstant.IntOne},
			Dst: dst,
		},
		Label{
			Ident: endLabel.Name,
		},
	}...)

	return dst, irs
}

func (c *TackyGen) EmitExpr(node parser.ASTExpression) (TackyVal, []Instruction) {
	switch expr := node.(type) {
	case *parser.ASTFnCall:
		irs := []Instruction{}
		dst := c.makeTemp(expr.Type)
		args := []TackyVal{}
		for _, arg := range expr.Args {
			argVal, argIrs := c.EmitExpr(arg)
			args = append(args, argVal)
			irs = append(irs, argIrs...)
		}
		irs = append(irs, FnCall{Name: expr.Ident, Dst: dst, Args: args})
		return dst, irs
	case *parser.ASTRangeExpr:
		irs := []Instruction{}
		start, startIrs := c.EmitExpr(expr.Start)
		irs = append(irs, startIrs...)
		end, endIrs := c.EmitExpr(expr.End)
		irs = append(irs, endIrs...)
		dst := c.makeTemp(&mtypes.Int32Type{})
		irs = append(irs, Binary{
			Op:   Sub,
			Src1: end,
			Src2: start,
			Dst:  dst,
		})
		irs = append(irs, Binary{
			Op:   Add,
			Src1: dst,
			Src2: Constant{Value: &mconstant.IntOne},
			Dst:  dst,
		})
		return dst, irs

	case *parser.ASTConditional:
		irs := []Instruction{}
		endLabel := c.makeLabel("conditional_end")
		elseLabel := c.makeLabel("conditional_else")
		dst := c.makeTemp(expr.Type)

		condEval, condIrs := c.EmitExpr(expr.Cond)
		irs = append(irs, condIrs...)
		irs = append(irs, JumpIfZero{Val: condEval, Ident: elseLabel.Name})

		evalThen, thenIrs := c.EmitExpr(expr.Then)
		irs = append(irs, thenIrs...)
		irs = append(irs, Copy{Src: evalThen, Dst: dst})
		irs = append(irs, Jump{Target: endLabel.Name})
		irs = append(irs, Label{Ident: elseLabel.Name})

		evalElse, elseIrs := c.EmitExpr(expr.Else)
		irs = append(irs, elseIrs...)
		irs = append(irs, Copy{Src: evalElse, Dst: dst})
		irs = append(irs, Label{Ident: endLabel.Name})
		return dst, irs
	case *parser.ASTVar:
		// Check if this is a global constant
		if constVal, ok := c.GlobalConstants[expr.Ident]; ok {
			return Constant{Value: constVal}, []Instruction{}
		}
		return Var{Name: expr.Ident}, []Instruction{}
	case *parser.ASTNewArray:
		irs := []Instruction{}
		sizeVal, sizeIrs := c.EmitExpr(expr.Size)
		irs = append(irs, sizeIrs...)
		// Sign-extend size to 64-bit if needed
		size64 := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, SignExtend{Src: sizeVal, Dst: size64})
		// byteSize = size * 4 (element size for Int32)
		byteSize := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, Binary{Op: Mul, Src1: size64, Src2: Constant{Value: &mconstant.Int64{Value: 4}}, Dst: byteSize})
		// Call malloc
		dst := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, FnCall{Name: "malloc", Args: []TackyVal{byteSize}, Dst: dst})
		return dst, irs

	case *parser.ASTArrayIndex:
		irs := []Instruction{}
		basePtr, baseIrs := c.EmitExpr(expr.Array)
		irs = append(irs, baseIrs...)
		indexVal, indexIrs := c.EmitExpr(expr.Index)
		irs = append(irs, indexIrs...)
		// Sign-extend index to 64-bit
		idx64 := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, SignExtend{Src: indexVal, Dst: idx64})
		// offset = index * 4
		offset := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, Binary{Op: Mul, Src1: idx64, Src2: Constant{Value: &mconstant.Int64{Value: 4}}, Dst: offset})
		// addr = base + offset
		addr := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, Binary{Op: Add, Src1: basePtr, Src2: offset, Dst: addr})
		// Load value from memory
		dst := c.makeTemp(expr.Type)
		irs = append(irs, Load{Src: addr, Dst: dst})
		return dst, irs

	case *parser.ASTAssignment:
		irs := []Instruction{}
		switch lhs := expr.Left.(type) {
		case *parser.ASTVar:
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			irs = append(irs, Copy{Src: rhsResult, Dst: Var{Name: lhs.Ident}})
			return Var{Name: lhs.Ident}, irs
		case *parser.ASTArrayIndex:
			// Compute address
			basePtr, baseIrs := c.EmitExpr(lhs.Array)
			irs = append(irs, baseIrs...)
			indexVal, indexIrs := c.EmitExpr(lhs.Index)
			irs = append(irs, indexIrs...)
			// Sign-extend index to 64-bit
			idx64 := c.makeTemp(&mtypes.Int64Type{})
			irs = append(irs, SignExtend{Src: indexVal, Dst: idx64})
			offset := c.makeTemp(&mtypes.Int64Type{})
			irs = append(irs, Binary{Op: Mul, Src1: idx64, Src2: Constant{Value: &mconstant.Int64{Value: 4}}, Dst: offset})
			addr := c.makeTemp(&mtypes.Int64Type{})
			irs = append(irs, Binary{Op: Add, Src1: basePtr, Src2: offset, Dst: addr})
			// Evaluate RHS
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			// Store
			irs = append(irs, Store{Src: rhsResult, Dst: addr})
			return rhsResult, irs
		default:
			panic("assignment left side must be var or array index")
		}
	case parser.ASTConst:
		exprType := expr.GetType()
		switch consttype := expr.(type) {
		case *parser.ASTConstInt:
			//this means the var decl's lhs is long
			_, isParentLong := exprType.(*mtypes.Int64Type)
			if isParentLong {
				return Constant{Value: &mconstant.Int64{Value: int64(consttype.Value)}}, []Instruction{}
			}
			return Constant{Value: &mconstant.Int32{Value: int32(consttype.Value)}}, []Instruction{}
		case *parser.ASTConstLong:
			return Constant{Value: &mconstant.Int64{Value: consttype.Value}}, []Instruction{}
		default:
			panic("unimplemented type")

		}
	case *parser.ASTStringExpression:
		// Handle string literals
		return StringConstant{Value: expr.Value}, []Instruction{}

	case *parser.ASTUnary:
		irs := []Instruction{}
		src, srcIrs := c.EmitExpr(expr.Inner)
		irs = append(irs, srcIrs...)
		dst := c.makeTemp(expr.Type)

		op, err := ToUnaryTackyOp(expr.Op)
		if err != nil {
			panic(err)
		}

		instr := Unary{
			Op:  op,
			Src: src,
			Dst: dst,
		}
		irs = append(irs, instr)
		return dst, irs
	case *parser.ASTBinary:
		irs := []Instruction{}
		if expr.Op == parser.ASTBinOp(parser.A_AND) {
			return c.EmitAndExpr(expr)
		} else if expr.Op == parser.ASTBinOp(parser.A_OR) {
			return c.EmitOrExpr(expr)
		} else {
			op, err := ToTackyOp(expr.Op)
			if err != nil {
				panic(err)
			}

			v1, v1Irs := c.EmitExpr(expr.Left)
			irs = append(irs, v1Irs...)
			v2, v2Irs := c.EmitExpr(expr.Right)
			irs = append(irs, v2Irs...)

			// Sign-extend if mixed 32/64-bit
			_, commonIs64 := expr.Type.(*mtypes.Int64Type)
			if commonIs64 {
				v1ext, v1extIrs := c.maybeSignExtend(v1, expr.Left.GetType(), expr.Type)
				irs = append(irs, v1extIrs...)
				v1 = v1ext
				v2ext, v2extIrs := c.maybeSignExtend(v2, expr.Right.GetType(), expr.Type)
				irs = append(irs, v2extIrs...)
				v2 = v2ext
			}

			dst := c.makeTemp(expr.Type)
			instr := Binary{
				Op:   op,
				Src1: v1,
				Src2: v2,
				Dst:  dst,
			}
			irs = append(irs, instr)
			return dst, irs
		}
	default:
		panic(fmt.Sprintf("unimplemented expr: %s", node.TokenLiteral()))
	}
}

func (c *TackyGen) EmitTackyParam(node *parser.Param) TackyVal {
	return Var{Name: node.Ident}
}

func (c *TackyGen) isReturnExistsIn(node *parser.FnDecl) bool {
	if node.Body == nil {
		return false
	}
	for _, stmt := range node.Body.BlockItems {
		switch stmt.(type) {
		case *parser.ASTReturnStmt:
			return true
		}
	}
	return false
}

func (c *TackyGen) PrettyPrint(program TackyProgram) {
	for _, fn := range program.FnDefs {
		fmt.Println(fn.Name + ":")
		for _, instr := range fn.Instructions {
			instr.Ir()
		}
		fmt.Println()
	}
}

func (c *TackyGen) maybeSignExtend(val TackyVal, fromType, toType mtypes.Type) (TackyVal, []Instruction) {
	_, fromIs32 := fromType.(*mtypes.Int32Type)
	_, toIs64 := toType.(*mtypes.Int64Type)
	if fromIs32 && toIs64 {
		dst := c.makeTemp(&mtypes.Int64Type{})
		return dst, []Instruction{SignExtend{Src: val, Dst: dst}}
	}
	return val, nil
}

func (c *TackyGen) makeTemp(mtype mtypes.Type) Var {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	c.SymbolTable.AddVar(mtype, temp)
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
