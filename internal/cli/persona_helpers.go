package cli

import "github.com/KevG1t/SpecAI/internal/model"

func isArgentinaConversationPersona(persona model.PersonaID) bool {
	return persona == model.PersonaArgentina
}
