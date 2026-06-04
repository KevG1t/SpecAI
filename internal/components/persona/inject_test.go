package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/agents"
	"github.com/KevG1t/specai/internal/agents/antigravity"
	"github.com/KevG1t/specai/internal/agents/claude"
	"github.com/KevG1t/specai/internal/agents/kilocode"
	"github.com/KevG1t/specai/internal/agents/kimi"
	"github.com/KevG1t/specai/internal/agents/openclaw"
	"github.com/KevG1t/specai/internal/agents/opencode"
	"github.com/KevG1t/specai/internal/assets"
	"github.com/KevG1t/specai/internal/model"
)

func antigravityAdapter() agents.Adapter { return antigravity.NewAdapter() }
func claudeAdapter() agents.Adapter      { return claude.NewAdapter() }
func kimiAdapter() agents.Adapter        { return kimi.NewAdapter() }
func kilocodeAdapter() agents.Adapter    { return kilocode.NewAdapter() }
func openclawAdapter() agents.Adapter    { return openclaw.NewAdapter() }
func opencodeAdapter() agents.Adapter    { return opencode.NewAdapter() }

func assertModismLanguageGuardrails(t *testing.T, text string, required []string, banned []string) {
	t.Helper()

	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("missing language guardrail %q", needle)
		}
	}

	for _, needle := range banned {
		if strings.Contains(text, needle) {
			t.Fatalf("contains drift-prone language instruction %q", needle)
		}
	}
}

func TestInjectClaudeModismWritesSectionWithRealContent(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("CLAUDE.md missing open marker for persona")
	}
	if !strings.Contains(text, "<!-- /specai:persona -->") {
		t.Fatal("CLAUDE.md missing close marker for persona")
	}
	// Real content check — the embedded persona has these patterns.
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("CLAUDE.md missing real persona content (expected 'Senior Architect')")
	}

	assertModismLanguageGuardrails(t, text,
		[]string{
			"Match the user's current language in your REPLY ONLY",
			"Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
			"When replying to the user in English, keep the full reply in natural English with the same warm energy.",
		},
		[]string{
			`Say "déjame verificar"`,
			"Spanish input → Rioplatense Spanish",
			"English input → same warm energy",
		},
	)
}

func TestInjectKimiModismIncludesProjectInstructionsAndLoadedSkills(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, kimiAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(kimi) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(kimi) changed = false")
	}

	// KIMI.md should be the static Jinja template (includes + variable placeholders).
	templatePath := filepath.Join(home, ".kimi", "KIMI.md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", templatePath, err)
	}

	text := string(content)
	if !strings.Contains(text, `{% include "output-style.md"`) {
		t.Fatal("KIMI.md template missing {% include \"output-style.md\" %}")
	}
	if !strings.Contains(text, "${KIMI_AGENTS_MD}") {
		t.Fatal("KIMI.md missing ${KIMI_AGENTS_MD} for project AGENTS.md parity")
	}
	if !strings.Contains(text, "${KIMI_SKILLS}") {
		t.Fatal("KIMI.md missing ${KIMI_SKILLS} for loaded-skills parity")
	}

	// output-style.md module should contain the Modism style content.
	outputStylePath := filepath.Join(home, ".kimi", "output-style.md")
	styleContent, err := os.ReadFile(outputStylePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outputStylePath, err)
	}
	if !strings.Contains(string(styleContent), "Modism Output Style") {
		t.Fatal("output-style.md missing Modism Output Style content")
	}
	assertModismLanguageGuardrails(t, string(styleContent),
		[]string{
			"Always match the user's current language in your reply.",
			"Do not drift into another language because of persona wording, examples, or stylistic momentum.",
			"When replying to the user in English, keep the full response in English unless the user explicitly asks for another language or you are translating/quoting.",
		},
		[]string{
			"### Spanish Input → Rioplatense Spanish (voseo)",
			`Use naturally: "Bien"`,
			`Use naturally: "Here's the thing"`,
		},
	)

	// persona.md module should exist and contain persona content.
	personaPath := filepath.Join(home, ".kimi", "persona.md")
	personaContent, err := os.ReadFile(personaPath)
	if err != nil {
		t.Fatalf("persona.md not written: %v", err)
	}
	assertModismLanguageGuardrails(t, string(personaContent),
		[]string{
			"Match the user's current language in your REPLY ONLY",
			"Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
			"When replying to the user in English, keep the full reply in natural English with the same warm energy.",
		},
		[]string{
			`Say "déjame verificar"`,
			"Spanish input → Rioplatense Spanish",
			"English input → same warm energy",
		},
	)
}

func TestInjectClaudeModismWritesOutputStyleFile(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Verify output-style file was written.
	stylePath := filepath.Join(home, ".claude", "output-styles", "modism.md")
	content, err := os.ReadFile(stylePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", stylePath, err)
	}

	text := string(content)
	if !strings.Contains(text, "name: Modism") {
		t.Fatal("Output style file missing YAML frontmatter 'name: Modism'")
	}
	if !strings.Contains(text, "keep-coding-instructions: true") {
		t.Fatal("Output style file missing 'keep-coding-instructions: true'")
	}
	if !strings.Contains(text, "Modism Output Style") {
		t.Fatal("Output style file missing 'Modism Output Style' heading")
	}
}

