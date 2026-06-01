package system

import (
	"os"
	"path/filepath"
	"strings"
)

// IDEAdapter models paths and characteristics for a specific IDE/agent globally and locally.
type IDEAdapter interface {
	Name() string
	ConfigDir(homeDir string) string
	GlobalRulesDir(homeDir string) string
	GlobalSkillsDir(homeDir string) string
	LocalRulesFile() string // relative rules file path in current directory, e.g. ".cursorrules"
	LocalSkillsDir() string // relative skills directory path in current directory, e.g. ".agents/skills"
}

// CursorAdapter implements IDEAdapter for Cursor.
type CursorAdapter struct{}

func (CursorAdapter) Name() string { return "Cursor" }
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

func (WindsurfAdapter) Name() string { return "Windsurf" }
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

func (CodexAdapter) Name() string { return "Codex" }
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

// GetAdapters returns all supported IDE adapters.
func GetAdapters() []IDEAdapter {
	return []IDEAdapter{
		CursorAdapter{},
		WindsurfAdapter{},
		CodexAdapter{},
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
		// Map adapter names to the agent IDs in knownAgentConfigDirs
		agentID := strings.ToLower(a.Name())
		if a.Name() == "Windsurf" {
			agentID = "windsurf"
		}
		if installedMap[agentID] {
			installed = append(installed, a)
		}
	}

	// Fallback: If nothing is detected, return all supported ones so the user can choose
	if len(installed) == 0 {
		return GetAdapters(), nil
	}

	return installed, nil
}
