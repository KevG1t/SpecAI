package steps

import (
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/spf13/afero"
)

// pluginFiles lists the OpenCode plugin files embedded in the assets FS.
var pluginFiles = []string{
	"background-agents.ts",
	"model-variants.ts",
}

// StepInjectOpenCodePlugins copies embedded OpenCode plugin files to the user's
// OpenCode plugins directory (~/.config/opencode/plugins/).
type StepInjectOpenCodePlugins struct {
	ctx *InstallContext
	fs  afero.Fs
}

// NewStepInjectOpenCodePlugins constructs a StepInjectOpenCodePlugins.
func NewStepInjectOpenCodePlugins(ctx *InstallContext) *StepInjectOpenCodePlugins {
	return &StepInjectOpenCodePlugins{ctx: ctx}
}

func (s *StepInjectOpenCodePlugins) ID() string {
	return "Injecting OpenCode plugins"
}

func (s *StepInjectOpenCodePlugins) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

// Run copies OpenCode plugin files to ~/.config/opencode/plugins/ when
// AgentOpenCode is in the selected IDEs.
func (s *StepInjectOpenCodePlugins) Run() error {
	hasOpenCode := false
	for _, ide := range s.ctx.IDEs {
		if ide.AgentID() == model.AgentOpenCode {
			hasOpenCode = true
			break
		}
	}
	if !hasOpenCode {
		return nil
	}

	destDir := filepath.Join(s.ctx.HomeDir, ".config", "opencode", "plugins")
	if err := s.filesystem().MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, fileName := range pluginFiles {
		srcPath := "opencode/plugins/" + fileName
		data, err := assets.FS.ReadFile(srcPath)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, fileName)
		if err := writeFileAtomic(s.filesystem(), destPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
