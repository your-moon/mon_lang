package parser

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type Decl struct {
	Token lexer.Token
	Ident string
	Expr  ASTExpression
}

// PrintAST implements BlockItem.
func (d *Decl) PrintAST(depth int) string {
	if d.Expr != nil {
		return fmt.Sprintf("%sDeclaration: %s\n%s└─ Initial Value:%s",
			indent(depth),
			d.Ident,
			indent(depth),
			d.Expr.PrintAST(depth+1))
	} else {
		return fmt.Sprintf("%sDeclaration: %s", indent(depth), d.Ident)
	}
}

// TokenLiteral implements BlockItem.
func (d *Decl) TokenLiteral() string {
	return d.Ident
}
