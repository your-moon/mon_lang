package parser

import (
	"bytes"
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ASTBlock struct {
	BlockItems []BlockItem
}

type ASTBreakStmt struct{
	Token lexer.Token
}
type ASTContinueStmt struct{
	Token lexer.Token
}

type ASTLoop struct {
	Token lexer.Token
	Var   *ASTVar
	Start ASTExpression
	End   ASTExpression
	Body  ASTStmt
}

type ASTWhile struct {
	Token lexer.Token
	Cond  ASTExpression
	Body  ASTStmt
}

type ASTCompoundStmt struct {
	Block ASTBlock
}

func (a *ASTWhile) statementNode()       {}
func (a *ASTWhile) TokenLiteral() string { return "WHILE" }
func (a *ASTWhile) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%sWhile:\n", indent(depth)))
	if a.Cond != nil {
		out.WriteString(fmt.Sprintf("%s├─ Cond:\n", indent(depth)))
		out.WriteString(a.Cond.PrintAST(depth+1) + "\n")
	}
	out.WriteString(fmt.Sprintf("%s└─ Body:\n", indent(depth)))
	out.WriteString(a.Body.PrintAST(depth + 1))
	return out.String()
}

func (a *ASTLoop) statementNode()       {}
func (a *ASTLoop) TokenLiteral() string { return "LOOP" }
func (a *ASTLoop) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%sLoop:\n", indent(depth)))
	if a.Var != nil {
		out.WriteString(fmt.Sprintf("%s├─ Var:\n", indent(depth)))
		out.WriteString(a.Var.PrintAST(depth+1) + "\n")
	}
	out.WriteString(fmt.Sprintf("%s├─ Start:\n", indent(depth)))
	out.WriteString(a.Start.PrintAST(depth+1) + "\n")
	out.WriteString(fmt.Sprintf("%s├─ End:\n", indent(depth)))
	out.WriteString(a.End.PrintAST(depth+1) + "\n")
	out.WriteString(fmt.Sprintf("%s└─ Body:\n", indent(depth)))
	out.WriteString(a.Body.PrintAST(depth + 1))
	return out.String()
}

func (a *ASTContinueStmt) statementNode()       {}
func (a *ASTContinueStmt) TokenLiteral() string { return "CONTINUE" }
func (a *ASTContinueStmt) PrintAST(depth int) string {
	return fmt.Sprintf("%sContinue", indent(depth))
}

func (a *ASTBreakStmt) statementNode()       {}
func (a *ASTBreakStmt) TokenLiteral() string { return "BREAK" }
func (a *ASTBreakStmt) PrintAST(depth int) string {
	return fmt.Sprintf("%sBreak", indent(depth))
}

func (a *ASTCompoundStmt) statementNode()       {}
func (a *ASTCompoundStmt) TokenLiteral() string { return "COMPOUND" }
func (a *ASTCompoundStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%s{\n", indent(depth)))
	for _, b := range a.Block.BlockItems {
		out.WriteString(b.PrintAST(depth+1) + "\n")
	}
	out.WriteString(fmt.Sprintf("%s}\n", indent(depth)))
	return out.String()
}

type ASTIfStmt struct {
	Cond ASTExpression
	Then ASTStmt
	Else ASTStmt
}

type ASTNullStmt struct {
	Token lexer.Token
}

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

func (a *ASTIfStmt) statementNode()       {}
func (a *ASTIfStmt) TokenLiteral() string { return "IF" }
func (a *ASTIfStmt) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sif %s then:\n", indent(depth), a.Cond.PrintAST(depth)))
	if a.Then != nil {
		out.WriteString(fmt.Sprintf("%s|   %s\n", indent(depth), a.Then.PrintAST(depth)))
	}

	if a.Else != nil {
		out.WriteString(fmt.Sprintf("%selse:\n", indent(depth)))
		out.WriteString(fmt.Sprintf("%s|   %s\n", indent(depth), a.Else.PrintAST(depth)))
	}

	return out.String()
}

func (a *ASTNullStmt) statementNode()       {}
func (a *ASTNullStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTNullStmt) PrintAST(depth int) string {
	return fmt.Sprintf("%s// empty statement", indent(depth))
}

func (a *ASTBlock) statementNode()       {}
func (a *ASTBlock) TokenLiteral() string { return "BLOCK" }
func (a *ASTBlock) PrintAST(depth int) string {
	var out bytes.Buffer

	for _, b := range a.BlockItems {
		out.WriteString(b.PrintAST(depth+1) + "\n")
	}

	return out.String()
}

func (a *ExpressionStmt) statementNode()       {}
func (a *ExpressionStmt) TokenLiteral() string { return "EXPRESSIONSTMT" }
func (a *ExpressionStmt) PrintAST(depth int) string {
	if a.Expression != nil {
		return fmt.Sprintf("%s%s", indent(depth), a.Expression.PrintAST(depth))
	}
	return fmt.Sprintf("%s// empty expression", indent(depth))
}

func (a *FNDef) statementNode() {}
func (a *FNDef) TokenLiteral() string {
	return string(*a.Token.Value)
}
func (a *FNDef) PrintAST(depth int) string {
	var out bytes.Buffer

	if a.Token.Value != nil {
		out.WriteString(fmt.Sprintf("%sfunction %s() -> int {\n", indent(depth), *a.Token.Value))
	}

	for _, b := range a.Block.BlockItems {
		out.WriteString(b.PrintAST(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s}\n", indent(depth)))
	return out.String()
}

func (a *ASTReturnStmt) statementNode()       {}
func (a *ASTReturnStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTReturnStmt) PrintAST(depth int) string {
	if a.ReturnValue != nil {
		return fmt.Sprintf("%sreturn %s", indent(depth), a.ReturnValue.PrintAST(depth))
	}
	return fmt.Sprintf("%sreturn", indent(depth))
}

func (a *ASTPrintStmt) statementNode()       {}
func (a *ASTPrintStmt) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTPrintStmt) PrintAST(depth int) string {
	if a.Value != nil {
		return fmt.Sprintf("%sprint %s", indent(depth), a.Value.PrintAST(depth))
	}
	return fmt.Sprintf("%sprint", indent(depth))
}
