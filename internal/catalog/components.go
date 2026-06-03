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
	{ID: model.ComponentNotion, Name: "Notion MCP", Description: "Notion MCP server for workspace context"},
	{ID: model.ComponentJira, Name: "Jira MCP", Description: "Jira/Confluence MCP server for issue tracking"},
}

func MVPComponents() []Component {
	components := make([]Component, len(mvpComponents))
	copy(components, mvpComponents)
	return components
}

var presetComponents = buildPresetComponents()

func buildPresetComponents() map[model.PresetID][]model.ComponentID {
	full := make([]model.ComponentID, len(mvpComponents))
	for i, c := range mvpComponents {
		full[i] = c.ID
	}
	return map[model.PresetID][]model.ComponentID{
		model.PresetFull: full,
		model.PresetEcosystemOnly: {
			model.ComponentSDDMemory,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentNotion,
			model.ComponentJira,
			model.ComponentPersona,
			model.ComponentPermission,
		},
		model.PresetMinimal: {
			model.ComponentSDDMemory,
			model.ComponentSDD,
			model.ComponentPersona,
		},
	}
}

// ComponentsForPreset returns the base ComponentIDs for a given preset.
// The returned slice is a fresh copy each call — callers may modify it without
// affecting subsequent calls or the internal catalog state.
func ComponentsForPreset(preset model.PresetID) []model.ComponentID {
	src := presetComponents[preset]
	return append([]model.ComponentID{}, src...)
}
