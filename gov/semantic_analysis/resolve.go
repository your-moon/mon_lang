package semanticanalysis

import (
	"bytes"
	"fmt"
	"strings"

	compilererrors "github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

const (
	ErrDuplicateVariable  = "хувьсагч '%s' нь давхардсан байна"
	ErrInvalidAssignment  = "хувьсагчид утга оноох үед зүүн талд хувьсагч байх ёстой, олдсон: '%s'"
	ErrUndeclaredVariable = "хувьсагч '%s'-г зарлаагүй байна"
	ErrUnknownExpression  = "үл мэдэгдэх илэрхийллийн төрөл: '%T'"
)

// SemanticError is kept for backward compatibility
type SemanticError struct {
	Message string
	Line    int
	Span    lexer.Span
	Source  []int32
}

func (e *SemanticError) Error() string {
	var buf bytes.Buffer

	lineStart, lineEnd := e.findLineBoundaries()
	lineContent := string(e.Source[lineStart:lineEnd])
	pointer := e.createErrorPointer(lineStart)

	fmt.Fprintf(&buf, "%d-р мөрөнд алдаа гарлаа:\n", e.Line)
	fmt.Fprintf(&buf, "%s\n", lineContent)
	fmt.Fprintf(&buf, "%s\n", pointer)
	fmt.Fprintf(&buf, "Алдааны мессеж: %s\n", e.Message)

	return buf.String()
}

func (e *SemanticError) findLineBoundaries() (start, end int) {
	start = 0
	end = 0
	for i := 0; i < len(e.Source); i++ {
		if e.Source[i] == '\n' {
			if i < e.Span.Start {
				start = i + 1
			}
			if i >= e.Span.End {
				end = i
				break
			}
		}
	}
	if end == 0 {
		end = len(e.Source)
	}
	return start, end
}

func (e *SemanticError) createErrorPointer(lineStart int) string {
	return strings.Repeat(" ", e.Span.Start-lineStart) +
		strings.Repeat("^", e.Span.End-e.Span.Start)
}

type VarEntry struct {
	UniqueName       string
	fromCurrentScope bool
}
type VariableMap struct {
	variableMap map[string]VarEntry
}

type Resolver struct {
	variableMap VariableMap
	tempCounter int
	source      []int32
}

func New(source []int32) Resolver {
	return Resolver{
		variableMap: VariableMap{
			variableMap: make(map[string]VarEntry),
		},
		tempCounter: 0,
		source:      source,
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

	block, err := r.ResolveBlock(fndef.Block)
	if err != nil {
		return parser.FNDef{}, err
	}
	fndef.Block = block

	return fndef, nil
}

func (r *Resolver) ResolveBlock(program parser.ASTBlock) (parser.ASTBlock, error) {
	for i, item := range program.BlockItems {
		blockitem, err := r.ResolveBlockItem(item)
		if err != nil {
			return parser.ASTBlock{}, err
		}
		program.BlockItems[i] = blockitem
	}
	return program, nil
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

	return nil, fmt.Errorf("unreachable point")
}

func (r *Resolver) ResolveStmt(program parser.ASTStmt) (parser.ASTStmt, error) {
	switch nodetype := program.(type) {
	case *parser.ASTCompoundStmt:
		// create a new variable map that inherit the varMap
		r.scopeSwitch()
		stmt, err := r.ResolveBlock(nodetype.Block)
		if err != nil {
			return nil, err
		}
		nodetype.Block = stmt
		return nodetype, nil
	case *parser.ASTReturnStmt:
		retval, err := r.ResolveExpr(nodetype.ReturnValue)
		if err != nil {
			return nodetype, err
		}
		nodetype.ReturnValue = retval
		return nodetype, nil
	case *parser.ASTIfStmt:
		cond, err := r.ResolveExpr(nodetype.Cond)
		if err != nil {
			return nodetype, err
		}

		nodetype.Cond = cond

		then, err := r.ResolveStmt(nodetype.Then)
		if err != nil {
			return nodetype, err
		}

		nodetype.Then = then
		klse, err := r.ResolveStmt(nodetype.Else)
		if err != nil {
			return nodetype, err
		}

		nodetype.Else = klse
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

func (r *Resolver) scopeSwitch() {
	newMap := make(map[string]VarEntry)

	for k, v := range r.variableMap.variableMap {
		entry := v
		entry.fromCurrentScope = false
		newMap[k] = entry
	}

	r.variableMap.variableMap = newMap
}

func (r *Resolver) endScope() {
	newMap := make(map[string]VarEntry)

	for k, v := range r.variableMap.variableMap {
		entry := v
		entry.fromCurrentScope = false
		newMap[k] = entry
	}

	r.variableMap.variableMap = newMap
}

// createSemanticError creates both a SemanticError and a CompilerError for the given parameters
func (r *Resolver) createSemanticError(message string, line int, span lexer.Span) error {
	// Create a SemanticError for backward compatibility
	semanticErr := &SemanticError{
		Message: message,
		Line:    line,
		Span:    span,
		Source:  r.source,
	}

	// Also create a CompilerError for the new error reporting system
	_ = compilererrors.New(message, line, span, r.source, "Semantic Analysis")

	return semanticErr
}

func (r *Resolver) ResolveDecl(program *parser.Decl) (*parser.Decl, error) {
	// variable that in current scope and redeclared
	if _, exists := r.variableMap.variableMap[program.Ident]; exists && r.variableMap.variableMap[program.Ident].fromCurrentScope {
		// Use token information from the parser
		return nil, r.createSemanticError(
			fmt.Sprintf(compilererrors.ErrDuplicateVariable, program.Ident),
			program.Token.Line,
			program.Token.Span,
		)
	}

	// variable that in outer scope or not defined in var map that generate unique name and add to map
	if !r.variableMap.variableMap[program.Ident].fromCurrentScope {
		uniqueName := r.makeNamedTemporary(program.Ident)
		r.variableMap.variableMap[program.Ident] = VarEntry{
			UniqueName:       uniqueName,
			fromCurrentScope: true,
		}

		if program.Expr != nil {
			resolved, err := r.ResolveExpr(program.Expr)
			if err != nil {
				return nil, err
			}
			program.Expr = resolved
		}

		program.Ident = uniqueName

	}

	return program, nil
}

func (r *Resolver) ResolveExpr(program parser.ASTExpression) (parser.ASTExpression, error) {
	if program == nil {
		return nil, nil
	}

	switch nodetype := program.(type) {
	case *parser.ASTConditional:
		cond, err := r.ResolveExpr(nodetype.Cond)
		if err != nil {
			return nodetype, err
		}

		nodetype.Cond = cond

		then, err := r.ResolveExpr(nodetype.Then)
		if err != nil {
			return nodetype, err
		}

		nodetype.Then = then
		klse, err := r.ResolveExpr(nodetype.Else)
		if err != nil {
			return nodetype, err
		}

		nodetype.Else = klse
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
		uniqueName, exists := r.variableMap.variableMap[nodetype.Ident]
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

	case *parser.ASTConstant:
		return nodetype, nil

		// case *parser.ASTIdent:
		// 	return nodetype, nil
	}

	// For unknown expression types, we don't have token information
	// This is a fallback case that should rarely occur
	return nil, r.createSemanticError(
		fmt.Sprintf(compilererrors.ErrUnknownExpression, program),
		1,                            // Default line
		lexer.Span{Start: 0, End: 0}, // Default span
	)
}
