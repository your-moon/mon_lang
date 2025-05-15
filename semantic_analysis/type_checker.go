package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/util/unique"
)

type TypeChecker struct {
	source      []int32
	uniqueGen   unique.UniqueGen
	symbolTable *SymbolTable
}

func NewTypeChecker(source []int32, uniqueGen unique.UniqueGen) *TypeChecker {
	return &TypeChecker{source: source, uniqueGen: uniqueGen, symbolTable: NewSymbolTable()}
}

func (c *TypeChecker) createSemanticError(message string, line int, span lexer.Span) error {
	return compilererrors.New(message, line, span, c.source, "Семантик шинжилгээ")
}

func (c *TypeChecker) Check(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	for _, decl := range program.Decls {
		switch decltype := decl.(type) {
		case *parser.FnDecl:
			if err := c.checkFnDecl(decltype); err != nil {
				return program, err
			}
		case *parser.ASTExtern:
			if err := c.checkExternFnDecl(decltype); err != nil {
				return program, err
			}
		}
	}
	return program, nil
}

func (c *TypeChecker) checkExternFnDecl(decl *parser.ASTExtern) error {
	fnType := &FnType{ParamCount: len(decl.FnDecl.Params)}
	c.symbolTable.AddFn(fnType, decl.FnDecl.Ident, false)
	return nil
}

func (c *TypeChecker) checkFnDecl(decl *parser.FnDecl) error {
	fnType := &FnType{ParamCount: len(decl.Params)}
	hasBody := decl.Body != nil
	alreadyDefined := false

	prev := c.symbolTable.GetOptional(decl.Ident)
	//decl is in symbol table
	if prev != nil {
		if !prev.Type.IsFn() {
			return c.createSemanticError("функц %s-ийг өөр төрөлтэйгөөр дахин зарласан байна", decl.Token.Line, decl.Token.Span)
		}
		alreadyDefined = prev.isDefined
		if alreadyDefined && hasBody {
			return c.createSemanticError(fmt.Sprintf("функц '%s'-ийг дахин зарласан байна", decl.Ident), decl.Token.Line, decl.Token.Span)
		}
	} else {
		c.symbolTable.AddFn(fnType, decl.Ident, hasBody)
	}

	if hasBody {
		for _, param := range decl.Params {
			c.symbolTable.AddVar(&IntType{}, param.Ident)
		}
		if err := c.checkBlock(decl.Body); err != nil {
			return err
		}
	}

	return nil
}

func (c *TypeChecker) checkBlock(block *parser.ASTBlock) error {
	for _, item := range block.BlockItems {
		switch item := item.(type) {
		case parser.ASTStmt:
			if err := c.checkStmt(item); err != nil {
				return err
			}
			return nil
		case parser.ASTDecl:
			if err := c.checkDecl(item); err != nil {
				return err
			}
			return nil
		default:
			return c.createSemanticError("unreachable block", 0, lexer.Span{})
		}
	}
	return nil
}

func (c *TypeChecker) checkStmt(stmt parser.ASTStmt) error {
	switch stmt := stmt.(type) {
	case *parser.ASTWhile:
		if stmt.Cond != nil {
			if err := c.checkExpr(stmt.Cond); err != nil {
				return err
			}
		}
		if err := c.checkBlock(&stmt.Body); err != nil {
			return err
		}
		return nil
	case *parser.ASTBreakStmt:
		return nil
	case *parser.ASTContinueStmt:
		return nil
	case *parser.ASTLoop:
		if stmt.Var != nil {
			if err := c.checkExpr(stmt.Var); err != nil {
				return err
			}
		}

		if err := c.checkExpr(stmt.Expr); err != nil {
			return err
		}

		if err := c.checkBlock(&stmt.Body); err != nil {
			return err
		}
		return nil
	case *parser.ASTCompoundStmt:
		if err := c.checkBlock(&stmt.Block); err != nil {
			return err
		}
		return nil
	case *parser.ASTIfStmt:
		if err := c.checkExpr(stmt.Cond); err != nil {
			return err
		}
		if err := c.checkStmt(stmt.Then); err != nil {
			return err
		}
		if stmt.Else != nil {
			if err := c.checkStmt(stmt.Else); err != nil {
				return err
			}
		}
		return nil
	case *parser.ExpressionStmt:
		if err := c.checkExpr(stmt.Expression); err != nil {
			return err
		}
		return nil
	case *parser.ASTReturnStmt:
		if stmt.ReturnValue != nil {
			if err := c.checkExpr(stmt.ReturnValue); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *TypeChecker) checkDecl(decl parser.ASTDecl) error {
	switch decl := decl.(type) {
	case *parser.VarDecl:
		c.symbolTable.AddVar(&IntType{}, decl.Ident)
		return nil
	case *parser.FnDecl:
		if err := c.checkFnDecl(decl); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown declaration type: %T", decl)
	}
}

func (c *TypeChecker) checkExpr(expr parser.ASTExpression) error {
	switch expr := expr.(type) {
	case *parser.ASTAssignment:
		if err := c.checkExpr(expr.Left); err != nil {
			return err
		}
		if err := c.checkExpr(expr.Right); err != nil {
			return err
		}
		return nil
	case *parser.ASTUnary:
		if err := c.checkExpr(expr.Inner); err != nil {
			return err
		}
		return nil
	case *parser.ASTConditional:
		if err := c.checkExpr(expr.Cond); err != nil {
			return err
		}
		if err := c.checkExpr(expr.Then); err != nil {
			return err
		}
		if err := c.checkExpr(expr.Else); err != nil {
			return err
		}
		return nil
	case parser.ASTConst:
		return nil
	case *parser.ASTBinary:
		if err := c.checkExpr(expr.Left); err != nil {
			return err
		}
		if err := c.checkExpr(expr.Right); err != nil {
			return err
		}
		return nil
	case *parser.ASTVar:
		if c.symbolTable.Get(expr.Ident).Type.IsFn() {
			return c.createSemanticError("%s-нь хувьсагч байна", expr.Token.Line, expr.Token.Span)
		}
		return nil
	case *parser.ASTFnCall:
		fn := c.symbolTable.Get(expr.Ident)
		if fn == nil {
			return c.createSemanticError("функц %s-ийг дуудаж байна", expr.Token.Line, expr.Token.Span)
		}

		fType := fn.Type
		switch fType := fType.(type) {
		case *IntType:
			return c.createSemanticError("хувьсагч %s-ийг дуудаж байна", expr.Token.Line, expr.Token.Span)
		case *FnType:
			if len(expr.Args) != fType.ParamCount {
				return c.createSemanticError("функц %s-ийг дуудахдаа аргументын тоог буруу оруулсан байна", expr.Token.Line, expr.Token.Span)
			}

			for _, arg := range expr.Args {
				if err := c.checkExpr(arg); err != nil {
					return err
				}
			}
			return nil
		default:
			return c.createSemanticError("функц %s-ийг өөр төрөлтэйгөөр дуудасан байна", expr.Token.Line, expr.Token.Span)
		}
	}
	return c.createSemanticError("unreachable expr", 0, lexer.Span{})
}
