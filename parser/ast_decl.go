package parser

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

type Param struct {
	Token lexer.Token
	Ident string
	Type  Type
}

func (p *Param) PrintAST(depth int) string {
	return fmt.Sprintf("%sParam: %s\n%s├─ Type: %s",
		indent(depth),
		p.Ident,
		indent(depth),
		p.Type)
}

type FnDecl struct {
	Token      lexer.Token
	Ident      string
	IsPublic   bool
	IsExtern   bool
	Params     []Param
	ReturnType Type
	Body       *ASTBlock
}

func (d *FnDecl) declNode() {}
func (d *FnDecl) TokenLiteral() string {
	return d.Ident
}

func (d *FnDecl) PrintAST(depth int) string {
	if d.Body == nil {
		paramsStr := ""
		if len(d.Params) > 0 {
			paramsStr = "\n" + indent(depth) + "├─ Parameters:"
			for i, param := range d.Params {
				prefix := "├─"
				if i == len(d.Params)-1 {
					prefix = "└─"
				}
				paramsStr += fmt.Sprintf("\n%s%s %s: %s",
					indent(depth+1),
					prefix,
					param.Ident,
					param.Type)
			}
		} else {
			paramsStr = "\n" + indent(depth) + "├─ Parameters: none"
		}

		return fmt.Sprintf("%sFunction: %s%s\n%s└─ Return Type: %s",
			indent(depth),
			d.Ident,
			paramsStr,
			indent(depth),
			d.ReturnType)
	}

	paramsStr := ""
	if len(d.Params) > 0 {
		paramsStr = "\n" + indent(depth) + "├─ Parameters:"
		for i, param := range d.Params {
			prefix := "├─"
			if i == len(d.Params)-1 {
				prefix = "└─"
			}
			paramsStr += fmt.Sprintf("\n%s%s %s: %s",
				indent(depth+1),
				prefix,
				param.Ident,
				param.Type)
		}
	} else {
		paramsStr = "\n" + indent(depth) + "├─ Parameters: none"
	}

	return fmt.Sprintf("%sFunction: %s%s\n%s├─ Return Type: %s\n%s├─ IsPublic: %v\n%s└─ Body:\n%s",
		indent(depth),
		d.Ident,
		paramsStr,
		indent(depth),
		d.ReturnType,
		indent(depth),
		d.IsPublic,
		indent(depth),
		d.Body.PrintAST(depth+1))
}

type VarDecl struct {
	Token   lexer.Token
	Ident   string
	VarType Type
	Expr    ASTExpression
}

func (d *VarDecl) declNode() {}
func (d *VarDecl) TokenLiteral() string {
	return d.Ident
}
func (d *VarDecl) PrintAST(depth int) string {
	if d.Expr != nil {
		return fmt.Sprintf("%sVariable: %s\n%s└─ Initial Value: %s",
			indent(depth),
			d.Ident,
			indent(depth),
			d.Expr.PrintAST(depth+1))
	} else {
		return fmt.Sprintf("%sVariable: %s", indent(depth), d.Ident)
	}
}

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
