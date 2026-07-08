/*
 * mon_lang - mtypes
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package mtypes

type Type interface {
	typecheck()
}

type VoidType struct{}

func (t *VoidType) typecheck() {}

type Int64Type struct{}

func (t *Int64Type) typecheck() {}

type Int32Type struct{}

func (t *Int32Type) typecheck() {}

type UInt32Type struct{}

func (t *UInt32Type) typecheck() {}

type UInt64Type struct{}

func (t *UInt64Type) typecheck() {}

// IsUnsigned reports whether t is an unsigned integer type.
func IsUnsigned(t Type) bool {
	switch t.(type) {
	case *UInt32Type, *UInt64Type:
		return true
	}
	return false
}

// IsInteger reports whether t is any integer type.
func IsInteger(t Type) bool {
	switch t.(type) {
	case *Int32Type, *Int64Type, *UInt32Type, *UInt64Type:
		return true
	}
	return false
}

type StringType struct{}

func (t *StringType) typecheck() {}

type ArrayType struct {
	ElementType Type
	Size        int64 // 0 = size not part of the type (unsized/decayed)
}

func (t *ArrayType) typecheck() {}

// SizeOf returns a type's storage size in bytes. Arrays are heap-backed and
// held by reference, so as values they are pointer-sized.
func SizeOf(t Type) int64 {
	switch t.(type) {
	case *Int32Type, *UInt32Type:
		return 4
	case *Int64Type, *UInt64Type, *PointerType, *ArrayType, *StringType:
		return 8
	}
	return 8
}

type PointerType struct {
	Referenced Type
}

func (t *PointerType) typecheck() {}

type FnType struct {
	ParamTypes []Type
	RetType    Type
}

func (t *FnType) typecheck() {}

// func (t FnType) IsFn() bool {
// 	return true
// }

// IsSameType reports structural equality of two types.
func IsSameType(a, b Type) bool {
	switch at := a.(type) {
	case *VoidType:
		_, ok := b.(*VoidType)
		return ok
	case *Int64Type:
		_, ok := b.(*Int64Type)
		return ok
	case *Int32Type:
		_, ok := b.(*Int32Type)
		return ok
	case *UInt32Type:
		_, ok := b.(*UInt32Type)
		return ok
	case *UInt64Type:
		_, ok := b.(*UInt64Type)
		return ok
	case *StringType:
		_, ok := b.(*StringType)
		return ok
	case *ArrayType:
		bt, ok := b.(*ArrayType)
		if !ok || !IsSameType(at.ElementType, bt.ElementType) {
			return false
		}
		// an unsized array type matches any size (decay)
		return at.Size == 0 || bt.Size == 0 || at.Size == bt.Size
	case *PointerType:
		bt, ok := b.(*PointerType)
		return ok && IsSameType(at.Referenced, bt.Referenced)
	case *FnType:
		bt, ok := b.(*FnType)
		if !ok || len(at.ParamTypes) != len(bt.ParamTypes) {
			return false
		}
		for i := range at.ParamTypes {
			if !IsSameType(at.ParamTypes[i], bt.ParamTypes[i]) {
				return false
			}
		}
		return IsSameType(at.RetType, bt.RetType)
	}
	return false
}
