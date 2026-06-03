package catalog

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

// TestMVPSkillsNoDuplicates ensures no skill is listed twice in mvpSkills.
func TestMVPSkillsNoDuplicates(t *testing.T) {
	seen := make(map[model.SkillID]bool)
	for _, s := range MVPSkills() {
		if seen[s.ID] {
			t.Errorf("duplicate skill %q in mvpSkills", s.ID)
		}
		seen[s.ID] = true
	}
}

// TestMVPSkillsContainsNewCodingSkills verifies that all 9 new coding skills
// are present in MVPSkills() with correct priority and category.
func TestMVPSkillsContainsNewCodingSkills(t *testing.T) {
	want := []struct {
		id       model.SkillID
		priority string
		category string
	}{
		{model.SkillTypeScript, "p0", "coding"},
		{model.SkillClaudeDevPlatform, "p0", "coding"},
		{model.SkillReact19, "p1", "coding"},
		{model.SkillNextjs15, "p1", "coding"},
		{model.SkillTailwind4, "p1", "coding"},
		{model.SkillZod4, "p1", "coding"},
		{model.SkillAiSdk5, "p1", "coding"},
		{model.SkillPlaywright, "p1", "coding"},
		{model.SkillPytest, "p1", "coding"},
	}

	skills := MVPSkills()
	index := make(map[model.SkillID]Skill, len(skills))
	for _, s := range skills {
		index[s.ID] = s
	}

	for _, w := range want {
		s, ok := index[w.id]
		if !ok {
			t.Errorf("skill %q not found in MVPSkills()", w.id)
			continue
		}
		if s.Priority != w.priority {
			t.Errorf("skill %q: Priority = %q, want %q", w.id, s.Priority, w.priority)
		}
		if s.Category != w.category {
			t.Errorf("skill %q: Category = %q, want %q", w.id, s.Category, w.category)
		}
	}
}

// TestMVPSkillsGoTestingExists verifies that go-testing is still present (not removed
// when adding the new coding skills).
func TestMVPSkillsGoTestingExists(t *testing.T) {
	skills := MVPSkills()
	for _, s := range skills {
		if s.ID == model.SkillGoTesting {
			return
		}
	}
	t.Fatal("go-testing skill must remain in MVPSkills()")
}
