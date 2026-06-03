package agents

import "github.com/KevG1t/SpecAI/internal/model"

// ValidationWarning is a non-fatal advisory from pre-flight agent selection checks.
type ValidationWarning struct {
	Code    string // machine-readable key, e.g. "AGENT_COLLISION"
	Message string // human-readable description
}

// ValidateAgentSelection checks for known problematic agent combinations.
// Returns a slice of warnings (empty = no issues).
func ValidateAgentSelection(agentIDs []model.AgentID) []ValidationWarning {
	var warnings []ValidationWarning

	hasGemini := containsAgentID(agentIDs, model.AgentGeminiCLI)
	hasAntigravity := containsAgentID(agentIDs, model.AgentAntigravity)

	if hasGemini && hasAntigravity {
		warnings = append(warnings, ValidationWarning{
			Code: "AGENT_COLLISION",
			Message: "Gemini CLI and Antigravity both write to ~/.gemini/. " +
				"Antigravity may override Gemini CLI config. Proceed with caution.",
		})
	}

	return warnings
}

func containsAgentID(agentIDs []model.AgentID, target model.AgentID) bool {
	for _, id := range agentIDs {
		if id == target {
			return true
		}
	}
	return false
}
