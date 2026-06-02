package steps

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

func TestStepInjectSDDMemory_Run(t *testing.T) {
	const startMarker = "<!-- specai:sdd-memory-protocol:start -->"
	const endMarker = "<!-- specai:sdd-memory-protocol:end -->"

	tests := []struct {
		name         string
		ides         []model.AgentID
		initialFile  string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name: "first-time injection into missing CLAUDE.md",
			ides: []model.AgentID{model.AgentClaudeCode},
			wantContains: []string{
				startMarker,
				endMarker,
			},
		},
		{
			name:        "appends to existing CLAUDE.md",
			ides:        []model.AgentID{model.AgentClaudeCode},
			initialFile: "# Existing rules\n\nSome content.",
			wantContains: []string{
				"# Existing rules",
				startMarker,
				endMarker,
			},
		},
		{
			name: "idempotent: no duplication on second run",
			ides: []model.AgentID{model.AgentClaudeCode},
			wantContains: []string{
				startMarker,
			},
		},
		{
			name: "non-claude IDE is skipped",
			ides: []model.AgentID{model.AgentCursor},
			wantAbsent: []string{
				startMarker,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := afero.NewMemMapFs()
			homeDir := "/home/testuser"
			claudeMDPath := filepath.Join(homeDir, ".claude", "CLAUDE.md")

			if tt.initialFile != "" {
				if err := afero.WriteFile(fsys, claudeMDPath, []byte(tt.initialFile), 0644); err != nil {
					t.Fatalf("setup WriteFile: %v", err)
				}
			}

			adapters := makeAdapters(tt.ides)
			ctx := &InstallContext{IDEs: adapters, HomeDir: homeDir}
			step := &StepInjectSDDMemory{ctx: ctx, fs: fsys}

			if err := step.Run(); err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			result, _ := afero.ReadFile(fsys, claudeMDPath)
			resultStr := string(result)

			for _, want := range tt.wantContains {
				if !strings.Contains(resultStr, want) {
					t.Errorf("result missing %q\nGot:\n%s", want, resultStr)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(resultStr, absent) {
					t.Errorf("result should not contain %q\nGot:\n%s", absent, resultStr)
				}
			}

			// Idempotency check: run again and verify no duplication.
			if strings.Contains(tt.name, "idempotent") || len(tt.wantContains) > 0 {
				if err := step.Run(); err != nil {
					t.Fatalf("second Run() error = %v", err)
				}
				result2, _ := afero.ReadFile(fsys, claudeMDPath)
				count := strings.Count(string(result2), startMarker)
				if count > 1 {
					t.Errorf("idempotent: start marker count = %d after second run, want 1", count)
				}
			}
		})
	}
}

// makeAdapters converts a slice of AgentIDs to IDEAdapters using the real adapter registry.
func makeAdapters(ids []model.AgentID) []system.IDEAdapter {
	var out []system.IDEAdapter
	for _, id := range ids {
		if a := system.GetAdapterByAgentID(id); a != nil {
			out = append(out, a)
		}
	}
	return out
}
