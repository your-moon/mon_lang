# Mon_lang Compiler Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix critical compiler bugs (control flow, type coercion), add memory management (`чөлөөлөх`), multi-error reporting, global mutable variables, import system, and a Go test suite.

**Architecture:** Changes span all 7 compiler stages. Control flow and type coercion fixes are in tackygen. Global variables add `.data` section support in codegen. Imports expand the parser and semantic analyzer to handle multi-file programs. Tests cover each stage independently plus integration tests that compile and run `.mn` programs.

**Tech Stack:** Go, x86-64 assembly, libc via C wrapper

---

### Task 1: Control Flow Codegen Audit

**Files:**
- Create: `test/control_flow.mn`
- Create: `control_flow_test.go`
- Modify: `tackygen/tack_gen.go` (if bugs found)

- [ ] **Step 1: Write the control flow test program**

Create `test/control_flow.mn`:
```mn
функц үндсэн() -> тоо {
    // Test 1: while with break
    зарла i: тоо = 0;
    давтах i < 10 бол {
        хэрэв i == 3 бол {
            зогс;
        }
        i = i + 1;
    }
    хэвлэ(i);
    мөр_хэвлэх(" ");

    // Test 2: while with continue
    зарла j: тоо = 0;
    зарла сум: тоо = 0;
    давтах j < 5 бол {
        j = j + 1;
        хэрэв j == 3 бол {
            үргэлжлүүл;
        }
        сум = сум + j;
    }
    хэвлэ(сум);
    мөр_хэвлэх(" ");

    // Test 3: nested loops with inner break
    зарла а: тоо = 0;
    зарла гадна: тоо = 0;
    давтах гадна < 3 бол {
        зарла дотор: тоо = 0;
        давтах дотор < 10 бол {
            хэрэв дотор == 2 бол {
                зогс;
            }
            дотор = дотор + 1;
        }
        а = а + дотор;
        гадна = гадна + 1;
    }
    хэвлэ(а);
    мөр_хэвлэх(" ");

    // Test 4: if / else-if / else
    зарла x: тоо = 2;
    хэрэв x == 1 бол {
        хэвлэ(10);
    } эсвэл хэрэв x == 2 бол {
        хэвлэ(20);
    } эсвэл {
        хэвлэ(30);
    }
    мөр_хэвлэх(" ");

    // Test 5: ternary
    зарла t: тоо = 1 == 1 ? 99 : 0;
    хэвлэ(t);
    мөр_хэвлэх("\n");

    буц 0;
}
```

Expected output: `3 12 6 20 99\n`

- [ ] **Step 2: Write the integration test**

Create `control_flow_test.go` in the project root:
```go
package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func compileAndRun(t *testing.T, srcFile string) string {
	t.Helper()
	outFile := t.TempDir() + "/out"

	// Compile
	cmd := exec.Command("go", "run", ".", "gen", srcFile, "-o", outFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("compile failed: %v\nstderr: %s", err, stderr.String())
	}

	// Run (x86_64 on ARM mac)
	runCmd := exec.Command("arch", "-x86_64", outFile)
	var stdout bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	if err := runCmd.Run(); err != nil {
		t.Fatalf("run failed: %v\nstderr: %s", err, stderr.String())
	}

	return stdout.String()
}

func TestControlFlow(t *testing.T) {
	output := compileAndRun(t, "test/control_flow.mn")
	expected := "3 12 6 20 99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
```

- [ ] **Step 3: Run test to verify it passes (or find bugs)**

Run: `go test -run TestControlFlow -v -timeout 30s`

If any value is wrong, trace the Tacky IR for that construct: `go run . tacky test/control_flow.mn --debug 2>&1 | less`

Fix any bugs found in `tackygen/tack_gen.go` by adding missing `Jump` instructions or correcting label references.

- [ ] **Step 4: Commit**

```bash
git add test/control_flow.mn control_flow_test.go
git commit -m "test: add control flow audit test"
```

---

### Task 2: Multi-Error Reporting

**Files:**
- Modify: `parser/parse.go:69-75`
- Modify: `cli/cli.go` (error display in runParser, runValidate, runGen)
- Create: `test/validate/multi_error.mn`

- [ ] **Step 1: Write a test file with multiple errors**

Create `test/validate/multi_error.mn`:
```mn
функц үндсэн() -> тоо {
    зарла а: тоо = 1;
    зарла а: тоо = 2;
    б = 5;
    буц 0;
}
```

