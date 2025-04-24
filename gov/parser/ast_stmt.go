package parser

import (
	"bytes"
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ASTIfStmt struct {
	Cond ASTExpression
	Then ASTExpression
	Else ASTExpression
}

func (a *ASTIfStmt) statementNode()       {}
func (a *ASTIfStmt) TokenLiteral() string { return "IF" }
func (a *ASTIfStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sIF[\n", indent(depth)))

	if a.Cond != nil {
		out.WriteString(a.Cond.PrintAST(depth+1) + "\n")
	}

	if a.Then != nil {
		out.WriteString(a.Then.PrintAST(depth+1) + "\n")
	}

	if a.Else != nil {
		out.WriteString(a.Else.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTNullStmt struct {
	Token lexer.Token
}

func (a *ASTNullStmt) statementNode()       {}
func (a *ASTNullStmt) TokenLiteral() string { return string(a.Token.Type) }

type ExpressionStmt struct {
	Expression ASTExpression
}

type ASTReturnStmt struct {
	Token       lexer.Token
	ReturnValue ASTExpression
}

type ASTPrintStmt struct {
	Token lexer.Token
	Value ASTExpression
}

func (a *ExpressionStmt) statementNode()       {}
func (a *ExpressionStmt) TokenLiteral() string { return "EXPRESSIONSTMT" }
func (a *ExpressionStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(indent(depth) + a.TokenLiteral() + " ")

	if a.Expression != nil {
		out.WriteString(a.Expression.PrintAST(depth + 1))
	}

	return out.String()
}

func (a *FNDef) statementNode() {}
func (a *FNDef) TokenLiteral() string {
	return string(*a.Token.Value)
}
func (a *FNDef) PrintAST(depth int) string {
	var out bytes.Buffer

	if a.Token.Value != nil {
		out.WriteString(fmt.Sprintf("%sFN %s[\n", indent(depth), *a.Token.Value))
	}

	for _, b := range a.BlockItems {
		out.WriteString(b.PrintAST(depth+1) + "\n")
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
