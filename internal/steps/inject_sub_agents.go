package steps

import (
	"io/fs"
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/spf13/afero"
)

// StepInjectSubAgents copies embedded sub-agent definition files to the
// user's agent directories for supported agents (Claude Code and Kiro).
type StepInjectSubAgents struct {
	ctx *InstallContext
	fs  afero.Fs
}

// NewStepInjectSubAgents constructs a StepInjectSubAgents.
func NewStepInjectSubAgents(ctx *InstallContext) *StepInjectSubAgents {
	return &StepInjectSubAgents{ctx: ctx}
}

func (s *StepInjectSubAgents) ID() string {
	return "Injecting sub-agent definitions"
}

func (s *StepInjectSubAgents) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

// Run copies agent files from embedded FS to user directories based on selected IDEs.
func (s *StepInjectSubAgents) Run() error {
	for _, ide := range s.ctx.IDEs {
		if !ide.SupportsSubAgents() {
			continue
		}
		srcDir := ide.EmbeddedSubAgentsDir()
		destDir := ide.SubAgentsDir(s.ctx.HomeDir)
		if err := s.copyAgentFiles(srcDir, destDir); err != nil {
			return err
		}
	}
	return nil
}

// copyAgentFiles walks the embedded srcDir and writes each .md file to destDir.
func (s *StepInjectSubAgents) copyAgentFiles(srcDir, destDir string) error {
	if err := s.filesystem().MkdirAll(destDir, 0755); err != nil {
		return err
	}

	return fs.WalkDir(assets.FS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, readErr := assets.FS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		destPath := filepath.Join(destDir, filepath.Base(path))
		return writeFileAtomic(s.filesystem(), destPath, data, 0644)
	})
}
