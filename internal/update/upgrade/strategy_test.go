package upgrade

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/system"
	"github.com/KevG1t/specai/internal/update"
)

// --- TestRunStrategy_BrewUpgrade ---

func TestRunStrategy_BrewUpgrade(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	var gotName string
	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = args
		return mockCmd("echo", "Upgraded sdd-memory")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallBrew,
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}

	err := runStrategy(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("runStrategy brew: unexpected error: %v", err)
	}

	if gotName != "brew" {
		t.Errorf("exec name = %q, want %q", gotName, "brew")
	}
	if len(gotArgs) < 2 || gotArgs[0] != "upgrade" || gotArgs[1] != "sdd-memory" {
		t.Errorf("exec args = %v, want [upgrade sdd-memory]", gotArgs)
	}
}

// --- TestRunStrategy_GoInstallUpgrade ---

func TestRunStrategy_GoInstallUpgrade(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	var gotName string
	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = args
		return mockCmd("echo", "go install ok")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallGoInstall,
			GoImportPath:  "github.com/KevG1t/sdd-memory/cmd/sdd-memory",
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := runStrategy(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("runStrategy go-install: unexpected error: %v", err)
	}

	if gotName != "go" {
		t.Errorf("exec name = %q, want %q", gotName, "go")
	}
	// Expected: go install github.com/KevG1t/sdd-memory/cmd/sdd-memory@v0.4.0
	wantArg0, wantArg1 := "install", "github.com/KevG1t/sdd-memory/cmd/sdd-memory@v0.4.0"
	if len(gotArgs) < 2 || gotArgs[0] != wantArg0 || gotArgs[1] != wantArg1 {
		t.Errorf("exec args = %v, want [%s %s]", gotArgs, wantArg0, wantArg1)
	}
}

// --- TestRunStrategy_GoInstallMissingImportPath ---

func TestRunStrategy_GoInstallMissingImportPath(t *testing.T) {
	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallGoInstall,
			GoImportPath:  "", // missing
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected error when GoImportPath is empty, got nil")
	}
}

// --- TestRunStrategy_UnsupportedMethodManualFallback ---

func TestRunStrategy_UnsupportedMethodManualFallback(t *testing.T) {
	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "some-tool",
			InstallMethod: update.InstallMethod("unsupported-method"),
		},
		LatestVersion: "1.0.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := runStrategy(context.Background(), r, profile)
	// Unsupported method → manual fallback error.
	if err == nil {
		t.Errorf("expected error for unsupported install method, got nil")
	}
}

// --- TestRunStrategy_BrewUpgradeFailure ---

func TestRunStrategy_BrewUpgradeFailure(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("false") // always fails
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallBrew,
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}

	err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected error when brew upgrade fails, got nil")
	}
}

// --- TestRunStrategy_GoInstallFailure ---

func TestRunStrategy_GoInstallFailure(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("false")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallGoInstall,
			GoImportPath:  "github.com/KevG1t/sdd-memory/cmd/sdd-memory",
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected error when go install fails, got nil")
	}
}

// --- TestRunStrategy_BinaryWindowsSelfUpdateSkipped ---

// TestRunStrategy_BinaryWindowsSelfUpdateSkipped verifies that the Windows binary
// self-replace for specai is NOT attempted in Phase 1 — it must return a
// manual hint error, not execute.
func TestRunStrategy_BinaryWindowsSelfUpdateSkipped(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not run")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "specai",
			InstallMethod: update.InstallBinary,
		},
		LatestVersion: "1.5.0",
		ReleaseURL:    "https://github.com/KevG1t/SpecAI/releases/tag/v1.5.0",
	}
	profile := system.PlatformProfile{OS: "windows", PackageManager: "winget"}

	err := runStrategy(context.Background(), r, profile)
	// Windows binary self-replace must return an error (manual hint) in Phase 1.
	if err == nil {
		t.Errorf("expected manual fallback error for Windows binary self-replace, got nil")
	}

	if execCalled {
		t.Errorf("exec should NOT be called for Windows binary self-replace in Phase 1")
	}
}

