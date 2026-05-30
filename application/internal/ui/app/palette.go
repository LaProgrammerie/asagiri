package app

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
)

type paletteEntry struct {
	ID          string
	ActionID    string
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
	GraphID        string
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

func fallbackPaletteEntries() []paletteEntry {
	return []paletteEntry{
		{ID: "nav.dashboard", ActionID: "nav.dashboard", Title: "Open dashboard", Description: "Navigate to dashboard live view", CLI: "asa dashboard", Keywords: []string{"screen", "dashboard", "nav"}},
		{ID: "nav.mission", ActionID: "nav.mission", Title: "Open mission control", Description: "Navigate to mission control", CLI: "asa mission", Keywords: []string{"screen", "mission", "nav"}},
		{ID: "nav.graph", ActionID: "nav.graph", Title: "Open graph explorer", Description: "Navigate to graph explorer", CLI: "asa graph", Keywords: []string{"screen", "graph", "nav"}},
		{ID: "nav.flow", ActionID: "nav.flow", Title: "Open flow explorer", Description: "Navigate to flow explorer", CLI: "asa flow", Keywords: []string{"screen", "flow", "nav"}},
		{ID: "nav.knowledge", ActionID: "nav.knowledge", Title: "Open knowledge", Description: "Navigate to knowledge explorer", CLI: "asa knowledge", Keywords: []string{"screen", "knowledge", "nav"}},
		{ID: "nav.trust", ActionID: "nav.trust", Title: "Open trust explorer", Description: "Navigate to trust explorer", CLI: "asa trust", Keywords: []string{"screen", "trust", "nav"}},
		{ID: "nav.explain", ActionID: "nav.explain", Title: "Open explain panel", Description: "Navigate to explainability panel", CLI: "asa explain", Keywords: []string{"screen", "explain", "nav"}},
		{ID: "flow.open-onboarding", ActionID: "flow.open-onboarding", Title: "Open flow onboarding", Description: "Open onboarding flow details", CLI: "asa flow open onboarding", Keywords: []string{"flow", "open", "onboarding"}},
		{ID: "cmd.verify-trust", ActionID: "cmd.verify-trust", Title: "Verify trust", Description: "Run trust verification for onboarding flow", CLI: "asa verify trust onboarding", Keywords: []string{"trust", "verify", "quality"}},
		{ID: "safe.graph-rollback", ActionID: "safe.graph-rollback", Title: "Rollback graph", Description: "Destructive action requiring explicit confirmation", CLI: "asa graph rollback <graph-id>", Keywords: []string{"graph", "rollback", "destructive", "safety"}},
	}
}

func (m *model) refreshPaletteEntries() {
	if m.queryBus == nil {
		m.paletteEntries = fallbackPaletteEntries()
		return
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetPaletteEntriesQuery{
		Screen: m.router.Current(),
		Limit:  96,
	})
	if err != nil {
		m.paletteEntries = fallbackPaletteEntries()
		return
	}
	typed, ok := res.(bus.PaletteEntriesResult)
	if !ok {
		m.paletteEntries = fallbackPaletteEntries()
		return
	}
	m.paletteEntries = toPaletteEntries(typed.Entries)
}

func toPaletteEntries(rows []bus.PaletteEntry) []paletteEntry {
	if len(rows) == 0 {
		return fallbackPaletteEntries()
	}
	out := make([]paletteEntry, 0, len(rows))
	for _, row := range rows {
		actionID := row.ActionID
		if actionID == "" {
			actionID = row.ID
		}
		out = append(out, paletteEntry{
			ID:          row.ID,
			ActionID:    actionID,
			Title:       row.Title,
			Description: row.Description,
			CLI:         row.CLI,
			Keywords:    row.Keywords,
		})
	}
	return out
}

func (m *model) openPalette() {
	m.showPalette = true
	m.paletteQuery = ""
	m.paletteCursor = 0
	m.refreshPaletteEntries()
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
	actionID := entry.ActionID
	if actionID == "" {
		actionID = entry.ID
	}
	return m.paletteActionCmd(actionID, entry.CLI, false)
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

func (m *model) graphIDForRollback() string {
	if id := strings.TrimSpace(m.snapshot.GraphExplorer.GraphID); id != "" {
		return id
	}
	return ""
}

func (m *model) loadRollbackConfirmation(cliEquivalent string) {
	graphID := m.graphIDForRollback()
	if graphID == "" {
		graphID = "graph-001"
	}
	if m.queryBus == nil {
		m.confirmation = &safetyConfirmation{
			Title:          fmt.Sprintf("You are about to rollback %s.", graphID),
			Impact:         []string{"Impact details unavailable without query bus."},
			CLIEquivalent:  cliEquivalent,
			RollbackPolicy: "Rollback can be replayed manually after confirmation.",
			ActionID:       "safe.graph-rollback",
			GraphID:        graphID,
		}
		return
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetGraphRollbackImpactQuery{GraphID: graphID})
	if err != nil {
		m.confirmation = &safetyConfirmation{
			Title:          fmt.Sprintf("You are about to rollback %s.", graphID),
			Impact:         []string{err.Error()},
			CLIEquivalent:  cliEquivalent,
			RollbackPolicy: "Rollback unavailable.",
			ActionID:       "safe.graph-rollback",
			GraphID:        graphID,
		}
		return
	}
	impact, ok := res.(bus.GraphRollbackImpactResult)
	if !ok {
		return
	}
	if strings.TrimSpace(impact.GraphID) != "" {
		graphID = impact.GraphID
	}
	if strings.TrimSpace(impact.CLIEquivalent) != "" {
		cliEquivalent = impact.CLIEquivalent
	}
	m.confirmation = &safetyConfirmation{
		Title:          impact.Title,
		Impact:         impact.ImpactLines,
		CLIEquivalent:  cliEquivalent,
		RollbackPolicy: impact.RollbackPolicy,
		ActionID:       "safe.graph-rollback",
		GraphID:        graphID,
	}
}

func (m model) handlePaletteAction(v paletteActionMsg) (tea.Model, tea.Cmd) {
	if v.actionID == "safe.graph-rollback" && m.cfg.ConfirmDestructiveActions && !v.skipSafetyPrompt {
		(&m).closePalette()
		(&m).loadRollbackConfirmation(v.cliEquivalent)
		return m, nil
	}
	(&m).closePalette()
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
		m.openExplainForFocus(bus.FocusKindDecision, m.explainSubject(), m.router.Current())
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
	case "cmd.start-work":
		return m, m.dispatchCommand(bus.StartWorkCommand{Intent: "add workspace invitations"}, cliEquivalent)
	case "cmd.run-investigation":
		return m, m.dispatchCommand(bus.RunInvestigationCommand{Symptom: "onboarding fails"}, cliEquivalent)
	case "cmd.verify-trust":
		return m, m.dispatchCommand(bus.VerifyTrustCommand{Target: "onboarding"}, cliEquivalent)
	case "cmd.build-knowledge", "ctx.knowledge-build":
		return m, m.dispatchCommand(bus.BuildKnowledgeGraphCommand{}, cliEquivalent)
	case "cmd.prototype-create":
		return m, m.dispatchCommand(bus.PrototypeCreateCommand{
			Intent:  "workspace onboarding prototype",
			Product: strings.TrimSpace(m.snapshot.Prototype.Product),
		}, cliEquivalent)
	case "cmd.flows-extract":
		product := strings.TrimSpace(m.snapshot.Prototype.Product)
		if product == "" {
			product = strings.TrimSpace(m.prototypeProduct)
		}
		return m, m.dispatchCommand(bus.FlowsExtractCommand{Product: product}, cliEquivalent)
	case "cmd.contracts-extract":
		product := strings.TrimSpace(m.snapshot.Prototype.Product)
		if product == "" {
			product = strings.TrimSpace(m.prototypeProduct)
		}
		return m, m.dispatchCommand(bus.ContractsExtractCommand{Product: product}, cliEquivalent)
	case "cmd.spec-generate":
		product := strings.TrimSpace(m.snapshot.Prototype.Product)
		if product == "" {
			product = strings.TrimSpace(m.prototypeProduct)
		}
		return m, m.dispatchCommand(bus.SpecGenerateFromProductCommand{Product: product}, cliEquivalent)
	case "cmd.prototype-pipeline", "rec.prototype-next":
		if len(m.snapshot.Prototype.SuggestedActions) == 0 {
			return m, m.dispatchPrototypePipelineAction(cliEquivalent)
		}
		return m, m.dispatchPrototypePipelineAction(m.snapshot.Prototype.SuggestedActions[0])
	case "rec.start-work":
		return m, m.dispatchCommand(bus.StartWorkCommand{Intent: "add workspace invitations"}, cliEquivalent)
	case "rec.investigate-flow":
		return m, m.dispatchCommand(bus.RunInvestigationCommand{Symptom: "onboarding fails"}, cliEquivalent)
	case "rec.verify-trust":
		return m, m.dispatchCommand(bus.VerifyTrustCommand{Target: "onboarding"}, cliEquivalent)
	case "rec.explain-security", "rec.explain-blocked":
		m.openExplainForFocus(bus.FocusKindTrustDimension, "Security", "")
		return m, nil
	case "rec.graph-resume":
		return m, m.dispatchCommand(bus.GraphResumeCommand{GraphID: m.snapshot.GraphExplorer.GraphID}, cliEquivalent)
	case "rec.export-events", "cmd.export-events":
		return m, m.dispatchCommand(bus.ExportEventsCommand{}, cliEquivalent)
	case "rec.dashboard":
		(&m).navigateTo(ScreenDashboard, "asa dashboard")
		return m, nil
	case "rec.prototype-create":
		return m, m.dispatchCommand(bus.PrototypeCreateCommand{Intent: "workspace onboarding prototype"}, cliEquivalent)
	case "cmd.complete-onboarding", "rec.complete-onboarding":
		return m, m.dispatchCommand(bus.ApplyOnboardingConfigCommand{Yes: true, Stack: "auto"}, cliEquivalent)
	case "cmd.show-readiness":
		(&m).navigateTo(ScreenOnboarding, "asa ready --plain")
		return m, nil
	case "cmd.doctor-full":
		m.lastCommandResult = cliEquivalent + " (exécutez dans un terminal)"
		return m, nil
	case "ctx.graph-resume":
		graphID := m.snapshot.GraphExplorer.GraphID
		return m, m.dispatchCommand(bus.GraphResumeCommand{GraphID: graphID}, cliEquivalent)
	case "ctx.graph-export":
		graphID := m.snapshot.GraphExplorer.GraphID
		return m, m.dispatchCommand(bus.ExportGraphCommand{GraphID: graphID, Format: "mermaid"}, cliEquivalent)
	case "safe.graph-rollback":
		graphID := m.graphIDForRollback()
		if m.confirmation != nil && strings.TrimSpace(m.confirmation.GraphID) != "" {
			graphID = m.confirmation.GraphID
		}
		if graphID == "" {
			graphID = "graph-001"
		}
		return m, m.dispatchCommand(bus.GraphRollbackCommand{GraphID: graphID}, cliEquivalent)
	default:
		if strings.HasPrefix(actionID, "flow.open.") {
			flowID := strings.TrimPrefix(actionID, "flow.open.")
			m.flowID = flowID
			(&m).navigateTo(ScreenFlow, "asa flow open "+flowID)
			return m, nil
		}
		if strings.HasPrefix(actionID, "replay.open.") {
			replayID := strings.TrimPrefix(actionID, "replay.open.")
			m.replayID = replayID
			(&m).navigateTo(ScreenReplay, "asa replay open "+replayID)
			return m, nil
		}
		if strings.HasPrefix(actionID, "replay.run.") {
			replayID := strings.TrimPrefix(actionID, "replay.run.")
			return m, m.dispatchCommand(bus.ReplayRunCommand{RunID: replayID}, cliEquivalent)
		}
		if actionID == "ctx.replay-run" && strings.TrimSpace(m.replayID) != "" {
			return m, m.dispatchCommand(bus.ReplayRunCommand{RunID: m.replayID}, cliEquivalent)
		}
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
