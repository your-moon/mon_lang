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

type IRPush struct {
	Value int64
}

func (i *IRPush) Ir() string { return fmt.Sprint("OP_PUSH ", i.Value) }
