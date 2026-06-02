package tui

import (
	"fmt"
	"time"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/pipeline"
	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/steps"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui/screens"
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	currentScreen Screen
	welcome       screens.WelcomeModel
	agentSelect   screens.AgentSelectModel

	install    screens.InstallModel
	setupLocal screens.SetupLocalModel
	upgrade    screens.UpgradeModel
	sync       screens.SyncModel
	upgSync    screens.UpgradeSyncModel
	uninstall  screens.UninstallModel

	selectedOpt     string
	installCtx      *steps.InstallContext
	progressCh      chan pipeline.ProgressEvent
	latestProg      screens.ProgressMsg
	completePayload *screens.CompletePayload

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
		m.currentScreen = ScreenWelcome
		m.selectedOpt = ""
		return m, nil

	case screens.AgentsSelectedMsg:
		// Build IDEAdapter slice from the selected AgentIDs.
		ctx, errCtx := steps.NewInstallContext()
		if errCtx != nil {
			// Fallback: transition to install screen which will surface the error.
			m.currentScreen = ScreenInstall
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
		m.currentScreen = ScreenInstall
		m.install = screens.NewInstallModel()
		return m, m.install.Init()

	case screens.OptionSelectedMsg:
		m.selectedOpt = msg.Option
		if msg.Option == "Backup" {
			m.currentScreen = ScreenBackups
			m.Backups = ListBackups()
			m.SelectedBackup = 0
			m.BackupScroll = 0
			return m, nil
		} else if msg.Option == "Install" {
			// Show agent selection before starting the install pipeline.
			detected := system.DetectedAgentIDs()
			m.agentSelect = screens.NewAgentSelectModel(detected)
			m.currentScreen = ScreenAgentSelect
			return m, nil
		} else if msg.Option == "Setup Local (Inyectar en este Repo)" {
			m.currentScreen = ScreenSetupLocal
			m.setupLocal = screens.NewSetupLocalModel()
			return m, m.setupLocal.Init()
		} else if msg.Option == "Sync" {
			m.currentScreen = ScreenSync
			m.sync = screens.NewSyncModel()
			return m, m.sync.Init()
		} else if msg.Option == "Upgrade" {
			m.currentScreen = ScreenUpgrade
			m.upgrade = screens.NewUpgradeModel()
			return m, m.upgrade.Init()
		} else if msg.Option == "Upgrade + Sync" {
			m.currentScreen = ScreenUpgradeSync
			m.upgSync = screens.NewUpgradeSyncModel()
			return m, m.upgSync.Init()
		} else if msg.Option == "Uninstall" {
			m.currentScreen = ScreenUninstall
			m.uninstall = screens.NewUninstallModel()
			return m, m.uninstall.Init()
		}
		return m, nil

	case screens.StartPipelineMsg:
		m.progressCh = make(chan pipeline.ProgressEvent, 100)
		m.latestProg = screens.ProgressMsg{TaskName: "Preparando pipeline...", Status: "Iniciando"}

		// Capture m.installCtx at closure creation time so it's safe to use inside goroutine.
		installCtx := m.installCtx

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

				// Extract AgentIDs from the install context for the resolver.
				agentIDs := make([]model.AgentID, len(ctx.IDEs))
				for i, ide := range ctx.IDEs {
					agentIDs[i] = ide.AgentID()
				}
				if len(agentIDs) == 0 {
					// No selected agents — treat as a fatal setup error.
					return screens.PipelineFinishedMsg{Err: fmt.Errorf("no agents selected for installation")}
				}

				resolver := planner.NewResolver(planner.MVPGraph())
				resolvedPlan, err := resolver.Resolve(model.Selection{
					Agents: agentIDs,
					Components: []model.ComponentID{
						model.ComponentSDDMemory,
						model.ComponentSDD,
						model.ComponentSkills,
						model.ComponentPersona,
					},
				})
				if err != nil {
					return screens.PipelineFinishedMsg{Err: err}
				}

				// Map ResolvedPlan to StagePlan manually
				var applySteps []pipeline.Step
				applySteps = append(applySteps, steps.NewStepInstallGlobalRules(ctx)) // Persona

				// Inject assets (Skills, SDD, sdd-memory config)
				injector := steps.NewAssetInjector(nil)
				applySteps = append(applySteps, steps.NewStepInjectAssets(ctx, injector, resolvedPlan))
				applySteps = append(applySteps, steps.NewStepInjectSDDMemory(ctx))
				applySteps = append(applySteps, steps.NewStepInjectOpenCodeOverlay(ctx))
				applySteps = append(applySteps, steps.NewStepInjectSlashCommands(ctx))
				applySteps = append(applySteps, steps.NewStepInjectMCP(ctx))

				prepareSteps := []pipeline.Step{}
				// Only scan if IDEs weren't pre-populated via agent selection.
				if len(ctx.IDEs) == 0 {
					prepareSteps = append(prepareSteps, steps.NewStepScanGlobalIDEs(ctx))
				}
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
				payload = &screens.CompletePayload{
					ConfiguredAgents:    len(installCtx.IDEs),
					InstalledComponents: len(plan.Apply),
					RollbackPerformed:   rollbackPerformed,
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
			m.currentScreen = ScreenComplete
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
