# Mon

Mon is a small statically-typed programming language with Mongolian keywords. This repo contains its compiler, written in Go, which lowers Mon source through a Tacky intermediate representation to x86-64 assembly, then assembles and links it into a native executable.

```mon
extern функц мөр_хэвлэх(м мөр) -> хоосон {}

функц үндсэн() -> тоо {
    мөр_хэвлэх("Өдрийн мэнд");
    буц 0;
}
```

## How it works

The compiler runs as a pipeline, and each stage is exposed as its own command:

| Stage | Command | Package |
|-------|---------|---------|
| Tokenize source | `lex` | `lexer/` |
| Build the AST | `parse` | `parser/` |
| Type check & resolve | `validate` | `semantic_analysis/`, `symbols/`, `mtypes/` |
| Lower to Tacky IR | `tacky` | `tackygen/` |
| Generate x86-64 assembly | `compile` | `code_gen/` |
| Assemble + link to a binary | `gen` | `linker/` |

`gen` runs the whole pipeline end to end. Assembly and linking shell out to the system `as` and `cc`; the standard library lives in `stdlib/`.

## Requirements

- Go 1.23.4+
- A C toolchain (`as` and `cc` — on Apple Silicon the linker invokes `arch -x86_64`, so Rosetta is required)

## Build

```bash
git clone https://github.com/your-moon/mon_lang.git
cd mon_lang
go build -o mon_lang .
```

## Usage

```bash
# Compile and run a program
./mon_lang gen program.mn --run

# Compile to an executable named "out"
./mon_lang gen program.mn -o out

# Inspect an intermediate stage
./mon_lang lex program.mn --debug
./mon_lang parse program.mn --debug
./mon_lang tacky program.mn --debug
```

| Flag | Description |
|------|-------------|
| `--debug` | Print the output of the stage |
| `--asm` | Keep the generated `.s` assembly file |
| `--obj` | Keep the generated object file |
| `--run` | Run the program after compiling |
| `-o` | Output file name |

## Language examples

### Arithmetic

```mon
extern функц хэвлэ(н тоо64) -> хоосон {}

функц үндсэн() -> тоо {
    зарла a: тоо64 = 10;
    зарла b: тоо64 = 5;

    хэвлэ(a + b);
    хэвлэ(a - b);
    хэвлэ(a * b);
    хэвлэ(a / b);

    буц 0;
}
```

### Fibonacci

```mon
extern функц хэвлэ(н тоо64) -> хоосон {}
extern функц унш() -> тоо64 {}

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
    хэвлэ(фибоначчи(унш()));
    буц 0;
}
```

## Repository layout

```
lexer/              Tokenizer
parser/             Recursive-descent parser → AST
semantic_analysis/  Type checking and validation
symbols/  mtypes/   Symbol table and type system
tackygen/           AST → Tacky IR lowering
code_gen/           Tacky IR → x86-64 assembly
linker/             Assemble and link via as/cc
stdlib/             Runtime / built-in functions
cli/                Command-line entry point
playground/         Web playground (Next.js frontend + Go server)
vscode/             VS Code syntax extension
```

## Development

```bash
go test ./...
go build ./...
```

## License

MIT — see [LICENSE](LICENSE).
