package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
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

func TestFree(t *testing.T) {
	output := compileAndRun(t, "test/free.mn")
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestGlobalMutableVars(t *testing.T) {
	output := compileAndRun(t, "test/globals.mn")
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestImport(t *testing.T) {
	output := compileAndRun(t, "test/import/main.mn")
	expected := "30 12\n"
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

func TestHelloWorld(t *testing.T) {
	output := compileAndRun(t, "test/hello_world.mn")
	if !strings.Contains(output, "Өдрийн мэнд") {
		t.Errorf("expected hello world output, got %q", output)
	}
}

func TestRule110(t *testing.T) {
	output := compileAndRun(t, "test/rule110.mn")
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "█") {
		t.Errorf("first line should contain █")
	}
}

func TestElseIf(t *testing.T) {
	src := `функц үндсэн() -> тоо {
    зарла x: тоо = 2;
    хэрэв x == 1 бол {
        хэвлэ(1);
    } эсвэл хэрэв x == 2 бол {
        хэвлэ(2);
    } эсвэл {
        хэвлэ(3);
    }
    мөр_хэвлэх("\n");
    буц 0;
}`
	tmpFile := t.TempDir() + "/elseif.mn"
	os.WriteFile(tmpFile, []byte(src), 0644)
	output := compileAndRun(t, tmpFile)
	if output != "2\n" {
		t.Errorf("expected \"2\\n\", got %q", output)
	}
}
