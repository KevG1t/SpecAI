package steps

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/system"
)

// TestStepInjectMCPComponents_NoComponents verifies that when plan has no Notion/Jira,
// AuthGuidance remains empty and no error is returned.
func TestStepInjectMCPComponents_NoComponents(t *testing.T) {
	ctx := &InstallContext{
		HomeDir: t.TempDir(),
		IDEs:    []system.IDEAdapter{stubInstallableIDE{agentID: model.AgentClaudeCode}},
	}
	plan := planner.ResolvedPlan{
		OrderedComponents: []model.ComponentID{model.ComponentSDD},
	}

	step := NewStepInjectMCPComponents(ctx, plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if len(ctx.AuthGuidance) != 0 {
		t.Errorf("expected empty AuthGuidance with no MCP components, got %v", ctx.AuthGuidance)
	}
}

// TestStepInjectMCPComponents_ID verifies the step has a non-empty ID.
func TestStepInjectMCPComponents_ID(t *testing.T) {
	ctx := &InstallContext{HomeDir: t.TempDir()}
	step := NewStepInjectMCPComponents(ctx, planner.ResolvedPlan{})
	if step.ID() == "" {
		t.Error("expected non-empty step ID")
	}
}

// TestStepInjectMCPComponents_NotionInPlan_AppliesToClaude verifies that when Notion is
// in the plan, the step runs InjectNotion for each IDE and appends guidance to AuthGuidance.
// This is an integration test that uses the real agents.Adapter and filesystem.
func TestStepInjectMCPComponents_NotionInPlan_AppliesToClaude(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MCP inject integration test in short mode")
	}

	homeDir := t.TempDir()
	ctx := &InstallContext{
		HomeDir: homeDir,
		IDEs:    []system.IDEAdapter{stubInstallableIDE{agentID: model.AgentClaudeCode}},
	}
	plan := planner.ResolvedPlan{
		OrderedComponents: []model.ComponentID{model.ComponentNotion},
	}

	step := NewStepInjectMCPComponents(ctx, plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// After running, AuthGuidance should contain Notion guidance.
	found := false
	for _, g := range ctx.AuthGuidance {
		if strings.Contains(g, "Notion") || strings.Contains(g, "notion") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Notion guidance in AuthGuidance, got %v", ctx.AuthGuidance)
	}
}
