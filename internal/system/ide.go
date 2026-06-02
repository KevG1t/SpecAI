package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
)

// vscodeUserConfigDirFromHome returns the OS-correct VS Code user config directory.
// Delegates to runtime.GOOS and env vars so it mirrors the behavior in inject_mcp.go.
func vscodeUserConfigDirFromHome(homeDir string) string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData + "/Code/User"
		}
		return homeDir + "/.vscode"
	case "darwin":
		return homeDir + "/Library/Application Support/Code/User"
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg + "/Code/User"
		}
		return homeDir + "/.config/Code/User"
	}
}

// IDEAdapter models paths and characteristics for a specific IDE/agent globally and locally.
type IDEAdapter interface {
	// Identity
	Name() string
	AgentID() model.AgentID
	AssetFolder() string

	// Config paths (existing)
	ConfigDir(homeDir string) string
	GlobalRulesDir(homeDir string) string
	GlobalSkillsDir(homeDir string) string
	LocalRulesFile() string // relative rules file path in current directory, e.g. ".cursorrules"
	LocalSkillsDir() string // relative skills directory path in current directory, e.g. ".agents/skills"

	// Strategy methods — HOW to inject content
	SystemPromptStrategy() model.SystemPromptStrategy
	MCPStrategy() model.MCPStrategy

	// Agent-specific path resolvers
	SystemPromptFile(homeDir string) string
	SubAgentsDir(homeDir string) string
	EmbeddedSubAgentsDir() string
	CommandsDir(homeDir string) string
	EmbeddedCommandsDir() string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string
	MCPConfigPath(homeDir string, serverName string) string

	// Capability flags
	SupportsSubAgents() bool
	SupportsSlashCommands() bool
	SupportsSkills() bool
	SupportsSystemPrompt() bool
	SupportsMCP() bool
}

// CursorAdapter implements IDEAdapter for Cursor.
type CursorAdapter struct{}

func (CursorAdapter) Name() string             { return "Cursor" }
func (CursorAdapter) AgentID() model.AgentID   { return model.AgentCursor }
func (CursorAdapter) AssetFolder() string      { return "cursor" }
func (CursorAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".cursor")
}
func (CursorAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".cursor", "rules")
}
func (CursorAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".cursor", "skills")
}
func (CursorAdapter) LocalRulesFile() string { return ".cursorrules" }
func (CursorAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (CursorAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (CursorAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (CursorAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".cursor", "rules", "specai.mdc")
}
func (CursorAdapter) SubAgentsDir(homeDir string) string    { return filepath.Join(homeDir, ".cursor", "agents") }
func (CursorAdapter) EmbeddedSubAgentsDir() string          { return "cursor/agents" }
func (CursorAdapter) CommandsDir(_ string) string           { return "" }
func (CursorAdapter) EmbeddedCommandsDir() string           { return "" }
func (CursorAdapter) SkillsDir(homeDir string) string       { return filepath.Join(homeDir, ".cursor", "skills") }
func (CursorAdapter) SettingsPath(homeDir string) string    { return filepath.Join(homeDir, ".cursor", "settings.json") }
func (CursorAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".cursor", "mcp.json")
}
func (CursorAdapter) SupportsSubAgents() bool    { return true }
func (CursorAdapter) SupportsSlashCommands() bool { return false }
func (CursorAdapter) SupportsSkills() bool        { return true }
func (CursorAdapter) SupportsSystemPrompt() bool  { return true }
func (CursorAdapter) SupportsMCP() bool           { return true }

// WindsurfAdapter implements IDEAdapter for Windsurf.
type WindsurfAdapter struct{}

func (WindsurfAdapter) Name() string             { return "Windsurf" }
func (WindsurfAdapter) AgentID() model.AgentID   { return model.AgentWindsurf }
func (WindsurfAdapter) AssetFolder() string      { return "windsurf" }
func (WindsurfAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".codeium", "windsurf")
}
func (WindsurfAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".codeium", "windsurf", "rules")
}
func (WindsurfAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".codeium", "windsurf", "skills")
}
func (WindsurfAdapter) LocalRulesFile() string { return ".windsurfrules" }
func (WindsurfAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (WindsurfAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyAppendToFile }
func (WindsurfAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (WindsurfAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".codeium", "windsurf", "rules", "specai.md")
}
func (WindsurfAdapter) SubAgentsDir(_ string) string       { return "" }
func (WindsurfAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (WindsurfAdapter) CommandsDir(_ string) string        { return "" }
func (WindsurfAdapter) EmbeddedCommandsDir() string        { return "" }
func (WindsurfAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".codeium", "windsurf", "skills") }
func (WindsurfAdapter) SettingsPath(_ string) string       { return "" }
func (WindsurfAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json")
}
func (WindsurfAdapter) SupportsSubAgents() bool    { return false }
func (WindsurfAdapter) SupportsSlashCommands() bool { return false }
func (WindsurfAdapter) SupportsSkills() bool        { return true }
func (WindsurfAdapter) SupportsSystemPrompt() bool  { return true }
func (WindsurfAdapter) SupportsMCP() bool           { return true }

// CodexAdapter implements IDEAdapter for Codex.
type CodexAdapter struct{}

func (CodexAdapter) Name() string             { return "Codex" }
func (CodexAdapter) AgentID() model.AgentID   { return model.AgentCodex }
func (CodexAdapter) AssetFolder() string      { return "codex" }
func (CodexAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".codex")
}
func (CodexAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "rules")
}
func (CodexAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "skills")
}
func (CodexAdapter) LocalRulesFile() string { return ".codexrules" }
func (CodexAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (CodexAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyAppendToFile }
func (CodexAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyTOMLFile }
func (CodexAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "AGENTS.md")
}
func (CodexAdapter) SubAgentsDir(_ string) string       { return "" }
func (CodexAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (CodexAdapter) CommandsDir(_ string) string        { return "" }
func (CodexAdapter) EmbeddedCommandsDir() string        { return "" }
func (CodexAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".codex", "skills") }
func (CodexAdapter) SettingsPath(_ string) string       { return "" }
func (CodexAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".codex", "config.toml")
}
func (CodexAdapter) SupportsSubAgents() bool    { return false }
func (CodexAdapter) SupportsSlashCommands() bool { return false }
func (CodexAdapter) SupportsSkills() bool        { return true }
func (CodexAdapter) SupportsSystemPrompt() bool  { return true }
func (CodexAdapter) SupportsMCP() bool           { return true }

