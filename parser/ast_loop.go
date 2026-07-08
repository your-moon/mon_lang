/*
 * mon_lang - parser
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package parser
type ASTInitDecl struct {
	Decl Decl
}

func (a *ASTInitDecl) loopinit()       {}
func (a *ASTInitDecl) TokenLiteral() string { return "INITDECL" }
func (a *ASTInitDecl) PrintAST(depth int) string { return ""}

type ASTInitExp struct {
	Expr ASTExpression
}

func (a *ASTInitExp) loopinit()       {}
func (a *ASTInitExp) TokenLiteral() string { return "INITEXP" }
func (a *ASTInitExp) PrintAST(depth int) string { return ""}
