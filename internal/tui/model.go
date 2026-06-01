package tui

import (
	"time"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/pipeline"
	"github.com/KevG1t/SpecAI/internal/steps"
	"github.com/KevG1t/SpecAI/internal/tui/screens"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressMsg struct {
	TaskName string
	Status   string
	Progress float64
}

type PipelineDoneMsg struct {
	Error error
}

type MainModel struct {
	currentScreen Screen
	welcome       screens.WelcomeModel
	
	install    screens.InstallModel
	setupLocal screens.SetupLocalModel
	upgrade    screens.UpgradeModel
	sync       screens.SyncModel
	upgSync    screens.UpgradeSyncModel
	uninstall  screens.UninstallModel

	selectedOpt   string
	progressCh    chan pipeline.ProgressEvent
	latestProg    ProgressMsg

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
		return ProgressMsg{
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

	case screens.OptionSelectedMsg:
		m.selectedOpt = msg.Option
		if msg.Option == "Backup" {
			m.currentScreen = ScreenBackups
			m.Backups = ListBackups()
			m.SelectedBackup = 0
			m.BackupScroll = 0
			return m, nil
		} else if msg.Option == "Install" {
			m.currentScreen = ScreenInstall
			m.install = screens.NewInstallModel()
			return m, m.install.Init()
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
		m.latestProg = ProgressMsg{TaskName: "Preparando pipeline...", Status: "Iniciando"}

		startCmd := func() tea.Msg {
			var plan pipeline.StagePlan

			if msg.Action == "Install" {
				ctx, errCtx := steps.NewInstallContext()
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanGlobalIDEs(ctx)},
					Apply:   []pipeline.Step{steps.NewStepInstallGlobalRules(ctx), steps.NewStepInstallGlobalSkills(ctx)},
				}
			} else if msg.Action == "SetupLocal" {
				ctx, errCtx := steps.NewInstallContext()
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanGlobalIDEs(ctx)},
					Apply:   []pipeline.Step{steps.NewStepSetupLocalRules(ctx)},
				}
			} else if msg.Action == "Sync" {
				syncCtx, errCtx := steps.NewSyncContext()
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepScanIDEsForSync(syncCtx), steps.NewStepSnapshotBeforeSync(syncCtx)},
					Apply:   []pipeline.Step{steps.NewStepSyncGlobalRules(syncCtx), steps.NewStepSyncGlobalSkills(syncCtx)},
				}
			} else if msg.Action == "Upgrade" {
				upgradeCtx, errCtx := steps.NewUpgradeContext(GetVersion())
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepCheckForUpdates(upgradeCtx)},
					Apply:   []pipeline.Step{steps.NewStepInstallUpdates(upgradeCtx)},
				}
			} else if msg.Action == "UpgradeSync" {
				upgradeCtx, errCtx := steps.NewUpgradeContext(GetVersion())
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
				}
				syncCtx, errCtx2 := steps.NewSyncContext()
				if errCtx2 != nil {
					return PipelineDoneMsg{Error: errCtx2}
				}
				plan = pipeline.StagePlan{
					Prepare: []pipeline.Step{steps.NewStepCheckForUpdates(upgradeCtx), steps.NewStepScanIDEsForSync(syncCtx), steps.NewStepSnapshotBeforeSync(syncCtx)},
					Apply:   []pipeline.Step{steps.NewStepInstallUpdates(upgradeCtx), steps.NewStepSyncGlobalRules(syncCtx), steps.NewStepSyncGlobalSkills(syncCtx)},
				}
			} else if msg.Action == "Uninstall" {
				uninstallCtx, errCtx := steps.NewUninstallContext()
				if errCtx != nil {
					return PipelineDoneMsg{Error: errCtx}
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
			return PipelineDoneMsg{Error: res.Err}
		}
		return m, tea.Batch(startCmd, waitForProgress(m.progressCh))

	case ProgressMsg:
		m.latestProg = msg
		return m, waitForProgress(m.progressCh)

	case PipelineDoneMsg:
		// Dispatch to the active module screen
		return m, func() tea.Msg {
			return screens.PipelineFinishedMsg{Err: msg.Error}
		}
	}

	// Route message to active screen if applicable
	switch m.currentScreen {
	case ScreenWelcome:
		var wModel tea.Model
		wModel, cmd = m.welcome.Update(msg)
		m.welcome = wModel.(screens.WelcomeModel)
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
	case ScreenBackups, ScreenRestoreConfirm, ScreenDeleteConfirm, ScreenBackupResult:
		return m.ViewBackups()
	default:
		return "Pantalla desconocida"
	}
}
