package steps

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

func TestStepInjectOpenCodeOverlay_Run(t *testing.T) {
	tests := []struct {
		name             string
		ides             []model.AgentID
		initialJSON      string
		wantKey          string   // top-level key that must exist after merge
		wantUserKeyPres  []string // user keys that must be preserved
		wantFileCreated  bool
	}{
		{
			name:            "creates file when absent",
			ides:            []model.AgentID{model.AgentOpenCode},
			initialJSON:     "",
			wantKey:         "agent",
			wantFileCreated: true,
		},
		{
			name: "merges into existing config preserving user keys",
			ides: []model.AgentID{model.AgentOpenCode},
			initialJSON: `{"theme":"dark","model":"anthropic/claude-opus-4"}`,
			wantKey:     "agent",
			wantUserKeyPres: []string{"theme", "model"},
		},
		{
			name: "idempotent: re-merge does not duplicate",
			ides: []model.AgentID{model.AgentOpenCode},
			wantKey: "agent",
		},
		{
			name: "skipped when opencode not in IDEs",
			ides: []model.AgentID{model.AgentClaudeCode},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := afero.NewMemMapFs()
			homeDir := "/home/testuser"
			configPath := filepath.Join(homeDir, ".config", "opencode", "opencode.json")

			if tt.initialJSON != "" {
				if err := fsys.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatalf("setup MkdirAll: %v", err)
				}
				if err := afero.WriteFile(fsys, configPath, []byte(tt.initialJSON), 0644); err != nil {
					t.Fatalf("setup WriteFile: %v", err)
				}
			}

			adapters := makeAdaptersFrom(tt.ides)
			ctx := &InstallContext{IDEs: adapters, HomeDir: homeDir}
			step := &StepInjectOpenCodeOverlay{ctx: ctx, fs: fsys}

			if err := step.Run(); err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if tt.wantKey == "" {
				// Step should have been skipped — file should not exist.
				if _, err := fsys.Stat(configPath); err == nil {
					t.Errorf("config file created unexpectedly when opencode not selected")
				}
				return
			}

			result, err := afero.ReadFile(fsys, configPath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			var resultMap map[string]any
			if err := json.Unmarshal(result, &resultMap); err != nil {
				t.Fatalf("result is not valid JSON: %v", err)
			}

			if _, ok := resultMap[tt.wantKey]; !ok {
				t.Errorf("result missing expected key %q; got keys: %v", tt.wantKey, mapKeys(resultMap))
			}

			for _, userKey := range tt.wantUserKeyPres {
				if _, ok := resultMap[userKey]; !ok {
					t.Errorf("user key %q was removed during merge", userKey)
				}
			}

			// Idempotency check.
			if err := step.Run(); err != nil {
				t.Fatalf("second Run() error = %v", err)
			}
			result2, _ := afero.ReadFile(fsys, configPath)
			var map2 map[string]any
			if err := json.Unmarshal(result2, &map2); err != nil {
				t.Fatalf("second result is not valid JSON: %v", err)
			}
			// Agent count should not have doubled.
			agents1, _ := resultMap["agent"].(map[string]any)
			agents2, _ := map2["agent"].(map[string]any)
			if len(agents2) > len(agents1) {
				t.Errorf("agents doubled after re-merge: before=%d, after=%d", len(agents1), len(agents2))
			}
		})
	}
}

func makeAdaptersFrom(ids []model.AgentID) []system.IDEAdapter {
	var out []system.IDEAdapter
	for _, id := range ids {
		if a := system.GetAdapterByAgentID(id); a != nil {
			out = append(out, a)
		}
	}
	return out
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
