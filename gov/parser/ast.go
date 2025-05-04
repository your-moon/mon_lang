package parser

import (
	"bytes"
	"strings"
)

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

type ASTProgram struct {
	Decls []FnDecl
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
