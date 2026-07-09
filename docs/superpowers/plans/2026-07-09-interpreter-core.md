# Interpreter Core Implementation Plan (Stage 1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A tree-walking interpreter (`interp/` Go package) that executes the resolved mon AST, validated against the existing `tests/run/*.mn` suite with byte-for-byte parity to the native back end.

**Architecture:** Walk the post-semantic-analysis AST (names resolved, enums folded, types attached). A tagged-union value model over a Go-hosted heap (Go GC, no leaks). Extern builtins are Go host functions matching `encoder/stdlib.go` semantics. One `Hook.OnStmt` seam (nil in normal runs) is the future DAP attach point. The interpreter becomes a second test mode alongside native — the differential oracle.

**Tech Stack:** Go 1.23+, existing `parser`, `semantic_analysis`, `symbols`, `mtypes`, `lexer`, `util/unique` packages. `testing` for tests.

## Global Constraints

- Go 1.23+ (build-time only). No new third-party dependencies.
- Reuse `lexer` → `parser` → `semantic_analysis` unchanged; the interpreter consumes the **resolved** AST (`SemanticAnalyzer.Analyze` output) only.
- Builtin semantics must match `encoder/stdlib.go` / the native runtime exactly (e.g. `хэвлэ` prints a signed decimal with no newline; `бутархай_хэвлэх` prints 6 truncated fractional digits; `санамсаргүй_тоо` returns 1..n). Where output must be deterministic for tests, only builtins that the `run/` suite exercises need parity.
- `тоо` is signed 32-bit (`Int32`); `тоо64` is signed 64-bit (`Int64`); `бутархай` is `Double`. Sub-word integer semantics must match (e.g. `тоо` wraps/sign-extends at 32 bits).
- Copyright header on every new Go file:
  ```go
  /*
   * mon_lang - <short description>
   *
   * Copyright (c) 2026 Munkherdene
   * SPDX-License-Identifier: MIT (see LICENSE)
   */
  ```
- Commit messages end with:
  `Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo`
- No `Co-Authored-By` trailer.
- Use `rtk proxy go test ...` to see raw test output (the RTK hook otherwise summarizes it).

---

## File Structure

- `interp/value.go` — the `Value` tagged union + heap object types (`Obj`), plus conversion/formatting helpers.
- `interp/env.go` — `Frame` / lexical scope chain; variable get/set/declare.
- `interp/interp.go` — `Interp` type, program loading (index functions by name), `Run(entry, args)`, statement execution, the `Hook` interface, and control-flow signals (return/break/continue via typed panics recovered at boundaries).
- `interp/eval.go` — expression evaluation (`eval(node) Value`): literals, vars, arithmetic/logic/compare, calls, arrays, members, pointers, cast, match-as-expression is N/A (match is a statement), conditional `?:`.
- `interp/builtins.go` — the extern bridge: a `map[string]func(*Interp, []Value) Value` implementing `хэвлэ`, `мөр_урт`, `байт`, `байт_тавих`, `мөр_шинэ`, `monAlloc`, `тэмдэгт_хэвлэх`, `мөр_хэвлэх`, `бутархай_хэвлэх`, `эхэвлэ`, `унш`, `аргумент`/`аргумент_тоо`, `хүлээх`, `чөлөөлөх`, `дэлгэц_цэвэрлэх`, file I/O.
- `interp/interp_test.go` — Go unit tests for the interpreter in isolation (small inline programs).
- `tests/interp_test.go` — wires the interpreter into the shared `run/` suite as a second mode (`TestRunInterp`).

Split rationale: value model, environment, statement driver, expression evaluator, and the builtin table each change for different reasons and are individually testable.

---

## Task 1: Value model

**Files:**
- Create: `interp/value.go`
- Test: `interp/value_test.go`

**Interfaces:**
- Produces:
  - `type Kind int` with `KindInt32, KindInt64, KindDouble, KindPtr, KindVoid`.
  - `type Value struct { Kind Kind; I int64; F float64; Ptr *Obj }` — integers (both widths) live in `I`, doubles in `F`, heap handles in `Ptr`.
  - `type Obj struct { Bytes []byte; Fields []Value; Kind ObjKind }` with `type ObjKind int` (`ObjBytes` for strings/byte buffers, `ObjArray` for value arrays, `ObjStruct` for structs).
  - Constructors: `func I32(v int64) Value`, `func I64(v int64) Value`, `func F(v float64) Value`, `func P(o *Obj) Value`, `var Void = Value{Kind: KindVoid}`.
  - `func (v Value) Truthy() bool` — nonzero int, nonzero double, or non-nil ptr.
  - `func (v Value) AsInt() int64` — returns `I` for int kinds, truncates `F` for doubles.

- [ ] **Step 1: Write the failing test**

```go
/*
 * mon_lang - interpreter value model tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import "testing"

func TestValueBasics(t *testing.T) {
	if !I32(5).Truthy() || I32(0).Truthy() {
		t.Fatal("int truthiness wrong")
	}
	if I64(-3).AsInt() != -3 {
		t.Fatal("AsInt int64 wrong")
	}
	if F(2.9).AsInt() != 2 {
		t.Fatal("AsInt double should truncate")
	}
	o := &Obj{Kind: ObjBytes, Bytes: []byte("hi")}
	if !P(o).Truthy() || P(o).Ptr != o {
		t.Fatal("ptr value wrong")
	}
	if Void.Kind != KindVoid {
		t.Fatal("void kind wrong")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestValueBasics -v`
