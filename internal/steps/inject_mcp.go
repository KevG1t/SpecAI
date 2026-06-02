package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/util"
	"github.com/spf13/afero"
)

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
		var err error
		switch ide.AgentID() {
		case model.AgentClaudeCode:
			err = s.writeSeparateMCPFile(cmd)
			if err == nil {
				err = s.writeClaudeHooks()
			}
		case model.AgentOpenCode, model.AgentKilocode, model.AgentGeminiCLI:
			err = s.writeMergeIntoSettings(ide, cmd)
		case model.AgentCodex:
			err = s.writeTOMLFile(cmd)
		case model.AgentAntigravity:
			err = s.writeMCPConfigFile(ide, cmd)
			if err == nil {
				err = s.writeAntigravityPluginFiles()
			}
		default:
			err = s.writeMCPConfigFile(ide, cmd)
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

// writeSeparateMCPFile handles Claude Code: ~/.claude/mcp/sdd-memory.json
func (s *StepInjectMCP) writeSeparateMCPFile(cmd string) error {
	dir := filepath.Join(s.ctx.HomeDir, ".claude", "mcp")
	if err := s.filesystem().MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := map[string]any{
		"command": cmd,
		"args":    []string{"mcp"},
	}
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.filesystem(), filepath.Join(dir, "sdd-memory.json"), data, 0644)
}

// writeMergeIntoSettings handles OpenCode, KiloCode, GeminiCLI.
func (s *StepInjectMCP) writeMergeIntoSettings(ide interface{ AgentID() model.AgentID }, cmd string) error {
	var settingsPath string
	switch ide.AgentID() {
	case model.AgentOpenCode:
		settingsPath = filepath.Join(s.ctx.HomeDir, ".config", "opencode", "opencode.json")
	case model.AgentKilocode:
		settingsPath = filepath.Join(s.ctx.HomeDir, ".config", "kilo", "opencode.json")
	case model.AgentGeminiCLI:
		settingsPath = filepath.Join(s.ctx.HomeDir, ".gemini", "settings.json")
	default:
		return fmt.Errorf("unsupported agent for StrategyMergeIntoSettings: %s", ide.AgentID())
	}

	overlay := map[string]any{
		"mcp": map[string]any{
			"sdd-memory": map[string]any{
				"command": []string{cmd, "mcp"},
				"type":    "local",
			},
		},
	}

	return s.mergeAndWrite(settingsPath, overlay)
}

// writeMCPConfigFile handles Cursor, Windsurf, Kiro, Antigravity, VSCodeCopilot, Kimi, Qwen, OpenClaw, Pi, Trae.
func (s *StepInjectMCP) writeMCPConfigFile(ide interface{ AgentID() model.AgentID }, cmd string) error {
	var configPath string
	switch ide.AgentID() {
	case model.AgentCursor:
		configPath = filepath.Join(s.ctx.HomeDir, ".cursor", "mcp.json")
	case model.AgentWindsurf:
		configPath = filepath.Join(s.ctx.HomeDir, ".codeium", "windsurf", "mcp_config.json")
	case model.AgentKiroIDE:
		configPath = filepath.Join(s.ctx.HomeDir, ".kiro", "settings", "mcp.json")
	case model.AgentAntigravity:
		configPath = filepath.Join(s.ctx.HomeDir, ".gemini", "antigravity-cli", "mcp_config.json")
	case model.AgentVSCodeCopilot:
		configPath = filepath.Join(s.ctx.HomeDir, ".vscode", "mcp.json")
	case model.AgentKimi:
		configPath = filepath.Join(s.ctx.HomeDir, ".kimi", "mcp.json")
	case model.AgentQwenCode:
		configPath = filepath.Join(s.ctx.HomeDir, ".qwen", "mcp.json")
	case model.AgentOpenClaw:
		configPath = filepath.Join(s.ctx.HomeDir, ".openclaw", "mcp.json")
	case model.AgentPi:
		configPath = filepath.Join(s.ctx.HomeDir, ".pi", "mcp.json")
	case model.AgentTrae:
		configPath = filepath.Join(s.ctx.HomeDir, ".trae", "mcp.json")
	default:
		return fmt.Errorf("unsupported agent for StrategyMCPConfigFile: %s", ide.AgentID())
	}

	var overlay map[string]any
	if ide.AgentID() == model.AgentVSCodeCopilot {
		overlay = map[string]any{
			"servers": map[string]any{
				"sdd-memory": map[string]any{
					"command": cmd,
					"args":    []string{"mcp"},
				},
			},
		}
	} else {
		overlay = map[string]any{
			"mcpServers": map[string]any{
				"sdd-memory": map[string]any{
					"command": cmd,
					"args":    []string{"mcp"},
				},
			},
		}
	}

	return s.mergeAndWrite(configPath, overlay)
}

// writeTOMLFile handles Codex: ~/.codex/config.toml via text-based upsert.
func (s *StepInjectMCP) writeTOMLFile(cmd string) error {
	path := filepath.Join(s.ctx.HomeDir, ".codex", "config.toml")

	block := fmt.Sprintf("\n[mcp_servers.sdd-memory]\ncommand = %q\nargs = [\"mcp\"]\n", cmd)

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

// writeClaudeHooks writes a UserPromptSubmit hook entry in ~/.claude/settings.json
// that executes `specai skill-registry refresh`. The write is idempotent via JSON merge.
func (s *StepInjectMCP) writeClaudeHooks() error {
	settingsPath := filepath.Join(s.ctx.HomeDir, ".claude", "settings.json")
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

// writeFileAtomic writes data to path via a temp file + rename.
func writeFileAtomic(fsys afero.Fs, path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := afero.WriteFile(fsys, tmp, data, perm); err != nil {
		return err
	}
	return fsys.Rename(tmp, path)
}
