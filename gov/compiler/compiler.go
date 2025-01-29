package compiler

import (
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
	case *parser.ASTProgram:
		for _, s := range ast.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *parser.ASTReturnStmt:
		err := c.Compile(ast.ReturnValue)
		if err != nil {
			return err
		}
		c.Emit(&gen.IRReturn{})
	case *parser.ASTIntLitExpression:
		c.Emit(&gen.IRPush{
			Value: ast.Value,
		})
	case *parser.ASTPrintStmt:
		err := c.Compile(ast.Value)
		if err != nil {
			return err
		}
		c.Emit(&gen.IRPrint{})
	default:
		return nil
	}

	return nil
}

func (c *Compiler) CompileExpr() {
}

func (c *Compiler) Emit(op gen.IR) {
	c.Irs = append(c.Irs, op)
}
