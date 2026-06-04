package sddmemory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/agents"
	"github.com/KevG1t/specai/internal/agents/antigravity"
	"github.com/KevG1t/specai/internal/agents/claude"
	"github.com/KevG1t/specai/internal/agents/codex"
	"github.com/KevG1t/specai/internal/agents/gemini"
	"github.com/KevG1t/specai/internal/agents/openclaw"
	"github.com/KevG1t/specai/internal/agents/opencode"
	"github.com/KevG1t/specai/internal/agents/pi"
	"github.com/KevG1t/specai/internal/agents/qwen"
	"github.com/KevG1t/specai/internal/agents/vscode"
)

func claudeAdapter() agents.Adapter   { return claude.NewAdapter() }
func opencodeAdapter() agents.Adapter { return opencode.NewAdapter() }
func codexAdapter() agents.Adapter    { return codex.NewAdapter() }
func geminiAdapter() agents.Adapter   { return gemini.NewAdapter() }
func qwenAdapter() agents.Adapter     { return qwen.NewAdapter() }
func openclawAdapter() agents.Adapter { return openclaw.NewAdapter() }
func antigravityAdapter() agents.Adapter {
	return antigravity.NewAdapter()
}

func piAdapter() agents.Adapter { return pi.NewAdapter() }

func TestInjectClaudeWritesMCPConfig(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	// Check MCP JSON file was created.
	mcpPath := filepath.Join(home, ".claude", "mcp", "sdd-memory.json")
	mcpContent, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory.json) error = %v", err)
	}

	// Parse the JSON and validate the "command" key exists and references sdd-memory.
	// The command may be an absolute path (if sdd-memory is on PATH) or the relative
	// string "sdd-memory" (if not found). Both are valid.
	var parsed map[string]any
	if err := json.Unmarshal(mcpContent, &parsed); err != nil {
		t.Fatalf("Unmarshal(sdd-memory.json) error = %v", err)
	}
	cmd, ok := parsed["command"].(string)
	if !ok || cmd == "" {
		t.Fatalf("sdd-memory.json missing or empty command field; got: %s", mcpContent)
	}
	// Command must either be the literal "sdd-memory" or an absolute path ending in "sdd-memory".
	base := filepath.Base(cmd)
	if base != "sdd-memory" && base != "sdd-memory.exe" {
		t.Fatalf("sdd-memory.json command %q does not reference sdd-memory binary; got: %s", cmd, mcpContent)
	}
	if _, ok := parsed["args"]; !ok {
		t.Fatal("sdd-memory.json missing args field")
	}
}

func TestInjectClaudeWritesProtocolSection(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	claudeMDPath := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("ReadFile(CLAUDE.md) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "<!-- specai:sdd-memory-protocol -->") {
		t.Fatal("CLAUDE.md missing open marker for sdd-memory-protocol")
	}
	if !strings.Contains(text, "<!-- /specai:sdd-memory-protocol -->") {
		t.Fatal("CLAUDE.md missing close marker for sdd-memory-protocol")
	}
	// Real content check.
	if !strings.Contains(text, "mem_save") {
		t.Fatal("CLAUDE.md missing real sdd-memory protocol content (expected 'mem_save')")
	}
}

func TestInjectClaudeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeMergesSDDMemoryToSettings(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	// Should include opencode.json and AGENTS.md (fallback protocol injection).
	if len(result.Files) != 2 {
		t.Fatalf("Inject() files = %v, want exactly 2 (opencode.json + AGENTS.md)", result.Files)
	}

	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	text := string(config)
	if !strings.Contains(text, `"sdd-memory"`) {
		t.Fatal("opencode.json missing sdd-memory server entry")
	}
	if !strings.Contains(text, `"mcp"`) {
		t.Fatal("opencode.json missing mcp key")
	}
	if strings.Contains(text, `"mcpServers"`) {
		t.Fatal("opencode.json should use 'mcp' key, not 'mcpServers'")
	}
	if !strings.Contains(text, `"type": "local"`) {
		t.Fatal("opencode.json sdd-memory missing type: local")
	}
	// OpenCode 1.3.3+: command must be an array, no separate "args" field.
	if strings.Contains(text, `"args"`) {
		t.Fatal("opencode.json must NOT have a separate args field — command must be an array")
	}

	// Verify NO plugin files or plugin arrays exist.
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "sdd-memory.ts")
	if _, err := os.Stat(pluginPath); err == nil {
		t.Fatal("plugin file should NOT exist — old approach removed")
	}
	if strings.Contains(text, `"plugins"`) {
		t.Fatal("opencode.json should NOT contain plugins key")
	}

	agentsPath := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	agentsText := string(agentsContent)
	if !strings.Contains(agentsText, "<!-- specai:sdd-memory-protocol -->") {
		t.Fatal("AGENTS.md missing sdd-memory protocol section marker")
	}
	if !strings.Contains(agentsText, "mem_save") {
		t.Fatal("AGENTS.md missing sdd-memory protocol content (expected 'mem_save')")
	}
}

func TestInjectOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectPiProvisioningCreatesMissingMCPAdapterFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, piAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	settings := readJSONFile(t, filepath.Join(home, ".pi", "agent", "settings.json"))
	assertNestedStrings(t, settings, []string{"npm:pi-mcp-adapter"}, "packages")

	npmPackage := readJSONFile(t, filepath.Join(home, ".pi", "npm", "package.json"))
	assertNestedString(t, npmPackage, "^2.6.0", "dependencies", "pi-mcp-adapter")
}

