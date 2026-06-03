package steps

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/system"
)

// OsCommandExecutor abstracts exec.Command for testability.
type OsCommandExecutor interface {
	Run(name string, args ...string) error
}

// defaultExecutor is the real implementation used in production.
type defaultExecutor struct{}

func (d defaultExecutor) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// StepInstallAgents iterates ctx.IDEs and executes the install command for each
// agent that SupportsAutoInstall(). Agents that return CapabilityNotSupportedError
// or AgentNotInstallableError are silently skipped.
type StepInstallAgents struct {
	ctx      *InstallContext
	executor OsCommandExecutor
}

// NewStepInstallAgents returns a production-ready StepInstallAgents.
func NewStepInstallAgents(ctx *InstallContext) *StepInstallAgents {
	return &StepInstallAgents{ctx: ctx, executor: defaultExecutor{}}
}

func (s *StepInstallAgents) ID() string {
	return "Instalando agentes (CLI installs)"
}

func (s *StepInstallAgents) Run() error {
	for _, ide := range s.ctx.IDEs {
		if err := s.installOne(ide); err != nil {
			s.ctx.Warnings = append(s.ctx.Warnings,
				fmt.Sprintf("install %s: %v", ide.AgentID(), err))
		}
	}
	return nil
}

func (s *StepInstallAgents) installOne(ide system.IDEAdapter) error {
	// Get the agents.Adapter to access SupportsAutoInstall and InstallCommand.
	adapter, err := agents.NewAdapter(ide.AgentID())
	if err != nil {
		// Unknown agent — skip silently.
		return nil
	}

	if !adapter.SupportsAutoInstall() {
		return nil
	}

	cmds, err := adapter.InstallCommand(s.ctx.Platform)
	if err != nil {
		// CapabilityNotSupportedError or similar — skip silently.
		if isSkippableInstallError(err) {
			return nil
		}
		return fmt.Errorf("resolve install command: %w", err)
	}

	if len(cmds) == 0 {
		return nil
	}

	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		if execErr := s.executor.Run(cmd[0], cmd[1:]...); execErr != nil {
			return fmt.Errorf("run %v: %w", cmd, execErr)
		}
	}
	return nil
}

func isSkippableInstallError(err error) bool {
	var capErr agents.CapabilityNotSupportedError
	if errors.As(err, &capErr) {
		return true
	}
	var notSupported agents.AgentNotSupportedError
	if errors.As(err, &notSupported) {
		return true
	}
	// Package-local AgentNotInstallableError types in adapter sub-packages (cursor,
	// windsurf, vscode, trae, antigravity, kiro, openclaw) are not checked here
	// because those adapters also return SupportsAutoInstall()=false, making
	// InstallCommand() unreachable before this guard. If a future adapter returns
	// true from SupportsAutoInstall() AND an AgentNotInstallableError, it should
	// either use agents.CapabilityNotSupportedError or expose a shared sentinel.
	return false
}
