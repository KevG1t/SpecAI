package steps

import (
	"os"
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/components/filemerge"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/spf13/afero"
)

// StepInjectOpenCodeOverlay deep-merges the SDD overlay JSON into
// ~/.config/opencode/opencode.json. The merge is additive — user-defined keys
// are preserved. Running it multiple times is idempotent.
type StepInjectOpenCodeOverlay struct {
	ctx *InstallContext
	fs  afero.Fs
}

func NewStepInjectOpenCodeOverlay(ctx *InstallContext) *StepInjectOpenCodeOverlay {
	return &StepInjectOpenCodeOverlay{ctx: ctx}
}

func (s *StepInjectOpenCodeOverlay) ID() string {
	return "Inyectando overlay SDD en opencode.json"
}

func (s *StepInjectOpenCodeOverlay) filesystem() afero.Fs {
	if s.fs != nil {
		return s.fs
	}
	return afero.NewOsFs()
}

// Run merges sdd-overlay-single.json into ~/.config/opencode/opencode.json.
// Only runs when OpenCode is among the selected IDEs.
func (s *StepInjectOpenCodeOverlay) Run() error {
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

	overlayBytes, err := assets.FS.ReadFile("opencode/sdd-overlay-single.json")
	if err != nil {
		return err
	}

	configPath := filepath.Join(s.ctx.HomeDir, ".config", "opencode", "opencode.json")
	fsys := s.filesystem()

	if err := fsys.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	existing, err := afero.ReadFile(fsys, configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	merged, err := filemerge.MergeJSONObjects(existing, overlayBytes)
	if err != nil {
		return err
	}

	return writeFileAtomic(fsys, configPath, merged, 0644)
}