func TestInjectPiProvisioningPreservesUnrelatedContent(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".pi", "agent", "settings.json"), `{"theme":"kanagawa","packages":["npm:other@1.0.0"]}`)
	writeFile(t, filepath.Join(home, ".pi", "npm", "package.json"), `{"name":"pi-user","dependencies":{"left-pad":"^1.0.0"},"devDependencies":{"vitest":"^1.0.0"}}`)

	_, err := Inject(home, piAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	settings := readJSONFile(t, filepath.Join(home, ".pi", "agent", "settings.json"))
	assertNestedString(t, settings, "kanagawa", "theme")
	assertNestedStringsUnordered(t, settings, []string{"npm:other@1.0.0", "npm:pi-mcp-adapter"}, "packages")

	npmPackage := readJSONFile(t, filepath.Join(home, ".pi", "npm", "package.json"))
	assertNestedString(t, npmPackage, "pi-user", "name")
	assertNestedString(t, npmPackage, "^1.0.0", "dependencies", "left-pad")
	assertNestedString(t, npmPackage, "^2.6.0", "dependencies", "pi-mcp-adapter")
	assertNestedString(t, npmPackage, "^1.0.0", "devDependencies", "vitest")
}

func TestInjectPiProvisioningCanonicalizesExistingEntriesAndIsIdempotent(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-mcp-adapter@2.0.0"]}`)
	writeFile(t, filepath.Join(home, ".pi", "npm", "package.json"), `{"dependencies":{"pi-mcp-adapter":"^2.0.0"}}`)

	first, err := Inject(home, piAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	settings := readJSONFile(t, filepath.Join(home, ".pi", "agent", "settings.json"))
	assertNestedStrings(t, settings, []string{"npm:pi-mcp-adapter"}, "packages")
	npmPackage := readJSONFile(t, filepath.Join(home, ".pi", "npm", "package.json"))
	assertNestedString(t, npmPackage, "^2.6.0", "dependencies", "pi-mcp-adapter")

	second, err := Inject(home, piAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectPiProvisioningMigratesLegacyObjectPackages(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".pi", "agent", "settings.json"), `{"theme":"kanagawa","packages":{"npm:other":"1.0.0","npm:pi-mcp-adapter":"2.0.0"}}`)

	_, err := Inject(home, piAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	settings := readJSONFile(t, filepath.Join(home, ".pi", "agent", "settings.json"))
	assertNestedString(t, settings, "kanagawa", "theme")
	assertNestedStringsUnordered(t, settings, []string{"npm:other@1.0.0", "npm:pi-mcp-adapter"}, "packages")
}

// TestInjectOpenCodeMigratesFromOldFormat verifies that when a user's
// opencode.json contains the old v1.11.3 format (separate "args" key),
// Inject() replaces mcp.sdd-memory atomically so that "args" is absent and
// "command" is an array — the format required by OpenCode 1.3.3+.
func TestInjectOpenCodeMigratesFromOldFormat(t *testing.T) {
	home := t.TempDir()

	mockSddMemoryLookPath(t, "/opt/homebrew/bin/sdd-memory", "")

	adapter := opencodeAdapter()
	configPath := adapter.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	// Pre-seed with the old v1.11.3 format.
	oldFormat := `{"mcp": {"sdd-memory": {"command": "/opt/homebrew/bin/sdd-memory", "args": ["mcp"], "type": "local"}}}`
	if err := os.WriteFile(configPath, []byte(oldFormat), 0o644); err != nil {
		t.Fatalf("WriteFile(opencode.json) error = %v", err)
	}

	result, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false; expected migration to produce a change")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	// (1) "args" key must be absent from mcp.sdd-memory.
	if strings.Contains(string(content), `"args"`) {
		t.Fatalf("mcp.sdd-memory still contains 'args' key after migration; got:\n%s", content)
	}

	// (2) command must be a []any containing the sdd-memory binary.
	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}
	mcpMap, _ := parsed["mcp"].(map[string]any)
	sddMemoryMap, _ := mcpMap["sdd-memory"].(map[string]any)
	cmdRaw, ok := sddMemoryMap["command"]
	if !ok {
		t.Fatalf("mcp.sdd-memory missing command key; got:\n%s", content)
	}
	cmdArr, ok := cmdRaw.([]any)
	if !ok {
		t.Fatalf("mcp.sdd-memory.command must be []any after migration, got %T; got:\n%s", cmdRaw, content)
	}
	if len(cmdArr) == 0 {
		t.Fatalf("mcp.sdd-memory.command array is empty; got:\n%s", content)
	}
	firstElem, _ := cmdArr[0].(string)
	if firstElem == "" {
		t.Fatalf("mcp.sdd-memory.command[0] is empty or not a string; got:\n%s", content)
	}
	// Must end with "sdd-memory".
	if filepath.Base(firstElem) != "sdd-memory" {
		t.Fatalf("mcp.sdd-memory.command[0] = %q does not end with 'sdd-memory'; got:\n%s", firstElem, content)
	}

	// (3) Second Inject() call must be idempotent (changed=false).
	second, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true; expected idempotent (no change)")
	}
}

