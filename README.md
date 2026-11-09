<div align="center">

# Mon

**A small, statically-typed programming language with Mongolian keywords,
and a compiler that writes native executables on its own.**

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.23%2B-00ADD8.svg)](go.mod)
[![Platform](https://img.shields.io/badge/platform-macOS%20arm64%20%7C%20x86--64-lightgrey.svg)](#supported-targets)
[![Status](https://img.shields.io/badge/status-educational-orange.svg)](#project-status-and-disclaimer)

[Quick start](#quick-start) ·
[Language tour](#language-tour) ·
[How it works](#how-the-compiler-works) ·
[Contributing](#contributing) ·
[License](#license)

</div>

> [!IMPORTANT]
> **Educational use only.** Mon is a learning project for studying how a
> compiler is built, from lexing to machine code. It is **not** intended for
> production use. It has not been audited for security or correctness. The
> language and compiler may change without notice, and the software is
> provided "as is", without warranty of any kind (see [LICENSE](LICENSE)).
>
> **Зөвхөн боловсролын зорилготой.** Mon нь хөрвүүлэгч (компилятор) хэрхэн
> бүтдэгийг судлах сургалтын төсөл юм. Үйлдвэрлэлийн орчинд ашиглахад
> зориулагдаагүй бөгөөд ямар нэгэн баталгаагүйгээр олгогдоно.

```mon
функц үндсэн() -> тоо {
    мөр_хэвлэх("Өдрийн мэнд\n");
    буц 0;
}
```

## About

Mon lets you write programs with Mongolian (Cyrillic) keywords and identifiers.
Its compiler, written in Go, handles every step itself: it parses the source,
type-checks it and lowers it to an intermediate representation, then encodes
machine code for x86-64 or arm64. It also writes and signs the Mach-O
executable. No `as`, `cc`, `ld`, `codesign` or libc is involved.

The project exists to make the whole compilation pipeline readable in one
repository, and to show that a programming language does not have to be built
on English keywords. A Mon compiler written in Mon itself lives in the
[`mon`](https://github.com/your-moon/mon) submodule.

## Project status and disclaimer

| | |
|---|---|
| **Purpose** | Education and research into compiler construction. |
| **Stability** | Experimental. Syntax, standard library and CLI can change between commits. |
| **Platforms** | macOS on Apple Silicon (arm64) and Intel (x86-64). Other operating systems are not supported. |
| **Memory** | The native heap is a bump allocator; memory is not reclaimed. |
| **Security** | Not audited. Do not compile or run untrusted Mon code with it, and do not use it to build software that handles real data, money or credentials. |
| **Warranty** | None. MIT licensed, provided "as is" ([LICENSE](LICENSE)). |

If you use Mon in a course, workshop or write-up, please link back to this
repository.

## Quick start

### Requirements

- **Go 1.23+**, needed only to build the compiler.
- **macOS**, arm64 or x86-64. Xcode and the Command Line Tools are not
  required.

### Build

```bash
git clone --recurse-submodules https://github.com/your-moon/mon_lang.git
cd mon_lang
go build -o mon .
```

`--recurse-submodules` fetches the self-hosted compiler in `mon/`. You can leave
it out if you only want the Go compiler.

### Compile and run a program

```bash
./mon gen hello.mn --run          # compile for this machine and run it
./mon gen hello.mn -o hello       # write a native executable
./mon gen hello.mn --arch x86_64  # pick a target explicitly (arm64 | x86_64)
```

## Language tour

```mon
тоочих Өнгө { УЛААН, НОГООН, ЦЭНХЭР }

бүтэц Цэг {
    х: тоо64,
    у: тоо64,
}

хэрэгжүүл Цэг {
    функц нийлбэр(өөрөө) -> тоо64 {
        буц өөрөө.х + өөрөө.у;
    }
}

функц өнгөний_нэр(ө Өнгө) -> мөр {
    тааруул ө {
        Өнгө.УЛААН  => { буц "улаан"; }
        Өнгө.НОГООН => { буц "ногоон"; }
        _           => { буц "бусад"; }
    }
    буц "";
}

функц үндсэн() -> тоо {
    зарла ц: Цэг;
    ц.х = 3;
    ц.у = 4;
    хэвлэ(ц.нийлбэр());                     // 7
    мөр_хэвлэх("\n");

    зарла нийт: тоо64 = 0;
    давт и бол 1..10 хүртэл {
        нийт = нийт + и;
    }
    хэвлэ(нийт);                            // 55
    мөр_хэвлэх("\n");

    мөр_хэвлэх(өнгөний_нэр(Өнгө.НОГООН));   // ногоон
    мөр_хэвлэх("\n");
    буц 0;
}
```

### Keywords

| Mon | Meaning | Mon | Meaning |
|-----|---------|-----|---------|
| `функц` | function | `буц` | return |
| `зарла` | declare a variable | `статик` | static |
| `хэрэв` … `бол` | if … then | `эсвэл` | else |
| `давтах` … `бол` | while | `давт` … `хүртэл` | range loop |
| `зогс` | break | `үргэлжлүүл` | continue |
| `тааруул` | match | `тоочих` | enum |
| `бүтэц` | struct | `хэрэгжүүл` | impl (methods) |
| `өөрөө` | self | `тунх` | public (export) |
| `ашигла` | import | `extern` | external declaration |

### Types

| Mon | Meaning |
|-----|---------|
| `тоо`, `тоо64` | signed 32-bit / 64-bit integer |
| `этоо`, `этоо64` | unsigned 32-bit / 64-bit integer |
| `бутархай` | 64-bit float |
| `тэмдэгт` | Unicode code point |
| `мөр` | string |
| `хоосон` | void |
| `тоо*`, `тоо[5]` | pointer, array (heap-backed reference) |

### More features

- **Modules:** `ашигла "файл.mn"`, folder packages (`ашигла "лексер"`), and
  named imports (`ашигла "файл.mn" гэж нэр` → `нэр.функц()`). `тунх` marks a
  declaration as public. Declarations are hoisted, so their order does not
  matter.
- **String escapes:** `\n \t \r \\ \" \' \0 \e \xNN \uNNNN`.
- **Optimizations** are on by default: constant folding, copy propagation and
  dead-code elimination.
### Built-in functions

You never write `extern` declarations. The core built-ins are always
available, and the rest come from standard packages that are built into the
compiler, so they work from any directory:

| Available | Functions |
|-----------|-----------|
| Always | `хэвлэ`, `эхэвлэ`, `мөр_хэвлэх`, `тэмдэгт_хэвлэх`, `бутархай_хэвлэх`, `унш`, `унш32` |
| `ашигла "тэмдэгт_мөр";` | `мөр_урт`, `байт`, `байт_тавих`, `мөр_шинэ` |
| `ашигла "файл";` | `файл_унших_бүтэн`, `файл_бичих`, `файл_бичих_байт` |
| `ашигла "систем";` | `аргумент_тоо`, `аргумент`, `одоо`, `хүлээх`, `санамсаргүй_тоо` (1..n), `дэлгэц_цэвэрлэх` |

```mon
ашигла "систем";

функц үндсэн() -> тоо {
    хэвлэ(санамсаргүй_тоо(6));   // a dice roll, 1..6
    мөр_хэвлэх("\n");
    буц 0;
}
```

Older programs that still declare built-ins with `extern` keep compiling.

Runnable programs that cover every feature are in [`tests/run/`](tests/run).
Each one states its expected output in a `// шалгах:` comment.

## How the compiler works

Every stage is a separate command, so you can inspect what each one produces.
`gen` runs the whole pipeline.

```
source.mn ─▶ lex ─▶ parse ─▶ validate ─▶ tacky ─▶ codegen ─▶ Mach-O executable
```

| Stage | Command | Package |
|-------|---------|---------|
| Tokenize | `lex` | [`lexer/`](lexer) |
| Parse into an AST | `parse` | [`parser/`](parser) |
| Resolve names and check types | `validate` | [`semantic_analysis/`](semantic_analysis), [`symbols/`](symbols), [`mtypes/`](mtypes) |
| Lower to Tacky IR (three-address code) | `tacky` | [`tackygen/`](tackygen) |
| x86-64 code generation | `compile`, `gen --arch x86_64` | [`code_gen/`](code_gen) → [`encoder/`](encoder) → [`macho/`](macho) |
| arm64 code generation | `gen --arch arm64` | [`armgen/`](armgen) → [`macho/`](macho) |

```bash
./mon lex   program.mn --debug   # tokens
./mon parse program.mn --debug   # AST
./mon tacky program.mn --debug   # intermediate representation
```

### Supported targets

| Target | Output | Notes |
|--------|--------|-------|
| **arm64** | Mach-O loaded by dyld, ad-hoc signed by Mon | Runs natively on Apple Silicon. |
| **x86-64** | Static Mach-O (`LC_UNIXTHREAD`) | Runs natively on Intel and under Rosetta 2 on Apple Silicon. |

When `--arch` is omitted, the host architecture is used.

### Self-hosting

- [`mon/`](https://github.com/your-moon/mon) (submodule): the main Mon compiler
  written in Mon (`mon/компилятор/`), with its own `bootstrap.sh`.
- [`selfhost/`](selfhost): a Mach-O writer written in Mon (`макхо.mn`), and
  `self_emit.mn`, which uses it to produce a runnable executable with no C
  toolchain. `TestSelfEmit` covers it.
- [`mc/`](mc): a modular rewrite of the self-hosted compiler using structs,
  methods, modules and enums. Work in progress.

### Learning path

To follow the code in the order the compiler runs:

1. [`lexer/`](lexer): how Cyrillic source becomes tokens.
2. [`parser/`](parser): a hand-written recursive-descent parser.
3. [`semantic_analysis/`](semantic_analysis): scopes, name resolution and type
   checking.
4. [`tackygen/`](tackygen): lowering the AST to a flat IR.
5. [`armgen/`](armgen) or [`code_gen/`](code_gen) + [`encoder/`](encoder):
   instruction selection and machine-code encoding.
6. [`macho/`](macho): building an executable file byte by byte.

## Examples

[`examples/`](examples) includes **гл**, a TinyGL-style software 3D renderer
written in Mon. It has an RGB framebuffer, z-buffered triangle rasterization,
flat shading, truecolor terminal output and PPM export.

```bash
./mon gen examples/гл_үзүүлэн.mn -o үзүүлэн && ./үзүүлэн   # a spinning shaded cube
```

## Tooling

| Tool | Location |
|------|----------|
| Language server (LSP) | [`lsp/`](lsp), [`cmd/mon-lsp/`](cmd/mon-lsp) |
| Tree-sitter grammar and Neovim setup | [`tools/`](tools) |
| VS Code extension | [`vscode/`](vscode) |
| Web playground (Go server + Next.js) | [`playground/`](playground); start with `playground/run.sh` |

## Repository layout

```
lexer/  parser/              Tokenizer, recursive-descent parser → AST
semantic_analysis/           Name resolution and type checking
symbols/  mtypes/            Symbol table and type system
tackygen/                    AST → Tacky IR
code_gen/  encoder/          x86-64: instruction selection and encoding
armgen/                      arm64: Tacky IR → machine code
macho/                       Mach-O writer and ad-hoc code signing
stdlib/                      Built-in functions
interp/                      Tree-walking interpreter (in progress)
selfhost/  mc/  mon/         Mon compilers written in Mon
examples/                    Demo programs, including the гл renderer
tests/                       Run and error test programs
cli/                         Command-line entry point
lsp/  cmd/  tools/  vscode/  Editor tooling
playground/                  Web playground
docs/                        Design specs and plans
```

## Development

```bash
go test ./...                                    # full test suite
MON_TEST_ARM64=1 go test ./tests/ -run ARM64     # native arm64 suite (arm64 host)
```

Tests are Mon programs: [`tests/run/`](tests/run) holds programs and their
expected output, and [`tests/error/`](tests/error) holds programs the compiler
must reject.

## Contributing

Contributions that make Mon a better teaching tool are welcome: bug fixes,
clearer code, new tests, documentation, and Mongolian or English explanations.

1. Open an issue first for anything larger than a small fix, so the design can
   be discussed.
2. Add a test in `tests/run/` or `tests/error/` for every behaviour change.
3. Run `gofmt` and `go test ./...` before opening a pull request.
4. Keep pull requests focused: one change per PR.

Bug reports should include the Mon source, the command you ran, what you
expected, and what happened.

## Security

Mon is a teaching compiler and has not had a security review. Generated
executables have no sandboxing, no memory reclamation and minimal runtime
checks. Please report suspected issues through a GitHub issue. There is no
private disclosure process, because the project is not meant for production
use.

## Acknowledgements

- Nora Sandler, [*Writing a C Compiler*](https://nostarch.com/writing-c-compiler)
  (No Starch Press, 2024). The pipeline structure and the Tacky IR follow this
  book.
- Everyone who has contributed code, tests and feedback.

## License

Mon is released under the [MIT License](LICENSE).
© 2024 e.munkherdene and Mon Compiler Contributors.

This software is provided for **educational purposes only**, "as is", without
warranty of any kind. The authors are not liable for any damage arising from
its use.
