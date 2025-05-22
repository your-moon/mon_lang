# 🚀 Mon Compiler

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Contributions Welcome](https://img.shields.io/badge/Contributions-Welcome-brightgreen.svg?style=flat)](CONTRIBUTING.md)

A modern compiler implementation written in Go for the Mon programming language.

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Examples](#code-examples) • [Contributing](#contributing)

</div>

## 📖 Overview

Mon Compiler is an open-source compiler that translates Mon source code into executable programs. Built with modern compiler design principles, it provides a complete compilation pipeline from source code to machine code.

### 🌟 Key Features

- **Lexical Analysis** - Efficient tokenization of source code
- **Parser Implementation** - Robust syntax analysis
- **Semantic Analysis** - Type checking and validation
- **Code Generation** - Optimized machine code generation
- **Standard Library** - Rich set of built-in functions
- **Error Handling** - Comprehensive error reporting
- **Symbol Table** - Efficient symbol management
- **Type System** - Strong static typing

## 🏗️ Project Structure

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

## ⚙️ Requirements

- Go 1.23.4 or higher
- Make (optional, for build scripts)

## 🚀 Installation

```bash
# Clone the repository
git clone https://github.com/your-moon/mn_compiler.git

# Navigate to the project directory
cd mn_compiler

# Build the compiler
go build
```

## 💻 Usage

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
```

### 🔧 Command Options

| Option | Description |
|--------|-------------|
| `--debug` | Enable debug mode |
| `--asm` | Generate assembly file |
| `--obj` | Generate object file |
| `--run` | Compile and run the program |
| `-o` | Specify output file name |

### 📝 Examples

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

## 📚 Code Examples

Here are some examples of Mon code to help you get started:

### 🌍 Hello World
```mon
extern функц мөр_хэвлэх(м мөр) -> хоосон {}

функц үндсэн() -> тоо {
    мөр_хэвлэх("Өдрийн мэнд");
    буц 0;
}
```

### 🔢 Basic Arithmetic
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

### 🔄 Fibonacci Example
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

## 🛠️ Development

```bash
# Run tests
go test ./...

# Build the project
go build
```

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### 📋 Contributing Guidelines

- Write clear, descriptive commit messages
- Follow the existing code style
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Thanks to all contributors who have helped shape this project
- Inspired by modern compiler design principles and practices

---

<div align="center">
Made with ❤️ by e.munkherdene
</div> 