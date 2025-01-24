package compiler

import "github.com/your-moon/mn_compiler_go_version/parser"

type Compiler struct {
	Ast parser.ASTNode
}

func NewCompiler(ast parser.ASTNode) Compiler {
	return Compiler{
		Ast: ast,
	}
}

// func (c *Compiler) Compile() {
// 	switch c.Ast.GenType {
// 	case parser.ASTprogram:
// 		c.CompileReturn()
// 	}
// }

func (c *Compiler) CompileStmt() {
	switch c.Ast.Op {
	case parser.ASTreturn:
		c.CompileReturn()
	}
}
func (c *Compiler) CompileReturn() {
}

func (c *Compiler) CompileExpr() {
}
