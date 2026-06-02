package steps

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

// mcpStubIDE is a minimal IDEAdapter stub for inject_mcp tests.
type mcpStubIDE struct {
	agentID model.AgentID
}

func (m mcpStubIDE) Name() string                    { return string(m.agentID) }
func (m mcpStubIDE) AgentID() model.AgentID          { return m.agentID }
func (m mcpStubIDE) AssetFolder() string             { return "" }
func (m mcpStubIDE) ConfigDir(_ string) string       { return "" }
func (m mcpStubIDE) GlobalRulesDir(_ string) string  { return "" }
func (m mcpStubIDE) GlobalSkillsDir(_ string) string { return "" }
func (m mcpStubIDE) LocalRulesFile() string          { return "" }
func (m mcpStubIDE) LocalSkillsDir() string          { return "" }

var _ system.IDEAdapter = mcpStubIDE{}

func newMCPStep(homeDir string, ids ...model.AgentID) *StepInjectMCP {
	var ides []system.IDEAdapter
	for _, id := range ids {
		ides = append(ides, mcpStubIDE{agentID: id})
	}
	return &StepInjectMCP{
		ctx: &InstallContext{HomeDir: homeDir, IDEs: ides},
		fs:  afero.NewMemMapFs(),
	}
}

func readJSON(t *testing.T, fsys afero.Fs, path string) map[string]any {
	t.Helper()
	data, err := afero.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("readJSON: cannot read %s: %v", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("readJSON: cannot unmarshal %s: %v", path, err)
	}
	return out
}

// TestInjectMCP_ClaudeCode verifies the separate file is written at ~/.claude/mcp/sdd-memory.json.
func TestInjectMCP_ClaudeCode(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentClaudeCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	path := "/home/user/.claude/mcp/sdd-memory.json"
	got := readJSON(t, step.fs, path)

	if got["command"] == nil {
		t.Fatal("expected 'command' key in sdd-memory.json")
	}
	args, ok := got["args"].([]any)
	if !ok || len(args) != 1 || args[0] != "mcp" {
		t.Errorf("args = %v, want [\"mcp\"]", got["args"])
	}
	// No mcpServers wrapper.
	if _, hasWrapper := got["mcpServers"]; hasWrapper {
		t.Error("Claude Code file must not have an mcpServers wrapper")
	}
}

// TestInjectMCP_OpenCode verifies merge into opencode.json preserves existing keys.
func TestInjectMCP_OpenCode(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentOpenCode)

	// Pre-seed an existing settings file with an unrelated key.
	existing := []byte(`{"theme":"dark","mcp":{"other-tool":{"command":"other","type":"local"}}}`)
	settingsPath := "/home/user/.config/opencode/opencode.json"
	if err := step.fs.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(step.fs, settingsPath, existing, 0644); err != nil {
		t.Fatal(err)
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	got := readJSON(t, step.fs, settingsPath)

	// Existing key must be preserved.
	if got["theme"] != "dark" {
		t.Errorf("existing 'theme' key lost, got %v", got["theme"])
	}

	mcp, ok := got["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("'mcp' key missing or wrong type: %T", got["mcp"])
	}

	// Existing MCP tool preserved.
	if _, exists := mcp["other-tool"]; !exists {
		t.Error("existing 'other-tool' MCP entry was removed")
	}

	// sdd-memory injected.
	sddMem, ok := mcp["sdd-memory"].(map[string]any)
	if !ok {
		t.Fatalf("'mcp.sdd-memory' missing or wrong type")
	}
	if sddMem["type"] != "local" {
		t.Errorf("type = %v, want 'local'", sddMem["type"])
	}
	cmdArr, ok := sddMem["command"].([]any)
	if !ok || len(cmdArr) < 2 || cmdArr[1] != "mcp" {
		t.Errorf("command = %v, want [\"<cmd>\", \"mcp\"]", sddMem["command"])
	}
}

// TestInjectMCP_Cursor verifies mcp.json is written with an mcpServers wrapper.
func TestInjectMCP_Cursor(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentCursor)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	path := "/home/user/.cursor/mcp.json"
	got := readJSON(t, step.fs, path)

	servers, ok := got["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("'mcpServers' missing or wrong type")
	}
	sddMem, ok := servers["sdd-memory"].(map[string]any)
	if !ok {
		t.Fatalf("'mcpServers.sdd-memory' missing")
	}
	args, ok := sddMem["args"].([]any)
	if !ok || len(args) != 1 || args[0] != "mcp" {
		t.Errorf("args = %v, want [\"mcp\"]", sddMem["args"])
	}
}

