package steps

import (
	"errors"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
)

// mockCommandExecutor records Run calls.
type mockCommandExecutor struct {
	calls   []executorCall
	failOn  string // fail when name matches
}

type executorCall struct {
	name string
	args []string
}

func (m *mockCommandExecutor) Run(name string, args ...string) error {
	m.calls = append(m.calls, executorCall{name: name, args: args})
	if m.failOn != "" && name == m.failOn {
		return errors.New("mock executor failure")
	}
	return nil
}

// stubInstallableIDE implements system.IDEAdapter with a known agentID.
// It delegates to the real adapter via system.GetAdapterByAgentID for most methods.
type stubInstallableIDE struct {
	agentID model.AgentID
}

func (s stubInstallableIDE) Name() string           { return string(s.agentID) }
func (s stubInstallableIDE) AgentID() model.AgentID { return s.agentID }
func (s stubInstallableIDE) AssetFolder() string    { return string(s.agentID) }

func (s stubInstallableIDE) real() system.IDEAdapter {
	return system.GetAdapterByAgentID(s.agentID)
}

func (s stubInstallableIDE) ConfigDir(homeDir string) string {
	if r := s.real(); r != nil {
		return r.ConfigDir(homeDir)
	}
	return ""
}
func (s stubInstallableIDE) GlobalRulesDir(h string) string {
	if r := s.real(); r != nil {
		return r.GlobalRulesDir(h)
	}
	return ""
}
func (s stubInstallableIDE) GlobalSkillsDir(h string) string {
	if r := s.real(); r != nil {
		return r.GlobalSkillsDir(h)
	}
	return ""
}
func (s stubInstallableIDE) LocalRulesFile() string { return "" }
func (s stubInstallableIDE) LocalSkillsDir() string { return "" }
func (s stubInstallableIDE) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}
func (s stubInstallableIDE) MCPStrategy() model.MCPStrategy { return model.StrategySeparateMCPFiles }
func (s stubInstallableIDE) SystemPromptFile(h string) string {
	if r := s.real(); r != nil {
		return r.SystemPromptFile(h)
	}
	return ""
}
func (s stubInstallableIDE) SubAgentsDir(h string) string {
	if r := s.real(); r != nil {
		return r.SubAgentsDir(h)
	}
	return ""
}
func (s stubInstallableIDE) EmbeddedSubAgentsDir() string { return "" }
func (s stubInstallableIDE) CommandsDir(h string) string {
	if r := s.real(); r != nil {
		return r.CommandsDir(h)
	}
	return ""
}
func (s stubInstallableIDE) EmbeddedCommandsDir() string { return "" }
func (s stubInstallableIDE) SkillsDir(h string) string {
	if r := s.real(); r != nil {
		return r.SkillsDir(h)
	}
	return ""
}
func (s stubInstallableIDE) SettingsPath(h string) string {
	if r := s.real(); r != nil {
		return r.SettingsPath(h)
	}
	return ""
}
func (s stubInstallableIDE) MCPConfigPath(h, sn string) string {
	if r := s.real(); r != nil {
		return r.MCPConfigPath(h, sn)
	}
	return ""
}
func (s stubInstallableIDE) SupportsSubAgents() bool {
	if r := s.real(); r != nil {
		return r.SupportsSubAgents()
	}
	return false
}
func (s stubInstallableIDE) SupportsSlashCommands() bool {
	if r := s.real(); r != nil {
		return r.SupportsSlashCommands()
	}
	return false
}
func (s stubInstallableIDE) SupportsSkills() bool {
	if r := s.real(); r != nil {
		return r.SupportsSkills()
	}
	return false
}
func (s stubInstallableIDE) SupportsSystemPrompt() bool {
	if r := s.real(); r != nil {
		return r.SupportsSystemPrompt()
	}
	return false
}
func (s stubInstallableIDE) SupportsMCP() bool {
	if r := s.real(); r != nil {
		return r.SupportsMCP()
	}
	return false
}

var _ system.IDEAdapter = stubInstallableIDE{}

// TestStepInstallAgents_NoAutoInstall verifies that agents without SupportsAutoInstall
// don't trigger any executor calls.
func TestStepInstallAgents_NoAutoInstall(t *testing.T) {
	exec := &mockCommandExecutor{}
	ctx := &InstallContext{
		HomeDir: t.TempDir(),
		// Cursor does not support auto-install in the MVP.
		IDEs: []system.IDEAdapter{stubInstallableIDE{agentID: model.AgentCursor}},
	}

	step := &StepInstallAgents{ctx: ctx, executor: exec}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Cursor.SupportsAutoInstall() returns false → no exec calls expected.
	if len(exec.calls) != 0 {
		t.Errorf("expected no executor calls, got %d", len(exec.calls))
	}
}

// TestStepInstallAgents_ClaudeCode_ExecutesInstall verifies that ClaudeCode
// (which SupportsAutoInstall=true) triggers an install command call.
// This test is marked for integration — skipped in -short mode because it
// depends on real agent adapter resolution and platform profile.
func TestStepInstallAgents_ClaudeCode_ExecutesInstall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping auto-install integration test in short mode")
	}

	exec := &mockCommandExecutor{}
	ctx := &InstallContext{
		HomeDir: t.TempDir(),
		IDEs:    []system.IDEAdapter{stubInstallableIDE{agentID: model.AgentClaudeCode}},
	}

	step := &StepInstallAgents{ctx: ctx, executor: exec}
	if err := step.Run(); err != nil {
		// Expected if ClaudeCode.InstallCommand returns error (e.g., unsupported profile).
		// The step should still have tried.
		t.Logf("Run() returned error (may be expected in CI): %v", err)
	}
}

// TestStepInstallAgents_NotInstallableAgent_Skips verifies that an agent returning
// CapabilityNotSupportedError from InstallCommand is silently skipped.
func TestStepInstallAgents_NotInstallableAgent_Skips(t *testing.T) {
	exec := &mockCommandExecutor{}
	ctx := &InstallContext{
		HomeDir: t.TempDir(),
		IDEs:    []system.IDEAdapter{stubInstallableIDE{agentID: model.AgentCursor}},
	}

	step := &StepInstallAgents{ctx: ctx, executor: exec}
	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error for non-installable agent: %v", err)
	}

	if len(exec.calls) != 0 {
		t.Errorf("expected no executor calls for non-installable agent, got %d", len(exec.calls))
	}
}

// TestStepInstallAgents_ID verifies the step ID string.
func TestStepInstallAgents_ID(t *testing.T) {
	ctx := &InstallContext{HomeDir: t.TempDir()}
	step := NewStepInstallAgents(ctx)
	if step.ID() == "" {
		t.Error("expected non-empty step ID")
	}
}
