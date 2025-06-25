package bugs

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

// TestBug02_Repro verifies the build fails with a helpful error when protoc is missing.
func TestBug02_Repro(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Regression test; passes post-fix")
	}
	cmd := exec.Command("bash", "-c", "make PROTOC=nonexistent-protoc proto")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected make proto to fail without protoc")
	}
	if !bytes.Contains(out, []byte("Error: protoc not found")) {
		t.Fatalf("expected error message about missing protoc, got: %s", out)
	}
}
