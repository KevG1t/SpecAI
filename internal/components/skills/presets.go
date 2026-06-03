package skills

import "github.com/KevG1t/SpecAI/internal/model"

var sddSkills = []model.SkillID{
	model.SkillSDDInit,
	model.SkillSDDExplore,
	model.SkillSDDPropose,
	model.SkillSDDSpec,
	model.SkillSDDDesign,
	model.SkillSDDTasks,
	model.SkillSDDApply,
	model.SkillSDDVerify,
	model.SkillSDDArchive,
	model.SkillSDDOnboard,
	model.SkillJudgmentDay,
}

var foundationSkills = []model.SkillID{
	model.SkillGoTesting,
	model.SkillCreator,
	model.SkillImprover,
	model.SkillBranchPR,
	model.SkillIssueCreation,
	model.SkillSkillRegistry,
	model.SkillChainedPR,
	model.SkillCognitiveDoc,
	model.SkillCommentWriter,
	model.SkillWorkUnitCommits,
}

// codingP0Skills are PRD Section 6.5 P0 coding skills — included in all non-custom presets.
var codingP0Skills = []model.SkillID{
	model.SkillTypeScript,
	model.SkillClaudeDevPlatform,
}

// codingP1Skills are PRD Section 6.5 P1 coding skills — included in Full preset only.
var codingP1Skills = []model.SkillID{
	model.SkillReact19,
	model.SkillNextjs15,
	model.SkillTailwind4,
	model.SkillZod4,
	model.SkillAiSdk5,
	model.SkillPlaywright,
	model.SkillPytest,
}

var presetSkills = buildPresetSkills()

func buildPresetSkills() map[model.PresetID][]model.SkillID {
	minimal := append(append([]model.SkillID{}, sddSkills...), codingP0Skills...)
	ecosystemOnly := append(append(append([]model.SkillID{}, sddSkills...), foundationSkills...), codingP0Skills...)
	full := append(append(append(append([]model.SkillID{}, sddSkills...), foundationSkills...), codingP0Skills...), codingP1Skills...)
	return map[model.PresetID][]model.SkillID{
		model.PresetMinimal:       minimal,
		model.PresetEcosystemOnly: ecosystemOnly,
		model.PresetFull:          full,
		model.PresetCustom:        nil,
	}
}

func SkillsForPreset(preset model.PresetID) []model.SkillID {
	src := presetSkills[preset]
	if src == nil {
		return nil
	}
	return append([]model.SkillID(nil), src...)
}

func AllSkillIDs() []model.SkillID {
	all := make([]model.SkillID, 0, len(sddSkills)+len(foundationSkills)+len(codingP0Skills)+len(codingP1Skills))
	all = append(all, sddSkills...)
	all = append(all, foundationSkills...)
	all = append(all, codingP0Skills...)
	all = append(all, codingP1Skills...)
	return all
}