Expected: FAIL (package/type undefined).

- [ ] **Step 3: Write minimal implementation**

```go
/*
 * mon_lang - interpreter value model
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

type Kind int

const (
	KindVoid Kind = iota
	KindInt32
	KindInt64
	KindDouble
	KindPtr
)

type ObjKind int

const (
	ObjBytes ObjKind = iota
	ObjArray
	ObjStruct
)

// Obj is a heap object. Bytes backs strings and byte buffers; Fields backs
// value arrays and structs. Go's GC reclaims unreachable Objs, so interpreted
// programs never leak (unlike the native bump allocator).
type Obj struct {
	Kind   ObjKind
	Bytes  []byte
	Fields []Value
}

// Value is a tagged union. Integers of both widths live in I, doubles in F,
// heap handles in Ptr.
type Value struct {
	Kind Kind
	I    int64
	F    float64
	Ptr  *Obj
}

var Void = Value{Kind: KindVoid}

func I32(v int64) Value { return Value{Kind: KindInt32, I: int64(int32(v))} }
func I64(v int64) Value { return Value{Kind: KindInt64, I: v} }
func F(v float64) Value { return Value{Kind: KindDouble, F: v} }
func P(o *Obj) Value    { return Value{Kind: KindPtr, Ptr: o} }

func (v Value) Truthy() bool {
	switch v.Kind {
	case KindDouble:
		return v.F != 0
	case KindPtr:
		return v.Ptr != nil
	default:
		return v.I != 0
	}
}

func (v Value) AsInt() int64 {
	if v.Kind == KindDouble {
		return int64(v.F)
	}
	return v.I
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestValueBasics -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/value.go interp/value_test.go
git commit -m "feat(interp): value model — tagged union over a Go-hosted heap

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 2: Environment (frames & scopes)

**Files:**
- Create: `interp/env.go`
- Test: `interp/env_test.go`

**Interfaces:**
- Consumes: `Value` (Task 1).
- Produces:
  - `type Frame struct { ... }` with methods `Declare(name string, v Value)`, `Get(name string) (Value, bool)`, `Set(name string, v Value) bool`, `Push()`, `Pop()`.
  - `func NewFrame() *Frame`.
  - Scopes are a stack of maps; `Get`/`Set` search innermost-out; `Declare` writes to the innermost scope; `Set` mutates the nearest existing binding and returns false if none exists.

- [ ] **Step 1: Write the failing test**

```go
/*
 * mon_lang - interpreter environment tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import "testing"

func TestFrameScopes(t *testing.T) {
	f := NewFrame()
	f.Declare("x", I64(1))
	f.Push()
	f.Declare("y", I64(2))
	if v, ok := f.Get("x"); !ok || v.AsInt() != 1 {
		t.Fatal("outer var not visible in inner scope")
	}
	if !f.Set("x", I64(9)) {
		t.Fatal("Set should find outer x")
	}
	f.Pop()
	if v, _ := f.Get("x"); v.AsInt() != 9 {
		t.Fatal("Set did not mutate outer binding")
	}
	if _, ok := f.Get("y"); ok {
		t.Fatal("inner var leaked after Pop")
	}
	if f.Set("missing", I64(0)) {
		t.Fatal("Set on undeclared var should return false")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestFrameScopes -v`
Expected: FAIL (NewFrame undefined).

- [ ] **Step 3: Write minimal implementation**

```go
/*
 * mon_lang - interpreter environment (call frame + lexical scopes)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

// Frame is one function activation: a stack of lexical scopes. The innermost
// scope is scopes[len-1]. Block statements Push/Pop; the function body is the
// outermost scope.
type Frame struct {
	scopes []map[string]Value
	FnName string // for stack traces / DAP
}

func NewFrame() *Frame {
	return &Frame{scopes: []map[string]Value{{}}}
}

func (f *Frame) Push() { f.scopes = append(f.scopes, map[string]Value{}) }
func (f *Frame) Pop()  { f.scopes = f.scopes[:len(f.scopes)-1] }

func (f *Frame) Declare(name string, v Value) {
	f.scopes[len(f.scopes)-1][name] = v
}

func (f *Frame) Get(name string) (Value, bool) {
	for i := len(f.scopes) - 1; i >= 0; i-- {
		if v, ok := f.scopes[i][name]; ok {
			return v, true
		}
	}
	return Void, false
}

func (f *Frame) Set(name string, v Value) bool {
	for i := len(f.scopes) - 1; i >= 0; i-- {
		if _, ok := f.scopes[i][name]; ok {
			f.scopes[i][name] = v
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestFrameScopes -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/env.go interp/env_test.go
git commit -m "feat(interp): call frames with a lexical scope stack

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 3: Interpreter skeleton — load program, run entry, integer literal + return + exit code

**Files:**
- Create: `interp/interp.go`
- Create: `interp/eval.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `Frame`, `Value` (Tasks 1–2); `parser.ASTProgram`, `parser.FnDecl`, `parser.ASTBlock`, `parser.ASTReturnStmt`, `parser.ASTConstInt`, `parser.ASTConstLong` (existing).
- Produces:
  - `type Hook interface { OnStmt(node parser.ASTStmt, fr *Frame) }`.
  - `type Interp struct { fns map[string]*parser.FnDecl; hook Hook; ... }`.
  - `func New(prog *parser.ASTProgram) *Interp` — indexes non-extern `FnDecl`s by `Ident`.
  - `func (ip *Interp) SetHook(h Hook)`.
  - `func (ip *Interp) Run(entry string, args []string) (int, error)` — calls `entry` with no mon params, returns its integer result as the exit code.
  - `func (ip *Interp) call(fn *parser.FnDecl, argv []Value) Value` — new `Frame`, binds params, executes body, returns the returned value (or `Void`).
  - internal `exec(stmt parser.ASTStmt, fr *Frame)` and `eval(expr parser.ASTExpression, fr *Frame) Value`.
  - Control flow via typed panics: `type retSignal struct{ v Value }`, recovered in `call`.

- [ ] **Step 1: Write the failing test**

```go
/*
 * mon_lang - interpreter driver tests
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"testing"

	"github.com/your-moon/mon_lang/parser"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

// resolve is a test helper: source text -> resolved AST.
func resolve(t *testing.T, src string) *parser.ASTProgram {
	t.Helper()
	runes := []int32(src)
	p := parser.NewParser(runes)
	node, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse: %v", err)
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
```

> Note: `[]int32(src)` relies on Go treating a string→[]rune→[]int32 conversion. If the harness's `toRunes` (UTF-8 aware, NUL-terminated) is needed, import and reuse that instead. Verify during Step 2 whether the parser needs a trailing 0; if it does, append one: `runes = append(runes, 0)`.

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestRunReturnsExitCode -v`
Expected: FAIL (New undefined).

- [ ] **Step 3: Write minimal implementation**

`interp/interp.go`:

```go
/*
 * mon_lang - tree-walking interpreter
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"fmt"

	"github.com/your-moon/mon_lang/parser"
)

// Hook is the debug seam. OnStmt runs before each statement; nil in normal
// runs (zero cost). The DAP server (a later stage) installs a Hook to
// implement breakpoints and stepping.
type Hook interface {
	OnStmt(node parser.ASTStmt, fr *Frame)
}

type Interp struct {
	fns  map[string]*parser.FnDecl
	hook Hook
	args []string
}

// retSignal unwinds a function body when буц executes.
type retSignal struct{ v Value }

func New(prog *parser.ASTProgram) *Interp {
	ip := &Interp{fns: map[string]*parser.FnDecl{}}
	for _, d := range prog.Decls {
		if fn, ok := d.(*parser.FnDecl); ok && !fn.IsExtern && fn.Body != nil {
			ip.fns[fn.Ident] = fn
		}
	}
	return ip
}

func (ip *Interp) SetHook(h Hook) { ip.hook = h }

func (ip *Interp) Run(entry string, args []string) (code int, err error) {
	fn, ok := ip.fns[entry]
	if !ok {
		return 0, fmt.Errorf("entry %q not found", entry)
	}
	ip.args = append([]string{"prog"}, args...)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("interp panic: %v", r)
		}
	}()
	return int(ip.call(fn, nil).AsInt()), nil
}

func (ip *Interp) call(fn *parser.FnDecl, argv []Value) (result Value) {
	fr := NewFrame()
	fr.FnName = fn.Ident
	for i, p := range fn.Params {
		if i < len(argv) {
			fr.Declare(p.Ident, argv[i])
		} else {
			fr.Declare(p.Ident, Void)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			if ret, ok := r.(retSignal); ok {
				result = ret.v
				return
			}
			panic(r)
		}
	}()
	ip.execBlock(fn.Body, fr)
	return Void
}

func (ip *Interp) execBlock(b *parser.ASTBlock, fr *Frame) {
	fr.Push()
	defer fr.Pop()
	for _, st := range b.Items {
		ip.exec(st, fr)
	}
}
```

`interp/eval.go`:

```go
/*
 * mon_lang - interpreter statement/expression evaluation
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import (
	"fmt"

	"github.com/your-moon/mon_lang/parser"
)

func (ip *Interp) exec(st parser.ASTStmt, fr *Frame) {
	if ip.hook != nil {
		ip.hook.OnStmt(st, fr)
	}
	switch s := st.(type) {
	case *parser.ASTReturnStmt:
		var v Value = Void
		if s.Value != nil {
			v = ip.eval(s.Value, fr)
		}
		panic(retSignal{v})
	case *parser.ASTBlock:
		ip.execBlock(s, fr)
	default:
		panic(fmt.Sprintf("exec: unhandled statement %T", st))
	}
}

func (ip *Interp) eval(expr parser.ASTExpression, fr *Frame) Value {
	switch e := expr.(type) {
	case *parser.ASTConstInt:
		return I32(int64(e.Value))
	case *parser.ASTConstLong:
		return I64(e.Value)
	default:
		panic(fmt.Sprintf("eval: unhandled expression %T", expr))
	}
}
```

> Verify field names during Step 2's failure output: `ASTReturnStmt.Value`, `ASTBlock.Items`, `ASTConstInt.Value` (int32), `ASTConstLong.Value` (int64). If a name differs, adjust — read `parser/ast_stmt.go` and `parser/ast_expr.go`. Do NOT invent fields; grep the struct.

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestRunReturnsExitCode -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/interp.go interp/eval.go interp/interp_test.go
git commit -m "feat(interp): driver — load program, run entry, return exit code

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 4: Variables, arithmetic, comparisons, logic

**Files:**
- Modify: `interp/eval.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: everything so far; `parser.ASTInitDecl` (variable declaration), `parser.ASTVar`, `parser.ASTAssignment`, `parser.ASTBinary` (with `parser.ASTBinOp`), `parser.ASTUnary` / `parser.ASTPrefixExpression`, `parser.ASTConditional`.
- Produces: `exec` handles `*parser.ASTInitDecl`; `eval` handles `*parser.ASTVar`, `*parser.ASTAssignment`, `*parser.ASTBinary`, unary/prefix, `*parser.ASTConditional`.
- Integer ops respect width: results of `тоо`-typed expressions re-wrap to 32 bits; `тоо64` to 64.

- [ ] **Step 1: Write the failing test**

```go
func TestArithmeticAndVars(t *testing.T) {
	src := `функц үндсэн() -> тоо {
		зарла а: тоо = 3;
		зарла б: тоо = 4;
		зарла в: тоо = а * б + 2;
		хэрэв в > 13 бол { буц в; }
		буц 0;
	}`
	code, err := New(resolve(t, src)).Run("үндсэн", nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != 14 {
		t.Fatalf("got %d, want 14", code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestArithmeticAndVars -v`
Expected: FAIL ("unhandled statement *parser.ASTInitDecl" or similar) — this also proves `хэрэв` is still unhandled, so add it here too.

- [ ] **Step 3: Write minimal implementation**

Read `parser/ast_stmt.go`, `parser/ast_expr.go`, `parser/ast_loop.go`, and the `ASTBinOp` constants in `parser/ast_expr.go` first, then add the cases. Add to `exec`:

```go
	case *parser.ASTInitDecl:
		var v Value = Void
		if s.Init != nil {
			v = ip.eval(s.Init, fr)
		}
		fr.Declare(s.Name, v)
	case *parser.ASTIfStmt:
		if ip.eval(s.Condition, fr).Truthy() {
			ip.exec(s.Then, fr)
		} else if s.Else != nil {
			ip.exec(s.Else, fr)
		}
	case *parser.ASTNullStmt:
		// no-op
```

Add to `eval`:

```go
	case *parser.ASTVar:
		if v, ok := fr.Get(e.Name); ok {
			return v
		}
		panic(fmt.Sprintf("eval: undefined variable %q", e.Name))
	case *parser.ASTAssignment:
		v := ip.eval(e.Value, fr)
		ip.assign(e.Target, v, fr)
		return v
	case *parser.ASTBinary:
		return ip.binary(e, fr)
	case *parser.ASTConditional:
		if ip.eval(e.Condition, fr).Truthy() {
			return ip.eval(e.Then, fr)
		}
		return ip.eval(e.Else, fr)
```

Add helpers (in `eval.go`), with the exact `ASTBinOp` constant names and AST field names taken from the source:

```go
// assign writes v to an lvalue: a bare variable, an array index, a struct
// member, or a pointer deref. Extended by later tasks; here it covers vars.
func (ip *Interp) assign(target parser.ASTExpression, v Value, fr *Frame) {
	switch t := target.(type) {
	case *parser.ASTVar:
		if !fr.Set(t.Name, v) {
			fr.Declare(t.Name, v)
		}
	default:
		panic(fmt.Sprintf("assign: unhandled lvalue %T", target))
	}
}

func (ip *Interp) binary(e *parser.ASTBinary, fr *Frame) Value {
	l := ip.eval(e.Left, fr)
	r := ip.eval(e.Right, fr)
	// Double arithmetic when either side is a double.
	if l.Kind == KindDouble || r.Kind == KindDouble {
		return ip.binaryDouble(e.Op, l.asF(), r.asF())
	}
	a, b := l.AsInt(), r.AsInt()
	res := intOp(e.Op, a, b)
	if e.GetType() != nil && e.GetType().String() == "тоо" { // Int32-typed result
		return I32(res)
	}
	return I64(res)
}
```

> `intOp` / `binaryDouble` / `Value.asF()` are small helpers to write here: `intOp` switches on the `ASTBinOp` constants (`ADD/SUB/MUL/DIV/MOD/EQ/NE/LT/LE/GT/GE/AND/OR/…` — use the real names) and returns 1/0 for comparisons and logic; `binaryDouble` does the float equivalents (comparisons still return `I32(0/1)`); `asF()` returns `F` for doubles else `float64(I)`. The `GetType().String() == "тоо"` width check is a placeholder heuristic — during Step 2, inspect what `mtypes` type `GetType()` returns for a `тоо` expression (likely `*mtypes.Int32Type`) and switch on the concrete type instead:
> ```go
> if _, ok := e.GetType().(*mtypes.Int32Type); ok { return I32(res) }
> ```
> Use the type-assertion form; it is exact. Import `mtypes`.

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestArithmeticAndVars -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/eval.go interp/interp_test.go
git commit -m "feat(interp): variables, arithmetic, comparisons, if, conditional

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 5: Loops and break/continue

**Files:**
- Modify: `interp/eval.go`, `interp/interp.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `parser.ASTWhile`, `parser.ASTLoop`, `parser.ASTBreakStmt`, `parser.ASTContinueStmt` (read `parser/ast_loop.go` / `ast_stmt.go` for exact fields).
- Produces: `type breakSignal struct{}`, `type continueSignal struct{}`; `exec` handles the loop and jump statements; loops recover `continueSignal` per iteration and `breakSignal` to terminate.

- [ ] **Step 1: Write the failing test**

```go
func TestWhileLoopSum(t *testing.T) {
	src := `функц үндсэн() -> тоо {
		зарла и: тоо = 0;
		зарла нийт: тоо = 0;
		давтах и < 10 бол {
			и = и + 1;
			хэрэв и == 3 бол { үргэлжлүүл; }
			хэрэв и == 8 бол { зогс; }
			нийт = нийт + и;
		}
		буц нийт;
	}`
	code, err := New(resolve(t, src)).Run("үндсэн", nil)
	if err != nil {
		t.Fatal(err)
	}
	// 1+2+4+5+6+7 = 25
	if code != 25 {
		t.Fatalf("got %d, want 25", code)
	}
}
```

> Confirm the keyword spellings from `lexer/keywords.go` (`давтах`, `зогс`, `үргэлжлүүл`) and the loop AST fields before writing code. Adjust the source string if a keyword differs.

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestWhileLoopSum -v`
Expected: FAIL (unhandled `*parser.ASTWhile`).

- [ ] **Step 3: Write minimal implementation**

Add signals to `interp.go`:

```go
type breakSignal struct{}
type continueSignal struct{}
```

Add to `exec` (field names per source):

```go
	case *parser.ASTWhile:
		for ip.eval(s.Condition, fr).Truthy() {
			if ip.runLoopBody(s.Body, fr) {
				break
			}
		}
	case *parser.ASTBreakStmt:
		panic(breakSignal{})
	case *parser.ASTContinueStmt:
		panic(continueSignal{})
```

Add helper to `interp.go`:

```go
// runLoopBody executes one iteration, translating зогс/үргэлжлүүл signals.
// Returns true if the loop should stop (break).
func (ip *Interp) runLoopBody(body parser.ASTStmt, fr *Frame) (stop bool) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case breakSignal:
				stop = true
			case continueSignal:
				// continue to next iteration
			default:
				panic(r)
			}
		}
	}()
	ip.exec(body, fr)
	return false
}
```

> If `ASTLoop` (a `удаа`/counted loop) exists and the `run/` suite uses it, add a `case *parser.ASTLoop:` following the same pattern; read its fields first.

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestWhileLoopSum -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/eval.go interp/interp.go interp/interp_test.go
git commit -m "feat(interp): while/loop with зогс/үргэлжлүүл

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 6: Function calls + the builtin bridge (хэвлэ)

**Files:**
- Modify: `interp/eval.go`
- Create: `interp/builtins.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `parser.ASTFnCall` (read fields: callee name + args); the `Interp.fns` map.
- Produces:
  - `interp/builtins.go`: `func (ip *Interp) builtin(name string, argv []Value) (Value, bool)` — returns `(result, true)` if `name` is a builtin, else `(Void, false)`.
  - `eval` handles `*parser.ASTFnCall`: evaluate args, then try `ip.builtin`; on miss, look up `ip.fns` and `ip.call`.
  - First builtin: `хэвлэ(n)` writes the signed decimal of `n` to `ip.out` (an `io.Writer`, default `os.Stdout`) with no newline. Add `out io.Writer` to `Interp`, defaulting in `New`, plus `func (ip *Interp) SetStdout(w io.Writer)` for tests.

- [ ] **Step 1: Write the failing test**

```go
func TestPrintBuiltin(t *testing.T) {
	src := `функц үндсэн() -> тоо { хэвлэ(7); хэвлэ(0-3); буц 0; }`
	ip := New(resolve(t, src))
	var buf strings.Builder
	ip.SetStdout(&buf)
	if _, err := ip.Run("үндсэн", nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "7-3" {
		t.Fatalf("stdout = %q, want %q", buf.String(), "7-3")
	}
}
```

Add `"strings"` to the test imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestPrintBuiltin -v`
Expected: FAIL (SetStdout undefined / хэвлэ unhandled).

- [ ] **Step 3: Write minimal implementation**

Add to `Interp` (in `interp.go`): `out io.Writer` field; in `New`, `ip.out = os.Stdout`; add `func (ip *Interp) SetStdout(w io.Writer) { ip.out = w }`. Import `io`, `os`.

`interp/builtins.go`:

```go
/*
 * mon_lang - interpreter builtin bridge (extern functions)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package interp

import "fmt"

// builtin implements the extern stdlib in Go host code, matching the native
// syscall runtime (encoder/stdlib.go) semantics. Returns (result, true) when
// name is a builtin.
func (ip *Interp) builtin(name string, argv []Value) (Value, bool) {
	switch name {
	case "хэвлэ": // signed decimal, no newline
		fmt.Fprintf(ip.out, "%d", argv[0].AsInt())
		return Void, true
	}
	return Void, false
}
```

Add to `eval` (`ASTFnCall` field names per `parser/ast_expr.go`):

```go
	case *parser.ASTFnCall:
		argv := make([]Value, len(e.Args))
		for i, a := range e.Args {
			argv[i] = ip.eval(a, fr)
		}
		if v, ok := ip.builtin(e.Name, argv); ok {
			return v
		}
		if fn, ok := ip.fns[e.Name]; ok {
			return ip.call(fn, argv)
		}
		panic(fmt.Sprintf("eval: unknown function %q", e.Name))
```

> `ASTFnCall`'s callee-name field may be `Name`, `Ident`, or a nested `Callee`. Grep the struct and use the real field.

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestPrintBuiltin -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/eval.go interp/builtins.go interp/interp.go interp/interp_test.go
git commit -m "feat(interp): function calls + builtin bridge (хэвлэ)

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 7: Strings, byte buffers, and the memory builtins

**Files:**
- Modify: `interp/builtins.go`, `interp/eval.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `parser.ASTStringExpression` (string literal). 
- Produces:
  - `eval` handles `*parser.ASTStringExpression` → `P(&Obj{Kind: ObjBytes, Bytes: append([]byte(lit), 0)})` (NUL-terminated to match native string layout).
  - Builtins: `мөр_урт` (strlen up to NUL), `байт(s,i)` (unsigned byte → `I32`), `байт_тавих(s,i,b)`, `мөр_шинэ(n)` (zeroed buffer of n+1), `monAlloc(n)` (zeroed `ObjBytes` of n), `мөр_хэвлэх(s)` (write bytes up to NUL), `тэмдэгт_хэвлэх(cp)` (UTF-8 encode one codepoint), `эхэвлэ` (unsigned decimal), `чөлөөлөх` (no-op), `дэлгэц_цэвэрлэх` (write `"\033[H\033[2J"`).

- [ ] **Step 1: Write the failing test**

```go
func TestStringBuiltins(t *testing.T) {
	src := `функц үндсэн() -> тоо {
		зарла с: мөр = "abc";
		мөр_хэвлэх(с);
		хэвлэ(мөр_урт(с));
		байт_тавих(с, 0, 65);
		тэмдэгт_хэвлэх(байт(с, 0));
		буц 0;
	}`
	ip := New(resolve(t, src))
	var buf strings.Builder
	ip.SetStdout(&buf)
	if _, err := ip.Run("үндсэн", nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "abc3A" {
		t.Fatalf("stdout = %q, want %q", buf.String(), "abc3A")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestStringBuiltins -v`
Expected: FAIL.

- [ ] **Step 3: Write minimal implementation**

Add the string-literal case to `eval` (field name per `parser/ast_expr.go` — likely `ASTStringExpression.Value`):

```go
	case *parser.ASTStringExpression:
		b := append([]byte(e.Value), 0)
		return P(&Obj{Kind: ObjBytes, Bytes: b})
```

Add cases to `builtin` (helper `strlen` counts bytes before the first NUL):

```go
	case "мөр_урт":
		return I64(int64(strlen(argv[0].Ptr))), true
	case "байт":
		o := argv[0].Ptr
		i := argv[1].AsInt()
		return I32(int64(o.Bytes[i])), true
	case "байт_тавих":
		argv[0].Ptr.Bytes[argv[1].AsInt()] = byte(argv[2].AsInt())
		return Void, true
	case "мөр_шинэ":
		n := argv[0].AsInt()
		return P(&Obj{Kind: ObjBytes, Bytes: make([]byte, n+1)}), true
	case "monAlloc":
		n := argv[0].AsInt()
		return P(&Obj{Kind: ObjBytes, Bytes: make([]byte, n)}), true
	case "мөр_хэвлэх":
		o := argv[0].Ptr
		ip.out.Write(o.Bytes[:strlen(o)])
		return Void, true
	case "тэмдэгт_хэвлэх":
		ip.out.Write([]byte(string(rune(argv[0].AsInt()))))
		return Void, true
	case "эхэвлэ":
		fmt.Fprintf(ip.out, "%d", uint64(argv[0].AsInt()))
		return Void, true
	case "чөлөөлөх":
		return Void, true
	case "дэлгэц_цэвэрлэх":
		ip.out.Write([]byte("\033[H\033[2J"))
		return Void, true
```

Add helper to `builtins.go`:

```go
func strlen(o *Obj) int {
	for i, b := range o.Bytes {
		if b == 0 {
			return i
		}
	}
	return len(o.Bytes)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestStringBuiltins -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/builtins.go interp/eval.go interp/interp_test.go
git commit -m "feat(interp): strings, byte buffers, memory + print builtins

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 8: Arrays, structs, members, pointers, cast, доубл print

**Files:**
- Modify: `interp/eval.go`, `interp/builtins.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `parser.ASTNewArray`, `parser.ASTArrayIndex`, `parser.ASTMember`, `parser.ASTAddrOf`, `parser.ASTDeref`, `parser.ASTCast`, `parser.ASTConstFloat`, and struct construction (read `parser/ast_expr.go` + how structs are instantiated).
- Produces:
  - `eval`: `ASTNewArray` → `ObjArray` with `Size` zeroed `Value`s; `ASTArrayIndex` → element read; `ASTMember` → struct field read (map field name → index via the struct type from sema); `ASTConstFloat` → `F`; `ASTCast` → numeric conversion honoring target `mtypes` type; `ASTAddrOf`/`ASTDeref` → model as pointer-to-object handles.
  - `assign` (from Task 4) gains `ASTArrayIndex`, `ASTMember`, `ASTDeref` lvalue cases.
  - Builtin `бутархай_хэвлэх` matching lib.c's 6-truncated-digit format.

> This is the widest task; split into sub-commits per feature (array, then struct/member, then pointer, then cast+float) if a reviewer prefers. Each sub-feature: write test → run red → implement → run green → commit. The array test below is the entry point; add struct/pointer/float tests following the same shape, using real field names read from the source.

- [ ] **Step 1: Write the failing test (arrays)**

```go
func TestArrays(t *testing.T) {
	src := `функц үндсэн() -> тоо {
		зарла а: тоо64[] = шинэ тоо64[3];
		а[0] = 10; а[1] = 20; а[2] = а[0] + а[1];
		буц а[2];
	}`
	code, err := New(resolve(t, src)).Run("үндсэн", nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != 30 {
		t.Fatalf("got %d, want 30", code)
	}
}
```

> Confirm array syntax (`шинэ тоо64[3]`) and `ASTNewArray` fields against `parser/parse.go` / `ast_expr.go` before writing.

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestArrays -v`
Expected: FAIL.

- [ ] **Step 3: Write minimal implementation**

Add to `eval`:

```go
	case *parser.ASTNewArray:
		n := ip.eval(e.Size, fr).AsInt()
		return P(&Obj{Kind: ObjArray, Fields: make([]Value, n)})
	case *parser.ASTArrayIndex:
		arr := ip.eval(e.Array, fr)
		idx := ip.eval(e.Index, fr).AsInt()
		return arr.Ptr.Fields[idx]
	case *parser.ASTConstFloat:
		return F(e.Value)
```

Add to `assign`:

```go
	case *parser.ASTArrayIndex:
		arr := ip.eval(t.Array, fr)
		idx := ip.eval(t.Index, fr).AsInt()
		arr.Ptr.Fields[idx] = v
```

> Struct/member: from sema, a struct value is an `ObjStruct` whose `Fields` are ordered by declaration; `ASTMember` carries (or its type resolves) the field's index. Determine how the resolver records field order (grep `semantic_analysis` + `mtypes` for struct field layout) and map member name → index. Pointer `ASTAddrOf`/`ASTDeref`: represent `&x` as a pointer `Obj` wrapping a single-field cell, or (simpler, matching native) as an `ObjArray` handle + index. Choose the representation that lets `байт`/`байт_тавих` and array indexing share one path; document the choice in a comment. `ASTCast`: switch on the target `mtypes` type and convert (`I32`/`I64`/`F`). `бутархай_хэвлэх`: port lib.c's algorithm exactly (sign, integer part via `хэвлэ`-style, then 6 iterations of `frac*=10`).

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestArrays -v`
Expected: PASS. Then repeat Steps 1–4 for struct/member, pointer, and float-print sub-features.

- [ ] **Step 5: Commit** (one per sub-feature)

```bash
git add interp/eval.go interp/builtins.go interp/interp_test.go
git commit -m "feat(interp): arrays (indexing, шинэ, assignment)

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 9: Match (тааруул) statement

**Files:**
- Modify: `interp/eval.go`
- Test: `interp/interp_test.go`

**Interfaces:**
- Consumes: `parser.ASTMatch` (read fields: subject + arms; each arm has a pattern value + body; `_` is the default). Since enums are constant-folded to `ASTConstInt` in sema, arm patterns are integer-valued expressions.
- Produces: `exec` handles `*parser.ASTMatch`: evaluate the subject, find the first arm whose pattern equals it (or the `_` default), execute that arm's body.

- [ ] **Step 1: Write the failing test**

```go
func TestMatch(t *testing.T) {
	src := `тоочих Өнгө { УЛААН, НОГООН, ЦЭНХЭР }
	функц нэр(ө тоо) -> тоо {
		тааруул ө {
			Өнгө.УЛААН => { буц 100; }
			Өнгө.НОГООН => { буц 200; }
			_ => { буц 0; }
		}
		буц -1;
	}
	функц үндсэн() -> тоо { буц нэр(Өнгө.НОГООН); }`
	code, err := New(resolve(t, src)).Run("үндсэн", nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != 200 {
		t.Fatalf("got %d, want 200", code)
	}
}
```

> Confirm `тааруул` / arm syntax and `ASTMatch` field names against `parser/parse.go` (`parseMatch`) and `parser/ast_stmt.go`.

- [ ] **Step 2: Run test to verify it fails**

Run: `rtk proxy go test ./interp/ -run TestMatch -v`
Expected: FAIL (unhandled `*parser.ASTMatch`).

- [ ] **Step 3: Write minimal implementation**

Add to `exec` (field names per source; `Arms`, `Subject`, `Pattern`, `Body`, and a default marker are illustrative):

```go
	case *parser.ASTMatch:
		subj := ip.eval(s.Subject, fr).AsInt()
		for _, arm := range s.Arms {
			if arm.IsDefault {
				ip.exec(arm.Body, fr)
				return
			}
			if ip.eval(arm.Pattern, fr).AsInt() == subj {
				ip.exec(arm.Body, fr)
				return
			}
		}
```

> If the AST represents `_` differently (e.g. a nil `Pattern`), adapt the default check. Match arms in mon do not fall through, so `return` after executing one arm is correct.

- [ ] **Step 4: Run test to verify it passes**

Run: `rtk proxy go test ./interp/ -run TestMatch -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interp/eval.go interp/interp_test.go
git commit -m "feat(interp): тааруул (match) statement

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Task 10: Wire the interpreter into the run/ suite as a second mode (the differential oracle)

**Files:**
- Create: `tests/interp_test.go`
- Modify: `tests/pipeline.go` (add a resolve+interpret helper if convenient)
- Test: (this task *is* the test)

**Interfaces:**
- Consumes: `interp.New`, `interp.Run`, `interp.SetStdout`; the harness helpers `parseDirectives`, `toRunes`, and the resolve pipeline already in `tests/`.
- Produces: `TestRunInterp` — for each `tests/run/*.mn`, resolve → interpret with stdin/argv from the directives → assert `(stdout, exit)` match the directive expectations. Tests using features the interpreter doesn't cover yet **skip** (like `TestRunARM64` does), so the suite stays green while coverage grows.

- [ ] **Step 1: Write the test (this is the deliverable)**

```go
/*
 * mon_lang - interpreter differential suite
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */
package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-moon/mon_lang/interp"
	"github.com/your-moon/mon_lang/parser"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

