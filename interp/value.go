/*
 * mon_lang - interpreter value model
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

type Kind int

const (
	KindVoid Kind = iota
	KindInt32
	KindInt64
	KindDouble
	KindPtr
)

type ObjKind int

const (
	ObjBytes ObjKind = iota
	ObjArray
	ObjStruct
)

// Obj is a heap object. Bytes backs strings and byte buffers; Fields backs
// value arrays and structs. Go's GC reclaims unreachable Objs, so interpreted
// programs never leak (unlike the native bump allocator).
type Obj struct {
	Kind   ObjKind
	Bytes  []byte
	Fields []Value
}

// Value is a tagged union. Integers of both widths live in I, doubles in F,
// heap handles in Ptr.
type Value struct {
	Kind Kind
	I    int64
	F    float64
	Ptr  *Obj
}

var Void = Value{Kind: KindVoid}

func I32(v int64) Value { return Value{Kind: KindInt32, I: int64(int32(v))} }
func I64(v int64) Value { return Value{Kind: KindInt64, I: v} }
func F(v float64) Value { return Value{Kind: KindDouble, F: v} }
func P(o *Obj) Value    { return Value{Kind: KindPtr, Ptr: o} }

func (v Value) Truthy() bool {
	switch v.Kind {
	case KindDouble:
		return v.F != 0
	case KindPtr:
		return v.Ptr != nil
	default:
		return v.I != 0
	}
}

func (v Value) AsInt() int64 {
	if v.Kind == KindDouble {
		return int64(v.F)
	}
	return v.I
}
