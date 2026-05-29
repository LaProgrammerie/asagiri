package app

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/agents"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/dashboard"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/explain"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/flows"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/graph"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/mission"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/prototype"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/replay"
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
	confirmation      *safetyConfirmation
	refreshEvery      time.Duration
	refreshThrottle   time.Duration
	lastRefreshAt     time.Time
	verticalSplit     float64
	horizontalSplit   float64
	mouseResizing     bool
}

const (
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
	return model{
		ctx:              ctx,
		now:              time.Now,
		cfg:              opts.Config,
		router:           newRouter(firstNonEmpty(opts.InitialScreen, opts.Config.DefaultScreen, ScreenMission)),
		state:            state.New(firstNonEmpty(opts.InitialScreen, opts.Config.DefaultScreen, ScreenMission)),
		theme:            th,
		layout:           layout.NewEngine(opts.Config.CompactThreshold),
		queryBus:         opts.QueryBus,
		commandBus:       opts.CommandBus,
		flowID:           strings.TrimSpace(opts.FlowID),
		replayID:         strings.TrimSpace(opts.ReplayID),
		prototypeProduct: strings.TrimSpace(opts.PrototypeProduct),
		lastError:        lastErr,
		paletteEntries:   defaultPaletteEntries(),
		refreshEvery:     time.Duration(refreshMs) * time.Millisecond,
		refreshThrottle: maxDuration(
			200*time.Millisecond,
			time.Duration(refreshMs/2)*time.Millisecond,
		),
		verticalSplit:   defaultSplitRatio,
		horizontalSplit: defaultSplitRatio,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.snapshotQueryCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(v)
	case tea.MouseMsg:
		return m.updateMouse(v)
	case tickMsg:
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
	default:
		return m, nil
	}
}

func (m model) updateKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := v.String()
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
	case keyNextPane:
		current := m.layout.Compute(m.currentLayout(), maxInt(1, m.width), maxInt(1, m.height))
		m.state.Focus = layout.NextFocus(current, m.state.Focus)
	case keyPrevPane:
		current := m.layout.Compute(m.currentLayout(), maxInt(1, m.width), maxInt(1, m.height))
		m.state.Focus = layout.PrevFocus(current, m.state.Focus)
	}
	return m, nil
}

func (m *model) navigateTo(screen string, cli string) {
	m.router.Set(screen)
	m.state.SetScreen(screen)
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
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.theme.Palette.Primary)).
		Render("ASAGIRI")
	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Muted)).
		Render(fmt.Sprintf("Screen: %s  |  Runtime: %s", m.router.Current(), m.runtimeStatusLabel()))

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
	frame := m.theme.BorderStyle().Padding(0, 1).Render(title + "\n" + statusLine + "\n\n" + body + "\n\n" + footer)
	return root.Render(frame)
}

