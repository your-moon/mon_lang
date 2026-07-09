/*
 * mon_lang - self-emission end-to-end test
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSelfEmit proves mon_lang can emit a native executable with zero external
// tooling: it compiles selfhost/self_emit.mn (which imports the mon_lang Mach-O
// writer макхо.mn), runs it to emit /tmp/mon_self_emit_out, then execs that
// emitted binary and checks it exits 42. macxо.mn emits an x86_64 Mach-O; on
// arm64 hosts it runs under Rosetta, so this is gated to amd64/Rosetta hosts.
func TestSelfEmit(t *testing.T) {
	driver := filepath.Join("..", "selfhost", "self_emit.mn")
	if _, err := os.Stat(driver); err != nil {
		t.Skipf("driver missing: %v", err)
	}
	out := filepath.Join(t.TempDir(), "emitter")
	if err := compile(driver, out, false); err != nil {
		t.Fatalf("compile emitter: %v", err)
	}
	emitted := "/tmp/mon_self_emit_out"
	os.Remove(emitted)
	if _, code := run(t, out, ""); code != 0 {
		t.Fatalf("emitter exited %d", code)
	}
	if _, err := os.Stat(emitted); err != nil {
		t.Fatalf("emitter did not produce %s: %v", emitted, err)
	}
	os.Chmod(emitted, 0755)

	// exec the mon_lang-emitted binary and check its exit code.
	cmd := exec.Command(emitted)
	err := cmd.Run()
	ee, ok := err.(*exec.ExitError)
	if !ok {
		if err != nil && runtime.GOARCH == "arm64" {
			t.Skipf("emitted x86_64 binary not runnable here: %v", err)
		}
		t.Fatalf("expected exit 42, got err=%v", err)
	}
	if ee.ExitCode() != 42 {
		t.Fatalf("emitted binary exit=%d, want 42", ee.ExitCode())
	}
}
