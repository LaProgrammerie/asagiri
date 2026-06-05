package app

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/version"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/agents"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/dashboard"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/explain"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/flows"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/graph"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/mission"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/prototype"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/replay"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/runs"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/settings"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/state"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Options configures Mission Control shell startup.
type Options struct {
	In               io.Reader
	Out              io.Writer
	Err              io.Writer
	Config           config.UIConfig
	InitialScreen    string
	FlowID           string
	ReplayID         string
	PrototypeProduct string
	CommandBus       bus.CommandBus
	QueryBus         bus.QueryBus
}

// Run starts the Bubble Tea lot-1 shell.
func Run(ctx context.Context, opts Options) error {
	return run(ctx, opts)
}

func run(ctx context.Context, opts Options) error {
	model := newModel(ctx, opts)
	programOpts := []tea.ProgramOption{
		tea.WithInput(opts.In),
		tea.WithOutput(opts.Out),
		tea.WithoutSignals(),
	}
	if model.wizardMode {
		programOpts = append(programOpts, tea.WithAltScreen())
	}
	if opts.Config.Mouse {
		programOpts = append(programOpts, tea.WithMouseCellMotion())
	}
	program := tea.NewProgram(model, programOpts...)
	_, err := program.Run()
	return err
}

type tickMsg time.Time
type snapshotMsg struct {
	result bus.MissionControlSnapshotResult
	err    error
}

type model struct {
	ctx context.Context
	now func() time.Time

	cfg    config.UIConfig
	router router
	state  state.UIState
	theme  theme.Theme
	layout layout.Engine

	queryBus         bus.QueryBus
	commandBus       bus.CommandBus
	flowID           string
	replayID         string
	prototypeProduct string

	width  int
	height int

	snapshot          bus.MissionControlSnapshotResult
	lastError         string
	lastWarning       string
	lastCommandResult string
	showHelp          bool
	showPalette       bool
	paletteQuery      string
	paletteCursor     int
	paletteEntries    []paletteEntry
	eventFeed         components.EventFeedModel
	eventFeedFocused  bool
	graphExplorer     graph.Model
	flowExplorer      flows.Model
	flowStep          bus.FlowStepDetail
	knowledgeExplorer knowledge.Model
	trustExplorer     trust.Model
	replayExplorer    replay.Model
	runsExplorer      runs.Model
	onboardingWizard  onboarding.Model
	wizardMode        bool
	confirmation      *safetyConfirmation
	refreshEvery      time.Duration
	refreshThrottle   time.Duration
	lastRefreshAt     time.Time
	verticalSplit     float64
	horizontalSplit   float64
	mouseResizing     bool
	mouse             input.MouseState
	animFrame         int
}

const (
	mouseListTopY        = 8
	defaultSplitRatio    = 0.50
	minSplitRatio        = 0.20
	maxSplitRatio        = 0.80
	splitResizeStep      = 0.05
	dividerHitTolerance  = 1
	minPaneWidthCells    = 24
	estimatedFrameMargin = 8
)