// ClaudeCodeAdapter implements IDEAdapter for Claude Code.
type ClaudeCodeAdapter struct{}

func (ClaudeCodeAdapter) Name() string             { return "Claude Code" }
func (ClaudeCodeAdapter) AgentID() model.AgentID   { return model.AgentClaudeCode }
func (ClaudeCodeAdapter) AssetFolder() string      { return "claude" }
func (ClaudeCodeAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude")
}
func (ClaudeCodeAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "rules")
}
func (ClaudeCodeAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
}
func (ClaudeCodeAdapter) LocalRulesFile() string { return "CLAUDE.md" }
func (ClaudeCodeAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (ClaudeCodeAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyMarkdownSections }
func (ClaudeCodeAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategySeparateMCPFiles }
func (ClaudeCodeAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "CLAUDE.md")
}
func (ClaudeCodeAdapter) SubAgentsDir(homeDir string) string { return filepath.Join(homeDir, ".claude", "agents") }
func (ClaudeCodeAdapter) EmbeddedSubAgentsDir() string       { return "claude/agents" }
func (ClaudeCodeAdapter) CommandsDir(homeDir string) string  { return filepath.Join(homeDir, ".claude", "commands") }
func (ClaudeCodeAdapter) EmbeddedCommandsDir() string        { return "claude/commands" }
func (ClaudeCodeAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".claude", "skills") }
func (ClaudeCodeAdapter) SettingsPath(homeDir string) string { return filepath.Join(homeDir, ".claude", "settings.json") }
func (ClaudeCodeAdapter) MCPConfigPath(homeDir string, serverName string) string {
	return filepath.Join(homeDir, ".claude", "mcp", serverName+".json")
}
func (ClaudeCodeAdapter) SupportsSubAgents() bool    { return true }
func (ClaudeCodeAdapter) SupportsSlashCommands() bool { return true }
func (ClaudeCodeAdapter) SupportsSkills() bool        { return true }
func (ClaudeCodeAdapter) SupportsSystemPrompt() bool  { return true }
func (ClaudeCodeAdapter) SupportsMCP() bool           { return true }

