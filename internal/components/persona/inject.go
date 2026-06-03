package persona

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/components/filemerge"
	"github.com/KevG1t/SpecAI/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

type bootstrapper interface {
	BootstrapTemplate(homeDir string) error
}

var outputStyleOverlayJSON = []byte("{\n  \"outputStyle\": \"Argentina\"\n}\n")

var openCodeAgentOverlayJSON = []byte("{\n  \"agent\": {\n    \"argentina\": {\n      \"mode\": \"primary\",\n      \"description\": \"Senior Architect mentor - helpful first, challenging when it matters\",\n      \"prompt\": \"{file:./AGENTS.md}\",\n      \"tools\": {\n        \"write\": true,\n        \"edit\": true\n      }\n    }\n  }\n}\n")

func Inject(homeDir string, adapter agents.Adapter, persona model.PersonaID) (InjectionResult, error) {
	return injectInternal(homeDir, adapter, persona, false)
}

func InjectForSync(homeDir string, adapter agents.Adapter, persona model.PersonaID) (InjectionResult, error) {
	return injectInternal(homeDir, adapter, persona, true)
}

func injectInternal(homeDir string, adapter agents.Adapter, persona model.PersonaID, syncManaged bool) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}
	if err := validateOpenClawWorkspacePath(homeDir, adapter); err != nil {
		return InjectionResult{}, err
	}

	if persona == model.PersonaCustom {
		return InjectionResult{}, nil
	}

	files := make([]string, 0, 3)
	changed := false

	content := personaContent(adapter.Agent(), persona)
	if content == "" {
		return InjectionResult{}, nil
	}

	if adapter.Agent() == model.AgentOpenClaw {
		return injectOpenClawSoulPersona(homeDir, content)
	}

	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		promptPath := adapter.SystemPromptFile(homeDir)
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, err
		}

		healed := filemerge.StripLegacyPersonaBlock(existing)
		healed = filemerge.StripLegacyATLBlock(healed)
		updated := filemerge.InjectMarkdownSection(healed, "persona", content)

		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyFileReplace:
		promptPath := adapter.SystemPromptFile(homeDir)

		if adapter.Agent() == model.AgentOpenCode {
			existing, err := readFileOrEmpty(promptPath)
			if err != nil {
				return InjectionResult{}, err
			}

			healed := existing
			if shouldStripManagedLegacyPersona(existing) {
				healed = filemerge.StripLegacyPersonaBlock(existing)
			} else if isExactLegacyPersonaAsset(existing) {
				healed = ""
			}

			healed = filemerge.StripLegacyATLBlock(healed)
			updated := filemerge.InjectMarkdownSection(healed, "persona", content)

			writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || writeResult.Changed
			files = append(files, promptPath)
			break
		}

		existing, readErr := readFileOrEmpty(promptPath)
		if readErr != nil {
			return InjectionResult{}, readErr
		}

		if preserved, ok := preserveManagedSections(existing, content, persona); ok {
			writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(preserved), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || writeResult.Changed
			files = append(files, promptPath)
			break
		}

		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyInstructionsFile:
		promptPath := adapter.SystemPromptFile(homeDir)

		if cleaned, cleanErr := cleanLegacyVSCodePersona(homeDir); cleanErr == nil && cleaned {
			changed = true
		}

		existing, readErr := readFileOrEmpty(promptPath)
		if readErr != nil {
			return InjectionResult{}, readErr
		}

		if preserved, ok := preserveManagedSections(existing, wrapInstructionsFile(content), persona); ok {
			writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(preserved), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || writeResult.Changed
			files = append(files, promptPath)
			break
		}

		instructionsContent := wrapInstructionsFile(content)
		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(instructionsContent), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategySteeringFile:
		promptPath := adapter.SystemPromptFile(homeDir)

		existing, readErr := readFileOrEmpty(promptPath)
		if readErr != nil {
			return InjectionResult{}, readErr
		}

		var steeringContent string
		if preserved, ok := preserveManagedSections(existing, wrapSteeringFile(content), persona); ok {
			steeringContent = preserved
		} else {
			steeringContent = wrapSteeringFile(content)
		}

		if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
			return InjectionResult{}, err
		}
		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(steeringContent), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyAppendToFile:
		promptPath := adapter.SystemPromptFile(homeDir)

		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, err
		}

		healed := filemerge.StripLegacyPersonaBlock(existing)
		healed = filemerge.StripLegacyATLBlock(healed)
		updated := filemerge.InjectMarkdownSection(healed, "persona", content)

		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyJinjaModules:
		if bs, ok := adapter.(bootstrapper); ok {
			if err := bs.BootstrapTemplate(homeDir); err != nil {
				return InjectionResult{}, fmt.Errorf("bootstrap template: %w", err)
			}
			files = append(files, adapter.SystemPromptFile(homeDir))
			files = append(files, adapter.SettingsPath(homeDir))
		}

		configDir := adapter.GlobalConfigDir(homeDir)

		personaPath := filepath.Join(configDir, "persona.md")
		wr1, err := filemerge.WriteFileAtomic(personaPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || wr1.Changed
		files = append(files, personaPath)

		outputStyleContent := ""
		if isArgentinaPersona(persona) {
			outputStyleContent = assets.MustRead("kimi/output-style-argentina.md")
		}
		outputStylePath := filepath.Join(configDir, "output-style.md")
		wr2, err := filemerge.WriteFileAtomic(outputStylePath, []byte(outputStyleContent), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || wr2.Changed
		files = append(files, outputStylePath)
	}

	// 2. OpenCode/Kilocode agent definitions.
	if !syncManaged && (adapter.Agent() == model.AgentOpenCode || adapter.Agent() == model.AgentKilocode) && persona != model.PersonaCustom {
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			if isArgentinaPersona(persona) {
				agentResult, err := mergeJSONFile(settingsPath, openCodeAgentOverlayJSON)
				if err != nil {
					return InjectionResult{}, err
				}
				changed = changed || agentResult.Changed
				files = append(files, settingsPath)
			} else {
				removed, err := removeJSONNestedSubKey(settingsPath, "agent", "argentina")
				if err != nil {
					return InjectionResult{}, fmt.Errorf("clean agent.argentina from settings: %w", err)
				}
				if removed {
					changed = true
					files = append(files, settingsPath)
				}
			}
		}
	}

	// 3. Argentina persona only: write output style.
	if isArgentinaPersona(persona) && adapter.Agent() != model.AgentOpenClaw && adapter.SupportsOutputStyles() {
		outputStyleDir := adapter.OutputStyleDir(homeDir)
		if outputStyleDir != "" {
			outputStylePath := outputStyleDir + "/argentina.md"
			outputStyleContent := assets.MustRead("claude/output-style-argentina.md")

			styleResult, err := filemerge.WriteFileAtomic(outputStylePath, []byte(outputStyleContent), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || styleResult.Changed
			files = append(files, outputStylePath)
		}

		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			settingsResult, err := mergeJSONFile(settingsPath, outputStyleOverlayJSON)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || settingsResult.Changed
			files = append(files, settingsPath)
		}
	}

	// 3b. Non-argentina cleanup: remove residual argentina artifacts.
	if !isArgentinaPersona(persona) && adapter.Agent() != model.AgentOpenClaw && adapter.SupportsOutputStyles() {
		outputStyleDir := adapter.OutputStyleDir(homeDir)
		if outputStyleDir != "" {
			outputStylePath := outputStyleDir + "/argentina.md"
			styleRemoved, err := removeFileAtomic(outputStylePath)
			if err != nil {
				return InjectionResult{}, fmt.Errorf("remove argentina output style: %w", err)
			}
			if styleRemoved {
				changed = true
				files = append(files, outputStylePath)
			}
		}

		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			removed, err := removeJSONKeyIfValue(settingsPath, "outputStyle", "Argentina")
			if err != nil {
				return InjectionResult{}, fmt.Errorf("clean outputStyle from settings: %w", err)
			}
			if removed {
				changed = true
				files = append(files, settingsPath)
			}
		}
	}

	// 3c. Managed persona (Argentina/Nicaragua) + Claude Code only: merge thinking-verbs overlay.
	if isManagedPersona(persona) && adapter.Agent() == model.AgentClaudeCode {
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			tvChanged, tvErr := injectThinkingVerbsIfArgentina(persona, adapter.Agent(), settingsPath)
			if tvErr != nil {
				return InjectionResult{}, tvErr
			}
			changed = changed || tvChanged
			if tvChanged {
				files = append(files, settingsPath)
			}
		}
	}

	// 3d. Non-managed persona + Claude Code: remove thinkingVerbs key.
	if !isManagedPersona(persona) && adapter.Agent() == model.AgentClaudeCode {
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			if err := cleanupThinkingVerbsIfNonArgentina(persona, adapter.Agent(), settingsPath); err != nil {
				return InjectionResult{}, err
			}
		}
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func validateOpenClawWorkspacePath(workspaceDir string, adapter agents.Adapter) error {
	if adapter.Agent() == model.AgentOpenClaw && strings.TrimSpace(workspaceDir) == "" {
		return fmt.Errorf("openclaw workspace path is required for workspace-first injection")
	}
	return nil
}

