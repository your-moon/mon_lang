# Mon_lang Compiler Improvements Design

## Context

The mon_lang compiler successfully compiles Rule 110 with array support, libc-based stdlib, and implicit stdlib registration. This spec covers the next set of improvements: memory management, control flow, type safety, error reporting, global variables, imports, and testing.

## 1. Manual Memory Free — `чөлөөлөх(arr)`

**Problem:** `шинэ тоо[N]` calls `malloc` but nothing ever frees. Every array leaks.

**Solution:** Add `чөлөөлөх` as an implicit stdlib function.

**Changes:**
- `semantic_analysis/sematic_analyze.go`: Add `чөлөөлөх` to `registerImplicitStdlib()` with return type `VoidType` and param type `Int64Type` (pointer).
- `tackygen/tack_gen.go`: Add `"чөлөөлөх"` to the implicit externs list in `EmitTacky`.
- `stdlib/lib.c`: Add `void chqlqqlqkh(void *p) { free(p); }`. The function name is the UTF-converted form of `чөлөөлөх`.

**No new IR instructions.** It's a regular function call: `чөлөөлөх(arr);` compiles to `FnCall{Name: "чөлөөлөх", Args: [ptr]}`.

**Verification:** Write a test that allocates an array, uses it, frees it, and exits cleanly.

## 2. Else-If Chains — `эсвэл хэрэв ... бол`

**Status: Already works.** Verified that `эсвэл хэрэв x == 2 бол { ... } эсвэл { ... }` compiles and runs correctly. The `parseStmt()` call after `эсвэл` naturally recurses into `parseIf()`. No changes needed — just add a test to the suite to prevent regression.

## 3. Fix 32/64-bit Type Coercion

**Problem:** `getCommonType()` returns `Int64Type` for any type mismatch, but no `SignExtend` instruction is emitted. The codegen uses 32-bit instructions on values that should be 64-bit, causing corrupt upper bits.

**Solution:** Insert `SignExtend` instructions in `tackygen/tack_gen.go` whenever a 32-bit value is used in a 64-bit context.

**Changes to `tack_gen.go`:**

1. **Binary operations** in `EmitExpr` for `*parser.ASTBinary`: After emitting both operands, check if one is Int32 and the other is Int64 (or the common type is Int64). If so, sign-extend the Int32 operand before the binary instruction.

2. **Function call arguments** in `EmitExpr` for `*parser.ASTFnCall`: For each arg, if the function's param type is Int64 but the arg expression is Int32, emit SignExtend.

3. **Variable declarations** in `EmitVarDecl`: If `decl.VarType` is Int64 but the init expression is Int32, emit SignExtend before the Copy.

4. **Assignments** in `EmitExpr` for `*parser.ASTAssignment`: If the target variable is Int64 and the RHS is Int32, emit SignExtend.

**Helper function:**
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

**Verification:** Test program that mixes `тоо` and `тоо64` arithmetic and prints results.

## 4. Multi-Error Reporting

**Problem:** Parser collects multiple errors in `parseErrors` but `ParseProgram()` returns only the first. Same pattern in semantic analysis.

**Solution:**

**Parser (`parser/parse.go`):**
- `ParseProgram()` line 63-67: Instead of returning `p.parseErrors[0]`, join all errors and return a combined error.
- Add panic recovery in `parseDecl` so a panic in one declaration doesn't stop parsing of subsequent declarations.

**Semantic analysis:**
- `resolve.go`: `Resolve()` should collect errors from each declaration and continue processing remaining declarations. Return all errors at the end.
- `type_checker.go`: Same pattern — collect errors from each top-level declaration, continue, return all.
- `sematic_analyze.go`: `Analyze()` should run all three passes even if earlier passes have non-fatal errors. Collect and return all.

**Error display (`cli/cli.go`):**
- When errors are returned, print each one on its own line with the source context.
- Cap at 10 errors to avoid flooding (like GCC/Clang).

**Verification:** Write a test file with 3 deliberate errors, verify all 3 are reported.

## 5. Control Flow Codegen Audit

**Problem:** The if-else had a missing `Jump` instruction after the "then" branch. Similar bugs may exist elsewhere.

**Audit targets:**
- `ASTWhile` in `EmitTackyStmt`: Verify the while loop has correct jump back to start and break label after body.
- `ASTLoop` (for-range): Verify increment and jump back are correct. Check the `else` branch of the range check (non-range loops).
- `ASTBreakStmt` / `ASTContinueStmt`: Verify they jump to the correct labels.
- Nested loops: Verify break/continue in inner loop don't affect outer loop.
- `ASTConditional` (ternary `? :`): Verify the jump-over-else is present.

**Method:** Write a comprehensive control flow test program and verify output matches expected.

**Test program should cover:**
- While loop with break
- While loop with continue
- Nested while loops with break from inner
- For-range loop
- If / else-if / else chain
- Ternary in assignment
- Return from inside a loop

**Verification:** Single test file that prints a sequence of numbers. Expected output is deterministic.

## 6. Global Mutable Variables

