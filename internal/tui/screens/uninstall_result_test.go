package screens

import (
	"strings"
	"testing"

	componentuninstall "github.com/KevG1t/specai/internal/components/uninstall"
	"github.com/KevG1t/specai/internal/model"
)

func TestRenderUninstallResultIncludesManualCleanup(t *testing.T) {
	out := RenderUninstallResult(componentuninstall.Result{
		RemovedDirectories: []string{"/tmp/skills"},
		ManualActions: []string{
			"Remove manually if no longer needed: /tmp/skills (directory still contains non-managed files)",
		},
	}, nil, "", nil, model.SddMemoryUninstallScopeGlobal, false, nil, nil)

	if !strings.Contains(out, "Manual cleanup required") {
		t.Fatalf("RenderUninstallResult() should include manual cleanup heading; got:\n%s", out)
	}
	if !strings.Contains(out, "/tmp/skills") {
		t.Fatalf("RenderUninstallResult() should include manual cleanup item; got:\n%s", out)
	}
}

func TestRenderUninstallConfirmIncludesSelectedProfiles(t *testing.T) {
	out := RenderUninstallConfirm(
		model.UninstallModePartial,
		[]model.AgentID{model.AgentOpenCode},
		[]model.ComponentID{model.ComponentSDD},
		[]string{"cheap"},
		model.SddMemoryUninstallScopeGlobal,
		false,
		0,
		false,
		0,
	)

	if !strings.Contains(out, "Profiles to remove") {
		t.Fatalf("RenderUninstallConfirm() should include profile section; got:\n%s", out)
	}
	if !strings.Contains(out, "cheap") {
		t.Fatalf("RenderUninstallConfirm() should include selected profile name; got:\n%s", out)
	}
}

func TestRenderUninstallConfirmIncludesSddMemoryProjectScopeDetails(t *testing.T) {
	out := RenderUninstallConfirm(
		model.UninstallModePartial,
		[]model.AgentID{model.AgentOpenCode},
		[]model.ComponentID{model.ComponentSddMemory},
		nil,
		model.SddMemoryUninstallScopeProject,
		true,
		0,
		false,
		0,
	)

	if !strings.Contains(out, "SddMemory cleanup scope") {
		t.Fatalf("RenderUninstallConfirm() should include SddMemory cleanup scope heading; got:\n%s", out)
	}
	if !strings.Contains(out, "Project-only") {
		t.Fatalf("RenderUninstallConfirm() should include project-only scope label; got:\n%s", out)
	}
	if !strings.Contains(out, ".sdd-memory/") {
		t.Fatalf("RenderUninstallConfirm() should mention .sdd-memory project data removal; got:\n%s", out)
	}
}

func TestRenderUninstallResultIncludesSelectedProfiles(t *testing.T) {
	out := RenderUninstallResult(componentuninstall.Result{}, nil, model.UninstallModePartial, []string{"cheap", "fast"}, model.SddMemoryUninstallScopeGlobal, false, nil, nil)

	if !strings.Contains(out, "Profiles removed") {
		t.Fatalf("RenderUninstallResult() should include profile summary heading; got:\n%s", out)
	}
	if !strings.Contains(out, "cheap") || !strings.Contains(out, "fast") {
		t.Fatalf("RenderUninstallResult() should include selected profile names; got:\n%s", out)
	}
}

func TestRenderUninstallResultIncludesSddMemoryScopeSummary(t *testing.T) {
	out := RenderUninstallResult(componentuninstall.Result{
		RemovedDirectories: []string{"/tmp/workspace/.sdd-memory"},
	}, nil, model.UninstallModePartial, nil, model.SddMemoryUninstallScopeProject, true, nil, nil)

	if !strings.Contains(out, "SddMemory scope: Project-only") {
		t.Fatalf("RenderUninstallResult() should include SddMemory project scope summary; got:\n%s", out)
	}
}
