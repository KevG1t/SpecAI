package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
)

// testPersonaFS is an in-memory fs.ReadFileFS for testing persona resolution.
type testPersonaFS struct {
	files map[string][]byte
}

func (t *testPersonaFS) ReadFile(name string) ([]byte, error) {
	data, ok := t.files[name]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	return data, nil
}

func newTestPersonaFS(files map[string][]byte) *testPersonaFS {
	return &testPersonaFS{files: files}
}

// makeInstallContext builds an InstallContext with a temp dir and given adapters.
func makeInstallContext(t *testing.T, adapters ...system.IDEAdapter) (*InstallContext, string) {
	t.Helper()
	base := t.TempDir()
	return &InstallContext{
		HomeDir: base,
		IDEs:    adapters,
	}, base
}

func TestResolvePersona_AgentHasDedicatedPersona(t *testing.T) {
	agentPersona := []byte("agent-specific persona content")
	pFS := newTestPersonaFS(map[string][]byte{
		"claude/persona-gentleman.md": agentPersona,
	})

	ide := makeStubIDEFull(t, t.TempDir(), "claude", model.AgentClaudeCode, "claude")
	got, err := resolvePersona(pFS, ide)
	if err != nil {
		t.Fatalf("resolvePersona returned unexpected error: %v", err)
	}
	if string(got) != string(agentPersona) {
		t.Errorf("resolvePersona = %q, want %q", got, agentPersona)
	}
}

func TestResolvePersona_FallbackToGenericGentleman(t *testing.T) {
	genericGentleman := []byte("generic gentleman persona")
	pFS := newTestPersonaFS(map[string][]byte{
		"generic/persona-gentleman.md": genericGentleman,
		// No agent-specific persona
	})

	ide := makeStubIDEFull(t, t.TempDir(), "claude", model.AgentClaudeCode, "claude")
	got, err := resolvePersona(pFS, ide)
	if err != nil {
		t.Fatalf("resolvePersona returned unexpected error: %v", err)
	}
	if string(got) != string(genericGentleman) {
		t.Errorf("resolvePersona = %q, want %q", got, genericGentleman)
	}
}

func TestResolvePersona_FallbackToNeutralPersona(t *testing.T) {
	neutralPersona := []byte("neutral persona content")
	pFS := newTestPersonaFS(map[string][]byte{
		"generic/persona-neutral.md": neutralPersona,
		// No agent-specific nor generic gentleman
	})

	ide := makeStubIDEFull(t, t.TempDir(), "cursor", model.AgentCursor, "cursor")
	got, err := resolvePersona(pFS, ide)
	if err != nil {
		t.Fatalf("resolvePersona returned unexpected error: %v", err)
	}
	if string(got) != string(neutralPersona) {
		t.Errorf("resolvePersona = %q, want %q", got, neutralPersona)
	}
}

func TestResolvePersona_ErrorWhenNoPersonaFound(t *testing.T) {
	pFS := newTestPersonaFS(map[string][]byte{
		// No persona files at all
	})

	ide := makeStubIDEFull(t, t.TempDir(), "cursor", model.AgentCursor, "cursor")
	_, err := resolvePersona(pFS, ide)
	if err == nil {
		t.Error("expected error when no persona file found, got nil")
	}
}

func TestStepInstallGlobalRules_WritesRuleFile(t *testing.T) {
	ctx, base := makeInstallContext(t,
		makeStubIDEFull(t, t.TempDir(), "vscode", model.AgentVSCodeCopilot, "generic"),
	)

	// Override the stub's rulesDir to be inside base so we can check it
	ide := makeStubIDEFull(t, base, "vscode", model.AgentVSCodeCopilot, "generic")
	ctx.IDEs = []system.IDEAdapter{ide}

	step := NewStepInstallGlobalRules(ctx)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	rulePath := filepath.Join(ide.rulesDir, "specai.md")
	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("rule file not written: %v", err)
	}
	if len(data) == 0 {
		t.Error("rule file is empty")
	}
}
