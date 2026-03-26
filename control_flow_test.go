package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func compileAndRun(t *testing.T, srcFile string) string {
	t.Helper()
	outFile := t.TempDir() + "/out"

	// Compile
	cmd := exec.Command("go", "run", ".", "gen", srcFile, "-o", outFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("compile failed: %v\nstderr: %s", err, stderr.String())
	}

	// Run (x86_64 on ARM mac)
	runCmd := exec.Command("arch", "-x86_64", outFile)
	var stdout bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	if err := runCmd.Run(); err != nil {
		t.Fatalf("run failed: %v\nstderr: %s", err, stderr.String())
	}

	return stdout.String()
}

func TestControlFlow(t *testing.T) {
	output := compileAndRun(t, "test/control_flow.mn")
	expected := "3 12 6 20 99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestTypeCoercion(t *testing.T) {
	output := compileAndRun(t, "test/types.mn")
	expected := "300 30 42 100\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