func (m model) renderScreen() string {
	switch m.router.Current() {
	case ScreenDashboard:
		return dashboard.Render(dashboard.ViewModel{
			Snapshot: m.snapshot,
			Theme:    m.theme,
			Width:    m.width,
			Animated: m.cfg.Animations,
		})
	case ScreenAgents:
		content := agents.Render(agents.ViewModel{
			Theatre: m.snapshot.AgentTheatre,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Agent Theatre", content, m.theme)
	case ScreenGraph:
		content := graph.Render(graph.ViewModel{
			Graph:   m.snapshot.GraphExplorer,
			Events:  m.snapshot.Events,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Graph Explorer", content, m.theme)
	case ScreenFlow:
		content := flows.Render(flows.ViewModel{
			Flow:    m.snapshot.FlowExplorer,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Flow Explorer", content, m.theme)
	case ScreenLogs:
		return components.Panel("Logs", "Logs panel placeholder.\nUse Command Palette to trigger actions.\nCLI: asa logs", m.theme)
	case ScreenExplain:
		content := explain.Render(explain.ViewModel{
			Explain: m.snapshot.Explain,
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Explain", content, m.theme)
	case ScreenReplay:
		content := replay.Render(replay.ViewModel{
			Replay:  m.snapshot.Replay,
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
			ShowCLI: m.cfg.ShowCLIEquivalents,
		})
		return components.Panel("Knowledge Explorer", content, m.theme)
	case ScreenTrust:
		content := trust.Render(trust.ViewModel{
			Trust:   m.snapshot.TrustExplorer,
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
			QueuedEvents:      m.snapshot.Runtime.Status.QueuedEvents,
			CostTodayEUR:      m.snapshot.CostTodayEUR,
			CostMonthEUR:      m.snapshot.CostMonthEUR,
			Warnings:          m.snapshot.Warnings,
			Warning:           m.snapshot.Runtime.Warning,
			Now:               fallbackTime(m.snapshot.UpdatedAt, m.now().UTC()),
			DisableAnimations: !m.cfg.Animations,
		}
		return m.renderMission(vm)
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
			Knowledge:          m.snapshot.Flow.FlowID,
			ExplainFor:         m.router.Current(),
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

func (m model) currentLayout() layout.Kind {
	if m.router.Current() == ScreenSettings {
		return layout.Single
	}
	if m.width > m.layout.CompactThreshold {
		return layout.SplitVertical
	}
	return layout.SplitHorizontal
}

func (m model) runtimeStatusLabel() string {
	if m.snapshot.Runtime.Status.Running {
		return "running"
	}
	return "stopped"
}

func (m model) renderMission(vm mission.ViewModel) string {
	headerPanel := components.Panel("ASAGIRI", mission.RenderHeader(vm), m.theme)
	runtimePanel := components.Panel("Runtime", mission.RenderRuntimeRuns(vm), m.theme)
	trustPanel := components.Panel("Trust", mission.RenderTrustPane(vm), m.theme)
	flowPanel := components.Panel("Active Flow", mission.RenderActiveFlowPane(vm), m.theme)
	agentsPanel := components.Panel("Agent Theatre", mission.RenderAgentTheatrePane(vm), m.theme)
	eventsPanel := components.Panel("Events", mission.RenderEventsPane(vm), m.theme)

	panes := []string{headerPanel}
	if m.width >= m.layout.CompactThreshold {
		if almostEqual(m.verticalSplit, defaultSplitRatio) {
			panes = append(panes, lipgloss.JoinHorizontal(lipgloss.Top, runtimePanel, trustPanel))
		} else {
			leftWidth, rightWidth := m.verticalPaneWidths()
			left := lipgloss.NewStyle().Width(leftWidth).Render(runtimePanel)
			right := lipgloss.NewStyle().Width(rightWidth).Render(trustPanel)
			panes = append(panes, lipgloss.JoinHorizontal(lipgloss.Top, left, right))
		}
	} else {
		panes = append(panes, runtimePanel, trustPanel)
	}
	panes = append(panes, flowPanel, agentsPanel, eventsPanel)
	return strings.Join(panes, "\n")
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
	b.WriteString("- wheel up/down: resize current split\n")
	b.WriteString("- drag divider with left click: basic panel resize\n")
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
	if !m.cfg.Mouse || m.showHelp || m.showPalette || m.confirmation != nil {
		return m, nil
	}
	currentLayout := m.currentLayout()
	if currentLayout != layout.SplitVertical && currentLayout != layout.SplitHorizontal {
		return m, nil
	}

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
	return m, nil
}

func (m *model) setSplitFromPointer(kind layout.Kind, x, y int) {
	switch kind {
	case layout.SplitVertical:
		if m.width <= 0 {
			return
		}
		m.verticalSplit = clampSplit(float64(x) / float64(m.width))
		m.lastCommandResult = splitStatusText(m.verticalSplit)
	case layout.SplitHorizontal:
		if m.height <= 0 {
			return
		}
		m.horizontalSplit = clampSplit(float64(y) / float64(m.height))
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
