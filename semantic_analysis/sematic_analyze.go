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
	"sort"
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

// importUnit is one source file pulled in by an import: local code brings in
// every declaration, whereas a stdlib package exports only public/extern ones.
type importUnit struct {
	key     string // dedup key (absolute path, or "std:name")
	text    string // source text
	isLocal bool
}

// resolveImport turns an import path into concrete source units.
//
//   - "лексер/токен.mn"  → a single local file
//   - "лексер"           → a folder package (every .mn in baseDir/лексер), or,
//     if no such directory exists, the embedded stdlib package of that name
//   - "лексер/дэд"       → a local sub-package directory
func (s *SemanticAnalyzer) resolveImport(path string, baseDir string) ([]importUnit, error) {
	// single local file
	if strings.HasSuffix(path, ".mn") {
		full := filepath.Join(baseDir, path)
		data, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("импорт файл уншихад алдаа: %s: %v", path, err)
		}
		return []importUnit{{key: full, text: string(data), isLocal: true}}, nil
	}

	// a directory under baseDir is a folder package
	dir := filepath.Join(baseDir, path)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return s.readPackageDir(dir)
	}

	// a slash path that is neither a .mn file nor a directory is an error
	if strings.Contains(path, "/") {
		return nil, fmt.Errorf("багц олдсонгүй: %s", path)
	}

	// bare name → embedded standard-library package
	pkgSrc, ok := stdlib.StdPackage(path)
	if !ok {
		return nil, fmt.Errorf("стандарт багц олдсонгүй: %s", path)
	}
	return []importUnit{{key: "std:" + path, text: pkgSrc, isLocal: false}}, nil
}

// readPackageDir reads every .mn file in a folder package, in a deterministic
// order so merged declarations are stable across runs.
func (s *SemanticAnalyzer) readPackageDir(dir string) ([]importUnit, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("багц уншихад алдаа: %s: %v", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".mn") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	units := make([]importUnit, 0, len(names))
	for _, name := range names {
		full := filepath.Join(dir, name)
		data, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("импорт файл уншихад алдаа: %s: %v", full, err)
		}
		units = append(units, importUnit{key: full, text: string(data), isLocal: true})
	}
	return units, nil
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

	// worklist of pending imports; imported files can add more. Each carries
	// the directory it was written in, so a package's relative imports resolve
	// against that package's location, not the entry file's.
	type pendingImport struct {
		path    string
		ident   string
		baseDir string
	}
	var queue []pendingImport
	for _, decl := range program.Decls {
		imp, ok := decl.(*parser.ASTImport)
		if !ok {
			ownDecls = append(ownDecls, decl)
			continue
		}
		queue = append(queue, pendingImport{path: imp.FilePath, ident: imp.Ident, baseDir: s.baseDir})
	}

	for len(queue) > 0 {
		imp := queue[0]
		queue = queue[1:]
		if imp.path == "" {
			continue
		}

		// Resolve an import to one or more source units. A ".mn" path is a
		// single local file; a bare name or slash path that names a directory
		// is a folder package (every .mn file in it, Rust/Go style); a bare
		// name that is not a local directory is an embedded stdlib package.
		units, err := s.resolveImport(imp.path, imp.baseDir)
		if err != nil {
			return nil, err
		}

		// named import (гэж нэр): нэр.х becomes a checked alias for the
		// exported symbol х - qualified access is validated against the
		// package's combined export list. Types and methods stay global.
		var exports map[string]bool
		if imp.ident != "" {
			exports = make(map[string]bool)
			s.moduleHandles[imp.ident] = exports
		}

		for _, u := range units {
			if s.importedFiles[u.key] {
				continue
			}
			s.importedFiles[u.key] = true

			runeStr := convertToRuneArray(u.text)
			p := parser.NewParser(runeStr)
			importedProg, err := p.ParseProgram()
			if err != nil {
				return nil, fmt.Errorf("импорт парсингийн алдаа: %s: %v", imp.path, err)
			}

			// a local package file's relative imports resolve against its own
			// directory; stdlib units keep the current base (their imports are
			// bare stdlib names anyway).
			unitDir := imp.baseDir
			if u.isLocal {
				unitDir = filepath.Dir(u.key)
			}

			for _, d := range importedProg.Decls {
				switch dt := d.(type) {
				case *parser.FnDecl:
					if u.isLocal || dt.IsPublic || dt.IsExtern {
						if exports != nil && !dt.IsMethod {
							exports[dt.Ident] = true
						}
						importedDecls = append(importedDecls, dt)
					}
				case *parser.VarDecl:
					if u.isLocal || dt.IsPublic {
						if exports != nil {
							exports[dt.Ident] = true
						}
						importedDecls = append(importedDecls, dt)
					}
				case *parser.ASTStructDecl:
					// struct layouts always travel with the module
					importedDecls = append(importedDecls, dt)
				case *parser.ASTEnumDecl:
					// enums (constant sets) always travel with the module
					importedDecls = append(importedDecls, dt)
				case *parser.ASTImport:
					// transitive import: process it too (a package may import
					// another package), resolved relative to this file's dir.
					// Named-import handles don't nest.
					if dt.FilePath != "" {
						queue = append(queue, pendingImport{path: dt.FilePath, baseDir: unitDir})
					}
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