**Problem:** Top-level `зарла` only works with constant integer initializers (they're inlined as constants). Mutable globals and non-constant initializers don't work.

**Solution:** Support global variables in the `.data` section with RIP-relative addressing.

**Changes:**

### Type system
- `tackygen/tacky.go`: Add `GlobalVar` struct: `{ Name string; InitValue TackyVal }`.
- `tackygen/tack_gen.go`: In `EmitTacky`, for top-level `VarDecl`:
  - If init is a constant: store in `GlobalConstants` map (existing behavior for inlining) AND emit a `GlobalVar`.
  - If init is non-constant: emit init code in a list that gets prepended to `үндсэн`'s body.

### Code generation
- `code_gen/emitter.go`: Add handling for global variable references. When `AsmType` encounters a global var (check a globals set), use RIP-relative addressing instead of stack-relative.
- `code_gen/asm_ast.go`: Add `RipRelative` operand type: `{ Label string; Offset int }`. Emits as `label(%rip)`.
- `code_gen/x86x64.go`: In `GenAsm`, emit `.data` section entries for global variables before `.text`.
- `code_gen/pseudo.go`: Global variables don't get stack slots. Skip them in pseudo replacement and use their RIP-relative label directly.

### Tacky gen
- `EmitExpr` for `ASTVar`: If the variable is a global (not a constant), emit a Load from the global's address instead of a stack Var.
- Global variable writes: emit a Store to the global's address.

**Verification:** Test with a mutable global counter incremented across function calls.

## 7. Import System — `импорт "file.mn"`

**Problem:** All code must be in a single file.

**Solution:** `импорт "file.mn";` reads, parses, and merges the imported file's public declarations.

**Changes:**

### Parser (`parser/parse.go`)
- `parseImport()` already partially exists. Fix it to extract the file path string.
- Return `ASTImport` with the file path.

### Semantic analyzer (`semantic_analysis/sematic_analyze.go`)
- Before calling `Resolve()`, process imports:
  1. For each `ASTImport` in `program.Decls`, read and parse the imported file.
  2. Track already-imported files in a `map[string]bool` to prevent cycles.
  3. Filter: only include declarations marked with `тунх` (public) or `IsPublic: true`.
  4. Prepend the imported declarations to `program.Decls` (before the current file's declarations).
  5. Remove the `ASTImport` nodes from the decl list.

### File resolution
- Import paths are relative to the importing file's directory.
- `импорт "math.mn"` looks for `math.mn` in the same directory as the source file.
- No search paths or package directories for now.

### CLI (`cli/cli.go`)
- Pass the source file's directory path through the pipeline so the import resolver knows where to look.

**Visibility rule:** Only `тунх`-prefixed declarations from imported files are visible. Non-public declarations are parsed but excluded from the importing file's scope.

**Example:**
```
// math.mn
тунх функц нэмэх(а: тоо, б: тоо) -> тоо {
    буц а + б;
}

// main.mn
импорт "math.mn";
функц үндсэн() -> тоо {
    хэвлэ(нэмэх(1, 2));
    буц 0;
}
```

**Verification:** Two-file program where main imports a utility file and calls its functions.

## 8. Go Test Suite

**Problem:** No automated tests. All testing is manual.

**Solution:** Add Go test files that test each compiler stage.

**Test files:**

| File | Tests |
|------|-------|
| `lexer/lexer_test.go` | Token sequences for array syntax, keywords, brackets |
| `parser/parser_test.go` | Already exists — extend with array, if-else, global var AST checks |
| `semantic_analysis/semantic_test.go` | Type checking, scope resolution, error detection |
| `tackygen/tacky_test.go` | IR output for array ops, Load/Store, SignExtend |
| `integration_test.go` (root) | Compile + run `.mn` files, check stdout against expected output |

**Integration test pattern:**
```go
func TestRule110(t *testing.T) {
    output := compileAndRun("test/rule110.mn")
    lines := strings.Split(output, "\n")
    assert(len(lines) == 50)
    assert(strings.Contains(lines[0], "█"))
}
```

**Test `.mn` files to add:**
- `test/elseif.mn` — else-if chain test
- `test/types.mn` — mixed 32/64-bit arithmetic
- `test/globals.mn` — mutable global variables
- `test/import/main.mn` + `test/import/lib.mn` — import test
- `test/control_flow.mn` — comprehensive control flow
- `test/free.mn` — malloc + free

**Verification:** `go test ./...` passes.

## Implementation Order

1. **Control flow audit** (Step 5) — find and fix bugs before building on top
2. **Multi-error reporting** (Step 4) — better DX for all subsequent work
3. **Type coercion** (Step 3) — correctness foundation
4. **`чөлөөлөх`** (Step 1) — simplest feature addition
5. **Global mutable variables** (Step 6) — new codegen capability
6. **Import system** (Step 7) — largest change, depends on everything else working
7. **Test suite** (Step 8) — written incrementally alongside each step, comprehensive pass at end

Note: Else-if chains (Step 2) already work. No implementation needed — just add regression test.
