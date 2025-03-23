package parser

import (
	"bytes"
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ASTExpressionStmt struct {
	Token      lexer.Token
	Expression ASTExpression
}

type ASTFNStmt struct {
	Token      lexer.Token
	ReturnType lexer.TokenType
	Stmt       ASTStmt
}

type ASTReturnStmt struct {
	Token       lexer.Token
	ReturnValue ASTExpression
}

type ASTPrintStmt struct {
	Token lexer.Token
	Value ASTExpression
}

func (a *ASTExpressionStmt) statementNode()       {}
func (a *ASTExpressionStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTExpressionStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(indent(depth) + a.TokenLiteral() + " ")

	if a.Expression != nil {
		out.WriteString(a.Expression.PrintAST(depth + 1))
	}

	return out.String()
}

func (a *ASTFNStmt) statementNode() {}
func (a *ASTFNStmt) TokenLiteral() string {
	return string(*a.Token.Value)
}
func (a *ASTFNStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	if a.Token.Value != nil {
		out.WriteString(fmt.Sprintf("%sFN %s[\n", indent(depth), *a.Token.Value))
	}

	if a.Stmt != nil {
		out.WriteString(a.Stmt.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

func (a *ASTReturnStmt) statementNode()       {}
func (a *ASTReturnStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTReturnStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sRET[\n", indent(depth)))

	if a.ReturnValue != nil {
		out.WriteString(a.ReturnValue.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

func (a *ASTPrintStmt) statementNode()       {}
func (a *ASTPrintStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTPrintStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(indent(depth) + a.TokenLiteral() + " ")

	if a.Value != nil {
		out.WriteString(a.Value.PrintAST(depth + 1))
	}

	return out.String()
}
