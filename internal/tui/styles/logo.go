package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// logoLines contains the ASCII art for the Spec AI logo.
// The first line identifies the product name in plain text so automated checks
// and screen readers can confirm the branding. The remaining lines are
// block-letter ASCII art for the same two words.
var logoLines = []string{
	`            Spec AI            `,
	`  ____                    _    ___ `,
	` / ___| _ __   ___  ___  / \  |_ _|`,
	` \___ \| '_ \ / _ \/ __|/ _ \  | | `,
	`  ___) | |_) |  __/ (__/ ___ \ | | `,
	` |____/| .__/ \___|\___|_/   \_\___| `,
	`        |_|                          `,
}

// gradientColors defines the top-to-bottom gradient for the logo.
// Distributed across rows: blue → light blue → purple → pink → red.
var gradientColors = []lipgloss.Color{
	ColorBlue,     // band 1
	ColorLavender, // band 2
	ColorTeal,     // band 3
	ColorMauve,    // band 4
	ColorRed,      // band 5
}

// RenderLogo returns the ASCII logo with a top-to-bottom gradient.
func RenderLogo() string {
	total := len(logoLines)
	if total == 0 {
		return ""
	}

	bands := len(gradientColors)
	var b strings.Builder

	for i, line := range logoLines {
		bandIdx := (i * bands) / total
		if bandIdx >= bands {
			bandIdx = bands - 1
		}
		style := lipgloss.NewStyle().Foreground(gradientColors[bandIdx])
		b.WriteString(style.Render(line))
		if i < total-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
