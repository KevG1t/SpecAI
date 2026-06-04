package catalog

import "github.com/KevG1t/specai/internal/model"

type Component struct {
	ID          model.ComponentID
	Name        string
	Description string
}

var mvpComponents = []Component{
	{ID: model.ComponentSddMemory, Name: "SddMemory", Description: "Persistent cross-session memory"},
	{ID: model.ComponentSDD, Name: "SDD", Description: "Spec-driven development workflow"},
	{ID: model.ComponentSkills, Name: "Skills", Description: "Curated coding skill library"},
	{ID: model.ComponentContext7, Name: "Context7", Description: "Latest framework and library docs"},
	{ID: model.ComponentPersona, Name: "Persona", Description: "Modism, neutral or custom persona"},
	{ID: model.ComponentPermission, Name: "Permissions", Description: "Security-first defaults and guardrails"},
	{ID: model.ComponentTheme, Name: "Theme", Description: "Kanagawa theme overlay"},
	{ID: model.ComponentClaudeTheme, Name: "Claude Theme", Description: "Claude Code custom theme"},
}

func MVPComponents() []Component {
	components := make([]Component, len(mvpComponents))
	copy(components, mvpComponents)
	return components
}