func injectOpenClawSoulPersona(workspaceDir, content string) (InjectionResult, error) {
	soulPath := filepath.Join(workspaceDir, "SOUL.md")
	existing, err := readFileOrEmpty(soulPath)
	if err != nil {
		return InjectionResult{}, err
	}

	healed := filemerge.StripLegacyPersonaBlock(existing)
	healed = filemerge.StripLegacyATLBlock(healed)
	updated := filemerge.InjectMarkdownSection(healed, "persona", content)

	writeResult, err := filemerge.WriteFileAtomic(soulPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{soulPath}}, nil
}

func isExactLegacyPersonaAsset(existing string) bool {
	trimmed := strings.TrimSpace(existing)
	if trimmed == "" {
		return false
	}
	for _, assetPath := range []string{
		"opencode/persona-argentina.md",
		"generic/persona-argentina.md",
		"claude/persona-nicaragua.md",
		"generic/persona-nicaragua.md",
		"generic/persona-neutral.md",
	} {
		asset := strings.TrimSpace(assets.MustRead(assetPath))
		if trimmed == asset {
			return true
		}
	}
	return false
}

func shouldStripManagedLegacyPersona(existing string) bool {
	return strings.Contains(existing, "<!-- specai:persona -->")
}

func isArgentinaPersona(persona model.PersonaID) bool {
	return persona == model.PersonaArgentina
}