func TestInjectClaudeModismMergesOutputStyleIntoSettings(t *testing.T) {
	home := t.TempDir()

	// Pre-create a settings.json with some existing content.
	settingsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	existingSettings := `{"permissions": {"allow": ["Read"]}, "syntaxHighlightingDisabled": true}`
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(existingSettings), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Verify settings.json has outputStyle merged in.
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", settingsPath, err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsContent, &settings); err != nil {
		t.Fatalf("Unmarshal settings.json error = %v", err)
	}

	outputStyle, ok := settings["outputStyle"]
	if !ok {
		t.Fatal("settings.json missing 'outputStyle' key")
	}
	if outputStyle != "Modism" {
		t.Fatalf("settings.json outputStyle = %q, want %q", outputStyle, "Modism")
	}

	// Verify existing keys were preserved.
	if _, ok := settings["permissions"]; !ok {
		t.Fatal("settings.json lost 'permissions' key during merge")
	}
	if _, ok := settings["syntaxHighlightingDisabled"]; !ok {
		t.Fatal("settings.json lost 'syntaxHighlightingDisabled' key during merge")
	}
}

func TestInjectClaudeModismReturnsAllFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Should return 3 files: CLAUDE.md, output-style, settings.json.
	if len(result.Files) != 3 {
		t.Fatalf("Inject() returned %d files, want 3: %v", len(result.Files), result.Files)
	}

	wantSuffixes := []string{"CLAUDE.md", "modism.md", "settings.json"}
	for _, suffix := range wantSuffixes {
		found := false
		for _, f := range result.Files {
			if strings.HasSuffix(f, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Inject() missing file with suffix %q in %v", suffix, result.Files)
		}
	}
}

func TestInjectClaudeNeutralWritesFullPersonaWithoutRegionalLanguage(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	// Neutral persona is the same teacher — should have Senior Architect.
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("Neutral persona should contain 'Senior Architect'")
	}
	// Should NOT have specai-specific regional language.
	if strings.Contains(text, "Rioplatense") {
		t.Fatal("Neutral persona should not contain Rioplatense language")
	}
}

func TestInjectClaudeNeutralDoesNotWriteOutputStyle(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Should only return CLAUDE.md, no output-style file.
	if len(result.Files) != 1 {
		t.Fatalf("Neutral persona returned %d files, want 1: %v", len(result.Files), result.Files)
	}

	// Output-style file should NOT exist.
	stylePath := filepath.Join(home, ".claude", "output-styles", "modism.md")
	if _, err := os.Stat(stylePath); !os.IsNotExist(err) {
		t.Fatal("Neutral persona should NOT write output-style file")
	}
}

func TestInjectCustomClaudeDoesNothing(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaCustom)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Custom persona should NOT change anything")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Custom persona should return no files, got %v", result.Files)
	}

	// CLAUDE.md should NOT be created.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if _, err := os.Stat(claudeMD); !os.IsNotExist(err) {
		t.Fatal("Custom persona should NOT create CLAUDE.md")
	}
}

func TestInjectCustomOpenCodeDoesNothing(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), model.PersonaCustom)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Custom persona (OpenCode) should NOT change anything")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Custom persona (OpenCode) should return no files, got %v", result.Files)
	}

	// AGENTS.md should NOT be created.
	agentsMD := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if _, err := os.Stat(agentsMD); !os.IsNotExist(err) {
		t.Fatal("Custom persona (OpenCode) should NOT create AGENTS.md")
	}
}

func TestInjectOpenCodeModismWritesAgentsFile(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("AGENTS.md missing real persona content")
	}
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("AGENTS.md missing persona marker")
	}
}

func TestInjectAntigravityModismWritesMarkedPersonaSection(t *testing.T) {
	home := t.TempDir()
	promptPath := filepath.Join(home, ".gemini", "GEMINI.md")
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(promptPath, []byte("# User Gemini rules\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Inject(home, antigravityAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"# User Gemini rules",
		"<!-- specai:persona -->",
		"Senior Architect",
		"<!-- /specai:persona -->",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("GEMINI.md missing %q; got:\n%s", want, text)
		}
	}

	second, err := Inject(home, antigravityAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true; want false")
	}

	content, err = os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() after second inject error = %v", err)
	}
	if got := strings.Count(string(content), "<!-- specai:persona -->"); got != 1 {
		t.Fatalf("persona marker count = %d, want 1", got)
	}
}

func TestInjectOpenCodeModismDoesNotCreateSDDConductor(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}
	text := string(content)
	if strings.Contains(text, `"sdd-orchestrator"`) {
		t.Fatal("persona injection must not create legacy sdd-orchestrator conductor")
	}
	if strings.Contains(text, `"specai-orchestrator"`) {
		t.Fatal("persona injection must not create SDD conductor; SDD component owns specai-orchestrator")
	}
	if strings.Contains(text, `"specai-orchestrator"`) {
		t.Fatal("persona injection must not create SDD conductor; SDD component owns specai-orchestrator")
	}
	if !strings.Contains(text, `"modism"`) {
		t.Fatal("persona injection should still create the modism persona agent")
	}
}

func TestInjectOpenCodePreservesUserContentInsteadOfOverwriting(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	userContent := "# My custom rules\n\nDo not overwrite this file.\n"
	if err := os.WriteFile(path, []byte(userContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Do not overwrite this file.") {
		t.Fatal("AGENTS.md user content was overwritten")
	}
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("AGENTS.md missing managed persona section after inject")
	}
}

