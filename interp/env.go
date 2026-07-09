/*
 * mon_lang - interpreter environment (call frame + lexical scopes)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

// Frame is one function activation: a stack of lexical scopes. The innermost
// scope is scopes[len-1]. Block statements Push/Pop; the function body is the
// outermost scope.
type Frame struct {
	scopes []map[string]Value
	FnName string // for stack traces / DAP
}

func NewFrame() *Frame {
	return &Frame{scopes: []map[string]Value{{}}}
}

func (f *Frame) Push() { f.scopes = append(f.scopes, map[string]Value{}) }
func (f *Frame) Pop()  { f.scopes = f.scopes[:len(f.scopes)-1] }

func (f *Frame) Declare(name string, v Value) {
	f.scopes[len(f.scopes)-1][name] = v
}

func (f *Frame) Get(name string) (Value, bool) {
	for i := len(f.scopes) - 1; i >= 0; i-- {
		if v, ok := f.scopes[i][name]; ok {
			return v, true
		}
	}
	return Void, false
}

func (f *Frame) Set(name string, v Value) bool {
	for i := len(f.scopes) - 1; i >= 0; i-- {
		if _, ok := f.scopes[i][name]; ok {
			f.scopes[i][name] = v
			return true
		}
	}
	return false
}
