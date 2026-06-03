package sdd

import (
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/components/filemerge"
)

func readSkillContent(phase string) (string, error) {
	return assets.Read("skills/" + phase + "/SKILL.md")
}

func SharedPromptDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "prompts", "sdd")
}

var subAgentPhaseOrder = profilePhaseOrder

func SharedPromptPhases() []string {
	return ProfilePhaseOrder()
}

func WriteSharedPromptFiles(homeDir string, phaseCapabilities map[string]string) (bool, error) {
	promptDir := SharedPromptDir(homeDir)
	anyChanged := false

	for _, phase := range subAgentPhaseOrder {
		skillContent, err := readSkillContent(phase)
		if err != nil {
			return false, err
		}

		capability := "capable"
		if phaseCapabilities != nil {
			if cap, ok := phaseCapabilities[phase]; ok && cap != "" {
				capability = cap
			}
		}

		content := extractModelSection(skillContent, capability)

		path := filepath.Join(promptDir, phase+".md")
		result, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
		if err != nil {
			return false, err
		}

		if result.Changed {
			anyChanged = true
		}
	}

	return anyChanged, nil
}
