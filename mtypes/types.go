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

type StringType struct{}

func (t *StringType) typecheck() {}

type ArrayType struct {
	ElementType Type
}

func (t *ArrayType) typecheck() {}

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
	case *StringType:
		_, ok := b.(*StringType)
		return ok
	case *ArrayType:
		bt, ok := b.(*ArrayType)
		return ok && IsSameType(at.ElementType, bt.ElementType)
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
