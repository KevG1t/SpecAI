package system

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

// adapterExpectation holds expected values for a single adapter in the table test.
type adapterExpectation struct {
	name        string
	agentID     model.AgentID
	assetFolder string
}

var expectedAdapters = []adapterExpectation{
	{name: "Cursor", agentID: model.AgentCursor, assetFolder: "cursor"},
	{name: "Windsurf", agentID: model.AgentWindsurf, assetFolder: "windsurf"},
	{name: "Codex", agentID: model.AgentCodex, assetFolder: "codex"},
	{name: "Claude Code", agentID: model.AgentClaudeCode, assetFolder: "claude"},
	{name: "Gemini CLI", agentID: model.AgentGeminiCLI, assetFolder: "gemini"},
	{name: "Antigravity", agentID: model.AgentAntigravity, assetFolder: "antigravity"},
	{name: "OpenCode", agentID: model.AgentOpenCode, assetFolder: "opencode"},
	{name: "Kiro IDE", agentID: model.AgentKiroIDE, assetFolder: "kiro"},
	{name: "Kimi Code", agentID: model.AgentKimi, assetFolder: "kimi"},
	{name: "Qwen Code", agentID: model.AgentQwenCode, assetFolder: "qwen"},
	{name: "Kilo Code", agentID: model.AgentKilocode, assetFolder: "kilo"},
	{name: "OpenClaw", agentID: model.AgentOpenClaw, assetFolder: "generic"},
	{name: "Pi", agentID: model.AgentPi, assetFolder: "generic"},
	{name: "Trae IDE", agentID: model.AgentTrae, assetFolder: "generic"},
	{name: "VS Code Copilot", agentID: model.AgentVSCodeCopilot, assetFolder: "generic"},
}

func TestGetAdapters_Returns15(t *testing.T) {
	adapters := GetAdapters()
	if len(adapters) != 15 {
		t.Fatalf("GetAdapters() returned %d adapters, want 15", len(adapters))
	}
}

func TestIDEAdapter_AgentID(t *testing.T) {
	adapters := GetAdapters()
	byName := make(map[string]IDEAdapter)
	for _, a := range adapters {
		byName[a.Name()] = a
	}

	for _, tc := range expectedAdapters {
		tc := tc
		t.Run(tc.name+"_AgentID", func(t *testing.T) {
			adapter, ok := byName[tc.name]
			if !ok {
				t.Fatalf("no adapter with name %q in GetAdapters()", tc.name)
			}
			got := adapter.AgentID()
			if got != tc.agentID {
				t.Errorf("AgentID() = %q, want %q", got, tc.agentID)
			}
		})
	}
}

func TestIDEAdapter_AssetFolder(t *testing.T) {
	adapters := GetAdapters()
	byName := make(map[string]IDEAdapter)
	for _, a := range adapters {
		byName[a.Name()] = a
	}

	for _, tc := range expectedAdapters {
		tc := tc
		t.Run(tc.name+"_AssetFolder", func(t *testing.T) {
			adapter, ok := byName[tc.name]
			if !ok {
				t.Fatalf("no adapter with name %q in GetAdapters()", tc.name)
			}
			got := adapter.AssetFolder()
			if got != tc.assetFolder {
				t.Errorf("AssetFolder() = %q, want %q", got, tc.assetFolder)
			}
		})
	}
}

func TestGenericAdapters_ReturnGenericAssetFolder(t *testing.T) {
	genericNames := []string{"OpenClaw", "Pi", "Trae IDE", "VS Code Copilot"}
	adapters := GetAdapters()
	byName := make(map[string]IDEAdapter)
	for _, a := range adapters {
		byName[a.Name()] = a
	}

	for _, name := range genericNames {
		adapter, ok := byName[name]
		if !ok {
			t.Fatalf("no adapter with name %q in GetAdapters()", name)
		}
		got := adapter.AssetFolder()
		if got != "generic" {
			t.Errorf("%s AssetFolder() = %q, want \"generic\"", name, got)
		}
	}
}

// capabilityRow is a capability table row — ground truth for all 15 adapters.
type capabilityRow struct {
	agentID              model.AgentID
	sysPromptStrategy    model.SystemPromptStrategy
	mcpStrategy          model.MCPStrategy
	supportsSubAgents    bool
	supportsSlashCmds    bool
	supportsSkills       bool
	supportsSystemPrompt bool
	supportsMCP          bool
}

