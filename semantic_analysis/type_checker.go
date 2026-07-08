/*
 * mon_lang - semantic_analysis
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mon_lang/errors"
	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mtypes"
	"github.com/your-moon/mon_lang/parser"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

type TypeChecker struct {
	source      []int32
	uniqueGen   unique.UniqueGen
	symbolTable *symbols.SymbolTable
}

func NewTypeChecker(source []int32, uniqueGen unique.UniqueGen, table *symbols.SymbolTable) *TypeChecker {
	return &TypeChecker{source: source, uniqueGen: uniqueGen, symbolTable: table}
}

func (c *TypeChecker) createSemanticError(message string, line int, span lexer.Span) error {
	return compilererrors.New(message, line, span, c.source, "Семантик шинжилгээ")
}

func (c *TypeChecker) CheckTopLevel(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	for i, decl := range program.Decls {
		switch decltype := decl.(type) {
		case *parser.FnDecl:
			decl, err := c.checkFnDecl(decltype)
			if err != nil {
				return nil, err
			}
			program.Decls[i] = decl
		case *parser.VarDecl:
			decl, err := c.checkDecl(decltype)
			if err != nil {
				return nil, err
			}
			program.Decls[i] = decl
		default:
			panic(fmt.Sprintf("unsupported top-level declaration: %T", decl))
		}
	}
	return program, nil
}

func (c *TypeChecker) checkFnDecl(decl *parser.FnDecl) (*parser.FnDecl, error) {
	paramTypes := make([]mtypes.Type, len(decl.Params))
	for i, param := range decl.Params {
		paramTypes[i] = param.Type
	}
	fnType := &mtypes.FnType{
		ParamTypes: paramTypes,
		RetType:    decl.ReturnType,
	}
	hasBody := decl.Body != nil && !decl.IsExtern
	alreadyDefined := false

	prev := c.symbolTable.GetOptional(decl.Ident)
	//decl is in symbol table
	if prev != nil {
		prevFn, ok := prev.Type.(*mtypes.FnType)
		if !ok {
			return nil, c.createSemanticError("функц %s-ийг өөр төрөлтэйгөөр дахин зарласан байна", decl.Token.Line, decl.Token.Span)
		}
		if !sameFnSignature(prevFn, fnType) {
			return nil, c.createSemanticError(fmt.Sprintf("функц '%s'-ийг өөр гарын үсэгтэйгээр дахин зарласан байна", decl.Ident), decl.Token.Line, decl.Token.Span)
		}
		alreadyDefined = prev.IsDefined
		if alreadyDefined && hasBody {
			return nil, c.createSemanticError(fmt.Sprintf("функц '%s'-ийг дахин зарласан байна", decl.Ident), decl.Token.Line, decl.Token.Span)
		}
		if hasBody {
			prev.IsDefined = true
		}
	} else {
		c.symbolTable.AddFn(fnType, decl.Ident, hasBody)
	}

	if hasBody {
		for _, param := range decl.Params {
			c.symbolTable.AddVar(param.Type, param.Ident)
		}
		block, err := c.checkBlock(decl.Body)
		if err != nil {
			return nil, err
		}
		decl.Body = block
	}

	return decl, nil
}

func isIntType(t mtypes.Type) bool {
	return mtypes.IsInteger(t)
}

// sameFnSignature reports whether two function types agree in arity and
// parameter/return types; redeclarations must match exactly.
func sameFnSignature(a, b *mtypes.FnType) bool {
	if len(a.ParamTypes) != len(b.ParamTypes) {
		return false
	}
	for i := range a.ParamTypes {
		if !mtypes.IsSameType(a.ParamTypes[i], b.ParamTypes[i]) {
			return false
		}
	}
	return mtypes.IsSameType(a.RetType, b.RetType)
}

func (c *TypeChecker) checkBlock(block *parser.ASTBlock) (*parser.ASTBlock, error) {
	for i, item := range block.BlockItems {
		blockItem, err := c.checkBlockItem(item)
		if err != nil {
			return nil, err
		}
		block.BlockItems[i] = blockItem
	}
	return block, nil
}

func (c *TypeChecker) checkBlockItem(blockItem parser.BlockItem) (parser.BlockItem, error) {
	switch blockItemType := blockItem.(type) {
	case parser.ASTStmt:
		stmt, err := c.checkStmt(blockItemType)
		if err != nil {
			return nil, err
		}
		return stmt, nil
	case parser.ASTDecl:
		decl, err := c.checkDecl(blockItemType)
		if err != nil {
			return nil, err
		}
		return decl, nil
	default:
		return nil, c.createSemanticError("unreachable block", 0, lexer.Span{})
	}
}

func (c *TypeChecker) checkStmt(stmt parser.ASTStmt) (parser.ASTStmt, error) {
	switch typestmt := stmt.(type) {
	case *parser.ASTWhile:
		if typestmt.Cond != nil {
			cond, err := c.checkExpr(typestmt.Cond)
			if err != nil {
				return nil, err
			}
			typestmt.Cond = cond
		}
		block, err := c.checkBlock(&typestmt.Body)
		if err != nil {
			return nil, err
		}
		typestmt.Body = *block
		return typestmt, nil
	case *parser.ASTBreakStmt:
		return typestmt, nil
	case *parser.ASTContinueStmt:
		return typestmt, nil
	case *parser.ASTLoop:
		if typestmt.Var != nil {
			// the loop variable is a fresh declaration, not a lookup
			if dvar, ok := typestmt.Var.(*parser.ASTVar); ok {
				c.symbolTable.AddVar(&mtypes.Int32Type{}, dvar.Ident)
				dvar.SetType(&mtypes.Int32Type{})
			} else {
				dvar, err := c.checkExpr(typestmt.Var)
				if err != nil {
					return nil, err
				}
				typestmt.Var = dvar
			}
		}

		// range bounds are plain int expressions; the range node itself
		// has no type of its own
		if rangeExpr, ok := typestmt.Expr.(*parser.ASTRangeExpr); ok {
			start, err := c.checkExpr(rangeExpr.Start)
			if err != nil {
				return nil, err
			}
			rangeExpr.Start = start
			end, err := c.checkExpr(rangeExpr.End)
			if err != nil {
				return nil, err
			}
			rangeExpr.End = end
		} else {
			expr, err := c.checkExpr(typestmt.Expr)
			if err != nil {
				return nil, err
			}
			typestmt.Expr = expr
		}

		block, err := c.checkBlock(&typestmt.Body)
		if err != nil {
			return nil, err
		}
		typestmt.Body = *block
		return typestmt, nil
	case *parser.ASTCompoundStmt:
		block, err := c.checkBlock(&typestmt.Block)
		if err != nil {
			return nil, err
		}
		if block != nil {
			typestmt.Block = *block
		}
		return typestmt, nil
	case *parser.ASTIfStmt:
		cond, err := c.checkExpr(typestmt.Cond)
		if err != nil {
			return nil, err
		}
		typestmt.Cond = cond

		then, err := c.checkStmt(typestmt.Then)
		if err != nil {
			return nil, err
		}
		typestmt.Then = then

		if typestmt.Else != nil {
			elseStmt, err := c.checkStmt(typestmt.Else)
			if err != nil {
				return nil, err
			}
			typestmt.Else = elseStmt
		}
		return typestmt, nil
	case *parser.ExpressionStmt:
		expr, err := c.checkExpr(typestmt.Expression)
		if err != nil {
			return nil, err
		}
		typestmt.Expression = expr
		return typestmt, nil
	case *parser.ASTReturnStmt:
		if typestmt.ReturnValue != nil {
			expr, err := c.checkExpr(typestmt.ReturnValue)
			if err != nil {
				return nil, err
			}
			typestmt.ReturnValue = expr
		}
		return typestmt, nil
	default:
		return nil, fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *TypeChecker) checkDecl(decl parser.ASTDecl) (parser.ASTDecl, error) {
	switch decl := decl.(type) {
	case *parser.VarDecl:
		c.symbolTable.AddVar(decl.VarType, decl.Ident)
		if decl.Expr != nil {
			exprCheck, err := c.checkExpr(decl.Expr)
			if err != nil {
				return nil, err
			}
			// keep the expression's true type: tackygen compares it
			// against the declared type and inserts SignExtend; relabeling
			// here would silently skip the widening (zero-extend bug)
			if decl.VarType != nil && !c.typesCompatible(exprCheck.GetType(), decl.VarType) {
				return nil, c.createSemanticError(
					fmt.Sprintf("'%s' төрлийн хувьсагчид '%s' төрлийн утга олгож болохгүй", c.typeName(decl.VarType), c.typeName(exprCheck.GetType())),
					decl.Token.Line, decl.Token.Span)
			}
			decl.Expr = exprCheck
		}
		return decl, nil
	case *parser.FnDecl:
		decl, err := c.checkFnDecl(decl)
		if err != nil {
			return nil, err
		}
		return decl, nil
	default:
		return nil, fmt.Errorf("unknown declaration type: %T", decl)
	}
}

func (c *TypeChecker) checkExpr(expr parser.ASTExpression) (parser.ASTExpression, error) {
	switch expr := expr.(type) {
	case *parser.ASTAssignment:
		left, err := c.checkExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.checkExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		expr.Left = left
		expr.Right = right
		if !c.typesCompatible(right.GetType(), left.GetType()) {
			return nil, c.createSemanticError(
				fmt.Sprintf("'%s' төрөлд '%s' төрлийн утга олгож болохгүй", c.typeName(left.GetType()), c.typeName(right.GetType())),
				expr.Token.Line, expr.Token.Span)
		}
		// For array index assignment, the type is the element type
		expr.Type = left.GetType()
		return expr, nil
	case *parser.ASTUnary:
		inner, err := c.checkExpr(expr.Inner)
		if err != nil {
			return nil, err
		}
		expr.Inner = inner
		expr.Type = &mtypes.Int32Type{}
		return expr, nil
	case *parser.ASTAddrOf:
		inner, err := c.checkExpr(expr.Inner)
		if err != nil {
			return nil, err
		}
		switch inner.(type) {
		case *parser.ASTVar, *parser.ASTDeref, *parser.ASTArrayIndex:
			// addressable
		default:
			return nil, c.createSemanticError("'&' зөвхөн хувьсагч, заагчийн утга эсвэл массивын элементэд хэрэглэнэ", expr.Token.Line, expr.Token.Span)
		}
		expr.Inner = inner
		expr.Type = &mtypes.PointerType{Referenced: inner.GetType()}
		return expr, nil
	case *parser.ASTDeref:
		inner, err := c.checkExpr(expr.Inner)
		if err != nil {
			return nil, err
		}
		ptr, ok := inner.GetType().(*mtypes.PointerType)
		if !ok {
			return nil, c.createSemanticError("'*' зөвхөн заагч төрөлд хэрэглэнэ", expr.Token.Line, expr.Token.Span)
		}
		expr.Inner = inner
		expr.Type = ptr.Referenced
		return expr, nil
	case *parser.ASTConditional:
		cond, err := c.checkExpr(expr.Cond)
		if err != nil {
			return nil, err
		}
		then, err := c.checkExpr(expr.Then)
		if err != nil {
			return nil, err
		}
		klse, err := c.checkExpr(expr.Else)
		if err != nil {
			return nil, err
		}
		expr.Cond = cond
		expr.Then = then
		expr.Else = klse
		common := c.getCommonType(then.GetType(), klse.GetType())
		expr.Type = common
		return expr, nil
	case parser.ASTConst:
		switch extype := expr.(type) {
		case *parser.ASTConstInt:
			extype.Type = &mtypes.Int32Type{}
		case *parser.ASTConstLong:
			extype.Type = &mtypes.Int64Type{}
		}
		return expr, nil
	case *parser.ASTStringExpression:
		expr.Type = &mtypes.StringType{}
		return expr, nil
	case *parser.ASTBinary:
		left, err := c.checkExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.checkExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		expr.Left = left
		expr.Right = right

		// pointer arithmetic: ptr±int (and int+ptr) stays a pointer;
		// ptr-ptr of the same type yields an element count
		lPtr, lIsPtr := left.GetType().(*mtypes.PointerType)
		rPtr, rIsPtr := right.GetType().(*mtypes.PointerType)
		if lIsPtr || rIsPtr {
			isAdd := expr.Op == parser.ASTBinOp(parser.A_PLUS)
			isSub := expr.Op == parser.ASTBinOp(parser.A_MINUS)
			switch {
			case lIsPtr && rIsPtr && isSub && mtypes.IsSameType(lPtr, rPtr):
				expr.Type = &mtypes.Int64Type{}
				return expr, nil
			case lIsPtr && !rIsPtr && (isAdd || isSub) && isIntType(right.GetType()):
				expr.Type = left.GetType()
				return expr, nil
			case !lIsPtr && rIsPtr && isAdd && isIntType(left.GetType()):
				expr.Type = right.GetType()
				return expr, nil
			default:
				return nil, c.createSemanticError("заагч дээр зөвхөн нэмэх, хасах үйлдэл хийж болно", expr.Token.Line, expr.Token.Span)
			}
		}

		//TODO: HANDLE DIFF CASES AND AND,OR | ADD,OR,MUL,DIV,MOD
		common := c.getCommonType(left.GetType(), right.GetType())
		expr.Type = common
		return expr, nil
	case *parser.ASTVar:
		dVar := c.symbolTable.Get(expr.Ident)
		if dVar == nil {
			return nil, c.createSemanticError(fmt.Sprintf("хувьсагч '%s' олдсонгүй", expr.Ident), expr.Token.Line, expr.Token.Span)
		}
		_, ok := dVar.Type.(*mtypes.FnType)
		if ok {
			return nil, c.createSemanticError(fmt.Sprintf("'%s' нь функц байна", expr.Ident), expr.Token.Line, expr.Token.Span)
		}
		expr.Type = dVar.Type
		return expr, nil
	case *parser.ASTArrayIndex:
		arr, err := c.checkExpr(expr.Array)
		if err != nil {
			return nil, err
		}
		idx, err := c.checkExpr(expr.Index)
		if err != nil {
			return nil, err
		}
		expr.Array = arr
		expr.Index = idx
		// Set type to element type
		if arrType, ok := arr.GetType().(*mtypes.ArrayType); ok {
			expr.Type = arrType.ElementType
		} else {
			expr.Type = &mtypes.Int32Type{}
		}
		return expr, nil

	case *parser.ASTNewArray:
		size, err := c.checkExpr(expr.Size)
		if err != nil {
			return nil, err
		}
		expr.Size = size
		expr.Type = &mtypes.ArrayType{ElementType: expr.ElementType}
		return expr, nil

	case *parser.ASTFnCall:
		fn := c.symbolTable.Get(expr.Ident)
		if fn == nil {
			return nil, c.createSemanticError("функц %s-ийг дуудаж байна", expr.Token.Line, expr.Token.Span)
		}
		// fType := fn.Type
		switch checkType := fn.Type.(type) {
		case *mtypes.Int32Type:
			return nil, c.createSemanticError("хувьсагч %s-ийг функц шиг болохгүй", expr.Token.Line, expr.Token.Span)
		case *mtypes.FnType:
			if len(checkType.ParamTypes) > 0 && len(expr.Args) != len(checkType.ParamTypes) {
				return nil, c.createSemanticError(
					fmt.Sprintf("'%s' функц %d аргумент авах ёстой, %d өгсөн байна", expr.Ident, len(checkType.ParamTypes), len(expr.Args)),
					expr.Token.Line, expr.Token.Span)
			}

			for i, arg := range expr.Args {
				checkedArg, err := c.checkExpr(arg)
				if err != nil {
					return nil, err
				}
				expr.Args[i] = checkedArg

				if i < len(checkType.ParamTypes) {
					argType := checkedArg.GetType()
					paramType := checkType.ParamTypes[i]
					if !c.typesCompatible(argType, paramType) {
						return nil, c.createSemanticError(
							fmt.Sprintf("'%s' функцийн %d-р аргумент '%s' төрөлтэй байх ёстой, '%s' төрөл өгсөн байна",
								expr.Ident, i+1, c.typeName(paramType), c.typeName(argType)),
							expr.Token.Line, expr.Token.Span)
					}
				}
			}
			expr.Type = checkType.RetType
			return expr, nil
		default:
			return nil, c.createSemanticError(fmt.Sprintf("unreachable expr %T", expr), expr.Token.Line, expr.Token.Span)
		}
	}
	return nil, c.createSemanticError(fmt.Sprintf("unreachable expr %T", expr), 0, lexer.Span{})
}

func (c *TypeChecker) typesCompatible(argType, paramType mtypes.Type) bool {
	// the integer family converts freely (widening/reinterpreting)
	if mtypes.IsInteger(argType) && mtypes.IsInteger(paramType) {
		return true
	}
	// everything else (pointers, strings, arrays) matches structurally -
	// a permissive default here would let integers flow into pointers
	return mtypes.IsSameType(argType, paramType)
}

func (c *TypeChecker) typeName(t mtypes.Type) string {
	switch t.(type) {
	case *mtypes.Int32Type:
		return "тоо"
	case *mtypes.Int64Type:
		return "тоо64"
	case *mtypes.StringType:
		return "мөр"
	case *mtypes.VoidType:
		return "хоосон"
	case *mtypes.PointerType:
		return "заагч"
	case *mtypes.UInt32Type:
		return "этоо"
	case *mtypes.UInt64Type:
		return "этоо64"
	case *mtypes.ArrayType:
		return "массив"
	default:
		return fmt.Sprintf("%T", t)
	}
}

// getCommonType implements the usual arithmetic conversions for the integer
// family: same type wins; otherwise the wider width wins; at equal width,
// unsigned wins (C semantics, Sandler ch12).
func (c *TypeChecker) getCommonType(t1, t2 mtypes.Type) mtypes.Type {
	if mtypes.IsSameType(t1, t2) {
		return t1
	}
	if !mtypes.IsInteger(t1) || !mtypes.IsInteger(t2) {
		return &mtypes.Int64Type{}
	}
	s1, s2 := mtypes.SizeOf(t1), mtypes.SizeOf(t2)
	if s1 == s2 {
		if mtypes.IsUnsigned(t1) {
			return t1
		}
		return t2
	}
	wider := t1
	if s2 > s1 {
		wider = t2
	}
	return wider
}
