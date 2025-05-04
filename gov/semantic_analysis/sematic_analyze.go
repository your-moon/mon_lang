package semanticanalysis

import (
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/unique"
)

type SemanticAnalyzer struct {
	resolver    *Resolver
	labelPass   *LoopPass
	typeChecker *TypeChecker
}

func NewSemanticAnalyzer(source []int32, uniqueGen unique.UniqueGen) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		resolver:    NewResolver(source, uniqueGen),
		labelPass:   NewLoopPass(source),
		typeChecker: NewTypeChecker(source, uniqueGen),
	}
}

func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	program, err := s.resolver.Resolve(program)
	if err != nil {
		return nil, err
	}
	program, err = s.labelPass.LabelLoops(program)
	if err != nil {
		return nil, err
	}
	program, err = s.typeChecker.Check(program)
	if err != nil {
		return nil, err
	}
	return program, nil
}