var capabilityTable = []capabilityRow{
	{model.AgentClaudeCode, model.StrategyMarkdownSections, model.StrategySeparateMCPFiles, true, true, true, true, true},
	{model.AgentOpenCode, model.StrategyFileReplace, model.StrategyMergeIntoSettings, false, true, true, true, true},
	{model.AgentKilocode, model.StrategyFileReplace, model.StrategyMergeIntoSettings, false, false, true, true, true},
	{model.AgentGeminiCLI, model.StrategyFileReplace, model.StrategyMergeIntoSettings, false, false, true, true, true},
	{model.AgentCursor, model.StrategyFileReplace, model.StrategyMCPConfigFile, true, false, true, true, true},
	{model.AgentVSCodeCopilot, model.StrategyInstructionsFile, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentCodex, model.StrategyAppendToFile, model.StrategyTOMLFile, false, false, true, true, true},
	{model.AgentWindsurf, model.StrategyAppendToFile, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentAntigravity, model.StrategyFileReplace, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentKimi, model.StrategyJinjaModules, model.StrategyMCPConfigFile, true, false, true, true, true},
	{model.AgentQwenCode, model.StrategyAppendToFile, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentKiroIDE, model.StrategySteeringFile, model.StrategyMCPConfigFile, true, false, true, true, true},
	{model.AgentOpenClaw, model.StrategyFileReplace, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentPi, model.StrategyInstructionsFile, model.StrategyMCPConfigFile, false, false, true, true, true},
	{model.AgentTrae, model.StrategyAppendToFile, model.StrategyMCPConfigFile, false, false, true, true, true},
}

// TestIDEAdapter_CapabilityTable verifies all 15 adapters against the design's capability table.
func TestIDEAdapter_CapabilityTable(t *testing.T) {
	homeDir := t.TempDir()

	for _, want := range capabilityTable {
		want := want
		t.Run(string(want.agentID), func(t *testing.T) {
			adapter := GetAdapterByAgentID(want.agentID)
			if adapter == nil {
				t.Fatalf("GetAdapterByAgentID(%q) returned nil", want.agentID)
			}

			// Strategy methods
			if got := adapter.SystemPromptStrategy(); got != want.sysPromptStrategy {
				t.Errorf("SystemPromptStrategy() = %v, want %v", got, want.sysPromptStrategy)
			}
			if got := adapter.MCPStrategy(); got != want.mcpStrategy {
				t.Errorf("MCPStrategy() = %v, want %v", got, want.mcpStrategy)
			}

			// Capability flags
			if got := adapter.SupportsMCP(); got != want.supportsMCP {
				t.Errorf("SupportsMCP() = %v, want %v", got, want.supportsMCP)
			}
			if got := adapter.SupportsSubAgents(); got != want.supportsSubAgents {
				t.Errorf("SupportsSubAgents() = %v, want %v", got, want.supportsSubAgents)
			}
			if got := adapter.SupportsSlashCommands(); got != want.supportsSlashCmds {
				t.Errorf("SupportsSlashCommands() = %v, want %v", got, want.supportsSlashCmds)
			}
			if got := adapter.SupportsSkills(); got != want.supportsSkills {
				t.Errorf("SupportsSkills() = %v, want %v", got, want.supportsSkills)
			}
			if got := adapter.SupportsSystemPrompt(); got != want.supportsSystemPrompt {
				t.Errorf("SupportsSystemPrompt() = %v, want %v", got, want.supportsSystemPrompt)
			}

			// SubAgentsDir non-empty iff SupportsSubAgents
			subDir := adapter.SubAgentsDir(homeDir)
			if want.supportsSubAgents && subDir == "" {
				t.Error("SubAgentsDir() must be non-empty when SupportsSubAgents() is true")
			}
			if !want.supportsSubAgents && subDir != "" {
				t.Errorf("SubAgentsDir() = %q, want empty when SupportsSubAgents() is false", subDir)
			}

			// CommandsDir non-empty iff SupportsSlashCommands
			cmdDir := adapter.CommandsDir(homeDir)
			if want.supportsSlashCmds && cmdDir == "" {
				t.Error("CommandsDir() must be non-empty when SupportsSlashCommands() is true")
			}
			if !want.supportsSlashCmds && cmdDir != "" {
				t.Errorf("CommandsDir() = %q, want empty when SupportsSlashCommands() is false", cmdDir)
			}

			// MCPConfigPath non-empty except for MergeIntoSettings agents
			mcpPath := adapter.MCPConfigPath(homeDir, "sdd-memory")
			if want.mcpStrategy != model.StrategyMergeIntoSettings && mcpPath == "" {
				t.Errorf("MCPConfigPath() must be non-empty for strategy %v", want.mcpStrategy)
			}

			// SettingsPath non-empty for MergeIntoSettings agents
			settingsPath := adapter.SettingsPath(homeDir)
			if want.mcpStrategy == model.StrategyMergeIntoSettings && settingsPath == "" {
				t.Error("SettingsPath() must be non-empty for StrategyMergeIntoSettings agents")
			}

			// SkillsDir always non-empty
			if skillsDir := adapter.SkillsDir(homeDir); skillsDir == "" {
				t.Error("SkillsDir() must be non-empty")
			}

			// SystemPromptFile always non-empty
			if sysFile := adapter.SystemPromptFile(homeDir); sysFile == "" {
				t.Error("SystemPromptFile() must be non-empty")
			}

			// EmbeddedSubAgentsDir non-empty iff SupportsSubAgents
			embSub := adapter.EmbeddedSubAgentsDir()
			if want.supportsSubAgents && embSub == "" {
				t.Error("EmbeddedSubAgentsDir() must be non-empty when SupportsSubAgents() is true")
			}
			if !want.supportsSubAgents && embSub != "" {
				t.Errorf("EmbeddedSubAgentsDir() = %q, want empty when SupportsSubAgents() is false", embSub)
			}

			// EmbeddedCommandsDir non-empty iff SupportsSlashCommands
			embCmd := adapter.EmbeddedCommandsDir()
			if want.supportsSlashCmds && embCmd == "" {
				t.Error("EmbeddedCommandsDir() must be non-empty when SupportsSlashCommands() is true")
			}
			if !want.supportsSlashCmds && embCmd != "" {
				t.Errorf("EmbeddedCommandsDir() = %q, want empty when SupportsSlashCommands() is false", embCmd)
			}
		})
	}
}

