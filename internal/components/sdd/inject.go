package sdd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/components/filemerge"
	"github.com/KevG1t/SpecAI/internal/components/skills"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/opencode"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

type InjectOptions struct {
	OpenCodeModelAssignments map[string]model.ModelAssignment
	ClaudeModelAssignments   map[string]model.ClaudeModelAlias
	KiroModelAssignments     map[string]model.ClaudeModelAlias

	WorkspaceDir string

	// StrictTDD enables Strict TDD mode marker injection.
	StrictTDD bool

	Profiles []model.Profile

	// PreserveOpenCodeOrchestratorPrompt keeps the existing
	// opencode.json agent.sdd-orchestrator.prompt value during sync.
	PreserveOpenCodeOrchestratorPrompt bool

	Capability string
}

type workflowInjector interface {
	SupportsWorkflows() bool
	WorkflowsDir(workspaceDir string) string
	EmbeddedWorkflowsDir() string
}

type kiroModelResolver interface {
	KiroModelID(alias model.ClaudeModelAlias) string
}

type claudeModelResolver interface {
	ClaudeModelID(alias model.ClaudeModelAlias) string
}

var monorepoRootMarkers = []string{
	"pnpm-workspace.yaml",
	"pnpm-workspace.yml",
	"nx.json",
	"turbo.json",
	"lerna.json",
	"rush.json",
}

var strongProjectMarkers = []string{
	".git",
	"go.mod",
	"Cargo.toml",
	"pyproject.toml",
	"pom.xml",
	"build.gradle",
}

const maxAncestorDepth = 20

type bootstrapper interface {
	BootstrapTemplate(homeDir string) error
}

func findProjectRoot(dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	current := filepath.Clean(dir)
	var bestCandidate string

	for i := 0; i < maxAncestorDepth; i++ {
		for _, marker := range monorepoRootMarkers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current, true
			}
		}
		for _, marker := range strongProjectMarkers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current, true
			}
		}
		if _, err := os.Stat(filepath.Join(current, "package.json")); err == nil {
			bestCandidate = current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	if bestCandidate != "" {
		return bestCandidate, true
	}
	return "", false
}

var (
	npmLookPath = exec.LookPath
	npmRun      = func(dir string, args ...string) ([]byte, error) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		return cmd.CombinedOutput()
	}
)

func overlayAssetPath(sddMode model.SDDModeID) string {
	if sddMode == model.SDDModeMulti {
		return "opencode/sdd-overlay-multi.json"
	}
	return "opencode/sdd-overlay-single.json"
}

