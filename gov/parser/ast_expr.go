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
	A_QUESTIONMARK
	A_DOTDOT
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
	case ASTBinOp(A_QUESTIONMARK):
		return "?"
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

type ASTConditional struct {
	Token lexer.Token
	Cond  ASTExpression
	Then  ASTExpression
	Else  ASTExpression
}

func (a *ASTConditional) expressionNode()      {}
func (a *ASTConditional) TokenLiteral() string { return "CONDITIONAL" }
func (a *ASTConditional) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sTernary Expression:\n", indent(depth)))
	out.WriteString(fmt.Sprintf("%s├── If %s then:\n", indent(depth), a.Cond.PrintAST(depth+1)))
	if a.Then != nil {
		out.WriteString(fmt.Sprintf("%s│   └── %s\n", indent(depth), a.Then.PrintAST(depth+1)))
	}

	out.WriteString(fmt.Sprintf("%s└── Else:\n", indent(depth)))
	if a.Else != nil {
		out.WriteString(fmt.Sprintf("%s    └── %s\n", indent(depth), a.Else.PrintAST(depth+1)))
	}

	return out.String()
}

type ASTConstant struct {
	Token lexer.Token
	Value int64
}

func (a *ASTConstant) expressionNode()      {}
func (a *ASTConstant) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTConstant) PrintAST(depth int) string {
	return fmt.Sprintf("%d", a.Value)
}

type ASTStringExpression struct {
	Token lexer.Token
	Value string
}

func (a *ASTStringExpression) expressionNode()      {}
func (a *ASTStringExpression) TokenLiteral() string { return string(a.Token.Type) }
func (a *ASTStringExpression) PrintAST(depth int) string {
	return fmt.Sprintf("%s└─ String: %s", indent(depth), *a.Token.Value)
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

	out.WriteString(fmt.Sprintf("%sPrefix Expression (%s):\n", indent(depth), a.Op))
	out.WriteString(fmt.Sprintf("%s└─ ", indent(depth)))

	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	return out.String()
}

type ASTBinary struct {
	Token lexer.Token
	Right ASTExpression
	Left  ASTExpression
	Op    ASTBinOp
}

func (a ASTBinary) expressionNode()      {}
func (a ASTBinary) TokenLiteral() string { return string(a.Op.String()) }
func (a ASTBinary) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("(%s %s %s)",
		a.Left.PrintAST(depth),
		a.Op.String(),
		a.Right.PrintAST(depth)))
	return out.String()
}

type ASTAssignment struct {
	Token lexer.Token
	Left  ASTExpression
	Right ASTExpression
}

func (a *ASTAssignment) expressionNode()      {}
func (a *ASTAssignment) TokenLiteral() string { return "VAR" }
func (a *ASTAssignment) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s = %s",
		a.Left.PrintAST(depth),
		a.Right.PrintAST(depth)))
	return out.String()
}

type ASTRangeExpr struct {
	Token lexer.Token
	Start ASTExpression
	End   ASTExpression
}

func (a *ASTRangeExpr) expressionNode()      {}
func (a *ASTRangeExpr) TokenLiteral() string { return "RANGE" }
func (a *ASTRangeExpr) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%sRange Expression:\n", indent(depth)))
	out.WriteString(fmt.Sprintf("%s├─ Start:\n", indent(depth)))
	out.WriteString(a.Start.PrintAST(depth+1) + "\n")
	out.WriteString(fmt.Sprintf("%s└─ End:\n", indent(depth)))
	out.WriteString(a.End.PrintAST(depth + 1))
	return out.String()
}

type ASTVar struct {
	Token lexer.Token
	Ident string
}

func (a ASTVar) expressionNode()      {}
func (a ASTVar) TokenLiteral() string { return "VAR" }
func (a ASTVar) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s%s", indent(depth), a.Ident))
	return out.String()
}

type ASTUnary struct {
	Token lexer.Token
	Inner ASTExpression
	Op    lexer.TokenType
}

func (a *ASTUnary) expressionNode()      {}
func (a *ASTUnary) TokenLiteral() string { return string(a.Op) }
func (a *ASTUnary) PrintAST(depth int) string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s%s",
		a.Op,
		a.Inner.PrintAST(depth)))
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

	out.WriteString(fmt.Sprintf("%sInfix Expression (%s):\n", indent(depth), a.Op))
	out.WriteString(fmt.Sprintf("%s├─ Left:\n", indent(depth)))
	if a.Left != nil {
		out.WriteString(a.Left.PrintAST(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s└─ Right:\n", indent(depth)))
	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	return out.String()
}