// TestIDEAdapter_MCPConfigPath_ClaudeCode verifies Claude's MCPConfigPath uses serverName.
func TestIDEAdapter_MCPConfigPath_ClaudeCode(t *testing.T) {
	homeDir := t.TempDir()
	adapter := GetAdapterByAgentID(model.AgentClaudeCode)
	got := adapter.MCPConfigPath(homeDir, "sdd-memory")
	want := filepath.Join(homeDir, ".claude", "mcp", "sdd-memory.json")
	if got != want {
		t.Errorf("MCPConfigPath(homeDir, \"sdd-memory\") = %q, want %q", got, want)
	}
}

// TestIDEAdapter_AllPathsContainHomeDir verifies all path methods include the homeDir prefix.
// VSCode Copilot's MCPConfigPath is OS-specific (uses APPDATA/XDG) and is exempt from this check.
func TestIDEAdapter_AllPathsContainHomeDir(t *testing.T) {
	for _, adapter := range GetAdapters() {
		adapter := adapter
		t.Run(string(adapter.AgentID()), func(t *testing.T) {
			homeDir := t.TempDir()
			slashHome := filepath.ToSlash(homeDir)

			checkPath := func(name, path string) {
				if path == "" {
					return // empty is valid for optional paths
				}
				if !strings.HasPrefix(filepath.ToSlash(path), slashHome) {
					t.Errorf("%s = %q does not contain homeDir %q", name, path, homeDir)
				}
			}

			checkPath("ConfigDir", adapter.ConfigDir(homeDir))
			checkPath("SkillsDir", adapter.SkillsDir(homeDir))
			checkPath("SystemPromptFile", adapter.SystemPromptFile(homeDir))
			checkPath("SettingsPath", adapter.SettingsPath(homeDir))
			// VSCode Copilot MCPConfigPath uses OS config dir (APPDATA/XDG), not homeDir
			if adapter.AgentID() != model.AgentVSCodeCopilot {
				checkPath("MCPConfigPath", adapter.MCPConfigPath(homeDir, "sdd-memory"))
			}
			checkPath("SubAgentsDir", adapter.SubAgentsDir(homeDir))
			checkPath("CommandsDir", adapter.CommandsDir(homeDir))
		})
	}
}
