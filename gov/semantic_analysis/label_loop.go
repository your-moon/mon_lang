package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mn_compiler_go_version/errors"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/unique"
)

const (
	ErrOutsideBreak    = "давталтаас гадуур зогсох үйлдэл орсон байна."
	ErrOutsideContinue = "давталтаас гадуур үргэлжлүүлэх үйлдэл орсон байна."
)

type LoopPass struct {
	source    []int32
	uniqueGen unique.UniqueGen
	currentId string
}

func NewLoopPass(source []int32) LoopPass {
	return LoopPass{
		source:    source,
		uniqueGen: unique.NewUniqueGen(),
	}
}

func (r *LoopPass) LabelLoops(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	for i, fndecl := range program.Decls {
		switch decl := fndecl.(type) {
		case *parser.FnDecl:
			fndef, err := r.LabelFnDecl(decl)
			if err != nil {
				return nil, err
			}
			program.Decls[i] = fndef
		}

	}

	return program, nil
}

func (r *LoopPass) LabelFnDecl(fndecl *parser.FnDecl) (*parser.FnDecl, error) {
	if fndecl.Block != nil {
		block, err := r.LabelBlock("", fndecl.Block)
		if err != nil {
			return nil, err
		}
		fndecl.Block = block
	}

	return fndecl, nil
}

func (r *LoopPass) LabelBlock(curLabel string, program *parser.ASTBlock) (*parser.ASTBlock, error) {
	for _, item := range program.BlockItems {
		_, err := r.LabelBlockItem(curLabel, item)
		if err != nil {
			return program, err
		}
	}
	return program, nil
}

func (r *LoopPass) LabelBlockItem(curLabel string, program parser.BlockItem) (parser.BlockItem, error) {
	switch nodetype := program.(type) {
	case parser.ASTStmt:
		curLabel, err := r.LabelStmt(curLabel, nodetype)
		if err != nil {
			return program, err
		}
		return curLabel, nil
	case *parser.Decl:
		return nodetype, nil
	}

	return nil, fmt.Errorf("unreachable point")
}

func (r *LoopPass) LabelStmt(currentLabel string, program parser.ASTStmt) (parser.ASTStmt, error) {
	switch nodetype := program.(type) {
	case *parser.ASTContinueStmt:
		if currentLabel == "" {
			err := r.createLoopError(ErrOutsideContinue, nodetype.Token.Line, nodetype.Token.Span)
			return nil, err
		}
		nodetype.Id = currentLabel
		return nodetype, nil
	case *parser.ASTBreakStmt:
		if currentLabel == "" {
			err := r.createLoopError(ErrOutsideBreak, nodetype.Token.Line, nodetype.Token.Span)
			return nil, err
		}
		nodetype.Id = currentLabel
		return nodetype, nil
	case *parser.ASTLoop:
		newID := r.uniqueGen.MakeLabel("loop")
		block, err := r.LabelBlock(newID, &nodetype.Body)
		if err != nil {
			return nil, err
		}
		nodetype.Body = *block
		nodetype.Id = newID
		return nodetype, nil
	case *parser.ASTWhile:
		newID := r.uniqueGen.MakeLabel("while")
		nodetype.Id = newID
		body, err := r.LabelBlock(newID, &nodetype.Body)
		if err != nil {
			return nil, err
		}
		nodetype.Body = *body
		return nodetype, nil
	case *parser.ASTCompoundStmt:
		r.LabelBlock(currentLabel, &nodetype.Block)
		return nodetype, nil
	case *parser.ASTIfStmt:
		then, err := r.LabelStmt(currentLabel, nodetype.Then)
		if err != nil {
			return nil, err
		}
		nodetype.Then = then
		klse, err := r.LabelStmt(currentLabel, nodetype.Else)
		if err != nil {
			return nil, err
		}
		nodetype.Else = klse
		return nodetype, nil
	default:
		return program, nil
	}
}

func (r *LoopPass) createLoopError(message string, line int, span lexer.Span) error {
	return compilererrors.New(message, line, span, r.source, "Семантик шинжилгээ")
}

func (r *LoopPass) ResolveDecl(program *parser.Decl) (*parser.Decl, error) {
	return nil, nil
}

func (r *LoopPass) ResolveExpr(program parser.ASTExpression) (parser.ASTExpression, error) {
	if program == nil {
		return nil, nil
	}

	switch nodetype := program.(type) {
	case *parser.ASTConditional:
		cond, err := r.ResolveExpr(nodetype.Cond)
		if err != nil {
			return nodetype, err
		}

		nodetype.Cond = cond

		then, err := r.ResolveExpr(nodetype.Then)
		if err != nil {
			return nodetype, err
		}

		nodetype.Then = then
		klse, err := r.ResolveExpr(nodetype.Else)
		if err != nil {
			return nodetype, err
		}

		nodetype.Else = klse
		return nodetype, nil
	case *parser.ASTAssignment:

		return nil, nil

	case *parser.ASTVar:
	case *parser.ASTUnary:
		resolvedInner, err := r.ResolveExpr(nodetype.Inner)
		if err != nil {
			return nil, err
		}
		return &parser.ASTUnary{
			Inner: resolvedInner,
			Op:    nodetype.Op,
		}, nil

	case *parser.ASTBinary:
		resolvedLeft, err := r.ResolveExpr(nodetype.Left)
		if err != nil {
			return nil, err
		}

		resolvedRight, err := r.ResolveExpr(nodetype.Right)
		if err != nil {
			return nil, err
		}

		return &parser.ASTBinary{
			Left:  resolvedLeft,
			Right: resolvedRight,
			Op:    nodetype.Op,
		}, nil

	case *parser.ASTConstant:
		return nodetype, nil

		// case *parser.ASTIdent:
		// 	return nodetype, nil
	}

	return nil, r.createLoopError(
		fmt.Sprintf(compilererrors.ErrUnknownExpression, program),
		1,
		lexer.Span{Start: 0, End: 0},
	)
}