// GeminiCLIAdapter implements IDEAdapter for Gemini CLI.
type GeminiCLIAdapter struct{}

func (GeminiCLIAdapter) Name() string             { return "Gemini CLI" }
func (GeminiCLIAdapter) AgentID() model.AgentID   { return model.AgentGeminiCLI }
func (GeminiCLIAdapter) AssetFolder() string      { return "gemini" }
func (GeminiCLIAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini")
}
func (GeminiCLIAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "rules")
}
func (GeminiCLIAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "skills")
}
func (GeminiCLIAdapter) LocalRulesFile() string { return "GEMINI.md" }
func (GeminiCLIAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (GeminiCLIAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (GeminiCLIAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMergeIntoSettings }
func (GeminiCLIAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "GEMINI.md")
}
func (GeminiCLIAdapter) SubAgentsDir(_ string) string       { return "" }
func (GeminiCLIAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (GeminiCLIAdapter) CommandsDir(_ string) string        { return "" }
func (GeminiCLIAdapter) EmbeddedCommandsDir() string        { return "" }
func (GeminiCLIAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".gemini", "skills") }
func (GeminiCLIAdapter) SettingsPath(homeDir string) string { return filepath.Join(homeDir, ".gemini", "settings.json") }
func (GeminiCLIAdapter) MCPConfigPath(_ string, _ string) string { return "" }
func (GeminiCLIAdapter) SupportsSubAgents() bool    { return false }
func (GeminiCLIAdapter) SupportsSlashCommands() bool { return false }
func (GeminiCLIAdapter) SupportsSkills() bool        { return true }
func (GeminiCLIAdapter) SupportsSystemPrompt() bool  { return true }
func (GeminiCLIAdapter) SupportsMCP() bool           { return true }

// AntigravityAdapter implements IDEAdapter for Google Antigravity (Gemini sub-agent env).
type AntigravityAdapter struct{}

func (AntigravityAdapter) Name() string             { return "Antigravity" }
func (AntigravityAdapter) AgentID() model.AgentID   { return model.AgentAntigravity }
func (AntigravityAdapter) AssetFolder() string      { return "antigravity" }
func (AntigravityAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli")
}
func (AntigravityAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "rules")
}
func (AntigravityAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills")
}
func (AntigravityAdapter) LocalRulesFile() string { return ".antigravity-rules" }
func (AntigravityAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (AntigravityAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (AntigravityAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (AntigravityAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "specai.md")
}
func (AntigravityAdapter) SubAgentsDir(_ string) string       { return "" }
func (AntigravityAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (AntigravityAdapter) CommandsDir(_ string) string        { return "" }
func (AntigravityAdapter) EmbeddedCommandsDir() string        { return "" }
func (AntigravityAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills") }
func (AntigravityAdapter) SettingsPath(homeDir string) string { return filepath.Join(homeDir, ".gemini", "settings.json") }
func (AntigravityAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "mcp_config.json")
}
func (AntigravityAdapter) SupportsSubAgents() bool    { return false }
func (AntigravityAdapter) SupportsSlashCommands() bool { return false }
func (AntigravityAdapter) SupportsSkills() bool        { return true }
func (AntigravityAdapter) SupportsSystemPrompt() bool  { return true }
func (AntigravityAdapter) SupportsMCP() bool           { return true }

// OpenCodeAdapter implements IDEAdapter for OpenCode.
type OpenCodeAdapter struct{}

func (OpenCodeAdapter) Name() string             { return "OpenCode" }
func (OpenCodeAdapter) AgentID() model.AgentID   { return model.AgentOpenCode }
func (OpenCodeAdapter) AssetFolder() string      { return "opencode" }
func (OpenCodeAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode")
}
func (OpenCodeAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "rules")
}
func (OpenCodeAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "skills")
}
func (OpenCodeAdapter) LocalRulesFile() string { return "AGENTS.md" }
func (OpenCodeAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (OpenCodeAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (OpenCodeAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMergeIntoSettings }
func (OpenCodeAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
}
func (OpenCodeAdapter) SubAgentsDir(_ string) string          { return "" }
func (OpenCodeAdapter) EmbeddedSubAgentsDir() string          { return "" }
func (OpenCodeAdapter) CommandsDir(homeDir string) string     { return filepath.Join(homeDir, ".config", "opencode", "commands") }
func (OpenCodeAdapter) EmbeddedCommandsDir() string           { return "opencode/commands" }
func (OpenCodeAdapter) SkillsDir(homeDir string) string       { return filepath.Join(homeDir, ".config", "opencode", "skills") }
func (OpenCodeAdapter) SettingsPath(homeDir string) string    { return filepath.Join(homeDir, ".config", "opencode", "opencode.json") }
func (OpenCodeAdapter) MCPConfigPath(_ string, _ string) string { return "" }
func (OpenCodeAdapter) SupportsSubAgents() bool    { return false }
func (OpenCodeAdapter) SupportsSlashCommands() bool { return true }
func (OpenCodeAdapter) SupportsSkills() bool        { return true }
func (OpenCodeAdapter) SupportsSystemPrompt() bool  { return true }
func (OpenCodeAdapter) SupportsMCP() bool           { return true }

// KiroIDEAdapter implements IDEAdapter for Kiro IDE.
type KiroIDEAdapter struct{}

func (KiroIDEAdapter) Name() string             { return "Kiro IDE" }
func (KiroIDEAdapter) AgentID() model.AgentID   { return model.AgentKiroIDE }
func (KiroIDEAdapter) AssetFolder() string      { return "kiro" }
func (KiroIDEAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro")
}
func (KiroIDEAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "rules")
}
func (KiroIDEAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "skills")
}
func (KiroIDEAdapter) LocalRulesFile() string { return ".kiro-rules.md" }
func (KiroIDEAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (KiroIDEAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategySteeringFile }
func (KiroIDEAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (KiroIDEAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "steering", "specai.md")
}
func (KiroIDEAdapter) SubAgentsDir(homeDir string) string { return filepath.Join(homeDir, ".kiro", "agents") }
func (KiroIDEAdapter) EmbeddedSubAgentsDir() string       { return "kiro/agents" }
func (KiroIDEAdapter) CommandsDir(_ string) string        { return "" }
func (KiroIDEAdapter) EmbeddedCommandsDir() string        { return "" }
func (KiroIDEAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".kiro", "skills") }
func (KiroIDEAdapter) SettingsPath(_ string) string       { return "" }
func (KiroIDEAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
}
func (KiroIDEAdapter) SupportsSubAgents() bool    { return true }
func (KiroIDEAdapter) SupportsSlashCommands() bool { return false }
func (KiroIDEAdapter) SupportsSkills() bool        { return true }
func (KiroIDEAdapter) SupportsSystemPrompt() bool  { return true }
func (KiroIDEAdapter) SupportsMCP() bool           { return true }

// KimiCodeAdapter implements IDEAdapter for Kimi Code.
type KimiCodeAdapter struct{}

func (KimiCodeAdapter) Name() string             { return "Kimi Code" }
func (KimiCodeAdapter) AgentID() model.AgentID   { return model.AgentKimi }
func (KimiCodeAdapter) AssetFolder() string      { return "kimi" }
func (KimiCodeAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".kimi")
}
func (KimiCodeAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".kimi", "rules")
}
func (KimiCodeAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".kimi", "skills")
}
func (KimiCodeAdapter) LocalRulesFile() string { return "KIMI.md" }
func (KimiCodeAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (KimiCodeAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyJinjaModules }
func (KimiCodeAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (KimiCodeAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".kimi", "KIMI.md")
}
func (KimiCodeAdapter) SubAgentsDir(homeDir string) string { return filepath.Join(homeDir, ".kimi", "agents") }
func (KimiCodeAdapter) EmbeddedSubAgentsDir() string       { return "kimi/agents" }
func (KimiCodeAdapter) CommandsDir(_ string) string        { return "" }
func (KimiCodeAdapter) EmbeddedCommandsDir() string        { return "" }
func (KimiCodeAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".kimi", "skills") }
func (KimiCodeAdapter) SettingsPath(_ string) string       { return "" }
func (KimiCodeAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".kimi", "mcp.json")
}
func (KimiCodeAdapter) SupportsSubAgents() bool    { return true }
func (KimiCodeAdapter) SupportsSlashCommands() bool { return false }
func (KimiCodeAdapter) SupportsSkills() bool        { return true }
func (KimiCodeAdapter) SupportsSystemPrompt() bool  { return true }
func (KimiCodeAdapter) SupportsMCP() bool           { return true }

