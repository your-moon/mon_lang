package semanticanalysis

import (
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/symbols"
	"github.com/your-moon/mn_compiler_go_version/util/unique"
)

type SemanticAnalyzer struct {
	resolver    *Resolver
	labelPass   *LoopPass
	typeChecker *TypeChecker
}

func NewSemanticAnalyzer(source []int32, uniqueGen unique.UniqueGen, table *symbols.SymbolTable) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		resolver:    NewResolver(source, uniqueGen),
		labelPass:   NewLoopPass(source),
		typeChecker: NewTypeChecker(source, uniqueGen, table),
	}
}

func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, *symbols.SymbolTable, error) {
	program, err := s.resolver.Resolve(program)
	if err != nil {
		return nil, nil, err
	}
	program, err = s.labelPass.LabelLoops(program)
	if err != nil {
		return nil, nil, err
	}
	program, err = s.typeChecker.Check(program)
	if err != nil {
		return nil, nil, err
	}
	return program, s.typeChecker.symbolTable, nil
}
