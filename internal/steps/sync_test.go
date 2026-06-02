package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

// stubIDE is a minimal IDEAdapter for testing.
// New interface methods delegate to the real adapter (via GetAdapterByAgentID) when available.
type stubIDE struct {
	name        string
	rulesDir    string
	skillsDir   string
	agentID     model.AgentID
	assetFolder string
}

func (s stubIDE) Name() string                    { return s.name }
func (s stubIDE) AgentID() model.AgentID          { return s.agentID }
func (s stubIDE) AssetFolder() string             { return s.assetFolder }
func (s stubIDE) ConfigDir(_ string) string       { return "" }
func (s stubIDE) GlobalRulesDir(_ string) string  { return s.rulesDir }
func (s stubIDE) GlobalSkillsDir(_ string) string { return s.skillsDir }
func (s stubIDE) LocalRulesFile() string          { return "" }
func (s stubIDE) LocalSkillsDir() string          { return "" }

func (s stubIDE) real() system.IDEAdapter {
	if a := system.GetAdapterByAgentID(s.agentID); a != nil {
		return a
	}
	return nil
}

func (s stubIDE) SystemPromptStrategy() model.SystemPromptStrategy {
	if r := s.real(); r != nil {
		return r.SystemPromptStrategy()
	}
	return model.StrategyFileReplace
}
func (s stubIDE) MCPStrategy() model.MCPStrategy {
	if r := s.real(); r != nil {
		return r.MCPStrategy()
	}
	return model.StrategyMCPConfigFile
}
func (s stubIDE) SystemPromptFile(homeDir string) string {
	if r := s.real(); r != nil {
		return r.SystemPromptFile(homeDir)
	}
	return ""
}
func (s stubIDE) SubAgentsDir(homeDir string) string {
	if r := s.real(); r != nil {
		return r.SubAgentsDir(homeDir)
	}
	return ""
}
func (s stubIDE) EmbeddedSubAgentsDir() string {
	if r := s.real(); r != nil {
		return r.EmbeddedSubAgentsDir()
	}
	return ""
}
func (s stubIDE) CommandsDir(homeDir string) string {
	if r := s.real(); r != nil {
		return r.CommandsDir(homeDir)
	}
	return ""
}
func (s stubIDE) EmbeddedCommandsDir() string {
	if r := s.real(); r != nil {
		return r.EmbeddedCommandsDir()
	}
	return ""
}
func (s stubIDE) SkillsDir(homeDir string) string {
	if r := s.real(); r != nil {
		return r.SkillsDir(homeDir)
	}
	return ""
}
func (s stubIDE) SettingsPath(homeDir string) string {
	if r := s.real(); r != nil {
		return r.SettingsPath(homeDir)
	}
	return ""
}
func (s stubIDE) MCPConfigPath(homeDir string, serverName string) string {
	if r := s.real(); r != nil {
		return r.MCPConfigPath(homeDir, serverName)
	}
	return ""
}
func (s stubIDE) SupportsSubAgents() bool {
	if r := s.real(); r != nil {
		return r.SupportsSubAgents()
	}
	return false
}
func (s stubIDE) SupportsSlashCommands() bool {
	if r := s.real(); r != nil {
		return r.SupportsSlashCommands()
	}
	return false
}
func (s stubIDE) SupportsSkills() bool {
	if r := s.real(); r != nil {
		return r.SupportsSkills()
	}
	return true
}
func (s stubIDE) SupportsSystemPrompt() bool {
	if r := s.real(); r != nil {
		return r.SupportsSystemPrompt()
	}
	return true
}
func (s stubIDE) SupportsMCP() bool {
	if r := s.real(); r != nil {
		return r.SupportsMCP()
	}
	return true
}

var _ system.IDEAdapter = stubIDE{}

// makeIDE creates a stubIDE with directories inside base.
func makeIDE(t *testing.T, base, name string) stubIDE {
	t.Helper()
	rulesDir := filepath.Join(base, name, "rules")
	skillsDir := filepath.Join(base, name, "skills")
	return stubIDE{name: name, rulesDir: rulesDir, skillsDir: skillsDir}
}

// ────────────────────────── StepSyncGlobalRules ──────────────────────────