func Inject(homeDir string, adapter agents.Adapter, sddMode model.SDDModeID, options ...InjectOptions) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}
	if err := validateOpenClawWorkspacePath(homeDir, adapter); err != nil {
		return InjectionResult{}, err
	}

	var opts InjectOptions
	if len(options) > 0 {
		opts = options[0]
	}

	files := make([]string, 0)
	changed := false

	// 1. Inject SDD orchestrator into global system prompt.
	if adapter.Agent() != model.AgentOpenCode && adapter.Agent() != model.AgentKilocode {
		switch adapter.SystemPromptStrategy() {
		case model.StrategyMarkdownSections:
			result, err := injectMarkdownSections(homeDir, adapter, opts.ClaudeModelAssignments)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || result.Changed
			files = append(files, result.Files...)

		case model.StrategyFileReplace, model.StrategyAppendToFile, model.StrategyInstructionsFile, model.StrategySteeringFile:
			result, err := injectFileAppend(homeDir, adapter)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || result.Changed
			files = append(files, result.Files...)

		case model.StrategyJinjaModules:
			if bs, ok := adapter.(bootstrapper); ok {
				if err := bs.BootstrapTemplate(homeDir); err != nil {
					return InjectionResult{}, fmt.Errorf("bootstrap template: %w", err)
				}
			}

			configDir := adapter.GlobalConfigDir(homeDir)
			content := assets.MustRead(sddOrchestratorAsset(adapter.Agent()))
			modulePath := filepath.Join(configDir, "sdd-orchestrator.md")
			writeResult, err := filemerge.WriteFileAtomic(modulePath, []byte(content), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || writeResult.Changed
			files = append(files, modulePath)
		}
	}

	// 1b. Inject strict-tdd-mode marker if enabled.
	if opts.StrictTDD && adapter.Agent() != model.AgentOpenCode && adapter.Agent() != model.AgentKilocode {
		if adapter.SystemPromptStrategy() == model.StrategyJinjaModules {
			configDir := adapter.GlobalConfigDir(homeDir)
			content := "Strict TDD Mode: enabled"
			modulePath := filepath.Join(configDir, "strict-tdd-mode.md")
			writeResult, err := filemerge.WriteFileAtomic(modulePath, []byte(content), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || writeResult.Changed
			files = append(files, modulePath)
		} else {
			promptPath := adapter.SystemPromptFile(homeDir)
			strictTDDContent := "Strict TDD Mode: enabled"
			existing, readErr := readFileOrEmpty(promptPath)
			if readErr != nil {
				return InjectionResult{}, readErr
			}
			updated := filemerge.InjectMarkdownSection(existing, "strict-tdd-mode", strictTDDContent)
			writeResult, writeErr := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
			if writeErr != nil {
				return InjectionResult{}, writeErr
			}
			changed = changed || writeResult.Changed
			alreadyInFiles := false
			for _, f := range files {
				if f == promptPath {
					alreadyInFiles = true
					break
				}
			}
			if !alreadyInFiles {
				files = append(files, promptPath)
			}
		}
	}

	// 2. Write slash commands.
	if adapter.SupportsSlashCommands() {
		commandsDir := adapter.CommandsDir(homeDir)
		if commandsDir != "" {
			commandsAssetDir := assets.SDDCommandsAssetDir(adapter.Agent())
			commandEntries, err := fs.ReadDir(assets.FS, commandsAssetDir)
			if err != nil {
				return InjectionResult{}, fmt.Errorf("read embedded %s: %w", commandsAssetDir, err)
			}

			for _, entry := range commandEntries {
				if entry.IsDir() {
					continue
				}

				content := assets.MustRead(commandsAssetDir + "/" + entry.Name())
				path := filepath.Join(commandsDir, entry.Name())
				writeResult, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
				if err != nil {
					return InjectionResult{}, err
				}

				changed = changed || writeResult.Changed
				files = append(files, path)
			}
		}
	}

	// 2b. OpenCode overlay injection.
	var mergedSettingsBytes []byte
	if adapter.Agent() == model.AgentOpenCode || adapter.Agent() == model.AgentKilocode {
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			overlayContent, err := assets.Read(overlayAssetPath(sddMode))
			if err != nil {
				return InjectionResult{}, fmt.Errorf("read SDD overlay asset: %w", err)
			}

			overlayBytes := []byte(overlayContent)
			if sddMode == model.SDDModeMulti {
				phaseCapabilities := make(map[string]string)
				for phase, assignment := range opts.OpenCodeModelAssignments {
					phaseCapabilities[phase] = model.ModelCapability(assignment.ModelID)
				}
				for _, profile := range opts.Profiles {
					for phase, assignment := range profile.PhaseAssignments {
						if assignment.ModelID != "" {
							phaseCapabilities[phase] = model.ModelCapability(assignment.ModelID)
						}
					}
				}
				promptsChanged, promptsErr := WriteSharedPromptFiles(homeDir, phaseCapabilities)
				if promptsErr != nil {
					return InjectionResult{}, fmt.Errorf("write shared SDD prompt files: %w", promptsErr)
				}
				changed = changed || promptsChanged
			}

			overlayBytes, err = inlineOpenCodeSDDPrompts(overlayBytes, homeDir, settingsPath, opts.PreserveOpenCodeOrchestratorPrompt)
			if err != nil {
				return InjectionResult{}, fmt.Errorf("inline OpenCode SDD prompts: %w", err)
			}
			assignments := opts.OpenCodeModelAssignments
			if sddMode != model.SDDModeMulti {
				assignments = nil
			}

			var rootModelID string
			var existingAgentKeys map[string]bool
			if sddMode == model.SDDModeMulti {
				rootModelID, err = readOpenCodeRootModel(settingsPath)
				if err != nil {
					return InjectionResult{}, err
				}
				existingAgentKeys, err = readExistingAgentModels(settingsPath)
				if err != nil {
					return InjectionResult{}, err
				}
			}

			if sddMode == model.SDDModeMulti && (len(assignments) > 0 || rootModelID != "") {
				overlayBytes, err = injectModelAssignments(overlayBytes, assignments, rootModelID, existingAgentKeys)
				if err != nil {
					return InjectionResult{}, fmt.Errorf("inject model assignments: %w", err)
				}
			}

			overlayBytes, err = defaultOpenCodeShareDisabled(settingsPath, overlayBytes)
			if err != nil {
				return InjectionResult{}, fmt.Errorf("default OpenCode share mode: %w", err)
			}

			agentResult, err := mergeJSONFile(settingsPath, overlayBytes)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || agentResult.writeResult.Changed
			files = append(files, settingsPath)
			mergedSettingsBytes = agentResult.merged

			pluginResult, err := installOpenCodePlugins(homeDir, adapter)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || pluginResult.Changed
			files = append(files, pluginResult.Files...)

			for _, profile := range opts.Profiles {
				if profile.Name == "" || profile.Name == "default" {
					continue
				}
				profileOverlay, profileErr := GenerateProfileOverlay(profile, homeDir)
				if profileErr != nil {
					return InjectionResult{}, fmt.Errorf("generate profile overlay %q: %w", profile.Name, profileErr)
				}
				profileResult, profileErr := mergeJSONFile(settingsPath, profileOverlay)
				if profileErr != nil {
					return InjectionResult{}, fmt.Errorf("merge profile overlay %q: %w", profile.Name, profileErr)
				}
				changed = changed || profileResult.writeResult.Changed
				mergedSettingsBytes = profileResult.merged
			}
		}
	}

	// 3. Write SDD skill files.
	if adapter.SupportsSkills() {
		skillDir := adapter.SkillsDir(homeDir)
		if skillDir != "" {
			sharedFiles := []string{
				"SKILL.md",
				"persistence-contract.md",
				"sdd-memory-convention.md",
				"openspec-convention.md",
				"sdd-phase-common.md",
				"skill-resolver.md",
			}
			sddSkillIDs := []model.SkillID{
				"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec",
				"sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive",
				"sdd-onboard", "judgment-day",
			}

			for _, fileName := range sharedFiles {
				assetPath := "skills/_shared/" + fileName
				content, readErr := assets.Read(assetPath)
				if readErr != nil {
					return InjectionResult{}, fmt.Errorf("required SDD shared file %q: embedded asset not found: %w", fileName, readErr)
				}
				if len(content) == 0 {
					return InjectionResult{}, fmt.Errorf("required SDD shared file %q: embedded asset is empty", fileName)
				}

				path := filepath.Join(skillDir, "_shared", fileName)
				writeResult, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
				if err != nil {
					return InjectionResult{}, err
				}

				changed = changed || writeResult.Changed
				files = append(files, path)
			}

			capability := opts.Capability
			if capability == "" {
				capability = "capable"
			}
			sddResult, sddErr := skills.InjectWithCapability(homeDir, adapter, sddSkillIDs, capability)
			if sddErr != nil {
				return InjectionResult{}, fmt.Errorf("inject SDD skills: %w", sddErr)
			}
			changed = changed || sddResult.Changed
			files = append(files, sddResult.Files...)
		}
	}

	// 3b. Write native workflow files.
	if wi, ok := adapter.(workflowInjector); ok && wi.SupportsWorkflows() {
		if projectRoot, found := findProjectRoot(opts.WorkspaceDir); found {
			workflowsDir := wi.WorkflowsDir(projectRoot)
			embedDir := wi.EmbeddedWorkflowsDir()
			entries, readErr := fs.ReadDir(assets.FS, embedDir)
			if readErr != nil {
				return InjectionResult{}, fmt.Errorf("read embedded %s: %w", embedDir, readErr)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				content, readErr := assets.Read(embedDir + "/" + entry.Name())
				if readErr != nil {
					return InjectionResult{}, fmt.Errorf("read embedded workflow %q: %w", entry.Name(), readErr)
				}
				path := filepath.Join(workflowsDir, entry.Name())
				writeResult, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
				if err != nil {
					return InjectionResult{}, fmt.Errorf("write workflow %q: %w", path, err)
				}
				changed = changed || writeResult.Changed
				files = append(files, path)
			}
		}
	}

	// 3c. Write native sub-agent files.
	var agentsDir string
	if adapter.SupportsSubAgents() {
		agentsDir = adapter.SubAgentsDir(homeDir)
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			return InjectionResult{}, fmt.Errorf("create agents dir: %w", err)
		}

		embeddedDir := adapter.EmbeddedSubAgentsDir()
		entries, err := assets.FS.ReadDir(embeddedDir)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("read embedded agents dir: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			contentStr := assets.MustRead(embeddedDir + "/" + entry.Name())

			if kmr, ok := adapter.(kiroModelResolver); ok {
				phase := strings.TrimSuffix(entry.Name(), ".md")
				alias := model.ClaudeModelSonnet
				if opts.KiroModelAssignments != nil {
					if a, hasAlias := opts.KiroModelAssignments[phase]; hasAlias {
						alias = a
					} else if d, hasDefault := opts.KiroModelAssignments["default"]; hasDefault {
						alias = d
					}
				} else if opts.ClaudeModelAssignments != nil {
					if a, hasAlias := opts.ClaudeModelAssignments[phase]; hasAlias {
						alias = a
					} else if d, hasDefault := opts.ClaudeModelAssignments["default"]; hasDefault {
						alias = d
					}
				}
				contentStr = strings.ReplaceAll(contentStr, "{{KIRO_MODEL}}", kmr.KiroModelID(alias))
			}

			if cmr, ok := adapter.(claudeModelResolver); ok {
				phase := strings.TrimSuffix(entry.Name(), ".md")
				alias := resolveClaudeModelAlias(opts.ClaudeModelAssignments, phase)
				contentStr = strings.ReplaceAll(contentStr, "{{CLAUDE_MODEL}}", cmr.ClaudeModelID(alias))
			}
			outPath := filepath.Join(agentsDir, entry.Name())
			writeResult, err := filemerge.WriteFileAtomic(outPath, []byte(contentStr), 0o644)
			if err != nil {
				return InjectionResult{}, fmt.Errorf("write agent %s: %w", entry.Name(), err)
			}
			changed = changed || writeResult.Changed
			if writeResult.Changed {
				files = append(files, outPath)
			}
		}

		for _, phase := range []string{"sdd-apply", "sdd-verify"} {
			found := false
			for _, ext := range []string{".md", ".yaml"} {
				checkPath := filepath.Join(agentsDir, phase+ext)
				if info, err := os.Stat(checkPath); err == nil && info.Size() >= 10 {
					found = true
					break
				}
			}
			if !found {
				return InjectionResult{}, fmt.Errorf("post-check: sub-agent %q not written correctly (missing or truncated)", phase)
			}
		}
	}

	// 4. Install skill-registry startup automation.
	automationResult, err := installSkillRegistryAutomation(homeDir, adapter)
	if err != nil {
		return InjectionResult{}, err
	}
	changed = changed || automationResult.Changed
	files = append(files, automationResult.Files...)

	// 5. Post-injection verification.
	if adapter.Agent() == model.AgentOpenCode {
		settingsPath := adapter.SettingsPath(homeDir)
		settingsText := string(mergedSettingsBytes)

		if len(mergedSettingsBytes) == 0 {
			if diskBytes, readErr := os.ReadFile(settingsPath); readErr == nil {
				settingsText = string(diskBytes)
			}
		}

		if !hasOpenCodeAgentKey(settingsText, "sdd-orchestrator") {
			if diskBytes, readErr := os.ReadFile(settingsPath); readErr == nil {
				settingsText = string(diskBytes)
			}
			if !hasOpenCodeAgentKey(settingsText, "sdd-orchestrator") {
				return InjectionResult{}, fmt.Errorf("post-check: %q missing sdd-orchestrator agent definition — OpenCode /sdd-* commands will fail", settingsPath)
			}
		}
		if sddMode == model.SDDModeMulti && !strings.Contains(settingsText, `"sdd-apply"`) {
			if diskBytes, readErr := os.ReadFile(settingsPath); readErr == nil {
				settingsText = string(diskBytes)
			}
			if !strings.Contains(settingsText, `"sdd-apply"`) {
				return InjectionResult{}, fmt.Errorf("post-check: %q missing sdd-apply sub-agent — multi-mode overlay was not injected correctly", settingsPath)
			}
		}

		for _, profile := range opts.Profiles {
			if profile.Name == "" || profile.Name == "default" {
				continue
			}
			orchKey := `"sdd-orchestrator-` + profile.Name + `"`
			if !strings.Contains(settingsText, orchKey) {
				if diskBytes, readErr := os.ReadFile(settingsPath); readErr == nil {
					settingsText = string(diskBytes)
				}
				if !strings.Contains(settingsText, orchKey) {
					return InjectionResult{}, fmt.Errorf("post-check: %q missing profile orchestrator %q — profile overlay was not injected correctly", settingsPath, "sdd-orchestrator-"+profile.Name)
				}
			}
		}
	}

	if adapter.SupportsSkills() {
		skillDir := adapter.SkillsDir(homeDir)
		if skillDir != "" {
			for _, skill := range []string{"sdd-init", "sdd-apply", "sdd-verify"} {
				path := filepath.Join(skillDir, skill, "SKILL.md")
				info, err := os.Stat(path)
				if err != nil {
					return InjectionResult{}, fmt.Errorf("post-check: SDD skill %q not found on disk: %w", skill, err)
				}
				if info.Size() < 100 {
					return InjectionResult{}, fmt.Errorf("post-check: SDD skill %q is too small (%d bytes) — content may be empty or corrupt", skill, info.Size())
				}
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

func inlineOpenCodeSDDPrompts(overlayBytes []byte, homeDir, settingsPath string, preserveExistingOrchestratorPrompt bool) ([]byte, error) {
	var overlay map[string]any
	if err := json.Unmarshal(overlayBytes, &overlay); err != nil {
		return nil, fmt.Errorf("unmarshal OpenCode SDD overlay: %w", err)
	}

	agentsRaw, ok := overlay["agent"]
	if !ok {
		return overlayBytes, nil
	}
	agentsMap, ok := agentsRaw.(map[string]any)
	if !ok {
		return overlayBytes, nil
	}

	orchestratorRaw, ok := agentsMap["sdd-orchestrator"]
	if !ok {
		return overlayBytes, nil
	}
	orchestratorMap, ok := orchestratorRaw.(map[string]any)
	if !ok {
		return overlayBytes, nil
	}
	if preserveExistingOrchestratorPrompt {
		existingPrompt, err := readOpenCodeAgentPrompt(settingsPath, "sdd-orchestrator")
		if err != nil {
			return nil, err
		}
		if existingPrompt != "" {
			orchestratorMap["prompt"] = migratePreservedOpenCodeOrchestratorPrompt(existingPrompt)
		} else {
			orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))
		}
	} else {
		orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))
	}

	if homeDir != "" {
		promptDir := SharedPromptDir(homeDir)
		for _, phase := range subAgentPhaseOrder {
			agentRaw, exists := agentsMap[phase]
			if !exists {
				continue
			}
			agentMap, ok := agentRaw.(map[string]any)
			if !ok {
				continue
			}
			placeholder := "__PROMPT_FILE_" + phase + "__"
			if prompt, _ := agentMap["prompt"].(string); prompt == placeholder {
				agentMap["prompt"] = "{file:" + filepath.ToSlash(filepath.Join(promptDir, phase+".md")) + "}"
			}
		}
	}

	result, err := json.MarshalIndent(overlay, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal OpenCode SDD overlay: %w", err)
	}

	return append(result, '\n'), nil
}

