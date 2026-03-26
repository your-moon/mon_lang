package semanticanalysis

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/your-moon/mon_lang/mtypes"
	"github.com/your-moon/mon_lang/parser"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

type SemanticAnalyzer struct {
	resolver      *Resolver
	labelPass     *LoopPass
	typeChecker   *TypeChecker
	importedFiles map[string]bool
	baseDir       string
}

func NewSemanticAnalyzer(source []int32, uniqueGen unique.UniqueGen, table *symbols.SymbolTable, baseDir string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		resolver:      NewResolver(source, uniqueGen),
		labelPass:     NewLoopPass(source),
		typeChecker:   NewTypeChecker(source, uniqueGen, table),
		importedFiles: make(map[string]bool),
		baseDir:       baseDir,
	}
}

func convertToRuneArray(dataString string) []int32 {
	var runeString []int32
	for len(dataString) > 0 {
		r, size := utf8.DecodeRuneInString(dataString)
		runeString = append(runeString, r)
		dataString = dataString[size:]
	}
	runeString = append(runeString, 0)
	return runeString
}

func (s *SemanticAnalyzer) processImports(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	var importedDecls []parser.ASTDecl
	var ownDecls []parser.ASTDecl

	for _, decl := range program.Decls {
		imp, ok := decl.(*parser.ASTImport)
		if !ok {
			ownDecls = append(ownDecls, decl)
			continue
		}
		if imp.FilePath == "" {
			continue
		}

		filePath := filepath.Join(s.baseDir, imp.FilePath)
		if s.importedFiles[filePath] {
			continue
		}
		s.importedFiles[filePath] = true

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("импорт файл уншихад алдаа: %s: %v", imp.FilePath, err)
		}

		runeStr := convertToRuneArray(string(data))
		p := parser.NewParser(runeStr)
		importedProg, err := p.ParseProgram()
		if err != nil {
			return nil, fmt.Errorf("импорт парсингийн алдаа: %s: %v", imp.FilePath, err)
		}

		for _, d := range importedProg.Decls {
			switch dt := d.(type) {
			case *parser.FnDecl:
				if dt.IsPublic {
					importedDecls = append(importedDecls, dt)
				}
			case *parser.VarDecl:
				if dt.IsPublic {
					importedDecls = append(importedDecls, dt)
				}
			}
		}
	}

	program.Decls = append(importedDecls, ownDecls...)
	return program, nil
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
		{"чөлөөлөх", &mtypes.VoidType{}},
	}

	for _, fn := range stdlibFns {
		fnType := &mtypes.FnType{RetType: fn.retType}
		s.typeChecker.symbolTable.AddFn(fnType, fn.name, false)
		s.resolver.RegisterBuiltin(fn.name)
	}
}

func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, *symbols.SymbolTable, error) {
	s.registerImplicitStdlib()

	program, err := s.processImports(program)
	if err != nil {
		return nil, nil, err
	}

	program, err = s.resolver.Resolve(program)
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
