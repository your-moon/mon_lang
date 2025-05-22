package parser

import (
	"bytes"
	"strings"

	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mtypes"
)

type StorageClass interface {
	storage()
}

type Static struct{}

func (i *Static) storage() {}

type Extern struct{}

func (i *Extern) storage() {}

type BlockItem interface {
	PrintAST(depth int) string
}

type ASTNode interface {
	TokenLiteral() string
	PrintAST(depth int) string
}

type ASTExpression interface {
	ASTNode
	expressionNode()
	GetType() mtypes.Type
	SetType(t mtypes.Type)
}

type ASTFactor interface {
	ASTNode
}

type ASTStmt interface {
	ASTNode
	statementNode()
}

type ASTDecl interface {
	ASTNode
	declNode()
}

type ASTImport struct {
	ASTNode
	Token      lexer.Token
	Ident      string
	SubImports []string
}

func (a *ASTImport) TokenLiteral() string {
	return a.Ident
}

func (a *ASTImport) PrintAST(depth int) string {
	return indent(depth) + "Import: " + a.Ident
}

func (a *ASTImport) declNode() {}

type ASTProgram struct {
	Decls []ASTDecl
}

func (a *ASTProgram) TokenLiteral() string {
	return a.Decls[0].TokenLiteral()
}

func (a *ASTProgram) PrintAST(depth int) string {
	var out bytes.Buffer
	for _, decl := range a.Decls {
		out.WriteString(decl.PrintAST(depth))
	}

	return out.String()
}

func indent(depth int) string {
	return strings.Repeat("  ", depth)
}
