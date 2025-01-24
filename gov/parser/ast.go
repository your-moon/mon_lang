package parser

type ASTtype string

const (
	ASTleaf  ASTtype = "ASTleaf"
	ASTunary ASTtype = "ASTunary"
)

type ASToptype string

const (
	ASTreturn ASToptype = "ASTreturn"
)

type ASTnode struct {
	Type     ASTtype
	Left     *ASTnode
	Right    *ASTnode
	Op       ASToptype
	IntValue int
}

func NewUnaryNode(op ASToptype, left *ASTnode, intval int) ASTnode {
	return ASTnode{
		Type:     ASTunary,
		Left:     left,
		Op:       op,
		IntValue: intval,
	}
}

func NewLeafNode(value int) ASTnode {
	return ASTnode{
		Type:     ASTleaf,
		IntValue: value,
	}
}