func TestInjectOpenCodeMigratesCellarSDDMemoryCommandToStablePath(t *testing.T) {
	home := t.TempDir()

	mockSddMemoryLookPath(t, "/opt/homebrew/bin/sdd-memory", "")

	adapter := opencodeAdapter()
	configPath := adapter.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	oldFormat := `{"mcp": {"sdd-memory": {"command": ["/opt/homebrew/Cellar/sdd-memory/1.14.1/bin/sdd-memory", "mcp"], "type": "local"}}}`
	if err := os.WriteFile(configPath, []byte(oldFormat), 0o644); err != nil {
		t.Fatalf("WriteFile(opencode.json) error = %v", err)
	}

	result, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false; expected Cellar command migration")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	text := string(content)
	if strings.Contains(text, "/Cellar/") {
		t.Fatalf("opencode.json still contains versioned Homebrew Cellar path; got:\n%s", text)
	}
	if !strings.Contains(text, "/opt/homebrew/bin/sdd-memory") {
		t.Fatalf("opencode.json did not migrate to stable Homebrew symlink; got:\n%s", text)
	}

	second, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true; expected idempotent Cellar migration")
	}
}

func TestInjectCursorMergesSDDMemoryToSettings(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter)
	if injectErr != nil {
		t.Fatalf("Inject(cursor) error = %v", injectErr)
	}

	// Cursor uses MCPConfigFile strategy — sdd-memory gets merged into mcp.json.
	if !result.Changed {
		t.Fatalf("Inject(cursor) changed = false")
	}
}

func TestInjectCursorWithMalformedMCPJsonRecovery(t *testing.T) {
	// Real Windows users may have a ~/.cursor/mcp.json that starts with non-JSON
	// content (e.g. "allow: all" or just "a"). The installer should recover by
	// treating the broken file as {} and proceeding with the overlay merge.
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	// Pre-create ~/.cursor/mcp.json with invalid (non-JSON) content.
	mcpPath := cursorAdapter.MCPConfigPath(home, "sdd-memory")
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(mcpPath, []byte("allow: all"), 0o644); err != nil {
		t.Fatalf("WriteFile(malformed mcp.json) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter)
	if injectErr != nil {
		t.Fatalf("Inject(cursor) with malformed mcp.json error = %v; want nil (should recover)", injectErr)
	}
	if !result.Changed {
		t.Fatalf("Inject(cursor) changed = false; want true")
	}

	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(mcp.json) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, `"mcpServers"`) {
		t.Fatalf("mcp.json missing mcpServers key after recovery; got:\n%s", text)
	}
	if !strings.Contains(text, `"sdd-memory"`) {
		t.Fatalf("mcp.json missing sdd-memory server after recovery; got:\n%s", text)
	}
}

func TestInjectVSCodeMergesSDDMemoryToMCPConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	adapter := vscode.NewAdapter()

	result, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject(vscode) error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject(vscode) changed = false")
	}

	mcpPath := adapter.MCPConfigPath(home, "sdd-memory")
	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(mcp.json) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, `"servers"`) {
		t.Fatal("mcp.json missing servers key")
	}
	if !strings.Contains(text, `"sdd-memory"`) {
		t.Fatal("mcp.json missing sdd-memory server")
	}
	if !strings.Contains(text, `"mcp"`) {
		t.Fatal("mcp.json missing sdd-memory args mcp")
	}
	if strings.Contains(text, `"mcpServers"`) {
		t.Fatal("mcp.json should use 'servers' key, not 'mcpServers'")
	}
}

// ─── Gemini tests ─────────────────────────────────────────────────────────────

func TestInjectGeminiSDDMemoryPresent(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, geminiAdapter())
	if err != nil {
		t.Fatalf("Inject(gemini) error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject(gemini) changed = false")
	}

	settingsPath := filepath.Join(home, ".gemini", "settings.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `"mcpServers"`) {
		t.Fatal("settings.json missing mcpServers key")
	}
	if !strings.Contains(text, `"sdd-memory"`) {
		t.Fatal("settings.json missing sdd-memory entry")
	}
}

func TestInjectAntigravityWritesMCPToCLIConfig(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, antigravityAdapter())
	if err != nil {
		t.Fatalf("Inject(antigravity) error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject(antigravity) changed = false")
	}

	cliMCPPath := filepath.Join(home, ".gemini", "antigravity-cli", "mcp_config.json")
	content, err := os.ReadFile(cliMCPPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", cliMCPPath, err)
	}
	text := string(content)
	if !strings.Contains(text, `"args": [`) || !strings.Contains(text, `"mcp"`) {
		t.Fatalf("Antigravity MCP config must launch sdd-memory MCP; got:\n%s", text)
	}
	if strings.Contains(text, `--tools=`) {
		t.Fatalf("Antigravity should use sdd-memory's default MCP invocation without tool-profile flags; got:\n%s", text)
	}

	pluginPath := filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "specai-sdd-memory", "plugin.json")
	if _, err := os.Stat(pluginPath); err != nil {
		t.Fatalf("Antigravity sdd-memory plugin manifest missing: %v", err)
	}

	pluginMCPPath := filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "specai-sdd-memory", "mcp_config.json")
	pluginMCPContent, err := os.ReadFile(pluginMCPPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", pluginMCPPath, err)
	}
	pluginMCPText := string(pluginMCPContent)
	if !strings.Contains(pluginMCPText, `"mcp"`) || strings.Contains(pluginMCPText, `--tools=`) {
		t.Fatalf("Antigravity sdd-memory plugin MCP config should expose default sdd-memory MCP tools; got:\n%s", pluginMCPText)
	}

	hooksPath := filepath.Join(home, ".gemini", "antigravity-cli", "plugins", "specai-sdd-memory", "hooks.json")
	hooksContent, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", hooksPath, err)
	}
	hooksText := string(hooksContent)
	for _, want := range []string{
		"PreInvocation",
		"injectSteps",
		"mem_save",
		"mem_search",
		"mem_context",
		"mem_session_summary",
		"mem_get_observation",
		"mem_current_project",
		"mem_judge",
	} {
		if !strings.Contains(hooksText, want) {
			t.Fatalf("Antigravity sdd-memory hook missing %q; got:\n%s", want, hooksText)
		}
	}

	desktopMCPPath := filepath.Join(home, ".gemini", "antigravity", "mcp_config.json")
	if _, err := os.Stat(desktopMCPPath); !os.IsNotExist(err) {
		t.Fatalf("legacy desktop MCP path %q should not be written for antigravity; stat err = %v", desktopMCPPath, err)
	}
}