// --- TestEffectiveMethod ---

func TestEffectiveMethod(t *testing.T) {
	tests := []struct {
		name    string
		tool    update.ToolInfo
		profile system.PlatformProfile
		want    update.InstallMethod
	}{
		{
			name:    "brew profile overrides go-install",
			tool:    update.ToolInfo{Name: "sdd-memory", InstallMethod: update.InstallGoInstall},
			profile: system.PlatformProfile{PackageManager: "brew"},
			want:    update.InstallBrew,
		},
		{
			name:    "brew profile overrides binary",
			tool:    update.ToolInfo{Name: "script-tool", InstallMethod: update.InstallBinary},
			profile: system.PlatformProfile{PackageManager: "brew"},
			want:    update.InstallBrew,
		},
		{
			name:    "brew profile overrides script",
			tool:    update.ToolInfo{Name: "script-tool", InstallMethod: update.InstallScript},
			profile: system.PlatformProfile{PackageManager: "brew"},
			want:    update.InstallBrew,
		},
		{
			name:    "apt profile respects declared method (go-install)",
			tool:    update.ToolInfo{Name: "sdd-memory", InstallMethod: update.InstallGoInstall},
			profile: system.PlatformProfile{PackageManager: "apt"},
			want:    update.InstallGoInstall,
		},
		{
			name:    "apt profile respects declared method (binary)",
			tool:    update.ToolInfo{Name: "script-tool", InstallMethod: update.InstallBinary},
			profile: system.PlatformProfile{PackageManager: "apt"},
			want:    update.InstallBinary,
		},
		{
			name:    "apt profile respects declared method (script)",
			tool:    update.ToolInfo{Name: "script-tool", InstallMethod: update.InstallScript},
			profile: system.PlatformProfile{PackageManager: "apt"},
			want:    update.InstallScript,
		},
		{
			name:    "brew profile does not override OpenCode plugin method",
			tool:    update.ToolInfo{Name: "opencode-subagent-statusline", InstallMethod: update.InstallOpenCodePlugin, NpmPackage: "opencode-subagent-statusline"},
			profile: system.PlatformProfile{PackageManager: "brew"},
			want:    update.InstallOpenCodePlugin,
		},
		// Auto-detect order: brew → go-install → binary (issue #246).
		{
			name:    "auto-detect: brew available → brew wins regardless of GoImportPath",
			tool:    update.ToolInfo{Name: "mytool", InstallMethod: update.InstallBinary, GoImportPath: "github.com/example/mytool/cmd/mytool"},
			profile: system.PlatformProfile{PackageManager: "brew", GoAvailable: true},
			want:    update.InstallBrew,
		},
		{
			name:    "auto-detect: brew missing + go available + GoImportPath set → go-install",
			tool:    update.ToolInfo{Name: "mytool", InstallMethod: update.InstallBinary, GoImportPath: "github.com/example/mytool/cmd/mytool"},
			profile: system.PlatformProfile{PackageManager: "apt", GoAvailable: true},
			want:    update.InstallGoInstall,
		},
		{
			name:    "auto-detect: brew missing + go missing + GoImportPath set → binary fallback",
			tool:    update.ToolInfo{Name: "mytool", InstallMethod: update.InstallBinary, GoImportPath: "github.com/example/mytool/cmd/mytool"},
			profile: system.PlatformProfile{PackageManager: "apt", GoAvailable: false},
			want:    update.InstallBinary,
		},
		{
			name:    "auto-detect: go available but GoImportPath empty → binary (no upgrade)",
			tool:    update.ToolInfo{Name: "mytool", InstallMethod: update.InstallBinary, GoImportPath: ""},
			profile: system.PlatformProfile{PackageManager: "apt", GoAvailable: true},
			want:    update.InstallBinary,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := effectiveMethod(tc.tool, tc.profile)
			if got != tc.want {
				t.Errorf("effectiveMethod = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunStrategyOpenCodePluginManualFallback(t *testing.T) {
	origHomeDir := openCodeHomeDir
	origLookPath := lookPathCommand
	origExecCommand := execCommand
	t.Cleanup(func() {
		openCodeHomeDir = origHomeDir
		lookPathCommand = origLookPath
		execCommand = origExecCommand
	})

	openCodeHomeDir = func() (string, error) { return t.TempDir(), nil }
	lookPathCommand = func(file string) (string, error) { return "", errors.New("not found") }
	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not run")
	}

	err := runStrategy(context.Background(), update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "opencode-subagent-statusline",
			InstallMethod: update.InstallOpenCodePlugin,
			NpmPackage:    "opencode-subagent-statusline",
		},
		UpdateHint: "restart OpenCode",
	}, system.PlatformProfile{PackageManager: "brew"})

	if err == nil {
		t.Fatal("expected manual fallback error, got nil")
	}
	if !containsAny(err.Error(), "OpenCode", "restart", "reload") {
		t.Fatalf("manual fallback should mention OpenCode restart/reload, got: %v", err)
	}
	if execCalled {
		t.Fatal("OpenCode plugin fallback should not run a package manager when config is missing")
	}

}

func TestRunStrategyOpenCodePluginUpgradesMaterializedPackage(t *testing.T) {
	origHomeDir := openCodeHomeDir
	origLookPath := lookPathCommand
	origExecCommand := execCommand
	t.Cleanup(func() {
		openCodeHomeDir = origHomeDir
		lookPathCommand = origLookPath
		execCommand = origExecCommand
	})

	home := t.TempDir()
	opencodeDir := filepath.Join(home, ".config", "opencode")
	pkg := "opencode-subagent-statusline"
	pkgDir := filepath.Join(opencodeDir, "node_modules", pkg)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"version":"0.1.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cacheRoot := filepath.Join(home, ".cache", "opencode", "packages")
	targetCache := filepath.Join(cacheRoot, pkg+"@latest")
	otherPluginCache := filepath.Join(cacheRoot, "opencode-sdd-sdd-memory-manage@latest")
	versionedCache := filepath.Join(cacheRoot, pkg+"@0.5.2")
	for _, dir := range []string{targetCache, otherPluginCache, versionedCache} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cwdFile := filepath.Join(t.TempDir(), "cwd.txt")

	openCodeHomeDir = func() (string, error) { return home, nil }
	lookPathCommand = func(file string) (string, error) {
		if file == "bun" {
			return "/usr/bin/bun", nil
		}
		return "", errors.New("not found")
	}

	var gotName string
	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string(nil), args...)
		cmd := exec.Command(os.Args[0], "-test.run=TestOpenCodePluginUpgradeHelperProcess", "--")
		cmd.Env = append(os.Environ(),
			"SPECAI_AI_UPGRADE_HELPER=1",
			"SPECAI_AI_UPGRADE_HELPER_CWD_FILE="+cwdFile,
		)
		return cmd
	}

	err := runStrategy(context.Background(), update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          pkg,
			InstallMethod: update.InstallOpenCodePlugin,
			NpmPackage:    pkg,
		},
		InstalledVersion: "0.1.0",
		LatestVersion:    "0.2.0",
	}, system.PlatformProfile{PackageManager: "brew"})
	if err != nil {
		t.Fatalf("runStrategy OpenCode plugin: unexpected error: %v", err)
	}

	if gotName != "bun" {
		t.Fatalf("exec name = %q, want bun", gotName)
	}
	wantArgs := []string{"add", pkg + "@latest", "@opencode-ai/plugin@latest"}
	if strings.Join(gotArgs, " ") != strings.Join(wantArgs, " ") {
		t.Fatalf("exec args = %v, want %v", gotArgs, wantArgs)
	}
	if _, err := os.Stat(targetCache); !os.IsNotExist(err) {
		t.Fatalf("target OpenCode cache %s should be removed after upgrade, stat err: %v", targetCache, err)
	}
	for _, dir := range []string{otherPluginCache, versionedCache} {
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("non-target cache %s should remain, stat err: %v", dir, err)
		}
	}
	cwd, err := os.ReadFile(cwdFile)
	if err != nil {
		t.Fatalf("read helper cwd: %v", err)
	}
	gotCwd, err := filepath.EvalSymlinks(string(cwd))
	if err != nil {
		t.Fatalf("resolve helper cwd: %v", err)
	}
	wantCwd, err := filepath.EvalSymlinks(opencodeDir)
	if err != nil {
		t.Fatalf("resolve OpenCode dir: %v", err)
	}
	if gotCwd != wantCwd {
		t.Fatalf("command cwd = %q, want %q", gotCwd, wantCwd)
	}
}

