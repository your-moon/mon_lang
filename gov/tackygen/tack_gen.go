package tackygen

import (
	"fmt"

	codegen "github.com/your-moon/mn_compiler_go_version/code_gen"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

type TackyGen struct {
	Irs       []codegen.IR
	TempCount uint64
}

func NewTackyGen() TackyGen {
	return TackyGen{
		TempCount: 0,
	}
}

func (c *TackyGen) makeTemp() string {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return temp

}

func (c *TackyGen) GenTacky(node parser.ASTProgram) TackyProgram {
	program := TackyProgram{}
	// for _, stmt := range node.Statements {
	// 	fndef := c.GenTackyFn(stmt)
	// 	program.FnDefs = append(program.FnDefs)
	// }
	return program
}

func (c *TackyGen) GenTackyFn(node parser.ASTFNStmt) TackyFn {
	fn := TackyFn{}

	return fn
}

func (c *TackyGen) CompileExpr(node parser.ASTNode) {
}

func (c *TackyGen) Emit(op codegen.IR) {
	c.Irs = append(c.Irs, op)
}

// switch ast := node.(type) {
// case *parser.ASTProgram:
// 	for _, s := range ast.Statements {
// 		c.GenTacky(s)
// 	}
// 	return nil
//
// case *parser.ASTFNStmt:
// 	ir := &codegen.TackyFn{
// 		Name: *ast.Token.Value,
// 	}
// 	c.Emit(ir)
// 	for _, s := range ast.Stmts {
// 		instruction := c.GenTacky(s)
// 		ir.Instructions = append(ir.Instructions, instruction)
// 	}
// 	return ir
// case *parser.ASTUnary:
// 	eval_inner := c.GenTacky(ast.Inner)
//
// 	dst_name := c.makeTemp()
// 	dst := &codegen.IRIdent{Name: dst_name}
//
// 	op := codegen.ConvertToUnOp(ast.Op)
// 	ir := &codegen.IRUnary{
// 		Op:  op,
// 		Src: eval_inner,
// 		Dst: dst,
// 	}
// 	c.Emit(ir)
// 	return ir
//
// case *parser.ASTIntLitExpression:
// 	ir := &codegen.IRConstant{
// 		Value: ast.Value,
// 	}
// 	return ir
// case *parser.ASTReturnStmt:
// 	ir := &codegen.IRReturn{}
// 	eval_exp := c.GenTacky(ast.ReturnValue)
// 	ir.Inner = eval_exp
// 	return ir
// default:
// 	panic(fmt.Sprintf("COMPILE ERROR: unknown ast %s", ast.PrintAST(0)))
// }