func migratePreservedOpenCodeOrchestratorPrompt(prompt string) string {
	if prompt == "" {
		return prompt
	}

	return ensurePreservedOpenCodeOrchestratorPreflight(prompt)
}

func ensurePreservedOpenCodeOrchestratorPreflight(prompt string) string {
	preflight := `

<!-- specai:sdd-session-preflight-migration -->
### SDD Session Preflight (HARD GATE)

Before executing ANY SDD command or natural-language SDD request, ensure this session has an explicit ` + "`SDD Session Preflight`" + ` decision block.

Required preflight choices: execution mode, artifact store, chained PR strategy, and review budget.

Ask the user directly with a compact, numbered preflight prompt. Match the user's current language for all user-facing prose. If the user writes Spanish, ask the preflight in Spanish. Keep option codes (` + "`A1`" + `, ` + "`B1`" + `, ` + "`C1`" + `, ` + "`D1`" + `) and canonical values unchanged. Do NOT ask the user to type raw keys like ` + "`execution mode`" + `, ` + "`artifact store`" + `, ` + "`chained PR strategy`" + `, or ` + "`review budget`" + `. Do NOT mention non-existent tools. Do NOT invent informal values; use only the canonical values after the user chooses.

Do NOT mix languages inside one preflight prompt: headings, option titles, descriptions, and follow-up text must all be in the user's current language.

` + "```text" + `
Before continuing with SDD, choose one option per group.
Reply with "use recommended" or with codes like: A1, B1, C1, D1.

A. Pace
   A1 Interactive (recommended): show each phase and wait for confirmation before continuing.
   A2 Automatic: run phases back-to-back and stop only on high risk.

B. Artifacts
   B1 OpenSpec (recommended): repo files, traceable in review.
   B2 sdd-memory: faster, no spec files in the repo.
   B3 Both: OpenSpec files plus sdd-memory copy.

C. PRs
   C1 Ask me (recommended): stop and ask if the forecast exceeds the budget.
   C2 Single PR: try to keep the change in one PR.
   C3 Chained: split into chained PRs from the start.
   C4 Auto: decide from the size forecast.

D. Review
   D1 400 lines (recommended): stop if forecast exceeds 400 changed lines.
   D2 800 lines: more permissive; useful for medium changes.
   D3 Other: ask for the number afterwards.
` + "```" + `

After asking this, STOP and wait for the user's answer.

If the user's current language is Spanish, use this localized shape:

` + "```text" + `
Antes de continuar con SDD, elija una opción por grupo.
Responda con "usar recomendado" o con códigos como: A1, B1, C1, D1.

A. Ritmo
   A1 Interactivo (recomendado): mostrar cada fase y esperar confirmación antes de continuar.
   A2 Automático: ejecutar las fases seguidas y frenar solo ante riesgo alto.

B. Artefactos
   B1 OpenSpec (recomendado): archivos en el repo, trazables en revisión.
   B2 sdd-memory: más rápido, sin archivos de especificación en el repo.
   B3 Ambos: archivos OpenSpec más copia en sdd-memory.

C. PRs
   C1 Preguntarme (recomendado): frenar y preguntar si la estimación supera el presupuesto.
   C2 Un solo PR: intentar mantener el cambio en un PR.
   C3 Encadenados: separar en PRs encadenados desde el inicio.
   C4 Auto: decidir según la estimación de tamaño.

D. Revisión
   D1 400 líneas (recomendado): frenar si la estimación supera 400 líneas cambiadas.
   D2 800 líneas: más permisivo; útil para cambios medianos.
   D3 Otro: preguntar el número después.
` + "```" + `

Map answers to canonical values: A1/Interactive -> ` + "`interactive`" + `; A2/Automatic -> ` + "`auto`" + `; B1/OpenSpec -> ` + "`openspec`" + `; B2/sdd-memory -> ` + "`sdd-memory`" + `; B3/Both -> ` + "`both`" + `; C1/Ask me -> ` + "`ask-always`" + `; C2/Single PR -> ` + "`single-pr-default`" + `; C3/Chained -> ` + "`force-chained`" + `; C4/Auto -> ` + "`auto-forecast`" + `; D1/400 lines -> ` + "`review_budget_lines: 400`" + `; D2/800 lines -> ` + "`review_budget_lines: 800`" + `; D3/Other -> ask one follow-up for the number.

Hard gate rules:

- ` + "`openspec/config.yaml`" + `, existing SDD artifacts, previous ` + "`sdd-init`" + ` results, or installed SDD assets do NOT satisfy session preflight.
- If the session has no preflight block, ask the localized user-facing preflight prompt above and STOP. Do not run init, delegate phases, edit files, or apply tasks in the same turn.
- For a new feature request that says to use SDD, start at preflight -> init guard -> explore/proposal. Never launch ` + "`sdd-apply`" + ` just because the user asked to implement a feature.
- In ` + "`interactive`" + ` mode, pause after each delegated phase returns, summarize the phase, ask before launching the next phase, and STOP. Match the user's language and active persona for direct conversation only. Do not run /sdd-ff phases back-to-back unless execution mode is ` + "`auto`" + `.
<!-- /specai:sdd-session-preflight-migration -->
`

	if strings.Contains(prompt, "### SDD Session Preflight (HARD GATE)") &&
		strings.Contains(prompt, "openspec/config.yaml") &&
		strings.Contains(prompt, "Never launch `sdd-apply`") &&
		strings.Contains(prompt, "Match the user's current language") &&
		strings.Contains(prompt, "pause after each delegated phase returns") &&
		strings.Contains(prompt, "Before continuing with SDD") {
		return prompt
	}

	start := "<!-- specai:sdd-session-preflight-migration -->"
	end := "<!-- /specai:sdd-session-preflight-migration -->"
	if startIdx := strings.Index(prompt, start); startIdx >= 0 {
		if relEndIdx := strings.Index(prompt[startIdx:], end); relEndIdx >= 0 {
			endIdx := startIdx + relEndIdx + len(end)
			return strings.TrimRight(prompt[:startIdx], "\n") + preflight + prompt[endIdx:]
		}
	}

	return strings.TrimRight(prompt, "\n") + preflight
}

