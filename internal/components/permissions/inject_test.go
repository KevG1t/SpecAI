package permissions

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

// TestAgentOverlayReturnsNonNilForNewAgents verifies that all 8 newly added agents
// have a non-nil permissions overlay (except Kimi which explicitly returns nil).
func TestAgentOverlayReturnsNonNilForNewAgents(t *testing.T) {
	tests := []struct {
		agent       model.AgentID
		expectNil   bool
		description string
	}{
		{model.AgentWindsurf, false, "Windsurf must have overlay"},
		{model.AgentKimi, true, "Kimi intentionally returns nil (TOML not supported)"},
		{model.AgentKiroIDE, false, "Kiro must have overlay"},
		{model.AgentTrae, false, "Trae must have overlay"},
		{model.AgentPi, false, "Pi must have overlay"},
		{model.AgentOpenClaw, false, "OpenClaw must have overlay"},
		{model.AgentAntigravity, false, "Antigravity must have overlay"},
		{model.AgentQwenCode, false, "QwenCode must have overlay"},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			overlay := agentOverlay(tc.agent)
			if tc.expectNil && overlay != nil {
				t.Fatalf("agentOverlay(%q): expected nil, got %d bytes", tc.agent, len(overlay))
			}
			if !tc.expectNil && overlay == nil {
				t.Fatalf("agentOverlay(%q): expected non-nil overlay, got nil", tc.agent)
			}
		})
	}
}

// TestAgentOverlayContainsDotEnvDenyRule verifies that each new agent overlay
// blocks .env file access.
func TestAgentOverlayContainsDotEnvDenyRule(t *testing.T) {
	agentsWithOverlay := []model.AgentID{
		model.AgentWindsurf,
		model.AgentKiroIDE,
		model.AgentTrae,
		model.AgentPi,
		model.AgentOpenClaw,
		model.AgentAntigravity,
		model.AgentQwenCode,
	}

	for _, id := range agentsWithOverlay {
		t.Run(string(id), func(t *testing.T) {
			overlay := agentOverlay(id)
			if overlay == nil {
				t.Fatalf("agentOverlay(%q) returned nil, expected overlay with deny rules", id)
			}
			content := string(overlay)
			if !strings.Contains(content, ".env") {
				t.Fatalf("agentOverlay(%q) overlay does not contain .env deny rule:\n%s", id, content)
			}
		})
	}
}

// TestAgentOverlayContainsDestructiveGitDenyRule verifies that each overlay
// blocks destructive git operations.
func TestAgentOverlayContainsDestructiveGitDenyRule(t *testing.T) {
	agentsWithOverlay := []model.AgentID{
		model.AgentWindsurf,
		model.AgentKiroIDE,
		model.AgentTrae,
		model.AgentPi,
		model.AgentOpenClaw,
		model.AgentAntigravity,
		model.AgentQwenCode,
	}

	for _, id := range agentsWithOverlay {
		t.Run(string(id), func(t *testing.T) {
			overlay := agentOverlay(id)
			if overlay == nil {
				t.Fatalf("agentOverlay(%q) returned nil", id)
			}
			content := string(overlay)
			hasForce := strings.Contains(content, "git push --force") || strings.Contains(content, "force")
			hasHard := strings.Contains(content, "reset --hard") || strings.Contains(content, "hard")
			if !hasForce && !hasHard {
				t.Fatalf("agentOverlay(%q) does not contain destructive git op rule:\n%s", id, content)
			}
		})
	}
}

// TestKimiOverlayIsNilWithComment verifies the Kimi case returns nil.
// The code comment is the documentation for this behavior.
func TestKimiOverlayIsNilWithComment(t *testing.T) {
	if kimiOverlayJSON != nil {
		t.Fatal("kimiOverlayJSON must be nil (TOML permissions not yet supported)")
	}
	overlay := agentOverlay(model.AgentKimi)
	if overlay != nil {
		t.Fatal("agentOverlay(AgentKimi) must return nil")
	}
}

// TestQwenCodeOverlayExtendedWithDenyList verifies that qwenCode overlay now
// includes a deny list (previously only had defaultMode).
func TestQwenCodeOverlayExtendedWithDenyList(t *testing.T) {
	overlay := agentOverlay(model.AgentQwenCode)
	if overlay == nil {
		t.Fatal("agentOverlay(AgentQwenCode) must not be nil")
	}
	content := string(overlay)
	if !strings.Contains(content, "deny") {
		t.Fatalf("qwenCode overlay must contain deny list, got:\n%s", content)
	}
}
