package parser

import (
	"bytes"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ASTExpressionStmt struct {
	Token      lexer.Token
	Expression ASTExpression
}

func (a *ASTExpressionStmt) statementNode()       {}
func (a *ASTExpressionStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTExpressionStmt) PrintAST() string {

	var out bytes.Buffer

	out.WriteString(a.TokenLiteral() + " ")

	if a.Expression != nil {
		out.WriteString(a.Expression.PrintAST())
	}

	return out.String()
}

/////////////////////////////

type ASTFNStmt struct {
	Token lexer.Token
	Block ASTExpression
}

func (a *ASTFNStmt) statementNode() {}
func (a *ASTFNStmt) TokenLiteral() string {
	token := a.Token.Type
	return string(token)
}
func (a *ASTFNStmt) PrintAST() string {
	var out bytes.Buffer

	out.WriteString(a.TokenLiteral() + " ")

	if a.Block != nil {
		out.WriteString(a.Block.PrintAST())
	}

	return out.String()
}

/////////////////////////////

type ASTReturnStmt struct {
	Token       lexer.Token
	ReturnValue ASTExpression
}

func (a *ASTReturnStmt) statementNode()       {}
func (a *ASTReturnStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTReturnStmt) PrintAST() string {
	var out bytes.Buffer

	out.WriteString(a.TokenLiteral() + " ")

	if a.ReturnValue != nil {
		out.WriteString(a.ReturnValue.PrintAST())
	}

	return out.String()
}

/////////////////////////////

type ASTPrintStmt struct {
	Token lexer.Token
	Value ASTExpression
}

func (a *ASTPrintStmt) statementNode()       {}
func (a *ASTPrintStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTPrintStmt) PrintAST() string {
	var out bytes.Buffer

	out.WriteString(a.TokenLiteral() + " ")

	if a.Value != nil {
		out.WriteString(a.Value.PrintAST())
	}

	// out.WriteString(token.Semicolon)
	return out.String()
}