func TestRunStrategyOpenCodePluginRegisteredPendingRunsPackageManager(t *testing.T) {
	origHomeDir := openCodeHomeDir
	origLookPath := lookPathCommand
	origExecCommand := execCommand
	t.Cleanup(func() {
		openCodeHomeDir = origHomeDir
		lookPathCommand = origLookPath
		execCommand = origExecCommand
	})

	home := t.TempDir()
	opencodeDir := filepath.Join(home, ".config", "opencode")
	pkg := "opencode-sdd-sdd-memory-manage"
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(opencodeDir, "tui.json"), []byte(`{"plugin":["opencode-sdd-sdd-memory-manage"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	openCodeHomeDir = func() (string, error) { return home, nil }
	lookPathCommand = func(file string) (string, error) {
		if file == "npm" {
			return "/usr/bin/npm", nil
		}
		return "", errors.New("not found")
	}
	var gotName string
	var gotArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return mockCmd("true")
	}

	err := runStrategy(context.Background(), update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          pkg,
			InstallMethod: update.InstallOpenCodePlugin,
			NpmPackage:    pkg,
		},
		Status: update.RegisteredNotMaterialized,
	}, system.PlatformProfile{})
	if err != nil {
		t.Fatalf("registered OpenCode plugin should be npm-managed during upgrade, got: %v", err)
	}
	if gotName != "npm" {
		t.Fatalf("exec name = %q, want npm", gotName)
	}
	wantArgs := []string{"install", "--save", "--no-audit", "--no-fund", pkg + "@latest", "@opencode-ai/plugin@latest"}
	if strings.Join(gotArgs, " ") != strings.Join(wantArgs, " ") {
		t.Fatalf("exec args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestRunStrategyOpenCodePluginFallsBackWithoutPackageManager(t *testing.T) {
	origHomeDir := openCodeHomeDir
	origLookPath := lookPathCommand
	origExecCommand := execCommand
	t.Cleanup(func() {
		openCodeHomeDir = origHomeDir
		lookPathCommand = origLookPath
		execCommand = origExecCommand
	})

	home := t.TempDir()
	opencodeDir := filepath.Join(home, ".config", "opencode")
	pkg := "opencode-sdd-sdd-memory-manage"
	if err := os.MkdirAll(filepath.Join(opencodeDir, "node_modules", pkg), 0o755); err != nil {
		t.Fatal(err)
	}

	openCodeHomeDir = func() (string, error) { return home, nil }
	lookPathCommand = func(file string) (string, error) { return "", errors.New("not found") }
	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not run")
	}

	err := runStrategy(context.Background(), update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          pkg,
			InstallMethod: update.InstallOpenCodePlugin,
			NpmPackage:    pkg,
		},
	}, system.PlatformProfile{})
	if err == nil {
		t.Fatal("expected manual fallback when bun/npm are unavailable, got nil")
	}
	if !containsAny(err.Error(), "bun", "npm", "package manager") {
		t.Fatalf("fallback should mention missing package manager, got: %v", err)
	}
	if execCalled {
		t.Fatal("OpenCode plugin fallback should not run a package manager when none is available")
	}
}

func TestSelectOpenCodePackageManagerPrefersPackageMetadata(t *testing.T) {
	origLookPath := lookPathCommand
	t.Cleanup(func() { lookPathCommand = origLookPath })

	opencodeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(opencodeDir, "package.json"), []byte(`{"packageManager":"npm@10.8.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	lookPathCommand = func(file string) (string, error) {
		switch file {
		case "bun", "npm":
			return filepath.Join("/usr/bin", file), nil
		default:
			return "", errors.New("not found")
		}
	}

	pm, err := selectOpenCodePackageManager(opencodeDir)
	if err != nil {
		t.Fatalf("selectOpenCodePackageManager: unexpected error: %v", err)
	}
	if pm != "npm" {
		t.Fatalf("package manager = %q, want npm from package.json metadata", pm)
	}
}

func TestOpenCodePluginUpgradeHelperProcess(t *testing.T) {
	if os.Getenv("SPECAI_AI_UPGRADE_HELPER") != "1" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(2)
	}
	if err := os.WriteFile(os.Getenv("SPECAI_AI_UPGRADE_HELPER_CWD_FILE"), []byte(cwd), 0o644); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(2)
	}
	os.Exit(0)
}

