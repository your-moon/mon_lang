<div align="center">

# Mon

**A small, statically-typed programming language with Mongolian keywords.**

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.23%2B-00ADD8.svg)](go.mod)
[![Target](https://img.shields.io/badge/target-x86--64-lightgrey.svg)](#requirements)
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
> **Зөвхөн сургалтын зорилгоор.** Mon нь хөрвүүлэгч (компилятор) лексик
> шинжилгээнээс машин код үүсгэх хүртэл хэрхэн ажилладгийг судлах сургалтын
> төсөл юм. Бодит бүтээгдэхүүн, үйлчилгээнд (production) ашиглахад
> зориулагдаагүй. Аюулгүй байдал, зөв ажиллагааны хувьд хөндлөнгийн шалгалт
> хийгдээгүй бөгөөд хэл болон хөрвүүлэгч урьдчилан мэдэгдэлгүйгээр өөрчлөгдөж
> болно. Програм хангамж нь ямар нэгэн баталгаагүй бөгөөд түүнийг ашигласнаас
> үүдэх аливаа хохирлыг зохиогч хариуцахгүй. Дэлгэрэнгүйг [LICENSE](LICENSE)
> файлаас үзнэ үү.

```mon
функц үндсэн() -> тоо {
    мөр_хэвлэх("Сайн байна уу, Дэлхий!\n");
    буц 0;
}
```

## About

Mon lets you write programs with Mongolian (Cyrillic) keywords and identifiers.
Its compiler, written in Go, lexes and parses the source, type-checks it,
lowers it to an intermediate representation called Tacky, and generates x86-64
assembly. The system assembler and C compiler then turn that assembly into a
native executable.

The project exists to make the whole compilation pipeline readable in one
repository, and to show that a programming language does not have to be built
on English keywords.

## Project status and disclaimer

| | |
|---|---|
| **Purpose** | Education and research into compiler construction. |
| **Stability** | Experimental. Syntax, standard library and CLI can change between commits. |
| **Platforms** | x86-64 output. On Apple Silicon it runs under Rosetta 2. |
| **Security** | Not audited. Do not compile or run untrusted Mon code with it, and do not use it to build software that handles real data, money or credentials. |
| **Warranty** | None. MIT licensed, provided "as is" ([LICENSE](LICENSE)). |

If you use Mon in a course, workshop or write-up, please link back to this
repository.

## Quick start

### Requirements

- **Go 1.23+** to build the compiler.
- **A C toolchain** (`as` and `cc`). On macOS, install the Xcode Command Line
  Tools with `xcode-select --install`.
- **Rosetta 2** on Apple Silicon, because the output is x86-64
  (`softwareupdate --install-rosetta`).

### Build

```bash
git clone https://github.com/your-moon/mon_lang.git
cd mon_lang
go build -o mon_lang .
```

The standard library is built into the compiler binary, so you can copy
`mon_lang` anywhere and compile from any directory.

### Compile and run a program

```bash
./mon_lang gen hello.mn --run      # compile and run
./mon_lang gen hello.mn -o hello   # write an executable named "hello"
```

| Flag | Description |
|------|-------------|
| `--run` | Run the program after compiling |
| `-o` | Output file name |
| `--debug` | Print the output of each stage |
| `--asm` | Keep the generated `.s` assembly file |
| `--obj` | Keep the generated object file |

## Language tour

```mon
функц фибоначчи(н тоо64) -> тоо64 {
    хэрэв н <= 1 бол {
        буц н;
    }
    зарла өмнөх: тоо64 = 0;
    зарла одоогийн: тоо64 = 1;
    зарла i: тоо64 = 2;
    давтах i <= н бол {
        зарла дараах: тоо64 = өмнөх + одоогийн;
        өмнөх = одоогийн;
        одоогийн = дараах;
        i = i + 1;
    }
    буц одоогийн;
}

функц үндсэн() -> тоо {
    зарла к: тоо64 = 0;
    давтах к < 10 бол {
        хэвлэ(фибоначчи(к));
        мөр_хэвлэх(" ");
        к = к + 1;
    }
    мөр_хэвлэх("\n");   // 0 1 1 2 3 5 8 13 21 34
    буц 0;
}
```

### Keywords

| Mon | Meaning | Mon | Meaning |
|-----|---------|-----|---------|
| `функц` | function | `буц` | return |
| `зарла` | declare a variable | `давтах` … `бол` | while |
| `хэрэв` … `бол` | if … then | `эсвэл` | else |
| `зогс` | break | `үргэлжлүүл` | continue |
| `ашигла` | import | `тунх` | public (export) |

### Types

| Mon | Meaning |
|-----|---------|
| `тоо` | 32-bit integer |
| `тоо64` | 64-bit integer |
| `мөр` | string |
| `хоосон` | void |

### Built-in functions

You don't need `extern` declarations. These built-ins are available in every
program:

| Function | Does |
|----------|------|
| `хэвлэ(н)` | print an integer |
| `мөр_хэвлэх(м)` | print a string |
| `унш()`, `унш32()` | read an integer from standard input |
| `санамсаргүйТоо(н)` | random number |
| `одоо()` | seconds since the Unix epoch |
| `хүлээх(мс)` | sleep for `мс` milliseconds |
| `дэлгэцЦэвэрлэх()` | clear the terminal |
| `чөлөөлөх(п)` | free memory |

### Modules

Mark a function `тунх` to export it, and load a file with `ашигла`. Imports are
looked up next to your program first, then in the standard library, which
ships with a math module:

```mon
ашигла "math.mn";

функц үндсэн() -> тоо {
    хэвлэ(их(10, 20));   // 20
    мөр_хэвлэх("\n");
    буц 0;
}
```

More programs are in [`playground/examples/`](playground/examples) and
[`test/`](test).

## How the compiler works

Every stage is a separate command, so you can inspect what each one produces.
`gen` runs the whole pipeline.

```
source.mn ─▶ lex ─▶ parse ─▶ validate ─▶ tacky ─▶ compile ─▶ as + cc ─▶ executable
```

| Stage | Command | Package |
|-------|---------|---------|
| Tokenize | `lex` | [`lexer/`](lexer) |
| Parse into an AST | `parse` | [`parser/`](parser) |
| Resolve names and check types | `validate` | [`semantic_analysis/`](semantic_analysis), [`symbols/`](symbols), [`mtypes/`](mtypes) |
| Lower to Tacky IR (three-address code) | `tacky` | [`tackygen/`](tackygen) |
| Generate x86-64 assembly | `compile` | [`code_gen/`](code_gen) |
| Assemble and link | `gen` | [`linker/`](linker) |

```bash
./mon_lang lex   program.mn --debug   # tokens
./mon_lang parse program.mn --debug   # AST
./mon_lang tacky program.mn --debug   # intermediate representation
```

### Learning path

To follow the code in the order the compiler runs:

1. [`lexer/`](lexer): how Cyrillic source becomes tokens.
2. [`parser/`](parser): a hand-written recursive-descent parser.
3. [`semantic_analysis/`](semantic_analysis): scopes, name resolution and type
   checking.
4. [`tackygen/`](tackygen): lowering the AST to a flat IR.
5. [`code_gen/`](code_gen): instruction selection and x86-64 assembly.
6. [`linker/`](linker) and [`stdlib/lib.c`](stdlib/lib.c): how the runtime is
   linked in.

## Repository layout

```
lexer/  parser/        Tokenizer, recursive-descent parser → AST
semantic_analysis/     Name resolution and type checking
symbols/  mtypes/      Symbol table and type system
tackygen/              AST → Tacky IR
code_gen/              Tacky IR → x86-64 assembly
linker/                Assemble and link with as / cc
stdlib/                Prelude, math module and C runtime (embedded in the compiler)
cli/                   Command-line entry point
playground/            Web playground (Go server + Next.js)
vscode/                VS Code syntax extension
test/                  Example and feature programs
```

## Development

```bash
go test ./...
```

The web playground starts with `playground/run.sh` and serves
http://localhost:8080.

## Contributing

Contributions that make Mon a better teaching tool are welcome: bug fixes,
clearer code, new tests, documentation, and Mongolian or English explanations.

1. Open an issue first for anything larger than a small fix, so the design can
   be discussed.
2. Add a test for every behaviour change.
3. Run `gofmt` and `go test ./...` before opening a pull request.
4. Keep pull requests focused: one change per PR.

Bug reports should include the Mon source, the command you ran, what you
expected, and what happened.

## Security

Mon is a teaching compiler and has not had a security review. Generated
executables have minimal runtime checks. Please report suspected issues through
a GitHub issue. There is no private disclosure process, because the project is
not meant for production use.

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
