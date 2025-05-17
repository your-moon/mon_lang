package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/util/unique"
)

const (
	ErrDuplicateVariable  = "хувьсагч '%s' нь давхардсан байна"
	ErrInvalidAssignment  = "хувьсагчид утга оноох үед зүүн талд хувьсагч байх ёстой, олдсон: '%s'"
	ErrUndeclaredVariable = "хувьсагч '%s'-г зарлаагүй байна"
	ErrUnknownExpression  = "үл мэдэгдэх илэрхийллийн төрөл: '%T'"
)

type IdMap map[string]VarEntry

type VarEntry struct {
	UniqueName       string
	fromCurrentScope bool
	hasLinkage       bool
	StorageClass     parser.StorageClass
}
type VariableMap struct {
	idMap map[string]VarEntry
}

type Resolver struct {
	// idMap       VariableMap
	tempCounter int
	source      []int32
	uniqueGen   unique.UniqueGen
	errors      []compilererrors.CompilerError
}

func NewResolver(source []int32, uniqueGen unique.UniqueGen) *Resolver {
	return &Resolver{
		tempCounter: 0,
		source:      source,
		uniqueGen:   uniqueGen,
	}
}

func (r *Resolver) makeNamedTemporary(name string) string {
	r.tempCounter++
	return fmt.Sprintf("%s_%d", name, r.tempCounter)
}

func (r *Resolver) resolveParams(params []parser.Param, innerMap map[string]VarEntry) (map[string]VarEntry, []parser.Param, error) {
	resolvedParams := []parser.Param{}
	for _, param := range params {
		if _, exists := innerMap[param.Ident]; exists {
			return nil, nil, r.createSemanticError(
				fmt.Sprintf(compilererrors.ErrDuplicateVariable, param.Ident),
				param.Token.Line,
				param.Token.Span,
			)
		}

		uniqueName := r.makeNamedTemporary(param.Ident)
		innerMap[param.Ident] = VarEntry{
			UniqueName:       uniqueName,
			fromCurrentScope: true,
			hasLinkage:       false,
		}

		resolvedParams = append(resolvedParams, parser.Param{
			Token: param.Token,
			Ident: uniqueName,
			Type:  param.Type,
		})
	}

	return innerMap, resolvedParams, nil
}

func (r *Resolver) Resolve(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	emptyMap := make(IdMap)
	for i, decl := range program.Decls {
		_, decl, err := r.ResolveDecl(decl, emptyMap)
		if err != nil {
			return nil, err
		}
		program.Decls[i] = decl
	}

	return program, nil
}

func (r *Resolver) ResolveDecl(decl parser.ASTDecl, innerMap IdMap) (IdMap, parser.ASTDecl, error) {
	switch declType := decl.(type) {
	case *parser.FnDecl:
		return r.ResolveFnDecl(declType, innerMap)
	case *parser.VarDecl:
		return r.ResolveFileScopeVarDecl(declType, innerMap)
	default:
		panic("unimplemented decl type on resolve")
	}
}

func (r *Resolver) ResolveFileScopeVarDecl(varDecl *parser.VarDecl, innerMap IdMap) (IdMap, *parser.VarDecl, error) {
	if _, exists := innerMap[varDecl.Ident]; exists && innerMap[varDecl.Ident].fromCurrentScope {
		return nil, nil, r.createSemanticError(
			fmt.Sprintf(compilererrors.ErrDuplicateVariable, varDecl.Ident),
			varDecl.Token.Line,
			varDecl.Token.Span,
		)
	}

	uniqueName := r.makeNamedTemporary(varDecl.Ident)
	innerMap[varDecl.Ident] = VarEntry{
		UniqueName:       uniqueName,
		fromCurrentScope: true,
		hasLinkage:       true,
	}

	return innerMap, varDecl, nil

}
func (r *Resolver) ResolveFnDecl(fndecl *parser.FnDecl, innerMap IdMap) (IdMap, *parser.FnDecl, error) {
	if found, exists := innerMap[fndecl.Ident]; exists && found.fromCurrentScope && !found.hasLinkage {
		return nil, nil, r.createSemanticError(
			fmt.Sprintf(compilererrors.ErrDuplicateFnDecl, fndecl.Ident),
			fndecl.Token.Line,
			fndecl.Token.Span,
		)
	}

	innerMap[fndecl.Ident] = VarEntry{
		UniqueName:       fndecl.Ident,
		fromCurrentScope: true,
		hasLinkage:       true,
	}

	newMap := r.copyIdMap(innerMap)

	solvedInnerMap, resolvedParams, err := r.resolveParams(fndecl.Params, newMap)
	if err != nil {
		return nil, nil, err
	}

	fndecl.Params = resolvedParams

	if fndecl.Body != nil {
		body, err := r.ResolveBlock(fndecl.Body, solvedInnerMap)
		if err != nil {
			return nil, nil, err
		}
		fndecl.Body = body
		return solvedInnerMap, fndecl, nil
	}

	return solvedInnerMap, fndecl, nil
}