func readOpenCodeAgentPrompt(settingsPath, agentKey string) (string, error) {
	if strings.TrimSpace(settingsPath) == "" || strings.TrimSpace(agentKey) == "" {
		return "", nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read OpenCode settings %q: %w", settingsPath, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return "", nil
	}

	agentsRaw, ok := root["agent"]
	if !ok {
		return "", nil
	}
	agentsMap, ok := agentsRaw.(map[string]any)
	if !ok {
		return "", nil
	}
	agentRaw, ok := agentsMap[agentKey]
	if !ok {
		return "", nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return "", nil
	}
	prompt, _ := agentMap["prompt"].(string)
	return prompt, nil
}


func installSkillRegistryAutomation(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if adapter.Agent() != model.AgentClaudeCode {
		return InjectionResult{}, nil
	}
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		return InjectionResult{}, nil
	}
	changed, err := ensureClaudeSkillRegistryHook(settingsPath)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("install Claude skill-registry hook: %w", err)
	}
	return InjectionResult{Changed: changed, Files: []string{settingsPath}}, nil
}

func ensureClaudeSkillRegistryHook(settingsPath string) (bool, error) {
	root := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &root); err != nil {
			return false, fmt.Errorf("parse Claude settings %q: %w", settingsPath, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	const command = `specai skill-registry refresh --quiet --no-gitignore --cwd "${CLAUDE_PROJECT_DIR:-$PWD}" || true`
	if claudeHookExists(root, command) {
		return false, nil
	}

	hooksRaw, hasHooks := root["hooks"]
	hooksMap, _ := hooksRaw.(map[string]any)
	if hasHooks && hooksMap == nil {
		return false, fmt.Errorf("Claude settings %q has unsupported hooks shape: want object", settingsPath)
	}
	if hooksMap == nil {
		hooksMap = map[string]any{}
	}
	promptRaw, hasUserPromptSubmit := hooksMap["UserPromptSubmit"]
	userPromptSubmit, _ := promptRaw.([]any)
	if hasUserPromptSubmit && userPromptSubmit == nil {
		return false, fmt.Errorf("Claude settings %q has unsupported hooks.UserPromptSubmit shape: want array", settingsPath)
	}
	userPromptSubmit = append(userPromptSubmit, map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": command,
			},
		},
	})
	hooksMap["UserPromptSubmit"] = userPromptSubmit
	root["hooks"] = hooksMap

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, err
	}
	out = append(out, '\n')
	wr, err := filemerge.WriteFileAtomic(settingsPath, out, 0o644)
	if err != nil {
		return false, err
	}
	return wr.Changed, nil
}

