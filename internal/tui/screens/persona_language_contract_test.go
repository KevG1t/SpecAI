package screens

import (
	"strings"
	"testing"

	"github.com/KevG1t/specai/internal/model"
)

func TestPersonaOptionsIncludeModismNeutralArtifacts(t *testing.T) {
	options := PersonaOptions()
	found := false
	for _, option := range options {
		if option == model.PersonaModismNeutralArtifacts {
			found = true
		}
	}
	if !found {
		t.Fatalf("PersonaOptions() = %v, missing %q", options, model.PersonaModismNeutralArtifacts)
	}
}

func TestRenderPersonaDescribesModismNeutralArtifacts(t *testing.T) {
	out := RenderPersona(model.PersonaModismNeutralArtifacts, 1)
	for _, want := range []string{
		"modism-neutral-artifacts",
		"Modism conversation",
		"English technical artifacts",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderPersona() missing %q; output:\n%s", want, out)
		}
	}
}