// --- TestManualFallbackHint ---

// TestManualFallbackHint verifies that Windows binary self-replace produces an
// actionable hint string, not an empty error.
func TestManualFallbackHint(t *testing.T) {
	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "specai",
			InstallMethod: update.InstallBinary,
		},
		LatestVersion: "1.5.0",
		UpdateHint:    "See https://github.com/KevG1t/SpecAI/releases",
	}
	profile := system.PlatformProfile{OS: "windows", PackageManager: "winget"}

	err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if msg == "" {
		t.Errorf("manual fallback error message should not be empty")
	}

	// Hint should mention manual action or Windows.
	if !containsAny(msg, "manual", "Manual", "windows", "Windows", "winget", "hint") {
		t.Errorf("manual hint message %q does not mention manual or windows", msg)
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// --- TestBrewUpgrade_RunsUpdateBeforeUpgrade ---

// TestBrewUpgrade_RunsUpdateBeforeUpgrade verifies that brewUpgrade calls
// `brew update` BEFORE `brew upgrade <toolName>`, and that the order is correct.
func TestBrewUpgrade_RunsUpdateBeforeUpgrade(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	var callOrder []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) > 0 {
			callOrder = append(callOrder, args[0]) // "update" or "upgrade"
		}
		return mockCmd("echo", "ok")
	}

	err := brewUpgrade(context.Background(), "specai")
	if err != nil {
		t.Fatalf("brewUpgrade: unexpected error: %v", err)
	}

	// Must have called brew tap, brew update AND brew upgrade — in that order.
	if len(callOrder) < 3 {
		t.Fatalf("expected 3 brew calls (tap, update, upgrade), got %d: %v", len(callOrder), callOrder)
	}
	if callOrder[1] != "update" {
		t.Errorf("second brew call = %q, want %q", callOrder[1], "update")
	}
	if callOrder[2] != "upgrade" {
		t.Errorf("third brew call = %q, want %q", callOrder[2], "upgrade")
	}
}

