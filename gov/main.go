package main

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/your-moon/mn_compiler_go_version/lexer"
)

func main() {
	// fmt.Println("оруулах стд")
	data, err := os.ReadFile("./examples/test.mn")
	if err != nil {
		panic("can't read file")
	}
	fmt.Println("BYTEARR:", data)

	dataString := string(data)

	var runeString []int32
	for len(dataString) > 0 {
		r, size := utf8.DecodeRuneInString(dataString)
		// fmt.Printf("%c %v\n", r, size)
		// if r == 'ф' {
		// 	fmt.Println("ITS F")
		// }
		runeString = append(runeString, r)
		dataString = dataString[size:]
	}

	scanner := lexer.NewScanner(runeString)

	fmt.Println("LEN:", len(runeString))
	for range len(runeString) - 1 {
		token, _ := scanner.Scan()
		fmt.Println("TOKEN:", token)
	}
	// fmt.Println(data)
	// fmt.Println(dataString)
}