This should report: (1) duplicate variable `а`, (2) undeclared variable `б`.

- [ ] **Step 2: Change ParseProgram to return all errors**

Modify `parser/parse.go` lines 69-75. Replace:
```go
	if len(p.parseErrors) > 0 {
		err := p.parseErrors[0]
		p.parseErrors = nil
		return nil, err
	}

	return program, nil
```

With:
```go
	if len(p.parseErrors) > 0 {
		msgs := make([]string, len(p.parseErrors))
		for i, e := range p.parseErrors {
			msgs[i] = e.Error()
		}
		p.parseErrors = nil
		return nil, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	return program, nil
```

Add `"strings"` to the imports at the top of the file.

- [ ] **Step 3: Make the resolver collect errors instead of returning early**

Modify `semantic_analysis/resolve.go` `Resolve()` function. Replace:
```go
func (r *Resolver) Resolve(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	emptyMap := make(IdMap)
	// Register builtins so they are available without extern declarations
	for name := range r.builtins {
		emptyMap[name] = VarEntry{
			UniqueName:       name,
			fromCurrentScope: true,
			hasLinkage:       true,
		}
	}
	for i, decl := range program.Decls {
		_, decl, err := r.ResolveDecl(decl, emptyMap)
		if err != nil {
			return nil, err
		}
		program.Decls[i] = decl
	}

	return program, nil
}
```

With:
```go
func (r *Resolver) Resolve(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	emptyMap := make(IdMap)
	for name := range r.builtins {
		emptyMap[name] = VarEntry{
			UniqueName:       name,
			fromCurrentScope: true,
			hasLinkage:       true,
		}
	}
	var allErrors []error
	for i, decl := range program.Decls {
		newMap, resolved, err := r.ResolveDecl(decl, emptyMap)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		emptyMap = newMap
		program.Decls[i] = resolved
	}

	if len(allErrors) > 0 {
		msgs := make([]string, len(allErrors))
		for i, e := range allErrors {
			msgs[i] = e.Error()
		}
		return nil, fmt.Errorf("%s", strings.Join(msgs, "\n"))
	}

	return program, nil
}
```

Add `"strings"` to the imports.

- [ ] **Step 4: Test multi-error output**

Run: `go run . validate test/validate/multi_error.mn 2>&1`

Expected: Output should contain both error messages (duplicate variable and undeclared variable). If only one shows, check which pass swallows the second.

- [ ] **Step 5: Commit**

```bash
git add parser/parse.go semantic_analysis/resolve.go test/validate/multi_error.mn
git commit -m "feat: report multiple errors instead of stopping at first"
```

---

### Task 3: Fix 32/64-bit Type Coercion

**Files:**
- Modify: `tackygen/tack_gen.go` (EmitExpr for ASTBinary, ASTFnCall, EmitVarDecl, ASTAssignment)
- Create: `test/types.mn`

- [ ] **Step 1: Write the type coercion test program**

Create `test/types.mn`:
```mn
функц тестТоо64(н тоо64) -> тоо64 {
    буц н + 1;
}

функц үндсэн() -> тоо {
    // Test 1: тоо64 literal arithmetic
    зарла а: тоо64 = 100;
    зарла б: тоо64 = 200;
    хэвлэ(а + б);
    мөр_хэвлэх(" ");

    // Test 2: mixed тоо + тоо64 arithmetic
    зарла в: тоо = 10;
    зарла г: тоо64 = 20;
    хэвлэ(в + г);
    мөр_хэвлэх(" ");

    // Test 3: тоо assigned to тоо64 var
    зарла д: тоо64 = 42;
    хэвлэ(д);
    мөр_хэвлэх(" ");

    // Test 4: тоо passed to тоо64 param
    зарла е: тоо = 99;
    хэвлэ(тестТоо64(е));
    мөр_хэвлэх("\n");

    буц 0;
}
```

Expected output: `300 30 42 100\n`

- [ ] **Step 2: Add maybeSignExtend helper to tack_gen.go**

Add this helper function after the `makeLabel` function (around line 669 in `tackygen/tack_gen.go`):
```go
func (c *TackyGen) maybeSignExtend(val TackyVal, fromType, toType mtypes.Type) (TackyVal, []Instruction) {
	_, fromIs32 := fromType.(*mtypes.Int32Type)
	_, toIs64 := toType.(*mtypes.Int64Type)
	if fromIs32 && toIs64 {
		dst := c.makeTemp(&mtypes.Int64Type{})
		return dst, []Instruction{SignExtend{Src: val, Dst: dst}}
	}
	return val, nil
}
```