// resolveLocalVarHelper is used to resolve local variables in a function
func (r *Resolver) resolveLocalVarHelper(innerMap map[string]VarEntry, varDecl *parser.VarDecl) (map[string]VarEntry, string, error) {
	if existVar, exists := innerMap[varDecl.Ident]; exists && innerMap[varDecl.Ident].fromCurrentScope {
		_, extrclass := existVar.StorageClass.(*parser.Extern)
		// end linkage bish bolon external bish uyd
		if !existVar.hasLinkage && !extrclass {
			return nil, "", r.createSemanticError(
				fmt.Sprintf(compilererrors.ErrDuplicateVariable, varDecl.Ident),
				varDecl.Token.Line,
				varDecl.Token.Span,
			)
		}
	}

	_, ok := varDecl.StorageClass.(*parser.Extern)
	if ok {
		innerMap[varDecl.Ident] = VarEntry{
			UniqueName:       varDecl.Ident,
			fromCurrentScope: true,
			hasLinkage:       false,
		}
	}

	uniqueName := r.makeNamedTemporary(varDecl.Ident)
	innerMap[varDecl.Ident] = VarEntry{
		UniqueName:       uniqueName,
		fromCurrentScope: true,
		hasLinkage:       false,
	}

	return innerMap, uniqueName, nil
}

func (r *Resolver) ResolveBlock(program *parser.ASTBlock, innerMap IdMap) (*parser.ASTBlock, error) {
	for i, item := range program.BlockItems {
		_, blockitem, err := r.ResolveBlockItem(item, innerMap)
		if err != nil {
			return nil, err
		}
		program.BlockItems[i] = blockitem
	}
	return program, nil
}

func (r *Resolver) ResolveBlockItem(program parser.BlockItem, innerMap IdMap) (IdMap, parser.BlockItem, error) {
	switch nodetype := program.(type) {
	case parser.ASTStmt:
		resolvedStmt, err := r.ResolveStmt(nodetype, innerMap)
		if err != nil {
			return nil, nil, err
		}
		return innerMap, resolvedStmt, nil
	case parser.ASTDecl:
		resolvedNewMap, decl, err := r.ResolveLocalDecl(nodetype, innerMap)
		if err != nil {
			return nil, nil, err
		}
		return resolvedNewMap, decl, nil
	}

	return innerMap, program, fmt.Errorf("unreachable point")
}

