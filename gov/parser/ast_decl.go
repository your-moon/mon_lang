package parser

import "fmt"

type Decl struct {
	Ident string
	Expr  ASTExpression
}

// PrintAST implements BlockItem.
func (d *Decl) PrintAST(depth int) string {
	if d.Expr != nil {
		return fmt.Sprintf("Decl: %s = %s", d.Ident, d.Expr.PrintAST(depth+1))
	} else {
		return fmt.Sprintf("Decl: %s", d.Ident)
	}
}

// TokenLiteral implements BlockItem.
func (d *Decl) TokenLiteral() string {
	return d.Ident
}

func (d Decl) delc() {}