- [ ] **Step 3: Fix EmitExpr for ASTBinary**

In `tackygen/tack_gen.go`, in the `EmitExpr` function's `*parser.ASTBinary` case (the else branch with normal binary ops), after emitting v1 and v2 but before the Binary instruction, add sign extension. Replace:

```go
		v1, v1Irs := c.EmitExpr(expr.Left)
		irs = append(irs, v1Irs...)
		v2, v2Irs := c.EmitExpr(expr.Right)
		irs = append(irs, v2Irs...)
		dst := c.makeTemp(expr.Type)
		instr := Binary{
			Op:   op,
			Src1: v1,
			Src2: v2,
			Dst:  dst,
		}
		irs = append(irs, instr)
		return dst, irs
```

With:
```go
		v1, v1Irs := c.EmitExpr(expr.Left)
		irs = append(irs, v1Irs...)
		v2, v2Irs := c.EmitExpr(expr.Right)
		irs = append(irs, v2Irs...)

		// Sign-extend if mixed 32/64-bit
		_, commonIs64 := expr.Type.(*mtypes.Int64Type)
		if commonIs64 {
			v1ext, v1extIrs := c.maybeSignExtend(v1, expr.Left.GetType(), expr.Type)
			irs = append(irs, v1extIrs...)
			v1 = v1ext
			v2ext, v2extIrs := c.maybeSignExtend(v2, expr.Right.GetType(), expr.Type)
			irs = append(irs, v2extIrs...)
			v2 = v2ext
		}

		dst := c.makeTemp(expr.Type)
		instr := Binary{
			Op:   op,
			Src1: v1,
			Src2: v2,
			Dst:  dst,
		}
		irs = append(irs, instr)
		return dst, irs
```

- [ ] **Step 4: Fix EmitVarDecl for mixed types**

In `tackygen/tack_gen.go`, replace `EmitVarDecl`:
```go
func (c *TackyGen) EmitVarDecl(node *parser.VarDecl) []Instruction {
	irs := []Instruction{}
	haveInit := node.Expr != nil
	if haveInit {
		rhsResult, rhsValIrs := c.EmitExpr(node.Expr)
		irs = append(irs, rhsValIrs...)

		// Sign-extend if assigning тоо to тоо64 variable
		if node.VarType != nil && node.Expr.GetType() != nil {
			extended, extIrs := c.maybeSignExtend(rhsResult, node.Expr.GetType(), node.VarType)
			irs = append(irs, extIrs...)
			rhsResult = extended
		}

		irs = append(irs, Copy{Src: rhsResult, Dst: Var{Name: node.Ident}})
	}
	return irs
}
```

- [ ] **Step 5: Add integration test**

Add to `control_flow_test.go`:
```go
func TestTypeCoercion(t *testing.T) {
	output := compileAndRun(t, "test/types.mn")
	expected := "300 30 42 100\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
```

- [ ] **Step 6: Run tests**

Run: `go test -run TestTypeCoercion -v -timeout 30s`

- [ ] **Step 7: Commit**

```bash
git add tackygen/tack_gen.go test/types.mn control_flow_test.go
git commit -m "fix: proper 32/64-bit type coercion with SignExtend"
```

---

### Task 4: Add `чөлөөлөх` (free)

**Files:**
- Modify: `semantic_analysis/sematic_analyze.go:25-44` (registerImplicitStdlib)
- Modify: `tackygen/tack_gen.go:36` (implicit externs list)
- Modify: `stdlib/lib.c`
- Create: `test/free.mn`

- [ ] **Step 1: Add `чөлөөлөх` to implicit stdlib**

In `semantic_analysis/sematic_analyze.go`, add to the `stdlibFns` slice in `registerImplicitStdlib()`:
```go
		{"чөлөөлөх", &mtypes.VoidType{}},
```

- [ ] **Step 2: Add to implicit externs in Tacky gen**

In `tackygen/tack_gen.go`, add `"чөлөөлөх"` to the `implicitExterns` slice:
```go
	implicitExterns := []string{"хэвлэ", "мөр_хэвлэх", "унш", "унш32", "санамсаргүйТоо", "одоо", "malloc", "чөлөөлөх"}
```

- [ ] **Step 3: Add C implementation**

In `stdlib/lib.c`, add at the bottom:
```c
// чөлөөлөх - free heap memory
void chqlqqlqkh(void *p) {
    free(p);
}
```

