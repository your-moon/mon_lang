package semanticanalysis

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

// Scope represents a lexical scope for variable declarations
type Scope struct {
	variables map[string]*SymbolInfo
	parent    *Scope
}

// SymbolInfo stores information about a declared variable
type SymbolInfo struct {
	Type        lexer.TokenType
	Declared    bool
	Initialized bool
}

// NewScope creates a new scope with an optional parent scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		variables: make(map[string]*SymbolInfo),
		parent:    parent,
	}
}

// Declare adds a variable to the current scope
func (s *Scope) Declare(name string, typ lexer.TokenType) error {
	if _, exists := s.variables[name]; exists {
		return fmt.Errorf("variable %s already declared in this scope", name)
	}
	s.variables[name] = &SymbolInfo{
		Type:        typ,
		Declared:    true,
		Initialized: false,
	}
	return nil
}

// Lookup searches for a variable in the current scope and parent scopes
func (s *Scope) Lookup(name string) (*SymbolInfo, bool) {
	if info, exists := s.variables[name]; exists {
		return info, true
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil, false
}

// Resolver performs semantic analysis on the AST
type Resolver struct {
	parser       *parser.Parser
	currentScope *Scope
	errors       []string
}

func NewResolver(parser *parser.Parser) *Resolver {
	return &Resolver{
		parser:       parser,
		currentScope: NewScope(nil),
		errors:       make([]string, 0),
	}
}

func (r *Resolver) pushScope() {
	r.currentScope = NewScope(r.currentScope)
}

func (r *Resolver) popScope() {
	if r.currentScope.parent != nil {
		r.currentScope = r.currentScope.parent
	}
}

func (r *Resolver) addError(msg string) {
	r.errors = append(r.errors, msg)
}

func (r *Resolver) Resolve() error {
	program, err := r.parser.ParseProgram()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	r.VisitProgram(program)

	if len(r.errors) > 0 {
		return fmt.Errorf("semantic errors: %v", r.errors)
	}

	return nil
}

// VisitProgram visits the program node
func (r *Resolver) VisitProgram(program *parser.ASTProgram) {
	r.pushScope() // Create global scope
	r.VisitFnDef(&program.FnDef)
	r.popScope()
}

func (r *Resolver) VisitFnDef(fn *parser.ASTFNDef) {
	r.pushScope() // Create function scope
	r.VisitBlockItems(fn.BlockItem)
	r.popScope()
}

func (r *Resolver) VisitBlockItems(blockItems []parser.BlockItem) {
	for _, item := range blockItems {
		r.VisitBlockItem(item)
	}
}

func (r *Resolver) VisitBlockItem(blockItem parser.BlockItem) {
	switch item := blockItem.(type) {
	case parser.ASTStmt:
		r.VisitStmt(item)
	case *parser.Decl:
		r.VisitDecl(item)
	}
}

func (r *Resolver) VisitStmt(stmt parser.ASTStmt) {
	switch s := stmt.(type) {
	case *parser.ASTExpressionStmt:
		r.VisitExpressionStmt(s)
	case *parser.ASTReturnStmt:
		if s.ReturnValue != nil {
			r.VisitExpr(s.ReturnValue)
		}
	case *parser.ASTPrintStmt:
		if s.Value != nil {
			r.VisitExpr(s.Value)
		}
	}
}

func (r *Resolver) VisitDecl(decl *parser.Decl) {
	// Handle variable declaration
	if err := r.currentScope.Declare(decl.Ident, lexer.INT_TYPE); err != nil {
		r.addError(err.Error())
		return
	}

	if decl.Expr != nil {
		exprType := r.VisitExpr(decl.Expr)
		if exprType != lexer.INT_TYPE {
			r.addError(fmt.Sprintf("type mismatch in declaration of %s: expected INT_TYPE, got %v",
				decl.Ident, exprType))
			return
		}

		if info, exists := r.currentScope.Lookup(decl.Ident); exists {
			info.Initialized = true
		}
	}
}

func (r *Resolver) VisitExpr(expr parser.ASTExpression) lexer.TokenType {
	switch e := expr.(type) {
	case *parser.ASTIdent:
		if e.Token.Value == nil {
			r.addError("identifier token has nil value")
			return lexer.NUMBER
		}

		if info, exists := r.currentScope.Lookup(*e.Token.Value); exists {
			if !info.Initialized {
				r.addError(fmt.Sprintf("use of uninitialized variable %s", *e.Token.Value))
			}
			return info.Type
		}
		r.addError(fmt.Sprintf("undefined variable %s", *e.Token.Value))
		return lexer.NUMBER

	case *parser.ASTIntLitExpression:
		return lexer.INT_TYPE

	case *parser.ASTBinary:
		leftType := r.VisitExpr(e.Left)
		rightType := r.VisitExpr(e.Right)

		if leftType != rightType {
			r.addError(fmt.Sprintf("type mismatch in binary expression: %v %s %v",
				leftType, e.Op.String(), rightType))
			return lexer.NUMBER
		}
		return leftType
	}

	return lexer.NUMBER
}

func (r *Resolver) VisitExpressionStmt(stmt *parser.ASTExpressionStmt) {
	r.VisitExpr(stmt.Expression)
}
