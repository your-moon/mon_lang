# Design: interpreter engine, native test framework, DAP debugging, arch auto-detect

**Date:** 2026-07-09
**Status:** approved (approach), pending spec review

## Goal

Give mon_lang three developer-facing capabilities that all lean on one new
keystone:

1. A **native test framework** — write tests in mon, run them with `mon шалга`,
   `go test` / `cargo test`-style output.
2. **Debugging via DAP** — set breakpoints, step, and inspect variables from
   VS Code, with zero external tooling (no LLDB, no DWARF).
3. **Hardware auto-detect** — the compiler picks the arm64 or x86_64 backend
   from the host automatically, on both the Go driver and the self-hosted `mc`.

The keystone is a **tree-walking interpreter** — the same choice Python
(debugpy), Java (JDWP), and JS (V8 Inspector) make: put debug hooks *inside your
own engine* instead of depending on an external native debugger. It keeps mon's
zero-external-tooling ethos, and it makes the test runner and DAP both fall out
of one build.

## Non-goals (this pass)

- Rewriting the interpreter or the DAP wire protocol **in mon**. Those are
  Go-hosted host tools first (like CPython's debug core is C). A self-hosted
  interpreter is a documented future epic, not part of this pass.
- Native (ptrace/Mach-exception) debugging of the compiled Mach-O binary.
- DWARF emission.

## Motivating defect found during design

The native backend has **no memory reclamation**: `monAlloc` is a bump
allocator (`armgen/stdlib.go`; equivalent on the x86 native path) and
`чөлөөлөх` is a no-op. A mark-sweep GC previously existed only on the legacy
`cc` path, which has since been removed entirely (native code is the only
target now), so the native runtime has *no* reclamation at all. **Any native
binary that allocates in a loop leaks.**
This surfaced as the `гл_үзүүлэн` 3D demo exhausting RAM (~24 MB/s): a per-frame
display buffer plus a local `тоо64[24]` scratch array — which lowers to
`monAlloc` (`tackygen/tack_gen.go:268`) — ran ~9000×/frame. The demo is fixed
(allocation-free), but the underlying gap makes native binaries unusable for
long-running programs. **Stage 5 (native heap reclamation) closes it.** The
interpreter (Stage 1) is Go-hosted and inherits Go's GC, so it never leaks —
programs run/tested through the interpreter are already safe.

## What is self-hosted vs host-tool

| Piece | Self-hosted (written in mon) | Host tool (Go) |
|---|---|---|
| Assertion library (`шалгалт` pkg) | ✅ | — |
| Test files (`*_шалгалт.mn`) | ✅ | — |
| Arch detection + arm64 emitter in `mc` | ✅ (stage 4 full port) | — |
| Test runner (`mon шалга`) | later | ✅ first |
| Interpreter eval loop | future epic | ✅ |
| DAP server | future epic | ✅ |

Everything mon-native mirrors into the `mon` submodule once stable
(mon_lang-now, mirror-later).

---

## Stage 1 — Interpreter core (`interp/` new Go package)

A tree-walker over the **resolved** AST (post-semantic-analysis: names resolved,
enums constant-folded, types attached). Reuses lexer → parser → sema unchanged;
only evaluation is new.

### Value model

Tagged union:

- `Int32`, `Int64` — integers (mon `тоо` is signed 32-bit; `тоо64` is 64-bit).
- `Double` — `бутархай`.
- `Ptr` — reference to a heap object: struct, array, or string buffer.

The heap is Go-hosted (Go maps/slices); Go's GC replaces the C mark-sweep
collector. Interpreted programs cannot forge pointers into native memory, so the
interpreter is memory-safe by construction. Pointer/`деref`/`addr-of` semantics
model mon's pointer ops over heap object handles + offsets, matching what the
native backend does with real addresses.

### Frames & scopes

A call stack of frames `{fn, locals (lexical scopes), currentNode}`. Block
scopes push/pop. `үндсэн` is the entry; argc/argv are threaded in; its return
value is the process exit code.

### Builtins bridge

Every `extern` function gets a Go host implementation matching the native
syscall stdlib (`encoder/stdlib.go`) semantics exactly (so interpreter and
native agree byte-for-byte):

- print family: `хэвлэ`, `эхэвлэ`, `мөр_хэвлэх`, `тэмдэгт_хэвлэх` (UTF-8),
  `бутархай_хэвлэх`.
- strings/bytes: `мөр_урт`, `байт`, `байт_тавих`, `мөр_шинэ`.
- files: `файл_унших_бүтэн`, `файл_бичих`, `файл_бичих_байт`.
- memory: `monAlloc` (returns a fresh zeroed heap buffer handle),
  `чөлөөлөх` (no-op).
- misc: `унш`, `унш32`, `санамсаргүй_тоо`, `одоо`, `хүлээх`,
  `дэлгэц_цэвэрлэх`, `аргумент_тоо`, `аргумент`.

`бутархай_хэвлэх` reproduces lib.c's 6-digit truncated formatting exactly so the
differential oracle stays green.

### The debug seam

One hook, defined as a Go interface:

```go
type Hook interface {
    OnStmt(node parser.ASTStmt, fr *Frame) // called before each statement
}
```

`nil` in normal runs (zero cost). This single seam is everything DAP needs —
breakpoint checks, step granularity, and the paused frame snapshot all key off
it. `node.Token.Line` gives the source line for breakpoint matching (verified:
tokens carry `Line` + `Span`; AST nodes carry a `Token`).

### Validation

The interpreter is wired into the existing Go test harness as a **second mode**
alongside the native back end: every `tests/run/*.mn` file runs interpreted and
its `(stdout, exit)` must match both the directive expectations and the native
result. Interpreter-vs-native is now the differential oracle (the legacy cc path
that formerly played that role has been removed). This proves parity before
anything is built on top of the interpreter.

---

## Stage 2 — Test framework (`mon шалга`)

Convention-driven, Go/Rust-style.

### Conventions

- **Files:** `*_шалгалт.mn` ("шалгалт" = test/inspection).
- **Test functions:** named `шалга_*`, signature `() -> тоо` (return 0 = pass by
  default; assertion failures flip it). Mirrors Go `TestX` / Rust `#[test]`.

### Assertion library — `шалгалт` stdlib package (written in mon)

- `батал(нөхцөл тоо)` — assert truthy.
- `тэнцүү(а тоо, б тоо)` — assert equal (int).
- `тэнцүү_мөр(а мөр, б мөр)` — assert equal (string).
- `бус(нөхцөл тоо)` — assert falsy.

Dual implementation:

- **Interpreter builtins:** on failure, throw a Go-level `TestFailure`
  (recovered per-test by the runner) carrying message + source line.
- **Native/C impls:** print a diagnostic to stderr and set a failure flag /
  exit non-zero (for `--төрөлх` native mode).

### Runner: `mon шалга [зам]`

1. Discover `*_шалгалт.mn` under the path (default `.`).
2. For each, compile front-end once, then for each `шалга_*` function: run it in
   the interpreter, isolated (one failure/panic doesn't abort the run).
3. Report `cargo test`-style: per-test `ok` / `FAILED`, failure messages with
   file:line, a summary line with pass/fail counts and elapsed time.
4. Exit non-zero if any test failed.

### `--төрөлх` (native) flag

Additionally compile+run each test through the native backend and assert the
native result agrees with the interpreter — the differential oracle, now a
first-class test mode.

---

## Stage 3 — DAP server (`dap/` new Go package) + VS Code adapter

### Wire

DAP over stdio: `Content-Length`-framed JSON messages. Launched via `mon dap`.

### Requests implemented

`initialize`, `launch`, `setBreakpoints`, `configurationDone`, `threads`,
`stackTrace`, `scopes`, `variables`, `continue`, `next`, `stepIn`, `stepOut`,
`evaluate`, `disconnect`.

### Mechanics

- The interpreter runs on a goroutine with a DAP `Hook` installed.
- `OnStmt` checks the statement's line against active breakpoints and, when a
  breakpoint is hit or a step is pending, blocks on a channel and emits a DAP
  `stopped` event.
- DAP requests (`continue`, `next`, `stepIn`, `stepOut`) set the step mode and
  unblock the channel.
- `stackTrace` / `scopes` / `variables` read the paused frame stack directly
  from the interpreter (single-threaded: one thread id).
- `evaluate` re-uses the interpreter to evaluate an expression in the paused
  frame's scope.

Step semantics: `next` = stop at next statement at ≤ current frame depth;
`stepIn` = next statement any depth; `stepOut` = next statement at < current
depth.

### VS Code adapter

Extend the existing `vscode/` extension with a `debuggers` contribution point
(type `mon`), a default `launch.json` configuration, and a breakpoint language
registration for `.mn`. `mon dap` is the adapter executable. Result:
set-breakpoint / F5 / step / hover-variable works in VS Code end to end.

---

## Stage 4 — Arch auto-detect

### Go driver (`cli/cli.go`) — immediate

When `--arch` is empty, default to the host instead of hardcoded x86_64:
`runtime.GOARCH == "arm64"` → arm64 backend; `"amd64"` → x86_64. Explicit
`--arch` still overrides. One small change; ship first.

### Self-hosted `mc` — full arm64 port

`mc` today emits **x86_64 only** (`компилятор/макхо.mn` is x86, `кодген.mn` is
x86 AT&T). To make self-hosted `mc` auto-emit arm64 on Apple Silicon:

1. **Host detection:** a builtin/extern `гэрийн_архитектур() -> мөр` (via
   `uname`/`sysctl`) returning `"arm64"` or `"x86_64"`.
2. **arm64 emitter in mon:** port `armgen/*.go` (AArch64 encoder + stack-machine
   codegen) into `компилятор/арм_кодген.mn`.
3. **arm64 Mach-O + signing in mon:** port `macho/arm64.go` (dyld-loaded signed
   Mach-O + ad-hoc CodeDirectory) into `компилятор/арм_макхо.mn`. The existing
   x86 `макхо.mn` is the template; the arm64 recipe is documented in the
   arm64-Mach-O reference.
4. **Dispatch:** `mc` selects backend by detected host arch (overridable by a
   flag), mirroring the Go driver.

This is the largest single chunk and the natural completion of self-hosting: mon
compiling itself to native code on both architectures with no external tooling.
Built last, on top of a proven interpreter/test/DAP stack.

---

## Build order & incremental verification

1. **Interpreter** → verify against the whole `tests/run/` suite as a
   differential oracle. Nothing proceeds until parity is green.
2. **Test framework** → assertion stdlib + `mon шалга` runner; add mon-native
   `*_шалгалт.mn` tests; verify.
3. **DAP** → server + VS Code adapter; manual breakpoint/step/inspect smoke.
4. **Arch auto-detect** → Go driver first (trivial), then the self-hosted arm64
   port (largest chunk), verified by emitting+running a native arm64 binary from
   `mc` on Apple Silicon.

Each stage is independently shippable and independently tested. mon-native
artifacts (`шалгалт` pkg, test files, `мс` arm64 emitter) mirror into the `mon`
submodule once stable.
```