- [ ] **Step 4: Write test program**

Create `test/free.mn`:
```mn
функц үндсэн() -> тоо {
    зарла а: тоо[] = шинэ тоо[10];
    а[0] = 42;
    хэвлэ(а[0]);
    чөлөөлөх(а);
    мөр_хэвлэх("\n");
    буц 0;
}
```

Expected output: `42\n`

- [ ] **Step 5: Add integration test**

Add to `control_flow_test.go`:
```go
func TestFree(t *testing.T) {
	output := compileAndRun(t, "test/free.mn")
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
```

- [ ] **Step 6: Run tests**

Run: `go test -run TestFree -v -timeout 30s`

- [ ] **Step 7: Commit**

```bash
git add semantic_analysis/sematic_analyze.go tackygen/tack_gen.go stdlib/lib.c test/free.mn control_flow_test.go
git commit -m "feat: add чөлөөлөх (free) for manual memory management"
```

---

### Task 5: Global Mutable Variables

**Files:**
- Modify: `tackygen/tacky.go` (add GlobalVar type)
- Modify: `tackygen/tack_gen.go` (emit globals, change var reads/writes for globals)
- Modify: `code_gen/asm_ast.go` (add RipRelative operand)
- Modify: `code_gen/emitter.go` (handle globals)
- Modify: `code_gen/pseudo.go` (skip globals in stack allocation)
- Modify: `code_gen/x86x64.go` (emit .data entries, RIP-relative addressing)
- Modify: `code_gen/asmsymbol/asmsymbols.go` (track globals)
- Create: `test/globals.mn`

- [ ] **Step 1: Write the test program**

Create `test/globals.mn`:
```mn
зарла тоолуур: тоо = 0;

функц нэмэх() -> хоосон {
    тоолуур = тоолуур + 1;
}

функц үндсэн() -> тоо {
    нэмэх();
    нэмэх();
    нэмэх();
    хэвлэ(тоолуур);
    мөр_хэвлэх("\n");
    буц 0;
}
```

Expected output: `3\n`

- [ ] **Step 2: Add GlobalVar to Tacky IR**

In `tackygen/tacky.go`, add after the `TackyProgram` struct:
```go
type GlobalVar struct {
	Name      string
	InitValue int64
	Type      string // "int32" or "int64"
}
```

Add `GlobalVars` field to `TackyProgram`:
```go
type TackyProgram struct {
	FnDefs     []TackyFn
	ExternDefs []TackyFn
	GlobalVars []GlobalVar
}
```

- [ ] **Step 3: Track globals in TackyGen**

In `tackygen/tack_gen.go`, add a `MutableGlobals` set to `TackyGen`:
```go
type TackyGen struct {
	TempCount       uint64
	LabelCount      uint64
	UniqueGen       unique.UniqueGen
	SymbolTable     *symbols.SymbolTable
	GlobalConstants map[string]mconstant.Const
	MutableGlobals  map[string]bool
}
```

Initialize in `NewTackyGen`:
```go
	MutableGlobals:  make(map[string]bool),
```

- [ ] **Step 4: Emit global variables in EmitTacky**

In `tackygen/tack_gen.go`, modify the `*parser.VarDecl` case in `EmitTacky` to always emit globals as `GlobalVar` (not just constants). Replace:
```go
		case *parser.VarDecl:
			// Top-level variable declarations with constant initializers become global constants
			if stmttype.Expr != nil {
				switch constExpr := stmttype.Expr.(type) {
				case *parser.ASTConstInt:
					c.GlobalConstants[stmttype.Ident] = &mconstant.Int32{Value: int32(constExpr.Value)}
				case *parser.ASTConstLong:
					c.GlobalConstants[stmttype.Ident] = &mconstant.Int64{Value: constExpr.Value}
				}
			}
```

With:
```go
		case *parser.VarDecl:
			c.MutableGlobals[stmttype.Ident] = true
			initVal := int64(0)
			varType := "int32"
			if stmttype.Expr != nil {
				switch constExpr := stmttype.Expr.(type) {
				case *parser.ASTConstInt:
					initVal = constExpr.Value
				case *parser.ASTConstLong:
					initVal = constExpr.Value
					varType = "int64"
				}
			}
			if _, isLong := stmttype.VarType.(*mtypes.Int64Type); isLong {
				varType = "int64"
			}
			program.GlobalVars = append(program.GlobalVars, GlobalVar{
				Name:      stmttype.Ident,
				InitValue: initVal,
				Type:      varType,
			})
```

