package kiro

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
)

type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) (os.FileInfo, error)
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: exec.LookPath,
		statPath: os.Stat,
	}
}

func (a *Adapter) Agent() model.AgentID {
	return model.AgentKiroIDE
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	configPath := filepath.Join(homeDir, ".kiro")

	binaryPath, err := a.lookPath("kiro")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, "", configPath, false, nil
		}
		return false, "", configPath, false, err
	}

	info, statErr := a.statPath(configPath)
	configFound := statErr == nil && info.IsDir()

	return true, binaryPath, configPath, configFound, nil
}

func (a *Adapter) SupportsAutoInstall() bool {
	return false
}

func (a *Adapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) {
	return nil, AgentNotInstallableError{Agent: model.AgentKiroIDE}
}

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return a.kiroConfigDir(homeDir)
}

func (a *Adapter) SystemPromptDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "steering")
}

func (a *Adapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(a.SystemPromptDir(homeDir), "specai.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(a.kiroConfigDir(homeDir), "settings.json")
}

func (a *Adapter) SupportsSubAgents() bool {
	return true
}

func (a *Adapter) SubAgentsDir(homeDir string) string {
	return filepath.Join(homeDir, ".kiro", "agents")
}

func (a *Adapter) EmbeddedSubAgentsDir() string {
	return "kiro/agents"
}

func (a *Adapter) KiroModelID(alias model.ClaudeModelAlias) string {
	return model.KiroModelID(alias)
}

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategySteeringFile
}

func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMCPConfigFile
}

func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
}

func (a *Adapter) kiroConfigDir(homeDir string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Kiro", "User")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "kiro", "User")
	default:
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			xdgConfigHome = filepath.Join(homeDir, ".config")
		}
		return filepath.Join(xdgConfigHome, "kiro", "user")
	}
}

func (a *Adapter) SupportsOutputStyles() bool     { return false }
func (a *Adapter) OutputStyleDir(_ string) string { return "" }
func (a *Adapter) SupportsSlashCommands() bool    { return false }
func (a *Adapter) CommandsDir(_ string) string    { return "" }
func (a *Adapter) SupportsSkills() bool           { return true }
func (a *Adapter) SupportsSystemPrompt() bool     { return true }
func (a *Adapter) SupportsMCP() bool              { return true }
