package steps

import (
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

func newOpenCodePluginsStep(homeDir string, ids ...model.AgentID) *StepInjectOpenCodePlugins {
	var ides []system.IDEAdapter
	for _, id := range ids {
		ides = append(ides, subAgentsStubIDE{agentID: id})
	}
	step := NewStepInjectOpenCodePlugins(&InstallContext{HomeDir: homeDir, IDEs: ides})
	step.fs = afero.NewMemMapFs()
	return step
}

// TestInjectOpenCodePlugins_WritesPluginFiles verifies both plugin files are deployed
// to ~/.config/opencode/plugins/ when OpenCode is selected.
func TestInjectOpenCodePlugins_WritesPluginFiles(t *testing.T) {
	step := newOpenCodePluginsStep("/home/user", model.AgentOpenCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	pluginDir := "/home/user/.config/opencode/plugins"
	for _, file := range []string{"background-agents.ts", "model-variants.ts"} {
		path := filepath.Join(pluginDir, file)
		if _, err := step.fs.Stat(path); err != nil {
			t.Errorf("expected plugin file %q to exist: %v", path, err)
		}
	}
}

// TestInjectOpenCodePlugins_CreatesDestDir verifies the plugin directory is created
// when it doesn't exist.
func TestInjectOpenCodePlugins_CreatesDestDir(t *testing.T) {
	step := newOpenCodePluginsStep("/home/newuser", model.AgentOpenCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	pluginDir := "/home/newuser/.config/opencode/plugins"
	if _, err := step.fs.Stat(pluginDir); err != nil {
		t.Errorf("expected directory %q to be created: %v", pluginDir, err)
	}
}

// TestInjectOpenCodePlugins_NoOp_WhenOpenCodeNotSelected verifies no files are written
// when OpenCode is not in the selected agents.
func TestInjectOpenCodePlugins_NoOp_WhenOpenCodeNotSelected(t *testing.T) {
	step := newOpenCodePluginsStep("/home/user", model.AgentClaudeCode)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	pluginDir := "/home/user/.config/opencode/plugins"
	if _, err := step.fs.Stat(pluginDir); err == nil {
		t.Error("expected no ~/.config/opencode/plugins when OpenCode not selected, but it exists")
	}
}
