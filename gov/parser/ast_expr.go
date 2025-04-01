package parser

import (
	"bytes"
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type ASTBinOp int

const (
	A_PLUS int = iota
	A_MINUS
	A_DIV
	A_MUL

	A_ASSIGN

	A_AND
	A_OR
	A_EQUALTO
	A_NOTEQUAL
	A_LESSTHAN
	A_LESSTHANEQUAL
	A_GREATERTHAN
	A_GREATERTHANEQUAL
)

func (op ASTBinOp) String() string {
	switch op {
	case ASTBinOp(A_PLUS):
		return "+"
	case ASTBinOp(A_MINUS):
		return "-"
	case ASTBinOp(A_DIV):
		return "/"
	case ASTBinOp(A_MUL):
		return "*"
	case ASTBinOp(A_AND):
		return "&&"
	case ASTBinOp(A_OR):
		return "||"
	case ASTBinOp(A_EQUALTO):
		return "=="
	case ASTBinOp(A_NOTEQUAL):
		return "!="
	case ASTBinOp(A_LESSTHAN):
		return "<"
	case ASTBinOp(A_LESSTHANEQUAL):
		return "<="
	case ASTBinOp(A_GREATERTHAN):
		return ">"
	case ASTBinOp(A_GREATERTHANEQUAL):
		return ">="
	default:
		return "unknown"
	}
}

type ASTIdent struct {
	Token lexer.Token
}

func (a *ASTIdent) expressionNode()      {}
func (a *ASTIdent) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTIdent) PrintAST(depth int) string {
	return indent(depth) + *a.Token.Value
}

type ASTIntLitExpression struct {
	Token lexer.Token
	Value int64
}

func (a *ASTIntLitExpression) expressionNode()      {}
func (a *ASTIntLitExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTIntLitExpression) PrintAST(depth int) string {
	return indent(depth) + *a.Token.Value
}

type ASTStringExpression struct {
	Token lexer.Token
	Value string
}

func (a *ASTStringExpression) expressionNode()      {}
func (a *ASTStringExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTStringExpression) PrintAST(depth int) string {
	return indent(depth) + *a.Token.Value
}

type ASTPrefixExpression struct {
	Token lexer.Token
	Right ASTExpression
	Op    string
}

func (a *ASTPrefixExpression) expressionNode()      {}
func (a *ASTPrefixExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTPrefixExpression) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sPREFIX %s[\n", indent(depth), a.Op))

	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTBinary struct {
	Right ASTExpression
	Left  ASTExpression
	Op    ASTBinOp
}

func (a ASTBinary) expressionNode()      {}
func (a ASTBinary) TokenLiteral() string { return string(a.Op.String()) }
func (a ASTBinary) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sBINARY %s[\n", indent(depth), a.Op.String()))

	if a.Left != nil {
		out.WriteString(a.Left.PrintAST(depth+1) + "\n")
	}
	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTAssignment struct {
	Left  ASTExpression
	Right ASTExpression
}

func (a *ASTAssignment) expressionNode()      {}
func (a *ASTAssignment) TokenLiteral() string { return "VAR" }
func (a *ASTAssignment) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sASSIGNMENT[\n", indent(depth)))

	if a.Left != nil {
		out.WriteString(a.Left.PrintAST(depth+1) + "\n")
	}

	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTVar struct {
	Ident ASTExpression
}

func (a *ASTVar) expressionNode()      {}
func (a *ASTVar) TokenLiteral() string { return "VAR" }
func (a *ASTVar) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sVAR[\n", indent(depth)))

	if a.Ident != nil {
		out.WriteString(a.Ident.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTUnary struct {
	Inner ASTExpression
	Op    lexer.TokenType
}

func (a *ASTUnary) expressionNode()      {}
func (a *ASTUnary) TokenLiteral() string { return string(a.Op) }
func (a *ASTUnary) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sUNARY %s[\n", indent(depth), a.Op))

	if a.Inner != nil {
		out.WriteString(a.Inner.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}

type ASTInfixExpression struct {
	Token lexer.Token
	Left  ASTExpression
	Right ASTExpression
	Op    string
}

func (a *ASTInfixExpression) expressionNode()      {}
func (a *ASTInfixExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTInfixExpression) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sINFIX %s[\n", indent(depth), a.Op))

	if a.Left != nil {
		out.WriteString(a.Left.PrintAST(depth+1) + "\n")
	}

	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	out.WriteString(indent(depth) + "]")
	return out.String()
}
