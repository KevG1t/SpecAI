package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/catalog"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// DependencyTreeConfirmedMsg is emitted when the user confirms the component selection.
type DependencyTreeConfirmedMsg struct {
	Components []model.ComponentID
}

// ResolveFunc is the injectable resolver function used by DependencyTreeModel.
type ResolveFunc func(selection model.Selection) (planner.ResolvedPlan, error)

// ComponentEntry holds display state for one component row.
type ComponentEntry struct {
	ID      model.ComponentID
	Label   string
	Badge   string // "included" | "auto"
	Checked bool   // only meaningful for Custom preset
}

// DependencyTreeModel is a standalone BubbleTea model for the dependency tree screen.
type DependencyTreeModel struct {
	components   []ComponentEntry
	preset       model.PresetID
	resolverFunc ResolveFunc
	baseAgents   []model.AgentID
	cursor       int
	resolved     planner.ResolvedPlan
}

// NewDependencyTreeModel constructs a DependencyTreeModel.
// For non-Custom presets the resolved plan drives the display.
// For Custom preset the model starts from the full MVP component list, all unchecked.
func NewDependencyTreeModel(
	preset model.PresetID,
	resolved planner.ResolvedPlan,
	agents []model.AgentID,
	resolverFunc ResolveFunc,
) DependencyTreeModel {
	m := DependencyTreeModel{
		preset:       preset,
		resolverFunc: resolverFunc,
		baseAgents:   agents,
		resolved:     resolved,
	}

	if preset == model.PresetCustom {
		// Start from the full MVP catalog, all unchecked.
		mvp := catalog.MVPComponents()
		entries := make([]ComponentEntry, len(mvp))
		for i, c := range mvp {
			entries[i] = ComponentEntry{
				ID:    c.ID,
				Label: c.Name,
			}
		}
		m.components = entries
	} else {
		// Non-custom: build from the resolved plan.
		m.components = buildEntriesFromResolved(resolved)
	}

	return m
}

// buildEntriesFromResolved converts a ResolvedPlan into ComponentEntry slices.
func buildEntriesFromResolved(resolved planner.ResolvedPlan) []ComponentEntry {
	autoSet := make(map[model.ComponentID]struct{}, len(resolved.AddedDependencies))
	for _, dep := range resolved.AddedDependencies {
		autoSet[dep] = struct{}{}
	}

	entries := make([]ComponentEntry, len(resolved.OrderedComponents))
	for i, id := range resolved.OrderedComponents {
		badge := "included"
		if _, isAuto := autoSet[id]; isAuto {
			badge = "auto"
		}
		entries[i] = ComponentEntry{
			ID:      id,
			Label:   string(id),
			Badge:   badge,
			Checked: true,
		}
	}
	return entries
}

// Init returns nil — no async initialization needed.
func (m DependencyTreeModel) Init() tea.Cmd { return nil }

// Update handles navigation and confirmation.
func (m DependencyTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.components)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ":
			if m.preset == model.PresetCustom {
				m = m.toggleCustom(m.cursor)
			}
		case "enter":
			components := m.confirmedComponents()
			return m, func() tea.Msg {
				return DependencyTreeConfirmedMsg{Components: components}
			}
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// toggleCustom handles a space-bar press in Custom preset mode.
func (m DependencyTreeModel) toggleCustom(idx int) DependencyTreeModel {
	if idx < 0 || idx >= len(m.components) {
		return m
	}

	entry := m.components[idx]

	// Cannot manually uncheck an auto-dependency.
	if entry.Badge == "auto" {
		return m
	}

	// Toggle the checked state.
	m.components = cloneEntries(m.components)
	m.components[idx].Checked = !m.components[idx].Checked

	// Re-resolve to enforce dependency constraints.
	m = m.reResolve()
	return m
}

// reResolve runs the resolver with the current custom checked state and updates
// auto-dependency badges in the component list.
func (m DependencyTreeModel) reResolve() DependencyTreeModel {
	if m.resolverFunc == nil {
		return m
	}

	checked := make([]model.ComponentID, 0, len(m.components))
	for _, e := range m.components {
		if e.Checked && e.Badge != "auto" {
			checked = append(checked, e.ID)
		}
	}

	sel := model.Selection{
		Agents:     m.baseAgents,
		Components: checked,
	}

	resolved, err := m.resolverFunc(sel)
	if err != nil {
		// On resolver error, keep existing state.
		return m
	}

	m.resolved = resolved

	// Rebuild the auto-dep set from the new resolved plan.
	autoSet := make(map[model.ComponentID]struct{}, len(resolved.AddedDependencies))
	for _, dep := range resolved.AddedDependencies {
		autoSet[dep] = struct{}{}
	}

	// Rebuild the resolved-present set.
	resolvedSet := make(map[model.ComponentID]struct{}, len(resolved.OrderedComponents))
	for _, id := range resolved.OrderedComponents {
		resolvedSet[id] = struct{}{}
	}

	entries := cloneEntries(m.components)
	for i := range entries {
		id := entries[i].ID
		if _, isAuto := autoSet[id]; isAuto {
			entries[i].Badge = "auto"
			entries[i].Checked = true
		} else if _, inResolved := resolvedSet[id]; inResolved {
			entries[i].Badge = ""
			entries[i].Checked = true
		} else {
			// Not in resolved plan — clear auto badge if it had one.
			if entries[i].Badge == "auto" {
				entries[i].Badge = ""
				entries[i].Checked = false
			}
		}
	}
	m.components = entries
	return m
}

// confirmedComponents returns the ordered component IDs for the confirmation message.
func (m DependencyTreeModel) confirmedComponents() []model.ComponentID {
	if m.preset != model.PresetCustom {
		return append([]model.ComponentID(nil), m.resolved.OrderedComponents...)
	}
	out := make([]model.ComponentID, 0, len(m.components))
	for _, e := range m.components {
		if e.Checked {
			out = append(out, e.ID)
		}
	}
	return out
}

// cloneEntries returns a deep copy of a ComponentEntry slice.
func cloneEntries(src []ComponentEntry) []ComponentEntry {
	out := make([]ComponentEntry, len(src))
	copy(out, src)
	return out
}

// View renders the dependency tree screen.
func (m DependencyTreeModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Component Dependencies"))
	b.WriteString("\n\n")

	if m.preset == model.PresetCustom {
		b.WriteString(styles.SubtextStyle.Render("Select components. Auto-dependencies are locked and cannot be deselected."))
		b.WriteString("\n\n")

		for idx, entry := range m.components {
			label := string(entry.ID)
			if entry.Badge == "auto" {
				label += " " + styles.SubtextStyle.Render("[auto]")
			}
			b.WriteString(renderCheckbox(label, entry.Checked, idx == m.cursor))
		}
	} else {
		b.WriteString(styles.SubtextStyle.Render("Components that will be installed:"))
		b.WriteString("\n\n")

		for idx, entry := range m.components {
			badge := styles.SubtextStyle.Render("[" + entry.Badge + "]")
			line := "  " + string(entry.ID) + " " + badge

			if idx == m.cursor {
				b.WriteString(styles.SelectedStyle.Render(styles.Cursor+string(entry.ID)+" "+badge) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: confirm"))
	if m.preset == model.PresetCustom {
		b.WriteString(styles.HelpStyle.Render(" • space: toggle • esc: back"))
	}

	return b.String()
}
