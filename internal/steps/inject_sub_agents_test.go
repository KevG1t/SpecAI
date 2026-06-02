package steps

import (
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

// subAgentsStubIDE is a minimal IDEAdapter stub for inject_sub_agents tests.
// New interface methods delegate to the real adapter so tests use real paths/flags.
type subAgentsStubIDE struct {
	agentID model.AgentID
}

func (s subAgentsStubIDE) Name() string                    { return string(s.agentID) }
func (s subAgentsStubIDE) AgentID() model.AgentID          { return s.agentID }
func (s subAgentsStubIDE) AssetFolder() string             { return "" }
func (s subAgentsStubIDE) ConfigDir(_ string) string       { return "" }
func (s subAgentsStubIDE) GlobalRulesDir(_ string) string  { return "" }
func (s subAgentsStubIDE) GlobalSkillsDir(_ string) string { return "" }
func (s subAgentsStubIDE) LocalRulesFile() string          { return "" }
func (s subAgentsStubIDE) LocalSkillsDir() string          { return "" }

func (s subAgentsStubIDE) real() system.IDEAdapter { return system.GetAdapterByAgentID(s.agentID) }

func (s subAgentsStubIDE) SystemPromptStrategy() model.SystemPromptStrategy {
	return s.real().SystemPromptStrategy()
}
func (s subAgentsStubIDE) MCPStrategy() model.MCPStrategy { return s.real().MCPStrategy() }
func (s subAgentsStubIDE) SystemPromptFile(homeDir string) string {
	return s.real().SystemPromptFile(homeDir)
}
func (s subAgentsStubIDE) SubAgentsDir(homeDir string) string    { return s.real().SubAgentsDir(homeDir) }
func (s subAgentsStubIDE) EmbeddedSubAgentsDir() string          { return s.real().EmbeddedSubAgentsDir() }
func (s subAgentsStubIDE) CommandsDir(homeDir string) string     { return s.real().CommandsDir(homeDir) }
func (s subAgentsStubIDE) EmbeddedCommandsDir() string           { return s.real().EmbeddedCommandsDir() }
func (s subAgentsStubIDE) SkillsDir(homeDir string) string       { return s.real().SkillsDir(homeDir) }
func (s subAgentsStubIDE) SettingsPath(homeDir string) string    { return s.real().SettingsPath(homeDir) }
func (s subAgentsStubIDE) MCPConfigPath(homeDir, serverName string) string {
	return s.real().MCPConfigPath(homeDir, serverName)
}
func (s subAgentsStubIDE) SupportsSubAgents() bool    { return s.real().SupportsSubAgents() }
func (s subAgentsStubIDE) SupportsSlashCommands() bool { return s.real().SupportsSlashCommands() }
func (s subAgentsStubIDE) SupportsSkills() bool        { return s.real().SupportsSkills() }
func (s subAgentsStubIDE) SupportsSystemPrompt() bool  { return s.real().SupportsSystemPrompt() }
func (s subAgentsStubIDE) SupportsMCP() bool           { return s.real().SupportsMCP() }

var _ system.IDEAdapter = subAgentsStubIDE{}

func newSubAgentStep(homeDir string, ids ...model.AgentID) *StepInjectSubAgents {
	var ides []system.IDEAdapter
	for _, id := range ids {
		ides = append(ides, subAgentsStubIDE{agentID: id})
	}
	step := NewStepInjectSubAgents(&InstallContext{HomeDir: homeDir, IDEs: ides})
	step.fs = afero.NewMemMapFs()
	return step
}

// TestInjectSubAgents_ClaudeCode_WritesAgentFiles verifies that ClaudeCode sub-agent
// .md files are written to ~/.claude/agents/.
func TestInjectSubAgents_ClaudeCode_WritesAgentFiles(t *testing.T) {
	step := newSubAgentStep("/home/user", model.AgentClaudeCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify at least one agent file was written
	agentDir := "/home/user/.claude/agents"
	infos, err := afero.ReadDir(step.fs, agentDir)
	if err != nil {
		t.Fatalf("expected ~/.claude/agents to exist, got: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("expected at least one file in ~/.claude/agents, got none")
	}
	// Verify sdd-apply.md is present
	path := filepath.Join(agentDir, "sdd-apply.md")
	if _, err := step.fs.Stat(path); err != nil {
		t.Errorf("expected %q to exist: %v", path, err)
	}
}

// TestInjectSubAgents_ClaudeCode_CreatesDestDir verifies the destination directory
// is created when it doesn't exist.
func TestInjectSubAgents_ClaudeCode_CreatesDestDir(t *testing.T) {
	step := newSubAgentStep("/home/newuser", model.AgentClaudeCode)
	// Directory does not exist — Run should create it
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	agentDir := "/home/newuser/.claude/agents"
	if _, err := step.fs.Stat(agentDir); err != nil {
		t.Errorf("expected directory %q to be created: %v", agentDir, err)
	}
}

// TestInjectSubAgents_Kiro_WritesAgentFiles verifies Kiro sub-agent files are written
// to ~/.kiro/agents/.
func TestInjectSubAgents_Kiro_WritesAgentFiles(t *testing.T) {
	step := newSubAgentStep("/home/user", model.AgentKiroIDE)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	agentDir := "/home/user/.kiro/agents"
	infos, err := afero.ReadDir(step.fs, agentDir)
	if err != nil {
		t.Fatalf("expected ~/.kiro/agents to exist, got: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("expected at least one file in ~/.kiro/agents, got none")
	}
}

// TestInjectSubAgents_NoOp_UnsupportedAgent verifies no files are written for
// agents that have no sub-agent assets (e.g. AgentCursor).
func TestInjectSubAgents_NoOp_UnsupportedAgent(t *testing.T) {
	step := newSubAgentStep("/home/user", model.AgentCursor)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// No agent directories should have been created
	if _, err := step.fs.Stat("/home/user/.claude"); err == nil {
		t.Error("expected no ~/.claude directory for unsupported agent, but it exists")
	}
	if _, err := step.fs.Stat("/home/user/.kiro"); err == nil {
		t.Error("expected no ~/.kiro directory for unsupported agent, but it exists")
	}
}
