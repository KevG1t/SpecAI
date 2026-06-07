package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/components/sddmemory"
	"github.com/KevG1t/specai/internal/system"
)

// TestRunInstallLinuxSddMemoryUsesDownloadNotGoInstall verifies that after the fix,
// Linux sdd-memory installation does NOT use "go install" but instead calls
// DownloadLatestBinary (i.e. no "go install" in recorder.get()).
func TestRunInstallLinuxSddMemoryUsesDownloadNotGoInstall(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
	recorder := &commandRecorder{}
	runCommand = recorder.record

	// Override the sdd-memory download function to succeed without hitting GitHub.
	origDownloadFn := sddMemoryDownloadFn
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		// Simulate a successful binary download to a temp path.
		return "/tmp/fake-sdd-memory", nil
	}
	t.Cleanup(func() { sddMemoryDownloadFn = origDownloadFn })

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "sdd-memory"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	// Must NOT have called "go install" for sdd-memory.
	for _, cmd := range recorder.get() {
		if strings.Contains(cmd, "go install") && strings.Contains(cmd, "sdd-memory") {
			t.Fatalf("Linux sdd-memory install should NOT use go install, got command: %s", cmd)
		}
	}
}

// TestRunInstallSddMemoryDownloadAddsBinDirToPath verifies that after downloading
// the sdd-memory binary, its directory is prepended to PATH so that subsequent
// commands (sdd-memory setup, resolveSddMemoryCommand) can find it.
func TestRunInstallSddMemoryDownloadAddsBinDirToPath(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	restorePath := os.Getenv("PATH")
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
		os.Setenv("PATH", restorePath)
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
	recorder := &commandRecorder{}
	runCommand = recorder.record

	fakeBinDir := filepath.Join(home, "sdd-memory-bin")
	os.MkdirAll(fakeBinDir, 0o755)
	fakeBinaryPath := filepath.Join(fakeBinDir, "sdd-memory")

	origDownloadFn := sddMemoryDownloadFn
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		return fakeBinaryPath, nil
	}
	t.Cleanup(func() { sddMemoryDownloadFn = origDownloadFn })

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "sdd-memory"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, fakeBinDir) {
		t.Fatalf("PATH should contain sdd-memory bin dir %q after download, got PATH=%q", fakeBinDir, currentPath)
	}
}

// TestRunInstallWindowsSddMemoryUsesDownloadNotGoInstall verifies Windows path.
func TestRunInstallWindowsSddMemoryUsesDownloadNotGoInstall(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
	recorder := &commandRecorder{}
	runCommand = recorder.record

	origDownloadFn := sddMemoryDownloadFn
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		return `C:\fake\sdd-memory.exe`, nil
	}
	t.Cleanup(func() { sddMemoryDownloadFn = origDownloadFn })

	detection := system.DetectionResult{
		System: system.SystemInfo{
			OS:        "windows",
			Arch:      "amd64",
			Supported: true,
			Profile: system.PlatformProfile{
				OS:             "windows",
				PackageManager: "winget",
				Supported:      true,
			},
		},
	}

	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "sdd-memory"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	// Must NOT have called "go install" for sdd-memory.
	for _, cmd := range recorder.get() {
		if strings.Contains(cmd, "go install") && strings.Contains(cmd, "sdd-memory") {
			t.Fatalf("Windows sdd-memory install should NOT use go install, got command: %s", cmd)
		}
	}
}

// TestRunInstallMacOSSddMemoryStillUsesBrew verifies macOS unchanged.
func TestRunInstallMacOSSddMemoryStillUsesBrew(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
	recorder := &commandRecorder{}
	runCommand = recorder.record

	// DownloadFn should NOT be called for macOS (brew handles it).
	origDownloadFn := sddMemoryDownloadFn
	sddMemoryDownloadFn = func(profile system.PlatformProfile) (string, error) {
		t.Error("DownloadLatestBinary should NOT be called on macOS (brew handles it)")
		return "", nil
	}
	t.Cleanup(func() { sddMemoryDownloadFn = origDownloadFn })

	detection := macOSDetectionResult()
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "sdd-memory"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}
	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	// Must use brew install sdd-memory.
	commands := recorder.get()
	foundBrew := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "brew install sdd-memory") {
			foundBrew = true
		}
	}
	if !foundBrew {
		t.Fatalf("expected brew install sdd-memory on macOS, got commands: %v", commands)
	}
}

// Make sure the sddmemory package's DownloadLatestBinary is accessible.
var _ = sddmemory.DownloadLatestBinary