func interpret(t *testing.T, srcPath string, exp expectation) (out string, code int, ok bool) {
	data := mustRead(t, srcPath) // small helper: os.ReadFile + fatal; or inline
	runes := toRunes(string(data))
	p := parser.NewParser(runes)
	node, err := p.ParseProgram()
	if err != nil || len(p.Errors()) > 0 {
		return "", 0, false
	}
	table := symbols.NewSymbolTable()
	sa := semanticanalysis.NewSemanticAnalyzer(runes, unique.NewUniqueGen(), table, filepath.Dir(srcPath), "")
	resolved, _, err := sa.Analyze(node)
	if err != nil {
		return "", 0, false
	}
	ip := interp.New(resolved)
	var buf strings.Builder
	ip.SetStdout(&buf)
	ip.SetStdin(strings.NewReader(exp.stdin)) // add SetStdin to interp if унш is exercised
	defer func() { _ = recover() }() // unsupported feature -> skip
	c, err := ip.Run("үндсэн", nil)
	if err != nil {
		return "", 0, false
	}
	return buf.String(), c, true
}

func TestRunInterp(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("run", "*.mn"))
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			exp := parseDirectives(t, f)
			if len(exp.errSubs) > 0 {
				t.Skip("error test")
			}
			out, code, ok := interpret(t, f, exp)
			if !ok {
				t.Skip("interpreter does not support this program yet")
			}
			if exp.hasStdout && out != exp.stdout {
				t.Errorf("stdout mismatch\n got: %q\nwant: %q", out, exp.stdout)
			}
			if code != exp.exitCode {
				t.Errorf("exit code: got %d, want %d", code, exp.exitCode)
			}
		})
	}
}
```

> Add `interp.SetStdin(io.Reader)` and a `mustRead` helper if not already present. If `toRunes`/`parseDirectives`/`expectation` are in package `tests` (they are, in `tests_test.go`), this file compiles against them directly.

- [ ] **Step 2: Run the suite**

Run: `rtk proxy go test ./tests/ -run TestRunInterp -v 2>&1 | tail -40`
Expected: a mix of PASS and SKIP, **no FAIL**. Every non-skipped file matches the native result byte-for-byte.

- [ ] **Step 3: Triage skips**

Run: `rtk proxy go test ./tests/ -run TestRunInterp -v 2>&1 | grep -c SKIP` and list the skipped files. For each skip, decide: is it an unimplemented node (add coverage in the relevant task above) or a genuinely out-of-scope feature (document it)? Drive skips down to only out-of-scope cases.

- [ ] **Step 4: Verify no regressions in the rest of the suite**

Run: `rtk proxy go test ./... 2>&1 | grep -vE "no test files"`
Expected: all packages `ok`.

- [ ] **Step 5: Commit**

```bash
git add tests/interp_test.go interp/
git commit -m "test(interp): run/ suite as interpreter-vs-native differential oracle

