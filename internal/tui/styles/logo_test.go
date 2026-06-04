package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// stripANSI removes ANSI escape sequences from a string for plain-text assertions.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// TestRenderLogoIsWellFormedArt verifies the logo renders as multi-line block
// art with vertically aligned columns. The art spells "Spec AI" visually; since
// block letters carry no literal text, this asserts structural integrity rather
// than a substring (SPEC-REQ1-R1, Scenario REQ1-S1).
func TestRenderLogoIsWellFormedArt(t *testing.T) {
	raw := RenderLogo()
	plain := stripANSI(raw)
	if plain == "" {
		t.Fatal("RenderLogo() returned empty output")
	}

	lines := strings.Split(plain, "\n")
	if len(lines) != len(logoLines) {
		t.Errorf("logo line count = %d, want %d", len(lines), len(logoLines))
	}

	// Every row must share the same visible width so the columns stay aligned.
	width := len(lines[0])
	for i, line := range lines {
		if len(line) != width {
			t.Errorf("logo line %d width = %d, want %d (misaligned columns); got:\n%s",
				i, len(line), width, plain)
		}
	}
}

// TestRenderLogoNarrowTerminalNoPanic verifies RenderLogo does not panic when
// called with no terminal width hint (width=0 equivalent) (SPEC-REQ1-R3,
// Scenario REQ1-S3).
func TestRenderLogoNarrowTerminalNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderLogo() panicked: %v", r)
		}
	}()
	got := RenderLogo()
	// An empty string is acceptable for narrow terminals.
	_ = got
}

// TestRenderLogoHasColorSequences verifies that the rendered logo contains
// ANSI escape sequences when color rendering is explicitly enabled
// (Scenario REQ1-S2).
func TestRenderLogoHasColorSequences(t *testing.T) {
	// Force lipgloss to render TrueColor output regardless of the test terminal.
	renderer := lipgloss.NewRenderer(nil, termenv.WithColorCache(true))
	renderer.SetColorProfile(termenv.TrueColor)
	renderer.SetHasDarkBackground(true)
	// Override the default renderer so RenderLogo uses it.
	lipgloss.SetDefaultRenderer(renderer)
	t.Cleanup(func() {
		lipgloss.SetDefaultRenderer(lipgloss.DefaultRenderer())
	})

	raw := RenderLogo()
	if !strings.Contains(raw, "\x1b[") {
		t.Error("RenderLogo() returned output with no ANSI color sequences; expected styled output")
	}
}
