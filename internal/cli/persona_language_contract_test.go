package cli

import (
	"testing"

	"github.com/KevG1t/specai/internal/model"
)

func TestNormalizePersonaAcceptsModismNeutralArtifacts(t *testing.T) {
	got, err := normalizePersona("modism-neutral-artifacts")
	if err != nil {
		t.Fatalf("normalizePersona() error = %v", err)
	}
	if got != model.PersonaModismNeutralArtifacts {
		t.Fatalf("normalizePersona() = %q, want %q", got, model.PersonaModismNeutralArtifacts)
	}
}