func TestInjectAntigravityInitializesEmptySettingsWhenGeminiMissing(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, antigravityAdapter())
	if err != nil {
		t.Fatalf("Inject(antigravity) first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject(antigravity) first changed = false")
	}

	settingsPath := filepath.Join(home, ".gemini", "antigravity-cli", "settings.json")
	got, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", settingsPath, err)
	}
	if strings.TrimSpace(string(got)) != "{}" {
		t.Fatalf("antigravity settings = %q, want empty JSON object", got)
	}

	second, err := Inject(home, antigravityAdapter())
	if err != nil {
		t.Fatalf("Inject(antigravity) second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject(antigravity) second changed = true; want false")
	}
}

// ─── Codex tests ──────────────────────────────────────────────────────────────

func TestInjectCodexWritesTOMLMCP(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, codexAdapter())
	if err != nil {
		t.Fatalf("Inject(codex) error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject(codex) changed = false")
	}

	configPath := filepath.Join(home, ".codex", "config.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config.toml) error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "[mcp_servers.sdd-memory]") {
		t.Fatalf("config.toml missing [mcp_servers.sdd-memory] block; got:\n%s", text)
	}
	// command must reference the sdd-memory binary — either relative ("sdd-memory") or an
	// absolute path (when sdd-memory is on PATH). Both are valid.
	if !strings.Contains(text, "command = ") {
		t.Fatalf("config.toml missing command field; got:\n%s", text)
	}
	cmdLine := ""
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "command = ") {
			cmdLine = strings.TrimSpace(line)
			break
		}
	}
	if cmdLine == "" {
		t.Fatalf("config.toml missing command line; got:\n%s", text)
	}
	// The command value must end with "sdd-memory" or "sdd-memory.exe".
	cmdVal := strings.TrimPrefix(cmdLine, "command = ")
	cmdVal = strings.Trim(cmdVal, `"`)
	base := filepath.Base(cmdVal)
	if base != "sdd-memory" && base != "sdd-memory.exe" {
		t.Fatalf("config.toml command %q does not reference sdd-memory binary; got:\n%s", cmdVal, text)
	}
}

func TestInjectCodexWritesInstructionFiles(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, codexAdapter())
	if err != nil {
		t.Fatalf("Inject(codex) error = %v", err)
	}

	instructionsPath := filepath.Join(home, ".codex", "sdd-memory-instructions.md")
	content, err := os.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory-instructions.md) error = %v", err)
	}
	if !strings.Contains(string(content), "mem_save") {
		t.Fatal("sdd-memory-instructions.md missing expected content (mem_save)")
	}

	compactPath := filepath.Join(home, ".codex", "sdd-memory-compact-prompt.md")
	compactContent, err := os.ReadFile(compactPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory-compact-prompt.md) error = %v", err)
	}
	if !strings.Contains(string(compactContent), "FIRST ACTION REQUIRED") {
		t.Fatal("sdd-memory-compact-prompt.md missing expected content (FIRST ACTION REQUIRED)")
	}
}

func TestInjectCodexInjectsTOMLKeys(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, codexAdapter())
	if err != nil {
		t.Fatalf("Inject(codex) error = %v", err)
	}

	configPath := filepath.Join(home, ".codex", "config.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config.toml) error = %v", err)
	}
	text := string(content)

	instructionsPath := filepath.Join(home, ".codex", "sdd-memory-instructions.md")
	if !strings.Contains(text, `model_instructions_file`) {
		t.Fatalf("config.toml missing model_instructions_file key; got:\n%s", text)
	}
	normText := strings.ReplaceAll(strings.ReplaceAll(text, "\\\\", "/"), "\\", "/")
	normInstrPath := filepath.ToSlash(instructionsPath)
	if !strings.Contains(normText, normInstrPath) {
		t.Fatalf("config.toml model_instructions_file does not reference %q; got:\n%s", instructionsPath, text)
	}

	compactPath := filepath.Join(home, ".codex", "sdd-memory-compact-prompt.md")
	if !strings.Contains(text, `experimental_compact_prompt_file`) {
		t.Fatalf("config.toml missing experimental_compact_prompt_file key; got:\n%s", text)
	}
	normCompactPath := filepath.ToSlash(compactPath)
	if !strings.Contains(normText, normCompactPath) {
		t.Fatalf("config.toml experimental_compact_prompt_file does not reference %q; got:\n%s", compactPath, text)
	}
}

