package steps

import (
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/spf13/afero"
)

// StepInjectSDDMemory injects the sdd-memory-protocol.md content as a
// marker-bounded section into the Claude Code CLAUDE.md system prompt file.
//
// Only Claude Code is supported — other IDEs do not use CLAUDE.md.
// The injection is idempotent: running it twice produces the same result.
type StepInjectSDDMemory struct {
	ctx *InstallContext
	fs  afero.Fs
}

func NewStepInjectSDDMemory(ctx *InstallContext) *StepInjectSDDMemory {
	return &StepInjectSDDMemory{ctx: ctx}
}

func (s *StepInjectSDDMemory) ID() string {
	return "Inyectando protocolo sdd-memory en CLAUDE.md"
}

func (s *StepInjectSDDMemory) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

// Run injects the sdd-memory-protocol into ~/.claude/CLAUDE.md for every
// Claude Code agent in the install context.
func (s *StepInjectSDDMemory) Run() error {
	content, err := assets.FS.ReadFile("claude/sdd-memory-protocol.md")
	if err != nil {
		return err
	}

	for _, ide := range s.ctx.IDEs {
		if ide.AgentID() != model.AgentClaudeCode {
			continue
		}
		claudeMDPath := filepath.Join(s.ctx.HomeDir, ".claude", "CLAUDE.md")
		if err := injectMarkdownSection(s.filesystem(), claudeMDPath, "sdd-memory-protocol", string(content)); err != nil {
			return err
		}
	}
	return nil
}
