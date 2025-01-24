package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/lexer"
	"github.com/your-moon/mn_compiler_go_version/parser"
)

func main() {
	filePath := "./examples/binary.mn"
	runeString := convertToRuneArray(func() string {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return ""
		}
		return string(data)
	}())

	parsed := parser.NewParser(runeString)
	node, err := parsed.ParseExpr()
	if err != nil {
		panic(err)
	}
	fmt.Println("NODE:", node)

	// scanner := lexer.NewScanner(runeString)
	// for {
	// 	token, err := scanner.Scan()
	// 	if err != nil {
	// 		fmt.Printf("Scanning error: %v\n", err)
	// 		break
	// 	}
	// 	if token.Type == lexer.EOF {
	// 		fmt.Println("Reached EOF")
	// 		break
	// 	}
	// 	fmt.Println("TOKEN:", token)
	// }
}

type FileCheck struct {
	File   string
	Status bool
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
	files, err := os.ReadDir("./examples")
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return nil
	}

	var fileChecks []FileCheck
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join("./examples", file.Name())
		fileChecks = append(fileChecks, processFile(filePath))
	}

	return fileChecks
}