// QwenCodeAdapter implements IDEAdapter for Qwen Code.
type QwenCodeAdapter struct{}

func (QwenCodeAdapter) Name() string             { return "Qwen Code" }
func (QwenCodeAdapter) AgentID() model.AgentID   { return model.AgentQwenCode }
func (QwenCodeAdapter) AssetFolder() string      { return "qwen" }
func (QwenCodeAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".qwen")
}
func (QwenCodeAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".qwen", "rules")
}
func (QwenCodeAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".qwen", "skills")
}
func (QwenCodeAdapter) LocalRulesFile() string { return ".qwen-rules.md" }
func (QwenCodeAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (QwenCodeAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyAppendToFile }
func (QwenCodeAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (QwenCodeAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".qwen", "specai.md")
}
func (QwenCodeAdapter) SubAgentsDir(_ string) string       { return "" }
func (QwenCodeAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (QwenCodeAdapter) CommandsDir(_ string) string        { return "" }
func (QwenCodeAdapter) EmbeddedCommandsDir() string        { return "" }
func (QwenCodeAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".qwen", "skills") }
func (QwenCodeAdapter) SettingsPath(_ string) string       { return "" }
func (QwenCodeAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".qwen", "mcp.json")
}
func (QwenCodeAdapter) SupportsSubAgents() bool    { return false }
func (QwenCodeAdapter) SupportsSlashCommands() bool { return false }
func (QwenCodeAdapter) SupportsSkills() bool        { return true }
func (QwenCodeAdapter) SupportsSystemPrompt() bool  { return true }
func (QwenCodeAdapter) SupportsMCP() bool           { return true }

