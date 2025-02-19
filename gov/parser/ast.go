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
	// if len(a.Statements) > 0 {
	// 	return a.Statements[0].TokenLiteral()
	// } else {
	// 	return ""
	// }
	return a.FnDef.TokenLiteral()
}

func (a *ASTProgram) PrintAST(depth int) string {
	var out bytes.Buffer

	// for _, s := range a.Statements {
	// 	out.WriteString(s.PrintAST(depth) + "\n")
	// }

	return out.String()
}

func indent(depth int) string {
	return fmt.Sprintf("%s", strings.Repeat("  ", depth))
}
