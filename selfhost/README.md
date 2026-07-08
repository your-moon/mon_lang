# Self-hosted mon_lang (боотстрап)

The mon_lang compiler, written in mon_lang. Bootstrap stages:

- **Stage 0 — lexer0.mn** (done): tokenizer over raw bytes with UTF-8
  Cyrillic identifiers, strings, comments. Lexes its own source.
- **Stage 1 — parser**: token stream → AST (structs + methods).
- **Stage 2 — checker + TACKY**: semantic analysis and IR lowering.
- **Stage 3 — codegen**: TACKY → x86_64 bytes + Mach-O writer,
  ported from the Go encoder/macho packages.
- **Stage 4 — bootstrap**: monc (Go) compiles monc.mn; monc.mn
  compiles itself; outputs must be identical.

Run stage 0:

    ./mon_lang gen selfhost/lexer0.mn -o lexer0
    ./lexer0 selfhost/lexer0.mn


## Bootstrap status (mc.mn)

`mc.mn` is a working mon_lang compiler written in mon_lang. It compiles a
self-hostable subset (functions, тоо64, locals/globals, arrays, arithmetic,
comparisons, if/else-if/while/break, string/char literals, calls, builtins,
prototypes) to x86_64 assembly.

- **Compiles real programs correctly** — recursion, loops, strings,
  arithmetic, conditionals — verified end to end (`mc_test.sh`).
- **Compiles its own source** — `mc` (built by the Go-hosted compiler)
  compiles `mc.mn` into ~7500 lines of assembly, which assembles and links
  into a stage-2 `mc` that runs.
- **Self-hosting fixed point reached**: `mc2` (mc compiled by mc) compiles
  `mc.mn` to assembly byte-identical to `mc`'s (7553 lines) — the compiler
  reproduces itself exactly. Verify with `./bootstrap.sh`.

Drive it with `./build.sh <program.mn> <output>` (entry function: үндсэн).
