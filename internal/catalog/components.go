package catalog

import "github.com/KevG1t/SpecAI/internal/model"

type Component struct {
	ID          model.ComponentID
	Name        string
	Description string
}

var mvpComponents = []Component{
	{ID: model.ComponentSDDMemory, Name: "SDD Memory", Description: "Persistent cross-session memory"},
	{ID: model.ComponentSDD, Name: "SDD", Description: "Spec-driven development workflow"},
	{ID: model.ComponentSkills, Name: "Skills", Description: "Curated coding skill library"},
	{ID: model.ComponentContext7, Name: "Context7", Description: "Latest framework and library docs"},
	{ID: model.ComponentPersona, Name: "Persona", Description: "Argentina, neutral or custom behavior"},
	{ID: model.ComponentPermission, Name: "Permissions", Description: "Security-first defaults and guardrails"},
	{ID: model.ComponentTheme, Name: "Theme", Description: "SpecAI Kanagawa theme overlay"},
	{ID: model.ComponentClaudeTheme, Name: "Claude SpecAI Theme", Description: "Claude Code SpecAI custom theme"},
	{ID: model.ComponentOpenCodeArgentinaLogo, Name: "OpenCode Argentina Logo", Description: "OpenCode home logo TUI plugin with Braille rose"},
}

func MVPComponents() []Component {
	components := make([]Component, len(mvpComponents))
	copy(components, mvpComponents)
	return components
}

// ComponentsForPreset returns the base ComponentIDs for a given preset.
// The returned slice is a fresh copy each call — callers may modify it without
// affecting subsequent calls or the internal catalog state.
func ComponentsForPreset(preset model.PresetID) []model.ComponentID {
	switch preset {
	case model.PresetFull:
		out := make([]model.ComponentID, len(mvpComponents))
		for i, c := range mvpComponents {
			out[i] = c.ID
		}
		return out
	case model.PresetEcosystemOnly:
		return []model.ComponentID{
			model.ComponentSDDMemory,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentPersona,
			model.ComponentPermission,
		}
	case model.PresetMinimal:
		return []model.ComponentID{
			model.ComponentSDDMemory,
			model.ComponentSDD,
			model.ComponentPersona,
		}
	case model.PresetCustom:
		return []model.ComponentID{}
	default:
		return []model.ComponentID{}
	}
}