func newModel(ctx context.Context, opts Options) model {
	if ctx == nil {
		ctx = context.Background()
	}
	th, err := theme.Resolve(opts.Config.Theme)
	lastErr := ""
	if err != nil {
		th = theme.Default()
		lastErr = err.Error()
	}
	refreshMs := opts.Config.RefreshIntervalMs
	if refreshMs <= 0 {
		refreshMs = 500
	}
	uiState := state.New(firstNonEmpty(opts.InitialScreen, opts.Config.DefaultScreen, ScreenMission))
	return model{
		ctx:              ctx,
		now:              time.Now,
		cfg:              opts.Config,
		router:           newRouter(firstNonEmpty(opts.InitialScreen, opts.Config.DefaultScreen, ScreenMission)),
		state:            uiState,
		theme:            th,
		layout:           layout.NewEngine(opts.Config.CompactThreshold),
		queryBus:         opts.QueryBus,
		commandBus:       opts.CommandBus,
		flowID:           strings.TrimSpace(opts.FlowID),
		replayID:         strings.TrimSpace(opts.ReplayID),
		prototypeProduct: strings.TrimSpace(opts.PrototypeProduct),
		lastError:        lastErr,
		eventFeed:        components.NewEventFeedModel(),
		graphExplorer:    graph.NewModel(),
		flowExplorer:     flows.NewModel(),
		knowledgeExplorer: knowledge.NewModel(),
		trustExplorer:    trust.NewModel(),
		replayExplorer:    replay.NewModel(),
		runsExplorer:      runs.NewModel(),
		onboardingWizard:  onboarding.NewModel(),
		wizardMode:        opts.InitialScreen == ScreenOnboarding,
		refreshEvery: time.Duration(refreshMs) * time.Millisecond,
		refreshThrottle: maxDuration(
			200*time.Millisecond,
			time.Duration(refreshMs/2)*time.Millisecond,
		),
		verticalSplit:   uiState.Panels.SizeRatio(layout.PaneMain, defaultSplitRatio),
		horizontalSplit: uiState.Panels.SizeRatio(layout.PaneSide, defaultSplitRatio),
		mouse:           input.NewMouseState(),
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadOnboardingWizardCmd()}
	if !m.wizardMode {
		cmds = append(cmds, m.tickCmd(), m.snapshotQueryCmd())
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.syncScreenTabs()
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(v)
	case tea.MouseMsg:
		return m.updateMouse(v)
	case tickMsg:
		if m.cfg.Animations {
			m.animFrame++
		}
		if m.wizardMode {
			return m, m.tickCmd()
		}
		cmds := []tea.Cmd{m.tickCmd()}
		if m.canRefresh(time.Time(v)) {
			cmds = append(cmds, m.snapshotQueryCmd())
			m.lastRefreshAt = time.Time(v)
		}
		return m, tea.Batch(cmds...)
	case snapshotMsg:
		if v.err != nil {
			m.lastError = v.err.Error()
			return m, nil
		}
		if !snapshotContentEqual(m.snapshot, v.result) {
			m.snapshot = v.result
		}
		m.lastError = ""
		if len(v.result.Warnings) > 0 {
			m.lastWarning = v.result.Warnings[0]
		} else {
			m.lastWarning = ""
		}
		return m, nil
	case commandDispatchMsg:
		if v.err != nil {
			m.lastError = v.err.Error()
			return m, nil
		}
		m.lastCommandResult = v.result.Message
		if m.cfg.ShowCLIEquivalents && strings.TrimSpace(v.result.CLIEquivalent) != "" {
			m.lastCommandResult = v.result.Message + " | CLI: " + v.result.CLIEquivalent
		}
		return m, nil
	case paletteActionMsg:
		return m.handlePaletteAction(v)
	case onboardingWizardMsg:
		return m.handleOnboardingWizardLoaded(v)
	case onboarding.OnboardingFooterMsg:
		return m.handleOnboardingFooter(v)
	case onboardingAdvanceMsg:
		return m.handleOnboardingAdvance(v)
	case onboardingApplyMsg:
		return m.handleOnboardingApply(v)
	case onboarding.OnboardingAutofixMsg:
		return m.handleOnboardingAutofixRequest()
	case onboardingAutofixMsg:
		return m.handleOnboardingAutofix(v)
	default:
		return m, nil
	}
}

func (m model) updateKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := v.String()
	if m.wizardMode {
		switch key {
		case "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			return m.updateOnboardingKey(v)
		}
	}
	if m.mouse.ContextMenu != nil {
		switch key {
		case "up":
			input.MoveContextMenuSelection(m.mouse.ContextMenu, -1)
			return m, nil
		case "down":
			input.MoveContextMenuSelection(m.mouse.ContextMenu, 1)
			return m, nil
		case "enter":
			if item := input.SelectedMenuItem(m.mouse.ContextMenu); item != nil {
				m.mouse.ClearContextMenu()
				return m.runContextMenuItem(*item)
			}
			return m, nil
		case keyClose:
			m.mouse.ClearContextMenu()
			return m, nil
		}
	}

	if m.confirmation != nil {
		switch key {
		case keyClose, "n", "N":
			m.confirmation = nil
			m.lastCommandResult = "action cancelled"
			return m, nil
		case "y", "Y":
			actionID := m.confirmation.ActionID
			m.confirmation = nil
			return m, m.paletteActionCmd(actionID, "", true)
		}
		return m, nil
	}

	if m.showPalette {
		switch key {
		case keyClose:
			m.closePalette()
			return m, nil
		case "up":
			m.movePaletteCursor(-1)
			return m, nil
		case "down":
			m.movePaletteCursor(1)
			return m, nil
		case "enter":
			return m, m.paletteSelectCmd()
		case "backspace":
			if len(m.paletteQuery) > 0 {
				m.paletteQuery = m.paletteQuery[:len(m.paletteQuery)-1]
			}
			m.paletteCursor = 0
			return m, nil
		}
		if v.Type == tea.KeyRunes && len(v.Runes) > 0 {
			m.paletteQuery += string(v.Runes)
			m.paletteCursor = 0
		}
		return m, nil
	}

	if v.Type == tea.KeyCtrlM {
		m.navigateTo(ScreenMission, "asa mission")
		return m, nil
	}

	switch key {
	case keyQuit, "ctrl+c":
		return m, tea.Quit
	case keyPalette:
		m.openPalette()
	case keyMission:
		m.navigateTo(ScreenMission, "asa mission")
	case keyDashboard, "d":
		m.navigateTo(ScreenDashboard, "asa dashboard")
	case keyGraph:
		m.navigateTo(ScreenGraph, "asa graph")
	case keyFlow:
		m.navigateTo(ScreenFlow, "asa flow")
	case keyLogs:
		m.navigateTo(ScreenLogs, "asa logs")
	case keyExplain:
		m.navigateTo(ScreenExplain, "asa explain")
	case keyReplay:
		m.navigateTo(ScreenReplay, "asa replay open <replay-id>")
	case keyKnowledge:
		m.navigateTo(ScreenKnowledge, "asa knowledge")
	case keyTrust:
		m.navigateTo(ScreenTrust, "asa trust")
	case keySettings:
		m.navigateTo(ScreenSettings, "asa settings")
	case keyHelp:
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.showPalette = false
			m.confirmation = nil
		}
	case keyCycleLayout:
		m.cycleLayoutMode()
	case keyToggleCollapse:
		m.togglePaneCollapse()
	case keyFullscreen:
		m.toggleFullscreen()
	case "ctrl+]":
		m.nextScreenTab()
	case "ctrl+[":
		m.prevScreenTab()
	case keyNextPane:
		current := m.computeLayout()
		m.state.Focus = layout.NextFocus(current, m.state.Focus)
	case keyPrevPane:
		current := m.computeLayout()
		m.state.Focus = layout.PrevFocus(current, m.state.Focus)
	default:
		if m.onboardingInputActive() {
			return m.updateOnboardingKey(v)
		}
		if m.prototypeInputActive() {
			return m.updatePrototypeKey(v)
		}
		if m.explorerInputActive() {
			return m.updateExplorerKey(v)
		}
		return m.updateEventFeedKey(v)
	}
	return m, nil
}