func claudeHookExists(root map[string]any, command string) bool {
	hooksMap, ok := root["hooks"].(map[string]any)
	if !ok {
		return false
	}
	for _, key := range []string{"UserPromptSubmit", "SessionStart"} {
		hookEntries, ok := hooksMap[key].([]any)
		if !ok {
			continue
		}
		if claudeHookListContains(hookEntries, command) {
			return true
		}
	}
	return false
}

func claudeHookListContains(hookEntries []any, command string) bool {
	for _, item := range hookEntries {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		hooks, ok := itemMap["hooks"].([]any)
		if !ok {
			continue
		}
		for _, hook := range hooks {
			hookMap, ok := hook.(map[string]any)
			if ok && hookMap["command"] == command {
				return true
			}
		}
	}
	return false
}

func installOpenCodePlugins(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	opencodeDir := adapter.GlobalConfigDir(homeDir)
	pluginsDir := filepath.Join(opencodeDir, "plugins")

	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return InjectionResult{}, fmt.Errorf("create plugins dir: %w", err)
	}

	var files []string
	var changed bool

	for _, name := range []string{"background-agents.ts", "model-variants.ts"} {
		content := assets.MustRead("opencode/plugins/" + name)
		pluginPath := filepath.Join(pluginsDir, name)

		writeResult, err := filemerge.WriteFileAtomic(pluginPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("write plugin %s: %w", name, err)
		}

		files = append(files, pluginPath)
		if writeResult.Changed {
			changed = true
		}
	}

	depPkg := "unique-names-generator"
	nmPath := filepath.Join(opencodeDir, "node_modules", depPkg)

	pkgMissing := false
	pkgMgrRan := false
	if _, statErr := os.Stat(nmPath); os.IsNotExist(statErr) {
		pkgMissing = true
		var installErr error
		pkgMgrRan, installErr = runPkgInstall(opencodeDir, depPkg)
		if installErr != nil {
			return InjectionResult{}, installErr
		}
	}

	if pkgMissing && pkgMgrRan {
		if _, statErr := os.Stat(nmPath); os.IsNotExist(statErr) {
			return InjectionResult{}, fmt.Errorf(
				"post-install check: %q was not found after install in %q — "+
					"the background-agents plugin will fail to load.\n"+
					"Fix: run `cd %s && bun add %s` (or npm install %s) manually",
				depPkg, nmPath, opencodeDir, depPkg, depPkg,
			)
		}
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func runPkgInstall(dir, pkg string) (ran bool, err error) {
	if bunPath, lookErr := npmLookPath("bun"); lookErr == nil {
		out, runErr := npmRun(dir, bunPath, "add", pkg)
		if runErr != nil {
			return true, fmt.Errorf(
				"bun add %s failed in %s: %w\nOutput: %s\nFix: run `cd %s && bun add %s` manually",
				pkg, dir, runErr, strings.TrimSpace(string(out)), dir, pkg,
			)
		}
		return true, nil
	}

	if npmPath, lookErr := npmLookPath("npm"); lookErr == nil {
		out, runErr := npmRun(dir, npmPath, "install", "--save", pkg)
		if runErr != nil {
			return true, fmt.Errorf(
				"npm install %s failed in %s: %w\nOutput: %s\nFix: run `cd %s && npm install %s` manually",
				pkg, dir, runErr, strings.TrimSpace(string(out)), dir, pkg,
			)
		}
		return true, nil
	}

	return false, nil
}

type mergeJSONResult struct {
	writeResult filemerge.WriteResult
	merged      []byte
}

func mergeJSONFile(path string, overlay []byte) (mergeJSONResult, error) {
	baseJSON, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return mergeJSONResult{}, fmt.Errorf("read json file %q: %w", path, err)
		}
		baseJSON = nil
	}

	baseJSON, err = migrateLegacyOpenCodeAgentsKey(baseJSON)
	if err != nil {
		return mergeJSONResult{}, fmt.Errorf("migrate opencode agents key: %w", err)
	}
	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return mergeJSONResult{}, err
	}

	writeResult, err := filemerge.WriteFileAtomic(path, merged, 0o644)
	if err != nil {
		return mergeJSONResult{}, err
	}

	return mergeJSONResult{writeResult: writeResult, merged: merged}, nil
}

