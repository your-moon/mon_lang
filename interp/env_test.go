/*
 * mon_lang - interpreter environment tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import "testing"

func TestFrameScopes(t *testing.T) {
	f := NewFrame()
	f.Declare("x", I64(1))
	f.Push()
	f.Declare("y", I64(2))
	if v, ok := f.Get("x"); !ok || v.AsInt() != 1 {
		t.Fatal("outer var not visible in inner scope")
	}
	if !f.Set("x", I64(9)) {
		t.Fatal("Set should find outer x")
	}
	f.Pop()
	if v, _ := f.Get("x"); v.AsInt() != 9 {
		t.Fatal("Set did not mutate outer binding")
	}
	if _, ok := f.Get("y"); ok {
		t.Fatal("inner var leaked after Pop")
	}
	if f.Set("missing", I64(0)) {
		t.Fatal("Set on undeclared var should return false")
	}
}
