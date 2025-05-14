package semanticanalysis

type Type interface {
	typecheck()
	IsFn() bool
}

type IntType struct{}

func (t *IntType) typecheck() {}

func (t *IntType) IsFn() bool {
	return false
}

type FnType struct {
	ParamCount int
}

func (t *FnType) typecheck() {}

func (t FnType) IsFn() bool {
	return true
}