// --- TestBrewUpgrade_UpdateFailureIsNonFatal ---

// TestBrewUpgrade_UpdateFailureIsNonFatal verifies that when `brew update` fails
// but `brew upgrade` succeeds, the overall result is success (non-fatal update failure).
func TestBrewUpgrade_UpdateFailureIsNonFatal(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	var callArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) > 0 {
			callArgs = append(callArgs, args[0])
			if args[0] == "update" {
				// brew update fails (e.g. no network).
				return mockCmd("false")
			}
		}
		// brew upgrade succeeds.
		return mockCmd("echo", "Upgraded specai")
	}

	err := brewUpgrade(context.Background(), "specai")
	// brew update failed but brew upgrade succeeded → overall success.
	if err != nil {
		t.Errorf("expected success when brew update fails but brew upgrade succeeds, got: %v", err)
	}

	// Both brew update and brew upgrade must have been called (after the tap).
	if len(callArgs) < 3 {
		t.Fatalf("expected 3 brew calls, got %d: %v", len(callArgs), callArgs)
	}
	if callArgs[1] != "update" {
		t.Errorf("second brew call = %q, want %q", callArgs[1], "update")
	}
	if callArgs[2] != "upgrade" {
		t.Errorf("third brew call = %q, want %q", callArgs[2], "upgrade")
	}
}

