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

type FnType struct {
	ParamTypes []Type
	RetType    Type
}

func (t *FnType) typecheck() {}

// func (t FnType) IsFn() bool {
// 	return true
// }
