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
// New interface methods delegate to the real adapter so tests use real paths/strategies.
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

func (m mcpStubIDE) real() system.IDEAdapter { return system.GetAdapterByAgentID(m.agentID) }

func (m mcpStubIDE) SystemPromptStrategy() model.SystemPromptStrategy {
	return m.real().SystemPromptStrategy()
}
func (m mcpStubIDE) MCPStrategy() model.MCPStrategy { return m.real().MCPStrategy() }
func (m mcpStubIDE) SystemPromptFile(homeDir string) string {
	return m.real().SystemPromptFile(homeDir)
}
func (m mcpStubIDE) SubAgentsDir(homeDir string) string    { return m.real().SubAgentsDir(homeDir) }
func (m mcpStubIDE) EmbeddedSubAgentsDir() string          { return m.real().EmbeddedSubAgentsDir() }
func (m mcpStubIDE) CommandsDir(homeDir string) string     { return m.real().CommandsDir(homeDir) }
func (m mcpStubIDE) EmbeddedCommandsDir() string           { return m.real().EmbeddedCommandsDir() }
func (m mcpStubIDE) SkillsDir(homeDir string) string       { return m.real().SkillsDir(homeDir) }
func (m mcpStubIDE) SettingsPath(homeDir string) string    { return m.real().SettingsPath(homeDir) }
func (m mcpStubIDE) MCPConfigPath(homeDir, serverName string) string {
	return m.real().MCPConfigPath(homeDir, serverName)
}
func (m mcpStubIDE) SupportsSubAgents() bool    { return m.real().SupportsSubAgents() }
func (m mcpStubIDE) SupportsSlashCommands() bool { return m.real().SupportsSlashCommands() }
func (m mcpStubIDE) SupportsSkills() bool        { return m.real().SupportsSkills() }
func (m mcpStubIDE) SupportsSystemPrompt() bool  { return m.real().SupportsSystemPrompt() }
func (m mcpStubIDE) SupportsMCP() bool           { return m.real().SupportsMCP() }

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
	if !ok || len(args) < 2 || args[0] != "mcp" || args[1] != "--tools=agent" {
		t.Errorf("args = %v, want [\"mcp\", \"--tools=agent\"]", got["args"])
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
	if !ok || len(args) < 2 || args[0] != "mcp" || args[1] != "--tools=agent" {
		t.Errorf("args = %v, want [\"mcp\", \"--tools=agent\"]", sddMem["args"])
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
	if !strings.Contains(content, `args = ["mcp", "--tools=agent"]`) {
		t.Errorf("missing args line, got:\n%s", content)
	}
	if !strings.Contains(content, `model_instructions_file = "sdd-memory-instructions.md"`) {
		t.Errorf("missing model_instructions_file, got:\n%s", content)
	}
	if !strings.Contains(content, `experimental_compact_prompt_file = "sdd-memory-compact-prompt.md"`) {
		t.Errorf("missing experimental_compact_prompt_file, got:\n%s", content)
	}
}

// TestInjectMCP_Codex_Upsert verifies an existing TOML block is replaced, not duplicated.
func TestInjectMCP_Codex_Upsert(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentCodex)

	tomlPath := "/home/user/.codex/config.toml"
	if err := step.fs.MkdirAll(filepath.Dir(tomlPath), 0755); err != nil {
		t.Fatal(err)
	}
	existing := "[mcp_servers.sdd-memory]\ncommand = \"old-cmd\"\nargs = [\"mcp\", \"--tools=agent\"]\n"
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

	// Path depends on OS — compute the same way the step does.
	vscodePath := filepath.Join(vscodeUserConfigDir("/home/user"), "mcp.json")
	got := readJSON(t, step.fs, vscodePath)

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

// TestSddMemoryArgs_ContainsToolsAgent verifies the package-level variable includes --tools=agent.
func TestSddMemoryArgs_ContainsToolsAgent(t *testing.T) {
	if len(sddMemoryArgs) < 2 {
		t.Fatalf("sddMemoryArgs len = %d, want at least 2", len(sddMemoryArgs))
	}
	if sddMemoryArgs[0] != "mcp" {
		t.Errorf("sddMemoryArgs[0] = %q, want \"mcp\"", sddMemoryArgs[0])
	}
	if sddMemoryArgs[1] != "--tools=agent" {
		t.Errorf("sddMemoryArgs[1] = %q, want \"--tools=agent\"", sddMemoryArgs[1])
	}
}

// TestInjectMCP_ClaudeCode_ArgsContainToolsAgent verifies --tools=agent is in Claude Code config.
func TestInjectMCP_ClaudeCode_ArgsContainToolsAgent(t *testing.T) {
	step := newMCPStep("/home/user", model.AgentClaudeCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	path := "/home/user/.claude/mcp/sdd-memory.json"
	got := readJSON(t, step.fs, path)

	args, ok := got["args"].([]any)
	if !ok || len(args) < 2 {
		t.Fatalf("args = %v, want at least 2 elements", got["args"])
	}
	if args[1] != "--tools=agent" {
		t.Errorf("args[1] = %q, want \"--tools=agent\"", args[1])
	}
}

// TestVscodeUserConfigDir_Windows verifies the Windows path uses APPDATA.
func TestVscodeUserConfigDir_Windows(t *testing.T) {
	result := vscodeUserConfigDir_forOS("windows", "/home/user", "C:\\Users\\user\\AppData\\Roaming", "")
	want := "C:\\Users\\user\\AppData\\Roaming/Code/User"
	if result != want {
		t.Errorf("Windows APPDATA path = %q, want %q", result, want)
	}
}

// TestVscodeUserConfigDir_WindowsFallback verifies fallback to ~/.vscode when APPDATA absent.
func TestVscodeUserConfigDir_WindowsFallback(t *testing.T) {
	result := vscodeUserConfigDir_forOS("windows", "/home/user", "", "")
	want := "/home/user/.vscode"
	if result != want {
		t.Errorf("Windows fallback path = %q, want %q", result, want)
	}
}

// TestVscodeUserConfigDir_macOS verifies the macOS path.
func TestVscodeUserConfigDir_macOS(t *testing.T) {
	result := vscodeUserConfigDir_forOS("darwin", "/Users/user", "", "")
	want := "/Users/user/Library/Application Support/Code/User"
	if result != want {
		t.Errorf("macOS path = %q, want %q", result, want)
	}
}

// TestVscodeUserConfigDir_Linux verifies the Linux path uses XDG_CONFIG_HOME when set.
func TestVscodeUserConfigDir_Linux(t *testing.T) {
	result := vscodeUserConfigDir_forOS("linux", "/home/user", "", "/home/user/.config")
	want := "/home/user/.config/Code/User"
	if result != want {
		t.Errorf("Linux XDG path = %q, want %q", result, want)
	}
}

// TestVscodeUserConfigDir_LinuxFallback verifies the Linux fallback when XDG absent.
func TestVscodeUserConfigDir_LinuxFallback(t *testing.T) {
	result := vscodeUserConfigDir_forOS("linux", "/home/user", "", "")
	want := "/home/user/.config/Code/User"
	if result != want {
		t.Errorf("Linux fallback path = %q, want %q", result, want)
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
