package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

// ── removeJSONKey helper tests (T07a) ────────────────────────────────────────

func TestRemoveJSONKey_KeyPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	initial := map[string]any{
		"outputStyle":   "Argentina",
		"thinkingVerbs": []any{"Analizando", "Evaluando"},
	}
	writeJSON(t, path, initial)

	removed, err := removeJSONKey(path, "thinkingVerbs")
	if err != nil {
		t.Fatalf("removeJSONKey returned error: %v", err)
	}
	if !removed {
		t.Error("expected removed=true when key is present")
	}

	var got map[string]any
	readJSON(t, path, &got)

	if _, ok := got["thinkingVerbs"]; ok {
		t.Error("thinkingVerbs key still present after removal")
	}
	if got["outputStyle"] != "Argentina" {
		t.Error("outputStyle should be preserved after removal")
	}
}

func TestRemoveJSONKey_KeyAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	initial := map[string]any{"outputStyle": "Argentina"}
	writeJSON(t, path, initial)

	removed, err := removeJSONKey(path, "thinkingVerbs")
	if err != nil {
		t.Fatalf("removeJSONKey returned error: %v", err)
	}
	if removed {
		t.Error("expected removed=false when key is absent")
	}
}

func TestRemoveJSONKey_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	// Don't create the file (osReadFile returns nil for missing files).

	removed, err := removeJSONKey(path, "thinkingVerbs")
	if err != nil {
		t.Fatalf("removeJSONKey on missing file returned error: %v", err)
	}
	if removed {
		t.Error("expected removed=false for missing file")
	}
}

// ── Thinking verbs inject/cleanup tests (T08) ─────────────────────────────────

func TestInjectThinkingVerbs_ArgentinaClaudeCode_Injects(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	writeJSON(t, settingsPath, map[string]any{})

	injected, err := injectThinkingVerbsIfArgentina(model.PersonaArgentina, model.AgentClaudeCode, settingsPath)
	if err != nil {
		t.Fatalf("injectThinkingVerbsIfArgentina returned error: %v", err)
	}
	if !injected {
		t.Error("expected inject=true for Argentina+ClaudeCode")
	}

	var got map[string]any
	readJSON(t, settingsPath, &got)

	if _, ok := got["thinkingVerbs"]; !ok {
		t.Error("thinkingVerbs key not present after inject")
	}
}

func TestInjectThinkingVerbs_NonArgentina_CleansUp(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")

	// Pre-populate with thinkingVerbs.
	writeJSON(t, settingsPath, map[string]any{
		"thinkingVerbs": []any{"Analizando"},
	})

	// Non-Argentina persona + ClaudeCode → should remove thinkingVerbs.
	err := cleanupThinkingVerbsIfNonArgentina(model.PersonaNeutral, model.AgentClaudeCode, settingsPath)
	if err != nil {
		t.Fatalf("cleanupThinkingVerbsIfNonArgentina returned error: %v", err)
	}

	var got map[string]any
	readJSON(t, settingsPath, &got)

	if _, ok := got["thinkingVerbs"]; ok {
		t.Error("thinkingVerbs key still present after cleanup for non-Argentina persona")
	}
}

func TestInjectThinkingVerbs_ArgentinaOpenCode_NoInject(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	writeJSON(t, settingsPath, map[string]any{})

	// Argentina + OpenCode → should NOT inject (only ClaudeCode gets thinking verbs).
	injected, err := injectThinkingVerbsIfArgentina(model.PersonaArgentina, model.AgentOpenCode, settingsPath)
	if err != nil {
		t.Fatalf("injectThinkingVerbsIfArgentina returned error: %v", err)
	}
	if injected {
		t.Error("expected inject=false for Argentina+OpenCode (not ClaudeCode)")
	}

	var got map[string]any
	readJSON(t, settingsPath, &got)
	if _, ok := got["thinkingVerbs"]; ok {
		t.Error("thinkingVerbs should not be injected for OpenCode")
	}
}

func TestInjectThinkingVerbs_Idempotent(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	writeJSON(t, settingsPath, map[string]any{})

	// First inject.
	_, err := injectThinkingVerbsIfArgentina(model.PersonaArgentina, model.AgentClaudeCode, settingsPath)
	if err != nil {
		t.Fatalf("first inject failed: %v", err)
	}
	data1, _ := os.ReadFile(settingsPath)

	// Second inject — should produce identical file.
	_, err = injectThinkingVerbsIfArgentina(model.PersonaArgentina, model.AgentClaudeCode, settingsPath)
	if err != nil {
		t.Fatalf("second inject failed: %v", err)
	}
	data2, _ := os.ReadFile(settingsPath)

	if string(data1) != string(data2) {
		t.Errorf("inject is not idempotent:\nfirst:  %s\nsecond: %s", data1, data2)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write JSON: %v", err)
	}
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
}
