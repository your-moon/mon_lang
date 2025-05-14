package linker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Linker struct {
	outputFile string
	asmContent string
	osType     string
	genAsm     bool
	genObj     bool
}

func NewLinker(outputFile string) *Linker {
	return &Linker{
		outputFile: outputFile,
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
	// Get just the filename without path
	outputName := filepath.Base(l.outputFile)

	if l.genAsm {
		return os.WriteFile(outputName+".s", []byte(l.asmContent), 0644)
	}

	tempAsmFile := outputName + ".asm"
	if err := os.WriteFile(tempAsmFile, []byte(l.asmContent), 0600); err != nil {
		return fmt.Errorf("failed to write temporary assembly file: %v", err)
	}
	defer os.Remove(tempAsmFile)

	objFile := outputName + ".o"
	asmCmd := exec.Command("as", "-o", objFile, tempAsmFile)
	if err := asmCmd.Run(); err != nil {
		return fmt.Errorf("failed to assemble: %v", err)
	}

	if l.genObj {
		return nil // Object file is already created with the correct name
	}

	defer os.Remove(objFile)

	var linkCmd *exec.Cmd
	if l.osType == "darwin" {
		linkCmd = exec.Command("ld", "-arch", "x86_64", "-o", outputName, objFile, "-e", "_start", "-no_pie", "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
	} else {
		linkCmd = exec.Command("ld", "-o", outputName, objFile)
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
