package macho

import (
	"encoding/binary"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestARM64ExitCode emits `movz w0,#42; ret` (dyld calls entry as main and
// exits with its return value) and checks the native process exits 42.
func TestARM64ExitCode(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		t.Skip("native arm64 only")
	}
	code := make([]byte, 0, 8)
	code = binary.LittleEndian.AppendUint32(code, 0x52800000|(42<<5)) // movz w0,#42
	code = binary.LittleEndian.AppendUint32(code, 0xD65F03C0)          // ret

	path := t.TempDir() + "/exit42"
	if err := WriteExecutableARM64(path, code, nil, nil, 0); err != nil {
		t.Fatal(err)
	}
	err := exec.Command(path).Run()
	if ee, ok := err.(*exec.ExitError); ok {
		if ee.ExitCode() != 42 {
			t.Fatalf("exit=%d want 42", ee.ExitCode())
		}
		return
	}
	t.Fatalf("expected exit 42, got err=%v", err)
}

var _ = os.Stdout