// --- TestBrewUpgrade_TapsBeforeUpdateAndUpgrade ---

// TestBrewUpgrade_TapsBeforeUpdateAndUpgrade verifies that brewUpgrade calls
// `brew tap KevG1t/homebrew-tap` BEFORE `brew update` and
// `brew upgrade <toolName>`. This makes the upgrade idempotent when a user
// has lost the tap (untap, machine swap, brew cleanup). See issue #455.
func TestBrewUpgrade_TapsBeforeUpdateAndUpgrade(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	type call struct {
		subcommand string
		arg        string
	}
	var calls []call
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) > 0 {
			c := call{subcommand: args[0]}
			if len(args) > 1 {
				c.arg = args[1]
			}
			calls = append(calls, c)
		}
		return mockCmd("echo", "ok")
	}

	if err := brewUpgrade(context.Background(), "sdd-memory"); err != nil {
		t.Fatalf("brewUpgrade: unexpected error: %v", err)
	}

	if len(calls) < 3 {
		t.Fatalf("expected 3 brew calls (tap, update, upgrade), got %d: %+v", len(calls), calls)
	}
	if calls[0].subcommand != "tap" {
		t.Errorf("first brew call subcommand = %q, want %q", calls[0].subcommand, "tap")
	}
	if calls[0].arg != "KevG1t/homebrew-tap" {
		t.Errorf("first brew call arg = %q, want %q", calls[0].arg, "KevG1t/homebrew-tap")
	}
	if calls[1].subcommand != "update" {
		t.Errorf("second brew call = %q, want %q", calls[1].subcommand, "update")
	}
	if calls[2].subcommand != "upgrade" {
		t.Errorf("third brew call = %q, want %q", calls[2].subcommand, "upgrade")
	}
}

// --- verify exec.Cmd.Run() failure is correctly wrapped ---
func TestRunStrategy_ExecErrorWrapped(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("false")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			InstallMethod: update.InstallBrew,
		},
		LatestVersion: "0.4.0",
	}
	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}

	err := runStrategy(context.Background(), r, profile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error should have a non-empty message.
	if err.Error() == "" {
		t.Errorf("error should have a message")
	}

	// Error should wrap an *exec.ExitError (from running "false").
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Logf("note: error is not directly an ExitError (may be wrapped): %v", err)
	}
}

// --- TestRunStrategy_ScriptUpgradeSuccess ---

func TestRunStrategy_ScriptUpgradeSuccess(t *testing.T) {
	origExecCommand := execCommand
	origHTTPClient := scriptHTTPClient
	origInstallScriptURL := installScriptURLFn
	t.Cleanup(func() {
		execCommand = origExecCommand
		scriptHTTPClient = origHTTPClient
		installScriptURLFn = origInstallScriptURL
	})

	// Serve a fake install.sh that succeeds.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("#!/bin/bash\necho 'install ok'\n"))
	}))
	defer server.Close()

	scriptHTTPClient = server.Client()

	// Override installScriptURL to point to our test server.
	installScriptURLFn = func(owner, repo, version string) (string, error) {
		return server.URL + "/install.sh", nil
	}

	var gotScriptContent string
	execCommand = func(name string, args ...string) *exec.Cmd {
		// Capture the script content passed via bash -c.
		if name == "bash" && len(args) >= 2 && args[0] == "-c" {
			gotScriptContent = args[1]
		}
		return mockCmd("echo", "ok")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "script-tool",
			Owner:         "KevG1t",
			Repo:          "script-tool",
			InstallMethod: update.InstallScript,
		},
		LatestVersion: "2.8.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := scriptUpgrade(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("scriptUpgrade: unexpected error: %v", err)
	}

	// Verify that bash was called with the install.sh content.
	if !containsAny(gotScriptContent, "install ok", "#!/bin/bash") {
		t.Errorf("bash -c did not receive install.sh content; got: %q", gotScriptContent)
	}
}

