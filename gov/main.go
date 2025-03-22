package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/base"
	codegen "github.com/your-moon/mn_compiler_go_version/code_gen"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
	"github.com/your-moon/mn_compiler_go_version/tackygen"
)

func main() {
	base.Debug = true

	filePath := "./test/ch1/return.mn"
	runeString := convertToRuneArray(func() string {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return ""
		}
		return string(data)
	}())

	if base.Debug {
		for _, item := range runeString {
			fmt.Printf("val: %s, code: %d\n", string(item), item)
		}
	}

	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseProgram()
	if err != nil {
		panic(err)
	}

	if base.Debug {
		fmt.Println("NODE:", node.PrintAST(0))
	}
	if len(parsed.Errors()) > 0 {
		panic(parsed.Errors())
	}

	fmt.Println("---- COMPILING ----:")
	compilerx := tackygen.NewTackyGen()
	tackyprogram := compilerx.EmitTacky(node)

	fmt.Println("---- TACKY LIST ----:")
	for _, ir := range tackyprogram.FnDef.Instructions {
		fmt.Println(ir)
		ir.Ir()
	}

	fmt.Println("---- ASMAST ----:")
	codegen := codegen.NewAsmGen()
	asmgen := codegen.GenAsm(tackyprogram)
	fmt.Println(asmgen)

	outfile := "out.asm"
	openFile, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}

	openFile.Close()

}

type FileCheck struct {
	File   string
	Status bool
	Error  error
}

func processFile(filePath string) FileCheck {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return FileCheck{
			File:   filePath,
			Status: false,
		}
	}

	runeString := convertToRuneArray(string(data))
	scanner := lexer.NewScanner(runeString)

	for {
		token, err := scanner.Scan()
		if err != nil {
			return FileCheck{
				File:   filePath,
				Status: false,
				Error:  err,
			}
		}
		if token.Type == lexer.EOF {
			return FileCheck{
				File:   filePath,
				Status: true,
			}
		}
	}
}

func convertToRuneArray(dataString string) []int32 {
	var runeString []int32
	for len(dataString) > 0 {
		r, size := utf8.DecodeRuneInString(dataString)
		runeString = append(runeString, r)
		dataString = dataString[size:]
	}
	runeString = append(runeString, 0)
	return runeString
}

func canLex() []FileCheck {
	dir := "./test/ch1/"
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return nil
	}

	var fileChecks []FileCheck
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		fileChecks = append(fileChecks, processFile(filePath))
	}

	return fileChecks
}