// ─── sdd-memory setup absolute path preservation tests ────────────────────────────

// TestInjectClaudePreservesAbsoluteCommandFromSDDMemorySetup verifies that when
// `sdd-memory setup claude-code` has already written an absolute-path command to
// ~/.claude/mcp/sdd-memory.json, a subsequent call to Inject() does NOT overwrite
// the absolute path with the relative "sdd-memory".
func TestInjectClaudePreservesAbsoluteCommandFromSDDMemorySetup(t *testing.T) {
	home := t.TempDir()

	// Simulate what `sdd-memory setup claude-code` writes:
	// an absolute path as the command value.
	absPath := "/opt/homebrew/bin/sdd-memory"
	mcpPath := filepath.Join(home, ".claude", "mcp", "sdd-memory.json")
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	setupContent := []byte(`{
  "command": "/opt/homebrew/bin/sdd-memory",
  "args": ["mcp"]
}
`)
	if err := os.WriteFile(mcpPath, setupContent, 0o644); err != nil {
		t.Fatalf("WriteFile(sdd-memory.json) error = %v", err)
	}

	// Now run Inject — should NOT overwrite the absolute command.
	_, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory.json) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, absPath) {
		t.Fatalf("Inject() overwrote absolute command path; want %q preserved, got:\n%s", absPath, text)
	}
}

// TestInjectClaudePreservesAbsoluteCommandIsIdempotent verifies that calling
// Inject() twice when an absolute-path sdd-memory.json already exists does not
// cause repeated writes (idempotency).
func TestInjectClaudePreservesAbsoluteCommandIsIdempotent(t *testing.T) {
	home := t.TempDir()

	absPath := "/usr/local/bin/sdd-memory"
	mcpPath := filepath.Join(home, ".claude", "mcp", "sdd-memory.json")
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	setupContent := []byte(`{
  "command": "/usr/local/bin/sdd-memory",
  "args": ["mcp"]
}
`)
	if err := os.WriteFile(mcpPath, setupContent, 0o644); err != nil {
		t.Fatalf("WriteFile(sdd-memory.json) error = %v", err)
	}

	first, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}

	second, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true after absolute-path setup; want idempotent (no change)")
	}

	// Absolute path must still be present.
	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory.json) error = %v", err)
	}
	if !strings.Contains(string(content), absPath) {
		t.Fatalf("absolute command path %q was lost after second Inject(); got:\n%s", absPath, string(content))
	}
	_ = first // first result not the focus of this test
}

func TestInjectClaudeMigratesCellarCommandToStablePath(t *testing.T) {
	home := t.TempDir()

	mockSddMemoryLookPath(t, "/usr/local/bin/sdd-memory", "")

	mcpPath := filepath.Join(home, ".claude", "mcp", "sdd-memory.json")
	if err := os.MkdirAll(filepath.Dir(mcpPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	setupContent := []byte(`{
  "command": "/usr/local/Cellar/sdd-memory/1.14.1/bin/sdd-memory",
  "args": ["mcp"]
}
`)
	if err := os.WriteFile(mcpPath, setupContent, 0o644); err != nil {
		t.Fatalf("WriteFile(sdd-memory.json) error = %v", err)
	}

	result, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false; expected Cellar command migration")
	}

	content, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-memory.json) error = %v", err)
	}
	text := string(content)
	if strings.Contains(text, "/Cellar/") {
		t.Fatalf("sdd-memory.json still contains versioned Homebrew Cellar path; got:\n%s", text)
	}
	if !strings.Contains(text, "/usr/local/bin/sdd-memory") {
		t.Fatalf("sdd-memory.json did not migrate to stable Homebrew symlink; got:\n%s", text)
	}
}

func TestInjectCodexIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, codexAdapter())
	if err != nil {
		t.Fatalf("Inject(codex) first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject(codex) first changed = false")
	}

	second, err := Inject(home, codexAdapter())
	if err != nil {
		t.Fatalf("Inject(codex) second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject(codex) second changed = true (should be idempotent)")
	}

	// Verify only one [mcp_servers.sdd-memory] block.
	configPath := filepath.Join(home, ".codex", "config.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config.toml) error = %v", err)
	}
	count := strings.Count(string(content), "[mcp_servers.sdd-memory]")
	if count != 1 {
		t.Fatalf("config.toml has %d [mcp_servers.sdd-memory] blocks, want exactly 1; got:\n%s", count, string(content))
	}
}

// ─── Absolute path resolution tests ──────────────────────────────────────────

// mockSddMemoryLookPath sets SddMemoryLookPath to a mock and restores it after the test.
func mockSddMemoryLookPath(t *testing.T, result string, errMsg string) {
	t.Helper()
	orig := SddMemoryLookPath
	SddMemoryLookPath = func(string) (string, error) {
		if errMsg != "" {
			return "", fmt.Errorf("%s", errMsg)
		}
		return result, nil
	}
	t.Cleanup(func() { SddMemoryLookPath = orig })
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v; content:\n%s", path, err, raw)
	}
	return parsed
}

