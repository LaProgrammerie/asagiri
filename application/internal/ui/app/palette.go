package app

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
)

type paletteEntry struct {
	ID          string
	Title       string
	Description string
	CLI         string
	Keywords    []string
}

type safetyConfirmation struct {
	Title          string
	Impact         []string
	CLIEquivalent  string
	RollbackPolicy string
	ActionID       string
}

type paletteActionMsg struct {
	actionID         string
	cliEquivalent    string
	skipSafetyPrompt bool
}

type commandDispatchMsg struct {
	result bus.CommandResult
	err    error
}

func defaultPaletteEntries() []paletteEntry {
	return []paletteEntry{
		{ID: "nav.dashboard", Title: "Open dashboard", Description: "Navigate to dashboard live view", CLI: "asa dashboard", Keywords: []string{"screen", "dashboard", "nav"}},
		{ID: "nav.mission", Title: "Open mission control", Description: "Navigate to mission control", CLI: "asa mission", Keywords: []string{"screen", "mission", "nav"}},
		{ID: "nav.agents", Title: "Open agent theatre", Description: "Navigate to live agent cards", CLI: "asa agents watch", Keywords: []string{"screen", "agents", "watch", "nav"}},
		{ID: "nav.graph", Title: "Open graph explorer", Description: "Navigate to graph explorer", CLI: "asa graph", Keywords: []string{"screen", "graph", "nav"}},
		{ID: "nav.flow", Title: "Open flow explorer", Description: "Navigate to flow explorer", CLI: "asa flow", Keywords: []string{"screen", "flow", "nav"}},
		{ID: "nav.logs", Title: "Open logs", Description: "Navigate to logs placeholder", CLI: "asa logs", Keywords: []string{"screen", "logs", "nav"}},
		{ID: "nav.explain", Title: "Open explain panel", Description: "Navigate to explainability panel", CLI: "asa explain", Keywords: []string{"screen", "explain", "nav"}},
		{ID: "nav.replay", Title: "Open replay", Description: "Navigate to replay explorer", CLI: "asa replay open <replay-id>", Keywords: []string{"screen", "replay", "nav"}},
		{ID: "nav.prototype", Title: "Open prototype mode", Description: "Navigate to prototype split view", CLI: "asa prototype", Keywords: []string{"screen", "prototype", "nav"}},
		{ID: "nav.knowledge", Title: "Open knowledge", Description: "Navigate to knowledge explorer", CLI: "asa knowledge", Keywords: []string{"screen", "knowledge", "nav"}},
		{ID: "nav.trust", Title: "Open trust explorer", Description: "Navigate to trust explorer", CLI: "asa trust", Keywords: []string{"screen", "trust", "nav"}},
		{ID: "flow.open-onboarding", Title: "Open flow onboarding", Description: "Open onboarding flow details", CLI: "asa flow open onboarding", Keywords: []string{"flow", "open", "onboarding"}},
		{ID: "cmd.start-work", Title: "Start work", Description: "Run workflow orchestration from intent", CLI: `asa work "add workspace invitations"`, Keywords: []string{"work", "implement", "dev"}},
		{ID: "cmd.run-investigation", Title: "Run investigation", Description: "Investigate onboarding failures", CLI: `asa investigate "onboarding fails"`, Keywords: []string{"investigate", "debug", "root cause"}},
		{ID: "cmd.verify-trust", Title: "Verify trust", Description: "Run trust verification for onboarding flow", CLI: "asa verify trust onboarding", Keywords: []string{"trust", "verify", "quality"}},
		{ID: "safe.graph-rollback", Title: "Rollback graph (stub)", Description: "Destructive action requiring explicit confirmation", CLI: "asa graph rollback graph-001", Keywords: []string{"graph", "rollback", "destructive", "safety"}},
	}
}

func (m *model) openPalette() {
	m.showPalette = true
	m.paletteQuery = ""
	m.paletteCursor = 0
}

func (m *model) closePalette() {
	m.showPalette = false
	m.paletteQuery = ""
	m.paletteCursor = 0
}

func (m model) filteredPaletteEntries() []paletteEntry {
	if strings.TrimSpace(m.paletteQuery) == "" {
		return m.paletteEntries
	}
	query := strings.ToLower(strings.TrimSpace(m.paletteQuery))
	out := make([]paletteEntry, 0, len(m.paletteEntries))
	for _, entry := range m.paletteEntries {
		if paletteMatches(entry, query) {
			out = append(out, entry)
		}
	}
	return out
}

