# Mon Compiler

A modern compiler implementation written in Go for the Mon programming language.

## Overview

Mon Compiler is an open-source compiler that translates Mon source code into executable programs. It features a complete compilation pipeline including lexical analysis, parsing, semantic analysis, and code generation.

## Features

- Lexical Analysis
- Parser Implementation
- Semantic Analysis
- Code Generation
- Standard Library Support
- Error Handling
- Symbol Table Management
- Type System

## Project Structure

```
.
├── base/           # Base utilities and common functionality
├── cli/            # Command-line interface implementation
├── codegen/        # Code generation components
├── errors/         # Error handling and reporting
├── lexer/          # Lexical analysis implementation
├── linker/         # Linker implementation
├── mconstant/      # Constant definitions
├── mn/             # Core language components
├── mtypes/         # Type system implementation
├── out/            # Output directory
├── parser/         # Parser implementation
├── rustv/          # Rust version compatibility
├── semantic_analysis/ # Semantic analysis implementation
├── stdlib/         # Standard library implementation
├── stringpool/     # String interning implementation
├── symbols/        # Symbol table management
├── tackygen/       # Target code generation
└── util/           # Utility functions
```

## Requirements

- Go 1.23.4 or higher
- Make (optional, for build scripts)

## Installation

```bash
# Clone the repository
git clone https://github.com/your-moon/mn_compiler.git

# Navigate to the project directory
cd mn_compiler

# Build the compiler
go build
```

## Usage

The compiler provides several commands for different stages of compilation:

```bash
# Lexical Analysis
compiler lex input.mn [--debug]

# Parsing
compiler parse input.mn [--debug]

# Semantic Analysis
compiler validate input.mn [--debug]

# Generate Tacky IR
compiler tacky input.mn [--debug]

# Compile to Assembly
compiler compile input.mn [--debug]

# Full Compilation Pipeline
compiler gen input.mn [options]

# Options for gen command:
  --debug    Enable debug mode
  --asm      Generate assembly file
  --obj      Generate object file
  --run      Compile and run the program
  -o         Specify output file name
```

### Examples

```bash
# Basic compilation
compiler gen input.mn

# Compile with debug output
compiler gen input.mn --debug

# Generate assembly file
compiler gen input.mn --asm

# Generate object file
compiler gen input.mn --obj

# Compile and run
compiler gen input.mn --run

# Specify output file
compiler gen input.mn -o myprogram
```

## Code Examples

Here are some examples of Mon code to help you get started:

### Hello World
```mon
extern функц мөр_хэвлэх(м мөр) -> хоосон {}

функц үндсэн() -> тоо {
    мөр_хэвлэх("Өдрийн мэнд");
    буц 0;
}
```

### Basic Arithmetic
```mon
extern функц хэвлэ(н тоо64) -> хоосон {}

функц үндсэн() -> тоо {
    зарла a: тоо64 = 10;
    зарла b: тоо64 = 5;
    
    зарла нийлбэр: тоо64 = a + b;
    зарла ялгавар: тоо64 = a - b;
    зарла үржвэр: тоо64 = a * b;
    зарла хуваарь: тоо64 = a / b;
    
    хэвлэ(нийлбэр);
    хэвлэ(ялгавар);
    хэвлэ(үржвэр);
    хэвлэ(хуваарь);
    
    буц 0;
}
```

### Function Definition
```mon
extern функц хэвлэ(н тоо64) -> хоосон {}

функц нэмэх(a тоо64, b тоо64) -> тоо64 {
    буц a + b;
}

функц үндсэн() -> тоо {
    зарла хариу: тоо64 = нэмэх(5, 3);
    хэвлэ(хариу);
    буц 0;
}
```

### Conditional Statements
```mon
extern функц хэвлэ(н тоо64) -> хоосон {}

функц үндсэн() -> тоо {
    зарла нас: тоо64 = 18;
    
    хэрэв нас >= 18 бол {
        хэвлэ(1); // Том хүн
    } эсвэл {
        хэвлэ(0); // Насанд хүрээгүй
    }
    
    буц 0;
}
```

### Loops
```mon
extern функц хэвлэ(н тоо64) -> хоосон {}

функц үндсэн() -> тоо {
    зарла тоолуур: тоо64 = 0;
    
    давтах тоолуур < 5 бол {
        хэвлэ(тоолуур);
        тоолуур = тоолуур + 1;
    }
    
    буц 0;
}
```

### Fibonacci Example
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
    зарла n: тоо64 = унш();
    зарла хариу: тоо64 = фибоначчи(n);
    хэвлэ(хариу);
    буц 0;
}
```

## Development

```bash
# Run tests
go test ./...

# Build the project
go build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped shape this project
- Inspired by modern compiler design principles and practices 