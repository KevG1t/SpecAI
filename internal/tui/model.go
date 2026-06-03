package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/KevG1t/SpecAI/internal/agents"
	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/catalog"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/pipeline"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/steps"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui/screens"
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	currentScreen  Screen
	previousScreen Screen
	welcome        screens.WelcomeModel
	agentSelect   screens.AgentSelectModel

	persona           screens.PersonaModel
	preset            screens.PresetModel
	claudeModelPicker screens.ClaudeModelPickerModel
	kiroModelPicker   screens.KiroModelPickerModel
	sddMode           screens.SDDModeModel
	strictTDD         screens.StrictTDDModel

	// New wizard screens (Phase 4)
	detection           screens.DetectionModel
	review              screens.ReviewModel
	dependencyTree      screens.DependencyTreeModel
	openCodeModelPicker screens.OpenCodeModelPickerModel
	skillPicker         screens.SkillPickerModel
	mcpPicker           screens.MCPPickerModel
	configPicker        screens.ConfigPickerModel

	install    screens.InstallModel
	setupLocal screens.SetupLocalModel
	upgrade    screens.UpgradeModel
	sync       screens.SyncModel
	upgSync    screens.UpgradeSyncModel
	uninstall  screens.UninstallModel

	selectedOpt     string
	installCtx      *steps.InstallContext
	selectedAgents  []model.AgentID
	progressCh      chan pipeline.ProgressEvent
	latestProg      screens.ProgressMsg
	completePayload *screens.CompletePayload

	// Accumulated wizard state (Phase 4)
	detectionResult *system.DetectionResult
	selectedPreset  model.PresetID
	resolvedPlan    planner.ResolvedPlan

	Backups        []backup.Manifest
	SelectedBackup int
	BackupScroll   int
	BackupErr      error
	RestoreMsg     string
}

