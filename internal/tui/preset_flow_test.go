package tui

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/KevG1t/specai/internal/model"
	"github.com/KevG1t/specai/internal/system"
	"github.com/KevG1t/specai/internal/tui/screens"
	tea "github.com/charmbracelet/bubbletea"
)

var updateTUIGoldens = flag.Bool("update", false, "update TUI golden files")

type flowAction struct {
	key       tea.KeyMsg
	cursor    int
	setCursor bool
	prepare   func(Model) Model
}

func TestPresetSelectionNextScreenFlowMatrix(t *testing.T) {
	tests := []struct {
		name       string
		agents     []model.AgentID
		preset     model.PresetID
		wantScreen Screen
		golden     string
	}{
		{
			name:       "full specai with opencode enters SDD mode before plugins",
			agents:     []model.AgentID{model.AgentOpenCode},
			preset:     model.PresetFullModism,
			wantScreen: ScreenSDDMode,
			golden:     "preset-full-specai-opencode-next.golden",
		},
		{
			name:       "ecosystem only with opencode enters SDD mode before plugins",
			agents:     []model.AgentID{model.AgentOpenCode},
			preset:     model.PresetEcosystemOnly,
			wantScreen: ScreenSDDMode,
			golden:     "preset-ecosystem-only-opencode-next.golden",
		},
		{
			name:       "minimal with opencode enters dependency tree",
			agents:     []model.AgentID{model.AgentOpenCode},
			preset:     model.PresetMinimal,
			wantScreen: ScreenDependencyTree,
			golden:     "preset-minimal-opencode-next.golden",
		},
		{
			name:       "custom with opencode enters component selection before plugins",
			agents:     []model.AgentID{model.AgentOpenCode},
			preset:     model.PresetCustom,
			wantScreen: ScreenDependencyTree,
			golden:     "preset-custom-opencode-next.golden",
		},
		{
			name:       "full specai without opencode enters strict TDD",
			agents:     []model.AgentID{model.AgentCursor},
			preset:     model.PresetFullModism,
			wantScreen: ScreenStrictTDD,
			golden:     "preset-full-specai-no-opencode-next.golden",
		},
		{
			name:       "ecosystem only without opencode enters strict TDD",
			agents:     []model.AgentID{model.AgentCursor},
			preset:     model.PresetEcosystemOnly,
			wantScreen: ScreenStrictTDD,
			golden:     "preset-ecosystem-only-no-opencode-next.golden",
		},
		{
			name:       "minimal without opencode enters dependency plan",
			agents:     []model.AgentID{model.AgentCursor},
			preset:     model.PresetMinimal,
			wantScreen: ScreenDependencyTree,
			golden:     "preset-minimal-no-opencode-next.golden",
		},
		{
			name:       "custom without opencode enters component selection",
			agents:     []model.AgentID{model.AgentCursor},
			preset:     model.PresetCustom,
			wantScreen: ScreenDependencyTree,
			golden:     "preset-custom-no-opencode-next.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(system.DetectionResult{}, "dev")
			m.Screen = ScreenPreset
			m.Selection.Agents = tt.agents
			m.Cursor = presetCursor(t, tt.preset)

			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			state := updated.(Model)

			if state.Screen != tt.wantScreen {
				t.Fatalf("screen = %v, want %v", state.Screen, tt.wantScreen)
			}
			assertTUIGolden(t, tt.golden, state.View())
		})
	}
}