func TestInjectOpenClawWritesPersonaToWorkspaceSoulAndNotAgents(t *testing.T) {
	workspace := t.TempDir()
	adapter := openclawAdapter()
	agentsPath := filepath.Join(workspace, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# Existing agent protocols\n\nKeep SDD here.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) error = %v", err)
	}

	result, err := Inject(workspace, adapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(openclaw) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(openclaw) changed = false")
	}

	soulPath := filepath.Join(workspace, "SOUL.md")
	soulContent, err := os.ReadFile(soulPath)
	if err != nil {
		t.Fatalf("ReadFile(SOUL.md) error = %v", err)
	}
	soulText := string(soulContent)
	if !strings.Contains(soulText, "<!-- specai:persona -->") {
		t.Fatalf("SOUL.md missing managed persona marker; got:\n%s", soulText)
	}
	if !strings.Contains(soulText, "Senior Architect") {
		t.Fatalf("SOUL.md missing real persona content; got:\n%s", soulText)
	}
	if !strings.Contains(soulText, "Match the user's current language in your REPLY ONLY") {
		t.Fatalf("SOUL.md missing persona language guardrail; got:\n%s", soulText)
	}

	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	agentsText := string(agentsContent)
	if !strings.Contains(agentsText, "Keep SDD here.") {
		t.Fatalf("AGENTS.md user protocol content was modified; got:\n%s", agentsText)
	}
	if strings.Contains(agentsText, "<!-- specai:persona -->") || strings.Contains(agentsText, "Senior Architect") {
		t.Fatalf("OpenClaw persona must not be written to AGENTS.md; got:\n%s", agentsText)
	}
}

func TestInjectOpenClawSoulPersonaIsIdempotentAndPreservesUserContent(t *testing.T) {
	workspace := t.TempDir()
	soulPath := filepath.Join(workspace, "SOUL.md")
	if err := os.WriteFile(soulPath, []byte("# Custom soul\n\nKeep my tone note.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SOUL.md) error = %v", err)
	}

	adapter := openclawAdapter()
	first, err := Inject(workspace, adapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(openclaw) first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Inject(openclaw) first changed = false")
	}
	second, err := Inject(workspace, adapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(openclaw) second error = %v", err)
	}
	if second.Changed {
		t.Fatal("OpenClaw SOUL.md persona injection should be idempotent")
	}

	content, err := os.ReadFile(soulPath)
	if err != nil {
		t.Fatalf("ReadFile(SOUL.md) error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Keep my tone note.") {
		t.Fatalf("SOUL.md user content was lost; got:\n%s", text)
	}
	if count := strings.Count(text, "<!-- specai:persona -->"); count != 1 {
		t.Fatalf("SOUL.md has %d persona markers, want exactly 1", count)
	}
}

func TestInjectOpenClawRejectsAmbiguousWorkspacePath(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)

	result, err := Inject("", openclawAdapter(), model.PersonaModism)
	if err == nil {
		t.Fatalf("Inject(openclaw, empty workspace) error = nil, want deterministic ambiguity error; result=%+v", result)
	}
	if _, statErr := os.Stat(filepath.Join(cwd, "SOUL.md")); !os.IsNotExist(statErr) {
		t.Fatalf("ambiguous OpenClaw workspace must not create relative SOUL.md; stat err=%v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(cwd, "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("ambiguous OpenClaw workspace must not create relative AGENTS.md; stat err=%v", statErr)
	}
}

func TestInjectOpenCodeDoesNotStripLookalikeUserContent(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	lookalike := "## Rules\n\n- Team rules.\n\n## Personality\n\nSenior Architect for my org.\n\nDo not delete this custom preface.\n"
	if err := os.WriteFile(path, []byte(lookalike), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "Do not delete this custom preface.") {
		t.Fatal("OpenCode AGENTS.md lookalike user content was stripped")
	}
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("AGENTS.md missing managed persona section after inject")
	}
}

func TestInjectOpenCodePreservesUserPrefaceAboveATLBlock(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// User has custom content with fingerprint-like headings ABOVE an old ATL block.
	// ATL markers must NOT trigger persona legacy stripping.
	existing := "## Rules\n\n- My team's custom rules.\n\n## Personality\n\nSenior Architect in my org.\n\n" +
		"<!-- BEGIN:agent-teams-lite -->\nOld ATL content.\n<!-- END:agent-teams-lite -->\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "My team's custom rules.") {
		t.Fatal("user preface above ATL block was stripped — ATL should not enable persona stripping")
	}
	if strings.Contains(text, "BEGIN:agent-teams-lite") {
		t.Fatal("ATL block should have been stripped by StripLegacyATLBlock")
	}
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("AGENTS.md missing managed persona section")
	}
}

func TestInjectOpenCodeReplacesExactLegacyAssetWithoutDuplication(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Write the exact legacy asset (no markers) — simulates old installer output.
	legacyContent := assets.MustRead("opencode/persona-modism.md")
	if err := os.WriteFile(path, []byte(legacyContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	// Must have exactly ONE persona marker — no duplication.
	if strings.Count(text, "<!-- specai:persona -->") != 1 {
		t.Fatalf("expected exactly 1 persona marker, got %d — legacy asset was not replaced cleanly",
			strings.Count(text, "<!-- specai:persona -->"))
	}
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("persona content missing after replacing legacy asset")
	}
}

func TestInjectOpenCodePreservesUserPrefaceAboveManagedMarkers(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Simulate: user has custom content with fingerprint-like headings ABOVE
	// existing managed markers. This is the exact scenario where aggressive
	// legacy stripping would destroy user content.
	existing := "## Rules\n\n- My team's custom rules.\n\n## Personality\n\nSenior Architect in my org.\n\n" +
		"<!-- specai:sdd-memory-protocol -->\nSddMemory protocol here.\n<!-- /specai:sdd-memory-protocol -->\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "My team's custom rules.") {
		t.Fatal("user preface above managed markers was stripped — should be preserved")
	}
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("AGENTS.md missing managed persona section after inject")
	}
	if !strings.Contains(text, "<!-- specai:sdd-memory-protocol -->") {
		t.Fatal("existing sdd-memory section was lost")
	}
}

