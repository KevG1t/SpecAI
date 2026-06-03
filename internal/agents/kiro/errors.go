package kiro

import (
	"fmt"

	"github.com/KevG1t/SpecAI/internal/model"
)

type AgentNotInstallableError struct {
	Agent model.AgentID
}

func (e AgentNotInstallableError) Error() string {
	return fmt.Sprintf("agent %q cannot be auto-installed; download from https://kiro.dev/downloads", e.Agent)
}
