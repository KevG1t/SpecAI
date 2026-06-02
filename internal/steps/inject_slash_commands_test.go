package steps

import (
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/spf13/afero"
)

func TestStepInjectSlashCommands_Run(t *testing.T) {
	tests := []struct {
		name            string
		ides            []model.AgentID
		wantClaudeFiles bool
		wantOpenFiles   bool
	}{
		{
			name:            "claude code gets commands",
			ides:            []model.AgentID{model.AgentClaudeCode},
			wantClaudeFiles: true,
			wantOpenFiles:   false,
		},
		{
			name:            "opencode gets commands",
			ides:            []model.AgentID{model.AgentOpenCode},
			wantClaudeFiles: false,
			wantOpenFiles:   true,
		},
		{
			name:            "both IDEs get commands",
			ides:            []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
			wantClaudeFiles: true,
			wantOpenFiles:   true,
		},
		{
			name:            "cursor is skipped",
			ides:            []model.AgentID{model.AgentCursor},
			wantClaudeFiles: false,
			wantOpenFiles:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := afero.NewMemMapFs()
			homeDir := "/home/testuser"

			adapters := makeAdapters(tt.ides)
			ctx := &InstallContext{IDEs: adapters, HomeDir: homeDir}
			step := &StepInjectSlashCommands{ctx: ctx, fs: fsys}

			if err := step.Run(); err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			claudeCommandsDir := filepath.Join(homeDir, ".claude", "commands")
			openCodeCommandsDir := filepath.Join(homeDir, ".config", "opencode", "commands")

			if tt.wantClaudeFiles {
				assertDirHasFiles(t, fsys, claudeCommandsDir)
			} else {
				assertDirAbsent(t, fsys, claudeCommandsDir)
			}

			if tt.wantOpenFiles {
				assertDirHasFiles(t, fsys, openCodeCommandsDir)
			} else {
				assertDirAbsent(t, fsys, openCodeCommandsDir)
			}
		})
	}
}

func TestStepInjectSlashCommands_OverwritesExisting(t *testing.T) {
	fsys := afero.NewMemMapFs()
	homeDir := "/home/testuser"
	claudeCommandsDir := filepath.Join(homeDir, ".claude", "commands")

	// Pre-create a stale file.
	if err := fsys.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	staleContent := []byte("outdated content that must be overwritten")
	if err := afero.WriteFile(fsys, filepath.Join(claudeCommandsDir, "sdd-init.md"), staleContent, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	adapters := makeAdapters([]model.AgentID{model.AgentClaudeCode})
	ctx := &InstallContext{IDEs: adapters, HomeDir: homeDir}
	step := &StepInjectSlashCommands{ctx: ctx, fs: fsys}

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	result, err := afero.ReadFile(fsys, filepath.Join(claudeCommandsDir, "sdd-init.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(result) == string(staleContent) {
		t.Error("file was not overwritten — still has stale content")
	}
}

func assertDirHasFiles(t *testing.T, fsys afero.Fs, dir string) {
	t.Helper()
	infos, err := afero.ReadDir(fsys, dir)
	if err != nil {
		t.Errorf("expected directory %q to exist with files, got error: %v", dir, err)
		return
	}
	if len(infos) == 0 {
		t.Errorf("expected directory %q to have files, but it is empty", dir)
	}
}

func assertDirAbsent(t *testing.T, fsys afero.Fs, dir string) {
	t.Helper()
	infos, _ := afero.ReadDir(fsys, dir)
	if len(infos) > 0 {
		t.Errorf("expected directory %q to be absent or empty, but it has %d files", dir, len(infos))
	}
}
