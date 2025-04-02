package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type BlockItem interface {
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

type ASTProgram struct {
	FnDef ASTFNDef
}

func (a *ASTProgram) TokenLiteral() string {
	return a.FnDef.TokenLiteral()
}

type ASTFNDef struct {
	Token      lexer.Token
	ReturnType lexer.TokenType
	BlockItem  []BlockItem
}

func (a *ASTProgram) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(a.FnDef.PrintAST(depth))

	return out.String()
}

func indent(depth int) string {
	return fmt.Sprintf("%s", strings.Repeat("  ", depth))
}
