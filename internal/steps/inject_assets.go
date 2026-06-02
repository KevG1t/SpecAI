package steps

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/spf13/afero"
)

type AssetInjector interface {
	// InjectAgentFolder copies {agentFolder}/ into targetDir.
	// If the folder does not exist in the embed.FS, it returns nil (graceful skip).
	InjectAgentFolder(agentFolder, targetDir string) error

	// InjectSharedSkills copies skills/ into targetDir.
	InjectSharedSkills(targetDir string) error
}

type assetInjector struct {
	fs afero.Fs
}

func NewAssetInjector(fsys afero.Fs) AssetInjector {
	if fsys == nil {
		fsys = afero.NewOsFs()
	}
	return &assetInjector{fs: fsys}
}

// InjectAgentFolder copies {agentFolder}/* into targetDir.
// If the embedded folder does not exist, returns nil (graceful skip).
func (a *assetInjector) InjectAgentFolder(agentFolder, targetDir string) error {
	return a.walkAndCopy(agentFolder, targetDir)
}

// InjectSharedSkills copies all embedded skills content into targetDir.
// The embedded FS root contains skills subdirectories directly.
func (a *assetInjector) InjectSharedSkills(targetDir string) error {
	return a.walkAndCopy("skills", targetDir)
}

// walkAndCopy recursively copies srcDir from the embedded FS into targetDir on the
// virtual filesystem. If srcDir does not exist in the embedded FS, returns nil.
func (a *assetInjector) walkAndCopy(srcDir, targetDir string) error {
	// Verify the source directory exists in the embed.FS before walking.
	if _, err := assets.FS.ReadDir(srcDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // graceful skip
		}
		return err
	}

	if err := a.fs.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	return fs.WalkDir(assets.FS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}

		relPath := strings.TrimPrefix(path, srcDir+"/")
		if relPath == "" {
			return nil
		}

		dstPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return a.fs.MkdirAll(dstPath, 0755)
		}

		data, err := assets.FS.ReadFile(path)
		if err != nil {
			return err
		}

		mode := fs.FileMode(0644)
		if info, err2 := d.Info(); err2 == nil {
			mode = info.Mode()
		}

		return afero.WriteFile(a.fs, dstPath, data, mode)
	})
}

// StepInjectAssets wraps AssetInjector to be a pipeline.Step for InstallContext.
// For each IDE in ctx.IDEs, it injects {adapter.AssetFolder()}/ into the
// adapter's GlobalSkillsDir. Then it injects skills/ exactly once into a
// shared location.
type StepInjectAssets struct {
	ctx      *InstallContext
	injector AssetInjector
	plan     planner.ResolvedPlan
}

func NewStepInjectAssets(ctx *InstallContext, injector AssetInjector, plan planner.ResolvedPlan) *StepInjectAssets {
	return &StepInjectAssets{
		ctx:      ctx,
		injector: injector,
		plan:     plan,
	}
}

func (s *StepInjectAssets) ID() string {
	return "Inyectando componentes (Skills, Persona, config)"
}

func (s *StepInjectAssets) Run() error {
	// Per-agent injection: copy {agentFolder}/ to each adapter's GlobalSkillsDir.
	for _, ide := range s.ctx.IDEs {
		targetDir := ide.GlobalSkillsDir(s.ctx.HomeDir)
		if err := s.injector.InjectAgentFolder(ide.AssetFolder(), targetDir); err != nil {
			return err
		}
	}

	// Shared skills injection: copy skills/ exactly once to a shared location.
	sharedSkillsDir := filepath.Join(s.ctx.HomeDir, ".specai", "skills")
	return s.injector.InjectSharedSkills(sharedSkillsDir)
}
