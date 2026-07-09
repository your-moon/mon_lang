/*
 * mon_lang - interpreter statement/expression evaluation
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"fmt"

	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mtypes"
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
	case *parser.ASTCompoundStmt:
		ip.execBlock(&s.Block, fr)
	case *parser.VarDecl:
		var v Value = Void
		if s.Expr != nil {
			v = ip.eval(s.Expr, fr)
		}
		fr.Declare(s.Ident, v)
	case *parser.ASTIfStmt:
		if ip.eval(s.Cond, fr).Truthy() {
			ip.exec(s.Then, fr)
		} else if s.Else != nil {
			ip.exec(s.Else, fr)
		}
	case *parser.ExpressionStmt:
		if s.Expression != nil {
			ip.eval(s.Expression, fr)
		}
	case *parser.ASTNullStmt:
		// no-op
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
	case *parser.ASTConstFloat:
		return F(e.Value)
	case *parser.ASTVar:
		if v, ok := fr.Get(e.Ident); ok {
			return v
		}
		panic(fmt.Sprintf("eval: undefined variable %q", e.Ident))
	case *parser.ASTAssignment:
		v := ip.eval(e.Right, fr)
		ip.assign(e.Left, v, fr)
		return v
	case *parser.ASTBinary:
		return ip.binary(e, fr)
	case *parser.ASTUnary:
		return ip.unary(e, fr)
	case *parser.ASTConditional:
		if ip.eval(e.Cond, fr).Truthy() {
			return ip.eval(e.Then, fr)
		}
		return ip.eval(e.Else, fr)
	default:
		panic(fmt.Sprintf("eval: unhandled expression %T", expr))
	}
}

// assign writes v to an lvalue: a bare variable, an array index, a struct
// member, or a pointer deref. Extended by later tasks; here it covers vars.
func (ip *Interp) assign(target parser.ASTExpression, v Value, fr *Frame) {
	switch t := target.(type) {
	case *parser.ASTVar:
		if !fr.Set(t.Ident, v) {
			fr.Declare(t.Ident, v)
		}
	default:
		panic(fmt.Sprintf("assign: unhandled lvalue %T", target))
	}
}

// binary evaluates a binary expression, doing double arithmetic if either
// operand is a double, else integer arithmetic re-wrapped to the result's
// static width (тоо -> 32 bits, тоо64 -> 64 bits).
func (ip *Interp) binary(e *parser.ASTBinary, fr *Frame) Value {
	l := ip.eval(e.Left, fr)
	r := ip.eval(e.Right, fr)
	if l.Kind == KindDouble || r.Kind == KindDouble {
		return ip.binaryDouble(e.Op, l.asF(), r.asF())
	}
	a, b := l.AsInt(), r.AsInt()
	res := intOp(e.Op, a, b)
	if _, ok := e.GetType().(*mtypes.Int32Type); ok {
		return I32(res)
	}
	return I64(res)
}

// unary evaluates a prefix expression: -x, ~x, or !x.
func (ip *Interp) unary(e *parser.ASTUnary, fr *Frame) Value {
	v := ip.eval(e.Inner, fr)
	if v.Kind == KindDouble {
		switch e.Op {
		case lexer.MINUS:
			return F(-v.F)
		default:
			panic(fmt.Sprintf("unary: unsupported op %q on double", e.Op))
		}
	}
	i := v.AsInt()
	var res int64
	switch e.Op {
	case lexer.MINUS:
		res = -i
	case lexer.TILDE:
		res = ^i
	case lexer.NOT:
		if i == 0 {
			res = 1
		} else {
			res = 0
		}
	default:
		panic(fmt.Sprintf("unary: unhandled op %q", e.Op))
	}
	if _, ok := e.GetType().(*mtypes.Int32Type); ok {
		return I32(res)
	}
	return I64(res)
}

// intOp applies an integer binary operator; comparisons and logic yield 1/0.
func intOp(op parser.ASTBinOp, a, b int64) int64 {
	switch op {
	case parser.ASTBinOp(parser.A_PLUS):
		return a + b
	case parser.ASTBinOp(parser.A_MINUS):
		return a - b
	case parser.ASTBinOp(parser.A_MUL):
		return a * b
	case parser.ASTBinOp(parser.A_DIV):
		return a / b
	case parser.ASTBinOp(parser.A_MOD):
		return a % b
	case parser.ASTBinOp(parser.A_EQUALTO):
		return boolToInt(a == b)
	case parser.ASTBinOp(parser.A_NOTEQUAL):
		return boolToInt(a != b)
	case parser.ASTBinOp(parser.A_LESSTHAN):
		return boolToInt(a < b)
	case parser.ASTBinOp(parser.A_LESSTHANEQUAL):
		return boolToInt(a <= b)
	case parser.ASTBinOp(parser.A_GREATERTHAN):
		return boolToInt(a > b)
	case parser.ASTBinOp(parser.A_GREATERTHANEQUAL):
		return boolToInt(a >= b)
	case parser.ASTBinOp(parser.A_AND):
		return boolToInt(a != 0 && b != 0)
	case parser.ASTBinOp(parser.A_OR):
		return boolToInt(a != 0 || b != 0)
	default:
		panic(fmt.Sprintf("intOp: unhandled operator %s", op))
	}
}

// binaryDouble applies a binary operator to two float64 operands; comparisons
// and logic still yield an I32(0/1), matching the type checker (relational
// ops always produce тоо regardless of operand type).
func (ip *Interp) binaryDouble(op parser.ASTBinOp, a, b float64) Value {
	switch op {
	case parser.ASTBinOp(parser.A_PLUS):
		return F(a + b)
	case parser.ASTBinOp(parser.A_MINUS):
		return F(a - b)
	case parser.ASTBinOp(parser.A_MUL):
		return F(a * b)
	case parser.ASTBinOp(parser.A_DIV):
		return F(a / b)
	case parser.ASTBinOp(parser.A_EQUALTO):
		return I32(boolToInt(a == b))
	case parser.ASTBinOp(parser.A_NOTEQUAL):
		return I32(boolToInt(a != b))
	case parser.ASTBinOp(parser.A_LESSTHAN):
		return I32(boolToInt(a < b))
	case parser.ASTBinOp(parser.A_LESSTHANEQUAL):
		return I32(boolToInt(a <= b))
	case parser.ASTBinOp(parser.A_GREATERTHAN):
		return I32(boolToInt(a > b))
	case parser.ASTBinOp(parser.A_GREATERTHANEQUAL):
		return I32(boolToInt(a >= b))
	case parser.ASTBinOp(parser.A_AND):
		return I32(boolToInt(a != 0 && b != 0))
	case parser.ASTBinOp(parser.A_OR):
		return I32(boolToInt(a != 0 || b != 0))
	default:
		panic(fmt.Sprintf("binaryDouble: unhandled operator %s", op))
	}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// asF returns the Value as a float64: F directly for doubles, else the widened int.
func (v Value) asF() float64 {
	if v.Kind == KindDouble {
		return v.F
	}
	return float64(v.I)
}
