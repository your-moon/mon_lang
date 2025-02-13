package parser

import (
	"bytes"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

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
func (a *ASTInfixExpression) PrintAST() string {

	var out bytes.Buffer

	// out.WriteString( " ")

	// if a.Left != nil {
	//
	// 	out.WriteString("LEFT:")
	// 	out.WriteString(a.Left.PrintAST())
	// }
	// if a.Right != nil {
	// 	out.WriteString("RIGHT:")
	// 	out.WriteString(a.Right.PrintAST())
	// }

	return out.String()
}
