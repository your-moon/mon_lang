package parser

import (
	"bytes"
	"strings"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type BlockItem interface {
	PrintAST(depth int) string
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

type ASTLoopInit interface {
	ASTNode
	loopinit()
}

type ASTProgram struct {
	FnDef FNDef
}

func (a *ASTProgram) TokenLiteral() string {
	return a.FnDef.TokenLiteral()
}

type FNDef struct {
	Token      lexer.Token
	ReturnType lexer.TokenType
	Block      ASTBlock
}

func (a *ASTProgram) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(a.FnDef.PrintAST(depth))

	return out.String()
}

func indent(depth int) string {
	return strings.Repeat("  ", depth)
}