- [ ] **Step 5: Emit Load/Store for global variable reads and writes**

In `tackygen/tack_gen.go`, modify `EmitExpr` for `*parser.ASTVar`. Replace:
```go
	case *parser.ASTVar:
		// Check if this is a global constant
		if constVal, ok := c.GlobalConstants[expr.Ident]; ok {
			return Constant{Value: constVal}, []Instruction{}
		}
		return Var{Name: expr.Ident}, []Instruction{}
```

With:
```go
	case *parser.ASTVar:
		if constVal, ok := c.GlobalConstants[expr.Ident]; ok {
			return Constant{Value: constVal}, []Instruction{}
		}
		if c.MutableGlobals[expr.Ident] {
			// Global variable: load from global address
			dst := c.makeTemp(expr.Type)
			irs := []Instruction{Load{Src: Var{Name: expr.Ident}, Dst: dst}}
			return dst, irs
		}
		return Var{Name: expr.Ident}, []Instruction{}
```

Modify `EmitExpr` for `*parser.ASTAssignment` in the `*parser.ASTVar` LHS case. Replace:
```go
		case *parser.ASTVar:
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			irs = append(irs, Copy{Src: rhsResult, Dst: Var{Name: lhs.Ident}})
			return Var{Name: lhs.Ident}, irs
```

With:
```go
		case *parser.ASTVar:
			rhsResult, rhsIrs := c.EmitExpr(expr.Right)
			irs = append(irs, rhsIrs...)
			if c.MutableGlobals[lhs.Ident] {
				irs = append(irs, Store{Src: rhsResult, Dst: Var{Name: lhs.Ident}})
			} else {
				irs = append(irs, Copy{Src: rhsResult, Dst: Var{Name: lhs.Ident}})
			}
			return rhsResult, irs
```

- [ ] **Step 6: Add RipRelative operand to ASM AST**

In `code_gen/asm_ast.go`, add:
```go
type RipRelative struct {
	Label string
}

func (r RipRelative) Op() string {
	return fmt.Sprintf("%s(%%rip)", r.Label)
}
```

- [ ] **Step 7: Track globals in ASM symbol table**

In `code_gen/asmsymbol/asmsymbols.go`, add a `Global` flag to `Obj`:
```go
type Obj struct {
	Type     asmtype.AsmType
	IsStatic bool
	IsGlobal bool
}
```

Add a method:
```go
func (s *SymbolTable) AddGlobal(name string, t asmtype.AsmType) {
	s.entries[name] = &Obj{
		Type:     t,
		IsStatic: false,
		IsGlobal: true,
	}
}

func (s *SymbolTable) IsGlobal(name string) bool {
	entry, ok := s.entries[name]
	if !ok {
		return false
	}
	obj, ok := entry.(*Obj)
	if !ok {
		return false
	}
	return obj.IsGlobal
}
```

- [ ] **Step 8: Handle globals in emitter**

In `code_gen/emitter.go`, modify `GenASTAsm` to register global variables in the ASM symbol table. After the existing symbol table population loop, add:
```go
	// Register global variables
	for _, gv := range program.GlobalVars {
		var asmType asmtype.AsmType
		if gv.Type == "int64" {
			asmType = &asmtype.QuadWord{}
		} else {
			asmType = &asmtype.LongWord{}
		}
		asmSymbols.AddGlobal(gv.Name, asmType)
	}
```

Note: `GenASTAsm` signature needs to also accept `TackyProgram` global vars. Since it already receives `tackygen.TackyProgram`, the `GlobalVars` field is already available.

Modify `GenASTVal` to emit RIP-relative for globals. Replace:
```go
	case tackygen.Var:
		return Pseudo{Ident: ast.Name}
```

With:
```go
	case tackygen.Var:
		if a.asmSymbols != nil && a.asmSymbols.IsGlobal(ast.Name) {
			return RipRelative{Label: ast.Name}
		}
		return Pseudo{Ident: ast.Name}
```

This requires storing `asmSymbols` on the `AsmASTGen` struct. Add it:
```go
type AsmASTGen struct {
	Registers   []AsmRegister
	SymbolTable *symbols.SymbolTable
	asmSymbols  *asmsymbol.SymbolTable
}
```

And set it at the start of `GenASTAsm`:
```go
	a.asmSymbols = asmSymbols
```

- [ ] **Step 9: Handle RipRelative in pseudo replacement**

