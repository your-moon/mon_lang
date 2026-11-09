package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestBuiltinsWithoutExternOutsideRepo compiles from a directory that has no
// stdlib/ folder, so the prelude, the C runtime and math.mn must come from the
// copies embedded in the compiler. compileAndRun runs inside the repo, where
// stdlib/ is on disk, so it cannot catch a missing embed.
func TestBuiltinsWithoutExternOutsideRepo(t *testing.T) {
	compiler := filepath.Join(t.TempDir(), "mon_lang")
	if out, err := exec.Command("go", "build", "-o", compiler, ".").CombinedOutput(); err != nil {
		t.Fatalf("build compiler: %v\n%s", err, out)
	}

	work := t.TempDir()
	src := `ашигла "math.mn";

функц үндсэн() -> тоо {
    мөр_хэвлэх("сайн ");
    хэвлэ(их(3, 9));
    мөр_хэвлэх("\n");
    буц 0;
}
`
	if err := os.WriteFile(filepath.Join(work, "prog.mn"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	gen := exec.Command(compiler, "gen", "prog.mn", "-o", "./prog")
	gen.Dir = work
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("compile outside repo: %v\n%s", err, out)
	}

	run := exec.Command("arch", "-x86_64", filepath.Join(work, "prog"))
	var stdout bytes.Buffer
	run.Stdout = &stdout
	if err := run.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got, want := stdout.String(), "сайн 9\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
