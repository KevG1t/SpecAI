package tui

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/tui/screens"
)

func TestTUIAsyncUpdates(t *testing.T) {
	// GIVEN the TUI is on the loading screen
	m := NewMainModel()

	// Simulate selecting an option (e.g., Install)
	msg := screens.OptionSelectedMsg{Option: "Install"}
	newModel, cmd := m.Update(msg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenLoading {
		t.Fatalf("expected screen to be ScreenLoading, got %v", m.currentScreen)
	}

	if cmd == nil {
		t.Fatal("expected tea.Cmd to start pipeline and wait for progress")
	}

	// WHEN the pipeline emits a progress event
	// Let's manually inject a ProgressMsg as if the channel yielded it
	progressMsg := ProgressMsg{
		TaskName: "Validando OS",
		Status:   "En progreso",
		Progress: 50.0,
	}

	newModel, _ = m.Update(progressMsg)
	m = newModel.(MainModel)

	// THEN the TUI receives it as a tea.Msg and updates the loading screen message without blocking the UI thread
	if m.latestProg.TaskName != "Validando OS" {
		t.Fatalf("expected latestProg.TaskName to be 'Validando OS', got %v", m.latestProg.TaskName)
	}

	view := m.View()
	if !strings.Contains(view, "Validando OS") {
		t.Fatalf("expected view to contain 'Validando OS', got: %s", view)
	}
	if !strings.Contains(view, "En progreso") {
		t.Fatalf("expected view to contain 'En progreso', got: %s", view)
	}

	// Test PipelineDoneMsg
	doneMsg := PipelineDoneMsg{Error: nil}
	newModel, _ = m.Update(doneMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenWelcome {
		t.Fatalf("expected screen to be ScreenWelcome after done, got %v", m.currentScreen)
	}
}