func defaultOpenCodeShareDisabled(settingsPath string, overlay []byte) ([]byte, error) {
	if openCodeSettingsHasShare(settingsPath) {
		return overlay, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(overlay, &root); err != nil {
		return nil, fmt.Errorf("unmarshal overlay json: %w", err)
	}
	if _, exists := root["share"]; !exists {
		root["share"] = "disabled"
	}

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal overlay json: %w", err)
	}
	return append(encoded, '\n'), nil
}

func openCodeSettingsHasShare(settingsPath string) bool {
	content, err := os.ReadFile(settingsPath)
	if err != nil || len(strings.TrimSpace(string(content))) == 0 {
		return false
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		return false
	}
	_, exists := root["share"]
	return exists
}


func hasOpenCodeAgentKey(settingsText, agentKey string) bool {
	root := map[string]any{}
	if err := json.Unmarshal([]byte(settingsText), &root); err != nil {
		return false
	}
	agentsRaw, ok := root["agent"]
	if !ok {
		return false
	}
	agentsMap, ok := agentsRaw.(map[string]any)
	if !ok {
		return false
	}
	_, exists := agentsMap[agentKey]
	return exists
}

func migrateLegacyOpenCodeAgentsKey(baseJSON []byte) ([]byte, error) {
	if len(strings.TrimSpace(string(baseJSON))) == 0 {
		return baseJSON, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(baseJSON, &root); err != nil {
		return baseJSON, nil
	}

	legacyRaw, hasLegacy := root["agents"]
	if !hasLegacy {
		return baseJSON, nil
	}

	legacy, ok := legacyRaw.(map[string]any)
	if !ok {
		delete(root, "agents")
		encoded, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return nil, err
		}
		return append(encoded, '\n'), nil
	}

	current := map[string]any{}
	if currentRaw, hasCurrent := root["agent"]; hasCurrent {
		if parsedCurrent, ok := currentRaw.(map[string]any); ok {
			current = parsedCurrent
		}
	}

	for key, value := range legacy {
		if _, exists := current[key]; !exists {
			current[key] = value
		}
	}

	root["agent"] = current
	delete(root, "agents")

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}

	return append(encoded, '\n'), nil
}

var sddOrchestratorMarkers = []string{
	"## Agent Teams Orchestrator",
	"## Spec-Driven Development (SDD) Orchestrator",
	"## Spec-Driven Development (SDD)",
	"# SDD Orchestrator for Cascade",
}

func hasSDDOrchestrator(content string) bool {
	for _, marker := range sddOrchestratorMarkers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
}