func TestInjectOpenCodeNeutralPreservesManagedSections(t *testing.T) {
	home := t.TempDir()

	// First install modism persona + simulate SDD/sdd-memory sections
	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")

	// Simulate SDD and sdd-memory sections appended by sdd.Inject and sdd-memory.Inject
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	withSections := string(existing) + "\n\n<!-- specai:sdd-orchestrator -->\nSDD orchestrator content here\n<!-- /specai:sdd-orchestrator -->\n\n<!-- specai:sdd-memory-protocol -->\nSddMemory protocol content here\n<!-- /specai:sdd-memory-protocol -->\n"
	if err := os.WriteFile(path, []byte(withSections), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Now switch to neutral persona
	result, err := Inject(home, opencodeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(neutral) should report changed")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after neutral error = %v", err)
	}
	text := string(content)

	// Neutral content should be present
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("AGENTS.md missing neutral persona content")
	}
	if strings.Contains(text, "Rioplatense") {
		t.Fatal("AGENTS.md has Rioplatense language in neutral persona — should be neutral tone")
	}

	// Managed sections MUST be preserved
	if !strings.Contains(text, "<!-- specai:sdd-orchestrator -->") {
		t.Fatal("AGENTS.md lost SDD orchestrator section after switching to neutral persona")
	}
	if !strings.Contains(text, "<!-- specai:sdd-memory-protocol -->") {
		t.Fatal("AGENTS.md lost sdd-memory protocol section after switching to neutral persona")
	}

	// Modism-specific language should be gone — neutral has the same personality but no regional language
	if strings.Contains(text, "Rioplatense") {
		t.Fatal("AGENTS.md still has Rioplatense language after switching to neutral")
	}
}

func TestInjectVSCodeNeutralPreservesManagedSections(t *testing.T) {
	home := t.TempDir()

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	_, err = Inject(home, vscodeAdapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	path := vscodeAdapter.SystemPromptFile(home)

	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	withSections := string(existing) + "\n\n<!-- specai:sdd-orchestrator -->\nSDD content\n<!-- /specai:sdd-orchestrator -->\n"
	if err := os.WriteFile(path, []byte(withSections), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = Inject(home, vscodeAdapter, model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after neutral error = %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("instructions file missing neutral persona content")
	}
	if strings.Contains(text, "Rioplatense") {
		t.Fatal("instructions file has Rioplatense language in neutral persona")
	}
	if !strings.Contains(text, "<!-- specai:sdd-orchestrator -->") {
		t.Fatal("instructions file lost SDD section after switching to neutral persona")
	}
	if !strings.Contains(text, "---\nname:") {
		t.Fatal("instructions file lost YAML frontmatter")
	}
}

func TestInjectNeutralPreservesWhenMarkerAtByteZero(t *testing.T) {
	home := t.TempDir()

	opencodeAdapter, err := agents.NewAdapter("opencode")
	if err != nil {
		t.Fatalf("NewAdapter(opencode) error = %v", err)
	}

	promptPath := opencodeAdapter.SystemPromptFile(home)
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// File starts DIRECTLY with a managed marker at byte 0 — no persona preamble.
	markerOnly := "<!-- specai:sdd-orchestrator -->\nSDD content\n<!-- /specai:sdd-orchestrator -->\n"
	if err := os.WriteFile(promptPath, []byte(markerOnly), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = Inject(home, opencodeAdapter, model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("missing neutral persona content")
	}
	if !strings.Contains(text, "<!-- specai:sdd-orchestrator -->") {
		t.Fatal("SDD section destroyed when marker was at byte 0")
	}
}

func TestInjectNeutralIdempotentWithManagedSections(t *testing.T) {
	home := t.TempDir()

	opencodeAdapter, err := agents.NewAdapter("opencode")
	if err != nil {
		t.Fatalf("NewAdapter(opencode) error = %v", err)
	}

	promptPath := opencodeAdapter.SystemPromptFile(home)
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Set up: neutral + managed sections
	// Simulate a file with neutral persona + managed sections.
	// Use a fingerprint from the real neutral asset so the test is realistic.
	neutralContent := assets.MustRead("generic/persona-neutral.md")
	initial := neutralContent + "\n\n<!-- specai:sdd-orchestrator -->\nSDD content\n<!-- /specai:sdd-orchestrator -->\n\n<!-- specai:sdd-memory-protocol -->\nSddMemory content\n<!-- /specai:sdd-memory-protocol -->\n"
	if err := os.WriteFile(promptPath, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// First neutral inject
	result1, err := Inject(home, opencodeAdapter, model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) first error = %v", err)
	}

	// Second neutral inject — should be idempotent
	result2, err := Inject(home, opencodeAdapter, model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) second error = %v", err)
	}

	if result2.Changed && !result1.Changed {
		t.Fatal("second neutral inject should not report changed when first didn't")
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	// Verify no duplication
	if strings.Count(text, "<!-- specai:sdd-orchestrator -->") != 1 {
		t.Fatal("SDD section duplicated after idempotent neutral inject")
	}
	if strings.Count(text, "## Rules") != 1 {
		t.Fatal("neutral persona duplicated after idempotent inject")
	}
	if strings.Count(text, "<!-- specai:sdd-memory-protocol -->") != 1 {
		t.Fatal("sdd-memory section duplicated after idempotent neutral inject")
	}
}

func TestInjectClaudeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectWindsurfIsIdempotent(t *testing.T) {
	home := t.TempDir()

	windsurfAdapter, err := agents.NewAdapter("windsurf")
	if err != nil {
		t.Fatalf("NewAdapter(windsurf) error = %v", err)
	}

	first, err := Inject(home, windsurfAdapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	promptPath := windsurfAdapter.SystemPromptFile(home)
	contentAfterFirst, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() after first inject error = %v", err)
	}

	second, err := Inject(home, windsurfAdapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true — persona was duplicated in global_rules.md")
	}

	contentAfterSecond, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile() after second inject error = %v", err)
	}

	if string(contentAfterFirst) != string(contentAfterSecond) {
		t.Fatal("global_rules.md content changed on second inject — persona was duplicated")
	}
}

func TestInjectCursorModismWritesRulesFileWithRealContent(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter, model.PersonaModism)
	if injectErr != nil {
		t.Fatalf("Inject(cursor) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatalf("Inject(cursor, specai) changed = false")
	}

	// Verify the generic persona content was used — not just neutral one-liner.
	path := filepath.Join(home, ".cursor", "rules", "specai.mdc")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("Cursor persona missing 'Senior Architect' — got neutral fallback instead of generic persona")
	}
	if !strings.Contains(text, "Contextual Skill Loading") {
		t.Fatal("Cursor persona missing contextual skill loading directive")
	}
}

func TestInjectGeminiModismWritesSystemPromptWithRealContent(t *testing.T) {
	home := t.TempDir()

	geminiAdapter, err := agents.NewAdapter("gemini-cli")
	if err != nil {
		t.Fatalf("NewAdapter(gemini-cli) error = %v", err)
	}

	result, injectErr := Inject(home, geminiAdapter, model.PersonaModism)
	if injectErr != nil {
		t.Fatalf("Inject(gemini) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(gemini, specai) changed = false")
	}

	path := filepath.Join(home, ".gemini", "GEMINI.md")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("Gemini persona missing 'Senior Architect'")
	}
	assertModismLanguageGuardrails(t, text,
		[]string{
			"Match the user's current language in your REPLY ONLY",
			"Do not switch languages unless the user does, asks you to, or you are quoting/translating content.",
			"When replying to the user in English, keep the full reply in natural English with the same warm energy.",
		},
		[]string{
			`Say "déjame verificar"`,
			"Spanish input → Rioplatense Spanish",
			"English input → same warm energy",
		},
	)
}

