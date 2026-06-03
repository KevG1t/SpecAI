package mcp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
)

// testAdapter is a minimal agents.Adapter implementation for mcp injection tests.
type testAdapter struct {
	agent       model.AgentID
	mcpStrategy model.MCPStrategy
	settingsPath string
	mcpConfigPath string
}

func (a testAdapter) Agent() model.AgentID { return a.agent }
func (a testAdapter) Tier() model.SupportTier { return model.TierFull }
func (a testAdapter) Detect(_ context.Context, _ string) (bool, string, string, bool, error) {
	return false, "", "", false, nil
}
func (a testAdapter) SupportsAutoInstall() bool { return false }
func (a testAdapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) { return nil, nil }
func (a testAdapter) GlobalConfigDir(_ string) string { return "" }
func (a testAdapter) SystemPromptDir(_ string) string { return "" }
func (a testAdapter) SystemPromptFile(_ string) string { return "" }
func (a testAdapter) SkillsDir(_ string) string { return "" }
func (a testAdapter) SettingsPath(_ string) string { return a.settingsPath }
func (a testAdapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}
func (a testAdapter) MCPStrategy() model.MCPStrategy { return a.mcpStrategy }
func (a testAdapter) MCPConfigPath(_ string, serverName string) string {
	if a.mcpConfigPath != "" {
		return filepath.Join(a.mcpConfigPath, serverName+".json")
	}
	return ""
}
func (a testAdapter) SupportsOutputStyles() bool { return false }
func (a testAdapter) OutputStyleDir(_ string) string { return "" }
func (a testAdapter) SupportsSlashCommands() bool { return false }
func (a testAdapter) CommandsDir(_ string) string { return "" }
func (a testAdapter) SupportsSubAgents() bool { return false }
func (a testAdapter) SubAgentsDir(_ string) string { return "" }
func (a testAdapter) EmbeddedSubAgentsDir() string { return "" }
func (a testAdapter) SupportsSkills() bool { return false }
func (a testAdapter) SupportsSystemPrompt() bool { return false }
func (a testAdapter) SupportsMCP() bool { return true }

// TestInjectNotionSeparateFile verifies that InjectNotion writes notion.json
// for agents using the StrategySeparateMCPFiles approach.
func TestInjectNotionSeparateFile(t *testing.T) {
	tmpDir := t.TempDir()
	mcpDir := filepath.Join(tmpDir, "mcp")

	adapter := testAdapter{
		agent:       model.AgentClaudeCode,
		mcpStrategy: model.StrategySeparateMCPFiles,
		mcpConfigPath: mcpDir,
	}

	result, guidance, err := InjectNotion(tmpDir, adapter)
	if err != nil {
		t.Fatalf("InjectNotion returned error: %v", err)
	}
	if !result.Changed {
		t.Fatal("expected Changed=true after first Notion injection")
	}
	if len(result.Files) == 0 {
		t.Fatal("expected at least one file path in result.Files")
	}
	if !strings.Contains(result.Files[0], "notion") {
		t.Fatalf("expected notion in file path, got %q", result.Files[0])
	}
	if guidance == "" {
		t.Fatal("expected non-empty auth guidance for Notion")
	}
	// Guidance must include config path and docs URL.
	if !strings.Contains(guidance, result.Files[0]) {
		t.Fatalf("auth guidance should include the config file path: %q", guidance)
	}
	if !strings.Contains(guidance, "https://developers.notion.com") {
		t.Fatalf("auth guidance should include Notion docs URL, got: %q", guidance)
	}
}

// TestInjectNotionMCPConfigFile verifies notion overlay is written for
// agents using the StrategyMCPConfigFile approach (e.g., Cursor-like agents).
func TestInjectNotionMCPConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	mcpDir := filepath.Join(tmpDir, "mcp")

	adapter := testAdapter{
		agent:       model.AgentCursor,
		mcpStrategy: model.StrategyMCPConfigFile,
		mcpConfigPath: mcpDir,
	}

	result, guidance, err := InjectNotion(tmpDir, adapter)
	if err != nil {
		t.Fatalf("InjectNotion (MCPConfigFile) returned error: %v", err)
	}
	_ = result
	_ = guidance
}