func sddOrchestratorAsset(agent model.AgentID) string {
	switch agent {
	case model.AgentClaudeCode:
		return "claude/sdd-orchestrator.md"
	case model.AgentGeminiCLI:
		return "gemini/sdd-orchestrator.md"
	case model.AgentCodex:
		return "codex/sdd-orchestrator.md"
	case model.AgentAntigravity:
		return "antigravity/sdd-orchestrator.md"
	case model.AgentWindsurf:
		return "windsurf/sdd-orchestrator.md"
	case model.AgentCursor:
		return "cursor/sdd-orchestrator.md"
	case model.AgentKimi:
		return "kimi/sdd-orchestrator.md"
	case model.AgentQwenCode:
		return "qwen/sdd-orchestrator.md"
	case model.AgentKiroIDE:
		return "kiro/sdd-orchestrator.md"
	case model.AgentOpenCode, model.AgentKilocode:
		return "opencode/sdd-orchestrator.md"
	default:
		return "generic/sdd-orchestrator.md"
	}
}

func injectFileAppend(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	promptPath := adapter.SystemPromptFile(homeDir)

	existing, err := readFileOrEmpty(promptPath)
	if err != nil {
		return InjectionResult{}, err
	}

	if adapter.SystemPromptStrategy() == model.StrategyInstructionsFile && strings.TrimSpace(existing) == "" {
		existing = instructionsFrontmatter
	}

	if adapter.SystemPromptStrategy() == model.StrategySteeringFile && strings.TrimSpace(existing) == "" {
		existing = steeringFrontmatter
	}

	content := assets.MustRead(sddOrchestratorAsset(adapter.Agent()))

	if hasLegacyBareOrchestrator(existing) {
		existing = stripBareOrchestratorForFilePrompt(existing)
	}

	updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content)

	writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{promptPath}}, nil
}

func hasLegacyBareOrchestrator(content string) bool {
	markedIdx := strings.Index(content, "<!-- specai:sdd-orchestrator -->")
	if markedIdx >= 0 {
		prefix := content[:markedIdx]
		if strings.Contains(prefix, "# Agent Teams Lite — Orchestrator Instructions") {
			return true
		}
	}

	firstHeading := -1
	for _, marker := range sddOrchestratorMarkers {
		idx := strings.Index(content, marker)
		if idx >= 0 && (firstHeading == -1 || idx < firstHeading) {
			firstHeading = idx
		}
	}
	if firstHeading < 0 {
		return false
	}

	if markedIdx < 0 {
		return true
	}

	return firstHeading < markedIdx
}

func stripBareOrchestratorForFilePrompt(content string) string {
	if markedIdx := strings.Index(content, "<!-- specai:sdd-orchestrator -->"); markedIdx >= 0 {
		prefix := content[:markedIdx]
		if start := strings.Index(prefix, "# Agent Teams Lite — Orchestrator Instructions"); start >= 0 {
			before := strings.TrimRight(content[:start], "\n")
			after := strings.TrimLeft(content[markedIdx:], "\n")
			if before == "" {
				if strings.HasSuffix(after, "\n") {
					return after
				}
				return after + "\n"
			}
			result := before + "\n\n" + after
			if !strings.HasSuffix(result, "\n") {
				result += "\n"
			}
			return result
		}
	}

	start := -1
	for _, marker := range sddOrchestratorMarkers {
		idx := strings.Index(content, marker)
		if idx >= 0 && (start == -1 || idx < start) {
			start = idx
		}
	}
	if start < 0 {
		return content
	}

	end := len(content)
	if rel := strings.Index(content[start:], "<!-- specai:"); rel >= 0 {
		end = start + rel
	}

	before := strings.TrimRight(content[:start], "\n")
	after := strings.TrimLeft(content[end:], "\n")

	if before == "" && after == "" {
		return ""
	}
	if before == "" {
		if strings.HasSuffix(after, "\n") {
			return after
		}
		return after + "\n"
	}
	if after == "" {
		return before + "\n"
	}

	result := before + "\n\n" + after
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

const instructionsFrontmatter = "---\n" +
	"name: SpecAI Persona\n" +
	"description: Argentina persona with SDD orchestration and sdd-memory protocol\n" +
	"applyTo: \"**\"\n" +
	"---\n"

const steeringFrontmatter = "---\n" +
	"inclusion: always\n" +
	"---\n"

func stripBareOrchestratorSection(content string) string {
	lines := strings.Split(content, "\n")

	startLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, marker := range sddOrchestratorMarkers {
			if trimmed == marker {
				startLine = i
				break
			}
		}
		if startLine >= 0 {
			break
		}
	}

	if startLine < 0 {
		return content
	}

	endLine := len(lines)
	for i := startLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "## ") {
			endLine = i
			break
		}
	}

	before := lines[:startLine]
	after := lines[endLine:]

	for len(before) > 0 && strings.TrimSpace(before[len(before)-1]) == "" {
		before = before[:len(before)-1]
	}

	var parts []string
	if len(before) > 0 {
		parts = append(parts, strings.Join(before, "\n"))
	}
	if len(after) > 0 {
		afterStr := strings.Join(after, "\n")
		afterStr = strings.TrimLeft(afterStr, "\n")
		if afterStr != "" {
			parts = append(parts, afterStr)
		}
	}

	result := strings.Join(parts, "\n\n")
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func injectMarkdownSections(homeDir string, adapter agents.Adapter, assignments map[string]model.ClaudeModelAlias) (InjectionResult, error) {
	promptPath := adapter.SystemPromptFile(homeDir)
	content := assets.MustRead(sddOrchestratorAsset(adapter.Agent()))
	if adapter.Agent() == model.AgentClaudeCode && len(assignments) > 0 {
		var err error
		content, err = injectClaudeModelAssignments(content, assignments)
		if err != nil {
			return InjectionResult{}, err
		}
	}

	existing, err := readFileOrEmpty(promptPath)
	if err != nil {
		return InjectionResult{}, err
	}

	existing = filemerge.StripLegacyATLBlock(existing)

	if hasSDDOrchestrator(existing) && !strings.Contains(existing, "<!-- specai:sdd-orchestrator -->") {
		existing = stripBareOrchestratorSection(existing)
	}

	updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content)

	writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{promptPath}}, nil
}

