package parser

type ASTtype string

const (
	ASTleaf ASTtype = "ASTleaf"
)

type ASTnode struct {
	Type     ASTtype
	Left     *ASTnode
	Right    *ASTnode
	Op       int
	IntValue int
}

func NewLeafNode(value int) ASTnode {
	return ASTnode{
		Type:     ASTleaf,
		IntValue: value,
	}
}
