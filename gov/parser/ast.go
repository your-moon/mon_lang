package parser

import (
	"bytes"
)

type ASTNode interface {
	TokenLiteral() string
	PrintAST() string
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
	Statements []ASTStmt
}

func (a *ASTProgram) TokenLiteral() string {
	if len(a.Statements) > 0 {
		return a.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}
func (a *ASTProgram) PrintAST() string {
	var out bytes.Buffer

	for _, s := range a.Statements {
		out.WriteString(s.PrintAST())
	}

	return out.String()
}
