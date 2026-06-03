package sdd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/components/filemerge"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/opencode"
)

var profileNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

var reservedProfileNames = func() map[string]bool {
	names := map[string]bool{
		"default":          true,
		"sdd-orchestrator": true,
	}
	for _, name := range opencode.JDPhases() {
		names[name] = true
	}
	return names
}()

func ValidateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name must not be empty")
	}
	if reservedProfileNames[name] {
		return fmt.Errorf("profile name %q is reserved", name)
	}
	if !profileNameRegex.MatchString(name) {
		return fmt.Errorf("profile name %q must match ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ (lowercase, hyphens only, no trailing hyphens, no underscores or spaces)", name)
	}
	return nil
}

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"sdd-onboard",
}

func ProfilePhaseOrder() []string {
	return append([]string(nil), profilePhaseOrder...)
}

func ResolveProfileStrategy(homeDir string, explicit model.SDDProfileStrategyID) model.SDDProfileStrategyID {
	if explicit != "" {
		return explicit
	}
	if HasExternalProfileFiles(homeDir) {
		return model.SDDProfileStrategyExternalSingleActive
	}
	return model.SDDProfileStrategyGeneratedMulti
}

func HasExternalProfileFiles(homeDir string) bool {
	if strings.TrimSpace(homeDir) == "" {
		return false
	}

	profilesDir := filepath.Join(homeDir, ".config", "opencode", "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			return true
		}
	}

	return false
}

func ProfileAgentKeys(name string) []string {
	suffix := ""
	if name != "" {
		suffix = "-" + name
	}

	keys := make([]string, 0, 11)
	keys = append(keys, "sdd-orchestrator"+suffix)
	for _, phase := range profilePhaseOrder {
		keys = append(keys, phase+suffix)
	}
	return keys
}

func DetectProfiles(settingsPath string) ([]model.Profile, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Profile{}, nil
		}
		return nil, fmt.Errorf("read settings %q: %w", settingsPath, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse settings %q: %w", settingsPath, err)
	}

	agentRaw, ok := root["agent"]
	if !ok {
		return []model.Profile{}, nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return []model.Profile{}, nil
	}

	const orchPrefix = "sdd-orchestrator-"
	profileNames := make([]string, 0)
	seen := make(map[string]bool)
	for key := range agentMap {
		if !strings.HasPrefix(key, orchPrefix) {
			continue
		}
		profileName := key[len(orchPrefix):]
		if profileName == "" || seen[profileName] {
			continue
		}
		seen[profileName] = true
		profileNames = append(profileNames, profileName)
	}

	if len(profileNames) == 0 {
		return []model.Profile{}, nil
	}

	sort.Strings(profileNames)

	profiles := make([]model.Profile, 0, len(profileNames))
	for _, profileName := range profileNames {
		orchKey := "sdd-orchestrator-" + profileName
		orchRaw := agentMap[orchKey]
		orchMap, _ := orchRaw.(map[string]any)

		orchModel := extractModelFromAgent(orchMap)
		phaseAssignments := make(map[string]model.ModelAssignment)
		for _, phase := range profilePhaseOrder {
			agentKey := phase + "-" + profileName
			agentRaw := agentMap[agentKey]
			agentMap2, _ := agentRaw.(map[string]any)
			if m := extractModelFromAgent(agentMap2); m.ProviderID != "" {
				phaseAssignments[phase] = m
			}
		}

		profiles = append(profiles, model.Profile{
			Name:              profileName,
			OrchestratorModel: orchModel,
			PhaseAssignments:  phaseAssignments,
		})
	}

	return profiles, nil
}

func extractModelFromAgent(agentMap map[string]any) model.ModelAssignment {
	if agentMap == nil {
		return model.ModelAssignment{}
	}
	modelStr, _ := agentMap["model"].(string)
	if modelStr == "" {
		return model.ModelAssignment{}
	}

	idx := strings.Index(modelStr, ":")
	if idx <= 0 {
		idx = strings.Index(modelStr, "/")
	}
	if idx <= 0 {
		return model.ModelAssignment{}
	}
	providerID := modelStr[:idx]
	modelID := modelStr[idx+1:]
	if modelID == "" {
		return model.ModelAssignment{}
	}
	effort, _ := agentMap["variant"].(string)
	return model.ModelAssignment{ProviderID: providerID, ModelID: modelID, Effort: effort}
}