func TestInjectVSCodeModismWritesInstructionsFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	result, injectErr := Inject(home, vscodeAdapter, model.PersonaModism)
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(vscode, specai) changed = false")
	}

	path := vscodeAdapter.SystemPromptFile(home)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "applyTo: \"**\"") {
		t.Fatal("VS Code instructions file missing YAML frontmatter applyTo pattern")
	}
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("VS Code persona missing 'Senior Architect'")
	}
}

// --- Auto-heal tests: Claude Code stale free-text persona ---

// legacyClaudePersonaBlock simulates a Modism persona block that was written
// directly (without markers) by an old installer or manually by the user.
const legacyClaudePersonaBlock = `## Rules

- NEVER add "Co-Authored-By" or any AI attribution to commits. Use conventional commits format only.

## Personality

Senior Architect, 15+ years experience, GDE & MVP.

## Language

- Spanish input → Rioplatense Spanish.

## Behavior

- Push back when user asks for code without context.

`

func TestInjectClaudeAutoHealsStaleFreeTextPersona(t *testing.T) {
	home := t.TempDir()

	// Pre-populate CLAUDE.md with legacy persona content (no markers) followed
	// by a properly-marked section from a previous installer run.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(claudeMD), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	// Simulate a stale install: free-text persona block at top, then a different
	// marked section below (e.g., from a previous SDD install).
	stalePreamble := legacyClaudePersonaBlock + "\n<!-- specai:sdd -->\nOld SDD content.\n<!-- /specai:sdd -->\n"
	if err := os.WriteFile(claudeMD, []byte(stalePreamble), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject() should have changed the file to remove the legacy block")
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	// The file should now have the persona inside markers, not as free text.
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("CLAUDE.md missing persona marker after heal")
	}
	if !strings.Contains(text, "<!-- /specai:persona -->") {
		t.Fatal("CLAUDE.md missing persona close marker after heal")
	}

	// The existing SDD section must be preserved.
	if !strings.Contains(text, "<!-- specai:sdd -->") {
		t.Fatal("CLAUDE.md lost the sdd section during heal")
	}
	if !strings.Contains(text, "Old SDD content.") {
		t.Fatal("CLAUDE.md lost the sdd section content during heal")
	}

	// The persona content must NOT appear twice (no duplicate blocks).
	firstPersonaIdx := strings.Index(text, "Senior Architect")
	if firstPersonaIdx < 0 {
		t.Fatal("CLAUDE.md missing 'Senior Architect' persona content")
	}
	// Verify there's no second occurrence outside the markers.
	lastPersonaIdx := strings.LastIndex(text, "Senior Architect")
	if firstPersonaIdx != lastPersonaIdx {
		// It's OK if the same string appears inside the single persona marker block
		// multiple times (e.g., content + newlines), but there must not be a
		// separate free-text block also containing it.
		// Check: everything before the open marker should NOT contain "Senior Architect".
		openMarkerIdx := strings.Index(text, "<!-- specai:persona -->")
		if openMarkerIdx >= 0 && strings.Contains(text[:openMarkerIdx], "Senior Architect") {
			t.Fatal("CLAUDE.md still has 'Senior Architect' before the persona marker — legacy block not fully stripped")
		}
	}
}

