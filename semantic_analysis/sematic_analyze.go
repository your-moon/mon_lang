/*
 * mon_lang - semantic_analysis
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package semanticanalysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/your-moon/mon_lang/parser"
	"github.com/your-moon/mon_lang/stdlib"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

type SemanticAnalyzer struct {
	resolver      *Resolver
	labelPass     *LoopPass
	typeChecker   *TypeChecker
	importedFiles map[string]bool
	baseDir       string
	stdlibDir     string
	moduleHandles map[string]map[string]bool // handle -> exported value names
}

func NewSemanticAnalyzer(source []int32, uniqueGen unique.UniqueGen, table *symbols.SymbolTable, baseDir string, stdlibDir string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		resolver:      NewResolver(source, uniqueGen),
		labelPass:     NewLoopPass(source),
		typeChecker:   NewTypeChecker(source, uniqueGen, table),
		importedFiles: make(map[string]bool),
		baseDir:       baseDir,
		stdlibDir:     stdlibDir,
		moduleHandles: make(map[string]map[string]bool),
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

	// Auto-import the prelude (builtin declarations). It is embedded in the
	// compiler binary; a stdlibDir override (tests, custom preludes) wins.
	preludeSrc := stdlib.Prelude
	if s.stdlibDir != "" {
		preludePath := filepath.Join(s.stdlibDir, "prelude.mn")
		if data, err := os.ReadFile(preludePath); err == nil {
			preludeSrc = string(data)
		}
	}
	runeStr := convertToRuneArray(preludeSrc)
	p := parser.NewParser(runeStr)
	preludeProg, err := p.ParseProgram()
	if err != nil {
		return nil, fmt.Errorf("prelude парсингийн алдаа: %v", err)
	}
	importedDecls = append(importedDecls, preludeProg.Decls...)

	// worklist of pending imports; imported files can add more
	var queue []*parser.ASTImport
	for _, decl := range program.Decls {
		imp, ok := decl.(*parser.ASTImport)
		if !ok {
			ownDecls = append(ownDecls, decl)
			continue
		}
		queue = append(queue, imp)
	}

	for len(queue) > 0 {
		imp := queue[0]
		queue = queue[1:]
		if imp.FilePath == "" {
			continue
		}

		// Bare names (no "/" and no ".mn") are standard-library packages,
		// resolved from the embedded pkg set - available from any directory.
		var srcText string
		if !strings.Contains(imp.FilePath, "/") && !strings.HasSuffix(imp.FilePath, ".mn") {
			pkgSrc, ok := stdlib.StdPackage(imp.FilePath)
			if !ok {
				return nil, fmt.Errorf("стандарт багц олдсонгүй: %s", imp.FilePath)
			}
			if s.importedFiles["std:"+imp.FilePath] {
				continue
			}
			s.importedFiles["std:"+imp.FilePath] = true
			srcText = pkgSrc
		} else {
			filePath := filepath.Join(s.baseDir, imp.FilePath)
			if s.importedFiles[filePath] {
				continue
			}
			s.importedFiles[filePath] = true
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("импорт файл уншихад алдаа: %s: %v", imp.FilePath, err)
			}
			srcText = string(data)
		}

		runeStr := convertToRuneArray(srcText)
		p := parser.NewParser(runeStr)
		importedProg, err := p.ParseProgram()
		if err != nil {
			return nil, fmt.Errorf("импорт парсингийн алдаа: %s: %v", imp.FilePath, err)
		}

		// named import (гэж нэр): нэр.х becomes a checked alias for the
		// exported symbol х - qualified access is validated against the
		// module's export list. Types and methods stay global.
		var exports map[string]bool
		if imp.Ident != "" {
			exports = make(map[string]bool)
			s.moduleHandles[imp.Ident] = exports
		}
		for _, d := range importedProg.Decls {
			switch dt := d.(type) {
			case *parser.FnDecl:
				if dt.IsPublic || dt.IsExtern {
					if exports != nil && !dt.IsMethod {
						exports[dt.Ident] = true
					}
					importedDecls = append(importedDecls, dt)
				}
			case *parser.VarDecl:
				if dt.IsPublic {
					if exports != nil {
						exports[dt.Ident] = true
					}
					importedDecls = append(importedDecls, dt)
				}
			case *parser.ASTStructDecl:
				// struct layouts always travel with the module
				importedDecls = append(importedDecls, dt)
			case *parser.ASTImport:
				// transitive import: process it too (a package may import
				// another package). Named-import handles don't nest.
				if dt.FilePath != "" {
					queue = append(queue, &parser.ASTImport{FilePath: dt.FilePath})
				}
			}
		}
	}

	// declaration-before-use: extern declarations must precede any function
	// that calls them. Transitively-imported package externs would otherwise
	// land after the library that uses them, so hoist all externs first.
	var externs, rest []parser.ASTDecl
	for _, d := range importedDecls {
		if fn, ok := d.(*parser.FnDecl); ok && fn.IsExtern {
			externs = append(externs, d)
		} else {
			rest = append(rest, d)
		}
	}
	ordered := append(externs, rest...)
	program.Decls = append(ordered, ownDecls...)
	return program, nil
}

func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, *symbols.SymbolTable, error) {
	program, err := s.processImports(program)
	if err != nil {
		return nil, nil, err
	}

	s.resolver.moduleHandles = s.moduleHandles
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
