package steps

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// spyRunner is a test double for OsCommandRunner that records calls.
type spyRunner struct {
	calls []spyCall
	errOn string // if non-empty, return error when name matches
}

type spyCall struct {
	name string
	args []string
}

func (s *spyRunner) RunCommand(name string, args ...string) error {
	s.calls = append(s.calls, spyCall{name: name, args: args})
	if s.errOn != "" && name == s.errOn {
		return fmt.Errorf("mock failure for %s", name)
	}
	return nil
}

func TestStepInjectSDDMemoryService_Darwin_WritesAndLoads(t *testing.T) {
	memFS := afero.NewMemMapFs()
	spy := &spyRunner{}
	ctx := &InstallContext{HomeDir: "/home/user"}

	step := &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     memFS,
		runner: spy,
		goos:   "darwin",
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Plist file should be written.
	plistPath := filepath.Join("/home/user", "Library", "LaunchAgents", "com.specai.sdd-memory.plist")
	exists, err := afero.Exists(memFS, plistPath)
	if err != nil {
		t.Fatalf("checking plist path: %v", err)
	}
	if !exists {
		t.Errorf("plist file not written at %s", plistPath)
	}

	// Plist content must contain required keys.
	data, _ := afero.ReadFile(memFS, plistPath)
	if !strings.Contains(string(data), "com.specai.sdd-memory") {
		t.Error("plist missing com.specai.sdd-memory label")
	}
	if !strings.Contains(string(data), "RunAtLoad") {
		t.Error("plist missing RunAtLoad key")
	}

	// launchctl must be called with the plist path.
	if len(spy.calls) == 0 {
		t.Fatal("expected launchctl call, got none")
	}
	if spy.calls[0].name != "launchctl" {
		t.Errorf("expected launchctl call, got %q", spy.calls[0].name)
	}
}

func TestStepInjectSDDMemoryService_Linux_WritesAndEnables(t *testing.T) {
	memFS := afero.NewMemMapFs()
	spy := &spyRunner{}
	ctx := &InstallContext{HomeDir: "/home/user"}

	step := &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     memFS,
		runner: spy,
		goos:   "linux",
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Service unit file should be written.
	unitPath := filepath.Join("/home/user", ".config", "systemd", "user", "sdd-memory.service")
	exists, err := afero.Exists(memFS, unitPath)
	if err != nil {
		t.Fatalf("checking unit path: %v", err)
	}
	if !exists {
		t.Errorf("unit file not written at %s", unitPath)
	}

	// Unit content must have [Service] and [Install] sections.
	data, _ := afero.ReadFile(memFS, unitPath)
	if !strings.Contains(string(data), "[Service]") {
		t.Error("unit missing [Service] section")
	}
	if !strings.Contains(string(data), "WantedBy=default.target") {
		t.Error("unit missing WantedBy=default.target")
	}

	// systemctl must be called.
	if len(spy.calls) == 0 {
		t.Fatal("expected systemctl call, got none")
	}
	if spy.calls[0].name != "systemctl" {
		t.Errorf("expected systemctl call, got %q", spy.calls[0].name)
	}
}

func TestStepInjectSDDMemoryService_Windows_CallsSchtasks(t *testing.T) {
	memFS := afero.NewMemMapFs()
	spy := &spyRunner{}
	ctx := &InstallContext{HomeDir: "/home/user"}

	step := &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     memFS,
		runner: spy,
		goos:   "windows",
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}

	if len(ctx.Warnings) != 0 {
		t.Errorf("expected no warnings on success, got: %v", ctx.Warnings)
	}

	if len(spy.calls) == 0 {
		t.Fatal("expected schtasks call, got none")
	}
	if spy.calls[0].name != "schtasks" {
		t.Errorf("expected schtasks call, got %q", spy.calls[0].name)
	}
}

func TestStepInjectSDDMemoryService_Windows_SchtasksError_SoftWarning(t *testing.T) {
	memFS := afero.NewMemMapFs()
	spy := &spyRunner{errOn: "schtasks"}
	ctx := &InstallContext{HomeDir: "/home/user"}

	step := &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     memFS,
		runner: spy,
		goos:   "windows",
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() should return nil even when schtasks fails, got: %v", err)
	}

	if len(ctx.Warnings) == 0 {
		t.Error("expected warning when schtasks fails, got none")
	}
}

func TestStepInjectSDDMemoryService_LaunchctlError_SoftWarning(t *testing.T) {
	memFS := afero.NewMemMapFs()
	spy := &spyRunner{errOn: "launchctl"}
	ctx := &InstallContext{HomeDir: "/home/user"}

	step := &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     memFS,
		runner: spy,
		goos:   "darwin",
	}

	// Should NOT return error — soft failure.
	if err := step.Run(); err != nil {
		t.Fatalf("Run() should return nil on launchctl error (soft), got: %v", err)
	}

	// Warning must be appended.
	if len(ctx.Warnings) == 0 {
		t.Error("expected warning on launchctl failure, got none")
	}
}
