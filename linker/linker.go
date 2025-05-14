package linker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	OUTPUT_DIR = "out"
)

type Linker struct {
	outputFile string
	asmContent string
	osType     string
	genAsm     bool
	genObj     bool
}

func NewLinker(outputFile string) *Linker {
	outputFile = filepath.Clean(outputFile)

	if filepath.IsAbs(outputFile) {
		outputFile = filepath.Base(outputFile)
	}

	return &Linker{
		outputFile: filepath.Join(OUTPUT_DIR, outputFile),
		osType:     runtime.GOOS,
	}
}

func (l *Linker) SetAssemblyContent(content string) {
	l.asmContent = content
}

func (l *Linker) SetGenerateAsm(genAsm bool) {
	l.genAsm = genAsm
}

func (l *Linker) SetGenerateObj(genObj bool) {
	l.genObj = genObj
}

func (l *Linker) Link() error {
	outputDir := filepath.Dir(l.outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	outputName := filepath.Base(l.outputFile)

	if l.genAsm {
		return os.WriteFile(l.outputFile+".s", []byte(l.asmContent), 0644)
	}

	tempAsmFile := filepath.Join(outputDir, outputName+".asm")
	if err := os.WriteFile(tempAsmFile, []byte(l.asmContent), 0600); err != nil {
		return fmt.Errorf("failed to write temporary assembly file: %v", err)
	}
	defer os.Remove(tempAsmFile)

	objFile := filepath.Join(outputDir, outputName+".o")
	asmCmd := exec.Command("as", "-o", objFile, tempAsmFile)
	if err := asmCmd.Run(); err != nil {
		return fmt.Errorf("failed to assemble: %v", err)
	}

	if l.genObj {
		return nil
	}

	defer os.Remove(objFile)

	var linkCmd *exec.Cmd
	if l.osType == "darwin" {
		linkCmd = exec.Command("ld", "-arch", "x86_64", "-o", l.outputFile, objFile, "-e", "_start", "-no_pie", "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
	} else {
		linkCmd = exec.Command("ld", "-o", l.outputFile, objFile)
	}

	var stdout, stderr bytes.Buffer
	linkCmd.Stdout = &stdout
	linkCmd.Stderr = &stderr

	if err := linkCmd.Run(); err != nil {
		return fmt.Errorf("failed to link: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}

func (l *Linker) MakeExecutable() error {
	return os.Chmod(l.outputFile, 0755)
}
