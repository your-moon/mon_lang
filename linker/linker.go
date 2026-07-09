/*
 * mon_lang - linker
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package linker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	OUTPUT_DIR = "out"
	STDLIB_DIR = "stdlib"
)

// Linker writes and runs native executables. The machine code and Mach-O
// container are produced entirely in-process by LinkNative (see native.go);
// there is no external as/cc/ld toolchain.
type Linker struct {
	outputFile string
}

func NewLinker(outputFile string) *Linker {
	if filepath.IsAbs(outputFile) || strings.HasPrefix(outputFile, ".") || strings.HasPrefix(outputFile, "..") {
		return &Linker{outputFile: outputFile}
	}
	return &Linker{outputFile: filepath.Join(OUTPUT_DIR, outputFile)}
}

func (l *Linker) MakeExecutable() error {
	return os.Chmod(l.outputFile, 0755)
}

// Run executes the built program, forwarding stdio. The program's exit code
// comes back as *ExitCodeError so main can propagate it instead of treating
// a nonzero exit as a compiler failure.
func (l *Linker) Run() error {
	return RunPath(l.outputFile)
}

// RunPath executes a built program by path, forwarding stdio and propagating
// the exit code as *ExitCodeError.
func RunPath(path string) error {
	if !strings.ContainsRune(path, '/') {
		path = "./" + path
	}
	cmd := exec.Command(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &ExitCodeError{Code: exitErr.ExitCode()}
	}
	return err
}

// ExitCodeError carries the compiled program's own exit status.
type ExitCodeError struct{ Code int }

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("програм %d кодоор дууслаа", e.Code)
}
