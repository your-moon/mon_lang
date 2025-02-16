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
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return temp

}

func (c *Compiler) Compile(node parser.ASTNode) gen.IR {
	switch ast := node.(type) {
	case *parser.ASTUnary:
		src := c.Compile(ast.Inner)
		dst_name := c.makeTemp()
		dst := &gen.IRIdent{Name: dst_name}
		op := gen.ConvertToUnOp(ast.Op)
		ir := &gen.IRUnary{
			Op:  op,
			Src: src,
			Dst: dst,
		}
		c.Emit(ir)
		return ir

	case *parser.ASTIntLitExpression:
		ir := &gen.IRConstant{
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
			c.Compile(s)
			// ir.Instructions = append(ir.Instructions, instruction)
		}
		return ir
	case *parser.ASTReturnStmt:
		ir := &gen.IRReturn{}
		inner := c.Compile(ast.ReturnValue)
		ir.Inner = inner
		// c.Emit(ir)
		return ir
	// case *parser.ASTBlockStmt:
	// 	for _, s := range ast.Statements {
	// 		instruction := c.Compile(s)
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
