package parser

import (
	"bytes"

	"github.com/your-moon/mn_compiler_go_version/lexer"
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

type ASTPrefixExpression struct {
	Token lexer.Token
	Right ASTExpression
	Op    string
}

func (a *ASTPrefixExpression) expressionNode()      {}
func (a *ASTPrefixExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTPrefixExpression) PrintAST() string     { return "" }

type ASTInfixExpression struct {
	Token lexer.Token
	Left  ASTExpression
	Right ASTExpression
	Op    string
}

func (a *ASTInfixExpression) expressionNode()      {}
func (a *ASTInfixExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTInfixExpression) PrintAST() string     { return "" }

type ASTIntLitExpression struct {
	Token lexer.Token
	Value int64
}

func (a *ASTIntLitExpression) expressionNode()      {}
func (a *ASTIntLitExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTIntLitExpression) PrintAST() string     { return *a.Token.Value }

type ASTStringExpression struct {
	Token lexer.Token
	Value string
}

func (a *ASTStringExpression) expressionNode()      {}
func (a *ASTStringExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTStringExpression) PrintAST() string     { return *a.Token.Value }

type ASTExpressionStmt struct {
	Token      lexer.Token
	Expression ASTExpression
}

func (a *ASTExpressionStmt) statementNode()       {}
func (a *ASTExpressionStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTExpressionStmt) PrintAST() string     { return "" }

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

	// out.WriteString(token.Semicolon)
	return out.String()
}

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
