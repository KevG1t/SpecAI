package sddmemory

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
)

const (
	SetupModeEnvVar   = "SPECAI_SDD_MEMORY_SETUP_MODE"
	SetupStrictEnvVar = "SPECAI_SDD_MEMORY_SETUP_STRICT"
)

type SetupMode string

const (
	SetupModeOff       SetupMode = "off"
	SetupModeOpenCode  SetupMode = "opencode"
	SetupModeSupported SetupMode = "supported"
)

func ParseSetupMode(value string) SetupMode {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case string(SetupModeOff):
		return SetupModeOff
	case string(SetupModeOpenCode):
		return SetupModeOpenCode
	case "", string(SetupModeSupported):
		return SetupModeSupported
	default:
		return SetupModeSupported
	}
}

func ParseSetupStrict(value string) bool {
	v := strings.TrimSpace(strings.ToLower(value))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func SetupAgentSlug(agent model.AgentID) (string, bool) {
	switch agent {
	case model.AgentOpenCode:
		return "opencode", true
	case model.AgentKilocode:
		return "kilocode", true
	case model.AgentClaudeCode:
		return "claude-code", true
	case model.AgentGeminiCLI:
		return "gemini-cli", true
	case model.AgentCodex:
		return "codex", true
	case model.AgentAntigravity:
		return "gemini-cli", true
	case model.AgentWindsurf:
		return "windsurf", true
	case model.AgentCursor, model.AgentVSCodeCopilot:
		return "", false
	case model.AgentQwenCode:
		return "", false
	default:
		return "", false
	}
}

func ShouldAttemptSetup(mode SetupMode, agent model.AgentID) bool {
	slug, ok := SetupAgentSlug(agent)
	if !ok {
		return false
	}

	switch mode {
	case SetupModeOff:
		return false
	case SetupModeSupported:
		return true
	case SetupModeOpenCode:
		return slug == "opencode"
	default:
		return slug == "opencode"
	}
}