func NewMainModel() MainModel {
	return MainModel{
		currentScreen: ScreenWelcome,
		welcome:       screens.NewWelcomeModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.welcome.Init()
}

// setCurrentScreen stores m.currentScreen as previousScreen before transitioning to next.
func (m *MainModel) setCurrentScreen(next Screen) {
	m.previousScreen = m.currentScreen
	m.currentScreen = next
}

type dummyStep struct {
	id string
}

func (s dummyStep) ID() string { return s.id }
func (s dummyStep) Run() error {
	time.Sleep(1 * time.Second)
	return nil
}

func waitForProgress(ch <-chan pipeline.ProgressEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return screens.ProgressMsg{
			TaskName: string(ev.Stage) + " / " + ev.StepID,
			Status:   string(ev.Status),
			Progress: 0,
		}
	}
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case screens.BackMsg:
		// No-op if previousScreen is zero value (user is already at the first screen).
		if m.previousScreen != "" {
			m.setCurrentScreen(m.previousScreen)
		}
		return m, nil

	case screens.AgentsSelectedMsg:
		// Build install context with the selected IDEs right away.
		ctx, errCtx := steps.NewInstallContext()
		if errCtx != nil {
			m.setCurrentScreen(ScreenInstall)
			m.install = screens.NewInstallModel()
			return m, m.install.Init()
		}
		var adapters []system.IDEAdapter
		for _, id := range msg.Agents {
			if a := system.GetAdapterByAgentID(id); a != nil {
				adapters = append(adapters, a)
			}
		}
		ctx.IDEs = adapters
		m.installCtx = ctx
		m.selectedAgents = msg.Agents
		m.persona = screens.NewPersonaModel()
		m.setCurrentScreen(ScreenPersona)
		return m, nil

	case screens.PersonaSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.Persona = msg.Persona
		}
		m.preset = screens.NewPresetModel()
		m.setCurrentScreen(ScreenPreset)
		return m, nil

	case screens.PresetSelectedMsg:
		// Store preset and build the initial resolved plan (Approach B).
		m.selectedPreset = msg.Preset
		if m.installCtx != nil {
			m.installCtx.Preset = msg.Preset
		}
		sel := m.buildSelection()
		resolved, err := planner.NewResolver(planner.MVPGraph()).Resolve(sel)
		if err == nil {
			m.resolvedPlan = resolved
		}

		// Route to agent-specific screens.
		next := m.nextScreenAfterConfig()
		switch next {
		case ScreenClaudeModelPicker:
			m.claudeModelPicker = screens.NewClaudeModelPickerModel(nil)
			m.setCurrentScreen(ScreenClaudeModelPicker)
		case ScreenKiroModelPicker:
			m.kiroModelPicker = screens.NewKiroModelPickerModel(nil)
			m.setCurrentScreen(ScreenKiroModelPicker)
		case ScreenSDDMode:
			m.sddMode = screens.NewSDDModeModel()
			m.setCurrentScreen(ScreenSDDMode)
		default:
			// No agent-specific screen — go directly to DependencyTree.
			m.dependencyTree = screens.NewDependencyTreeModel(
				m.selectedPreset, m.resolvedPlan, m.selectedAgents,
				planner.NewResolver(planner.MVPGraph()).Resolve,
			)
			m.setCurrentScreen(ScreenDependencyTree)
		}
		return m, nil

	case screens.ClaudeModelsSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.ClaudeModelAssignments = stringifyAliasMap(msg.Assignments)
		}
		hasKiro := agentInList(m.selectedAgents, model.AgentKiroIDE)
		hasOpenCode := agentInList(m.selectedAgents, model.AgentOpenCode)
		if hasKiro {
			m.kiroModelPicker = screens.NewKiroModelPickerModel(nil)
			m.setCurrentScreen(ScreenKiroModelPicker)
			return m, nil
		} else if hasOpenCode {
			m.sddMode = screens.NewSDDModeModel()
			m.setCurrentScreen(ScreenSDDMode)
			return m, nil
		}
		// No further agent-specific screens — go to DependencyTree.
		m.dependencyTree = screens.NewDependencyTreeModel(
			m.selectedPreset, m.resolvedPlan, m.selectedAgents,
			planner.NewResolver(planner.MVPGraph()).Resolve,
		)
		m.setCurrentScreen(ScreenDependencyTree)
		return m, nil

	case screens.KiroModelsSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.KiroModelAssignments = stringifyAliasMap(msg.Assignments)
		}
		hasOpenCode := agentInList(m.selectedAgents, model.AgentOpenCode)
		if hasOpenCode {
			m.sddMode = screens.NewSDDModeModel()
			m.setCurrentScreen(ScreenSDDMode)
			return m, nil
		}
		// No further agent-specific screens — go to DependencyTree.
		m.dependencyTree = screens.NewDependencyTreeModel(
			m.selectedPreset, m.resolvedPlan, m.selectedAgents,
			planner.NewResolver(planner.MVPGraph()).Resolve,
		)
		m.setCurrentScreen(ScreenDependencyTree)
		return m, nil

	case screens.SDDModeSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.SDDMode = string(msg.Mode)
		}
		// If OpenCode is selected, route to OpenCode model picker before StrictTDD.
		if agentInList(m.selectedAgents, model.AgentOpenCode) {
			m.openCodeModelPicker = screens.NewOpenCodeModelPickerModel()
			m.setCurrentScreen(ScreenOpenCodeModelPicker)
			return m, nil
		}
		m.strictTDD = screens.NewStrictTDDModel()
		m.setCurrentScreen(ScreenStrictTDD)
		return m, nil

	case screens.DetectionConfirmedMsg:
		m.detectionResult = msg.Result
		detected := system.DetectedAgentIDs()
		m.agentSelect = screens.NewAgentSelectModel(detected)
		m.setCurrentScreen(ScreenAgentSelect)
		return m, nil

	case screens.StrictTDDSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.StrictTDD = msg.Enabled
		}
		next := m.nextScreenAfterStrictTDD()
		m.setCurrentScreen(next)
		m.dependencyTree = screens.NewDependencyTreeModel(
			m.selectedPreset, m.resolvedPlan, m.selectedAgents,
			planner.NewResolver(planner.MVPGraph()).Resolve,
		)
		return m, nil

	case screens.OpenCodeModelsSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.ModelAssignments = msg.Assignments
		}
		m.strictTDD = screens.NewStrictTDDModel()
		m.setCurrentScreen(ScreenStrictTDD)
		return m, nil

	case screens.DependencyTreeConfirmedMsg:
		// Re-resolve with the confirmed component list.
		sel := m.buildSelection()
		sel.Components = msg.Components
		resolved, err := planner.NewResolver(planner.MVPGraph()).Resolve(sel)
		if err == nil {
			m.resolvedPlan = resolved
		}
		if m.selectedPreset == model.PresetCustom {
			m.skillPicker = screens.NewSkillPickerModel(nil)
			m.setCurrentScreen(ScreenSkillPicker)
			return m, nil
		}
		payload := planner.BuildReviewPayload(sel, m.resolvedPlan)
		m.review = screens.NewReviewModel(payload)
		m.setCurrentScreen(ScreenReview)
		return m, nil

	case screens.SkillsSelectedMsg:
		if m.installCtx != nil {
			m.installCtx.SelectedComponents = nil // reset; skills are kept in sel
		}
		// In the Custom preset flow, route through MCP picker before review.
		if m.selectedPreset == model.PresetCustom {
			m.mcpPicker = screens.NewMCPPickerModel()
			m.setCurrentScreen(ScreenMCPPicker)
			return m, nil
		}
		// Non-custom presets: build review payload and go directly to review.
		sel := m.buildSelection()
		sel.Components = m.resolvedPlan.OrderedComponents
		sel.Skills = msg.Skills
		payload := planner.BuildReviewPayload(sel, m.resolvedPlan)
		m.review = screens.NewReviewModel(payload)
		m.setCurrentScreen(ScreenReview)
		return m, nil

	case screens.MCPServersSelectedMsg:
		// Store selected MCP component IDs (Custom flow).
		if m.installCtx != nil {
			m.installCtx.SelectedMCPServers = msg.ComponentIDs
		}
		m.configPicker = screens.NewConfigPickerModel()
		m.setCurrentScreen(ScreenConfigPicker)
		return m, nil

	case screens.ConfigSelectedMsg:
		// Store config customization (Custom flow).
		if m.installCtx != nil {
			m.installCtx.Theme = msg.Theme
			m.installCtx.PermissionsLevel = msg.PermissionsLevel
			m.installCtx.EditorMode = msg.EditorMode
		}
		// Proceed to dependency tree / review.
		m.dependencyTree = screens.NewDependencyTreeModel(
			m.selectedPreset, m.resolvedPlan, m.selectedAgents,
			planner.NewResolver(planner.MVPGraph()).Resolve,
		)
		m.setCurrentScreen(ScreenDependencyTree)
		return m, nil

	case screens.ReviewConfirmedMsg:
		m.install = screens.NewInstallModel()
		m.install.Started = true
		m.setCurrentScreen(ScreenInstall)
		return m, tea.Batch(m.install.Init(), func() tea.Msg { return screens.StartPipelineMsg{Action: "Install"} })

	case screens.ReviewBackMsg:
		m.setCurrentScreen(ScreenDependencyTree)
		return m, nil

	case screens.OptionSelectedMsg:
		m.selectedOpt = msg.Option
		if msg.Option == "Backup" {
			m.setCurrentScreen(ScreenBackups)
			m.Backups = ListBackups()
			m.SelectedBackup = 0
			m.BackupScroll = 0
			return m, nil
		} else if msg.Option == "Install" {
			// Run system detection before agent selection.
			m.detection = screens.NewDetectionModel(func(ctx context.Context) (*system.DetectionResult, error) {
				r, err := system.Detect(ctx)
				return &r, err
			})
			m.setCurrentScreen(ScreenDetection)
			return m, m.detection.Init()
		} else if msg.Option == "Setup Local (Inyectar en este Repo)" {
			m.setupLocal = screens.NewSetupLocalModel()
			m.setCurrentScreen(ScreenSetupLocal)
			return m, m.setupLocal.Init()
		} else if msg.Option == "Sync" {
			m.sync = screens.NewSyncModel()
			m.setCurrentScreen(ScreenSync)
			return m, m.sync.Init()
		} else if msg.Option == "Upgrade" {
			m.upgrade = screens.NewUpgradeModel()
			m.setCurrentScreen(ScreenUpgrade)
			return m, m.upgrade.Init()
		} else if msg.Option == "Upgrade + Sync" {
			m.upgSync = screens.NewUpgradeSyncModel()
			m.setCurrentScreen(ScreenUpgradeSync)
			return m, m.upgSync.Init()
		} else if msg.Option == "Uninstall" {
			m.uninstall = screens.NewUninstallModel()
			m.setCurrentScreen(ScreenUninstall)
			return m, m.uninstall.Init()
		}
		return m, nil

	case screens.StartPipelineMsg:
		m.progressCh = make(chan pipeline.ProgressEvent, 100)
		m.latestProg = screens.ProgressMsg{TaskName: "Preparando pipeline...", Status: "Iniciando"}

		// Capture at closure creation time so it's safe to use inside goroutine.
		installCtx := m.installCtx
		capturedPlan := m.resolvedPlan

		startCmd := func() tea.Msg {
			var plan pipeline.StagePlan

			if msg.Action == "Install" {
				var ctx *steps.InstallContext
				if installCtx != nil {
					ctx = installCtx
				} else {
					var errCtx error
					ctx, errCtx = steps.NewInstallContext()
					if errCtx != nil {
						return screens.PipelineFinishedMsg{Err: errCtx}
					}
				}

				if len(ctx.IDEs) == 0 {
					return screens.PipelineFinishedMsg{Err: fmt.Errorf("no agents selected for installation")}
				}

				// Use the plan built through the wizard (DependencyTree → Review).
				// Fall back to resolving from context if the wizard plan is empty.
				resolvedPlan := capturedPlan
				if len(resolvedPlan.OrderedComponents) == 0 {
					agentIDs := make([]model.AgentID, len(ctx.IDEs))
					for i, ide := range ctx.IDEs {
						agentIDs[i] = ide.AgentID()
					}
					var err error
					resolvedPlan, err = planner.NewResolver(planner.MVPGraph()).Resolve(model.Selection{
						Agents:     agentIDs,
						Components: catalog.ComponentsForPreset(ctx.Preset),
						Persona:    ctx.Persona,
						Preset:     ctx.Preset,
					})
					if err != nil {
						return screens.PipelineFinishedMsg{Err: err}
					}
				}

				// Pre-flight: validate agent selection for known conflicts.
				agentIDs := make([]model.AgentID, len(ctx.IDEs))
				for i, ide := range ctx.IDEs {
					agentIDs[i] = ide.AgentID()
				}
				agentWarnings := agents.ValidateAgentSelection(agentIDs)
				for _, w := range agentWarnings {
					ctx.Warnings = append(ctx.Warnings, w.Message)
				}

				// Map ResolvedPlan to StagePlan manually
				var applySteps []pipeline.Step
				applySteps = append(applySteps, steps.NewStepInstallGlobalRules(ctx)) // Persona

				// Inject assets (Skills, SDD, sdd-memory config) — idempotent via ForceAssets flag.
				injector := steps.NewAssetInjectorWithOpts(nil, ctx.ForceAssets)
				applySteps = append(applySteps, steps.NewStepInjectAssets(ctx, injector, resolvedPlan))
				applySteps = append(applySteps, steps.NewStepInstallAgents(ctx))
				applySteps = append(applySteps, steps.NewStepInjectSubAgents(ctx))
				applySteps = append(applySteps, steps.NewStepInjectOpenCodePlugins(ctx))
				applySteps = append(applySteps, steps.NewStepInjectSDDMemory(ctx))
				applySteps = append(applySteps, steps.NewStepInjectSDDMemoryService(ctx))
				applySteps = append(applySteps, steps.NewStepInjectOpenCodeOverlay(ctx))
				applySteps = append(applySteps, steps.NewStepInjectSlashCommands(ctx))
				applySteps = append(applySteps, steps.NewStepInjectMCP(ctx))
				applySteps = append(applySteps, steps.NewStepInjectMCPComponents(ctx, resolvedPlan))

				prepareSteps := []pipeline.Step{}
				// Scan first if IDEs weren't pre-populated (fallback path).
				if len(ctx.IDEs) == 0 {
					prepareSteps = append(prepareSteps, steps.NewStepScanGlobalIDEs(ctx))
				}
				prepareSteps = append(prepareSteps, steps.NewStepSnapshotBeforeInstall(ctx))
				plan = pipeline.StagePlan{
					Prepare: prepareSteps,
					Apply:   applySteps,
				}
			} else if msg.Action == "SetupLocal" {
				ctx, errCtx := steps.NewInstallContext()
				if errCtx != nil {
					return screens.PipelineFinishedMsg{Err: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanGlobalIDEs(ctx)},
					Apply:   []pipeline.Step{steps.NewStepSetupLocalRules(ctx)},
				}
			} else if msg.Action == "Sync" {
				syncCtx, errCtx := steps.NewSyncContext()
				if errCtx != nil {
					return screens.PipelineFinishedMsg{Err: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanIDEsForSync(syncCtx), steps.NewStepSnapshotBeforeSync(syncCtx)},
					Apply:   []pipeline.Step{steps.NewStepSyncGlobalRules(syncCtx), steps.NewStepSyncGlobalSkills(syncCtx)},
				}
			} else if msg.Action == "Upgrade" {
				upgradeCtx, errCtx := steps.NewUpgradeContext(GetVersion())
				if errCtx != nil {
					return screens.PipelineFinishedMsg{Err: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepCheckForUpdates(upgradeCtx)},
					Apply:   []pipeline.Step{steps.NewStepInstallUpdates(upgradeCtx)},
				}
			} else if msg.Action == "UpgradeSync" {
				upgradeCtx, errCtx := steps.NewUpgradeContext(GetVersion())
				if errCtx != nil {
					return screens.PipelineFinishedMsg{Err: errCtx}
				}
				syncCtx, errCtx2 := steps.NewSyncContext()
				if errCtx2 != nil {
					return screens.PipelineFinishedMsg{Err: errCtx2}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepCheckForUpdates(upgradeCtx), steps.NewStepScanIDEsForSync(syncCtx), steps.NewStepSnapshotBeforeSync(syncCtx)},
					Apply:   []pipeline.Step{steps.NewStepInstallUpdates(upgradeCtx), steps.NewStepSyncGlobalRules(syncCtx), steps.NewStepSyncGlobalSkills(syncCtx)},
				}
			} else if msg.Action == "Uninstall" {
				uninstallCtx, errCtx := steps.NewUninstallContext()
				if errCtx != nil {
					return screens.PipelineFinishedMsg{Err: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanIDEsForUninstall(uninstallCtx), steps.NewStepSnapshotBeforeUninstall(uninstallCtx)},
					Apply:   []pipeline.Step{steps.NewStepRemoveGlobalRules(uninstallCtx), steps.NewStepRemoveGlobalSkills(uninstallCtx)},
				}
			}

			orch := pipeline.NewOrchestrator(
				pipeline.DefaultRollbackPolicy(),
				pipeline.WithFailurePolicy(pipeline.StopOnError),
				pipeline.WithProgressFunc(func(ev pipeline.ProgressEvent) {
					m.progressCh <- ev
				}),
			)
			res := orch.Execute(plan)
			close(m.progressCh)

			var payload *screens.CompletePayload
			if msg.Action == "Install" && installCtx != nil {
				rollbackPerformed := res.Rollback.Stage == pipeline.StageRollback

				// Convert agents.ValidationWarning → screens.ValidationWarning.
				var screenWarnings []screens.ValidationWarning
				for _, w := range installCtx.Warnings {
					screenWarnings = append(screenWarnings, screens.ValidationWarning{Message: w})
				}

				payload = &screens.CompletePayload{
					ConfiguredAgents:    len(installCtx.IDEs),
					InstalledComponents: len(plan.Apply),
					RollbackPerformed:   rollbackPerformed,
					AuthGuidance:        installCtx.AuthGuidance,
					ValidationWarnings:  screenWarnings,
				}
				if res.Err != nil {
					payload.FailedSteps = append(payload.FailedSteps, screens.FailedStep{
						StepName: "pipeline",
						Err:      res.Err.Error(),
					})
				}
			}
			return screens.PipelineFinishedMsg{Err: res.Err, Payload: payload}
		}
		return m, tea.Batch(startCmd, waitForProgress(m.progressCh))

	case screens.ProgressMsg:
		m.latestProg = msg
		// Dispatch to the active module screen
		return m, tea.Batch(
			func() tea.Msg { return msg },
			waitForProgress(m.progressCh),
		)

	case screens.PipelineFinishedMsg:
		// If the message carries a CompletePayload, transition to the completion screen.
		if msg.Payload != nil {
			m.completePayload = msg.Payload
			m.setCurrentScreen(ScreenComplete)
			return m, nil
		}
		// Legacy: dispatch to the active module screen.
		return m, func() tea.Msg {
			return msg
		}
	}

	// Route message to active screen if applicable
	switch m.currentScreen {
	case ScreenWelcome:
		var wModel tea.Model
		wModel, cmd = m.welcome.Update(msg)
		m.welcome = wModel.(screens.WelcomeModel)
		cmds = append(cmds, cmd)
	case ScreenAgentSelect:
		var aModel tea.Model
		aModel, cmd = m.agentSelect.Update(msg)
		m.agentSelect = aModel.(screens.AgentSelectModel)
		cmds = append(cmds, cmd)
	case ScreenPersona:
		var pModel tea.Model
		pModel, cmd = m.persona.Update(msg)
		m.persona = pModel.(screens.PersonaModel)
		cmds = append(cmds, cmd)
	case ScreenPreset:
		var prModel tea.Model
		prModel, cmd = m.preset.Update(msg)
		m.preset = prModel.(screens.PresetModel)
		cmds = append(cmds, cmd)
	case ScreenClaudeModelPicker:
		var cModel tea.Model
		cModel, cmd = m.claudeModelPicker.Update(msg)
		m.claudeModelPicker = cModel.(screens.ClaudeModelPickerModel)
		cmds = append(cmds, cmd)
	case ScreenKiroModelPicker:
		var kModel tea.Model
		kModel, cmd = m.kiroModelPicker.Update(msg)
		m.kiroModelPicker = kModel.(screens.KiroModelPickerModel)
		cmds = append(cmds, cmd)
	case ScreenSDDMode:
		var sModel tea.Model
		sModel, cmd = m.sddMode.Update(msg)
		m.sddMode = sModel.(screens.SDDModeModel)
		cmds = append(cmds, cmd)
	case ScreenStrictTDD:
		var tModel tea.Model
		tModel, cmd = m.strictTDD.Update(msg)
		m.strictTDD = tModel.(screens.StrictTDDModel)
		cmds = append(cmds, cmd)
	case ScreenInstall:
		var mModel tea.Model
		mModel, cmd = m.install.Update(msg)
		m.install = mModel.(screens.InstallModel)
		cmds = append(cmds, cmd)
	case ScreenSetupLocal:
		var mModel tea.Model
		mModel, cmd = m.setupLocal.Update(msg)
		m.setupLocal = mModel.(screens.SetupLocalModel)
		cmds = append(cmds, cmd)
	case ScreenUpgrade:
		var mModel tea.Model
		mModel, cmd = m.upgrade.Update(msg)
		m.upgrade = mModel.(screens.UpgradeModel)
		cmds = append(cmds, cmd)
	case ScreenSync:
		var mModel tea.Model
		mModel, cmd = m.sync.Update(msg)
		m.sync = mModel.(screens.SyncModel)
		cmds = append(cmds, cmd)
	case ScreenUpgradeSync:
		var mModel tea.Model
		mModel, cmd = m.upgSync.Update(msg)
		m.upgSync = mModel.(screens.UpgradeSyncModel)
		cmds = append(cmds, cmd)
	case ScreenUninstall:
		var mModel tea.Model
		mModel, cmd = m.uninstall.Update(msg)
		m.uninstall = mModel.(screens.UninstallModel)
		cmds = append(cmds, cmd)
	case ScreenDetection:
		var dModel tea.Model
		dModel, cmd = m.detection.Update(msg)
		m.detection = dModel.(screens.DetectionModel)
		cmds = append(cmds, cmd)
	case ScreenReview:
		var rModel tea.Model
		rModel, cmd = m.review.Update(msg)
		m.review = rModel.(screens.ReviewModel)
		cmds = append(cmds, cmd)
	case ScreenDependencyTree:
		var dtModel tea.Model
		dtModel, cmd = m.dependencyTree.Update(msg)
		m.dependencyTree = dtModel.(screens.DependencyTreeModel)
		cmds = append(cmds, cmd)
	case ScreenOpenCodeModelPicker:
		var ocModel tea.Model
		ocModel, cmd = m.openCodeModelPicker.Update(msg)
		m.openCodeModelPicker = ocModel.(screens.OpenCodeModelPickerModel)
		cmds = append(cmds, cmd)
	case ScreenSkillPicker:
		var spModel tea.Model
		spModel, cmd = m.skillPicker.Update(msg)
		m.skillPicker = spModel.(screens.SkillPickerModel)
		cmds = append(cmds, cmd)
	case ScreenMCPPicker:
		var mpModel tea.Model
		mpModel, cmd = m.mcpPicker.Update(msg)
		m.mcpPicker = mpModel.(screens.MCPPickerModel)
		cmds = append(cmds, cmd)
	case ScreenConfigPicker:
		var cpModel tea.Model
		cpModel, cmd = m.configPicker.Update(msg)
		m.configPicker = cpModel.(screens.ConfigPickerModel)
		cmds = append(cmds, cmd)
	case ScreenBackups, ScreenRestoreConfirm, ScreenDeleteConfirm, ScreenBackupResult:
		m, cmd = m.UpdateBackups(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.currentScreen {
	case ScreenWelcome:
		return m.welcome.View()
	case ScreenAgentSelect:
		return m.agentSelect.View()
	case ScreenPersona:
		return m.persona.View()
	case ScreenPreset:
		return m.preset.View()
	case ScreenClaudeModelPicker:
		return m.claudeModelPicker.View()
	case ScreenKiroModelPicker:
		return m.kiroModelPicker.View()
	case ScreenSDDMode:
		return m.sddMode.View()
	case ScreenStrictTDD:
		return m.strictTDD.View()
	case ScreenInstall:
		return m.install.View()
	case ScreenSetupLocal:
		return m.setupLocal.View()
	case ScreenUpgrade:
		return m.upgrade.View()
	case ScreenSync:
		return m.sync.View()
	case ScreenUpgradeSync:
		return m.upgSync.View()
	case ScreenUninstall:
		return m.uninstall.View()
	case ScreenDetection:
		return m.detection.View()
	case ScreenReview:
		return m.review.View()
	case ScreenDependencyTree:
		return m.dependencyTree.View()
	case ScreenOpenCodeModelPicker:
		return m.openCodeModelPicker.View()
	case ScreenSkillPicker:
		return m.skillPicker.View()
	case ScreenMCPPicker:
		return m.mcpPicker.View()
	case ScreenConfigPicker:
		return m.configPicker.View()
	case ScreenLoading:
		// Not used anymore as each screen handles its own "running" state
		return ""
	case ScreenComplete:
		if m.completePayload != nil {
			return screens.RenderComplete(*m.completePayload)
		}
		return ""
	case ScreenBackups, ScreenRestoreConfirm, ScreenDeleteConfirm, ScreenBackupResult:
		return m.ViewBackups()
	default:
		return "Pantalla desconocida"
	}
}

// agentInList reports whether the given AgentID is present in the slice.
func agentInList(agents []model.AgentID, target model.AgentID) bool {
	for _, a := range agents {
		if a == target {
			return true
		}
	}
	return false
}

// stringifyAliasMap converts a map[string]model.ClaudeModelAlias to map[string]string
// so it can be stored in InstallContext without importing screens from steps.
func stringifyAliasMap(m map[string]model.ClaudeModelAlias) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = string(v)
	}
	return out
}