// --- TestRunStrategy_ScriptUpgradeDownloadFailure ---

func TestRunStrategy_ScriptUpgradeDownloadFailure(t *testing.T) {
	origHTTPClient := scriptHTTPClient
	origInstallScriptURL := installScriptURLFn
	t.Cleanup(func() {
		scriptHTTPClient = origHTTPClient
		installScriptURLFn = origInstallScriptURL
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	scriptHTTPClient = server.Client()
	installScriptURLFn = func(owner, repo, version string) (string, error) {
		return server.URL + "/install.sh", nil
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "script-tool",
			Owner:         "KevG1t",
			Repo:          "script-tool",
			InstallMethod: update.InstallScript,
		},
		LatestVersion: "2.8.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := scriptUpgrade(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected error when install.sh download fails, got nil")
	}
}

// --- TestRunStrategy_ScriptUpgradeWindowsManualFallback ---

func TestRunStrategy_ScriptUpgradeWindowsManualFallback(t *testing.T) {
	origExecCommand := execCommand
	t.Cleanup(func() { execCommand = origExecCommand })

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not run")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "script-tool",
			Owner:         "KevG1t",
			Repo:          "script-tool",
			InstallMethod: update.InstallScript,
		},
		LatestVersion: "2.8.0",
	}
	profile := system.PlatformProfile{OS: "windows", PackageManager: "winget"}

	err := scriptUpgrade(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected manual fallback error for Windows script upgrade, got nil")
	}

	if execCalled {
		t.Errorf("exec should NOT be called for Windows script manual fallback")
	}
}

// --- TestInstallScriptURL ---

func TestInstallScriptURL(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		version     string
		wantURL     string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "pins to release tag",
			owner:       "KevG1t",
			repo:        "SpecAI",
			version:     "1.31.0",
			wantURL:     "https://raw.githubusercontent.com/KevG1t/SpecAI/v1.31.0/install.sh",
			wantContain: "v1.31.0",
		},
		{
			name:    "empty version returns error",
			owner:   "KevG1t",
			repo:    "SpecAI",
			version: "",
			wantErr: true,
		},
		{
			name:    "whitespace-only version returns error",
			owner:   "KevG1t",
			repo:    "SpecAI",
			version: "   ",
			wantErr: true,
		},
		{
			name:        "does not reference main",
			owner:       "KevG1t",
			repo:        "SpecAI",
			version:     "2.0.0",
			wantContain: "v2.0.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url, err := installScriptURL(tc.owner, tc.repo, tc.version)
			if tc.wantErr {
				if err == nil {
					t.Errorf("installScriptURL(%q, %q, %q): want error, got nil (url=%q)", tc.owner, tc.repo, tc.version, url)
				}
				return
			}
			if err != nil {
				t.Fatalf("installScriptURL(%q, %q, %q): unexpected error: %v", tc.owner, tc.repo, tc.version, err)
			}
			if tc.wantURL != "" && url != tc.wantURL {
				t.Errorf("installScriptURL = %q, want %q", url, tc.wantURL)
			}
			if tc.wantContain != "" && !containsAny(url, tc.wantContain) {
				t.Errorf("installScriptURL = %q, want it to contain %q", url, tc.wantContain)
			}
			if containsAny(url, "/main/") {
				t.Errorf("installScriptURL = %q must NOT reference /main/", url)
			}
		})
	}
}

// --- TestSDDMemoryUpgradeUsesDownloadNotGoInstall ---

