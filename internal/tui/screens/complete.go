package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
)

// RenderComplete produces the post-install completion screen content as a string.
// It is a pure render function — no Bubbletea model needed for the static layout.
func RenderComplete(data CompletePayload) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Install Complete"))
	b.WriteString("\n\n")

	// Summary section.
	b.WriteString(styles.HeadingStyle.Render("Summary"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Configured agents:    %d\n", data.ConfiguredAgents))
	b.WriteString(fmt.Sprintf("  Installed components: %d\n", data.InstalledComponents))

	if data.RollbackPerformed {
		b.WriteString("\n")
		b.WriteString(styles.WarningStyle.Render("  Rollback was performed due to errors."))
		b.WriteString("\n")
	}

	// Failed steps section.
	if len(data.FailedSteps) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("Errors"))
		b.WriteString("\n")
		for _, fs := range data.FailedSteps {
			errDetail := fs.Err
			if len(errDetail) > 120 {
				errDetail = errDetail[:120] + "..."
			}
			b.WriteString(fmt.Sprintf("  %s %s\n", styles.ErrorStyle.Render("✗"), fs.StepName))
			b.WriteString(fmt.Sprintf("    %s\n", styles.SubtextStyle.Render(errDetail)))
		}
	}

	// Missing dependencies section.
	if len(data.MissingDeps) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.WarningStyle.Render("Missing dependencies"))
		b.WriteString("\n")
		for _, dep := range data.MissingDeps {
			b.WriteString(fmt.Sprintf("  - %s\n", dep.Name))
		}
	}

	// Auth guidance section.
	if len(data.AuthGuidance) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.HeadingStyle.Render("Auth setup required"))
		b.WriteString("\n")
		for _, g := range data.AuthGuidance {
			for _, line := range strings.Split(g, "\n") {
				b.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	// Pipeline warnings section.
	if len(data.ValidationWarnings) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.WarningStyle.Render("Warnings"))
		b.WriteString("\n")
		for _, w := range data.ValidationWarnings {
			b.WriteString(fmt.Sprintf("  - %s\n", w.Message))
		}
	}

	// Next steps section — always visible.
	b.WriteString("\n")
	b.WriteString(styles.HeadingStyle.Render("Next steps"))
	b.WriteString("\n")
	b.WriteString("  1. Open your IDE and start a new conversation.\n")
	b.WriteString("  2. Use /sdd-init to initialize SDD in your project.\n")
	b.WriteString("  3. Use /sdd-new <change> to start a new change.\n")
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("Run `specai --help` for all available commands."))
	b.WriteString("\n")

	return b.String()
}
