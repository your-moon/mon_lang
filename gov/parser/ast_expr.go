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
	Cond ASTExpression
	Then ASTExpression
	Else ASTExpression
}

func (a *ASTConditional) expressionNode()      {}
func (a *ASTConditional) TokenLiteral() string { return "CONDITIONAL" }
func (a *ASTConditional) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sConditional Expression:\n", indent(depth)))
	out.WriteString(fmt.Sprintf("%s├─ Condition:\n", indent(depth)))
	if a.Cond != nil {
		out.WriteString(a.Cond.PrintAST(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s├─ Then:\n", indent(depth)))
	if a.Then != nil {
		out.WriteString(a.Then.PrintAST(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s└─ Else:\n", indent(depth)))
	if a.Else != nil {
		out.WriteString(a.Else.PrintAST(depth+1) + "\n")
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
	return fmt.Sprintf("%s└─ Constant: %s", indent(depth), *a.Token.Value)
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
	Right ASTExpression
	Left  ASTExpression
	Op    ASTBinOp
}

func (a ASTBinary) expressionNode()      {}
func (a ASTBinary) TokenLiteral() string { return string(a.Op.String()) }
func (a ASTBinary) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sBinary Expression (%s):\n", indent(depth), a.Op.String()))
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

type ASTAssignment struct {
	Left  ASTExpression
	Right ASTExpression
}

func (a *ASTAssignment) expressionNode()      {}
func (a *ASTAssignment) TokenLiteral() string { return "VAR" }
func (a *ASTAssignment) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sAssignment:\n", indent(depth)))
	out.WriteString(fmt.Sprintf("%s├─ Variable:\n", indent(depth)))
	if a.Left != nil {
		out.WriteString(a.Left.PrintAST(depth+1) + "\n")
	}

	out.WriteString(fmt.Sprintf("%s└─ Value:\n", indent(depth)))
	if a.Right != nil {
		out.WriteString(a.Right.PrintAST(depth+1) + "\n")
	}

	return out.String()
}

type ASTVar struct {
	Ident string
}

func (a ASTVar) expressionNode()      {}
func (a ASTVar) TokenLiteral() string { return "VAR" }
func (a ASTVar) PrintAST(depth int) string {
	return fmt.Sprintf("%s└─ Variable: %s", indent(depth), a.Ident)
}

type ASTUnary struct {
	Inner ASTExpression
	Op    lexer.TokenType
}

func (a *ASTUnary) expressionNode()      {}
func (a *ASTUnary) TokenLiteral() string { return string(a.Op) }
func (a *ASTUnary) PrintAST(depth int) string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%sUnary Expression (%s):\n", indent(depth), a.Op))
	out.WriteString(fmt.Sprintf("%s└─ ", indent(depth)))

	if a.Inner != nil {
		out.WriteString(a.Inner.PrintAST(depth+1) + "\n")
	}

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
