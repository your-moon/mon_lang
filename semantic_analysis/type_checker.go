package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/mtypes"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/util/unique"
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

func (c *TypeChecker) Check(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	for i, decl := range program.Decls {
		switch decltype := decl.(type) {
		case *parser.FnDecl:
			decl, err := c.checkFnDecl(decltype)
			if err != nil {
				program.Decls[i] = decl
			}
		default:
			program.Decls[i] = decltype
		}
	}
	return program, nil
}

func (c *TypeChecker) checkFnDecl(decl *parser.FnDecl) (*parser.FnDecl, error) {
	//TODO: ADD PARAM TYPES
	fnType := &mtypes.FnType{}
	hasBody := decl.Body != nil
	alreadyDefined := false

	prev := c.symbolTable.GetOptional(decl.Ident)
	//decl is in symbol table
	if prev != nil {
		_, ok := prev.Type.(*mtypes.FnType)
		if !ok {
			return nil, c.createSemanticError("функц %s-ийг өөр төрөлтэйгөөр дахин зарласан байна", decl.Token.Line, decl.Token.Span)
		}
		alreadyDefined = prev.IsDefined
		if alreadyDefined && hasBody {
			return nil, c.createSemanticError(fmt.Sprintf("функц '%s'-ийг дахин зарласан байна", decl.Ident), decl.Token.Line, decl.Token.Span)
		}
	} else {
		c.symbolTable.AddFn(fnType, decl.Ident, hasBody)
	}

	if hasBody {
		for _, param := range decl.Params {
			c.symbolTable.AddVar(&mtypes.IntType{}, param.Ident)
		}
		block, err := c.checkBlock(decl.Body)
		if err != nil {
			return nil, err
		}
		decl.Body = block
	}

	return decl, nil
}

func (c *TypeChecker) checkBlock(block *parser.ASTBlock) (*parser.ASTBlock, error) {
	for _, item := range block.BlockItems {
		switch item := item.(type) {
		case parser.ASTStmt:
			stmt, err := c.checkStmt(item)
			if err != nil {
				return nil, err
			}
			return &parser.ASTBlock{BlockItems: []parser.BlockItem{stmt}}, nil
		case parser.ASTDecl:
			decl, err := c.checkDecl(item)
			if err != nil {
				return nil, err
			}
			return &parser.ASTBlock{BlockItems: []parser.BlockItem{decl}}, nil
		default:
			return nil, c.createSemanticError("unreachable block", 0, lexer.Span{})
		}
	}
	return nil, nil
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
			dvar, err := c.checkExpr(typestmt.Var)
			if err != nil {
				return nil, err
			}
			typestmt.Var = dvar
		}

		expr, err := c.checkExpr(typestmt.Expr)
		if err != nil {
			return nil, err
		}
		typestmt.Expr = expr

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
		typestmt.Block = *block
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
		expr.Type = left.GetType()
		return expr, nil
	case *parser.ASTUnary:
		inner, err := c.checkExpr(expr.Inner)
		if err != nil {
			return nil, err
		}
		expr.Inner = inner
		expr.Type = &mtypes.IntType{}
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
			extype.Type = &mtypes.IntType{}
		case *parser.ASTConstLong:
			extype.Type = &mtypes.LongType{}
		}
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
		//TODO: HANDLE DIFF CASES AND AND,OR | ADD,OR,MUL,DIV,MOD
		common := c.getCommonType(left.GetType(), right.GetType())
		expr.Type = common
		return expr, nil
	case *parser.ASTVar:
		dVar := c.symbolTable.Get(expr.Ident)
		_, ok := dVar.Type.(*mtypes.FnType)
		if ok {
			return nil, c.createSemanticError("%s-нь хувьсагч байна", expr.Token.Line, expr.Token.Span)
		}
		expr.Type = dVar.Type
		return expr, nil
	case *parser.ASTFnCall:
		fn := c.symbolTable.Get(expr.Ident)
		if fn == nil {
			return nil, c.createSemanticError("функц %s-ийг дуудаж байна", expr.Token.Line, expr.Token.Span)
		}

		fType := fn.Type
		switch fType.(type) {
		case *mtypes.IntType:
			return nil, c.createSemanticError("хувьсагч %s-ийг дуудаж байна", expr.Token.Line, expr.Token.Span)
		case *mtypes.FnType:
			//TODO: ADD CHECK
			// if len(expr.Args) != len(fType.ParamTypes) {
			// 	return c.createSemanticError("функц %s-ийг дуудахдаа аргументын тоог буруу оруулсан байна", expr.Token.Line, expr.Token.Span)
			// }

			for i, arg := range expr.Args {
				checkedArg, err := c.checkExpr(arg)
				if err != nil {
					return nil, err
				}
				expr.Args[i] = checkedArg
			}
			return expr, nil
		default:
			return nil, c.createSemanticError("функц %s-ийг өөр төрөлтэйгөөр дуудасан байна", expr.Token.Line, expr.Token.Span)
		}
	}
	return nil, c.createSemanticError("unreachable expr", 0, lexer.Span{})
}

func (c *TypeChecker) getCommonType(t1, t2 mtypes.Type) mtypes.Type {
	if t1 == t2 {
		return t1
	}
	return &mtypes.LongType{}
}
