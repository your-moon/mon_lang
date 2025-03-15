package parser

import (
	"bytes"
	"fmt"
	"strings"
)

type ASTNode interface {
	TokenLiteral() string
	PrintAST(depth int) string
}

type ASTExpression interface {
	ASTNode
	expressionNode()
}

type ASTStmt interface {
	ASTNode
	statementNode()
}

type ASTProgram struct {
	FnDef ASTFNStmt
}

func (a *ASTProgram) TokenLiteral() string {
	return a.FnDef.TokenLiteral()
}

func (a *ASTProgram) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(a.FnDef.PrintAST(depth))

	return out.String()
}

func indent(depth int) string {
	return fmt.Sprintf("%s", strings.Repeat("  ", depth))
}
