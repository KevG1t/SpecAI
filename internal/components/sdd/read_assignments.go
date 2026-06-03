package sdd

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/opencode"
)

var configurableAgentSet = buildConfigurableAgentSet()

func buildConfigurableAgentSet() map[string]bool {
	phases := opencode.ConfigurableAgentPhases()
	set := make(map[string]bool, len(phases)+1)
	for _, p := range phases {
		set[p] = true
	}
	set["sdd-orchestrator"] = true
	return set
}

func ReadCurrentProfiles(settingsPath string) ([]model.Profile, error) {
	return DetectProfiles(settingsPath)
}

func ReadCurrentModelAssignments(settingsPath string) (map[string]model.ModelAssignment, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]model.ModelAssignment{}, nil
		}
		return nil, err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return map[string]model.ModelAssignment{}, nil
	}

	agentRaw, ok := root["agent"]
	if !ok {
		return map[string]model.ModelAssignment{}, nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return map[string]model.ModelAssignment{}, nil
	}

	result := make(map[string]model.ModelAssignment)
	for name, defRaw := range agentMap {
		if !configurableAgentSet[name] {
			continue
		}
		defMap, ok := defRaw.(map[string]any)
		if !ok {
			continue
		}
		modelStr, ok := defMap["model"].(string)
		if !ok || modelStr == "" {
			continue
		}
		idx := strings.Index(modelStr, ":")
		if idx <= 0 {
			idx = strings.Index(modelStr, "/")
		}
		if idx <= 0 {
			continue
		}
		providerID := modelStr[:idx]
		modelID := modelStr[idx+1:]
		if modelID == "" {
			continue
		}
		assignmentKey := name
		effort, _ := defMap["variant"].(string)
		result[assignmentKey] = model.ModelAssignment{
			ProviderID: providerID,
			ModelID:    modelID,
			Effort:     effort,
		}
	}

	return result, nil
}
