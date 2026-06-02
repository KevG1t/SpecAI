package system

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
)

// IDEAdapter models paths and characteristics for a specific IDE/agent globally and locally.
type IDEAdapter interface {
	Name() string
	AgentID() model.AgentID
	AssetFolder() string
	ConfigDir(homeDir string) string
	GlobalRulesDir(homeDir string) string
	GlobalSkillsDir(homeDir string) string
	LocalRulesFile() string // relative rules file path in current directory, e.g. ".cursorrules"
	LocalSkillsDir() string // relative skills directory path in current directory, e.g. ".agents/skills"
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