func (m *model) navigateTo(screen string, cli string) {
	m.router.Set(screen)
	m.state.SetScreen(screen)
	m.syncScreenTabs()
	if m.cfg.ShowCLIEquivalents && strings.TrimSpace(cli) != "" {
		m.lastCommandResult = "open " + screen + " | CLI: " + cli
	}
}

func (m model) View() string {
	if m.width == 0 {
		m.width = 100
	}
	if m.height == 0 {
		m.height = 30
	}
	root := lipgloss.NewStyle().Padding(0, 1)
	st := m.theme.Styles()
	heroTitle := "Mission Control"
	if m.wizardMode {
		heroTitle = "Project Onboarding"
	}
	title := st.RenderHero("ASAGIRI", heroTitle, fmt.Sprintf("Screen · %s", m.router.Current()))
	statusBar := st.RenderStatusBar(
		"RUNTIME",
		m.runtimeStatusLabel(),
		fmt.Sprintf("Screen: %s", m.router.Current()),
	)

	body := m.renderScreen()
	if m.showHelp {
		body = m.renderHelpScreen()
	}
	if m.showPalette {
		body += "\n\n" + m.renderPalette()
	}
	if m.confirmation != nil {
		body += "\n\n" + m.renderSafetyConfirmation()
	}
	if menu := m.renderContextMenu(); menu != "" {
		body += "\n\n" + menu
	}
	footer := "Screen: " + m.router.Current()
	if m.cfg.ShowCLIEquivalents {
		footer += "  |  CLI: asa " + m.router.Current()
	}
	if m.lastError != "" {
		footer += "  |  error: " + m.lastError
	}
	if m.lastWarning != "" {
		footer += "  |  warn: " + m.lastWarning
	}
	if m.lastCommandResult != "" {
		footer += "  |  action: " + m.lastCommandResult
	}
	tabBar := m.renderTabBar()
	frameBody := title + "\n" + statusBar
	if tabBar != "" {
		frameBody += "\n\n" + tabBar
	}
	middle := body
	if rw := m.railWidth(); rw > 0 {
		rail := m.renderNavRail(rw, lipgloss.Height(body))
		middle = lipgloss.JoinHorizontal(lipgloss.Top, rail, body)
	}
	frameBody += "\n\n" + middle + "\n\n" + st.Hint.Render(footer)
	frame := st.Theme.BorderStyle().
		Padding(0, 1).
		Background(lipgloss.Color(m.theme.Palette.Background)).
		Render(frameBody)
	return root.Render(frame)
}

func (m model) renderOnboardingBody() string {
	m.ensureOnboardingWizard()
	w := m.onboardingWizard
	bodyH := m.height - 8
	if bodyH < 12 {
		bodyH = 12
	}
	inner := onboarding.Render(onboarding.ViewModel{
		Model:      w,
		Readiness:  m.snapshot.Readiness,
		ShowCLI:    m.cfg.ShowCLIEquivalents,
		WizardMode: m.wizardMode,
		InAppShell: m.wizardMode,
		FullScreen: m.wizardMode,
		Width:      m.bodyWidth(),
		Height:     bodyH,
		Theme:      m.theme,
		Shell: onboarding.ShellContext{
			Workspace:    firstNonEmpty(m.snapshot.Workspace, w.Fields["project_name"]),
			Branch:       firstNonEmpty(m.snapshot.Branch, w.Fields["default_branch"], "main"),
			Directory:    firstNonEmpty(w.Fields["project_name"], "."),
			Clock:        m.now().Format("15:04:05"),
			Version:      version.Version,
			CostTodayEUR: m.snapshot.CostTodayEUR,
			APIProvider:  "OpenRouter",
			Online:       true,
		},
	})
	if m.lastError != "" {
		errLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.Palette.Error)).
			Render("Erreur: " + m.lastError)
		inner += "\n" + errLine
	}
	return inner
}

