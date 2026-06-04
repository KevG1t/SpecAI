package styles

import "github.com/charmbracelet/lipgloss"

// Gemini CLI color palette.
var (
	ColorBase     = lipgloss.Color("#0D0D0D")
	ColorSurface  = lipgloss.Color("#161616")
	ColorOverlay  = lipgloss.Color("#4B4B4B")
	ColorText     = lipgloss.Color("#E8EAED")
	ColorSubtext  = lipgloss.Color("#9AA0A6")
	ColorLavender = lipgloss.Color("#8AB4F8")
	ColorGreen    = lipgloss.Color("#34A853")
	ColorPeach    = lipgloss.Color("#FBBC04")
	ColorRed      = lipgloss.Color("#EA4335")
	ColorBlue     = lipgloss.Color("#4796E4")
	ColorMauve    = lipgloss.Color("#EC4899")
	ColorYellow   = lipgloss.Color("#FBBC04")
	ColorTeal     = lipgloss.Color("#8B5CF6")
)

// Cursor is the prefix used for the currently focused item.
const Cursor = "▸ "

// Tagline returns the welcome screen tagline with the given version.
func Tagline(version string) string {
	return "SpecAI " + version + " — Ecosystem, Frameworks, Workflows"
}

// Pre-built reusable styles.
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true)

	HeadingStyle = lipgloss.NewStyle().
			Foreground(ColorMauve).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	SubtextStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorLavender).
			Padding(1, 2)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorOverlay).
			Padding(0, 1)

	ProgressFilled = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ProgressEmpty = lipgloss.NewStyle().
			Foreground(ColorOverlay)

	PercentStyle = lipgloss.NewStyle().
			Foreground(ColorPeach).
			Bold(true)
)
