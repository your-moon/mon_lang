/*
 * mon_lang - tree-walking interpreter
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"fmt"

	"github.com/your-moon/mon_lang/parser"
)

// Hook is the debug seam. OnStmt runs before each statement; nil in normal
// runs (zero cost). The DAP server (a later stage) installs a Hook to
// implement breakpoints and stepping.
type Hook interface {
	OnStmt(node parser.ASTStmt, fr *Frame)
}

type Interp struct {
	fns  map[string]*parser.FnDecl
	hook Hook
	args []string
}

// retSignal unwinds a function body when буц executes.
type retSignal struct{ v Value }

func New(prog *parser.ASTProgram) *Interp {
	ip := &Interp{fns: map[string]*parser.FnDecl{}}
	for _, d := range prog.Decls {
		if fn, ok := d.(*parser.FnDecl); ok && !fn.IsExtern && fn.Body != nil {
			ip.fns[fn.Ident] = fn
		}
	}
	return ip
}

func (ip *Interp) SetHook(h Hook) { ip.hook = h }

func (ip *Interp) Run(entry string, args []string) (code int, err error) {
	fn, ok := ip.fns[entry]
	if !ok {
		return 0, fmt.Errorf("entry %q not found", entry)
	}
	ip.args = append([]string{"prog"}, args...)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("interp panic: %v", r)
		}
	}()
	return int(ip.call(fn, nil).AsInt()), nil
}

func (ip *Interp) call(fn *parser.FnDecl, argv []Value) (result Value) {
	fr := NewFrame()
	fr.FnName = fn.Ident
	for i, p := range fn.Params {
		if i < len(argv) {
			fr.Declare(p.Ident, argv[i])
		} else {
			fr.Declare(p.Ident, Void)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			if ret, ok := r.(retSignal); ok {
				result = ret.v
				return
			}
			panic(r)
		}
	}()
	ip.execBlock(fn.Body, fr)
	return Void
}

func (ip *Interp) execBlock(b *parser.ASTBlock, fr *Frame) {
	fr.Push()
	defer fr.Pop()
	for _, it := range b.BlockItems {
		ip.exec(it, fr)
	}
}