func TestInjectClaudeAutoHealStalePersonaOnlyFile(t *testing.T) {
	home := t.TempDir()

	// CLAUDE.md contains ONLY the legacy persona block (no markers at all).
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(claudeMD), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(claudeMD, []byte(legacyClaudePersonaBlock), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject() should have changed the file")
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	// Must have markers now.
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("CLAUDE.md missing persona marker")
	}

	// Must NOT have the legacy free-text block before markers.
	openMarkerIdx := strings.Index(text, "<!-- specai:persona -->")
	if openMarkerIdx >= 0 {
		before := text[:openMarkerIdx]
		if strings.Contains(before, "## Rules") {
			t.Fatal("legacy '## Rules' block still present before persona marker")
		}
	}
}

func TestInjectClaudeHealDoesNotTouchNonPersonaContent(t *testing.T) {
	home := t.TempDir()

	// CLAUDE.md has user content that does NOT match persona fingerprints.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(claudeMD), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	userContent := "# My custom config\n\nI like turtles.\n"
	if err := os.WriteFile(claudeMD, []byte(userContent), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject() should write persona section")
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(content)

	// User content must be preserved.
	if !strings.Contains(text, "I like turtles.") {
		t.Fatal("user content was erased — heal was too aggressive")
	}
	// Persona section must be appended.
	if !strings.Contains(text, "<!-- specai:persona -->") {
		t.Fatal("persona section not appended")
	}
}

// --- Auto-heal tests: VSCode stale legacy path cleanup ---

func TestInjectVSCodeCleansLegacyGitHubPersonaFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Plant an old-style Modism persona file at the legacy path.
	legacyPath := filepath.Join(home, ".github", "copilot-instructions.md")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	// Old installer wrote raw persona content without YAML frontmatter.
	oldContent := "## Personality\n\nSenior Architect, 15+ years experience.\n"
	if err := os.WriteFile(legacyPath, []byte(oldContent), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	result, injectErr := Inject(home, vscodeAdapter, model.PersonaModism)
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}
	if !result.Changed {
		t.Fatal("Inject(vscode) should report changed (legacy cleanup + new file write)")
	}

	// Legacy file must be gone.
	if _, statErr := os.Stat(legacyPath); !os.IsNotExist(statErr) {
		t.Fatal("legacy ~/.github/copilot-instructions.md was NOT removed by auto-heal")
	}

	// New file must exist at the current path.
	newPath := vscodeAdapter.SystemPromptFile(home)
	content, readErr := os.ReadFile(newPath)
	if readErr != nil {
		t.Fatalf("ReadFile new path %q error = %v", newPath, readErr)
	}
	if !strings.Contains(string(content), "applyTo: \"**\"") {
		t.Fatal("new VSCode instructions file missing YAML frontmatter")
	}
}

func TestInjectVSCodePreservesNonPersonaGitHubFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Plant a .github/copilot-instructions.md that has user content (not a
	// Modism persona) — it must NOT be deleted.
	legacyPath := filepath.Join(home, ".github", "copilot-instructions.md")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	userContent := "# My custom Copilot instructions\n\nAlways be concise.\n"
	if err := os.WriteFile(legacyPath, []byte(userContent), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	_, injectErr := Inject(home, vscodeAdapter, model.PersonaModism)
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}

	// User's file must still exist.
	remaining, readErr := os.ReadFile(legacyPath)
	if readErr != nil {
		t.Fatalf("legacy user file was deleted: ReadFile error = %v", readErr)
	}
	if string(remaining) != userContent {
		t.Fatalf("user file content was modified: got %q", string(remaining))
	}
}

func TestNeutralAndModismToneSectionsMatch(t *testing.T) {
	neutral := assets.MustRead("generic/persona-neutral.md")
	specai := assets.MustRead("generic/persona-modism.md")

	extractSection := func(content, section string) string {
		idx := strings.Index(content, "## "+section)
		if idx < 0 {
			return ""
		}
		rest := content[idx:]
		nextIdx := strings.Index(rest[1:], "\n## ")
		if nextIdx < 0 {
			return rest
		}
		return rest[:nextIdx+1]
	}

	neutralTone := extractSection(neutral, "Tone")
	specaiTone := extractSection(specai, "Tone")

	if neutralTone != specaiTone {
		t.Fatalf("## Tone sections diverged:\nneutral:\n%s\nspecai:\n%s", neutralTone, specaiTone)
	}
}

func TestInjectVSCodeIdempotentAfterHeal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Plant legacy file and run inject twice — second run should be idempotent.
	legacyPath := filepath.Join(home, ".github", "copilot-instructions.md")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte("## Personality\n\nSenior Architect.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	first, err := Inject(home, vscodeAdapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("first inject should have changed")
	}

	second, err := Inject(home, vscodeAdapter, model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("second inject should be idempotent (changed = false), but changed = true")
	}
}