var claudeModelAssignmentRowOrder = []string{
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"sdd-onboard",
	"jd-judge-a",
	"jd-judge-b",
	"jd-fix-agent",
	"default",
}

var claudeModelAssignmentReasons = map[string]string{
	"orchestrator": "Coordinates, makes decisions",
	"sdd-explore":  "Reads code, structural - not architectural",
	"sdd-propose":  "Architectural decisions",
	"sdd-spec":     "Structured writing",
	"sdd-design":   "Architecture decisions",
	"sdd-tasks":    "Mechanical breakdown",
	"sdd-apply":    "Implementation",
	"sdd-verify":   "Validation against spec",
	"sdd-archive":  "Copy and close",
	"sdd-onboard":  "Guided walkthrough, pedagogical",
	"jd-judge-a":   "Adversarial review — blind judge A",
	"jd-judge-b":   "Adversarial review — blind judge B",
	"jd-fix-agent": "Surgical fixes from confirmed issues",
	"default":      "Non-SDD general delegation",
}

func injectClaudeModelAssignments(content string, assignments map[string]model.ClaudeModelAlias) (string, error) {
	const openMarker = "<!-- specai:sdd-model-assignments -->"
	const closeMarker = "<!-- /specai:sdd-model-assignments -->"

	start := strings.Index(content, openMarker)
	end := strings.Index(content, closeMarker)
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("sdd orchestrator asset missing model assignment markers")
	}

	merged := model.ClaudeModelPresetBalanced()
	for key, alias := range assignments {
		if alias.Valid() {
			merged[key] = alias
		}
	}

	replacement := renderClaudeModelAssignmentsSection(merged)
	start += len(openMarker)
	return content[:start] + "\n" + replacement + content[end:], nil
}

func resolveClaudeModelAlias(assignments map[string]model.ClaudeModelAlias, phase string) model.ClaudeModelAlias {
	merged := model.ClaudeModelPresetBalanced()
	for key, alias := range assignments {
		if alias.Valid() {
			merged[key] = alias
		}
	}

	if alias, ok := merged[phase]; ok && alias.Valid() {
		return alias
	}
	if alias, ok := merged["default"]; ok && alias.Valid() {
		return alias
	}
	return model.ClaudeModelSonnet
}

func renderClaudeModelAssignmentsSection(assignments map[string]model.ClaudeModelAlias) string {
	var b strings.Builder
	b.WriteString("## Model Assignments\n\n")
	b.WriteString("Read this table at session start (or before first delegation), cache it for the session, and pass the mapped alias in every Agent tool call via the `model` parameter. If a phase is missing, use the `default` row. If you do not have access to the assigned model (for example, no Opus access), substitute `sonnet` and continue.\n\n")
	b.WriteString("The Claude Code session model is controlled by Claude Code itself; SpecAI does not configure the main orchestrator model. This table applies only to Agent tool calls for SDD phase sub-agents and general delegation.\n\n")
	b.WriteString("**Mandatory model gate:** Every Agent tool call MUST include `model`. Calling Agent without `model` is invalid. Before each Agent call, resolve the target phase to an alias from this table; for general/non-SDD delegation use `default`. If you are about to call Agent and have not chosen a `model`, STOP and choose the mapped alias first.\n\n")
	b.WriteString("| Phase | Default Model | Reason |\n")
	b.WriteString("|-------|---------------|--------|\n")
	for _, key := range claudeModelAssignmentRowOrder {
		alias := assignments[key]
		if !alias.Valid() {
			alias = model.ClaudeModelSonnet
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", key, alias, claudeModelAssignmentReasons[key]))
	}
	b.WriteString("\n")
	return b.String()
}

var jdAgentSet = buildJDAgentSet()

func buildJDAgentSet() map[string]bool {
	phases := opencode.JDPhases()
	set := make(map[string]bool, len(phases))
	for _, p := range phases {
		set[p] = true
	}
	return set
}

func isJDAgent(name string) bool {
	return jdAgentSet[name]
}

func injectModelAssignments(overlayBytes []byte, assignments map[string]model.ModelAssignment, rootModelID string, existingAgentKeys map[string]bool) ([]byte, error) {
	var overlay map[string]any
	if err := json.Unmarshal(overlayBytes, &overlay); err != nil {
		return nil, fmt.Errorf("unmarshal overlay for model injection: %w", err)
	}

	agentsRaw, ok := overlay["agent"]
	if !ok {
		return overlayBytes, nil
	}
	agents, ok := agentsRaw.(map[string]any)
	if !ok {
		return overlayBytes, nil
	}

	for phase, agentDef := range agents {
		agentMap, ok := agentDef.(map[string]any)
		if !ok {
			continue
		}

		assignment, hasExplicitAssignment := assignments[phase]

		switch {
		case hasExplicitAssignment && assignment.ProviderID != "" && assignment.ModelID != "":
			agentMap["model"] = assignment.FullID()
			if assignment.Effort != "" {
				agentMap["variant"] = assignment.Effort
			} else {
				agentMap["variant"] = ""
			}
		case existingAgentKeys[phase]:
			// Agent already exists in user's config — preserve whatever they have.
		case rootModelID != "":
			if !isJDAgent(phase) {
				agentMap["model"] = rootModelID
				agentMap["variant"] = ""
			}
		}
	}

	result, err := json.MarshalIndent(overlay, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal overlay after model injection: %w", err)
	}
	return append(result, '\n'), nil
}


func readOpenCodeRootModel(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read opencode root model from %q: %w", path, err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(data, &root); err != nil {
		return "", nil
	}

	rootModelID, _ := root["model"].(string)
	return rootModelID, nil
}

func readExistingAgentModels(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("read existing agent keys from %q: %w", path, err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(data, &root); err != nil {
		return map[string]bool{}, nil
	}

	agentRaw, ok := root["agent"]
	if !ok {
		return map[string]bool{}, nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return map[string]bool{}, nil
	}

	result := make(map[string]bool, len(agentMap))
	for name := range agentMap {
		result[name] = true
	}
	return result, nil
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
