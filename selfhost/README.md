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
