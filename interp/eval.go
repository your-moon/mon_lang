/*
 * mon_lang - interpreter statement/expression evaluation
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"fmt"

	"github.com/your-moon/mon_lang/parser"
)

// exec runs one block item. Block items are ASTStmt (буц, if, while, ...) or
// local var declarations (VarDecl); only the parser.BlockItem interface
// (PrintAST) is common to both, so that is what execBlock hands us.
func (ip *Interp) exec(item parser.BlockItem, fr *Frame) {
	if st, ok := item.(parser.ASTStmt); ok && ip.hook != nil {
		ip.hook.OnStmt(st, fr)
	}
	switch s := item.(type) {
	case *parser.ASTReturnStmt:
		var v Value = Void
		if s.ReturnValue != nil {
			v = ip.eval(s.ReturnValue, fr)
		}
		panic(retSignal{v})
	case *parser.ASTBlock:
		ip.execBlock(s, fr)
	default:
		panic(fmt.Sprintf("exec: unhandled statement %T", item))
	}
}

func (ip *Interp) eval(expr parser.ASTExpression, fr *Frame) Value {
	switch e := expr.(type) {
	case *parser.ASTConstInt:
		return I32(e.Value)
	case *parser.ASTConstLong:
		return I64(e.Value)
	default:
		panic(fmt.Sprintf("eval: unhandled expression %T", expr))
	}
}
