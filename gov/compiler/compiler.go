package compiler

import (
	"fmt"

	"github.com/your-moon/mn_compiler_go_version/gen"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

type Compiler struct {
	Irs       []gen.IR
	TempCount uint64
}

func NewCompiler() Compiler {
	return Compiler{
		TempCount: 0,
	}
}

func (c *Compiler) makeTemp() string {
	return "tmp.0"
}

func (c *Compiler) Compile(node parser.ASTNode) gen.IR {
	switch ast := node.(type) {
	case *parser.ASTUnary:
		src := c.Compile(ast.Inner)
		dst_name := c.makeTemp()
		dst := &gen.IRVar{Value: dst_name}
		op := gen.ConvertToUnOp(ast.Op)
		ir := &gen.IRUnary{
			Op:  op,
			Src: src,
			Dst: dst,
		}
		return ir

	case *parser.ASTIntLitExpression:
		ir := &gen.IRMov{
			Value: ast.Value,
		}
		return ir
	case *parser.ASTProgram:
		for _, s := range ast.Statements {
			c.Compile(s)
		}
		return nil

	case *parser.ASTFNStmt:
		ir := &gen.IRFn{
			Name: *ast.Token.Value,
		}
		for _, s := range ast.Stmts {
			instruction := c.Compile(s)
			ir.Stmts = append(ir.Stmts, instruction)
		}

		c.Emit(ir)
		return ir
	case *parser.ASTReturnStmt:
		ir := &gen.IRReturn{}
		inner := c.Compile(ast.ReturnValue)
		ir.Inner = inner
		return ir
	// case *parser.ASTBlockStmt:
	// 	for _, s := range ast.Statements {
	// 		instruction := c.Compile(s)
	// 		c.Emit(instruction)
	// 	}
	// 	return nil
	default:
		panic(fmt.Sprintf("COMPILE ERROR: unknown ast %s", ast.PrintAST(0)))
	}

}

func (c *Compiler) CompileExpr(node parser.ASTNode) {
}

func (c *Compiler) Emit(op gen.IR) {
	c.Irs = append(c.Irs, op)
}
