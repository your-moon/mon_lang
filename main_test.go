package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/your-moon/mn_compiler/cli"
)

func TestLex(t *testing.T) {
	dir := "./test/"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Error reading directory: %v", err)
	}

	compiler := cli.New()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		err := compiler.Run([]string{"compiler", "lex", filePath})
		assert.NoError(t, err, "Failed to lex file: %s", filePath)
	}
}
