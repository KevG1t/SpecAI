package system

import (
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
