package cli

import (
	"context"
	"os"
	"testing"
)

// TestMain stubs the sdd-memory version probe for the whole cli test package so
// tests never invoke the real sdd-memory binary. On a machine where sdd-memory
// is installed, the real "sdd-memory version" command can hang, which would
// otherwise freeze RunInstall verification tests until the suite timeout.
// Individual tests can still override sddMemoryVerifyVersionFn locally.
func TestMain(m *testing.M) {
	sddMemoryVerifyVersionFn = func(context.Context) (string, error) {
		return "0.0.0-test", nil
	}
	os.Exit(m.Run())
}