// isManagedPersona returns true for any persona that has its own thinking verbs
// and active personality injection (currently Argentina and Nicaragua).
func isManagedPersona(persona model.PersonaID) bool {
	return persona == model.PersonaArgentina || persona == model.PersonaNicaragua
}

func thinkingVerbsAsset(persona model.PersonaID) string {
	switch persona {
	case model.PersonaNicaragua:
		return "claude/thinking-verbs-nicaragua.json"
	default:
		return "claude/thinking-verbs-argentina.json"
	}
}

func personaContent(agent model.AgentID, persona model.PersonaID) string {
	switch persona {
	case model.PersonaNeutral:
		return assets.MustRead("generic/persona-neutral.md")
	case model.PersonaCustom:
		return ""
	case model.PersonaNicaragua:
		switch agent {
		case model.AgentClaudeCode:
			return assets.MustRead("claude/persona-nicaragua.md")
		default:
			return assets.MustRead("generic/persona-nicaragua.md")
		}
	default:
		switch agent {
		case model.AgentClaudeCode:
			return assets.MustRead("claude/persona-argentina.md")
		case model.AgentOpenCode, model.AgentKilocode:
			return assets.MustRead("opencode/persona-argentina.md")
		case model.AgentKimi:
			return assets.MustRead("kimi/persona-argentina.md")
		case model.AgentKiroIDE:
			return assets.MustRead("kiro/persona-argentina.md")
		default:
			return assets.MustRead("generic/persona-argentina.md")
		}
	}
}

func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	return filemerge.WriteFileAtomic(path, merged, 0o644)
}

var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}
	return content, nil
}

func preserveManagedSections(existing, newPersona string, persona model.PersonaID) (string, bool) {
	if existing == "" || isArgentinaPersona(persona) {
		return "", false
	}

	idx := strings.Index(existing, "<!-- specai:")
	if idx < 0 {
		return "", false
	}

	managedSuffix := existing[idx:]
	updated := newPersona
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	if idx > 0 {
		updated += "\n"
	}
	updated += managedSuffix

	return updated, true
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read file %q: %w", path, err)
	}
	return string(data), nil
}

func wrapInstructionsFile(content string) string {
	frontmatter := "---\n" +
		"name: SpecAI Persona\n" +
		"description: Teaching-oriented persona with SDD orchestration and sdd-memory protocol\n" +
		"applyTo: \"**\"\n" +
		"---\n\n"
	return frontmatter + content
}

func wrapSteeringFile(content string) string {
	frontmatter := "---\n" +
		"inclusion: always\n" +
		"---\n\n"
	return frontmatter + content
}

