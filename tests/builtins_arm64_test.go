package tests

import (
	"path/filepath"
	"testing"
)

// TestARM64CompilesEveryStdPackage guards against a builtin that is declared in
// a standard package but has no arm64 implementation. TestRunARM64 skips files
// that fail to compile, so such a gap would otherwise stay invisible.
// Compiling needs no arm64 host; only running does.
func TestARM64CompilesEveryStdPackage(t *testing.T) {
	for _, f := range []string{"систем_цаг.mn", "file_io.mn", "argv.mn"} {
		t.Run(f, func(t *testing.T) {
			bin := filepath.Join(t.TempDir(), "prog")
			if err := compileARM64(filepath.Join("run", f), bin); err != nil {
				t.Fatalf("arm64 compile failed: %v", err)
			}
		})
	}
}
