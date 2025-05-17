package mtypes

type Type interface {
	typecheck()
}

type VoidType struct{}

func (t *VoidType) typecheck() {}

type LongType struct{}

func (t *LongType) typecheck() {}

type IntType struct{}

func (t *IntType) typecheck() {}

type FnType struct {
	ParamTypes []Type
	RetType    Type
}

func (t *FnType) typecheck() {}

// func (t FnType) IsFn() bool {
// 	return true
// }
