package sddmemory

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KevG1t/specai/internal/agents"
	"github.com/KevG1t/specai/internal/assets"
	"github.com/KevG1t/specai/internal/components/filemerge"
	"github.com/KevG1t/specai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

// bootstrapper is an optional adapter capability: if an adapter implements
// this interface, any injector that writes Jinja modules will first ensure
// the base template (entry point) exists.
type bootstrapper interface {
	BootstrapTemplate(homeDir string) error
}

type piSddMemoryProvisioner interface {
	ProvisionSddMemoryMCP(homeDir string) (changed bool, files []string, err error)
}

// SddMemoryLookPath is the function used to resolve the sdd-memory binary path.
// It is a package-level variable so it can be replaced in tests — both from
// within the sddmemory package and from external test packages (e.g. golden_test.go).
// In production it is set to exec.LookPath.
var SddMemoryLookPath = exec.LookPath

// SetLookPathForTest replaces SddMemoryLookPath with a mock for the duration of
// a test and restores the original after the test completes. Exported so that
// external test packages (e.g. golden_test.go in components) can control the
// resolved sdd-memory path.
func SetLookPathForTest(t interface {
	Helper()
	Cleanup(func())
}, result, errMsg string) {
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

// resolveSddMemoryCommand attempts to resolve the sdd-memory binary to an absolute
// path using exec.LookPath. If found, it returns the absolute path and true.
// If not found (e.g. binary not yet installed), it returns "sdd-memory" and false.
// This is used to write the most stable command possible into MCP configs:
// an absolute path survives across environments where PATH is not fully
// inherited (e.g. Windsurf, IDEs that launch without a login shell).
func resolveSddMemoryCommand() (string, bool) {
	p, err := SddMemoryLookPath("sdd-memory")
	if err != nil || p == "" {
		return "sdd-memory", false
	}
	if isVersionedHomebrewCellarPath(p) {
		return "sdd-memory", false
	}
	return p, true
}

// sddMemoryServerJSON returns the MCP server config bytes, using the absolute
// path to the sdd-memory binary if it can be resolved via PATH.
func sddMemoryServerJSON() []byte {
	cmd, _ := resolveSddMemoryCommand()
	return sddMemoryServerJSONWithCmd(cmd)
}

// sddMemoryServerJSONWithCmd returns the MCP server config bytes for a specific
// command.
func sddMemoryServerJSONWithCmd(cmd string) []byte {
	cfg := map[string]any{
		"command": cmd,
		"args":    []string{"mcp"},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return append(b, '\n')
}

// sddMemoryOverlayJSON returns the settings overlay JSON (used for merge-into-settings
// and MCPConfigFile strategies), with the resolved sdd-memory command.
func sddMemoryOverlayJSON(agentID model.AgentID, cmd string) []byte {
	var cfg map[string]any
	if agentID == model.AgentOpenCode || agentID == model.AgentKilocode {
		// OpenCode 1.3.3+ requires command as an array for type:local servers.
		// The separate "args" field is not accepted; all args must be in the
		// command array itself.
		//
		// Use the __replace__ sentinel so that MergeJSONObjects replaces the
		// entire mcp.sdd-memory object atomically instead of deep-merging into it.
		// Without this, users upgrading from v1.11.3 (which had a separate
		// "args" key) would end up with both "args" and the new array "command"
		// in their config, which is invalid for OpenCode 1.3.3.
		cfg = map[string]any{
			"mcp": map[string]any{
				"sdd-memory": map[string]any{
					"__replace__": map[string]any{
						"command": []string{cmd, "mcp"},
						"type":    "local",
					},
				},
			},
		}
	} else if agentID == model.AgentOpenClaw {
		cfg = map[string]any{
			"mcp": map[string]any{
				"servers": map[string]any{
					"sdd-memory": map[string]any{
						"__replace__": map[string]any{
							"command": cmd,
							"args":    []string{"mcp"},
						},
					},
				},
			},
		}
	} else {
		cfg = map[string]any{
			"mcpServers": map[string]any{
				"sdd-memory": map[string]any{
					"command": cmd,
					"args":    []string{"mcp"},
				},
			},
		}
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return append(b, '\n')
}

// vsCodeSddMemoryOverlayJSON is the VS Code mcp.json overlay using the "servers" key.
// Uses the sdd-memory MCP server without narrowing the exposed tool set.
// VS Code uses a fixed "servers" key structure rather than mcpServers, so it
// is kept as a separate helper.
func vsCodeSddMemoryOverlayJSON(cmd string) []byte {
	cfg := map[string]any{
		"servers": map[string]any{
			"sdd-memory": map[string]any{
				"command": cmd,
				"args":    []string{"mcp"},
			},
		},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return append(b, '\n')
}

func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	return inject(homeDir, homeDir, adapter)
}

// InjectWithPromptDir writes sdd-memory's MCP configuration using configHomeDir and
// writes prompt protocol files using promptDir. This is needed for agents such
// as OpenClaw where MCP is loaded from the global config but instructions are
// read from an active workspace.
func InjectWithPromptDir(configHomeDir, promptDir string, adapter agents.Adapter) (InjectionResult, error) {
	return inject(configHomeDir, promptDir, adapter)
}

const antigravitySddMemoryPluginJSON = `{
  "name": "specai-sdd-memory",
  "description": "Loads sdd-memory MCP memory tools for Antigravity sessions.",
  "version": "0.1.0"
}
`

const antigravitySddMemoryToolsMessage = "CRITICAL FIRST ACTION — Ensure these sdd-memory MCP tools are available before responding: mem_save, mem_search, mem_context, mem_session_summary, mem_session_start, mem_session_end, mem_get_observation, mem_suggest_topic_key, mem_capture_passive, mem_save_prompt, mem_update, mem_current_project, mem_judge. If Antigravity defers MCP tools, load/select these tools from the sdd-memory MCP server first. Then call mem_context when the user asks about prior work or the session needs project memory."

func antigravitySddMemoryHooksJSON() []byte {
	cfg := map[string]any{
		"specai-sdd-memory-tools": map[string]any{
			"PreInvocation": []any{
				map[string]any{
					"type": "command",
					"command": "printf '%s\\n' '" + mustJSONString(map[string]any{
						"injectSteps": []any{
							map[string]any{"ephemeralMessage": antigravitySddMemoryToolsMessage},
						},
					}) + "'",
				},
			},
		},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return append(b, '\n')
}

func mustJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func ensureJSONFileIfMissing(path string) (filemerge.WriteResult, error) {
	if _, err := os.Stat(path); err == nil {
		return filemerge.WriteResult{Changed: false}, nil
	} else if !os.IsNotExist(err) {
		return filemerge.WriteResult{}, err
	}
	return filemerge.WriteFileAtomic(path, []byte("{}\n"), 0o644)
}

func installAntigravitySddMemoryPlugin(homeDir, sddMemoryCommand string) (bool, []string, error) {
	pluginDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins", "specai-sdd-memory")
	files := make([]string, 0, 3)
	changed := false

	pluginPath := filepath.Join(pluginDir, "plugin.json")
	pluginWrite, err := filemerge.WriteFileAtomic(pluginPath, []byte(antigravitySddMemoryPluginJSON), 0o644)
	if err != nil {
		return false, nil, fmt.Errorf("write Antigravity sdd-memory plugin manifest: %w", err)
	}
	changed = changed || pluginWrite.Changed
	files = append(files, pluginPath)

	pluginMCPPath := filepath.Join(pluginDir, "mcp_config.json")
	mcpWrite, err := filemerge.WriteFileAtomic(pluginMCPPath, sddMemoryOverlayJSON(model.AgentAntigravity, sddMemoryCommand), 0o644)
	if err != nil {
		return false, nil, fmt.Errorf("write Antigravity sdd-memory plugin MCP config: %w", err)
	}
	changed = changed || mcpWrite.Changed
	files = append(files, pluginMCPPath)

	hooksPath := filepath.Join(pluginDir, "hooks.json")
	hooksWrite, err := filemerge.WriteFileAtomic(hooksPath, antigravitySddMemoryHooksJSON(), 0o644)
	if err != nil {
		return false, nil, fmt.Errorf("write Antigravity sdd-memory hooks: %w", err)
	}
	changed = changed || hooksWrite.Changed
	files = append(files, hooksPath)

	return changed, files, nil
}

func inject(configHomeDir, promptDir string, adapter agents.Adapter) (InjectionResult, error) {
	if provisioner, ok := adapter.(piSddMemoryProvisioner); ok {
		changed, files, err := provisioner.ProvisionSddMemoryMCP(configHomeDir)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: changed, Files: files}, nil
	}

	if !adapter.SupportsMCP() {
		return InjectionResult{}, nil
	}
	if err := validateOpenClawWorkspacePath(promptDir, adapter); err != nil {
		return InjectionResult{}, err
	}

	files := make([]string, 0, 2)
	changed := false

	// 1. Write MCP server config using the adapter's strategy.
	switch adapter.MCPStrategy() {
	case model.StrategySeparateMCPFiles:
		// sdd-memory v1.10.3+ writes an absolute path for the command field when
		// `sdd-memory setup <agent>` is invoked. specai's Inject() runs after
		// sdd-memory setup, so we must preserve any absolute command path already
		// present instead of silently overwriting it with the relative "sdd-memory".
		mcpPath := adapter.MCPConfigPath(configHomeDir, "sdd-memory")
		cmd := stableSddMemoryCommandForMergedConfig(mcpPath, adapter.Agent())
		content := buildSeparateMCPContent(mcpPath, sddMemoryServerJSONWithCmd(cmd))
		mcpWrite, err := filemerge.WriteFileAtomic(mcpPath, content, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || mcpWrite.Changed
		files = append(files, mcpPath)

	case model.StrategyMergeIntoSettings:
		settingsPath := adapter.SettingsPath(configHomeDir)
		if settingsPath == "" {
			break
		}
		overlay := sddMemoryOverlayJSON(adapter.Agent(), stableSddMemoryCommandForMergedConfig(settingsPath, adapter.Agent()))
		settingsWrite, err := mergeJSONFile(settingsPath, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || settingsWrite.Changed
		files = append(files, settingsPath)

	case model.StrategyMCPConfigFile:
		mcpPath := adapter.MCPConfigPath(configHomeDir, "sdd-memory")
		if mcpPath == "" {
			break
		}
		sddMemoryCommand := stableSddMemoryCommandForMergedConfig(mcpPath, adapter.Agent())
		var overlay []byte
		if adapter.Agent() == model.AgentVSCodeCopilot {
			overlay = vsCodeSddMemoryOverlayJSON(sddMemoryCommand)
		} else {
			overlay = sddMemoryOverlayJSON(adapter.Agent(), sddMemoryCommand)
		}

		mcpWrite, err := mergeJSONFile(mcpPath, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || mcpWrite.Changed
		files = append(files, mcpPath)

		if adapter.Agent() == model.AgentAntigravity {
			settingsWrite, settingsErr := ensureJSONFileIfMissing(adapter.SettingsPath(configHomeDir))
			if settingsErr != nil {
				return InjectionResult{}, fmt.Errorf("ensure Antigravity settings: %w", settingsErr)
			}
			changed = changed || settingsWrite.Changed
			files = append(files, adapter.SettingsPath(configHomeDir))

			pluginChanged, pluginFiles, pluginErr := installAntigravitySddMemoryPlugin(configHomeDir, sddMemoryCommand)
			if pluginErr != nil {
				return InjectionResult{}, pluginErr
			}
			changed = changed || pluginChanged
			files = append(files, pluginFiles...)
		}

	case model.StrategyTOMLFile:
		// Codex: upsert [mcp_servers.sdd-memory] block and instruction-file keys
		// in ~/.codex/config.toml, then write instruction files.
		// All TOML mutations are composed in a single pass before writing to
		// ensure idempotency (no intermediate states that differ on re-run).
		configPath := adapter.MCPConfigPath(configHomeDir, "sdd-memory")
		if configPath == "" {
			break
		}

		// Determine instruction file paths before mutating the config.
		instructionsPath, compactPath, instrErr := writeCodexInstructionFiles(configHomeDir)
		if instrErr != nil {
			return InjectionResult{}, instrErr
		}

		// Read existing config and apply all mutations in one pass.
		existing, err := readFileOrEmpty(configPath)
		if err != nil {
			return InjectionResult{}, err
		}
		sddMemoryCmd := stableSddMemoryCommandForMergedConfig(configPath, adapter.Agent())
		withMCP := filemerge.UpsertCodexSddMemoryBlock(existing, sddMemoryCmd)
		withInstr := filemerge.UpsertTopLevelTOMLString(withMCP, "model_instructions_file", instructionsPath)
		withCompact := filemerge.UpsertTopLevelTOMLString(withInstr, "experimental_compact_prompt_file", compactPath)

		tomlWrite, err := filemerge.WriteFileAtomic(configPath, []byte(withCompact), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || tomlWrite.Changed
		files = append(files, configPath)
	}

	// 2. Inject sdd-memory protocol into system prompt (if supported).
	if adapter.SupportsSystemPrompt() {
		switch adapter.SystemPromptStrategy() {
		case model.StrategyMarkdownSections:
			promptPath := adapter.SystemPromptFile(promptDir)
			protocolContent := assets.MustRead("claude/sdd-memory-protocol.md")

			existing, err := readFileOrEmpty(promptPath)
			if err != nil {
				return InjectionResult{}, err
			}

			updated := filemerge.InjectMarkdownSection(existing, "sdd-memory-protocol", protocolContent)

			mdWrite, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || mdWrite.Changed
			files = append(files, promptPath)

		case model.StrategyJinjaModules:
			// Ensure the base template exists for Jinja-based agents.
			if bs, ok := adapter.(bootstrapper); ok {
				if err := bs.BootstrapTemplate(promptDir); err != nil {
					return InjectionResult{}, fmt.Errorf("bootstrap template: %w", err)
				}
			}

			// Write the sdd-memory protocol as a standalone Jinja include module.
			// The static KIMI.md template references it via {% include "sdd-memory-protocol.md" %}.
			configDir := adapter.GlobalConfigDir(promptDir)
			protocolContent := assets.MustRead("claude/sdd-memory-protocol.md")
			modulePath := filepath.Join(configDir, "sdd-memory-protocol.md")
			mdWrite, err := filemerge.WriteFileAtomic(modulePath, []byte(protocolContent), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || mdWrite.Changed
			files = append(files, modulePath)

		default:
			promptPath := adapter.SystemPromptFile(promptDir)
			protocolContent := assets.MustRead("claude/sdd-memory-protocol.md")

			existing, err := readFileOrEmpty(promptPath)
			if err != nil {
				return InjectionResult{}, err
			}

			updated := filemerge.InjectMarkdownSection(existing, "sdd-memory-protocol", protocolContent)

			mdWrite, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || mdWrite.Changed
			files = append(files, promptPath)
		}
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func validateOpenClawWorkspacePath(workspaceDir string, adapter agents.Adapter) error {
	if adapter.Agent() == model.AgentOpenClaw && strings.TrimSpace(workspaceDir) == "" {
		return fmt.Errorf("openclaw workspace path is required for workspace-first injection")
	}
	return nil
}

type settingsBootstrapResult struct {
	Changed bool
	Path    string
}

func ensureAntigravitySettings(homeDir string, adapter agents.Adapter) (settingsBootstrapResult, error) {
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		return settingsBootstrapResult{}, nil
	}

	if _, err := os.Stat(settingsPath); err == nil {
		return settingsBootstrapResult{Path: settingsPath}, nil
	} else if !os.IsNotExist(err) {
		return settingsBootstrapResult{}, fmt.Errorf("stat antigravity settings %q: %w", settingsPath, err)
	}

	sourcePath := filepath.Join(homeDir, ".gemini", "settings.json")
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return settingsBootstrapResult{}, fmt.Errorf("read gemini settings %q: %w", sourcePath, err)
		}
		content = []byte("{}")
	}

	writeResult, err := filemerge.WriteFileAtomic(settingsPath, content, 0o644)
	if err != nil {
		return settingsBootstrapResult{}, err
	}

	return settingsBootstrapResult{Changed: writeResult.Changed, Path: settingsPath}, nil
}

// writeCodexInstructionFiles writes the sdd-memory protocol and compact prompt
// files to ~/.codex/ and returns their paths.
func writeCodexInstructionFiles(homeDir string) (instructionsPath, compactPath string, err error) {
	codexDir := filepath.Join(homeDir, ".codex")
	instructionsPath = filepath.Join(codexDir, "sdd-memory-instructions.md")
	compactPath = filepath.Join(codexDir, "sdd-memory-compact-prompt.md")

	instrContent := assets.MustRead("codex/sdd-memory-instructions.md")
	instrWrite, err := filemerge.WriteFileAtomic(instructionsPath, []byte(instrContent), 0o644)
	if err != nil {
		return "", "", fmt.Errorf("write codex sdd-memory-instructions.md: %w", err)
	}
	_ = instrWrite

	compactContent := assets.MustRead("codex/sdd-memory-compact-prompt.md")
	compactWrite, err := filemerge.WriteFileAtomic(compactPath, []byte(compactContent), 0o644)
	if err != nil {
		return "", "", fmt.Errorf("write codex sdd-memory-compact-prompt.md: %w", err)
	}
	_ = compactWrite

	return instructionsPath, compactPath, nil
}

func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	return filemerge.WriteFileAtomic(path, merged, 0o644)
}

var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}

	return content, nil
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read file %q: %w", path, err)
	}
	return string(data), nil
}

func stableSddMemoryCommandForMergedConfig(path string, agentID model.AgentID) string {
	raw, err := osReadFile(path)
	if err == nil {
		if cmd, ok := existingMergedSddMemoryCommand(raw, agentID); ok {
			return stableSddMemoryCommandForExisting(cmd, agentID)
		}
	}

	if isStandardAgent(agentID) {
		return preferredStableSddMemoryCommand()
	}

	cmd, _ := resolveSddMemoryCommand()
	return cmd
}

func stableSddMemoryCommandForExisting(cmd string, agentID model.AgentID) string {
	if isVersionedHomebrewCellarPath(cmd) {
		if stable := preferredStableSddMemoryCommand(); stable != "" {
			return stable
		}
		return "sdd-memory"
	}

	return cmd
}

func preferredStableSddMemoryCommand() string {
	p, err := SddMemoryLookPath("sdd-memory")
	if err == nil && isStableHomebrewSddMemoryPath(p) {
		return p
	}
	return "sdd-memory"
}

func existingMergedSddMemoryCommand(raw []byte, agentID model.AgentID) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}

	normalized, err := filemerge.MergeJSONObjects(raw, []byte("{}"))
	if err != nil {
		return "", false
	}

	var root map[string]any
	if err := json.Unmarshal(normalized, &root); err != nil {
		return "", false
	}

	var server any
	switch agentID {
	case model.AgentOpenCode:
		mcp, ok := root["mcp"].(map[string]any)
		if !ok {
			return "", false
		}
		server = mcp["sdd-memory"]
	case model.AgentOpenClaw:
		mcp, ok := root["mcp"].(map[string]any)
		if !ok {
			return "", false
		}
		servers, ok := mcp["servers"].(map[string]any)
		if !ok {
			return "", false
		}
		server = servers["sdd-memory"]
	case model.AgentVSCodeCopilot:
		servers, ok := root["servers"].(map[string]any)
		if !ok {
			return "", false
		}
		server = servers["sdd-memory"]
	default:
		mcpServers, ok := root["mcpServers"].(map[string]any)
		if !ok {
			return "", false
		}
		server = mcpServers["sdd-memory"]
	}

	serverMap, ok := server.(map[string]any)
	if !ok {
		return "", false
	}

	return executableFromCommandValue(serverMap["command"])
}

func executableFromCommandValue(command any) (string, bool) {
	switch value := command.(type) {
	case string:
		if value == "" {
			return "", false
		}
		return value, true
	case []any:
		if len(value) == 0 {
			return "", false
		}
		first, ok := value[0].(string)
		if !ok || first == "" {
			return "", false
		}
		return first, true
	default:
		return "", false
	}
}

func isStandardAgent(id model.AgentID) bool {
	switch id {
	case model.AgentOpenCode, model.AgentQwenCode, model.AgentCodex, model.AgentGeminiCLI, model.AgentAntigravity, model.AgentClaudeCode, model.AgentOpenClaw:
		return true
	default:
		return false
	}
}

// buildSeparateMCPContent returns the content to write to the MCP server JSON
// file for agents that use the StrategySeparateMCPFiles strategy (e.g. Claude
// Code).
//
// sdd-memory writes an absolute command path when `sdd-memory setup` is run.
// specai runs Inject() after setup, so we must not overwrite that absolute
// path with the relative "sdd-memory" string from defaultSddMemoryServerJSON.
//
// Logic:
//   - If the file does not exist yet, return defaultContent unchanged.
//   - If the file exists but cannot be parsed as JSON, return defaultContent.
//   - If the parsed JSON has a "command" value that is an absolute path to the
//     sdd-memory binary, rebuild the config using that command and the canonical
//     args (["mcp"]) so that the absolute path is preserved and the correct
//     flags are always present.
//   - Otherwise (relative command or other value), return defaultContent.
func buildSeparateMCPContent(mcpPath string, defaultContent []byte) []byte {
	raw, err := os.ReadFile(mcpPath)
	if err != nil {
		// File does not exist or is not readable — use the default.
		return defaultContent
	}

	var existing map[string]any
	if err := json.Unmarshal(raw, &existing); err != nil {
		// Malformed JSON — use the default.
		return defaultContent
	}

	cmd, ok := executableFromCommandValue(existing["command"])
	if !ok || !isSddMemoryCommand(cmd) {
		// No command, or not a sdd-memory command — use the default.
		return defaultContent
	}
	cmd = stableSddMemoryCommandForExisting(cmd, "")

	// Rebuild with the preserved command and the canonical args (["mcp"]).
	rebuilt := map[string]any{
		"command": cmd,
		"args":    []string{"mcp"},
	}
	encoded, err := json.MarshalIndent(rebuilt, "", "  ")
	if err != nil {
		// Should be impossible with a plain map — use the default as fallback.
		return defaultContent
	}
	return append(encoded, '\n')
}

// isSddMemoryCommand reports whether cmd is either a relative "sdd-memory" command
// or an absolute path pointing to a sdd-memory binary.
func isSddMemoryCommand(cmd string) bool {
	if cmd == "" {
		return false
	}
	base := filepath.Base(cmd)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(base, "sdd-memory.exe") || strings.EqualFold(base, "sdd-memory")
	}
	return base == "sdd-memory"
}

// isAbsoluteSddMemoryPath reports whether path is an absolute filesystem path
// that points to a sdd-memory binary.
func isAbsoluteSddMemoryPath(path string) bool {
	return filepath.IsAbs(path) && isSddMemoryCommand(path)
}

func isVersionedHomebrewCellarPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(clean, "/Cellar/sdd-memory/") && isSddMemoryCommand(clean)
}

func isStableHomebrewSddMemoryPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return (clean == "/opt/homebrew/bin/sdd-memory" || clean == "/usr/local/bin/sdd-memory") && isSddMemoryCommand(clean)
}
