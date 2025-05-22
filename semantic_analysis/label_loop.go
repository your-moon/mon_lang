package semanticanalysis

import (
	"fmt"

	compilererrors "github.com/your-moon/mn_compiler/errors"
	"github.com/your-moon/mn_compiler/lexer"
	"github.com/your-moon/mn_compiler/parser"
	"github.com/your-moon/mn_compiler/util/unique"
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

func NewLoopPass(source []int32) *LoopPass {
	return &LoopPass{
		source:    source,
		uniqueGen: unique.NewUniqueGen(),
	}
}

func (r *LoopPass) createLoopError(message string, line int, span lexer.Span) *compilererrors.CompilerError {
	return compilererrors.New(message, line, span, r.source, "Семантик шинжилгээ")
}

func (r *LoopPass) LabelLoops(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	for i, decl := range program.Decls {
		switch decltype := decl.(type) {
		case *parser.FnDecl:
			fndef, err := r.LabelFnDecl(decltype)
			if err != nil {
				return program, err
			}
			program.Decls[i] = fndef
		}

	}

	return program, nil
}

func (r *LoopPass) LabelFnDecl(fndecl *parser.FnDecl) (*parser.FnDecl, error) {
	if fndecl.Body != nil {
		block, err := r.LabelBlock("", fndecl.Body)
		if err != nil {
			return nil, err
		}
		fndecl.Body = block
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
	case parser.ASTDecl:
		return nodetype, nil
	}

	return nil, r.createLoopError(
		fmt.Sprintf(compilererrors.ErrUnknownExpression, program),
		1,
		lexer.Span{Start: 0, End: 0},
	)
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
		block, err := r.LabelBlock(currentLabel, &nodetype.Block)
		if err != nil {
			return nil, err
		}
		nodetype.Block = *block
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