In `code_gen/pseudo.go`, modify `ReplaceOperand` to pass through `RipRelative` (it's not a Pseudo, so it won't match the isPseudo check and will be returned as-is). No changes needed — the existing `else` branch already returns the operand unchanged.

- [ ] **Step 10: Handle RipRelative in fixup pass**

In `code_gen/fixup.go`, `RipRelative` is a memory operand. The fixup pass needs to handle it in memory-to-memory moves. The existing `Stack` checks won't catch it. Add `RipRelative` checks alongside `Stack` checks in the `AsmMov` case. After the stack-to-stack block, add:
```go
		// Handle RipRelative as memory source/dest
		if _, isRipSrc := ast.Src.(RipRelative); isRipSrc {
			if _, isDstStack := ast.Dst.(Stack); isDstStack {
				return []AsmInstruction{
					AsmMov{Type: ast.Type, Src: ast.Src, Dst: Register{Reg: R10}},
					AsmMov{Type: ast.Type, Src: Register{Reg: R10}, Dst: ast.Dst},
				}
			}
			if _, isDstRip := ast.Dst.(RipRelative); isDstRip {
				return []AsmInstruction{
					AsmMov{Type: ast.Type, Src: ast.Src, Dst: Register{Reg: R10}},
					AsmMov{Type: ast.Type, Src: Register{Reg: R10}, Dst: ast.Dst},
				}
			}
		}
		if _, isRipDst := ast.Dst.(RipRelative); isRipDst {
			if _, isSrcStack := ast.Src.(Stack); isSrcStack {
				return []AsmInstruction{
					AsmMov{Type: ast.Type, Src: ast.Src, Dst: Register{Reg: R10}},
					AsmMov{Type: ast.Type, Src: Register{Reg: R10}, Dst: ast.Dst},
				}
			}
		}
```

- [ ] **Step 11: Emit .data section and RIP-relative in x86x64.go**

In `code_gen/x86x64.go`, modify `GenAsm` to accept global vars and emit them. The `GenAsm` method needs the global vars from the Tacky program. Change the signature or pass them separately.

Simplest approach: add a `GlobalVars` field to `AsmProgram` in `asm_ast.go`:
```go
type AsmProgram struct {
	AsmFnDef    []AsmFnDef
	AsmExternFn []AsmExternFn
	GlobalVars  []GlobalVarAsm
}

type GlobalVarAsm struct {
	Label     string
	InitValue int64
	Size      int // 4 or 8
}
```

In `emitter.go` `GenASTAsm`, populate it:
```go
	for _, gv := range program.GlobalVars {
		size := 4
		if gv.Type == "int64" {
			size = 8
		}
		asmprogram.GlobalVars = append(asmprogram.GlobalVars, GlobalVarAsm{
			Label:     gv.Name,
			InitValue: gv.InitValue,
			Size:      size,
		})
	}
```

In `x86x64.go` `GenAsm`, after `GenStringData()` and before `.text`, emit global vars in the `.data` section:
```go
	// Emit global variables in .data section
	if len(program.GlobalVars) > 0 {
		for _, gv := range program.GlobalVars {
			label := utfconvert.UtfConvert(gv.Label)
			if a.ostype == util.Darwin {
				label = "_" + label
			}
			a.Write(fmt.Sprintf("%s:", label))
			if gv.Size == 8 {
				a.Write(fmt.Sprintf("    .quad %d", gv.InitValue))
			} else {
				a.Write(fmt.Sprintf("    .long %d", gv.InitValue))
			}
		}
		a.Write("")
	}
```

Import `utfconvert` in x86x64.go.

In `GenOperand`, add `RipRelative` handling:
```go
	case RipRelative:
		label := utfconvert.UtfConvert(ast.Label)
		if a.ostype == util.Darwin {
			label = "_" + label
		}
		return fmt.Sprintf("%s(%%rip)", label)
```

- [ ] **Step 12: Add integration test**

Add to `control_flow_test.go`:
```go
func TestGlobalMutableVars(t *testing.T) {
	output := compileAndRun(t, "test/globals.mn")
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
```

- [ ] **Step 13: Run tests**

Run: `go test -run TestGlobalMutableVars -v -timeout 30s`

Debug with: `go run . gen test/globals.mn --asm && cat out/globals.s`

- [ ] **Step 14: Commit**

```bash
git add tackygen/ code_gen/ test/globals.mn control_flow_test.go
git commit -m "feat: global mutable variables with .data section support"
```

---

### Task 6: Import System

