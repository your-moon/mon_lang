package parser

type IRType string

const (
	OP_PUSH IRType = "OP_PUSH"
	OP_ADD  IRType = "OP_ADD"
	OP_SUB  IRType = "OP_SUB"
	OP_MUL  IRType = "OP_MUL"
	OP_DIV  IRType = "OP_DIV"
)