func nestedValue(t *testing.T, root map[string]any, path ...string) (any, bool) {
	t.Helper()
	var current any = root
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func assertNestedString(t *testing.T, root map[string]any, want string, path ...string) {
	t.Helper()
	got, ok := nestedValue(t, root, path...)
	if !ok {
		t.Fatalf("missing JSON path %v in %#v", path, root)
	}
	if got != want {
		t.Fatalf("JSON path %v = %#v, want %q", path, got, want)
	}
}

func assertNestedStrings(t *testing.T, root map[string]any, want []string, path ...string) {
	t.Helper()
	got, ok := nestedValue(t, root, path...)
	if !ok {
		t.Fatalf("missing JSON path %v in %#v", path, root)
	}
	items, ok := got.([]any)
	if !ok {
		t.Fatalf("JSON path %v = %#v, want string array", path, got)
	}
	if len(items) != len(want) {
		t.Fatalf("JSON path %v length = %d, want %d (%#v)", path, len(items), len(want), got)
	}
	for i, wantItem := range want {
		if items[i] != wantItem {
			t.Fatalf("JSON path %v[%d] = %#v, want %q", path, i, items[i], wantItem)
		}
	}
}

func assertNestedStringsUnordered(t *testing.T, root map[string]any, want []string, path ...string) {
	t.Helper()
	got, ok := nestedValue(t, root, path...)
	if !ok {
		t.Fatalf("missing JSON path %v in %#v", path, root)
	}
	items, ok := got.([]any)
	if !ok {
		t.Fatalf("JSON path %v = %#v, want string array", path, got)
	}
	if len(items) != len(want) {
		t.Fatalf("JSON path %v length = %d, want %d (%#v)", path, len(items), len(want), got)
	}
	remaining := make(map[string]int, len(want))
	for _, item := range want {
		remaining[item]++
	}
	for _, item := range items {
		itemString, ok := item.(string)
		if !ok {
			t.Fatalf("JSON path %v contains non-string item %#v", path, item)
		}
		remaining[itemString]--
	}
	for item, count := range remaining {
		if count != 0 {
			t.Fatalf("JSON path %v missing/extra %q count delta %d; got %#v", path, item, count, got)
		}
	}
}

func assertNestedBool(t *testing.T, root map[string]any, want bool, path ...string) {
	t.Helper()
	got, ok := nestedValue(t, root, path...)
	if !ok {
		t.Fatalf("missing JSON path %v in %#v", path, root)
	}
	if got != want {
		t.Fatalf("JSON path %v = %#v, want %v", path, got, want)
	}
}

func assertNestedMissing(t *testing.T, root map[string]any, path ...string) {
	t.Helper()
	if got, ok := nestedValue(t, root, path...); ok {
		t.Fatalf("JSON path %v present = %#v, want missing", path, got)
	}
}

// TestSDDMemoryInjectUsesAbsolutePathWhenAvailable verifies that when sdd-memory is
// resolvable on PATH, its absolute path is written into the MCP config file
// for agents that use StrategyMCPConfigFile (e.g. Windsurf).
func TestSDDMemoryInjectUsesAbsolutePathWhenAvailable(t *testing.T) {
	home := t.TempDir()

	absPath := "/usr/local/bin/sdd-memory"
	mockSddMemoryLookPath(t, absPath, "")

	windsurfAdapter, err := agents.NewAdapter("windsurf")
	if err != nil {
		t.Fatalf("NewAdapter(windsurf) error = %v", err)
	}

	result, injectErr := Inject(home, windsurfAdapter)
	if injectErr != nil {
		t.Fatalf("Inject(windsurf) error = %v", injectErr)
	}
	if !result.Changed {
		t.Fatalf("Inject(windsurf) changed = false")
	}

	mcpPath := windsurfAdapter.MCPConfigPath(home, "sdd-memory")
	content, readErr := os.ReadFile(mcpPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", mcpPath, readErr)
	}

	// Parse and validate the command field contains the absolute path.
	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v", mcpPath, err)
	}

	mcpServersRaw, ok := parsed["mcpServers"]
	if !ok {
		t.Fatalf("mcp_config.json missing mcpServers key; got:\n%s", content)
	}
	mcpServers, ok := mcpServersRaw.(map[string]any)
	if !ok {
		t.Fatalf("mcpServers has unexpected type: %T", mcpServersRaw)
	}
	sddMemoryServerRaw, ok := mcpServers["sdd-memory"]
	if !ok {
		t.Fatalf("mcpServers missing sdd-memory entry; got:\n%s", content)
	}
	sddMemoryServer, ok := sddMemoryServerRaw.(map[string]any)
	if !ok {
		t.Fatalf("sdd-memory server has unexpected type: %T", sddMemoryServerRaw)
	}

	cmd, _ := sddMemoryServer["command"].(string)
	if cmd != absPath {
		t.Fatalf("mcp_config.json command = %q, want absolute path %q", cmd, absPath)
	}
}

// TestSDDMemoryInjectFallsBackToRelativeWhenNotFound verifies that when sdd-memory
// cannot be resolved on PATH, the config falls back to the relative "sdd-memory"
// command string.
func TestSDDMemoryInjectFallsBackToRelativeWhenNotFound(t *testing.T) {
	home := t.TempDir()

	mockSddMemoryLookPath(t, "", "not found")

	windsurfAdapter, err := agents.NewAdapter("windsurf")
	if err != nil {
		t.Fatalf("NewAdapter(windsurf) error = %v", err)
	}

	result, injectErr := Inject(home, windsurfAdapter)
	if injectErr != nil {
		t.Fatalf("Inject(windsurf) error = %v", injectErr)
	}
	if !result.Changed {
		t.Fatalf("Inject(windsurf) changed = false")
	}

	mcpPath := windsurfAdapter.MCPConfigPath(home, "sdd-memory")
	content, readErr := os.ReadFile(mcpPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", mcpPath, readErr)
	}

	text := string(content)
	if !strings.Contains(text, `"command": "sdd-memory"`) {
		t.Fatalf("mcp_config.json should use relative fallback 'sdd-memory'; got:\n%s", text)
	}
}

