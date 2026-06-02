package steps

import (
	"io/fs"
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/spf13/afero"
)

// StepInjectSlashCommands copies SDD slash command files from embedded assets to
// the IDE-specific commands directories:
//
//   - Claude Code: assets/claude/commands/ → ~/.claude/commands/
//   - OpenCode:    assets/opencode/commands/ → ~/.config/opencode/commands/
//
// Existing files at the destination are overwritten with the latest version.
type StepInjectSlashCommands struct {
	ctx *InstallContext
	fs  afero.Fs
}

func NewStepInjectSlashCommands(ctx *InstallContext) *StepInjectSlashCommands {
	return &StepInjectSlashCommands{ctx: ctx}
}

func (s *StepInjectSlashCommands) ID() string {
	return "Inyectando slash commands SDD"
}

func (s *StepInjectSlashCommands) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

func (s *StepInjectSlashCommands) Run() error {
	for _, ide := range s.ctx.IDEs {
		if !ide.SupportsSlashCommands() {
			continue
		}
		srcDir := ide.EmbeddedCommandsDir()
		destDir := ide.CommandsDir(s.ctx.HomeDir)
		if err := s.copyCommands(srcDir, destDir); err != nil {
			return err
		}
	}
	return nil
}

func (s *StepInjectSlashCommands) copyCommands(assetDir, destDir string) error {
	fsys := s.filesystem()
	if err := fsys.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	return fs.WalkDir(assets.FS, assetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := assets.FS.ReadFile(path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, filepath.Base(path))
		return afero.WriteFile(fsys, destPath, data, 0644)
	})
}