func TestStepSyncGlobalRules(t *testing.T) {
	embedded := []byte(templates.MustRead("base/persona.md"))

	tests := []struct {
		name          string
		setupFile     func(t *testing.T, rulePath string)
		wantChanged   int
		wantSkipped   int
		wantFileOnDisk bool
		wantErr       bool
	}{
		{
			name: "file already up to date skips write",
			setupFile: func(t *testing.T, rulePath string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(rulePath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(rulePath, embedded, 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantChanged:    0,
			wantSkipped:    1,
			wantFileOnDisk: true,
		},
		{
			name: "outdated file gets rewritten",
			setupFile: func(t *testing.T, rulePath string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(rulePath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(rulePath, []byte("old content"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantChanged:    1,
			wantSkipped:    0,
			wantFileOnDisk: true,
		},
		{
			name: "missing file gets written",
			setupFile:      nil, // no setup, file absent
			wantChanged:    1,
			wantSkipped:    0,
			wantFileOnDisk: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			ide := makeIDE(t, base, "VSCode")
			rulePath := filepath.Join(ide.rulesDir, "specai.md")

			if tc.setupFile != nil {
				tc.setupFile(t, rulePath)
			}

			ctx := &SyncContext{
				HomeDir: base,
				IDEs:    []system.IDEAdapter{ide},
			}
			step := NewStepSyncGlobalRules(ctx)

			err := step.Run()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Run() error = %v, wantErr %v", err, tc.wantErr)
			}
			if ctx.Changed != tc.wantChanged {
				t.Errorf("Changed = %d, want %d", ctx.Changed, tc.wantChanged)
			}
			if ctx.Skipped != tc.wantSkipped {
				t.Errorf("Skipped = %d, want %d", ctx.Skipped, tc.wantSkipped)
			}
			if tc.wantFileOnDisk {
				got, err := os.ReadFile(rulePath)
				if err != nil {
					t.Fatalf("file not on disk: %v", err)
				}
				if string(got) != string(embedded) {
					t.Errorf("file content mismatch: got %q, want embedded persona", string(got))
				}
			}
		})
	}
}

func TestStepSyncGlobalRules_CursorUsesMDC(t *testing.T) {
	base := t.TempDir()
	ide := makeIDE(t, base, "Cursor")
	embedded := []byte(templates.MustRead("base/persona.md"))

	ctx := &SyncContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
	step := NewStepSyncGlobalRules(ctx)

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	mdcPath := filepath.Join(ide.rulesDir, "specai.mdc")
	got, err := os.ReadFile(mdcPath)
	if err != nil {
		t.Fatalf("specai.mdc not created: %v", err)
	}
	if string(got) != string(embedded) {
		t.Error("specai.mdc content does not match embedded persona")
	}
	if ctx.Changed != 1 {
		t.Errorf("Changed = %d, want 1", ctx.Changed)
	}
}

func TestStepSyncGlobalRules_NoIDEs(t *testing.T) {
	ctx := &SyncContext{HomeDir: t.TempDir(), IDEs: nil}
	step := NewStepSyncGlobalRules(ctx)
	if err := step.Run(); err == nil {
		t.Fatal("expected error with no IDEs, got nil")
	}
}

// ────────────────────────── StepSyncGlobalSkills ─────────────────────────

func TestStepSyncGlobalSkills(t *testing.T) {
	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		t.Fatalf("WalkFiles: %v", err)
	}
	if len(embeddedSkills) == 0 {
		t.Skip("no embedded skills to test")
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T, skillsDir string)
		wantChanged int
		wantSkipped int
	}{
		{
			name: "all skills already up to date",
			setup: func(t *testing.T, skillsDir string) {
				t.Helper()
				for _, ef := range embeddedSkills {
					dest := filepath.Join(skillsDir, ef.RelPath)
					if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(dest, ef.Content, 0644); err != nil {
						t.Fatal(err)
					}
				}
			},
			wantChanged: 0,
			wantSkipped: len(embeddedSkills),
		},
		{
			name:        "all skills missing",
			setup:       nil,
			wantChanged: len(embeddedSkills),
			wantSkipped: 0,
		},
		{
			name: "one skill outdated",
			setup: func(t *testing.T, skillsDir string) {
				t.Helper()
				// Write all up-to-date, then corrupt the first one.
				for _, ef := range embeddedSkills {
					dest := filepath.Join(skillsDir, ef.RelPath)
					if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(dest, ef.Content, 0644); err != nil {
						t.Fatal(err)
					}
				}
				first := filepath.Join(skillsDir, embeddedSkills[0].RelPath)
				if err := os.WriteFile(first, []byte("outdated"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantChanged: 1,
			wantSkipped: len(embeddedSkills) - 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			ide := makeIDE(t, base, "VSCode")

			if tc.setup != nil {
				if err := os.MkdirAll(ide.skillsDir, 0755); err != nil {
					t.Fatal(err)
				}
				tc.setup(t, ide.skillsDir)
			}

			ctx := &SyncContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
			step := NewStepSyncGlobalSkills(ctx)

			if err := step.Run(); err != nil {
				t.Fatalf("Run() error: %v", err)
			}
			if ctx.Changed != tc.wantChanged {
				t.Errorf("Changed = %d, want %d", ctx.Changed, tc.wantChanged)
			}
			if ctx.Skipped != tc.wantSkipped {
				t.Errorf("Skipped = %d, want %d", ctx.Skipped, tc.wantSkipped)
			}
		})
	}
}

func TestStepSyncGlobalSkills_NoIDEs(t *testing.T) {
	ctx := &SyncContext{HomeDir: t.TempDir(), IDEs: nil}
	step := NewStepSyncGlobalSkills(ctx)
	if err := step.Run(); err == nil {
		t.Fatal("expected error with no IDEs, got nil")
	}
}

// ────────────────────────── StepSnapshotBeforeSync ───────────────────────

func TestStepSnapshotBeforeSync_NoIDEsDoesNotPanic(t *testing.T) {
	ctx := &SyncContext{HomeDir: t.TempDir(), IDEs: nil}
	step := NewStepSnapshotBeforeSync(ctx)

	// Should not panic and should return nil (non-fatal).
	if err := step.Run(); err != nil {
		t.Errorf("expected nil error (non-fatal), got %v", err)
	}
}

func TestStepSnapshotBeforeSync_CreatesSnapshotDir(t *testing.T) {
	base := t.TempDir()
	ide := makeIDE(t, base, "VSCode")

	// Pre-create a rules file so the snapshot has something to archive.
	embedded := []byte(templates.MustRead("base/persona.md"))
	rulePath := filepath.Join(ide.rulesDir, "specai.md")
	if err := os.MkdirAll(ide.rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rulePath, embedded, 0644); err != nil {
		t.Fatal(err)
	}

	ctx := &SyncContext{HomeDir: base, IDEs: []system.IDEAdapter{ide}}
	step := NewStepSnapshotBeforeSync(ctx)

	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// A backup directory should have been created under <base>/.specai/backups/.
	backupsRoot := filepath.Join(base, ".specai", "backups")
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatalf("backup root not created: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one snapshot directory, got none")
	}
}