func TestCustomPresetPostComponentFlowMatrix(t *testing.T) {
	tests := []struct {
		name       string
		agents     []model.AgentID
		components []model.ComponentID
		actions    []flowAction
		wantScreen Screen
		golden     string
	}{
		{
			name:       "opencode with SddMemory only shows review after component selection",
			agents:     []model.AgentID{model.AgentOpenCode},
			components: []model.ComponentID{model.ComponentSddMemory},
			actions:    []flowAction{{key: tea.KeyMsg{Type: tea.KeyEnter}}},
			wantScreen: ScreenReview,
			golden:     "custom-opencode-sdd-memory-next.golden",
		},
		{
			name:       "opencode with SDD reaches review after SDD and strict TDD stages",
			agents:     []model.AgentID{model.AgentOpenCode},
			components: []model.ComponentID{model.ComponentSDD},
			actions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // DependencyTree Continue -> SDDMode
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // SDDMode single -> StrictTDD
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // StrictTDD enable -> Review
			},
			wantScreen: ScreenReview,
			golden:     "custom-opencode-sdd-after-strict-next.golden",
		},
		{
			name:       "opencode with SDD and Skills reaches skill picker after strict TDD",
			agents:     []model.AgentID{model.AgentOpenCode},
			components: []model.ComponentID{model.ComponentSDD, model.ComponentSkills},
			actions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // DependencyTree Continue -> SDDMode
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // SDDMode single -> StrictTDD
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // StrictTDD enable -> SkillPicker
			},
			wantScreen: ScreenSkillPicker,
			golden:     "custom-opencode-sdd-skills-after-plugins-next.golden",
		},
		{
			name:       "no opencode with SDD and Skills reaches skill picker after strict TDD",
			agents:     []model.AgentID{model.AgentCursor},
			components: []model.ComponentID{model.ComponentSDD, model.ComponentSkills},
			actions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // DependencyTree Continue -> StrictTDD
				{key: tea.KeyMsg{Type: tea.KeyEnter}}, // StrictTDD enable -> SkillPicker
			},
			wantScreen: ScreenSkillPicker,
			golden:     "custom-no-opencode-sdd-skills-next.golden",
		},
		{
			name:       "no opencode with SddMemory only reaches review",
			agents:     []model.AgentID{model.AgentCursor},
			components: []model.ComponentID{model.ComponentSddMemory},
			actions:    []flowAction{{key: tea.KeyMsg{Type: tea.KeyEnter}}},
			wantScreen: ScreenReview,
			golden:     "custom-no-opencode-sdd-memory-next.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(system.DetectionResult{}, "dev")
			m.Screen = ScreenDependencyTree
			m.Selection.Preset = model.PresetCustom
			m.Selection.Agents = tt.agents
			m.Selection.Components = tt.components
			m.Cursor = len(screens.AllComponents())

			state := m
			for _, action := range tt.actions {
				if action.setCursor {
					state.Cursor = action.cursor
				}
				updated, _ := state.Update(action.key)
				state = updated.(Model)
			}

			if state.Screen != tt.wantScreen {
				t.Fatalf("screen = %v, want %v", state.Screen, tt.wantScreen)
			}
			assertTUIGolden(t, tt.golden, state.View())
		})
	}
}