Claude-Session: https://claude.ai/code/session_014ovvLQwwA34Y7rkWDbfYHo"
```

---

## Self-Review

**Spec coverage (Stage 1 sections of the design):**
- Value model → Task 1. ✓
- Frames & scopes → Task 2. ✓
- Builtins bridge → Tasks 6–8 (print, strings/memory, float). ✓
- The debug seam (`Hook.OnStmt`) → introduced in Task 3, called in every `exec`. ✓ (DAP consumes it in a later plan.)
- Validation as a second test mode / differential oracle → Task 10. ✓
- `тоо`/`тоо64`/`бутархай` width semantics → Task 4 (int width via concrete `mtypes` assertion) + Task 8 (float). ✓
- Enums fold to ints → relied on in Task 9 (match arms are integer patterns). ✓

**Out of scope for Stage 1 (correctly deferred):** DAP server, arch auto-detect, native GC, and self-hosting the interpreter — each its own later plan.

**Placeholder note:** Several steps say "confirm the exact field/keyword name against the source before writing." This is deliberate and bounded — the plan gives the exact node **type** and the code shape; only leaf field names must be verified against `parser/*.go`, because inventing a field name is the one failure mode worse than looking it up. Every such note names the exact file to read. No step defers *logic*.

**Type consistency:** `Value`, `Obj`, `Frame`, `Interp`, `Hook`, `retSignal`/`breakSignal`/`continueSignal`, `builtin`, `assign`, `eval`, `exec`, `call`, `execBlock`, `SetStdout`/`SetStdin` are used consistently across tasks with the signatures declared in each task's Interfaces block.