// KiloAdapter implements IDEAdapter for Kilo Code.
type KiloAdapter struct{}

func (KiloAdapter) Name() string             { return "Kilo Code" }
func (KiloAdapter) AgentID() model.AgentID   { return model.AgentKilocode }
func (KiloAdapter) AssetFolder() string      { return "kilo" }
func (KiloAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "kilo")
}
func (KiloAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "kilo", "rules")
}
func (KiloAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "kilo", "skills")
}
func (KiloAdapter) LocalRulesFile() string { return ".kilo-rules.md" }
func (KiloAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (KiloAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (KiloAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMergeIntoSettings }
func (KiloAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".config", "kilo", "AGENTS.md")
}
func (KiloAdapter) SubAgentsDir(_ string) string          { return "" }
func (KiloAdapter) EmbeddedSubAgentsDir() string          { return "" }
func (KiloAdapter) CommandsDir(_ string) string           { return "" }
func (KiloAdapter) EmbeddedCommandsDir() string           { return "" }
func (KiloAdapter) SkillsDir(homeDir string) string       { return filepath.Join(homeDir, ".config", "kilo", "skills") }
func (KiloAdapter) SettingsPath(homeDir string) string    { return filepath.Join(homeDir, ".config", "kilo", "opencode.json") }
func (KiloAdapter) MCPConfigPath(_ string, _ string) string { return "" }
func (KiloAdapter) SupportsSubAgents() bool    { return false }
func (KiloAdapter) SupportsSlashCommands() bool { return false }
func (KiloAdapter) SupportsSkills() bool        { return true }
func (KiloAdapter) SupportsSystemPrompt() bool  { return true }
func (KiloAdapter) SupportsMCP() bool           { return true }

// OpenClawAdapter implements IDEAdapter for OpenClaw (uses generic asset folder).
type OpenClawAdapter struct{}

func (OpenClawAdapter) Name() string             { return "OpenClaw" }
func (OpenClawAdapter) AgentID() model.AgentID   { return model.AgentOpenClaw }
func (OpenClawAdapter) AssetFolder() string      { return "generic" }
func (OpenClawAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".openclaw")
}
func (OpenClawAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".openclaw", "rules")
}
func (OpenClawAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".openclaw", "skills")
}
func (OpenClawAdapter) LocalRulesFile() string { return ".openclaw-rules.md" }
func (OpenClawAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (OpenClawAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (OpenClawAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (OpenClawAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".openclaw", "AGENTS.md")
}
func (OpenClawAdapter) SubAgentsDir(_ string) string       { return "" }
func (OpenClawAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (OpenClawAdapter) CommandsDir(_ string) string        { return "" }
func (OpenClawAdapter) EmbeddedCommandsDir() string        { return "" }
func (OpenClawAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".openclaw", "skills") }
func (OpenClawAdapter) SettingsPath(_ string) string       { return "" }
func (OpenClawAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".openclaw", "mcp.json")
}
func (OpenClawAdapter) SupportsSubAgents() bool    { return false }
func (OpenClawAdapter) SupportsSlashCommands() bool { return false }
func (OpenClawAdapter) SupportsSkills() bool        { return true }
func (OpenClawAdapter) SupportsSystemPrompt() bool  { return true }
func (OpenClawAdapter) SupportsMCP() bool           { return true }

// PiAdapter implements IDEAdapter for Pi (uses generic asset folder).
type PiAdapter struct{}

func (PiAdapter) Name() string             { return "Pi" }
func (PiAdapter) AgentID() model.AgentID   { return model.AgentPi }
func (PiAdapter) AssetFolder() string      { return "generic" }
func (PiAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".pi")
}
func (PiAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".pi", "rules")
}
func (PiAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".pi", "skills")
}
func (PiAdapter) LocalRulesFile() string { return ".pi-rules.md" }
func (PiAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (PiAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyInstructionsFile }
func (PiAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (PiAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".pi", "instructions.md")
}
func (PiAdapter) SubAgentsDir(_ string) string       { return "" }
func (PiAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (PiAdapter) CommandsDir(_ string) string        { return "" }
func (PiAdapter) EmbeddedCommandsDir() string        { return "" }
func (PiAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".pi", "skills") }
func (PiAdapter) SettingsPath(_ string) string       { return "" }
func (PiAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".pi", "mcp.json")
}
func (PiAdapter) SupportsSubAgents() bool    { return false }
func (PiAdapter) SupportsSlashCommands() bool { return false }
func (PiAdapter) SupportsSkills() bool        { return true }
func (PiAdapter) SupportsSystemPrompt() bool  { return true }
func (PiAdapter) SupportsMCP() bool           { return true }