func TestQwenSDDMemoryIdempotency(t *testing.T) {
	orig := SddMemoryLookPath
	t.Cleanup(func() { SddMemoryLookPath = orig })

	homeDir := t.TempDir()
	adapter := qwenAdapter()
	settingsPath := adapter.SettingsPath(homeDir)

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		t.Fatal(err)
	}

	SddMemoryLookPath = func(string) (string, error) {
		return "", os.ErrNotExist
	}

	_, err := Inject(homeDir, adapter)
	if err != nil {
		t.Fatalf("First injection failed: %v", err)
	}

	content1, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate sdd-memory being found later (e.g. after go install or manual install)
	absPath := "/usr/local/bin/sdd-memory"
	SddMemoryLookPath = func(string) (string, error) {
		return absPath, nil
	}

	_, err = Inject(homeDir, adapter)
	if err != nil {
		t.Fatalf("Second injection failed: %v", err)
	}

	content2, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(content1) != string(content2) {
		t.Errorf("Idempotency failure! Settings changed between runs despite sdd-memory command being stable-relative.\nRun 1:\n%s\nRun 2:\n%s", string(content1), string(content2))
	}
}

func TestInjectOpenClawMergesSDDMemoryIntoMCPServersPreservingStdioAndRemoteFields(t *testing.T) {
	mockSddMemoryLookPath(t, "sdd-memory", "")

	home := t.TempDir()
	adapter := openclawAdapter()
	configPath := adapter.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existing := `{
  "mcp": {
    "sessionIdleTtlMs": 120000,
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem"],
        "env": {"ROOT": "/workspace"},
        "unknownStdioField": true
      },
      "linear": {
        "url": "https://mcp.linear.app/sse",
        "transport": "sse",
        "headers": {"Authorization": "Bearer existing-token"},
        "unknownRemoteField": "preserve-me"
      }
    }
  },
  "theme": "kanagawa"
}`
	if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile(openclaw.json) error = %v", err)
	}

	result, err := Inject(home, adapter)
	if err != nil {
		t.Fatalf("Inject(openclaw) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(openclaw) changed = false")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(openclaw.json) error = %v", err)
	}
	root := unmarshalObjectForTest(t, content)
	mcp := objectAtForTest(t, root, "mcp")
	if got := mcp["sessionIdleTtlMs"]; got != float64(120000) {
		t.Fatalf("mcp.sessionIdleTtlMs = %v, want preserved 120000", got)
	}
	servers := objectAtForTest(t, mcp, "servers")

	filesystem := objectAtForTest(t, servers, "filesystem")
	if got := filesystem["command"]; got != "npx" {
		t.Fatalf("filesystem.command = %v, want npx", got)
	}
	if got := filesystem["unknownStdioField"]; got != true {
		t.Fatalf("filesystem.unknownStdioField = %v, want true", got)
	}
	args, ok := filesystem["args"].([]any)
	if !ok || len(args) != 2 || args[1] != "@modelcontextprotocol/server-filesystem" {
		t.Fatalf("filesystem.args = %#v, want preserved stdio args", filesystem["args"])
	}

	linear := objectAtForTest(t, servers, "linear")
	if got := linear["url"]; got != "https://mcp.linear.app/sse" {
		t.Fatalf("linear.url = %v, want preserved remote url", got)
	}
	if got := linear["transport"]; got != "sse" {
		t.Fatalf("linear.transport = %v, want sse", got)
	}
	headers := objectAtForTest(t, linear, "headers")
	if got := headers["Authorization"]; got != "Bearer existing-token" {
		t.Fatalf("linear Authorization header = %v, want preserved token", got)
	}
	if got := linear["unknownRemoteField"]; got != "preserve-me" {
		t.Fatalf("linear.unknownRemoteField = %v, want preserve-me", got)
	}

	sddMemory := objectAtForTest(t, servers, "sdd-memory")
	if got := sddMemory["command"]; got != "sdd-memory" {
		t.Fatalf("sdd-memory.command = %v, want sdd-memory", got)
	}
	sddMemoryArgs, ok := sddMemory["args"].([]any)
	if !ok || len(sddMemoryArgs) != 1 || sddMemoryArgs[0] != "mcp" {
		t.Fatalf("sdd-memory.args = %#v, want [mcp]", sddMemory["args"])
	}
	if _, hasMCPServers := root["mcpServers"]; hasMCPServers {
		t.Fatal("OpenClaw config must use mcp.servers, not top-level mcpServers")
	}
}