**Files:**
- Modify: `parser/parse.go` (fix parseImport to handle string paths)
- Modify: `parser/ast.go` (add FilePath to ASTImport)
- Modify: `semantic_analysis/sematic_analyze.go` (process imports before resolve)
- Modify: `cli/cli.go` (pass source file path through pipeline)
- Create: `test/import/lib.mn`
- Create: `test/import/main.mn`

- [ ] **Step 1: Write the test files**

Create `test/import/lib.mn`:
```mn
тунх функц нэмэх(а: тоо, б: тоо) -> тоо {
    буц а + б;
}

тунх функц үржүүлэх(а: тоо, б: тоо) -> тоо {
    буц а * б;
}

функц нууц() -> тоо {
    буц 999;
}
```

Create `test/import/main.mn`:
```mn
импорт "lib.mn";

функц үндсэн() -> тоо {
    хэвлэ(нэмэх(10, 20));
    мөр_хэвлэх(" ");
    хэвлэ(үржүүлэх(3, 4));
    мөр_хэвлэх("\n");
    буц 0;
}
```

Expected output: `30 12\n`

- [ ] **Step 2: Fix parseImport to accept string paths**

The current `parseImport` expects an IDENT, but we need it to accept a STRING (e.g., `"lib.mn"`). In `parser/parse.go`, add a `FilePath` field to `ASTImport` in `parser/ast.go`:

In `parser/ast.go`:
```go
type ASTImport struct {
	ASTNode
	Token      lexer.Token
	Ident      string
	SubImports []string
	FilePath   string
}
```

Replace `parseImport()` in `parser/parse.go`:
```go
func (p *Parser) parseImport() *ASTImport {
	ast := &ASTImport{
		Token: p.current,
	}

	// Accept either a string path or an identifier
	if p.peekIs(lexer.STRING) {
		p.nextToken()
		ast.FilePath = *p.current.Value
	} else if p.expect(lexer.IDENT) {
		ast.Ident = *p.current.Value
		for !p.peekIs(lexer.SEMICOLON) {
			if p.peekIs(lexer.DOT) {
				p.nextToken()
				if !p.expect(lexer.IDENT) {
					p.appendError(ErrMissingIdentifier)
					return nil
				}
				ast.SubImports = append(ast.SubImports, *p.current.Value)
			}
		}
	} else {
		p.appendError("файлын зам эсвэл нэр байх ёстой")
		return nil
	}

	if !p.expect(lexer.SEMICOLON) {
		p.appendError(ErrMissingSemicolon)
	}

	return ast
}
```

Also fix `ParseProgram` — the `case lexer.IMPORT` currently calls `p.nextToken()` after `parseImport`, but `parseImport` now consumes the semicolon. Remove the extra `p.nextToken()`:
```go
		case lexer.IMPORT:
			stmt := p.parseImport()
			if stmt != nil {
				program.Decls = append(program.Decls, stmt)
			}
```

Remove the `p.nextToken()` that was on the line after `program.Decls = append(...)`.

- [ ] **Step 3: Add import resolution to semantic analyzer**

In `semantic_analysis/sematic_analyze.go`, add an import processing method. Add fields:

```go
type SemanticAnalyzer struct {
	resolver      *Resolver
	labelPass     *LoopPass
	typeChecker   *TypeChecker
	importedFiles map[string]bool
	baseDir       string
}
```

Update `NewSemanticAnalyzer` to accept a base directory:
```go
func NewSemanticAnalyzer(source []int32, uniqueGen unique.UniqueGen, table *symbols.SymbolTable, baseDir string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		resolver:      NewResolver(source, uniqueGen),
		labelPass:     NewLoopPass(source),
		typeChecker:   NewTypeChecker(source, uniqueGen, table),
		importedFiles: make(map[string]bool),
		baseDir:       baseDir,
	}
}
```

