package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/your-moon/mon_lang/cli"
)

func TestLex(t *testing.T) {
	dir := "./test/"
	compiler := cli.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".mn") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			err := compiler.Run([]string{"compiler", "lex", path})
			assert.NoError(t, err, "Failed to lex file: %s", path)
		})
		return nil
	})
	if err != nil {
		t.Fatalf("Error walking directory: %v", err)
	}
}