func (m model) renderScreen() string {
	switch m.router.Current() {
	case ScreenDashboard:
		return dashboard.Render(dashboard.ViewModel{
			Snapshot:  m.snapshot,
			EventFeed: m.eventFeedViewModel(4),
			Theme:     m.theme,
			Width:     m.width,
			Animated:  m.cfg.Animations,
			AnimFrame: m.animFrame,
			TabIndex:  m.activeScreenTab(),
			TabLabels: m.screenTabLabels(),
			Compact:   m.layout.CompactThreshold,
		})
	case ScreenRuns:
		return runs.Render(runs.ViewModel{
			Runs:      m.snapshot.Runs,
			Detail:    m.currentRunDetail(),
			Model:     m.runsExplorer,
			Readiness: m.snapshot.Readiness,
			ShowCLI:   m.cfg.ShowCLIEquivalents,
			Width:     m.bodyWidth(),
			Height:    m.height,
			Theme:     m.theme,
		})
	case ScreenAgents:
		content := agents.Render(agents.ViewModel{
			Theatre: m.snapshot.AgentTheatre,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Agent Theatre", content, m.theme)
	case ScreenGraph:
		content := graph.Render(graph.ViewModel{
			Graph:      m.snapshot.GraphExplorer,
			View:       m.refreshGraphView(),
			Events:     m.snapshot.Events,
			NodeDetail: m.graphExplorer.Detail,
			Model:      m.graphExplorer,
			ShowCLI:    m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Graph Explorer", content, m.theme)
	case ScreenFlow:
		step := m.flowStep
		if step.ID == "" {
			step = m.queryFlowStepDetail(m.snapshot.FlowExplorer.FlowID, m.flowExplorer.SelectedStepID(m.snapshot.FlowExplorer))
		}
		content := flows.Render(flows.ViewModel{
			Flow:    m.snapshot.FlowExplorer,
			Step:    step,
			Model:   m.flowExplorer,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Flow Explorer", content, m.theme)
	case ScreenLogs:
		return components.Panel("Logs", m.renderLogs(), m.theme)
	case ScreenExplain:
		content := explain.Render(explain.ViewModel{
			Explain: m.snapshot.Explain,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Explain", content, m.theme)
	case ScreenReplay:
		content := replay.Render(replay.ViewModel{
			Replay:  m.snapshot.Replay,
			Detail:  m.replayExplorer.Detail,
			Compare: m.replayExplorer.Compare,
			Model:   m.replayExplorer,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Replay Explorer", content, m.theme)
	case ScreenPrototype:
		content := prototype.Render(prototype.ViewModel{
			Pipeline: m.snapshot.Prototype,
			ShowCLI:  m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Prototype Mode", content, m.theme)
	case ScreenKnowledge:
		content := knowledge.Render(knowledge.ViewModel{
			Search:  m.snapshot.Knowledge,
			Detail:  m.knowledgeExplorer.Detail,
			Model:   m.knowledgeExplorer,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Knowledge Explorer", content, m.theme)
	case ScreenTrust:
		content := trust.Render(trust.ViewModel{
			Trust:   m.snapshot.TrustExplorer,
			Detail:  m.trustExplorer.Detail,
			Model:   m.trustExplorer,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Trust Explorer", content, m.theme)
	case ScreenSettings:
		content := settings.Render(settings.ViewModel{
			ThemeName:       m.theme.Name,
			MouseEnabled:    m.cfg.Mouse,
			Animations:      m.cfg.Animations,
			AvailableThemes: theme.Names(),
		})
		return components.Panel("Settings", content, m.theme)
	case ScreenOnboarding:
		if m.wizardMode {
			return m.renderOnboardingBody()
		}
		m.ensureOnboardingWizard()
		content := onboarding.Render(onboarding.ViewModel{
			Model:      m.onboardingWizard,
			Readiness:  m.snapshot.Readiness,
			ShowCLI:    m.cfg.ShowCLIEquivalents,
			WizardMode: m.wizardMode,
		})
		return components.Panel("Onboarding", content, m.theme)
	default:
		vm := mission.ViewModel{
			Workspace:         m.snapshot.Workspace,
			Branch:            m.snapshot.Branch,
			SessionStatus:     m.snapshot.SessionStatus,
			RuntimeStatus:     m.runtimeStatusLabel(),
			ActiveAgents:      m.snapshot.ActiveAgents,
			Trust:             m.snapshot.Trust,
			Flow:              m.snapshot.Flow,
			Runs:              m.snapshot.Runs,
			Events:            m.snapshot.Events,
			EventFeed:         m.eventFeedViewModel(5),
			QueuedEvents:      m.snapshot.Runtime.Status.QueuedEvents,
			CostTodayEUR:      m.snapshot.CostTodayEUR,
			CostMonthEUR:      m.snapshot.CostMonthEUR,
			Warnings:          m.snapshot.Warnings,
			Warning:           m.snapshot.Runtime.Warning,
			Recommended:       m.snapshot.RecommendedActions,
			Readiness:         m.snapshot.Readiness,
			Now:               fallbackTime(m.snapshot.UpdatedAt, m.now().UTC()),
			DisableAnimations: !m.cfg.Animations,
			AnimFrame:         m.animFrame,
			Width:             m.bodyWidth(),
			Height:            m.height,
			CompactThreshold:  m.layout.CompactThreshold,
			Theme:             m.theme,
		}
		return m.renderMissionWithTabs(vm)
	}
}

func (m model) renderPalette() string {
	var b strings.Builder
	b.WriteString("Command Palette (Ctrl+P)\n")
	b.WriteString("> " + m.paletteQuery + "\n")
	entries := m.filteredPaletteEntries()
	if len(entries) == 0 {
		b.WriteString("- no results")
		return b.String()
	}
	for i, entry := range entries {
		prefix := "  "
		if i == m.paletteCursor {
			prefix = "> "
		}
		b.WriteString(prefix + entry.Title + "\n")
		b.WriteString("    " + entry.Description + "\n")
		if m.cfg.ShowCLIEquivalents && entry.CLI != "" {
			b.WriteString("    CLI: " + entry.CLI + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m model) renderSafetyConfirmation() string {
	var b strings.Builder
	b.WriteString("Safety confirmation\n")
	b.WriteString(m.confirmation.Title + "\n")
	b.WriteString("Impacted:\n")
	for _, impact := range m.confirmation.Impact {
		b.WriteString("- " + impact + "\n")
	}
	if m.cfg.ShowCLIEquivalents && m.confirmation.CLIEquivalent != "" {
		b.WriteString("CLI equivalent:\n")
		b.WriteString(m.confirmation.CLIEquivalent + "\n")
	}
	if m.confirmation.RollbackPolicy != "" {
		b.WriteString("Rollback:\n")
		b.WriteString(m.confirmation.RollbackPolicy + "\n")
	}
	b.WriteString("Proceed? [y/N]")
	return strings.TrimRight(b.String(), "\n")
}

func (m model) snapshotQueryCmd() tea.Cmd {
	return func() tea.Msg {
		if m.queryBus == nil {
			return snapshotMsg{}
		}
		res, err := m.queryBus.Query(m.ctx, bus.GetMissionControlSnapshotQuery{
			RunsLimit:          8,
			EventsLimit:        12,
			AgentsLimit:        200,
			Knowledge:          firstNonEmpty(m.knowledgeExplorer.Query, m.snapshot.Flow.FlowID),
			ExplainFor:         m.explainSubject(),
			ExplainContext:     m.explainContext(),
			FlowID:             m.flowID,
			ReplayID:           m.replayID,
			PrototypeProduct:   m.prototypeProduct,
			PrototypeFlowLimit: 24,
		})
		if err != nil {
			return snapshotMsg{err: err}
		}
		typed, ok := res.(bus.MissionControlSnapshotResult)
		if !ok {
			return snapshotMsg{err: fmt.Errorf("unexpected snapshot query result: %T", res)}
		}
		return snapshotMsg{result: typed}
	}
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshEvery, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) renderLogs() string {
	lines := logLinesFromEvents(m.snapshot.Events)
	return components.RenderLogView(components.LogViewModel{
		Lines:  lines,
		Cursor: m.mouse.HoverRow,
		Limit:  maxInt(8, m.height-6),
	})
}

func (m model) runtimeStatusLabel() string {
	if m.snapshot.Runtime.Status.Running {
		return "running"
	}
	return "stopped"
}

func (m model) renderHelpScreen() string {
	var b strings.Builder
	b.WriteString("Keyboard\n")
	b.WriteString("- ctrl+p: command palette\n")
	b.WriteString("- ctrl+m: mission control\n")
	b.WriteString("- ctrl+d: dashboard\n")
	b.WriteString("- ctrl+g: graph explorer\n")
	b.WriteString("- ctrl+f: flow explorer\n")
	b.WriteString("- ctrl+k: knowledge explorer\n")
	b.WriteString("- ctrl+t: trust explorer\n")
	b.WriteString("- ctrl+e: explain\n")
	b.WriteString("- ctrl+r: replay explorer\n")
	b.WriteString("- ctrl+l: logs\n")
	b.WriteString("- tab / shift+tab: move focus\n")
	b.WriteString("- ctrl+[ / ctrl+]: previous / next layout tab\n")
	b.WriteString("- ctrl+shift+l: cycle layout mode\n")
	b.WriteString("- ctrl+\\: collapse/expand focused pane\n")
	b.WriteString("- f11: toggle fullscreen layout\n")
	b.WriteString("- esc: close modal or palette\n")
	b.WriteString("- q: quit\n")
	b.WriteString("\nAccessibility\n")
	b.WriteString("- theme: " + m.theme.Name + "\n")
	if m.theme.IsHighContrast() {
		b.WriteString("- high-contrast mode: active\n")
	} else {
		b.WriteString("- high-contrast mode: available via ui.theme=high-contrast\n")
	}
	b.WriteString("- animations: " + onOffLabel(m.cfg.Animations) + "\n")
	b.WriteString("- mouse support: " + onOffLabel(m.cfg.Mouse) + "\n")
	b.WriteString("- plain/json mode: available with ui.mode=plain|json for CI/screen readers\n")
	b.WriteString("\nMouse\n")
	b.WriteString("- left click: select list row / recommended action\n")
	b.WriteString("- double click: open detail or explain focused item\n")
	b.WriteString("- right click: context menu (explorers, mission, prototype)\n")
	b.WriteString("- wheel up/down: resize current split\n")
	b.WriteString("- drag divider with left click: panel resize\n")
	return components.Panel("Accessibility & key help", b.String(), m.theme)
}

func onOffLabel(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func (m model) verticalPaneWidths() (int, int) {
	available := maxInt(minPaneWidthCells*2, m.width-estimatedFrameMargin)
	left := int(float64(available) * m.verticalSplit)
	if left < minPaneWidthCells {
		left = minPaneWidthCells
	}
	right := available - left
	if right < minPaneWidthCells {
		right = minPaneWidthCells
		left = available - right
	}
	return left, right
}

func (m model) updateMouse(v tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.cfg.Mouse {
		return m, nil
	}
	if m.showHelp || m.showPalette {
		return m, nil
	}
	if m.confirmation != nil {
		return m, nil
	}
	if m.mouse.ContextMenu != nil {
		return m.updateContextMenuMouse(v)
	}

	currentLayout := m.currentLayout()
	if currentLayout == layout.SplitVertical || currentLayout == layout.SplitHorizontal {
		switch {
		case v.Button == tea.MouseButtonWheelUp && v.Action == tea.MouseActionPress:
			if m.adjustSplit(currentLayout, splitResizeStep) {
				return m, nil
			}
		case v.Button == tea.MouseButtonWheelDown && v.Action == tea.MouseActionPress:
			if m.adjustSplit(currentLayout, -splitResizeStep) {
				return m, nil
			}
		case v.Button == tea.MouseButtonLeft && v.Action == tea.MouseActionPress:
			if m.isOnDivider(currentLayout, v.X, v.Y) {
				m.mouseResizing = true
				m.setSplitFromPointer(currentLayout, v.X, v.Y)
				return m, nil
			}
		case v.Button == tea.MouseButtonLeft && v.Action == tea.MouseActionMotion && m.mouseResizing:
			m.setSplitFromPointer(currentLayout, v.X, v.Y)
			return m, nil
		case (v.Action == tea.MouseActionRelease || (v.Button == tea.MouseButtonNone && v.Action == tea.MouseActionMotion)) && m.mouseResizing:
			m.mouseResizing = false
			return m, nil
		}
	}

	switch v.Action {
	case tea.MouseActionMotion:
		if v.Button == tea.MouseButtonNone {
			m.mouse.HoverRow = input.ListRowFromY(v.Y, mouseListTopY)
			m.mouse.HoverCol = v.X
		}
	case tea.MouseActionPress:
		if v.Button == tea.MouseButtonRight {
			items := m.contextMenuItems()
			m.mouse.ContextMenu = input.OpenContextMenu(v.X, v.Y, items)
			return m, nil
		}
		if v.Button == tea.MouseButtonLeft {
			if screen := m.railScreenAt(v.X, v.Y); screen != "" {
				m.navigateTo(screen, "asa "+screen)
				return m, nil
			}
			if input.IsDoubleClick(m.mouse.LastClickAt, m.mouse.LastClickX, m.mouse.LastClickY, v.X, v.Y, time.Now()) {
				return m.handleMouseDoubleClick(v)
			}
			m.mouse.LastClickAt = time.Now()
			m.mouse.LastClickX = v.X
			m.mouse.LastClickY = v.Y
			if m.onboardingInputActive() {
				return m.updateOnboardingMouse(v)
			}
			m.selectExplorerRowFromMouse(v.Y)
			if m.eventFeedScreenActive() {
				m.eventFeedFocused = true
				m.eventFeed.Focused = true
				m.eventFeed.SelectRow(v.Y)
			}
		}
	}
	return m, nil
}

func (m model) updateContextMenuMouse(v tea.MouseMsg) (tea.Model, tea.Cmd) {
	menu := m.mouse.ContextMenu
	if menu == nil {
		return m, nil
	}
	switch v.Action {
	case tea.MouseActionPress:
		if v.Button == tea.MouseButtonLeft {
			row := v.Y - menu.Y
			if row >= 0 && row < len(menu.Items) {
				item := menu.Items[row]
				m.mouse.ClearContextMenu()
				return m.runContextMenuItem(item)
			}
			m.mouse.ClearContextMenu()
		}
		if v.Button == tea.MouseButtonRight {
			m.mouse.ClearContextMenu()
		}
	case tea.MouseActionMotion:
		row := v.Y - menu.Y
		if row >= 0 && row < len(menu.Items) {
			menu.Selected = row
		}
	}
	return m, nil
}

func (m *model) selectExplorerRowFromMouse(y int) {
	row := input.ListRowFromY(y, mouseListTopY)
	if row < 0 {
		return
	}
	m.mouse.SelectRow(row)
	switch m.router.Current() {
	case ScreenGraph:
		nodes := m.refreshGraphView().Nodes
		if len(nodes) == 0 {
			nodes = m.snapshot.GraphExplorer.Nodes
		}
		m.graphExplorer.SelectIndex(row, len(nodes))
	case ScreenFlow:
		m.flowExplorer.SelectIndex(row, len(m.snapshot.FlowExplorer.Steps))
	case ScreenKnowledge:
		m.knowledgeExplorer.SelectIndex(row, len(m.snapshot.Knowledge.Matches))
	case ScreenTrust:
		m.trustExplorer.SelectIndex(row, len(m.snapshot.TrustExplorer.Dimensions))
	case ScreenReplay:
		m.replayExplorer.SelectIndex(row, len(m.snapshot.Replay.Timeline))
	case ScreenRuns:
		m.runsExplorer.SelectIndex(row, len(m.snapshot.Runs))
	}
}

func (m model) runContextMenuItem(item input.MenuItem) (tea.Model, tea.Cmd) {
	if item.ID == "nav.explain" {
		m.openExplainForFocus(bus.FocusKindDecision, m.explainSubject(), m.router.Current())
		return m, nil
	}
	return m.runPaletteAction(item.ID, item.CLI)
}

func (m model) handleMouseDoubleClick(v tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch m.router.Current() {
	case ScreenGraph:
		selected := m.graphExplorer.SelectedNode(m.refreshGraphView(), m.snapshot.GraphExplorer.Nodes)
		if selected == nil {
			return m, nil
		}
		detail := m.queryGraphNodeDetail(m.snapshot.GraphExplorer.GraphID, selected.ID)
		m.graphExplorer.Detail = detail
		m.graphExplorer.ShowDetail = true
		return m, nil
	case ScreenFlow:
		stepID := m.flowExplorer.SelectedStepID(m.snapshot.FlowExplorer)
		if stepID == "" {
			return m, nil
		}
		m.flowStep = m.queryFlowStepDetail(m.snapshot.FlowExplorer.FlowID, stepID)
		return m, nil
	case ScreenTrust:
		label := m.trustExplorer.SelectedLabel(m.snapshot.TrustExplorer)
		if label == "" {
			return m, nil
		}
		m.openExplainForFocus(bus.FocusKindTrustDimension, label, "")
		return m, nil
	case ScreenKnowledge:
		match := m.knowledgeExplorer.SelectedMatch(m.snapshot.Knowledge)
		if match == nil {
			return m, nil
		}
		m.openExplainForFocus(bus.FocusKindDecision, match.Name, match.ID)
		return m, nil
	case ScreenReplay:
		idx := m.replayExplorer.SelectedEventIndex(len(m.snapshot.Replay.Timeline))
		m.openExplainForFocus(bus.FocusKindReplayEvent, m.snapshot.Replay.ReplayID, fmt.Sprintf("event-%d", idx))
		return m, nil
	case ScreenMission, ScreenDashboard:
		m.eventFeedFocused = true
		m.eventFeed.Focused = true
		ev := m.eventFeed.SelectedEvent(m.snapshot.Events)
		if ev == nil {
			if m.router.Current() == ScreenMission {
				row := input.ListRowFromY(v.Y, mouseListTopY)
				if row >= 0 && row < len(m.snapshot.RecommendedActions) {
					action := m.snapshot.RecommendedActions[row]
					return m.runPaletteAction(action.ActionID, action.CLIEquivalent)
				}
			}
			return m, nil
		}
		screen, cli := components.ArtifactNavigation(*ev)
		m.navigateTo(screen, cli)
		return m, nil
	}
	return m, nil
}

func (m model) contextMenuItems() []input.MenuItem {
	screen := m.router.Current()
	items := []input.MenuItem{
		{ID: "nav.explain", Label: "Explain selection", CLI: "asa explain"},
	}
	switch screen {
	case ScreenGraph:
		items = append(items,
			input.MenuItem{ID: "ctx.graph-resume", Label: "Resume graph", CLI: "asa graph resume <graph-id>"},
			input.MenuItem{ID: "ctx.graph-export", Label: "Export graph", CLI: "asa graph visualize <graph-id> --format mermaid"},
		)
	case ScreenFlow:
		items = append(items, input.MenuItem{ID: "nav.flow", Label: "Open flow detail", CLI: "asa flow"})
	case ScreenTrust:
		items = append(items, input.MenuItem{ID: "cmd.verify-trust", Label: "Verify trust", CLI: "asa verify trust onboarding"})
	case ScreenPrototype:
		items = append(items,
			input.MenuItem{ID: "cmd.prototype-create", Label: "Prototype create", CLI: `asa prototype create "<intent>"`},
			input.MenuItem{ID: "cmd.flows-extract", Label: "Flows extract", CLI: "asa flows extract <product>"},
			input.MenuItem{ID: "cmd.contracts-extract", Label: "Contracts extract", CLI: "asa contracts extract <product>"},
			input.MenuItem{ID: "cmd.spec-generate", Label: "Spec generate from product", CLI: "asa spec generate-from-product <product>"},
		)
	case ScreenReplay:
		items = append(items, input.MenuItem{ID: "ctx.replay-run", Label: "Replay offline", CLI: "asa replay run --offline"})
	}
	return items
}

func (m model) renderContextMenu() string {
	menu := m.mouse.ContextMenu
	if menu == nil || len(menu.Items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nContext menu\n")
	for i, item := range menu.Items {
		prefix := "  "
		if i == menu.Selected {
			prefix = "> "
		}
		b.WriteString(prefix + item.Label + "\n")
		if m.cfg.ShowCLIEquivalents && item.CLI != "" {
			b.WriteString("    CLI: " + item.CLI + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m *model) setSplitFromPointer(kind layout.Kind, x, y int) {
	switch kind {
	case layout.SplitVertical:
		if m.width <= 0 {
			return
		}
		m.verticalSplit = clampSplit(float64(x) / float64(m.width))
		m.state.Panels.SetSizeRatio(layout.PaneMain, m.verticalSplit)
		m.lastCommandResult = splitStatusText(m.verticalSplit)
	case layout.SplitHorizontal:
		if m.height <= 0 {
			return
		}
		m.horizontalSplit = clampSplit(float64(y) / float64(m.height))
		m.state.Panels.SetSizeRatio(layout.PaneSide, m.horizontalSplit)
		m.lastCommandResult = splitStatusText(m.horizontalSplit)
	}
}

func (m *model) adjustSplit(kind layout.Kind, delta float64) bool {
	switch kind {
	case layout.SplitVertical:
		next := clampSplit(m.verticalSplit + delta)
		if next == m.verticalSplit {
			return false
		}
		m.verticalSplit = next
		m.lastCommandResult = splitStatusText(next)
		return true
	case layout.SplitHorizontal:
		next := clampSplit(m.horizontalSplit + delta)
		if next == m.horizontalSplit {
			return false
		}
		m.horizontalSplit = next
		m.lastCommandResult = splitStatusText(next)
		return true
	default:
		return false
	}
}

func splitStatusText(ratio float64) string {
	mainPct := int(ratio * 100)
	return fmt.Sprintf("pane resize: %d/%d", mainPct, 100-mainPct)
}

func clampSplit(v float64) float64 {
	if v < minSplitRatio {
		return minSplitRatio
	}
	if v > maxSplitRatio {
		return maxSplitRatio
	}
	return v
}

func almostEqual(a, b float64) bool {
	const epsilon = 0.001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= epsilon
}

func (m model) isOnDivider(kind layout.Kind, x, y int) bool {
	switch kind {
	case layout.SplitVertical:
		if m.width <= 0 || y < 0 || y >= m.height {
			return false
		}
		divider := int(float64(m.width) * m.verticalSplit)
		return x >= divider-dividerHitTolerance && x <= divider+dividerHitTolerance
	case layout.SplitHorizontal:
		if m.height <= 0 || x < 0 || x >= m.width {
			return false
		}
		divider := int(float64(m.height) * m.horizontalSplit)
		return y >= divider-dividerHitTolerance && y <= divider+dividerHitTolerance
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) canRefresh(now time.Time) bool {
	if m.lastRefreshAt.IsZero() {
		return true
	}
	return now.Sub(m.lastRefreshAt) >= m.refreshThrottle
}

func fallbackTime(primary time.Time, fallback time.Time) time.Time {
	if primary.IsZero() {
		return fallback
	}
	return primary
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func snapshotContentEqual(a, b bus.MissionControlSnapshotResult) bool {
	aCopy := a
	bCopy := b
	aCopy.UpdatedAt = time.Time{}
	bCopy.UpdatedAt = time.Time{}
	return reflect.DeepEqual(aCopy, bCopy)
}