func TestInjectClaude_SwitchModismToNeutral_CleansOutputStyle(t *testing.T) {
	home := t.TempDir()

	// Step 1: install specai — creates output-styles/modism.md and sets outputStyle in settings.json.
	_, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	stylePath := filepath.Join(home, ".claude", "output-styles", "modism.md")
	if _, statErr := os.Stat(stylePath); os.IsNotExist(statErr) {
		t.Fatal("precondition: modism.md must exist after specai install")
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	settingsRaw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("precondition: settings.json must exist after specai install: %v", err)
	}
	var settingsBefore map[string]any
	if err := json.Unmarshal(settingsRaw, &settingsBefore); err != nil {
		t.Fatalf("precondition: unmarshal settings.json: %v", err)
	}
	if settingsBefore["outputStyle"] != "Modism" {
		t.Fatalf("precondition: outputStyle must be 'Modism', got %v", settingsBefore["outputStyle"])
	}

	// Step 2: switch to neutral — should clean both residuals.
	result, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(neutral) should report changed when cleaning specai residuals")
	}

	// output-styles/modism.md must be gone.
	if _, statErr := os.Stat(stylePath); !os.IsNotExist(statErr) {
		t.Fatal("modism.md must be removed when switching to neutral")
	}

	// outputStyle key must be absent from settings.json.
	settingsRaw, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) after neutral: %v", err)
	}
	var settingsAfter map[string]any
	if err := json.Unmarshal(settingsRaw, &settingsAfter); err != nil {
		t.Fatalf("Unmarshal settings.json after neutral: %v", err)
	}
	if _, ok := settingsAfter["outputStyle"]; ok {
		t.Fatalf("outputStyle key must be removed from settings.json after switching to neutral, got %v", settingsAfter["outputStyle"])
	}
}

func TestInjectClaude_Neutral_PreservesUserOutputStyle(t *testing.T) {
	home := t.TempDir()

	// Pre-create settings.json with a user-defined outputStyle that is NOT "Modism".
	settingsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	userSettings := `{"outputStyle": "MyCustom", "syntaxHighlightingDisabled": true}`
	settingsPath := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(userSettings), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}

	settingsRaw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(settingsRaw, &settings); err != nil {
		t.Fatalf("Unmarshal settings.json error = %v", err)
	}

	// User's custom outputStyle must be preserved.
	if settings["outputStyle"] != "MyCustom" {
		t.Fatalf("user outputStyle 'MyCustom' was modified: got %v", settings["outputStyle"])
	}
	// Other user keys must also survive.
	if settings["syntaxHighlightingDisabled"] != true {
		t.Fatal("syntaxHighlightingDisabled was lost")
	}
}

func TestInjectClaude_SwitchModismToNeutral_IsIdempotent(t *testing.T) {
	home := t.TempDir()

	// Install specai, then switch to neutral twice — second switch must be a no-op.
	_, err := Inject(home, claudeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	first, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("first neutral inject after specai should report changed")
	}

	second, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) second error = %v", err)
	}
	if second.Changed {
		t.Fatal("second neutral inject should be idempotent (no residuals to clean)")
	}
}

func TestInjectOpenCode_SwitchModismToNeutral_CleansAgentOverlay(t *testing.T) {
	home := t.TempDir()

	// Step 1: install specai — agent.specai key must appear in opencode.json.
	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	settingsRaw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("precondition: opencode.json must exist after specai install: %v", err)
	}
	var before map[string]any
	if err := json.Unmarshal(settingsRaw, &before); err != nil {
		t.Fatalf("precondition: unmarshal opencode.json: %v", err)
	}
	agentBefore, ok := before["agent"].(map[string]any)
	if !ok {
		t.Fatal("precondition: 'agent' key must be present after specai install")
	}
	if _, ok := agentBefore["modism"]; !ok {
		t.Fatal("precondition: agent.specai must be present after specai install")
	}

	// Pre-populate a user-defined agent to verify it survives the cleanup.
	agentBefore["my-custom-agent"] = map[string]any{"mode": "secondary"}
	before["agent"] = agentBefore
	before["someUserKey"] = "preserved"
	encoded, _ := json.MarshalIndent(before, "", "  ")
	if err := os.WriteFile(settingsPath, append(encoded, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() setup error = %v", err)
	}

	// Step 2: switch to neutral — agent.specai must be removed.
	result, err := Inject(home, opencodeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(neutral) should report changed when cleaning agent.specai residual")
	}

	settingsRaw, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) after neutral: %v", err)
	}
	var after map[string]any
	if err := json.Unmarshal(settingsRaw, &after); err != nil {
		t.Fatalf("Unmarshal opencode.json after neutral: %v", err)
	}

	// agent.specai must be gone.
	if agentAfter, ok := after["agent"].(map[string]any); ok {
		if _, stillPresent := agentAfter["modism"]; stillPresent {
			t.Fatal("agent.specai must be removed from opencode.json after switching to neutral")
		}
		// User-defined agent must survive.
		if _, ok := agentAfter["my-custom-agent"]; !ok {
			t.Fatal("user-defined agent 'my-custom-agent' was removed — only agent.specai should be cleaned")
		}
	}

	// Other top-level user keys must survive.
	if after["someUserKey"] != "preserved" {
		t.Fatalf("user key 'someUserKey' was lost: got %v", after["someUserKey"])
	}
}

