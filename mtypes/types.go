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
	case *Int64Type, *UInt64Type, *PointerType, *ArrayType, *StringType, *StructType:
		return 8 // structs are references
	}
	return 8
}

type PointerType struct {
	Referenced Type
}

func (t *PointerType) typecheck() {}

// NamedType is an unresolved type reference (e.g. a struct name in source);
// the type checker replaces it with the real type.
type NamedType struct {
	Name string
}

func (t *NamedType) typecheck() {}

type StructField struct {
	Name   string
	Type   Type
	Offset int64
}

// StructType has reference semantics like arrays: a value of struct type is
// a pointer to its heap storage, so codegen only ever moves 8-byte handles.
type StructType struct {
	Name   string
	Fields []StructField
	Size   int64 // laid-out payload size in bytes
}

func (t *StructType) typecheck() {}

// Field returns the named field, or nil.
func (t *StructType) Field(name string) *StructField {
	for i := range t.Fields {
		if t.Fields[i].Name == name {
			return &t.Fields[i]
		}
	}
	return nil
}

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
	case *StructType:
		bt, ok := b.(*StructType)
		return ok && at.Name == bt.Name
	case *NamedType:
		bt, ok := b.(*NamedType)
		return ok && at.Name == bt.Name
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
