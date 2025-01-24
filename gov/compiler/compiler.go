package compiler

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/gen"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

type Compiler struct {
	Irs []gen.IR
}

func NewCompiler() Compiler {
	return Compiler{}
}

func (c *Compiler) Compile(node parser.ASTNode) error {
	switch ast := node.(type) {
	case *parser.ASTReturnStmt:
		err := c.Compile(ast.ReturnValue)
		if err != nil {
			return err
		}
		c.Emit(&gen.IRReturn{})
	case *parser.ASTIntLitExpression:
		fmt.Println("INTLIT")
		c.Emit(&gen.IRPush{
			Value: ast.Value,
		})
	default:
		return nil
		// return fmt.Errorf("unreachable")
	}

	return nil
}

func (c *Compiler) CompileExpr() {
}

func (c *Compiler) Emit(op gen.IR) {
	c.Irs = append(c.Irs, op)
}
