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

// TestRenderLogoContainsSpecAI verifies the logo text spells "Spec" and "AI"
// (SPEC-REQ1-R1, Scenario REQ1-S1).
func TestRenderLogoContainsSpecAI(t *testing.T) {
	raw := RenderLogo()
	plain := stripANSI(raw)
	upperPlain := strings.ToUpper(plain)
	if !strings.Contains(upperPlain, "SPEC") {
		t.Errorf("logo plain text does not contain 'SPEC'; got:\n%s", plain)
	}
	if !strings.Contains(upperPlain, "AI") {
		t.Errorf("logo plain text does not contain 'AI'; got:\n%s", plain)
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