// TraeAdapter implements IDEAdapter for Trae IDE (uses generic asset folder).
type TraeAdapter struct{}

func (TraeAdapter) Name() string             { return "Trae IDE" }
func (TraeAdapter) AgentID() model.AgentID   { return model.AgentTrae }
func (TraeAdapter) AssetFolder() string      { return "generic" }
func (TraeAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".trae")
}
func (TraeAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".trae", "rules")
}
func (TraeAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".trae", "skills")
}
func (TraeAdapter) LocalRulesFile() string { return ".trae-rules.md" }
func (TraeAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (TraeAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyAppendToFile }
func (TraeAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (TraeAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".trae", "specai.md")
}
func (TraeAdapter) SubAgentsDir(_ string) string       { return "" }
func (TraeAdapter) EmbeddedSubAgentsDir() string       { return "" }
func (TraeAdapter) CommandsDir(_ string) string        { return "" }
func (TraeAdapter) EmbeddedCommandsDir() string        { return "" }
func (TraeAdapter) SkillsDir(homeDir string) string    { return filepath.Join(homeDir, ".trae", "skills") }
func (TraeAdapter) SettingsPath(_ string) string       { return "" }
func (TraeAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".trae", "mcp.json")
}
func (TraeAdapter) SupportsSubAgents() bool    { return false }
func (TraeAdapter) SupportsSlashCommands() bool { return false }
func (TraeAdapter) SupportsSkills() bool        { return true }
func (TraeAdapter) SupportsSystemPrompt() bool  { return true }
func (TraeAdapter) SupportsMCP() bool           { return true }

// VSCodeCopilotAdapter implements IDEAdapter for VS Code Copilot (uses generic asset folder).
type VSCodeCopilotAdapter struct{}

