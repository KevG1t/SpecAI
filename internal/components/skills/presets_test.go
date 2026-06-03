package skills_test

import (
	"slices"
	"testing"

	"github.com/KevG1t/SpecAI/internal/components/skills"
	"github.com/KevG1t/SpecAI/internal/model"
)

func TestSkillsForPresetMinimal(t *testing.T) {
	got := skills.SkillsForPreset(model.PresetMinimal)

	// Must include P0 coding skills.
	for _, id := range []model.SkillID{model.SkillTypeScript, model.SkillClaudeDevPlatform} {
		if !slices.Contains(got, id) {
			t.Errorf("PresetMinimal missing P0 coding skill %q", id)
		}
	}

	// Must NOT include P1 coding skills.
	for _, id := range []model.SkillID{
		model.SkillReact19, model.SkillNextjs15, model.SkillTailwind4,
		model.SkillZod4, model.SkillAiSdk5, model.SkillPlaywright, model.SkillPytest,
	} {
		if slices.Contains(got, id) {
			t.Errorf("PresetMinimal should not contain P1 coding skill %q", id)
		}
	}
}

func TestSkillsForPresetFull(t *testing.T) {
	got := skills.SkillsForPreset(model.PresetFull)

	allCoding := []model.SkillID{
		model.SkillTypeScript, model.SkillClaudeDevPlatform,
		model.SkillReact19, model.SkillNextjs15, model.SkillTailwind4,
		model.SkillZod4, model.SkillAiSdk5, model.SkillPlaywright, model.SkillPytest,
	}
	for _, id := range allCoding {
		if !slices.Contains(got, id) {
			t.Errorf("PresetFull missing coding skill %q", id)
		}
	}
}

func TestSkillsForPresetCustomReturnsNil(t *testing.T) {
	if skills.SkillsForPreset(model.PresetCustom) != nil {
		t.Error("PresetCustom should return nil (user picks manually)")
	}
}

func TestAllSkillIDsIncludesCodingSkills(t *testing.T) {
	all := skills.AllSkillIDs()
	for _, id := range []model.SkillID{
		model.SkillTypeScript, model.SkillClaudeDevPlatform, model.SkillReact19,
	} {
		if !slices.Contains(all, id) {
			t.Errorf("AllSkillIDs missing %q", id)
		}
	}
}

func TestSkillsForPreset_UnknownPresetReturnsNil(t *testing.T) {
	got := skills.SkillsForPreset(model.PresetID("unknown-preset"))
	if got != nil {
		t.Errorf("expected nil for unknown preset, got slice of len %d", len(got))
	}
}

func TestSkillsForPreset_Independence(t *testing.T) {
	first := skills.SkillsForPreset(model.PresetFull)
	if len(first) == 0 {
		t.Fatal("expected non-empty slice for PresetFull")
	}
	first[0] = model.SkillID("tampered")

	second := skills.SkillsForPreset(model.PresetFull)
	if len(second) == 0 {
		t.Fatal("expected non-empty second slice")
	}
	if second[0] == model.SkillID("tampered") {
		t.Error("second call returned a slice that shares backing array with first call")
	}
}