func (r *Resolver) ResolveStmt(program parser.ASTStmt, innerMap IdMap) (parser.ASTStmt, error) {
	switch nodetype := program.(type) {
	case *parser.ASTWhile:
		if nodetype.Cond != nil {
			resolvedCond, err := r.ResolveExpr(nodetype.Cond, innerMap)
			if err != nil {
				return nil, err
			}
			nodetype.Cond = resolvedCond
		}

		body, err := r.ResolveBlock(&nodetype.Body, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.Body = *body
		return nodetype, nil
	case *parser.ASTLoop:
		resolvedExpr, err := r.ResolveExpr(nodetype.Expr, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.Expr = resolvedExpr

		if nodetype.Var != nil {
			varExpr, ok := nodetype.Var.(*parser.ASTVar)
			if !ok {
				return nil, r.createSemanticError(
					"loop variable must be an identifier",
					nodetype.Token.Line,
					nodetype.Token.Span,
				)
			}

			uniqueName := r.makeNamedTemporary(varExpr.Ident)
			innerMap[varExpr.Ident] = VarEntry{
				UniqueName:       uniqueName,
				fromCurrentScope: true,
			}
			varExpr.Ident = uniqueName
			nodetype.Var = varExpr
		}

		body, err := r.ResolveBlock(&nodetype.Body, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.Body = *body

		return nodetype, nil
	case *parser.ASTCompoundStmt:
		newMap := r.copyIdMap(innerMap)
		body, err := r.ResolveBlock(&nodetype.Block, newMap)
		if err != nil {
			return nil, err
		}
		nodetype.Block = *body
		return nodetype, nil
	case *parser.ASTReturnStmt:
		resolvedReturnValue, err := r.ResolveExpr(nodetype.ReturnValue, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.ReturnValue = resolvedReturnValue
		return nodetype, nil
	case *parser.ASTIfStmt:
		resolvedCond, err := r.ResolveExpr(nodetype.Cond, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Cond = resolvedCond

		resolvedThen, err := r.ResolveStmt(nodetype.Then, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Then = resolvedThen
		resolvedElse, err := r.ResolveStmt(nodetype.Else, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Else = resolvedElse
		return nodetype, nil
	case *parser.ExpressionStmt:
		resolvedExpr, err := r.ResolveExpr(nodetype.Expression, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.Expression = resolvedExpr
		return nodetype, nil
	}
	return program, nil
}

func (r *Resolver) copyIdMap(mapToCopy map[string]VarEntry) map[string]VarEntry {
	newMap := make(map[string]VarEntry)

	for k, v := range mapToCopy {
		entry := v
		entry.fromCurrentScope = false
		newMap[k] = entry
	}

	return newMap
}

func (r *Resolver) createSemanticError(message string, line int, span lexer.Span) *compilererrors.CompilerError {
	return compilererrors.New(message, line, span, r.source, "Семантик шинжилгээ")
}

func (r *Resolver) ResolveLocalDecl(program parser.ASTDecl, innerMap IdMap) (IdMap, parser.ASTDecl, error) {
	switch decl := program.(type) {
	case *parser.FnDecl:
		return nil, nil, r.createSemanticError(
			fmt.Sprintf(compilererrors.ErrFnDeclCanNotBeInsideFnDecl, decl.Ident),
			decl.Token.Line,
			decl.Token.Span,
		)
	case *parser.VarDecl:
		return r.ResolveLocalVarDecl(decl, innerMap)
	}

	return innerMap, program, fmt.Errorf("unreachable point")
}

func (r *Resolver) ResolveLocalVarDecl(decl *parser.VarDecl, innerMap IdMap) (IdMap, *parser.VarDecl, error) {
	newMap, uniqueName, err := r.resolveLocalVarHelper(innerMap, decl)
	if err != nil {
		return nil, nil, err
	}

	decl.Ident = uniqueName
	if decl.Expr != nil {
		resolvedExpr, err := r.ResolveExpr(decl.Expr, newMap)
		if err != nil {
			return nil, nil, err
		}
		decl.Expr = resolvedExpr
	}

	return newMap, decl, nil
}

func (r *Resolver) ResolveExpr(program parser.ASTExpression, innerMap IdMap) (parser.ASTExpression, error) {
	if program == nil {
		return nil, nil
	}

	switch nodetype := program.(type) {
	case *parser.ASTFnCall:
		if _, exists := innerMap[nodetype.Ident]; !exists {
			return nil, r.createSemanticError(
				fmt.Sprintf(compilererrors.ErrNotDeclaredFnCall, nodetype.Ident),
				nodetype.Token.Line,
				nodetype.Token.Span,
			)
		}

		for i, arg := range nodetype.Args {
			resolvedArg, err := r.ResolveExpr(arg, innerMap)
			if err != nil {
				return nil, err
			}
			nodetype.Args[i] = resolvedArg
		}
		return nodetype, nil
	case *parser.ASTRangeExpr:
		resolvedStart, err := r.ResolveExpr(nodetype.Start, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Start = resolvedStart

		resolvedEnd, err := r.ResolveExpr(nodetype.End, innerMap)
		if err != nil {
			return nil, err
		}
		nodetype.End = resolvedEnd
		return nodetype, nil
	case *parser.ASTConditional:
		resolvedCond, err := r.ResolveExpr(nodetype.Cond, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Cond = resolvedCond

		resolvedThen, err := r.ResolveExpr(nodetype.Then, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Then = resolvedThen
		resolvedElse, err := r.ResolveExpr(nodetype.Else, innerMap)
		if err != nil {
			return nil, err
		}

		nodetype.Else = resolvedElse
		return nodetype, nil
	case *parser.ASTAssignment:
		left, ok := nodetype.Left.(*parser.ASTVar)
		if !ok {
			return nil, r.createSemanticError(
				fmt.Sprintf(compilererrors.ErrInvalidAssignment, nodetype.Left.TokenLiteral()),
				nodetype.Token.Line,
				nodetype.Token.Span,
			)
		}

		resolvedLeft, err := r.ResolveExpr(left, innerMap)
		if err != nil {
			return nil, err
		}

		resolvedRight, err := r.ResolveExpr(nodetype.Right, innerMap)
		if err != nil {
			return nil, err
		}

		return &parser.ASTAssignment{
			Left:  resolvedLeft,
			Right: resolvedRight,
		}, nil

	case *parser.ASTVar:
		uniqueName, exists := innerMap[nodetype.Ident]
		if !exists {
			return nil, r.createSemanticError(
				fmt.Sprintf(compilererrors.ErrUndeclaredVariable, nodetype.Ident),
				nodetype.Token.Line,
				nodetype.Token.Span,
			)
		}

		return &parser.ASTVar{
			Token: nodetype.Token,
			Ident: uniqueName.UniqueName,
		}, nil

	case *parser.ASTUnary:
		resolvedInner, err := r.ResolveExpr(nodetype.Inner, innerMap)
		if err != nil {
			return nil, err
		}
		return &parser.ASTUnary{
			Inner: resolvedInner,
			Op:    nodetype.Op,
		}, nil

	case *parser.ASTBinary:
		resolvedLeft, err := r.ResolveExpr(nodetype.Left, innerMap)
		if err != nil {
			return nil, err
		}

		resolvedRight, err := r.ResolveExpr(nodetype.Right, innerMap)
		if err != nil {
			return nil, err
		}

		return &parser.ASTBinary{
			Left:  resolvedLeft,
			Right: resolvedRight,
			Op:    nodetype.Op,
		}, nil

	case parser.ASTConst:
		return nodetype, nil

		// case *parser.ASTIdent:
		// 	return nodetype, nil
	}

	return nil, r.createSemanticError(
		fmt.Sprintf(compilererrors.ErrUnknownExpression, program),
		0,
		lexer.Span{Start: 0, End: 0},
	)
}