func (VSCodeCopilotAdapter) Name() string             { return "VS Code Copilot" }
func (VSCodeCopilotAdapter) AgentID() model.AgentID   { return model.AgentVSCodeCopilot }
func (VSCodeCopilotAdapter) AssetFolder() string      { return "generic" }
func (VSCodeCopilotAdapter) ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".copilot")
}
func (VSCodeCopilotAdapter) GlobalRulesDir(homeDir string) string {
	return filepath.Join(homeDir, ".copilot", "rules")
}
func (VSCodeCopilotAdapter) GlobalSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".copilot", "skills")
}
func (VSCodeCopilotAdapter) LocalRulesFile() string { return ".copilot-rules.md" }
func (VSCodeCopilotAdapter) LocalSkillsDir() string { return ".agents/skills" }

func (VSCodeCopilotAdapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyInstructionsFile }
func (VSCodeCopilotAdapter) MCPStrategy() model.MCPStrategy                   { return model.StrategyMCPConfigFile }
func (VSCodeCopilotAdapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".copilot", "instructions.md")
}
func (VSCodeCopilotAdapter) SubAgentsDir(_ string) string { return "" }
func (VSCodeCopilotAdapter) EmbeddedSubAgentsDir() string { return "" }
func (VSCodeCopilotAdapter) CommandsDir(_ string) string  { return "" }
func (VSCodeCopilotAdapter) EmbeddedCommandsDir() string  { return "" }
func (VSCodeCopilotAdapter) SkillsDir(homeDir string) string { return filepath.Join(homeDir, ".copilot", "skills") }
func (VSCodeCopilotAdapter) SettingsPath(_ string) string { return "" }
func (v VSCodeCopilotAdapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(vscodeUserConfigDirFromHome(homeDir), "mcp.json")
}
func (VSCodeCopilotAdapter) SupportsSubAgents() bool    { return false }
func (VSCodeCopilotAdapter) SupportsSlashCommands() bool { return false }
func (VSCodeCopilotAdapter) SupportsSkills() bool        { return true }
func (VSCodeCopilotAdapter) SupportsSystemPrompt() bool  { return true }
func (VSCodeCopilotAdapter) SupportsMCP() bool           { return true }

// GetAdapters returns all 15 supported IDE/agent adapters.
func GetAdapters() []IDEAdapter {
	return []IDEAdapter{
		CursorAdapter{},
		WindsurfAdapter{},
		CodexAdapter{},
		ClaudeCodeAdapter{},
		GeminiCLIAdapter{},
		AntigravityAdapter{},
		OpenCodeAdapter{},
		KiroIDEAdapter{},
		KimiCodeAdapter{},
		QwenCodeAdapter{},
		KiloAdapter{},
		OpenClawAdapter{},
		PiAdapter{},
		TraeAdapter{},
		VSCodeCopilotAdapter{},
	}
}

// GetAdapterByName returns an adapter by its case-insensitive name or nil.
func GetAdapterByName(name string) IDEAdapter {
	lowerName := strings.ToLower(name)
	for _, a := range GetAdapters() {
		if strings.ToLower(a.Name()) == lowerName {
			return a
		}
	}
	return nil
}

// GetAdapterByAgentID returns an adapter by its AgentID or nil.
func GetAdapterByAgentID(id model.AgentID) IDEAdapter {
	for _, a := range GetAdapters() {
		if a.AgentID() == id {
			return a
		}
	}
	return nil
}

// DetectInstalledIDEs returns a list of installed IDE adapters by scanning the user's home directory.
func DetectInstalledIDEs() ([]IDEAdapter, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configs := ScanConfigs(homeDir)
	installedMap := make(map[string]bool)
	for _, c := range configs {
		if c.Exists {
			installedMap[c.Agent] = true
		}
	}

	var installed []IDEAdapter
	for _, a := range GetAdapters() {
		if installedMap[string(a.AgentID())] {
			installed = append(installed, a)
		}
	}

	// Fallback: If nothing is detected, return all supported ones so the user can choose
	if len(installed) == 0 {
		return GetAdapters(), nil
	}

	return installed, nil
}

// DetectedAgentIDs returns the AgentIDs of currently installed IDEs, without an error.
// If detection fails, it returns nil (empty slice). Useful for pre-populating TUI selections.
func DetectedAgentIDs() []model.AgentID {
	adapters, err := DetectInstalledIDEs()
	if err != nil {
		return nil
	}
	ids := make([]model.AgentID, len(adapters))
	for i, a := range adapters {
		ids[i] = a.AgentID()
	}
	return ids
}
