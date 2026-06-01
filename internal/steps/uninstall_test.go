package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

func TestStepRemoveGlobalRules(t *testing.T) {
	tests := []struct {
		name      string
		ideName   string
		setup     func(t *testing.T, rulesDir string) string
		wantExist bool
	}{
		{
			name:    "removes standard specai.md",
			ideName: "VSCode",
			setup: func(t *testing.T, rulesDir string) string {
				t.Helper()
				rulePath := filepath.Join(rulesDir, "specai.md")
				if err := os.MkdirAll(rulesDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(rulePath, []byte("rules content"), 0644); err != nil {
					t.Fatal(err)
				}
				return rulePath
			},
			wantExist: false,
		},
		{
			name:    "removes specai.mdc for Cursor",
			ideName: "Cursor",
			setup: func(t *testing.T, rulesDir string) string {
				t.Helper()
				rulePath := filepath.Join(rulesDir, "specai.mdc")
				if err := os.MkdirAll(rulesDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(rulePath, []byte("rules content"), 0644); err != nil {
					t.Fatal(err)
				}
				return rulePath
			},
			wantExist: false,
		},
		{
			name:    "no error if file does not exist",
			ideName: "VSCode",
			setup: func(t *testing.T, rulesDir string) string {
				return filepath.Join(rulesDir, "specai.md")
			},
			wantExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			ide := makeIDE(t, base, tc.ideName)

			rulePath := tc.setup(t, ide.rulesDir)

			ctx := &UninstallContext{
				HomeDir: base,
				IDEs:    []system.IDEAdapter{ide},
			}
			step := NewStepRemoveGlobalRules(ctx)

			if err := step.Run(); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			if _, err := os.Stat(rulePath); !os.IsNotExist(err) && !tc.wantExist {
				t.Errorf("expected %s to be removed, but it exists or stat failed: %v", rulePath, err)
			}
		})
	}
}

func TestStepRemoveGlobalSkills(t *testing.T) {
	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		t.Fatalf("WalkFiles: %v", err)
	}
	if len(embeddedSkills) == 0 {
		t.Skip("no embedded skills to test")
	}

	t.Run("removes embedded skills and empty directory", func(t *testing.T) {
		base := t.TempDir()
		ide := makeIDE(t, base, "VSCode")

		// Create skills and write them
		if err := os.MkdirAll(ide.skillsDir, 0755); err != nil {
			t.Fatal(err)
		}

		var createdFiles []string
		for _, ef := range embeddedSkills {
			dest := filepath.Join(ide.skillsDir, ef.RelPath)
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dest, ef.Content, 0644); err != nil {
				t.Fatal(err)
			}
			createdFiles = append(createdFiles, dest)
		}

		ctx := &UninstallContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
		step := NewStepRemoveGlobalSkills(ctx)

		if err := step.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}

		// Verify files are removed
		for _, f := range createdFiles {
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				t.Errorf("expected file %s to be removed", f)
			}
		}

		// Verify directory is removed if empty
		if _, err := os.Stat(ide.skillsDir); !os.IsNotExist(err) {
			t.Errorf("expected directory %s to be removed because it should be empty", ide.skillsDir)
		}
	})

	t.Run("leaves unrelated files in skills directory intact", func(t *testing.T) {
		base := t.TempDir()
		ide := makeIDE(t, base, "VSCode")

		if err := os.MkdirAll(ide.skillsDir, 0755); err != nil {
			t.Fatal(err)
		}

		for _, ef := range embeddedSkills {
			dest := filepath.Join(ide.skillsDir, ef.RelPath)
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dest, ef.Content, 0644); err != nil {
				t.Fatal(err)
			}
		}

		// Create a file not managed by SpecAI
		unrelatedPath := filepath.Join(ide.skillsDir, "user-skill.md")
		if err := os.WriteFile(unrelatedPath, []byte("user content"), 0644); err != nil {
			t.Fatal(err)
		}

		ctx := &UninstallContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
		step := NewStepRemoveGlobalSkills(ctx)

		if err := step.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}

		// Verify unrelated file still exists
		if _, err := os.Stat(unrelatedPath); err != nil {
			t.Errorf("expected unrelated file %s to still exist: %v", unrelatedPath, err)
		}

		// Verify directory still exists because it's not empty
		if _, err := os.Stat(ide.skillsDir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to still exist because it has a user file", ide.skillsDir)
		}
	})

	t.Run("no error if skills directory doesn't exist", func(t *testing.T) {
		base := t.TempDir()
		ide := makeIDE(t, base, "VSCode")

		ctx := &UninstallContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
		step := NewStepRemoveGlobalSkills(ctx)

		if err := step.Run(); err != nil {
			t.Fatalf("expected nil error when skills directory doesn't exist, got %v", err)
		}
	})
}

func TestStepSnapshotBeforeUninstall(t *testing.T) {
	base := t.TempDir()
	ide := makeIDE(t, base, "VSCode")

	// Pre-create rules file
	rulePath := filepath.Join(ide.rulesDir, "specai.md")
	if err := os.MkdirAll(ide.rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rulePath, []byte("rules"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &UninstallContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
	step := NewStepSnapshotBeforeUninstall(ctx)

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	backupsRoot := filepath.Join(base, ".specai", "backups")
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatalf("backup root not created: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one snapshot directory, got none")
	}
}

func TestStepScanIDEsForUninstall(t *testing.T) {
	ctx := &UninstallContext{HomeDir: t.TempDir()}
	step := NewStepScanIDEsForUninstall(ctx)

	if got := step.ID(); got == "" {
		t.Error("ID() should not be empty")
	}
}