func TestInjectOpenClawHandlesJSON5ConfigAndMissingConfigPath(t *testing.T) {
	t.Run("preserves JSON5 compatible config content as normalized JSON", func(t *testing.T) {
		home := t.TempDir()
		adapter := openclawAdapter()
		configPath := adapter.SettingsPath(home)
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}

		existing := `{
  // OpenClaw user config with JSON5-style comments.
  "ui": {
    "theme": "kanagawa", // trailing comma survives via normalization
  },
  "mcp": {
    "servers": {
      "remoteDocs": {
        "url": "https://docs.example/mcp",
        "transport": "http",
      },
    },
  },
}`
		if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
			t.Fatalf("WriteFile(openclaw.json) error = %v", err)
		}

		if _, err := Inject(home, adapter); err != nil {
			t.Fatalf("Inject(openclaw) error = %v", err)
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("ReadFile(openclaw.json) error = %v", err)
		}
		root := unmarshalObjectForTest(t, content)
		ui := objectAtForTest(t, root, "ui")
		if got := ui["theme"]; got != "kanagawa" {
			t.Fatalf("ui.theme = %v, want preserved kanagawa", got)
		}
		servers := objectAtForTest(t, objectAtForTest(t, root, "mcp"), "servers")
		remoteDocs := objectAtForTest(t, servers, "remoteDocs")
		if got := remoteDocs["transport"]; got != "http" {
			t.Fatalf("remoteDocs.transport = %v, want preserved http", got)
		}
		if _, ok := servers["sdd-memory"]; !ok {
			t.Fatal("mcp.servers.sdd-memory missing after JSON5 merge")
		}
	})

	t.Run("creates canonical OpenClaw config when missing", func(t *testing.T) {
		home := t.TempDir()
		adapter := openclawAdapter()
		configPath := filepath.Join(home, ".openclaw", "openclaw.json")

		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Fatalf("expected missing OpenClaw config before inject, got err=%v", err)
		}
		if _, err := Inject(home, adapter); err != nil {
			t.Fatalf("Inject(openclaw) error = %v", err)
		}
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("ReadFile(created openclaw.json) error = %v", err)
		}
		servers := objectAtForTest(t, objectAtForTest(t, unmarshalObjectForTest(t, content), "mcp"), "servers")
		if _, ok := servers["sdd-memory"]; !ok {
			t.Fatal("created OpenClaw config missing mcp.servers.sdd-memory")
		}
	})
}

func TestInjectOpenClawWritesSDDMemoryProtocolToWorkspaceAgentsOnly(t *testing.T) {
	workspace := t.TempDir()
	adapter := openclawAdapter()
	toolsPath := filepath.Join(workspace, "TOOLS.md")
	if err := os.WriteFile(toolsPath, []byte("# Tool guidance\n\nUser-owned tool notes.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(TOOLS.md) error = %v", err)
	}

	first, err := Inject(workspace, adapter)
	if err != nil {
		t.Fatalf("Inject(openclaw) first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Inject(openclaw) first changed = false")
	}

	agentsPath := filepath.Join(workspace, "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	agentsText := string(agentsContent)
	for _, want := range []string{
		"<!-- specai:sdd-memory-protocol -->",
		"<!-- /specai:sdd-memory-protocol -->",
		"mem_save",
	} {
		if !strings.Contains(agentsText, want) {
			t.Fatalf("OpenClaw AGENTS.md missing sdd-memory protocol content %q; got:\n%s", want, agentsText)
		}
	}
	if _, err := os.Stat(filepath.Join(workspace, ".openclaw", "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("OpenClaw sdd-memory injection must not write global .openclaw/AGENTS.md; stat err=%v", err)
	}

	toolsContent, err := os.ReadFile(toolsPath)
	if err != nil {
		t.Fatalf("ReadFile(TOOLS.md) error = %v", err)
	}
	toolsText := string(toolsContent)
	if strings.Contains(toolsText, "specai:sdd-memory-protocol") || strings.Contains(toolsText, "mem_save") {
		t.Fatalf("TOOLS.md must not receive sdd-memory protocol sections; got:\n%s", toolsText)
	}
	if !strings.Contains(toolsText, "User-owned tool notes.") {
		t.Fatalf("TOOLS.md user content was modified; got:\n%s", toolsText)
	}

	second, err := Inject(workspace, adapter)
	if err != nil {
		t.Fatalf("Inject(openclaw) second error = %v", err)
	}
	if second.Changed {
		t.Fatal("OpenClaw sdd-memory injection should be idempotent")
	}
	updated, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) second error = %v", err)
	}
	if count := strings.Count(string(updated), "<!-- specai:sdd-memory-protocol -->"); count != 1 {
		t.Fatalf("AGENTS.md has %d sdd-memory protocol markers, want exactly 1", count)
	}
}

func TestInjectOpenClawRejectsAmbiguousWorkspacePath(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)

	result, err := Inject("", openclawAdapter())
	if err == nil {
		t.Fatalf("Inject(openclaw, empty workspace) error = nil, want deterministic ambiguity error; result=%+v", result)
	}
	if _, statErr := os.Stat(filepath.Join(cwd, ".openclaw", "openclaw.json")); !os.IsNotExist(statErr) {
		t.Fatalf("ambiguous OpenClaw workspace must not create relative config; stat err=%v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(cwd, "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("ambiguous OpenClaw workspace must not create relative AGENTS.md; stat err=%v", statErr)
	}
}

func unmarshalObjectForTest(t *testing.T, content []byte) map[string]any {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal JSON error = %v; content:\n%s", err, content)
	}
	return root
}

func objectAtForTest(t *testing.T, root map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := root[key]
	if !ok {
		t.Fatalf("missing object key %q in %#v", key, root)
	}
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("key %q has type %T, want object", key, value)
	}
	return object
}
