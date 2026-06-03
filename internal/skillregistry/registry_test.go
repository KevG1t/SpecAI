package skillregistry_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/skillregistry"
)

// writeSkillFile creates a SKILL.md with minimal frontmatter under dir/name/SKILL.md.
func writeSkillFile(t *testing.T, dir, name, description string) string {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", skillDir, err)
	}
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n## Body\n"
	path := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	return path
}

// TestRegenerate_CreatesRegistry verifies that Regenerate writes .atl/skill-registry.md.
func TestRegenerate_CreatesRegistry(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()

	// Create a skill inside the project so it is scanned.
	writeSkillFile(t, filepath.Join(cwd, "skills"), "my-skill", "A test skill")

	result, err := skillregistry.Regenerate(cwd, home, false)
	if err != nil {
		t.Fatalf("Regenerate() error: %v", err)
	}
	if !result.Regenerated {
		t.Fatal("expected Regenerated=true on first run")
	}
	if result.SkillCount < 1 {
		t.Errorf("SkillCount = %d, want >= 1", result.SkillCount)
	}

	registryPath := filepath.Join(cwd, skillregistry.RegistryRelPath)
	data, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("cannot read registry: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Skill Registry") {
		t.Error("registry must contain '# Skill Registry' header")
	}
	if !strings.Contains(content, "my-skill") {
		t.Errorf("registry must list 'my-skill', got:\n%s", content)
	}
}

// TestRegenerate_CacheHit verifies that a second run without changes is a no-op.
func TestRegenerate_CacheHit(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()

	writeSkillFile(t, filepath.Join(cwd, "skills"), "cached-skill", "Cached")

	// First run — must regenerate.
	if _, err := skillregistry.Regenerate(cwd, home, false); err != nil {
		t.Fatalf("first Regenerate() error: %v", err)
	}

	// Second run — same files, must be a cache hit.
	result, err := skillregistry.Regenerate(cwd, home, false)
	if err != nil {
		t.Fatalf("second Regenerate() error: %v", err)
	}
	if result.Regenerated {
		t.Error("expected Regenerated=false on cache hit")
	}
	if result.Reason != "cache-hit" {
		t.Errorf("Reason = %q, want 'cache-hit'", result.Reason)
	}
}

// TestRegenerate_ForceBypassesCache verifies that force=true always regenerates.
func TestRegenerate_ForceBypassesCache(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()

	writeSkillFile(t, filepath.Join(cwd, "skills"), "force-skill", "Forced")

	// First run.
	if _, err := skillregistry.Regenerate(cwd, home, false); err != nil {
		t.Fatalf("first Regenerate() error: %v", err)
	}

	// Second run with force.
	result, err := skillregistry.Regenerate(cwd, home, true)
	if err != nil {
		t.Fatalf("forced Regenerate() error: %v", err)
	}
	if !result.Regenerated {
		t.Error("expected Regenerated=true when force=true")
	}
	if result.Reason != "forced" {
		t.Errorf("Reason = %q, want 'forced'", result.Reason)
	}
}

// TestLoadSkill_ParsesFrontmatter verifies frontmatter parsing.
func TestLoadSkill_ParsesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillFile(t, dir, "test-skill", "Does useful things")

	entry, ok := skillregistry.LoadSkill(path)
	if !ok {
		t.Fatal("LoadSkill() returned ok=false")
	}
	if entry.Name != "test-skill" {
		t.Errorf("Name = %q, want 'test-skill'", entry.Name)
	}
	if entry.Description != "Does useful things" {
		t.Errorf("Description = %q, want 'Does useful things'", entry.Description)
	}
}

// TestFingerprint_Deterministic verifies the same file list produces the same fingerprint.
func TestFingerprint_Deterministic(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillFile(t, dir, "fp-skill", "FP test")

	files := []string{path}
	fp1 := skillregistry.Fingerprint(files)
	fp2 := skillregistry.Fingerprint(files)
	if fp1 != fp2 {
		t.Errorf("Fingerprint not deterministic: %q != %q", fp1, fp2)
	}
	if fp1 == "" {
		t.Error("Fingerprint must not be empty")
	}
}

// TestRenderRegistry_ContainsExpectedSections verifies the rendered output structure.
func TestRenderRegistry_ContainsExpectedSections(t *testing.T) {
	cwd := t.TempDir()
	entries := []skillregistry.SkillEntry{
		{Name: "my-skill", Path: "/home/user/.claude/skills/my-skill/SKILL.md", Description: "Does stuff"},
	}
	output := skillregistry.RenderRegistry(cwd, []string{".claude/skills"}, entries)

	for _, want := range []string{
		"# Skill Registry",
		"## Sources scanned",
		"## Contract",
		"## Skills",
		"## Loading protocol",
		"my-skill",
		"Does stuff",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("RenderRegistry output missing %q", want)
		}
	}
	if !strings.Contains(output, "specai") {
		t.Error("RenderRegistry output must reference 'specai'")
	}
}
