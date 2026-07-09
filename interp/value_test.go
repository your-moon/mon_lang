/*
 * mon_lang - interpreter value model tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import "testing"

func TestValueBasics(t *testing.T) {
	if !I32(5).Truthy() || I32(0).Truthy() {
		t.Fatal("int truthiness wrong")
	}
	if I64(-3).AsInt() != -3 {
		t.Fatal("AsInt int64 wrong")
	}
	if F(2.9).AsInt() != 2 {
		t.Fatal("AsInt double should truncate")
	}
	o := &Obj{Kind: ObjBytes, Bytes: []byte("hi")}
	if !P(o).Truthy() || P(o).Ptr != o {
		t.Fatal("ptr value wrong")
	}
	if Void.Kind != KindVoid {
		t.Fatal("void kind wrong")
	}
}