func TestInjectKilocode_SwitchModismToNeutral_CleansAgentOverlay(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, kilocodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "kilo", "opencode.json")
	data, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(data), `"modism"`) {
		t.Fatal("precondition: kilo/opencode.json should have specai agent after Modism install")
	}

	result, err := Inject(home, kilocodeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(neutral) should report changed when cleaning up specai agent overlay")
	}

	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile kilo/opencode.json error = %v", err)
	}
	if strings.Contains(string(data), `"modism"`) {
		t.Fatal("kilo/opencode.json must not have specai agent key after switching to Neutral")
	}
}

func TestInjectOpenCode_NeutralFresh_IsNoOp(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, opencodeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) on fresh install error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if _, statErr := os.Stat(settingsPath); !os.IsNotExist(statErr) {
		data, _ := os.ReadFile(settingsPath)
		if strings.Contains(string(data), `"modism"`) {
			t.Fatal("Neutral fresh install must not create specai agent key")
		}
	}
}

func TestInjectOpenCode_ModismOnly_WritesAgentOverlay(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, opencodeAdapter(), model.PersonaModism)
	if err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}
	if !strings.Contains(string(data), `"modism"`) {
		t.Fatal("Modism install must write specai agent overlay in opencode.json")
	}
}

func TestInjectOpenCode_MalformedJSON_DoesNotPanic(t *testing.T) {
	home := t.TempDir()

	settingsDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	malformed := `{ "agent": { "modism": {invalid json`
	if err := os.WriteFile(filepath.Join(settingsDir, "opencode.json"), []byte(malformed), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, opencodeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) with malformed JSON must not error, got: %v", err)
	}
}

func TestInjectClaude_MalformedJSON_DoesNotPanic(t *testing.T) {
	home := t.TempDir()

	settingsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	malformed := `{ "outputStyle": "Modism", invalid`
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(malformed), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject(neutral) with malformed settings.json must not error, got: %v", err)
	}
}

func TestInjectKimi_SwitchModismToNeutral_NoResidualPersonaContent(t *testing.T) {
	home := t.TempDir()

	if _, err := Inject(home, kimiAdapter(), model.PersonaModism); err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	if _, err := Inject(home, kimiAdapter(), model.PersonaNeutral); err != nil {
		t.Fatalf("Inject(neutral) error = %v", err)
	}

	outputStylePath := filepath.Join(home, ".kimi", "output-style.md")
	data, err := os.ReadFile(outputStylePath)
	if err != nil {
		t.Fatalf("ReadFile(output-style.md) error = %v", err)
	}
	content := string(data)

	if len(strings.TrimSpace(content)) != 0 {
		t.Errorf("output-style.md should be empty after switching to neutral; got %d bytes:\n%s", len(content), content)
	}
	if strings.Contains(content, "Rioplatense") {
		t.Error("output-style.md still contains 'Rioplatense' after switching to neutral")
	}
	if strings.Contains(content, "Modism Output Style") {
		t.Error("output-style.md still contains 'Modism Output Style' after switching to neutral")
	}
	if strings.Contains(content, "voseo") {
		t.Error("output-style.md still contains 'voseo' after switching to neutral")
	}
}

func TestInjectForSync_OpenCodeNeutral_DoesNotCleanAgentModism(t *testing.T) {
	home := t.TempDir()

	if _, err := Inject(home, opencodeAdapter(), model.PersonaModism); err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	before, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) after install error = %v", err)
	}
	if !strings.Contains(string(before), `"modism"`) {
		t.Fatalf("opencode.json missing specai agent after install; got:\n%s", string(before))
	}

	if _, err := InjectForSync(home, opencodeAdapter(), model.PersonaNeutral); err != nil {
		t.Fatalf("InjectForSync(neutral) error = %v", err)
	}

	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) after sync error = %v", err)
	}
	if !strings.Contains(string(after), `"modism"`) {
		t.Fatalf("opencode.json lost specai agent after InjectForSync(neutral) — this is by-design and must not regress;\ngot:\n%s", string(after))
	}
}

func TestInjectForSync_ClaudeModismToNeutral_CleansOutputStyle(t *testing.T) {
	home := t.TempDir()

	if _, err := Inject(home, claudeAdapter(), model.PersonaModism); err != nil {
		t.Fatalf("Inject(specai) error = %v", err)
	}

	stylePath := filepath.Join(home, ".claude", "output-styles", "modism.md")
	if _, err := os.Stat(stylePath); os.IsNotExist(err) {
		t.Fatal("modism.md not written by Inject(specai) — precondition failed")
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}
	if !strings.Contains(string(raw), `"outputStyle"`) {
		t.Fatal("settings.json missing outputStyle after install — precondition failed")
	}

	if _, err := InjectForSync(home, claudeAdapter(), model.PersonaNeutral); err != nil {
		t.Fatalf("InjectForSync(neutral) error = %v", err)
	}

	if _, err := os.Stat(stylePath); !os.IsNotExist(err) {
		t.Fatal("modism.md still present after InjectForSync(neutral) — residue not cleaned")
	}

	afterRaw, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadFile(settings.json) after sync error = %v", err)
	}
	if strings.Contains(string(afterRaw), `"outputStyle"`) {
		t.Fatal("settings.json still has outputStyle key after InjectForSync(neutral) — residue not cleaned")
	}
}