func GenerateProfileOverlay(profile model.Profile, homeDir string) ([]byte, error) {
	if profile.Name == "" || profile.Name == "default" {
		return nil, fmt.Errorf("GenerateProfileOverlay: profile name must be non-empty and not 'default'")
	}

	suffix := "-" + profile.Name
	orchestratorKey := "sdd-orchestrator" + suffix

	orchestratorPrompt, err := buildProfileOrchestratorPrompt(profile)
	if err != nil {
		return nil, fmt.Errorf("build orchestrator prompt for profile %q: %w", profile.Name, err)
	}

	agentMap := make(map[string]any, 11)

	taskPerms := map[string]any{
		"*": "deny",
	}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range opencode.JDPhases() {
		taskPerms[jd] = "allow"
	}

	orchEntry := map[string]any{
		"mode":        "primary",
		"description": "SDD Orchestrator (" + profile.Name + " profile) - coordinates sub-agents, never does work inline",
		"prompt":      orchestratorPrompt,
		"permission": map[string]any{
			"task": map[string]any{
				"__replace__": taskPerms,
			},
		},
		"tools": map[string]any{
			"read":            true,
			"write":           true,
			"edit":            true,
			"bash":            true,
			"delegate":        true,
			"delegation_read": true,
			"delegation_list": true,
		},
	}
	if profile.OrchestratorModel.ProviderID != "" && profile.OrchestratorModel.ModelID != "" {
		orchEntry["model"] = profile.OrchestratorModel.FullID()
		if profile.OrchestratorModel.Effort != "" {
			orchEntry["variant"] = profile.OrchestratorModel.Effort
		} else {
			orchEntry["variant"] = ""
		}
	}
	agentMap[orchestratorKey] = orchEntry

	promptDir := SharedPromptDir(homeDir)
	phaseDescriptions := map[string]string{
		"sdd-init":    "Bootstrap SDD context and project configuration",
		"sdd-explore": "Investigate codebase and think through ideas",
		"sdd-propose": "Create change proposals from explorations",
		"sdd-spec":    "Write detailed specifications from proposals",
		"sdd-design":  "Create technical design from proposals",
		"sdd-tasks":   "Break down specs and designs into implementation tasks",
		"sdd-apply":   "Implement code changes from task definitions",
		"sdd-verify":  "Validate implementation against specs",
		"sdd-archive": "Archive completed change artifacts",
		"sdd-onboard": "Guide user through a complete SDD cycle using their real codebase",
	}

	for _, phase := range profilePhaseOrder {
		key := phase + suffix
		prompt := "{file:" + filepath.ToSlash(filepath.Join(promptDir, phase+".md")) + "}"
		entry := map[string]any{
			"mode":        "subagent",
			"hidden":      true,
			"description": phaseDescriptions[phase],
			"prompt":      prompt,
			"tools": map[string]any{
				"read":  true,
				"write": true,
				"edit":  true,
				"bash":  true,
			},
		}
		if assignment, ok := profile.PhaseAssignments[phase]; ok && assignment.ProviderID != "" && assignment.ModelID != "" {
			entry["model"] = assignment.FullID()
			if assignment.Effort != "" {
				entry["variant"] = assignment.Effort
			} else {
				entry["variant"] = ""
			}
		}
		agentMap[key] = entry
	}

	overlay := map[string]any{
		"agent": agentMap,
	}

	result, err := json.MarshalIndent(overlay, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal profile overlay: %w", err)
	}
	return append(result, '\n'), nil
}

