package steps

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/spf13/afero"
)

// makeStubIDEFull creates a stubIDE with agentID and assetFolder set.
func makeStubIDEFull(t *testing.T, base, name string, agentID model.AgentID, assetFolder string) stubIDE {
	t.Helper()
	rulesDir := filepath.Join(base, name, "rules")
	skillsDir := filepath.Join(base, name, "skills")
	return stubIDE{
		name:        name,
		rulesDir:    rulesDir,
		skillsDir:   skillsDir,
		agentID:     agentID,
		assetFolder: assetFolder,
	}
}

// --- Mock injector for behavior testing ---

// mockAssetInjector records all InjectAgentFolder and InjectSharedSkills calls.
type mockAssetInjector struct {
	agentCalls  []agentCall
	sharedCalls []string
}

type agentCall struct {
	agentFolder string
	targetDir   string
}

func (m *mockAssetInjector) InjectAgentFolder(agentFolder, targetDir string) error {
	m.agentCalls = append(m.agentCalls, agentCall{agentFolder: agentFolder, targetDir: targetDir})
	return nil
}

func (m *mockAssetInjector) InjectSharedSkills(targetDir string) error {
	m.sharedCalls = append(m.sharedCalls, targetDir)
	return nil
}

// TestInjectAssets_PerAgentWalk verifies that each adapter triggers one InjectAgentFolder
// call with agentFolder and ConfigDir as the targetDir.
func TestInjectAssets_PerAgentWalk(t *testing.T) {
	mock := &mockAssetInjector{}

	base := t.TempDir()
	claudeIDE := makeStubIDEFull(t, base, "claude", model.AgentClaudeCode, "claude")
	codexIDE := makeStubIDEFull(t, base, "codex", model.AgentCodex, "codex")

	ctx := &InstallContext{
		HomeDir: base,
		IDEs:    []system.IDEAdapter{claudeIDE, codexIDE},
	}
	plan := planner.ResolvedPlan{}

	step := NewStepInjectAssets(ctx, mock, plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Expect exactly 2 agent calls (one per adapter)
	if len(mock.agentCalls) != 2 {
		t.Fatalf("expected 2 InjectAgentFolder calls, got %d", len(mock.agentCalls))
	}

	// First call should be for claude — target is ConfigDir, not GlobalSkillsDir
	if mock.agentCalls[0].agentFolder != "claude" {
		t.Errorf("first agent call folder = %q, want %q", mock.agentCalls[0].agentFolder, "claude")
	}
	if mock.agentCalls[0].targetDir != claudeIDE.ConfigDir(base) {
		t.Errorf("first agent call targetDir = %q, want ConfigDir %q", mock.agentCalls[0].targetDir, claudeIDE.ConfigDir(base))
	}

	// Second call should be for codex — target is ConfigDir
	if mock.agentCalls[1].agentFolder != "codex" {
		t.Errorf("second agent call folder = %q, want %q", mock.agentCalls[1].agentFolder, "codex")
	}
	if mock.agentCalls[1].targetDir != codexIDE.ConfigDir(base) {
		t.Errorf("second agent call targetDir = %q, want ConfigDir %q", mock.agentCalls[1].targetDir, codexIDE.ConfigDir(base))
	}
}

// TestInjectAssets_SharedSkillsPerAgent verifies InjectSharedSkills is called once per
// adapter that SupportsSkills() — skills are written to each adapter's SkillsDir.
func TestInjectAssets_SharedSkillsPerAgent(t *testing.T) {
	mock := &mockAssetInjector{}

	base := t.TempDir()
	ide1 := makeStubIDEFull(t, base, "claude", model.AgentClaudeCode, "claude")
	ide2 := makeStubIDEFull(t, base, "codex", model.AgentCodex, "codex")
	ide3 := makeStubIDEFull(t, base, "cursor", model.AgentCursor, "cursor")

	ctx := &InstallContext{
		HomeDir: base,
		IDEs:    []system.IDEAdapter{ide1, ide2, ide3},
	}
	plan := planner.ResolvedPlan{}

	step := NewStepInjectAssets(ctx, mock, plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// All 3 adapters support skills — InjectSharedSkills called 3 times (once per adapter)
	if len(mock.sharedCalls) != 3 {
		t.Errorf("InjectSharedSkills called %d times, want 3 (once per adapter)", len(mock.sharedCalls))
	}

	// Each call must target the adapter's SkillsDir, not ~/.specai/skills
	if mock.sharedCalls[0] != ide1.SkillsDir(base) {
		t.Errorf("first skills call target = %q, want SkillsDir %q", mock.sharedCalls[0], ide1.SkillsDir(base))
	}
	if mock.sharedCalls[1] != ide2.SkillsDir(base) {
		t.Errorf("second skills call target = %q, want SkillsDir %q", mock.sharedCalls[1], ide2.SkillsDir(base))
	}
	if mock.sharedCalls[2] != ide3.SkillsDir(base) {
		t.Errorf("third skills call target = %q, want SkillsDir %q", mock.sharedCalls[2], ide3.SkillsDir(base))
	}

	// Must NOT use ~/.specai/skills
	for _, call := range mock.sharedCalls {
		if strings.Contains(call, ".specai") {
			t.Errorf("skills must not write to ~/.specai/skills, got: %q", call)
		}
	}
}

// TestInjectAssets_MissingAgentFolderSkipped verifies that the real AssetInjector
// handles a missing embed.FS folder gracefully (no error).
func TestInjectAssets_MissingAgentFolderSkipped(t *testing.T) {
	memFS := afero.NewMemMapFs()

	base := t.TempDir()
	kiloIDE := makeStubIDEFull(t, base, "kilo", model.AgentKilocode, "nonexistent-agent-folder-xyz")

	ctx := &InstallContext{
		HomeDir: base,
		IDEs:    []system.IDEAdapter{kiloIDE},
	}
	plan := planner.ResolvedPlan{}

	step := NewStepInjectAssets(ctx, NewAssetInjector(memFS), plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() returned unexpected error for missing agent folder: %v", err)
	}
}

// TestInjectAssets_NoAdapters verifies zero adapters → zero agent calls and zero shared calls.
func TestInjectAssets_NoAdapters(t *testing.T) {
	mock := &mockAssetInjector{}

	base := t.TempDir()
	ctx := &InstallContext{
		HomeDir: base,
		IDEs:    nil,
	}
	plan := planner.ResolvedPlan{}

	step := NewStepInjectAssets(ctx, mock, plan)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() with no adapters returned unexpected error: %v", err)
	}

	// No agent calls when there are no adapters
	if len(mock.agentCalls) != 0 {
		t.Errorf("expected 0 agent calls with no adapters, got %d", len(mock.agentCalls))
	}
	// No shared skills calls when there are no adapters
	if len(mock.sharedCalls) != 0 {
		t.Errorf("InjectSharedSkills called %d times, want 0 with no adapters", len(mock.sharedCalls))
	}
}

// TestAssetInjector_InjectSharedSkills_RealFS verifies the real injector writes
// shared skills content into the target directory.
func TestAssetInjector_InjectSharedSkills_RealFS(t *testing.T) {
	memFS := afero.NewMemMapFs()
	injector := NewAssetInjector(memFS)

	targetDir := "/target/shared-skills"
	if err := injector.InjectSharedSkills(targetDir); err != nil {
		t.Fatalf("InjectSharedSkills failed: %v", err)
	}

	// The target directory should exist because skills/sdd-tasks/SKILL.md etc. are embedded
	exists, err := afero.DirExists(memFS, targetDir)
	if err != nil {
		t.Fatalf("checking dir existence: %v", err)
	}
	if !exists {
		t.Errorf("expected shared skills dir %s to be created", targetDir)
	}

	// Check if a specific known skill file was copied
	expectedFile := filepath.Join(targetDir, "sdd-tasks", "SKILL.md")
	fileExists, err := afero.Exists(memFS, expectedFile)
	if err != nil {
		t.Fatalf("checking file existence: %v", err)
	}
	if !fileExists {
		t.Errorf("expected skill file %s to be injected", expectedFile)
	}
}
