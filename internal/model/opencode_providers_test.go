package model

import "testing"

func TestOpenCodeProviders_NonEmpty(t *testing.T) {
	if len(OpenCodeProviders) == 0 {
		t.Fatal("OpenCodeProviders must not be empty")
	}
}

func TestOpenCodeProviders_AllHaveModels(t *testing.T) {
	for _, p := range OpenCodeProviders {
		if len(p.Models) == 0 {
			t.Errorf("provider %q has no models", p.ID)
		}
	}
}

func TestOpenCodeProviders_AllIDsNonEmpty(t *testing.T) {
	for _, p := range OpenCodeProviders {
		if p.ID == "" {
			t.Error("provider has empty ID")
		}
		if p.Name == "" {
			t.Errorf("provider %q has empty Name", p.ID)
		}
		for _, m := range p.Models {
			if m.ID == "" {
				t.Errorf("provider %q has model with empty ID", p.ID)
			}
			if m.Name == "" {
				t.Errorf("provider %q model %q has empty Name", p.ID, m.ID)
			}
		}
	}
}

func TestOpenCodeProviders_AnthropicPresent(t *testing.T) {
	for _, p := range OpenCodeProviders {
		if p.ID == "anthropic" {
			return
		}
	}
	t.Fatal("anthropic provider must be present in OpenCodeProviders")
}

func TestOpenCodeProviders_SupportsEffortOnlyForO3(t *testing.T) {
	for _, p := range OpenCodeProviders {
		for _, m := range p.Models {
			if m.ID == "o3" {
				if !m.SupportsEffort {
					t.Errorf("model o3 must have SupportsEffort=true")
				}
			} else {
				if m.SupportsEffort {
					t.Errorf("model %q (provider %q) must have SupportsEffort=false", m.ID, p.ID)
				}
			}
		}
	}
}

func TestOpenCodeProviders_ExactlyFiveProviders(t *testing.T) {
	if len(OpenCodeProviders) != 5 {
		t.Errorf("expected 5 providers, got %d", len(OpenCodeProviders))
	}
}
