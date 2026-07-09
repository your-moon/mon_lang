/*
 * mon_lang - parser
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package parser

import (
	"fmt"

	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mtypes"
)

type Param struct {
	Token lexer.Token
	Ident string
	Type  mtypes.Type
}

func (p *Param) PrintAST(depth int) string {
	return fmt.Sprintf("%sParam: %s\n%s├─ Type: %s",
		indent(depth),
		p.Ident,
		indent(depth),
		p.Type)
}

type FnDecl struct {
	Token        lexer.Token
	Ident        string
	Params       []Param
	ReturnType   mtypes.Type
	Body         *ASTBlock
	StorageClass StorageClass
	IsPublic     bool
	IsExtern     bool
	IsMethod     bool // defined in a хэрэгжүүл block; stays with its type
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
	Token        lexer.Token
	Ident        string
	VarType      mtypes.Type
	Expr         ASTExpression
	StorageClass StorageClass
	IsExtern     bool
	IsPublic     bool
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

// ASTStructDecl is a top-level `бүтэц Нэр { талбар: төрөл, ... }`.
type ASTStructDecl struct {
	Token      lexer.Token
	Name       string
	FieldNames []string
	FieldTypes []mtypes.Type
}

func (d *ASTStructDecl) declNode()            {}
func (d *ASTStructDecl) TokenLiteral() string { return "бүтэц" }
func (d *ASTStructDecl) PrintAST(depth int) string {
	out := indent(depth) + "бүтэц " + d.Name + " {\n"
	for _, n := range d.FieldNames {
		out += indent(depth+1) + n + "\n"
	}
	return out + indent(depth) + "}"
}

// ASTEnumDecl is a top-level `тоочих Нэр { ВАР1, ВАР2, ... }`. Each variant is
// an integer constant equal to its index; `Нэр.ВАРi` folds to that constant.
type ASTEnumDecl struct {
	Token    lexer.Token
	Name     string
	Variants []string
}

func (d *ASTEnumDecl) declNode()            {}
func (d *ASTEnumDecl) TokenLiteral() string { return "тоочих" }
func (d *ASTEnumDecl) PrintAST(depth int) string {
	out := indent(depth) + "тоочих " + d.Name + " {\n"
	for i, n := range d.Variants {
		out += indent(depth+1) + fmt.Sprintf("%s = %d\n", n, i)
	}
	return out + indent(depth) + "}"
}