func buildProfileOrchestratorPrompt(profile model.Profile) (string, error) {
	base := assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))

	capability := "capable"
	if profile.OrchestratorModel.ModelID != "" {
		capability = model.ModelCapability(profile.OrchestratorModel.ModelID)
	}
	base = extractModelSection(base, capability)

	const openMarker = "<!-- specai:sdd-model-assignments -->"
	const closeMarker = "<!-- /specai:sdd-model-assignments -->"

	start := strings.Index(base, openMarker)
	end := strings.Index(base, closeMarker)
	if start != -1 && end != -1 && end > start {
		table := renderProfileModelAssignmentsSection(profile)
		afterOpen := start + len(openMarker)
		base = base[:afterOpen] + "\n" + table + base[end:]
	}

	suffix := "-" + profile.Name
	for _, phase := range profilePhaseOrder {
		base = replacePhaseRef(base, phase, phase+suffix)
	}
	base = replacePhaseRef(base, "sdd-orchestrator", "sdd-orchestrator"+suffix)

	return base, nil
}

func extractModelSection(content, capability string) string {
	openMarker := "<!-- section:model-" + capability + " -->"
	closeMarker := "<!-- /section:model-" + capability + " -->"
	start := strings.Index(content, openMarker)
	end := strings.Index(content, closeMarker)
	if start == -1 || end == -1 || end <= start {
		return content
	}
	afterOpen := start + len(openMarker)
	return strings.TrimLeft(content[afterOpen:end], " \t\r\n")
}

func replacePhaseRef(content, from, to string) string {
	suffix := strings.TrimPrefix(to, from)
	if suffix == "" {
		return content
	}

	var sb strings.Builder
	remaining := content
	for {
		idx := strings.Index(remaining, from)
		if idx < 0 {
			sb.WriteString(remaining)
			break
		}
		afterIdx := idx + len(from)
		if afterIdx <= len(remaining) && strings.HasPrefix(remaining[afterIdx:], suffix) {
			sb.WriteString(remaining[:afterIdx])
			remaining = remaining[afterIdx:]
			continue
		}
		sb.WriteString(remaining[:idx])
		sb.WriteString(to)
		remaining = remaining[afterIdx:]
	}
	return sb.String()
}

func renderProfileModelAssignmentsSection(profile model.Profile) string {
	var b strings.Builder
	b.WriteString("## Model Assignments\n\n")
	b.WriteString("Read this table at session start (or before first delegation) and cache it for the session. Treat each row as the authoritative configured model for that agent. If a phase is missing, use the default OpenCode runtime model and continue.\n\n")
	b.WriteString("| Phase | Model | Reason |\n")
	b.WriteString("|-------|-------|--------|\n")

	orchModel := "—"
	if profile.OrchestratorModel.ProviderID != "" {
		orchModel = profile.OrchestratorModel.FullID()
	}
	b.WriteString(fmt.Sprintf("| orchestrator | %s | Coordinates, makes decisions |\n", orchModel))

	phaseReasons := map[string]string{
		"sdd-init":    "Bootstrap SDD context",
		"sdd-explore": "Reads code, structural - not architectural",
		"sdd-propose": "Architectural decisions",
		"sdd-spec":    "Structured writing",
		"sdd-design":  "Architecture decisions",
		"sdd-tasks":   "Mechanical breakdown",
		"sdd-apply":   "Implementation",
		"sdd-verify":  "Validation against spec",
		"sdd-archive": "Copy and close",
		"sdd-onboard": "Guided walkthrough",
	}

	for _, phase := range profilePhaseOrder {
		phaseModel := "—"
		if m, ok := profile.PhaseAssignments[phase]; ok && m.ProviderID != "" {
			phaseModel = m.FullID()
		}
		reason := phaseReasons[phase]
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", phase, phaseModel, reason))
	}
	b.WriteString("\n")
	return b.String()
}

func RemoveProfileAgents(settingsPath string, profileName string) error {
	if profileName == "" || profileName == "default" {
		return fmt.Errorf("RemoveProfileAgents: cannot remove default profile (name=%q)", profileName)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read settings %q: %w", settingsPath, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse settings %q: %w", settingsPath, err)
	}

	agentRaw, ok := root["agent"]
	if !ok {
		return nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return nil
	}

	keysToDelete := ProfileAgentKeys(profileName)
	deleted := 0
	for _, key := range keysToDelete {
		if _, exists := agentMap[key]; exists {
			delete(agentMap, key)
			deleted++
		}
	}

	if deleted == 0 {
		return nil
	}

	root["agent"] = agentMap
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	out = append(out, '\n')

	_, err = filemerge.WriteFileAtomic(settingsPath, out, 0o644)
	return err
}
