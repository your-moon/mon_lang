package semanticanalysis

import (
	"errors"
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/parser"
)

type Resolver struct {
	variableMap map[string]string
	tempCounter int
}

func New() Resolver {
	return Resolver{
		variableMap: make(map[string]string),
		tempCounter: 0,
	}
}

func (r *Resolver) makeNamedTemporary(name string) string {
	r.tempCounter++
	return fmt.Sprintf("%s_%d", name, r.tempCounter)
}

func (r *Resolver) Resolve(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	fndef, err := r.ResolveFnDef(program.FnDef)
	if err != nil {
		return nil, err
	}

	program.FnDef = fndef
	return program, nil
}

func (r *Resolver) ResolveFnDef(fndef parser.FNDef) (parser.FNDef, error) {
	// Create a new scope for the function
	oldMap := r.variableMap
	r.variableMap = make(map[string]string)
	defer func() { r.variableMap = oldMap }()

	for i, item := range fndef.BlockItems {
		blockitem, err := r.ResolveBlockItem(item)
		if err != nil {
			return parser.FNDef{}, err
		}
		fndef.BlockItems[i] = blockitem
	}
	return fndef, nil
}

func (r *Resolver) ResolveBlockItem(program parser.BlockItem) (parser.BlockItem, error) {
	switch nodetype := program.(type) {
	case parser.ASTStmt:
		stmt, err := r.ResolveStmt(nodetype)
		if err != nil {
			return stmt, err
		}
		return stmt, nil
	case *parser.Decl:
		decl, err := r.ResolveDecl(nodetype)
		if err != nil {
			return decl, err
		}
		return decl, nil
	}

	return nil, errors.New("unreachable point")
}

func (r *Resolver) ResolveStmt(program parser.ASTStmt) (parser.ASTStmt, error) {
	switch nodetype := program.(type) {
	case *parser.ASTReturnStmt:
		retval, err := r.ResolveExpr(nodetype.ReturnValue)
		if err != nil {
			return nodetype, err
		}
		nodetype.ReturnValue = retval
		return nodetype, nil
	case *parser.ExpressionStmt:
		expr, err := r.ResolveExpr(nodetype.Expression)
		if err != nil {
			return nodetype, err
		}
		nodetype.Expression = expr
		return nodetype, nil
	}
	return program, nil
}

func (r *Resolver) ResolveDecl(program *parser.Decl) (*parser.Decl, error) {
	// Check for duplicate declaration
	if _, exists := r.variableMap[program.Ident]; exists {
		return nil, fmt.Errorf("duplicate variable declaration: %s", program.Ident)
	}

	uniqueName := r.makeNamedTemporary(program.Ident)
	r.variableMap[program.Ident] = uniqueName

	if program.Expr != nil {
		resolved, err := r.ResolveExpr(program.Expr)
		if err != nil {
			return nil, err
		}
		program.Expr = resolved
	}

	program.Ident = uniqueName
	return program, nil
}

func (r *Resolver) ResolveExpr(program parser.ASTExpression) (parser.ASTExpression, error) {
	if program == nil {
		return nil, nil
	}

	switch nodetype := program.(type) {
	case *parser.ASTAssignment:
		left, ok := nodetype.Left.(parser.ASTVar)
		if !ok {
			return nil, fmt.Errorf("expected variable on left hand side of assignment statement, found %s", nodetype.Left.TokenLiteral())
		}

		resolvedLeft, err := r.ResolveExpr(left)
		if err != nil {
			return nil, err
		}

		resolvedRight, err := r.ResolveExpr(nodetype.Right)
		if err != nil {
			return nil, err
		}

		return &parser.ASTAssignment{
			Left:  resolvedLeft,
			Right: resolvedRight,
		}, nil

	case *parser.ASTVar:
		ident, ok := nodetype.Ident.(*parser.ASTIdent)
		if !ok {
			return nil, fmt.Errorf("expected identifier in variable reference, found %s", nodetype.Ident.TokenLiteral())
		}

		uniqueName, exists := r.variableMap[*ident.Token.Value]
		if !exists {
			return nil, fmt.Errorf("undeclared variable: %s", *ident.Token.Value)
		}

		newIdent := &parser.ASTIdent{Token: ident.Token}
		*newIdent.Token.Value = uniqueName
		return &parser.ASTVar{Ident: newIdent}, nil

	case *parser.ASTUnary:
		resolvedInner, err := r.ResolveExpr(nodetype.Inner)
		if err != nil {
			return nil, err
		}
		return &parser.ASTUnary{
			Inner: resolvedInner,
			Op:    nodetype.Op,
		}, nil

	case *parser.ASTBinary:
		resolvedLeft, err := r.ResolveExpr(nodetype.Left)
		if err != nil {
			return nil, err
		}

		resolvedRight, err := r.ResolveExpr(nodetype.Right)
		if err != nil {
			return nil, err
		}

		return &parser.ASTBinary{
			Left:  resolvedLeft,
			Right: resolvedRight,
			Op:    nodetype.Op,
		}, nil

	case *parser.ASTIntLitExpression:
		return nodetype, nil

	case *parser.ASTIdent:
		return nodetype, nil
	}

	return nil, fmt.Errorf("unknown expression type: %T", program)
}
