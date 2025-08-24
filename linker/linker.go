package linker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	OUTPUT_DIR = "out"
	STDLIB_DIR = "stdlib"
)

type Linker struct {
	outputFile string
	asmContent string
	osType     string
	genAsm     bool
	genObj     bool
}

func NewLinker(outputFile string) *Linker {
	if filepath.IsAbs(outputFile) || strings.HasPrefix(outputFile, ".") || strings.HasPrefix(outputFile, "..") {
		return &Linker{
			outputFile: outputFile,
			osType:     runtime.GOOS,
		}
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
		return fmt.Errorf("гаралтын хавтас үүсгэхэд алдаа гарлаа: %v", err)
	}

	outputName := filepath.Base(l.outputFile)

	if l.genAsm {
		return os.WriteFile(l.outputFile+".s", []byte(l.asmContent), 0644)
	}

	tempAsmFile := filepath.Join(outputDir, outputName+".s")
	if err := os.WriteFile(tempAsmFile, []byte(l.asmContent), 0600); err != nil {
		return fmt.Errorf("гаралтын түр ассемблер файл үүсгэхэд алдаа гарлаа: %v", err)
	}
	defer os.Remove(tempAsmFile)

	objFile := filepath.Join(outputDir, outputName+".o")
	var asmCmd *exec.Cmd
	if l.osType == "darwin" && runtime.GOARCH == "arm64" {
		asmCmd = exec.Command("arch", "-x86_64", "as", "-o", objFile, tempAsmFile)
	} else {
		asmCmd = exec.Command("as", "-o", objFile, tempAsmFile)
	}
	var asmStdout, asmStderr bytes.Buffer
	asmCmd.Stdout = &asmStdout
	asmCmd.Stderr = &asmStderr
	if err := asmCmd.Run(); err != nil {
		return fmt.Errorf("ассембле хийхэд алдаа гарлаа: %v\nstdout: %s\nstderr: %s", err, asmStdout.String(), asmStderr.String())
	}

	if l.genObj {
		return nil
	}

	defer os.Remove(objFile)

	stdlibFile := filepath.Join(STDLIB_DIR, "lib_"+l.osType+".asm")
	stdlibObj := filepath.Join(outputDir, "lib_"+l.osType+".o")
	var stdlibAsmCmd *exec.Cmd
	if l.osType == "darwin" && runtime.GOARCH == "arm64" {
		stdlibAsmCmd = exec.Command("arch", "-x86_64", "as", "-o", stdlibObj, stdlibFile)
	} else {
		stdlibAsmCmd = exec.Command("as", "-o", stdlibObj, stdlibFile)
	}
	var stdlibStdout, stdlibStderr bytes.Buffer
	stdlibAsmCmd.Stdout = &stdlibStdout
	stdlibAsmCmd.Stderr = &stdlibStderr
	if err := stdlibAsmCmd.Run(); err != nil {
		return fmt.Errorf("стандарт сан ассембле хийхэд алдаа гарлаа: %v\nstdout: %s\nstderr: %s", err, stdlibStdout.String(), stdlibStderr.String())
	}
	defer os.Remove(stdlibObj)

	var linkCmd *exec.Cmd
	if l.osType == "darwin" && runtime.GOARCH == "arm64" {
		linkCmd = exec.Command("arch", "-x86_64", "ld", "-arch", "x86_64", "-o", l.outputFile, objFile, stdlibObj, "-e", "_main", "-no_pie", "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
	} else if l.osType == "darwin" {
		linkCmd = exec.Command("ld", "-arch", "x86_64", "-o", l.outputFile, objFile, stdlibObj, "-e", "_main", "-no_pie", "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk")
	} else {
		linkCmd = exec.Command("ld", "-o", l.outputFile, objFile, stdlibObj)
	}

	var stdout, stderr bytes.Buffer
	linkCmd.Stdout = &stdout
	linkCmd.Stderr = &stderr

	if err := linkCmd.Run(); err != nil {
		return fmt.Errorf("линк хийхэд алдаа гарлаа: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}

func (l *Linker) MakeExecutable() error {
	return os.Chmod(l.outputFile, 0755)
}

func (l *Linker) Run() error {
	cmd := exec.Command(l.outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