func paletteMatches(entry paletteEntry, query string) bool {
	fields := []string{
		entry.Title,
		entry.Description,
		entry.CLI,
		strings.Join(entry.Keywords, " "),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func (m *model) movePaletteCursor(delta int) {
	entries := m.filteredPaletteEntries()
	if len(entries) == 0 {
		m.paletteCursor = 0
		return
	}
	m.paletteCursor += delta
	if m.paletteCursor < 0 {
		m.paletteCursor = len(entries) - 1
	}
	if m.paletteCursor >= len(entries) {
		m.paletteCursor = 0
	}
}

func (m model) paletteSelectCmd() tea.Cmd {
	entries := m.filteredPaletteEntries()
	if len(entries) == 0 {
		return nil
	}
	idx := m.paletteCursor
	if idx < 0 || idx >= len(entries) {
		idx = 0
	}
	entry := entries[idx]
	return m.paletteActionCmd(entry.ID, entry.CLI, false)
}

func (m model) paletteActionCmd(actionID, cliEquivalent string, skipSafetyPrompt bool) tea.Cmd {
	return func() tea.Msg {
		return paletteActionMsg{
			actionID:         actionID,
			cliEquivalent:    cliEquivalent,
			skipSafetyPrompt: skipSafetyPrompt,
		}
	}
}

func (m model) handlePaletteAction(v paletteActionMsg) (tea.Model, tea.Cmd) {
	if v.actionID == "safe.graph-rollback" && m.cfg.ConfirmDestructiveActions && !v.skipSafetyPrompt {
		m.closePalette()
		m.confirmation = &safetyConfirmation{
			Title: "You are about to rollback graph-001.",
			Impact: []string{
				"2 worktrees will be impacted",
				"4 generated reports may become stale",
				"1 active session will require a manual refresh",
			},
			CLIEquivalent:  v.cliEquivalent,
			RollbackPolicy: "Rollback can be replayed manually after confirmation.",
			ActionID:       v.actionID,
		}
		return m, nil
	}
	m.closePalette()
	return m.runPaletteAction(v.actionID, v.cliEquivalent)
}

func (m model) runPaletteAction(actionID string, cliEquivalent string) (tea.Model, tea.Cmd) {
	switch actionID {
	case "nav.dashboard":
		(&m).navigateTo(ScreenDashboard, "asa dashboard")
		return m, nil
	case "nav.mission":
		(&m).navigateTo(ScreenMission, "asa mission")
		return m, nil
	case "nav.agents":
		(&m).navigateTo(ScreenAgents, "asa agents watch")
		return m, nil
	case "nav.graph":
		(&m).navigateTo(ScreenGraph, "asa graph")
		return m, nil
	case "nav.flow":
		(&m).navigateTo(ScreenFlow, "asa flow")
		return m, nil
	case "nav.logs":
		(&m).navigateTo(ScreenLogs, "asa logs")
		return m, nil
	case "nav.explain":
		(&m).navigateTo(ScreenExplain, "asa explain")
		return m, nil
	case "nav.replay":
		(&m).navigateTo(ScreenReplay, "asa replay open <replay-id>")
		return m, nil
	case "nav.prototype":
		(&m).navigateTo(ScreenPrototype, "asa prototype")
		return m, nil
	case "nav.knowledge":
		(&m).navigateTo(ScreenKnowledge, "asa knowledge")
		return m, nil
	case "nav.trust":
		(&m).navigateTo(ScreenTrust, "asa trust")
		return m, nil
	case "flow.open-onboarding":
		(&m).navigateTo(ScreenFlow, "asa flow open onboarding")
		return m, nil
	case "safe.graph-rollback":
		m.lastCommandResult = "graph rollback stub confirmed"
		return m, nil
	case "cmd.start-work":
		return m, m.dispatchCommand(bus.StartWorkCommand{Intent: "add workspace invitations"}, cliEquivalent)
	case "cmd.run-investigation":
		return m, m.dispatchCommand(bus.RunInvestigationCommand{Symptom: "onboarding fails"}, cliEquivalent)
	case "cmd.verify-trust":
		return m, m.dispatchCommand(bus.VerifyTrustCommand{Target: "onboarding"}, cliEquivalent)
	default:
		m.lastCommandResult = fmt.Sprintf("palette action not supported: %s", actionID)
		return m, nil
	}
}

func (m model) dispatchCommand(cmd bus.Command, fallbackCLI string) tea.Cmd {
	return func() tea.Msg {
		if m.commandBus == nil {
			return commandDispatchMsg{
				err: fmt.Errorf("command bus unavailable"),
			}
		}
		res, err := m.commandBus.Dispatch(m.ctx, cmd)
		if strings.TrimSpace(res.CLIEquivalent) == "" {
			res.CLIEquivalent = fallbackCLI
		}
		return commandDispatchMsg{
			result: res,
			err:    err,
		}
	}
}