Add import processing before `Resolve`:
```go
func (s *SemanticAnalyzer) processImports(program *parser.ASTProgram) (*parser.ASTProgram, error) {
	var newDecls []parser.ASTDecl
	for _, decl := range program.Decls {
		imp, ok := decl.(*parser.ASTImport)
		if !ok {
			newDecls = append(newDecls, decl)
			continue
		}
		if imp.FilePath == "" {
			continue
		}

		filePath := filepath.Join(s.baseDir, imp.FilePath)
		if s.importedFiles[filePath] {
			continue // already imported, skip
		}
		s.importedFiles[filePath] = true

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("импорт файл уншихад алдаа: %s: %v", imp.FilePath, err)
		}

		runeStr := convertToRuneArray(string(data))
		p := parser.NewParser(runeStr)
		importedProgram, err := p.ParseProgram()
		if err != nil {
			return nil, fmt.Errorf("импорт файл парсингийн алдаа: %s: %v", imp.FilePath, err)
		}

		// Only include public declarations
		for _, d := range importedProgram.Decls {
			switch dt := d.(type) {
			case *parser.FnDecl:
				if dt.IsPublic {
					newDecls = append(newDecls, dt)
				}
			case *parser.VarDecl:
				if dt.IsPublic {
					newDecls = append(newDecls, dt)
				}
			}
		}
	}

	program.Decls = append(newDecls[:0:0], newDecls...)
	// Re-add non-import decls from original that we already added
	return program, nil
}
```

Wait — the above has a bug. Let me restructure. The imported public decls should be prepended, then the current file's non-import decls follow:

```go
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
```

Add the `convertToRuneArray` helper (same as in cli.go):
```go
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
```

Add imports: `"os"`, `"path/filepath"`, `"unicode/utf8"`, `"fmt"`.

In `Analyze`, call `processImports` first:
```go
func (s *SemanticAnalyzer) Analyze(program *parser.ASTProgram) (*parser.ASTProgram, *symbols.SymbolTable, error) {
	s.registerImplicitStdlib()

	program, err := s.processImports(program)
	if err != nil {
		return nil, nil, err
	}

	program, err = s.resolver.Resolve(program)
	// ... rest unchanged
```

- [ ] **Step 4: Update CLI to pass base directory**

In `cli/cli.go`, all functions that create `SemanticAnalyzer` need to pass the source file's directory. For `runGen` (and other commands), extract the directory:

```go
	baseDir := filepath.Dir(args[0])
	resolver := semanticanalysis.NewSemanticAnalyzer(runeString, uniqueGen, table, baseDir)
```

Do this for `runValidate`, `runTacky`, `runCompiler`, and `runGen`.

- [ ] **Step 5: Add integration test**

Add to `control_flow_test.go`:
```go
func TestImport(t *testing.T) {
	output := compileAndRun(t, "test/import/main.mn")
	expected := "30 12\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
```

- [ ] **Step 6: Run tests**

Run: `go test -run TestImport -v -timeout 30s`

- [ ] **Step 7: Commit**

```bash
git add parser/ semantic_analysis/ cli/cli.go test/import/ control_flow_test.go
git commit -m "feat: import system with импорт \"file.mn\" syntax"
```

---

### Task 7: Comprehensive Test Suite

**Files:**
- Modify: `control_flow_test.go` (rename to `integration_test.go`, add all tests)
- Verify all previous tests pass together

- [ ] **Step 1: Rename and consolidate test file**

Rename `control_flow_test.go` to `integration_test.go` (it already has all tests from previous tasks).

Add remaining integration tests:

```go
func TestHelloWorld(t *testing.T) {
	output := compileAndRun(t, "test/hello_world.mn")
	if !strings.Contains(output, "Өдрийн мэнд") {
		t.Errorf("expected hello world output, got %q", output)
	}
}

func TestRule110(t *testing.T) {
	output := compileAndRun(t, "test/rule110.mn")
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}
	// First line should have a single █ in the middle
	if !strings.Contains(lines[0], "█") {
		t.Errorf("first line should contain █")
	}
}

func TestElseIf(t *testing.T) {
	src := `функц үндсэн() -> тоо {
    зарла x: тоо = 2;
    хэрэв x == 1 бол {
        хэвлэ(1);
    } эсвэл хэрэв x == 2 бол {
        хэвлэ(2);
    } эсвэл {
        хэвлэ(3);
    }
    мөр_хэвлэх("\n");
    буц 0;
}`
	tmpFile := t.TempDir() + "/elseif.mn"
	os.WriteFile(tmpFile, []byte(src), 0644)
	output := compileAndRun(t, tmpFile)
	if output != "2\n" {
		t.Errorf("expected \"2\\n\", got %q", output)
	}
}
```

- [ ] **Step 2: Run the full test suite**

Run: `go test -v -timeout 60s`

All tests should pass:
- `TestControlFlow`
- `TestTypeCoercion`
- `TestFree`
- `TestGlobalMutableVars`
- `TestImport`
- `TestHelloWorld`
- `TestRule110`
- `TestElseIf`

- [ ] **Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: comprehensive integration test suite"
```

---
