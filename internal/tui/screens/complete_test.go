package screens

import (
	"strings"
	"testing"
)

func TestRenderComplete_AllSuccess(t *testing.T) {
	data := CompletePayload{
		ConfiguredAgents:    2,
		InstalledComponents: 4,
	}

	result := RenderComplete(data)

	assertContains(t, result, "Install Complete")
	assertContains(t, result, "Summary")
	assertContains(t, result, "Configured agents:    2")
	assertContains(t, result, "Installed components: 4")
	assertContains(t, result, "Next steps")
	assertContains(t, result, "/sdd-init")
	assertContains(t, result, "specai --help")

	// No error section.
	assertAbsent(t, result, "Errors")
	assertAbsent(t, result, "Rollback")
}

func TestRenderComplete_PartialFailure(t *testing.T) {
	data := CompletePayload{
		ConfiguredAgents:    1,
		InstalledComponents: 2,
		FailedSteps: []FailedStep{
			{StepName: "Inyectando MCP", Err: "sdd-memory binary not found"},
			{StepName: "Inyectando overlay", Err: "permission denied"},
		},
	}

	result := RenderComplete(data)

	assertContains(t, result, "Errors")
	assertContains(t, result, "Inyectando MCP")
	assertContains(t, result, "sdd-memory binary not found")
	assertContains(t, result, "Inyectando overlay")
	assertContains(t, result, "permission denied")
	assertContains(t, result, "Next steps")
}

func TestRenderComplete_RollbackPerformed(t *testing.T) {
	data := CompletePayload{
		ConfiguredAgents:    0,
		InstalledComponents: 0,
		RollbackPerformed:   true,
		FailedSteps: []FailedStep{
			{StepName: "Detectando IDEs", Err: "critical failure"},
		},
	}

	result := RenderComplete(data)

	assertContains(t, result, "Rollback")
	assertContains(t, result, "Errors")
	assertContains(t, result, "Next steps")
}

func TestRenderComplete_LongErrorIsTruncated(t *testing.T) {
	longErr := strings.Repeat("x", 200)
	data := CompletePayload{
		FailedSteps: []FailedStep{
			{StepName: "some step", Err: longErr},
		},
	}

	result := RenderComplete(data)

	// Truncated error must end with "..."
	if !strings.Contains(result, "...") {
		t.Errorf("long error should be truncated with '...', got:\n%s", result)
	}
	// Full 200-char error should not appear.
	if strings.Contains(result, longErr) {
		t.Error("full long error should not appear — must be truncated")
	}
}

func TestRenderComplete_NextStepsAlwaysVisible(t *testing.T) {
	// Both success and failure scenarios must show Next steps.
	for _, tc := range []struct {
		name string
		data CompletePayload
	}{
		{"all-success", CompletePayload{ConfiguredAgents: 2, InstalledComponents: 3}},
		{"with-errors", CompletePayload{FailedSteps: []FailedStep{{StepName: "s", Err: "e"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := RenderComplete(tc.data)
			assertContains(t, result, "Next steps")
		})
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("output missing %q\nGot:\n%s", needle, haystack)
	}
}

func assertAbsent(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("output unexpectedly contains %q\nGot:\n%s", needle, haystack)
	}
}