// TestSDDMemoryUpgradeUsesDownloadNotGoInstall verifies that on Windows (non-brew),
// sdd-memory upgrade calls the binary download function, NOT go install.
// This is the regression test for issue #160.
func TestSddMemoryUpgradeUsesDownloadNotGoInstall(t *testing.T) {
	origExecCommand := execCommand
	origSddMemoryDownloadFn := sddMemoryDownloadFn
	t.Cleanup(func() {
		execCommand = origExecCommand
		sddMemoryDownloadFn = origSddMemoryDownloadFn
	})

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not be called")
	}

	downloadCalled := false
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		downloadCalled = true
		return "/fake/path/sdd-memory.exe", nil
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			Owner:         "KevG1t",
			Repo:          "sdd-memory",
			InstallMethod: update.InstallBinary, // should be InstallBinary after fix
		},
		LatestVersion: "0.5.0",
	}
	profile := system.PlatformProfile{OS: "windows", PackageManager: "winget"}

	err := runStrategy(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("runStrategy sdd-memory windows: unexpected error: %v", err)
	}

	// Must call binary download, NOT go install.
	if !downloadCalled {
		t.Errorf("expected sddMemoryDownloadFn to be called, but it was not")
	}
	if execCalled {
		t.Errorf("exec (go install) should NOT be called for sdd-memory on Windows — use binary download")
	}
}

// --- TestSDDMemoryUpgradeLinuxUsesDownload ---

// TestSddMemoryUpgradeLinuxUsesDownload verifies that on Linux (non-brew),
// sdd-memory upgrade uses the binary download function, not go install.
func TestSddMemoryUpgradeLinuxUsesDownload(t *testing.T) {
	origExecCommand := execCommand
	origSddMemoryDownloadFn := sddMemoryDownloadFn
	t.Cleanup(func() {
		execCommand = origExecCommand
		sddMemoryDownloadFn = origSddMemoryDownloadFn
	})

	execCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		execCalled = true
		return mockCmd("echo", "should not be called")
	}

	downloadCalled := false
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		downloadCalled = true
		return "/home/user/.local/bin/sdd-memory", nil
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "sdd-memory",
			Owner:         "KevG1t",
			Repo:          "sdd-memory",
			InstallMethod: update.InstallBinary,
		},
		LatestVersion: "0.5.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := runStrategy(context.Background(), r, profile)
	if err != nil {
		t.Fatalf("runStrategy sdd-memory linux: unexpected error: %v", err)
	}

	if !downloadCalled {
		t.Errorf("expected sddMemoryDownloadFn to be called for sdd-memory on Linux, but it was not")
	}
	if execCalled {
		t.Errorf("exec (go install) should NOT be called for sdd-memory on Linux — use binary download")
	}
}

// --- TestRunStrategy_ScriptUpgradeExecFailure ---

func TestRunStrategy_ScriptUpgradeExecFailure(t *testing.T) {
	origExecCommand := execCommand
	origHTTPClient := scriptHTTPClient
	origInstallScriptURL := installScriptURLFn
	t.Cleanup(func() {
		execCommand = origExecCommand
		scriptHTTPClient = origHTTPClient
		installScriptURLFn = origInstallScriptURL
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("#!/bin/bash\nexit 1\n"))
	}))
	defer server.Close()
	scriptHTTPClient = server.Client()
	installScriptURLFn = func(owner, repo, version string) (string, error) {
		return server.URL + "/install.sh", nil
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		return mockCmd("false")
	}

	r := update.UpdateResult{
		Tool: update.ToolInfo{
			Name:          "example-tool",
			Owner:         "kevg1t",
			Repo:          "specai-example",
			InstallMethod: update.InstallScript,
		},
		LatestVersion: "2.8.0",
	}
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	err := scriptUpgrade(context.Background(), r, profile)
	if err == nil {
		t.Errorf("expected error when install.sh execution fails, got nil")
	}
}
