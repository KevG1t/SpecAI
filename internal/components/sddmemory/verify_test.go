package sddmemory

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func TestVerifyInstalled(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = func(string) (string, error) { return "/opt/homebrew/bin/sdd-memory", nil }
	if err := VerifyInstalled(); err != nil {
		t.Fatalf("VerifyInstalled() error = %v", err)
	}

	lookPath = func(string) (string, error) { return "", errors.New("missing") }
	if err := VerifyInstalled(); err == nil {
		t.Fatalf("VerifyInstalled() expected missing binary error")
	}
}

// TestVerifyVersionBoundsContextAndDetachesStdin is the regression test for the
// frozen-install bug: a hung "sdd-memory version" (one that starts a server or
// blocks on stdin) used to block RunInstall forever. VerifyVersion must run the
// command with a bounded (deadline) context and with stdin detached.
func TestVerifyVersionBoundsContextAndDetachesStdin(t *testing.T) {
	original := commandContext
	t.Cleanup(func() { commandContext = original })

	var hadDeadline bool
	var captured *exec.Cmd
	commandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		_, hadDeadline = ctx.Deadline()
		// A binary that does not exist fails fast on Output(), so the test
		// never depends on a real sdd-memory and never hangs.
		captured = exec.Command("specai-nonexistent-binary-xyz")
		return captured
	}

	if _, err := VerifyVersion(context.Background()); err == nil {
		t.Fatal("VerifyVersion() expected error for missing binary")
	}
	if !hadDeadline {
		t.Fatal("VerifyVersion must run the command with a bounded context deadline")
	}
	if captured != nil && captured.Stdin != nil {
		t.Fatal("VerifyVersion must detach stdin (nil) so an interactive binary cannot block")
	}
}

func TestVerifyHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := VerifyHealth(context.Background(), server.URL); err != nil {
		t.Fatalf("VerifyHealth() error = %v", err)
	}

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer badServer.Close()

	if err := VerifyHealth(context.Background(), badServer.URL); err == nil {
		t.Fatalf("VerifyHealth() expected non-200 error")
	}
}
