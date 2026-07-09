# Mon

Mon is a small statically-typed programming language with **Mongolian keywords**.
It compiles straight to native machine code for **x86-64 and arm64 (Apple
Silicon)** and writes the executable itself — its own instruction encoder, its
own Mach-O writer, its own ad-hoc code signing. No `as`, no `cc`, no `ld`, no
`codesign`, no libc. And it **self-hosts**: Mon is written in Mon.

```mon
функц үндсэн() -> тоо {
    мөр_хэвлэх("Өдрийн мэнд\n");
    буц 0;
}
```

```bash
go build -o mon .
./mon gen сайн.mn --arch arm64 -o сайн   # native Apple Silicon binary
./сайн
```

## Language

- **Types:** `тоо`/`тоо64` (int32/64), `этоо`/`этоо64` (unsigned), `тэмдэгт`
  (Unicode codepoint), `бутархай` (float64), `мөр` (string), pointers (`тоо*`),
  arrays (`тоо[5]`, heap-backed references).
- **Structs + methods** (`бүтэц` / `хэрэгжүүл` / `өөрөө`):

  ```mon
  бүтэц Цэг { х: тоо, у: тоо64 }
  хэрэгжүүл Цэг {
      функц нийлбэр(өөрөө) -> тоо64 { буц өөрөө.х + өөрөө.у; }
  }
  ```

- **Enums** (`тоочих`) — variants are integer constants, usable as `match` patterns:

  ```mon
  тоочих Өнгө { УЛААН, НОГООН, ЦЭНХЭР }
  тааруул ө {
      Өнгө.УЛААН => { мөр_хэвлэх("улаан"); }
      _         => { мөр_хэвлэх("бусад"); }
  }
  ```

- **Rust-style `match`** (`тааруул`) on integers, chars, and enum variants.
- **Control flow:** `хэрэв`/`эсвэл`, `давтах` (while), `давт..хүртэл` (range
  for), `зогс`/`үргэлжлүүл` (break/continue).
- **Modules:** `ашигла "файл.mn"`, folder packages (`ашигла "лексер"`), named
  imports `ашигла "файл.mn" гэж нэр` → `нэр.функц()`, `тунх` for public exports.
  Declarations are order-independent (Go-style hoisting).
- **String escapes:** `\n \t \\ \" \0 \r \e \xNN \uNNNN`.
- Optimizations on by default (constant folding, copy propagation, DCE).
- Bump-allocated heap on the native path (conservative mark-sweep GC on the
  legacy `--cc` runtime).
- **Builtins:** printing (`хэвлэ`, `эхэвлэ`, `мөр_хэвлэх`, `тэмдэгт_хэвлэх`,
  `бутархай_хэвлэх`), strings/bytes (`мөр_урт`, `байт`, `байт_тавих`, `мөр_шинэ`),
  file I/O (`файл_унших_бүтэн`, `файл_бичих`, `файл_бичих_байт`), stdin (`унш`),
  argv (`аргумент`), and time/random/sleep.

## Targets

| Target | How | Notes |
|--------|-----|-------|
| **arm64** | `gen … --arch arm64` | Native on Apple Silicon. dyld-loaded, self ad-hoc-signed (required by modern macOS). |
| **x86-64** | `gen …` (default) | Static `LC_UNIXTHREAD` binary; runs natively on Intel, under Rosetta 2 on Apple Silicon. |

Both back ends are fully self-contained: Mon encodes the machine code
(`encoder/`, `armgen/`), writes the Mach-O container (`macho/`), and links a
built-in syscall standard library. Nothing external is invoked.

## Pipeline

Each stage is its own command; `gen` runs them all end to end.

| Stage | Command | Package |
|-------|---------|---------|
| Tokenize | `lex` | `lexer/` |
| Parse → AST | `parse` | `parser/` |
| Type check & resolve | `validate` | `semantic_analysis/`, `symbols/`, `mtypes/` |
| Lower to Tacky IR | `tacky` | `tackygen/` |
| x86-64 codegen | `compile`/`gen` | `code_gen/` → `encoder/` → `macho/` |
| arm64 codegen | `gen --arch arm64` | `armgen/` → `macho/` |

Pass `--asm` to emit AT&T text, or `--cc` for the legacy `as`+`cc` path.

## Self-hosting

Mon compiles itself. `selfhost/mc.mn` (plus `лексер.mn`, `кодген.mn`) is a Mon
compiler **written in Mon**; it bootstraps to a byte-identical fixed point:

```bash
./selfhost/bootstrap.sh
# SELF-HOSTING OK: stage1 == stage3 (…, byte-identical)
```

`selfhost/макхо.mn` is a Mach-O writer written in Mon — the language can emit a
native executable entirely on its own (see `tests` → `TestSelfEmit`). `mc/` is a
cleaner, modular rewrite (structs/methods/modules/enums) in progress.

## Requirements

- Go 1.23+ (build-time only).
- macOS (arm64 or x86-64). The `--cc` legacy path additionally needs Xcode
  Command Line Tools.

## Build & run

```bash
git clone git@github.com:your-moon/mon.git
cd mon
go build -o mon .

./mon gen program.mn --arch arm64 -o out && ./out   # native arm64
./mon gen program.mn -o out && ./out                # x86-64
./mon gen program.mn --run                           # compile + run

# inspect a stage
./mon tacky program.mn --debug
```

## Examples

`examples/` includes **`гл.mn`** — a TinyGL-style software 3D renderer written in
Mon: RGB framebuffer, z-buffered triangle rasterization, per-face colour +
shading, 24-bit truecolor terminal output and PPM image save.

```bash
./mon gen examples/гл_үзүүлэн.mn --arch arm64 -o үзүүлэн && ./үзүүлэн
# a colour-shaded spinning cube, in your terminal
```

## Repository layout

```
lexer/  parser/            Tokenizer, recursive-descent parser → AST
semantic_analysis/         Resolve, type-check (structs, enums, modules)
symbols/  mtypes/           Symbol table and type system
tackygen/                  AST → Tacky IR
code_gen/ encoder/ macho/  x86-64: asm AST → machine code → Mach-O
armgen/                    arm64: Tacky IR → machine code (+ signed Mach-O)
stdlib/                    Built-in functions (native + C shim)
selfhost/  mc/             Mon compiler written in Mon (self-hosting)
examples/                  Demos incl. the гл 3D renderer
cli/                       Command-line entry point
lsp/  tools/               Language server, tree-sitter grammar, Neovim setup
playground/  vscode/       Web playground, VS Code extension
```

## Development

```bash
go test ./...                              # full suite
MON_TEST_ARM64=1 go test ./tests/ -run ARM64   # native arm64 suite (arm64 host)
./selfhost/bootstrap.sh                    # self-hosting check
```

## License

MIT — see [LICENSE](LICENSE).
