package screens

import (
	"github.com/KevG1t/SpecAI/internal/tui/styles"
)

type BackMsg struct{}

type InstallState int

const (
	InstallStateConfirm InstallState = iota
	InstallStateRunning
	InstallStateResult
)

type StartPipelineMsg struct {
	Action string
}

// FailedStep holds the step name and a truncated error message for display.
type FailedStep struct {
	StepName string
	Err      string
}

// MissingDep holds a dependency name that was expected but not found.
type MissingDep struct {
	Name string
}

// CompletePayload carries post-install summary data used by the completion screen.
type CompletePayload struct {
	ConfiguredAgents    int
	InstalledComponents int
	FailedSteps         []FailedStep
	RollbackPerformed   bool
	MissingDeps         []MissingDep
}

type PipelineFinishedMsg struct {
	Err     error
	Payload *CompletePayload // nil means legacy behavior (no completion screen)
}

type ProgressMsg struct {
	TaskName string
	Status   string
	Progress float64
}

func renderOptions(options []string, cursor int) string {
	output := ""
	for idx, option := range options {
		if idx == cursor {
			output += styles.SelectedStyle.Render(styles.Cursor+option) + "\n"
		} else {
			output += styles.UnselectedStyle.Render("  "+option) + "\n"
		}
	}

	return output
}

func renderCheckbox(label string, checked bool, focused bool) string {
	marker := "[ ]"
	markerStyle := styles.UnselectedStyle
	if checked {
		marker = "[x]"
		markerStyle = styles.SuccessStyle
	}

	prefix := "  "
	if focused {
		prefix = styles.Cursor
		return styles.SelectedStyle.Render(prefix+markerStyle.Render(marker)+" "+label) + "\n"
	}

	return styles.UnselectedStyle.Render(prefix+markerStyle.Render(marker)+" "+label) + "\n"
}

func renderRadio(label string, selected bool, focused bool) string {
	marker := "( )"
	markerStyle := styles.UnselectedStyle
	if selected {
		marker = "(*)"
		markerStyle = styles.SelectedStyle
	}

	prefix := "  "
	if focused {
		prefix = styles.Cursor
		return styles.SelectedStyle.Render(prefix+markerStyle.Render(marker)+" "+label) + "\n"
	}

	return styles.UnselectedStyle.Render(prefix+markerStyle.Render(marker)+" "+label) + "\n"
}
