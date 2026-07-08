/*
 * mon_lang - tackygen
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

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
	currentRetType mtypes.Type // return type of the function being lowered
}

func NewTackyGen(uniquegen unique.UniqueGen, table *symbols.SymbolTable) TackyGen {
	return TackyGen{
		TempCount:   0,
		LabelCount:  0,
		UniqueGen:   uniquegen,
		SymbolTable: table,
	}
}

func (c *TackyGen) EmitTacky(node *parser.ASTProgram) TackyProgram {
	program := TackyProgram{}

	// Only malloc is implicit (used by шинэ keyword internally)
	program.ExternDefs = append(program.ExternDefs, TackyFn{Name: "malloc", IsExtern: true})

	for _, stmt := range node.Decls {
		switch stmttype := stmt.(type) {
		case *parser.FnDecl:
			if !stmttype.IsExtern {
				program.FnDefs = append(program.FnDefs, c.EmitTackyFn(stmttype))
			} else {
				program.ExternDefs = append(program.ExternDefs, c.EmitTackyFn(stmttype))
			}
		case *parser.ASTStructDecl:
			// layout only; nothing to emit
		case *parser.VarDecl:
			// Top-level variable declarations become global variables in .data section
			var initValue int64
			size := 4 // default Int32
			if stmttype.Expr != nil {
				if f, isFloat := stmttype.Expr.(*parser.ASTConstFloat); isFloat {
					initValue = (&mconstant.Float64{Value: f.Value}).GetValue() // bit pattern
					size = 8
				} else {
					v, ok := foldConstExpr(stmttype.Expr)
					if !ok {
						panic(fmt.Sprintf("глобал хувьсагч '%s'-ийн анхны утга тогтмол илэрхийлэл байх ёстой", stmttype.Ident))
					}
					initValue = v
					if _, isLong := stmttype.Expr.(*parser.ASTConstLong); isLong {
						size = 8
					}
				}
			}
			if stmttype.VarType != nil {
				if _, is64 := stmttype.VarType.(*mtypes.Int64Type); is64 {
					size = 8
				}
			}
			program.GlobalVars = append(program.GlobalVars, GlobalVar{
				Name:      stmttype.Ident,
				InitValue: initValue,
				Size:      size,
			})
		}
	}

	return program
}

// foldConstExpr evaluates a constant initializer expression at compile time.
// Globals live in .data, so their value must be known before the program
// runs; anything non-constant is rejected by the caller.
func foldConstExpr(e parser.ASTExpression) (int64, bool) {
	switch n := e.(type) {
	case *parser.ASTConstInt:
		return int64(n.Value), true
	case *parser.ASTConstLong:
		return n.Value, true
	case *parser.ASTUnary:
		v, ok := foldConstExpr(n.Inner)
		if !ok {
			return 0, false
		}
		switch n.Op {
		case lexer.MINUS:
			return -v, true
		case lexer.TILDE:
			return ^v, true
		case lexer.NOT:
			if v == 0 {
				return 1, true
			}
			return 0, true
		}
		return 0, false
	case *parser.ASTBinary:
		l, ok := foldConstExpr(n.Left)
		if !ok {
			return 0, false
		}
		r, ok := foldConstExpr(n.Right)
		if !ok {
			return 0, false
		}
		switch int(n.Op) {
		case parser.A_PLUS:
			return l + r, true
		case parser.A_MINUS:
			return l - r, true
		case parser.A_MUL:
			return l * r, true
		case parser.A_DIV:
			if r == 0 {
				return 0, false
			}
			return l / r, true
		case parser.A_MOD:
			if r == 0 {
				return 0, false
			}
			return l % r, true
		}
		return 0, false
	}
	return 0, false
}

func (c *TackyGen) EmitTackyFn(node *parser.FnDecl) TackyFn {
	irs := []Instruction{}
	c.currentRetType = node.ReturnType
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

// emitPointerArith lowers ptr±int (integer scaled by element size) and
// ptr-ptr (byte difference divided down to an element count).
func (c *TackyGen) emitPointerArith(expr *parser.ASTBinary, op TackyBinaryOp, v1, v2 TackyVal) ([]Instruction, TackyVal, bool) {
	lPtr, lIsPtr := expr.Left.GetType().(*mtypes.PointerType)
	_, rIsPtr := expr.Right.GetType().(*mtypes.PointerType)
	if !lIsPtr && !rIsPtr {
		return nil, nil, false
	}
	irs := []Instruction{}

	if lIsPtr && rIsPtr { // ptr - ptr
		diff := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, Binary{Op: Sub, Src1: v1, Src2: v2, Dst: diff})
		dst := c.makeTemp(&mtypes.Int64Type{})
		size := mtypes.SizeOf(lPtr.Referenced)
		irs = append(irs, Binary{Op: Div, Src1: diff, Src2: Constant{Value: &mconstant.Int64{Value: size}}, Dst: dst})
		return irs, dst, true
	}

	// normalize to ptr op int
	ptrVal, intVal, intType := v1, v2, expr.Right.GetType()
	if rIsPtr {
		ptrVal, intVal, intType = v2, v1, expr.Left.GetType()
	}
	ptrType := expr.Type.(*mtypes.PointerType)

	idx64, extIrs := c.maybeSignExtend(intVal, intType, &mtypes.Int64Type{})
	irs = append(irs, extIrs...)
	size := mtypes.SizeOf(ptrType.Referenced)
	offset := c.makeTemp(&mtypes.Int64Type{})
	irs = append(irs, Binary{Op: Mul, Src1: idx64, Src2: Constant{Value: &mconstant.Int64{Value: size}}, Dst: offset})
	dst := c.makeTemp(expr.Type)
	irs = append(irs, Binary{Op: op, Src1: ptrVal, Src2: offset, Dst: dst})
	return irs, dst, true
}

// emitMemberAddr computes &s.f: the struct reference plus the field's
// laid-out offset.
func (c *TackyGen) emitMemberAddr(m *parser.ASTMember) (TackyVal, []Instruction) {
	irs := []Instruction{}
	base, baseIrs := c.EmitExpr(m.Inner)
	irs = append(irs, baseIrs...)
	if m.Offset == 0 {
		return base, irs
	}
	addr := c.makeTemp(&mtypes.Int64Type{})
	irs = append(irs, Binary{Op: Add, Src1: base, Src2: Constant{Value: &mconstant.Int64{Value: m.Offset}}, Dst: addr})
	return addr, irs
}

// emitElementAddr computes &array[index]: base + index * sizeof(element).
func (c *TackyGen) emitElementAddr(idx *parser.ASTArrayIndex) (TackyVal, []Instruction) {
	irs := []Instruction{}
	basePtr, baseIrs := c.EmitExpr(idx.Array)
	irs = append(irs, baseIrs...)
	indexVal, indexIrs := c.EmitExpr(idx.Index)
	irs = append(irs, indexIrs...)
	idx64 := c.makeTemp(&mtypes.Int64Type{})
	irs = append(irs, SignExtend{Src: indexVal, Dst: idx64})
	elemSize := mtypes.SizeOf(idx.GetType())
	offset := c.makeTemp(&mtypes.Int64Type{})
	irs = append(irs, Binary{Op: Mul, Src1: idx64, Src2: Constant{Value: &mconstant.Int64{Value: elemSize}}, Dst: offset})
	addr := c.makeTemp(&mtypes.Int64Type{})
	irs = append(irs, Binary{Op: Add, Src1: basePtr, Src2: offset, Dst: addr})
	return addr, irs
}

func (c *TackyGen) EmitVarDecl(node *parser.VarDecl) []Instruction {
	irs := []Instruction{}

	// `зарла а: тоо[5];` - a sized array declaration allocates its backing
	// store immediately (heap-backed, like шинэ тоо[5])
	if arr, isArr := node.VarType.(*mtypes.ArrayType); isArr && arr.Size > 0 && node.Expr == nil {
		byteSize := arr.Size * mtypes.SizeOf(arr.ElementType)
		dst := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, FnCall{Name: "malloc",
			Args: []TackyVal{Constant{Value: &mconstant.Int64{Value: byteSize}}}, Dst: dst})
		irs = append(irs, Copy{Src: dst, Dst: Var{Name: node.Ident}})
		return irs
	}

	// `зарла ц: Цэг;` - struct declarations allocate their payload the
	// same way (structs are references)
	if st, isStruct := node.VarType.(*mtypes.StructType); isStruct && node.Expr == nil {
		dst := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, FnCall{Name: "malloc",
			Args: []TackyVal{Constant{Value: &mconstant.Int64{Value: st.Size}}}, Dst: dst})
		irs = append(irs, Copy{Src: dst, Dst: Var{Name: node.Ident}})
		return irs
	}

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
			// operand-defining irs must precede their consumers
			irs = append(irs, rangeStartIrs...)
			irs = append(irs, Copy{
				Src: rangeStart,
				Dst: Var{Name: loopVar.Ident},
			})

			irs = append(irs, Label{Ident: startLabel.Name})

			endVal, endValIrs := c.EmitExpr(rangeExpr.End)
			loopVarVal, loopVarValIrs := c.EmitExpr(ast.Var)
			temp := c.makeTemp(ast.Var.GetType())
			irs = append(irs, loopVarValIrs...)
			irs = append(irs, endValIrs...)
			irs = append(irs, Binary{
				Op:   LessThanEqual,
				Src1: loopVarVal,
				Src2: endVal,
				Dst:  temp,
			})
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

			irs = append(irs, c.EmitTackyBlock(ast.Body)...)

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
	case *parser.ASTMatch:
		irs := []Instruction{}
		endLabel := c.makeLabel("match_end")

		// evaluate the scrutinee exactly once
		val, valIrs := c.EmitExpr(ast.Scrutinee)
		irs = append(irs, valIrs...)
		scrutinee := c.makeTemp(ast.Scrutinee.GetType())
		irs = append(irs, Copy{Src: val, Dst: scrutinee})

		for _, arm := range ast.Arms {
			if arm.Pattern == nil { // wildcard: unconditional
				irs = append(irs, c.EmitTackyBlock(arm.Body)...)
				irs = append(irs, Jump{Target: endLabel.Name})
				continue
			}
			next := c.makeLabel("match_next")
			patVal, patIrs := c.EmitExpr(arm.Pattern)
			irs = append(irs, patIrs...)
			cond := c.makeTemp(&mtypes.Int32Type{})
			irs = append(irs, Binary{Op: Equal, Src1: scrutinee, Src2: patVal, Dst: cond})
			irs = append(irs, JumpIfZero{Val: cond, Ident: next.Name})
			irs = append(irs, c.EmitTackyBlock(arm.Body)...)
			irs = append(irs, Jump{Target: endLabel.Name})
			irs = append(irs, Label{Ident: next.Name})
		}

		irs = append(irs, Label{Ident: endLabel.Name})
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
			if c.currentRetType != nil {
				widened, widenIrs := c.maybeSignExtend(val, ast.ReturnValue.GetType(), c.currentRetType)
				irs = append(irs, widenIrs...)
				val = widened
			}
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
		var paramTypes []mtypes.Type
		if entry := c.SymbolTable.Get(expr.Ident); entry != nil {
			if fnType, ok := entry.Type.(*mtypes.FnType); ok {
				paramTypes = fnType.ParamTypes
			}
		}
		for i, arg := range expr.Args {
			argVal, argIrs := c.EmitExpr(arg)
			irs = append(irs, argIrs...)
			if i < len(paramTypes) {
				widened, widenIrs := c.maybeSignExtend(argVal, arg.GetType(), paramTypes[i])
				irs = append(irs, widenIrs...)
				argVal = widened
			}
			args = append(args, argVal)
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
		// Global mutable variables are accessed via Var - the emitter will convert to RipRelative
		return Var{Name: expr.Ident}, []Instruction{}
	case *parser.ASTNewArray:
		irs := []Instruction{}
		sizeVal, sizeIrs := c.EmitExpr(expr.Size)
		irs = append(irs, sizeIrs...)
		// Sign-extend size to 64-bit if needed
		size64 := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, SignExtend{Src: sizeVal, Dst: size64})
		elemSize := mtypes.SizeOf(expr.ElementType)
		byteSize := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, Binary{Op: Mul, Src1: size64, Src2: Constant{Value: &mconstant.Int64{Value: elemSize}}, Dst: byteSize})
		// Call malloc
		dst := c.makeTemp(&mtypes.Int64Type{})
		irs = append(irs, FnCall{Name: "malloc", Args: []TackyVal{byteSize}, Dst: dst})
		return dst, irs

	case *parser.ASTArrayIndex:
		irs := []Instruction{}
		addr, addrIrs := c.emitElementAddr(expr)
		irs = append(irs, addrIrs...)
		dst := c.makeTemp(expr.Type)
		irs = append(irs, Load{Src: addr, Dst: dst})
		return dst, irs

	case *parser.ASTMember:
		irs := []Instruction{}
		addr, addrIrs := c.emitMemberAddr(expr)
		irs = append(irs, addrIrs...)
		dst := c.makeTemp(expr.Type)
		irs = append(irs, Load{Src: addr, Dst: dst})
		return dst, irs

	case *parser.ASTAddrOf:
		irs := []Instruction{}
		switch inner := expr.Inner.(type) {
		case *parser.ASTVar:
			dst := c.makeTemp(expr.Type)
			irs = append(irs, GetAddress{Src: Var{Name: inner.Ident}, Dst: dst})
			return dst, irs
		case *parser.ASTDeref:
			// &*p is just p
			return c.EmitExpr(inner.Inner)
		case *parser.ASTArrayIndex:
			// &a[i] is the element address itself
			return c.emitElementAddr(inner)
		case *parser.ASTMember:
			return c.emitMemberAddr(inner)
		default:
			panic("addr-of: unsupported operand (semantic pass should reject)")
		}

	case *parser.ASTDeref:
		irs := []Instruction{}
		ptr, ptrIrs := c.EmitExpr(expr.Inner)
		irs = append(irs, ptrIrs...)
		dst := c.makeTemp(expr.Type)
		irs = append(irs, Load{Src: ptr, Dst: dst})
		return dst, irs

	case *parser.ASTAssignment:
		irs := []Instruction{}
		switch lhs := expr.Left.(type) {
		case *parser.ASTVar:
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			widened, widenIrs := c.maybeSignExtend(rhsResult, expr.Right.GetType(), lhs.GetType())
			irs = append(irs, widenIrs...)
			irs = append(irs, Copy{Src: widened, Dst: Var{Name: lhs.Ident}})
			return Var{Name: lhs.Ident}, irs
		case *parser.ASTArrayIndex:
			addr, addrIrs := c.emitElementAddr(lhs)
			irs = append(irs, addrIrs...)
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			widened, widenIrs := c.maybeSignExtend(rhsResult, expr.Right.GetType(), lhs.GetType())
			irs = append(irs, widenIrs...)
			irs = append(irs, Store{Src: widened, Dst: addr})
			return widened, irs
		case *parser.ASTMember:
			addr, addrIrs := c.emitMemberAddr(lhs)
			irs = append(irs, addrIrs...)
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			widened, widenIrs := c.maybeSignExtend(rhsResult, expr.Right.GetType(), lhs.GetType())
			irs = append(irs, widenIrs...)
			irs = append(irs, Store{Src: widened, Dst: addr})
			return widened, irs
		case *parser.ASTDeref:
			// *p = rhs: evaluate the pointer, then store through it
			ptr, ptrIrs := c.EmitExpr(lhs.Inner)
			irs = append(irs, ptrIrs...)
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			widened, widenIrs := c.maybeSignExtend(rhsResult, expr.Right.GetType(), lhs.GetType())
			irs = append(irs, widenIrs...)
			irs = append(irs, Store{Src: widened, Dst: ptr})
			return widened, irs
		default:
			panic("assignment left side must be var or array index")
		}
	case parser.ASTConst:
		if f, isFloat := expr.(*parser.ASTConstFloat); isFloat {
			return Constant{Value: &mconstant.Float64{Value: f.Value}}, []Instruction{}
		}
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

			if ptrIrs, dst, isPtrArith := c.emitPointerArith(expr, op, v1, v2); isPtrArith {
				irs = append(irs, ptrIrs...)
				return dst, irs
			}

			op = unsignedOp(op, expr.Left.GetType(), expr.Right.GetType())

			// relational ops on doubles: promote operands to double when
			// either side is one (expr.Type is int32 for comparisons)
			_, lf := expr.Left.GetType().(*mtypes.Float64Type)
			_, rf := expr.Right.GetType().(*mtypes.Float64Type)
			if lf || rf {
				v1c, v1cIrs := c.maybeSignExtend(v1, expr.Left.GetType(), &mtypes.Float64Type{})
				irs = append(irs, v1cIrs...)
				v1 = v1c
				v2c, v2cIrs := c.maybeSignExtend(v2, expr.Right.GetType(), &mtypes.Float64Type{})
				irs = append(irs, v2cIrs...)
				v2 = v2c
				dst := c.makeTemp(expr.Type)
				irs = append(irs, Binary{Op: op, Src1: v1, Src2: v2, Dst: dst})
				return dst, irs
			}

			// widen if mixed 32/64-bit
			commonIs64 := mtypes.IsInteger(expr.Type) && mtypes.SizeOf(expr.Type) == 8
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

// maybeSignExtend converts a value to the context's numeric type:
// 32->64 integer widening (sign- or zero-extension by signedness),
// int->double, and double->int truncation. Same-type pairs pass through.
func (c *TackyGen) maybeSignExtend(val TackyVal, fromType, toType mtypes.Type) (TackyVal, []Instruction) {
	if fromType == nil || toType == nil {
		return val, nil
	}
	_, fromF := fromType.(*mtypes.Float64Type)
	_, toF := toType.(*mtypes.Float64Type)

	switch {
	case fromF && toF:
		return val, nil
	case !fromF && toF:
		if !mtypes.IsInteger(fromType) {
			return val, nil
		}
		irs := []Instruction{}
		// widen to 64-bit first; cvtsi2sd takes a full register
		if mtypes.SizeOf(fromType) == 4 {
			wide := c.makeTemp(&mtypes.Int64Type{})
			if mtypes.IsUnsigned(fromType) {
				irs = append(irs, ZeroExtend{Src: val, Dst: wide})
			} else {
				irs = append(irs, SignExtend{Src: val, Dst: wide})
			}
			val = wide
		}
		dst := c.makeTemp(toType)
		irs = append(irs, IntToDouble{Src: val, Dst: dst})
		return dst, irs
	case fromF && !toF:
		if !mtypes.IsInteger(toType) {
			return val, nil
		}
		wide := c.makeTemp(&mtypes.Int64Type{})
		irs := []Instruction{DoubleToInt{Src: val, Dst: wide}}
		if mtypes.SizeOf(toType) == 4 {
			narrow := c.makeTemp(toType)
			irs = append(irs, Copy{Src: wide, Dst: narrow})
			return narrow, irs
		}
		return wide, irs
	}

	if !mtypes.IsInteger(fromType) || !mtypes.IsInteger(toType) {
		return val, nil
	}
	if mtypes.SizeOf(fromType) != 4 || mtypes.SizeOf(toType) != 8 {
		return val, nil
	}
	dst := c.makeTemp(toType)
	if mtypes.IsUnsigned(fromType) {
		return dst, []Instruction{ZeroExtend{Src: val, Dst: dst}}
	}
	return dst, []Instruction{SignExtend{Src: val, Dst: dst}}
}

// unsignedOp swaps division and ordered comparisons for their unsigned
// variants when the operands' common type is unsigned.
func unsignedOp(op TackyBinaryOp, l, r mtypes.Type) TackyBinaryOp {
	lu, ru := mtypes.IsUnsigned(l), mtypes.IsUnsigned(r)
	unsigned := (lu && mtypes.SizeOf(l) >= mtypes.SizeOf(r)) ||
		(ru && mtypes.SizeOf(r) >= mtypes.SizeOf(l))
	if !unsigned {
		return op
	}
	switch op {
	case Div:
		return UDiv
	case Modulo:
		return UModulo
	case LessThan:
		return ULessThan
	case LessThanEqual:
		return ULessThanEqual
	case GreaterThan:
		return UGreaterThan
	case GreaterThanEqual:
		return UGreaterThanEqual
	}
	return op
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
