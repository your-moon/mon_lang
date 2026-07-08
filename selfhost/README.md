# Self-hosted mon_lang (боотстрап)

`mc` is the mon_lang compiler **written in mon_lang**. It compiles a
self-hostable subset — functions and prototypes, `тоо64`, locals and globals,
arrays, arithmetic, comparisons, `хэрэв`/`эсвэл`/`давтах`/`зогс`, calls, string
and char literals, and the runtime builtins — to x86_64 AT&T assembly.

## Modules

The compiler is split with the language's own `ашигла` imports:

| File         | Role                                                        |
| ------------ | ----------------------------------------------------------- |
| `лексер.mn`  | tokeniser (lexer state + scanning)                          |
| `кодген.mn`  | symbol tables + x86_64 emitter + expression/statement code  |
| `mc.mn`      | driver: resolves local module imports, then compiles        |

`mc.mn` inlines the source of each local `.mn` import before lexing — the same
way the Go-hosted front end merges modules — so the split compiler still
reproduces itself.

## Self-hosting is verified

```
./bootstrap.sh
```

runs the full chain:

- **stage 1** — `mc` (built by the Go-hosted compiler) compiles `mc.mn`
- **stage 2** — assemble that into `mc2` (mc compiled by mc)
- **stage 3** — `mc2` compiles `mc.mn`

Self-hosting holds when **stage 1 == stage 3**, byte-identical — a reproducing
fixed point (the compiler compiles itself into exactly itself).

## Scripts

- `bootstrap.sh` — build fresh and verify the fixed point.
- `mc_test.sh` — smoke test: `mc` compiles a program and it runs correctly.
- `build.sh <program.mn> <output>` — compile any program with `mc`
  (entry function must be named `үндсэн`).
