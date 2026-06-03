package steps

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"
)

// OsCommandRunner is an injectable abstraction for launchctl/systemctl calls.
type OsCommandRunner interface {
	RunCommand(name string, args ...string) error
}

type defaultOsRunner struct{}

func (d defaultOsRunner) RunCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// StepInjectSDDMemoryService registers the sdd-memory binary as an OS daemon.
// macOS: writes a launchd plist and calls launchctl load.
// Linux: writes a systemd user unit and calls systemctl --user enable --now.
// Other OS: appends a warning and returns nil (no-op, soft skip).
type StepInjectSDDMemoryService struct {
	ctx    *InstallContext
	fs     afero.Fs
	runner OsCommandRunner
	goos   string // injected for testing; defaults to runtime.GOOS
}

// NewStepInjectSDDMemoryService returns the production-ready step.
func NewStepInjectSDDMemoryService(ctx *InstallContext) *StepInjectSDDMemoryService {
	return &StepInjectSDDMemoryService{
		ctx:    ctx,
		fs:     afero.NewOsFs(),
		runner: defaultOsRunner{},
		goos:   runtime.GOOS,
	}
}

func (s *StepInjectSDDMemoryService) ID() string {
	return "Registrando sdd-memory como servicio del sistema"
}

func (s *StepInjectSDDMemoryService) Run() error {
	switch s.goos {
	case "darwin":
		return s.runDarwin()
	case "linux":
		return s.runLinux()
	case "windows":
		return s.runWindows()
	default:
		s.ctx.Warnings = append(s.ctx.Warnings,
			fmt.Sprintf("SDD-memory daemon registration is not supported on %s — skipping", s.goos),
		)
		return nil
	}
}

const launchdPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.specai.sdd-memory</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>server</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <false/>
</dict>
</plist>
`

const systemdUnit = `[Unit]
Description=sdd-memory MCP server
After=network.target

[Service]
ExecStart=%s server
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

func (s *StepInjectSDDMemoryService) runDarwin() error {
	binaryPath := s.resolveBinaryPath()

	plistDir := filepath.Join(s.ctx.HomeDir, "Library", "LaunchAgents")
	if err := s.fs.MkdirAll(plistDir, 0755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}

	plistPath := filepath.Join(plistDir, "com.specai.sdd-memory.plist")
	content := []byte(fmt.Sprintf(launchdPlist, binaryPath))
	if err := afero.WriteFile(s.fs, plistPath, content, 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	if err := s.runner.RunCommand("launchctl", "load", plistPath); err != nil {
		s.ctx.Warnings = append(s.ctx.Warnings,
			fmt.Sprintf("launchctl load failed (non-fatal): %v", err),
		)
	}
	return nil
}

func (s *StepInjectSDDMemoryService) runLinux() error {
	binaryPath := s.resolveBinaryPath()

	unitDir := filepath.Join(s.ctx.HomeDir, ".config", "systemd", "user")
	if err := s.fs.MkdirAll(unitDir, 0755); err != nil {
		return fmt.Errorf("create systemd user dir: %w", err)
	}

	unitPath := filepath.Join(unitDir, "sdd-memory.service")
	content := []byte(fmt.Sprintf(systemdUnit, binaryPath))
	if err := afero.WriteFile(s.fs, unitPath, content, 0644); err != nil {
		return fmt.Errorf("write systemd unit: %w", err)
	}

	if err := s.runner.RunCommand("systemctl", "--user", "enable", "--now", "sdd-memory.service"); err != nil {
		s.ctx.Warnings = append(s.ctx.Warnings,
			fmt.Sprintf("systemctl enable failed (non-fatal): %v", err),
		)
	}
	return nil
}

func (s *StepInjectSDDMemoryService) runWindows() error {
	binaryPath := s.resolveBinaryPath()

	// Register as a Task Scheduler task that runs at user logon.
	// /f overwrites any existing task with the same name.
	err := s.runner.RunCommand("schtasks",
		"/create",
		"/tn", `SpecAI\sdd-memory`,
		"/tr", fmt.Sprintf(`"%s" server`, binaryPath),
		"/sc", "onlogon",
		"/rl", "limited",
		"/f",
	)
	if err != nil {
		s.ctx.Warnings = append(s.ctx.Warnings,
			fmt.Sprintf("schtasks create failed (non-fatal): %v", err),
		)
	}
	return nil
}

// resolveBinaryPath attempts to find the sdd-memory binary. Falls back gracefully.
func (s *StepInjectSDDMemoryService) resolveBinaryPath() string {
	if path, err := exec.LookPath("sdd-memory"); err == nil {
		return path
	}
	return "sdd-memory"
}
