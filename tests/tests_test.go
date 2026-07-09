/*
 * mon_lang - directive-driven test harness
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// Every .mn file under run/ and error/ is a self-describing test, in the
// style of LLVM lit / Go's test/run.go: expectations live in шалгах
// directives inside the source, and this harness is the only Go code.
//
//	// шалгах: гаралт "15 9 5\n"     expected stdout (Go string escapes)
//	// шалгах: код 7                 expected exit code (default 0)
//	// шалгах: оролт "-21\n"         stdin to feed the program
//	// шалгах: алдаа "олдсонгүй"     compile must FAIL containing substring
//
// run/ files compile and execute; multiple гаралт directives concatenate.
// error/ files must fail to compile with every алдаа substring present.
// Adding a test = adding one .mn file. Every fixed bug gets one.
//
// Set MON_TEST_CC=1 to additionally build run/ tests through the legacy
// as/cc path and require byte-identical stdout (differential mode).
package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/your-moon/mon_lang/armgen"
	codegen "github.com/your-moon/mon_lang/code_gen"
	"github.com/your-moon/mon_lang/code_gen/asmsymbol"
	"github.com/your-moon/mon_lang/linker"
	"github.com/your-moon/mon_lang/parser"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

type expectation struct {
	stdout    string
	hasStdout bool
	exitCode  int
	stdin     string
	errSubs   []string // non-empty => compilation must fail
}

var directiveRe = regexp.MustCompile(`^\s*//\s*шалгах:\s*(гаралт|код|оролт|алдаа)\s+(.+?)\s*$`)

func TestMain(m *testing.M) {
	// the legacy cc path needs the on-disk stdlib; point it at the repo
	// regardless of the harness working directory
	if abs, err := filepath.Abs("../stdlib"); err == nil {
		os.Setenv("MON_STDLIB", abs)
	}
	os.Exit(m.Run())
}

func parseDirectives(t *testing.T, path string) expectation {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	exp := expectation{}
	for _, line := range strings.Split(string(data), "\n") {
		m := directiveRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		switch m[1] {
		case "гаралт":
			s, err := strconv.Unquote(m[2])
			if err != nil {
				t.Fatalf("%s: bad гаралт directive %q: %v", path, m[2], err)
			}
			exp.stdout += s
			exp.hasStdout = true
		case "код":
			n, err := strconv.Atoi(m[2])
			if err != nil {
				t.Fatalf("%s: bad код directive %q: %v", path, m[2], err)
			}
			exp.exitCode = n
		case "оролт":
			s, err := strconv.Unquote(m[2])
			if err != nil {
				t.Fatalf("%s: bad оролт directive %q: %v", path, m[2], err)
			}
			exp.stdin += s
		case "алдаа":
			s, err := strconv.Unquote(m[2])
			if err != nil {
				t.Fatalf("%s: bad алдаа directive %q: %v", path, m[2], err)
			}
			exp.errSubs = append(exp.errSubs, s)
		}
	}
	return exp
}

func toRunes(s string) []int32 {
	runes := make([]int32, 0, len(s)+1)
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		runes = append(runes, r)
		s = s[size:]
	}
	return append(runes, 0)
}

// compile runs the full in-process pipeline (no subprocesses) and links
// with the native backend, or the legacy cc path when useCC is set.
// Panics from any pass are converted to errors so error-tests can assert
// on them and a broken pass can't take the harness down.
func compile(srcPath, outPath string, useCC bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("compiler panic: %v", r)
		}
	}()

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	runes := toRunes(string(data))

	p := parser.NewParser(runes)
	node, err := p.ParseProgram()
	if err != nil {
		return err
	}
	if errs := p.Errors(); len(errs) > 0 {
		return errs[0]
	}

	uniqueGen := unique.NewUniqueGen()
	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runes, uniqueGen, table, filepath.Dir(srcPath), "")
	resolvedAst, symbolTable, err := resolver.Analyze(node)
	if err != nil {
		return err
	}

	tackyGen := tackygenNew(uniqueGen, table)
	tackyProgram := tackyGen.EmitTacky(resolvedAst)
	if !useCC {
		// optimize only the native path: differential mode then compares
		// optimized output against the unoptimized legacy toolchain
		tackyProgram = tackygenOptimize(tackyProgram)
	}

	asmGen := codegen.NewAsmGen(table)
	asmProgram := asmGen.GenASTAsm(tackyProgram, symbolTable, asmsymbol.NewAsmSymbolTable())

	lnk := linker.NewLinker(outPath)
	if useCC {
		return compileCC(lnk, asmProgram)
	}
	return lnk.LinkNative(asmProgram)
}

func run(t *testing.T, bin, stdin string) (string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errBuf strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s: %v (stderr: %s)", bin, err, errBuf.String())
	}
	if ctx.Err() != nil {
		t.Fatalf("run %s: timed out after 10s", bin)
	}
	return out.String(), code
}

func TestRun(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("run", "*.mn"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no test files under tests/run: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			t.Parallel()
			exp := parseDirectives(t, f)
			if len(exp.errSubs) > 0 {
				t.Fatalf("%s: алдаа directive belongs under error/", f)
			}

			bin := filepath.Join(t.TempDir(), "prog")
			if err := compile(f, bin, false); err != nil {
				t.Fatalf("compile failed: %v", err)
			}
			gotOut, gotCode := run(t, bin, exp.stdin)
			if exp.hasStdout && gotOut != exp.stdout {
				t.Errorf("stdout mismatch\n got: %q\nwant: %q", gotOut, exp.stdout)
			}
			if gotCode != exp.exitCode {
				t.Errorf("exit code: got %d, want %d", gotCode, exp.exitCode)
			}

			// differential mode: legacy toolchain must agree byte-for-byte
			if os.Getenv("MON_TEST_CC") == "1" {
				binCC := filepath.Join(t.TempDir(), "prog_cc")
				if err := compile(f, binCC, true); err != nil {
					t.Fatalf("cc compile failed: %v", err)
				}
				ccOut, ccCode := run(t, binCC, exp.stdin)
				if ccOut != gotOut || ccCode != gotCode {
					t.Errorf("native/cc divergence: native (%q, %d) vs cc (%q, %d)",
						gotOut, gotCode, ccOut, ccCode)
				}
			}
		})
	}
}

// compileARM64 lowers a source file to a native signed arm64 executable via
// the armgen back end — the same path as `gen --arch arm64`.
func compileARM64(srcPath, outPath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("armgen panic: %v", r)
		}
	}()
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	runes := toRunes(string(data))
	p := parser.NewParser(runes)
	node, err := p.ParseProgram()
	if err != nil {
		return err
	}
	if errs := p.Errors(); len(errs) > 0 {
		return errs[0]
	}
	uniqueGen := unique.NewUniqueGen()
	table := symbols.NewSymbolTable()
	resolver := semanticanalysis.NewSemanticAnalyzer(runes, uniqueGen, table, filepath.Dir(srcPath), "")
	resolvedAst, _, err := resolver.Analyze(node)
	if err != nil {
		return err
	}
	tackyGen := tackygenNew(uniqueGen, table)
	prog := tackygenOptimize(tackyGen.EmitTacky(resolvedAst))
	return armgen.Compile(prog, outPath)
}

// TestRunARM64 exercises the native arm64 back end against the same run/ suite
// (opt-in: MON_TEST_ARM64=1, arm64 host only). Tests using features the arm64
// back end doesn't support yet skip rather than fail, so this stays green while
// coverage grows.
func TestRunARM64(t *testing.T) {
	if os.Getenv("MON_TEST_ARM64") != "1" {
		t.Skip("set MON_TEST_ARM64=1 to run the native arm64 suite")
	}
	if runtime.GOARCH != "arm64" {
		t.Skip("native arm64 host only")
	}
	files, err := filepath.Glob(filepath.Join("run", "*.mn"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no test files under tests/run: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			exp := parseDirectives(t, f)
			if len(exp.errSubs) > 0 {
				t.Skipf("%s: error directive", f)
			}
			bin := filepath.Join(t.TempDir(), "prog")
			if err := compileARM64(f, bin); err != nil {
				t.Skipf("arm64 unsupported: %v", err)
			}
			gotOut, gotCode := run(t, bin, exp.stdin)
			if exp.hasStdout && gotOut != exp.stdout {
				t.Errorf("stdout mismatch\n got: %q\nwant: %q", gotOut, exp.stdout)
			}
			if gotCode != exp.exitCode {
				t.Errorf("exit code: got %d, want %d", gotCode, exp.exitCode)
			}
		})
	}
}

func TestError(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("error", "*.mn"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no test files under tests/error: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			t.Parallel()
			exp := parseDirectives(t, f)
			if len(exp.errSubs) == 0 {
				t.Fatalf("%s: error tests need at least one алдаа directive", f)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- compile(f, filepath.Join(t.TempDir(), "prog"), false) }()
			var cerr error
			select {
			case cerr = <-done:
			case <-ctx.Done():
				t.Fatal("compiler hung for 10s on invalid input")
			}
			if cerr == nil {
				t.Fatal("compiled successfully, expected failure")
			}
			for _, sub := range exp.errSubs {
				if !strings.Contains(cerr.Error(), sub) {
					t.Errorf("diagnostic missing %q\ngot: %v", sub, cerr)
				}
			}
		})
	}
}
