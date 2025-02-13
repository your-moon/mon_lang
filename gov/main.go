package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/base"
	"github.com/your-moon/mn_compiler_go_version/compiler"
	"github.com/your-moon/mn_compiler_go_version/gen"
	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
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
		fmt.Println("NODE:", node.PrintAST())
	}
	if len(parsed.Errors()) > 0 {
		panic(parsed.Errors()[0])
	}

	fmt.Println("---- COMPILING ----:")
	compilerx := compiler.NewCompiler()
	err = compilerx.Compile(node)
	if err != nil {
		panic(err)
	}

	outfile := "out.asm"
	openFile, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	// target := gen.QBE
	emitter := gen.NewX86Emitter(openFile, compilerx.Irs)
	emitter.Emit()

	openFile.Close()

	// if target == gen.QBE {
	// 	// cmd := exec.Command("qbe", "-t", "arm64_apple", "-o", "out.s", "out.asm")
	// 	cmd := exec.Command("qbe", "-o", "out.s", "out.asm")
	// 	if err := cmd.Run(); err != nil {
	// 		fmt.Println("Error: ", err)
	// 	}
	// 	cmd = exec.Command("as", "-o", "out.o", "out.s")
	// 	if err := cmd.Run(); err != nil {
	// 		fmt.Println("Error: ", err)
	// 	}
	// 	// xcrunCmd := exec.Command("xcrun", "--show-sdk-path")
	// 	// syslibrootPath, err := xcrunCmd.Output()
	// 	// if err != nil {
	// 	// 	fmt.Println("Error getting syslibroot path:", err)
	// 	// 	os.Exit(1)
	// 	// }
	// 	// syslibroot := strings.TrimSpace(string(syslibrootPath))
	//
	// 	// Prepare the `ld` command
	// 	// cmd = exec.Command("ld", "-o", "out", "out.o", "-syslibroot", syslibroot, "-lSystem")
	//
	// 	// cmd = exec.Command("./out")
	//
	// 	// Set the output to stdout/stderr
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Stderr = os.Stderr
	//
	// 	// Run the command
	// 	if err := cmd.Run(); err != nil {
	// 		fmt.Println("Error executing ld command:", err)
	// 		os.Exit(1)
	// 	}
	// }

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
