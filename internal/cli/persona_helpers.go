package cli

import "github.com/KevG1t/specai/internal/model"

func isGentlemanConversationPersona(persona model.PersonaID) bool {
	return persona == model.PersonaGentleman || persona == model.PersonaGentlemanNeutralArtifacts
}