func isLegacyUnwrappedPersona(content string) bool {
	if strings.HasPrefix(content, "---\n") {
		return false
	}
	personaFingerprints := []string{
		"## Personality",
		"Senior Architect",
	}
	for _, fp := range personaFingerprints {
		if !strings.Contains(content, fp) {
			return false
		}
	}
	return true
}

func legacyVSCodePersonaPaths(homeDir string) []string {
	return []string{
		filepath.Join(homeDir, ".github", "copilot-instructions.md"),
	}
}

func removeFileAtomic(path string) (bool, error) {
	err := os.Remove(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// removeJSONKey removes key from the JSON object at path unconditionally.
// Returns true if the key was present and removed, false if absent or file missing.
func removeJSONKey(path, key string) (bool, error) {
	raw, err := osReadFile(path)
	if err != nil {
		return false, err
	}
	if len(raw) == 0 {
		return false, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return false, nil
	}

	if _, ok := root[key]; !ok {
		return false, nil
	}

	delete(root, key)

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal settings after key removal: %w", err)
	}
	encoded = append(encoded, '\n')

	if _, err := filemerge.WriteFileAtomic(path, encoded, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func removeJSONKeyIfValue(path, key, wantValue string) (bool, error) {
	raw, err := osReadFile(path)
	if err != nil {
		return false, err
	}
	if len(raw) == 0 {
		return false, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return false, nil
	}

	current, ok := root[key]
	if !ok {
		return false, nil
	}
	if current != wantValue {
		return false, nil
	}

	delete(root, key)

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal settings after cleanup: %w", err)
	}
	encoded = append(encoded, '\n')

	if _, err := filemerge.WriteFileAtomic(path, encoded, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func removeJSONNestedSubKey(path, parentKey, subKey string) (bool, error) {
	raw, err := osReadFile(path)
	if err != nil {
		return false, err
	}
	if len(raw) == 0 {
		return false, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return false, nil
	}

	parent, ok := root[parentKey]
	if !ok {
		return false, nil
	}
	parentMap, ok := parent.(map[string]any)
	if !ok {
		return false, nil
	}
	if _, exists := parentMap[subKey]; !exists {
		return false, nil
	}

	delete(parentMap, subKey)
	if len(parentMap) == 0 {
		delete(root, parentKey)
	} else {
		root[parentKey] = parentMap
	}

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal settings after cleanup: %w", err)
	}
	encoded = append(encoded, '\n')

	if _, err := filemerge.WriteFileAtomic(path, encoded, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// injectThinkingVerbsIfArgentina merges the persona-specific thinking-verbs JSON into the
// settings file at settingsPath when persona is a managed persona AND agent is ClaudeCode.
// Returns (true, nil) if the verbs were merged, (false, nil) if skipped.
func injectThinkingVerbsIfArgentina(persona model.PersonaID, agent model.AgentID, settingsPath string) (bool, error) {
	if !isManagedPersona(persona) || agent != model.AgentClaudeCode {
		return false, nil
	}
	if settingsPath == "" {
		return false, nil
	}

	assetPath := thinkingVerbsAsset(persona)
	tvData, err := assets.FS.ReadFile(assetPath)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", assetPath, err)
	}

	result, err := mergeJSONFile(settingsPath, tvData)
	if err != nil {
		return false, fmt.Errorf("merge thinking verbs: %w", err)
	}

	return result.Changed, nil
}

// cleanupThinkingVerbsIfNonArgentina removes the thinkingVerbs key from the settings
// file at settingsPath when persona is not a managed persona AND agent is ClaudeCode.
func cleanupThinkingVerbsIfNonArgentina(persona model.PersonaID, agent model.AgentID, settingsPath string) error {
	if isManagedPersona(persona) || agent != model.AgentClaudeCode {
		return nil
	}
	if settingsPath == "" {
		return nil
	}

	_, err := removeJSONKey(settingsPath, "thinkingVerbs")
	if err != nil {
		return fmt.Errorf("clean thinkingVerbs from settings: %w", err)
	}
	return nil
}

func cleanLegacyVSCodePersona(homeDir string) (bool, error) {
	cleaned := false
	for _, oldPath := range legacyVSCodePersonaPaths(homeDir) {
		data, err := os.ReadFile(oldPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return cleaned, fmt.Errorf("read legacy vscode persona %q: %w", oldPath, err)
		}

		if !isLegacyUnwrappedPersona(string(data)) {
			continue
		}

		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			return cleaned, fmt.Errorf("remove legacy vscode persona %q: %w", oldPath, err)
		}
		cleaned = true
	}
	return cleaned, nil
}
