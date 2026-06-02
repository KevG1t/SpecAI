package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/util"
	"github.com/spf13/afero"
)

// sddMemoryArgs are the arguments passed to the sdd-memory MCP server.
// The --tools=agent flag restricts the exposed tools to the agent subset.
var sddMemoryArgs = []string{"mcp", "--tools=agent"}

// StepInjectMCP writes the sdd-memory MCP config into each IDE's config file.
type StepInjectMCP struct {
	ctx *InstallContext
	fs  afero.Fs
}

func NewStepInjectMCP(ctx *InstallContext) *StepInjectMCP {
	return &StepInjectMCP{ctx: ctx}
}

func (s *StepInjectMCP) ID() string {
	return "Inyectando configuración MCP de sdd-memory"
}

func (s *StepInjectMCP) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

func (s *StepInjectMCP) Run() error {
	cmd := resolveBinary("sdd-memory")

	for _, ide := range s.ctx.IDEs {
		if !ide.SupportsMCP() {
			continue
		}
		var err error
		switch ide.MCPStrategy() {
		case model.StrategySeparateMCPFiles:
			err = s.writeSeparateMCPFile(ide.MCPConfigPath(s.ctx.HomeDir, "sdd-memory"), cmd)
			if err == nil {
				err = s.writeClaudeHooks(ide.SettingsPath(s.ctx.HomeDir))
			}
		case model.StrategyMergeIntoSettings:
			err = s.writeMergeIntoSettings(ide.SettingsPath(s.ctx.HomeDir), cmd)
		case model.StrategyMCPConfigFile:
			err = s.writeMCPConfigFile(ide.MCPConfigPath(s.ctx.HomeDir, "sdd-memory"), ide.AgentID(), cmd)
			if err == nil && ide.AgentID() == model.AgentAntigravity {
				err = s.writeAntigravityPluginFiles()
			}
		case model.StrategyTOMLFile:
			err = s.writeTOMLFile(ide.MCPConfigPath(s.ctx.HomeDir, "sdd-memory"), cmd)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", ide.Name(), err)
		}
	}
	return nil
}

func resolveBinary(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return name
}

// writeSeparateMCPFile writes a per-server MCP JSON file to configPath.
func (s *StepInjectMCP) writeSeparateMCPFile(configPath string, cmd string) error {
	dir := filepath.Dir(configPath)
	if err := s.filesystem().MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := map[string]any{
		"command": cmd,
		"args":    sddMemoryArgs,
	}
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.filesystem(), configPath, data, 0644)
}

// writeMergeIntoSettings merges MCP config into the agent's settings file at settingsPath.
func (s *StepInjectMCP) writeMergeIntoSettings(settingsPath string, cmd string) error {
	cmdArgs := append([]string{cmd}, sddMemoryArgs...)
	overlay := map[string]any{
		"mcp": map[string]any{
			"sdd-memory": map[string]any{
				"command": cmdArgs,
				"type":    "local",
			},
		},
	}

	return s.mergeAndWrite(settingsPath, overlay)
}

// writeMCPConfigFile writes MCP config to a dedicated config file at configPath.
// VSCodeCopilot uses "servers" as the root key; all other agents use "mcpServers".
func (s *StepInjectMCP) writeMCPConfigFile(configPath string, agentID model.AgentID, cmd string) error {
	var overlay map[string]any
	if agentID == model.AgentVSCodeCopilot {
		overlay = map[string]any{
			"servers": map[string]any{
				"sdd-memory": map[string]any{
					"command": cmd,
					"args":    sddMemoryArgs,
				},
			},
		}
	} else {
		overlay = map[string]any{
			"mcpServers": map[string]any{
				"sdd-memory": map[string]any{
					"command": cmd,
					"args":    sddMemoryArgs,
				},
			},
		}
	}

	return s.mergeAndWrite(configPath, overlay)
}

