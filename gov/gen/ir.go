package gen

import "fmt"

type IRType string

type IR interface {
	Ir() string
}

type IRPrint struct{}

func (i *IRPrint) Ir() string { return fmt.Sprint("OP_PRINT") }

type IRReturn struct{}

func (i *IRReturn) Ir() string { return fmt.Sprint("OP_RETURN") }

type IRIntConst struct {
	Value int64
}

func (i *IRIntConst) Ir() string { return fmt.Sprint("OP_PUSH ", i.Value) }

type IRBinary struct {
	Value  int64
	Binary int64
}

func (i *IRBinary) Ir() string { return fmt.Sprint("OP_BIN ", i.Value) }