// buildSelection constructs a model.Selection from the accumulated MainModel state.
// It is called at PresetSelectedMsg time (Approach B) and before entering ScreenReview.
func (m MainModel) buildSelection() model.Selection {
	components := catalog.ComponentsForPreset(m.selectedPreset)

	var persona model.PersonaID
	var sddMode model.SDDModeID
	var strictTDD bool
	var modelAssignments map[string]model.ModelAssignment
	var claudeAssignments map[string]model.ClaudeModelAlias

	if m.installCtx != nil {
		persona = m.installCtx.Persona
		sddMode = model.SDDModeID(m.installCtx.SDDMode)
		strictTDD = m.installCtx.StrictTDD
		modelAssignments = m.installCtx.ModelAssignments

		// Convert ClaudeModelAssignments map[string]string → map[string]model.ClaudeModelAlias
		if len(m.installCtx.ClaudeModelAssignments) > 0 {
			claudeAssignments = make(map[string]model.ClaudeModelAlias, len(m.installCtx.ClaudeModelAssignments))
			for k, v := range m.installCtx.ClaudeModelAssignments {
				if k == "orchestrator" {
					continue // Claude Code manages its own orchestrator model
				}
				claudeAssignments[k] = model.ClaudeModelAlias(v)
			}
		}
	}

	return model.Selection{
		Agents:                 m.selectedAgents,
		Components:             components,
		Persona:                persona,
		Preset:                 m.selectedPreset,
		SDDMode:                sddMode,
		StrictTDD:              strictTDD,
		ModelAssignments:       modelAssignments,
		ClaudeModelAssignments: claudeAssignments,
	}
}

// nextScreenAfterConfig returns the screen that should follow the preset selection,
// based on which agents the user selected.
func (m MainModel) nextScreenAfterConfig() Screen {
	if agentInList(m.selectedAgents, model.AgentClaudeCode) {
		return ScreenClaudeModelPicker
	}
	if agentInList(m.selectedAgents, model.AgentKiroIDE) {
		return ScreenKiroModelPicker
	}
	if agentInList(m.selectedAgents, model.AgentOpenCode) {
		return ScreenSDDMode
	}
	return ScreenDependencyTree
}

// nextScreenAfterStrictTDD returns the screen that should follow ScreenStrictTDD.
// ScreenOpenCodePlugins is excluded from this change per scope constraint.
func (m MainModel) nextScreenAfterStrictTDD() Screen {
	return ScreenDependencyTree
}
