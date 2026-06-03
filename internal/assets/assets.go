package assets

import (
	"embed"

	"github.com/KevG1t/SpecAI/internal/model"
)

//go:embed all:claude all:opencode all:generic all:skills all:gemini all:codex all:antigravity all:windsurf all:cursor all:kimi all:qwen all:kiro
var FS embed.FS

// MustRead returns the content of an embedded file or panics.
func MustRead(path string) string {
	data, err := FS.ReadFile(path)
	if err != nil {
		panic("assets: " + err.Error())
	}
	return string(data)
}

// Read returns the content of an embedded file.
func Read(path string) (string, error) {
	data, err := FS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SDDCommandsAssetDir returns the embedded slash-command asset directory for an
// agent. Claude uses Claude-native frontmatter under claude/commands; agents
// without a dedicated command set fall back to the OpenCode-compatible assets.
func SDDCommandsAssetDir(agent model.AgentID) string {
	switch agent {
	case model.AgentClaudeCode:
		return "claude/commands"
	default:
		return "opencode/commands"
	}
}