func TestInstallNavigationRoundTrips(t *testing.T) {
	withModelCache := func(t *testing.T) {
		t.Helper()
		cacheFile := filepath.Join(t.TempDir(), "models.json")
		if err := os.WriteFile(cacheFile, []byte(`{}`), 0o644); err != nil {
			t.Fatalf("WriteFile(models cache) error = %v", err)
		}

		origStat := osStatModelCache
		osStatModelCache = func(name string) (os.FileInfo, error) {
			return os.Stat(cacheFile)
		}
		t.Cleanup(func() { osStatModelCache = origStat })
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) Model
		forwardActions []flowAction
		forwardScreens []Screen
		reverseScreens []Screen
	}{
		{
			name: "Pi-only agents fast path returns to agent selection",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenAgents
				m.Selection.Agents = []model.AgentID{model.AgentPi}
				m.Selection.Components = componentsForPreset(model.PresetFullModism, model.PersonaModism)
				m.Cursor = len(screens.AgentOptions())
				return m
			},
			forwardActions: []flowAction{{key: tea.KeyMsg{Type: tea.KeyEnter}}},
			forwardScreens: []Screen{ScreenDependencyTree},
			reverseScreens: []Screen{ScreenAgents},
		},
		{
			name: "non-custom minimal without OpenCode returns from dependency plan to preset",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenPreset
				m.Selection.Agents = []model.AgentID{model.AgentCursor}
				m.Cursor = presetCursor(t, model.PresetMinimal)
				return m
			},
			forwardActions: []flowAction{{key: tea.KeyMsg{Type: tea.KeyEnter}}},
			forwardScreens: []Screen{ScreenDependencyTree},
			reverseScreens: []Screen{ScreenPreset},
		},
		{
			name: "non-custom minimal with OpenCode goes directly to dependency tree",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenPreset
				m.Selection.Agents = []model.AgentID{model.AgentOpenCode}
				m.Cursor = presetCursor(t, model.PresetMinimal)
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenDependencyTree},
			reverseScreens: []Screen{ScreenPreset},
		},
		{
			name: "OpenCode SDD single returns through strict TDD and SDD mode to preset",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenPreset
				m.Selection.Agents = []model.AgentID{model.AgentOpenCode}
				m.Cursor = presetCursor(t, model.PresetFullModism)
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenSDDMode, ScreenStrictTDD, ScreenDependencyTree},
			reverseScreens: []Screen{ScreenStrictTDD, ScreenSDDMode, ScreenPreset},
		},
		{
			name: "OpenCode SDD multi with model cache returns through strict TDD model picker and SDD mode",
			setup: func(t *testing.T) Model {
				withModelCache(t)
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenPreset
				m.Selection.Agents = []model.AgentID{model.AgentOpenCode}
				m.Cursor = presetCursor(t, model.PresetFullModism)
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}, cursor: sddMultiCursor(t), setCursor: true},
				{
					key:       tea.KeyMsg{Type: tea.KeyEnter},
					cursor:    len(screens.ModelPickerRows()),
					setCursor: true,
					prepare: func(state Model) Model {
						state.ModelPicker.AvailableIDs = []string{"opencode"}
						return state
					},
				},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenSDDMode, ScreenModelPicker, ScreenStrictTDD, ScreenDependencyTree},
			reverseScreens: []Screen{ScreenStrictTDD, ScreenModelPicker, ScreenSDDMode, ScreenPreset},
		},
		{
			name: "non-OpenCode SDD returns through strict TDD to preset",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenPreset
				m.Selection.Agents = []model.AgentID{model.AgentCursor}
				m.Cursor = presetCursor(t, model.PresetFullModism)
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenStrictTDD, ScreenDependencyTree},
			reverseScreens: []Screen{ScreenStrictTDD, ScreenPreset},
		},
		{
			name: "custom SDD skills returns from skill picker through strict TDD to component selector",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenDependencyTree
				m.Selection.Preset = model.PresetCustom
				m.Selection.Agents = []model.AgentID{model.AgentCursor}
				m.Selection.Components = []model.ComponentID{model.ComponentSDD, model.ComponentSkills}
				m.Cursor = len(screens.AllComponents())
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenStrictTDD, ScreenSkillPicker},
			reverseScreens: []Screen{ScreenStrictTDD, ScreenDependencyTree},
		},
		{
			name: "custom OpenCode SDD skills returns from skill picker through strict TDD and SDD mode to component selector",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenDependencyTree
				m.Selection.Preset = model.PresetCustom
				m.Selection.Agents = []model.AgentID{model.AgentOpenCode}
				m.Selection.Components = []model.ComponentID{model.ComponentSDD, model.ComponentSkills}
				m.Cursor = len(screens.AllComponents())
				return m
			},
			forwardActions: []flowAction{
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
				{key: tea.KeyMsg{Type: tea.KeyEnter}},
			},
			forwardScreens: []Screen{ScreenSDDMode, ScreenStrictTDD, ScreenSkillPicker},
			reverseScreens: []Screen{ScreenStrictTDD, ScreenSDDMode, ScreenDependencyTree},
		},
		{
			name: "custom SddMemory only returns from review to component selector",
			setup: func(t *testing.T) Model {
				m := NewModel(system.DetectionResult{}, "dev")
				m.Screen = ScreenDependencyTree
				m.Selection.Preset = model.PresetCustom
				m.Selection.Agents = []model.AgentID{model.AgentCursor}
				m.Selection.Components = []model.ComponentID{model.ComponentSddMemory}
				m.Cursor = len(screens.AllComponents())
				return m
			},
			forwardActions: []flowAction{{key: tea.KeyMsg{Type: tea.KeyEnter}}},
			forwardScreens: []Screen{ScreenReview},
			reverseScreens: []Screen{ScreenDependencyTree},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup(t)
			for idx, action := range tt.forwardActions {
				state = applyFlowAction(t, state, action)
				if state.Screen != tt.forwardScreens[idx] {
					t.Fatalf("forward step %d: screen = %v, want %v", idx+1, state.Screen, tt.forwardScreens[idx])
				}
			}

			for idx, want := range tt.reverseScreens {
				state = applyFlowAction(t, state, flowAction{key: tea.KeyMsg{Type: tea.KeyEsc}})
				if state.Screen != want {
					t.Fatalf("reverse step %d: screen = %v, want %v", idx+1, state.Screen, want)
				}
			}
		})
	}
}

func TestPiOnlyDependencyTreeBackRowReturnsToAgentSelection(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenAgents
	m.Selection.Agents = []model.AgentID{model.AgentPi}
	m.Selection.Components = componentsForPreset(model.PresetFullModism, model.PersonaModism)
	m.Cursor = len(screens.AgentOptions())

	state := applyFlowAction(t, m, flowAction{key: tea.KeyMsg{Type: tea.KeyEnter}})
	if state.Screen != ScreenDependencyTree {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenDependencyTree)
	}

	state = applyFlowAction(t, state, flowAction{key: tea.KeyMsg{Type: tea.KeyEnter}, cursor: 1, setCursor: true})
	if state.Screen != ScreenAgents {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenAgents)
	}
}

func applyFlowAction(t *testing.T, state Model, action flowAction) Model {
	t.Helper()
	if action.prepare != nil {
		state = action.prepare(state)
	}
	if action.setCursor {
		state.Cursor = action.cursor
	}
	updated, _ := state.Update(action.key)
	return updated.(Model)
}

func presetCursor(t *testing.T, preset model.PresetID) int {
	t.Helper()
	for idx, option := range screens.PresetOptions() {
		if option == preset {
			return idx
		}
	}
	t.Fatalf("preset %q not found", preset)
	return 0
}

func assertTUIGolden(t *testing.T, name string, actual string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name)

	if *updateTUIGoldens {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(goldenPath), err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", goldenPath, err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", goldenPath, err)
	}
	if string(expected) != actual {
		t.Fatalf("golden mismatch for %s\n\nexpected:\n%s\n\nactual:\n%s", name, string(expected), actual)
	}
}
