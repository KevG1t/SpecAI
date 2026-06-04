package screens

import (
	"strings"
	"testing"
)

func TestRenderCompleteSuccessRenders(t *testing.T) {
	out := RenderComplete(CompletePayload{
		ConfiguredAgents:    1,
		InstalledComponents: 1,
	})

	if !strings.Contains(out, "Done!") {
		t.Fatalf("expected 'Done!' in output: %q", out)
	}
}