// TestInjectMCP_Codex verifies the TOML block is appended when no prior block exists.
func TestInjectMCP_Codex(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentCodex)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	path := "/home/user/.codex/config.toml"
	data, err := afero.ReadFile(step.fs, path)
	if err != nil {
		t.Fatalf("cannot read config.toml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "[mcp_servers.sdd-memory]") {
		t.Errorf("missing [mcp_servers.sdd-memory] block, got:\n%s", content)
	}
	if !strings.Contains(content, `args = ["mcp"]`) {
		t.Errorf("missing args line, got:\n%s", content)
	}
}

// TestInjectMCP_Codex_Upsert verifies an existing TOML block is replaced, not duplicated.
func TestInjectMCP_Codex_Upsert(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentCodex)

	tomlPath := "/home/user/.codex/config.toml"
	if err := step.fs.MkdirAll(filepath.Dir(tomlPath), 0755); err != nil {
		t.Fatal(err)
	}
	existing := "[mcp_servers.sdd-memory]\ncommand = \"old-cmd\"\nargs = [\"mcp\"]\n"
	if err := afero.WriteFile(step.fs, tomlPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	data, err := afero.ReadFile(step.fs, tomlPath)
	if err != nil {
		t.Fatalf("cannot read config.toml: %v", err)
	}
	content := string(data)

	count := strings.Count(content, "[mcp_servers.sdd-memory]")
	if count != 1 {
		t.Errorf("expected exactly 1 [mcp_servers.sdd-memory] block, got %d:\n%s", count, content)
	}
	if strings.Contains(content, "old-cmd") {
		t.Errorf("old command not replaced:\n%s", content)
	}
}

// TestInjectMCP_BinaryNotOnPATH verifies fallback to literal "sdd-memory" when binary is missing.
func TestInjectMCP_BinaryNotOnPATH(t *testing.T) {
	// resolveBinary is called with a name that definitely won't be on PATH.
	resolved := resolveBinary("sdd-memory-definitely-not-installed-xyz-12345")
	if resolved != "sdd-memory-definitely-not-installed-xyz-12345" {
		t.Errorf("expected literal fallback, got %q", resolved)
	}
}

// TestInjectMCP_VSCode_WritesServersKey verifies that VS Code Copilot config uses
// "servers" (not "mcpServers") per the VS Code MCP specification.
func TestInjectMCP_VSCode_WritesServersKey(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentVSCodeCopilot)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	path := "/home/user/.vscode/mcp.json"
	got := readJSON(t, step.fs, path)

	if _, hasMCPServers := got["mcpServers"]; hasMCPServers {
		t.Error("VS Code config must NOT have 'mcpServers' key — use 'servers' instead")
	}
	servers, ok := got["servers"].(map[string]any)
	if !ok {
		t.Fatalf("VS Code config must have 'servers' key, got keys: %v", mapKeysOf(got))
	}
	if _, ok := servers["sdd-memory"]; !ok {
		t.Fatal("'servers.sdd-memory' must be present")
	}
}

// TestInjectMCP_ClaudeCode_WritesHooks verifies that the Claude Code install writes
// a UserPromptSubmit hook to ~/.claude/settings.json.
func TestInjectMCP_ClaudeCode_WritesHooks(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentClaudeCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	settingsPath := "/home/user/.claude/settings.json"
	got := readJSON(t, step.fs, settingsPath)

	hooks, ok := got["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("'hooks' key missing in settings.json, got keys: %v", mapKeysOf(got))
	}
	if _, ok := hooks["UserPromptSubmit"]; !ok {
		t.Fatal("'hooks.UserPromptSubmit' must be present after Claude Code install")
	}
}

// TestInjectMCP_Antigravity_WritesPluginFiles verifies that Antigravity install
// writes plugin.json and hooks.json in addition to mcp_config.json.
func TestInjectMCP_Antigravity_WritesPluginFiles(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentAntigravity)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	pluginDir := "/home/user/.gemini/antigravity-cli"
	for _, file := range []string{"mcp_config.json", "plugin.json", "hooks.json"} {
		path := pluginDir + "/" + file
		if _, err := step.fs.Stat(path); err != nil {
			t.Errorf("Antigravity plugin file %q must exist after install: %v", file, err)
		}
	}
}

func mapKeysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestInjectMCP_AllStrategies is a smoke test verifying every supported agent runs without error.
func TestInjectMCP_AllStrategies(t *testing.T) {
	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentKilocode,
		model.AgentGeminiCLI,
		model.AgentCursor,
		model.AgentWindsurf,
		model.AgentKiroIDE,
		model.AgentAntigravity,
		model.AgentVSCodeCopilot,
		model.AgentKimi,
		model.AgentQwenCode,
		model.AgentOpenClaw,
		model.AgentPi,
		model.AgentTrae,
		model.AgentCodex,
	}
	for _, id := range agents {
		t.Run(fmt.Sprintf("agent=%s", id), func(t *testing.T) {
			step := newMCPStep("/home/user", id)
			if err := step.Run(); err != nil {
				t.Errorf("Run() error for %s: %v", id, err)
			}
		})
	}
}