// writeTOMLFile writes MCP config as a TOML block to path via text-based upsert.
func (s *StepInjectMCP) writeTOMLFile(path string, cmd string) error {

	// Build TOML args array from sddMemoryArgs: ["mcp", "--tools=agent"]
	tomlArgs := ""
	for i, a := range sddMemoryArgs {
		if i > 0 {
			tomlArgs += ", "
		}
		tomlArgs += fmt.Sprintf("%q", a)
	}
	block := fmt.Sprintf("\n[mcp_servers.sdd-memory]\ncommand = %q\nargs = [%s]\nmodel_instructions_file = \"sdd-memory-instructions.md\"\nexperimental_compact_prompt_file = \"sdd-memory-compact-prompt.md\"\n", cmd, tomlArgs)

	existing, err := afero.ReadFile(s.filesystem(), path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(existing)

	const header = "[mcp_servers.sdd-memory]"
	if idx := strings.Index(content, header); idx != -1 {
		// Replace the existing block from the header to the next section or EOF.
		end := len(content)
		// Look for the next top-level section after idx.
		nextSection := strings.Index(content[idx+len(header):], "\n[")
		if nextSection != -1 {
			end = idx + len(header) + nextSection + 1 // include the newline before [
		}
		content = content[:idx] + strings.TrimLeft(block, "\n") + content[end:]
	} else {
		content += block
	}

	if err := s.filesystem().MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return writeFileAtomic(s.filesystem(), path, []byte(content), 0644)
}

// writeClaudeHooks writes a UserPromptSubmit hook entry in the agent's settings file
// that executes `specai skill-registry refresh`. The write is idempotent via JSON merge.
func (s *StepInjectMCP) writeClaudeHooks(settingsPath string) error {
	overlay := map[string]any{
		"hooks": map[string]any{
			"UserPromptSubmit": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "specai skill-registry refresh",
						},
					},
				},
			},
		},
	}
	return s.mergeAndWrite(settingsPath, overlay)
}

// writeAntigravityPluginFiles writes plugin.json and hooks.json to the Antigravity
// plugin directory. mcp_config.json is written by writeMCPConfigFile.
func (s *StepInjectMCP) writeAntigravityPluginFiles() error {
	pluginDir := filepath.Join(s.ctx.HomeDir, ".gemini", "antigravity-cli")
	if err := s.filesystem().MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	for _, assetPath := range []string{
		"antigravity/plugin.json",
		"antigravity/hooks.json",
	} {
		data, err := assets.FS.ReadFile(assetPath)
		if err != nil {
			return fmt.Errorf("read asset %s: %w", assetPath, err)
		}
		destPath := filepath.Join(pluginDir, filepath.Base(assetPath))
		if err := writeFileAtomic(s.filesystem(), destPath, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}
	}
	return nil
}

// mergeAndWrite reads an existing JSON file (if any), deep-merges the overlay into it, and writes back atomically.
func (s *StepInjectMCP) mergeAndWrite(path string, overlay map[string]any) error {
	if err := s.filesystem().MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	existing, err := afero.ReadFile(s.filesystem(), path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	merged, err := util.MergeJSON(existing, overlay)
	if err != nil {
		return err
	}

	return writeFileAtomic(s.filesystem(), path, merged, 0644)
}

// vscodeUserConfigDir returns the OS-correct directory for VS Code user config.
// Windows: %APPDATA%/Code/User (fallback ~/.vscode)
// macOS:   ~/Library/Application Support/Code/User
// Linux:   $XDG_CONFIG_HOME/Code/User (fallback ~/.config/Code/User)
func vscodeUserConfigDir(homeDir string) string {
	return vscodeUserConfigDir_forOS(runtime.GOOS, homeDir, os.Getenv("APPDATA"), os.Getenv("XDG_CONFIG_HOME"))
}

// vscodeUserConfigDir_forOS is the testable variant that accepts OS/env values explicitly.
func vscodeUserConfigDir_forOS(goos, homeDir, appData, xdgConfig string) string {
	switch goos {
	case "windows":
		if appData != "" {
			return appData + "/Code/User"
		}
		return homeDir + "/.vscode"
	case "darwin":
		return homeDir + "/Library/Application Support/Code/User"
	default: // linux and others
		if xdgConfig != "" {
			return xdgConfig + "/Code/User"
		}
		return homeDir + "/.config/Code/User"
	}
}

// writeFileAtomic writes data to path via a temp file + rename.
func writeFileAtomic(fsys afero.Fs, path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := afero.WriteFile(fsys, tmp, data, perm); err != nil {
		return err
	}
	return fsys.Rename(tmp, path)
}
