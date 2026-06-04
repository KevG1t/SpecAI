package cli

import "github.com/KevG1t/specai/internal/model"

func isModismPersona(persona model.PersonaID) bool {
	return persona == model.PersonaModism || persona == model.PersonaModismNeutralArtifacts
}

