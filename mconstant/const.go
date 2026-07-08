/*
 * mon_lang - mconstant
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package mconstant

import "math"

var IntZero = Int32{
	Value: 0,
}

var IntOne = Int32{
	Value: 1,
}

type Const interface {
	constant()
	GetValue() int64
}

type Int64 struct {
	Value int64
}

func (i Int64) constant()       {}
func (i Int64) GetValue() int64 { return i.Value }

type Float64 struct {
	Value float64
}

func (f Float64) constant() {}

// GetValue returns the bit pattern; float constants flow through the
// integer-shaped plumbing as bits and only codegen interprets them.
func (f Float64) GetValue() int64 {
	return int64(math.Float64bits(f.Value))
}

type Int32 struct {
	Value int32
}

func (i Int32) constant()       {}
func (i Int32) GetValue() int64 { return int64(i.Value) }
