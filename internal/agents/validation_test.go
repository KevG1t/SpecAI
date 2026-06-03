package agents_test

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/model"
)

func TestValidateAgentSelection(t *testing.T) {
	tests := []struct {
		name          string
		agents        []model.AgentID
		wantWarnings  int
		wantCode      string
	}{
		{
			name:         "gemini and antigravity selected → AGENT_COLLISION warning",
			agents:       []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
			wantWarnings: 1,
			wantCode:     "AGENT_COLLISION",
		},
		{
			name:         "gemini only → no warnings",
			agents:       []model.AgentID{model.AgentGeminiCLI},
			wantWarnings: 0,
		},
		{
			name:         "antigravity only → no warnings",
			agents:       []model.AgentID{model.AgentAntigravity},
			wantWarnings: 0,
		},
		{
			name:         "neither gemini nor antigravity → no warnings",
			agents:       []model.AgentID{model.AgentClaudeCode, model.AgentCursor},
			wantWarnings: 0,
		},
		{
			name:         "empty selection → no warnings",
			agents:       []model.AgentID{},
			wantWarnings: 0,
		},
		{
			name:         "all three plus others → AGENT_COLLISION warning",
			agents:       []model.AgentID{model.AgentClaudeCode, model.AgentGeminiCLI, model.AgentAntigravity},
			wantWarnings: 1,
			wantCode:     "AGENT_COLLISION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := agents.ValidateAgentSelection(tt.agents)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateAgentSelection() returned %d warnings, want %d", len(warnings), tt.wantWarnings)
			}
			if tt.wantCode != "" && len(warnings) > 0 {
				if warnings[0].Code != tt.wantCode {
					t.Errorf("warning.Code = %q, want %q", warnings[0].Code, tt.wantCode)
				}
				if warnings[0].Message == "" {
					t.Error("warning.Message is empty")
				}
			}
		})
	}
}
