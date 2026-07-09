/*
 * mon_lang - interpreter driver tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"testing"
	"unicode/utf8"

	"github.com/your-moon/mon_lang/parser"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

// toRunes converts source text to a NUL-terminated []int32; the lexer detects
// EOF via a NUL sentinel (Scanner.Current == 0), so the trailing 0 is required.
func toRunes(s string) []int32 {
	runes := make([]int32, 0, len(s)+1)
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		runes = append(runes, r)
		s = s[size:]
	}
	return append(runes, 0)
}

// resolve is a test helper: source text -> resolved AST.
func resolve(t *testing.T, src string) *parser.ASTProgram {
	t.Helper()
	runes := toRunes(src)
	p := parser.NewParser(runes)
	node, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	table := symbols.NewSymbolTable()
	sa := semanticanalysis.NewSemanticAnalyzer(runes, unique.NewUniqueGen(), table, ".", "")
	resolved, _, err := sa.Analyze(node)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	return resolved
}

func TestRunReturnsExitCode(t *testing.T) {
	prog := resolve(t, "функц үндсэн() -> тоо { буц 42; }")
	code, err := New(prog).Run("үндсэн", nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != 42 {
		t.Fatalf("exit code = %d, want 42", code)
	}
}
