package screens_test

import (
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/tui/screens"
)

// ─── WelcomeOptions ──────────────────────────────────────────────────────────

func TestWelcomeOptions_OptionCount(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	// Expected: Start installation, Upgrade tools, Sync configs, Upgrade + Sync,
	// Configure models, Manage backups, Managed uninstall, Quit = 8
	want := 8
	if len(opts) != want {
		t.Errorf("WelcomeOptions() = %d options, want %d; opts: %v", len(opts), want, opts)
	}
}

func TestWelcomeOptions_IncludesStartInstallation(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	if !containsOption(opts, "Start installation") {
		t.Fatalf("expected 'Start installation' option; got: %v", opts)
	}
}

func TestWelcomeOptions_IncludesManagedUninstall(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	if !containsOption(opts, "Managed uninstall") {
		t.Fatalf("expected 'Managed uninstall' option; got: %v", opts)
	}
}

func TestWelcomeOptions_IncludesManageBackups(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	if !containsOption(opts, "Manage backups") {
		t.Fatalf("expected 'Manage backups' option; got: %v", opts)
	}
}

func TestWelcomeOptions_DoesNotIncludeOpenCodePlugins(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	for _, opt := range opts {
		if strings.Contains(opt, "OpenCode") {
			t.Errorf("unexpected OpenCode option in welcome menu: %q", opt)
		}
	}
}

func TestWelcomeOptions_DoesNotIncludeAgentBuilder(t *testing.T) {
	opts := screens.WelcomeOptions(nil, true, false, 0, false)
	for _, opt := range opts {
		if strings.Contains(opt, "Agent") {
			t.Errorf("unexpected Agent option in welcome menu: %q", opt)
		}
	}
}

func containsOption(opts []string, want string) bool {
	for _, opt := range opts {
		if opt == want {
			return true
		}
	}
	return false
}

// ─── RenderWelcome ────────────────────────────────────────────────────────────

func TestRenderWelcome_ContainsStartInstallation(t *testing.T) {
	output := screens.RenderWelcome(0, "1.0.0", "", nil, true, false, 0, false)
	if !strings.Contains(output, "Start installation") {
		t.Errorf("RenderWelcome missing 'Start installation'")
	}
}

func TestRenderWelcome_ContainsSpecAITagline(t *testing.T) {
	output := screens.RenderWelcome(0, "1.0.0", "", nil, true, false, 0, false)
	if !strings.Contains(output, "SpecAI") {
		t.Errorf("RenderWelcome missing 'SpecAI' in tagline")
	}
}
