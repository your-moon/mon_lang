package semanticanalysis

import (
	"github.com/your-moon/mon_lang/mtypes"
	"github.com/your-moon/mon_lang/parser"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
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

// registerImplicitStdlib registers built-in stdlib functions so they don't need extern declarations
func (s *SemanticAnalyzer) registerImplicitStdlib() {
	stdlibFns := []struct {
		name    string
		retType mtypes.Type
	}{
		{"хэвлэ", &mtypes.VoidType{}},
		{"мөр_хэвлэх", &mtypes.VoidType{}},
		{"унш", &mtypes.Int64Type{}},
		{"унш32", &mtypes.Int32Type{}},
		{"санамсаргүйТоо", &mtypes.Int32Type{}},
		{"одоо", &mtypes.Int64Type{}},
		{"malloc", &mtypes.Int64Type{}},
	}

	for _, fn := range stdlibFns {
		fnType := &mtypes.FnType{RetType: fn.retType}
		s.typeChecker.symbolTable.AddFn(fnType, fn.name, false)
		s.resolver.RegisterBuiltin(fn.name)
	}
}

func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, *symbols.SymbolTable, error) {
	s.registerImplicitStdlib()

	program, err := s.resolver.Resolve(program)
	if err != nil {
		return nil, nil, err
	}
	program, err = s.labelPass.LabelLoops(program)
	if err != nil {
		return nil, nil, err
	}
	program, err = s.typeChecker.CheckTopLevel(program)
	if err != nil {
		return nil, nil, err
	}
	return program, s.typeChecker.symbolTable, nil
}
