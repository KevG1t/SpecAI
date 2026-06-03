package catalog

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

func TestComponentsForPreset(t *testing.T) {
	tests := []struct {
		name        string
		preset      model.PresetID
		wantLen     int
		wantFirst   model.ComponentID
		wantNotNil  bool
		mustContain []model.ComponentID
		mustMiss    []model.ComponentID
	}{
		{
			name:       "Full preset returns all 11 components",
			preset:     model.PresetFull,
			wantLen:    11,
			wantFirst:  model.ComponentSDDMemory,
			wantNotNil: true,
		},
		{
			name:        "EcosystemOnly returns 8 components without Theme",
			preset:      model.PresetEcosystemOnly,
			wantLen:     8,
			wantNotNil:  true,
			mustContain: []model.ComponentID{model.ComponentSDDMemory, model.ComponentSDD, model.ComponentPersona, model.ComponentNotion, model.ComponentJira},
			mustMiss:    []model.ComponentID{model.ComponentTheme},
		},
		{
			name:        "Minimal returns 3 components",
			preset:      model.PresetMinimal,
			wantLen:     3,
			wantNotNil:  true,
			mustContain: []model.ComponentID{model.ComponentSDDMemory, model.ComponentSDD, model.ComponentPersona},
		},
		{
			name:       "Custom returns empty non-nil slice",
			preset:     model.PresetCustom,
			wantLen:    0,
			wantNotNil: true,
		},
		{
			name:       "Unknown preset returns empty non-nil slice",
			preset:     model.PresetID("unknown"),
			wantLen:    0,
			wantNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComponentsForPreset(tt.preset)

			if tt.wantNotNil && got == nil {
				t.Fatal("expected non-nil slice, got nil")
			}
			if len(got) != tt.wantLen {
				t.Errorf("len=%d, want %d", len(got), tt.wantLen)
			}
			if tt.wantFirst != "" && len(got) > 0 && got[0] != tt.wantFirst {
				t.Errorf("first element=%q, want %q", got[0], tt.wantFirst)
			}
			for _, must := range tt.mustContain {
				if !componentIDIn(got, must) {
					t.Errorf("expected %q in result", must)
				}
			}
			for _, miss := range tt.mustMiss {
				if componentIDIn(got, miss) {
					t.Errorf("did not expect %q in result", miss)
				}
			}
		})
	}
}

func TestComponentsForPreset_Independence(t *testing.T) {
	first := ComponentsForPreset(model.PresetFull)
	if len(first) == 0 {
		t.Fatal("expected non-empty slice")
	}
	// Mutate the returned slice.
	first[0] = model.ComponentID("tampered")

	second := ComponentsForPreset(model.PresetFull)
	if second[0] == model.ComponentID("tampered") {
		t.Error("second call returned a slice that shares backing array with first call")
	}
}

func TestComponentsForPreset_FullDerivedFromMVP(t *testing.T) {
	presets := buildPresetComponents()
	fullIDs := presets[model.PresetFull]

	mvp := MVPComponents()
	if len(fullIDs) != len(mvp) {
		t.Fatalf("PresetFull has %d IDs, want %d (len of mvpComponents)", len(fullIDs), len(mvp))
	}

	for i, c := range mvp {
		if fullIDs[i] != c.ID {
			t.Errorf("PresetFull[%d] = %q, want %q (from mvpComponents)", i, fullIDs[i], c.ID)
		}
	}
}

func componentIDIn(list []model.ComponentID, target model.ComponentID) bool {
	for _, id := range list {
		if id == target {
			return true
		}
	}
	return false
}
