package assets

import (
	"embed"
	"strings"
)

//go:embed all:claude all:opencode all:generic all:skills all:gemini all:codex all:antigravity all:windsurf all:cursor all:kimi all:qwen all:kiro
var FS embed.FS

// MustRead returns the content of an embedded file or panics.
func MustRead(path string) string {
	data, err := FS.ReadFile(path)
	if err != nil {
		panic("assets: " + err.Error())
	}
	return normalizeText(string(data))
}

// Read returns the content of an embedded file.
func Read(path string) (string, error) {
	data, err := FS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizeText(string(data)), nil
}

// normalizeText canonicalizes embedded text assets to LF line endings so
// generated prompts, markdown sections, and golden outputs stay deterministic
// across platforms.
func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}