// TestInjectJiraSeparateFile verifies InjectJira writes jira.json for
// StrategySeparateMCPFiles agents.
func TestInjectJiraSeparateFile(t *testing.T) {
	tmpDir := t.TempDir()
	mcpDir := filepath.Join(tmpDir, "mcp")

	adapter := testAdapter{
		agent:       model.AgentClaudeCode,
		mcpStrategy: model.StrategySeparateMCPFiles,
		mcpConfigPath: mcpDir,
	}

	result, guidance, err := InjectJira(tmpDir, adapter)
	if err != nil {
		t.Fatalf("InjectJira returned error: %v", err)
	}
	if !result.Changed {
		t.Fatal("expected Changed=true after first Jira injection")
	}
	if len(result.Files) == 0 {
		t.Fatal("expected at least one file path in result.Files")
	}
	if !strings.Contains(result.Files[0], "jira") {
		t.Fatalf("expected jira in file path, got %q", result.Files[0])
	}
	if guidance == "" {
		t.Fatal("expected non-empty auth guidance for Jira")
	}
	if !strings.Contains(guidance, "https://github.com/sooperset/mcp-atlassian") {
		t.Fatalf("auth guidance should include Jira docs URL, got: %q", guidance)
	}
}

// TestNotionOverlayForAgentReturnsCorrectFormat verifies that different agents get
// the correct Notion overlay format.
func TestNotionOverlayForAgentReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		agent    model.AgentID
		wantKey  string
	}{
		{model.AgentOpenCode, "\"type\""},
		{model.AgentOpenClaw, "\"servers\""},
		{model.AgentVSCodeCopilot, "\"type\""},
		{model.AgentClaudeCode, "\"mcpServers\""},
		{model.AgentKimi, "\"mcpServers\""},
	}
	for _, tc := range tests {
		t.Run(string(tc.agent), func(t *testing.T) {
			adapter := testAdapter{agent: tc.agent}
			overlay := notionOverlayForAgent(adapter)
			if overlay == nil {
				t.Fatalf("notionOverlayForAgent(%q) returned nil", tc.agent)
			}
			if !strings.Contains(string(overlay), tc.wantKey) {
				t.Fatalf("overlay for %q should contain %q:\n%s", tc.agent, tc.wantKey, string(overlay))
			}
		})
	}
}

// TestInjectJiraGuidanceContainsConfigPath verifies that auth guidance includes the
// config file path.
func TestInjectJiraGuidanceContainsConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	mcpDir := filepath.Join(tmpDir, "mcp")

	adapter := testAdapter{
		agent:       model.AgentClaudeCode,
		mcpStrategy: model.StrategySeparateMCPFiles,
		mcpConfigPath: mcpDir,
	}

	result, guidance, err := InjectJira(tmpDir, adapter)
	if err != nil {
		t.Fatalf("InjectJira returned error: %v", err)
	}
	if len(result.Files) == 0 {
		t.Fatal("expected file path in result")
	}
	if !strings.Contains(guidance, result.Files[0]) {
		t.Fatalf("guidance should contain the config file path %q, got: %q", result.Files[0], guidance)
	}
}

// TestUvxAvailableFnIsInjectable verifies the uvxAvailableFn var can be overridden.
func TestUvxAvailableFnIsInjectable(t *testing.T) {
	original := uvxAvailableFn
	defer func() { uvxAvailableFn = original }()

	uvxAvailableFn = func() bool { return true }
	if !UvxAvailable() {
		t.Fatal("expected UvxAvailable()=true when uvxAvailableFn returns true")
	}

	uvxAvailableFn = func() bool { return false }
	if UvxAvailable() {
		t.Fatal("expected UvxAvailable()=false when uvxAvailableFn returns false")
	}
}
