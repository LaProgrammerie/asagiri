package app

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/state"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) explorerInputActive() bool {
	if m.showHelp || m.showPalette || m.confirmation != nil {
		return false
	}
	switch m.router.Current() {
	case ScreenGraph, ScreenFlow, ScreenKnowledge, ScreenTrust, ScreenReplay, ScreenRuns:
		return true
	default:
		return false
	}
}

func (m *model) updateExplorerKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.explorerInputActive() {
		return m, nil
	}
	key := v.String()
	switch m.router.Current() {
	case ScreenGraph:
		return m.updateGraphExplorerKey(v, key)
	case ScreenFlow:
		return m.updateFlowExplorerKey(v, key)
	case ScreenKnowledge:
		return m.updateKnowledgeExplorerKey(v, key)
	case ScreenTrust:
		return m.updateTrustExplorerKey(v, key)
	case ScreenReplay:
		return m.updateReplayExplorerKey(v, key)
	case ScreenRuns:
		return m.updateRunsExplorerKey(v, key)
	default:
		return m, nil
	}
}

func (m *model) updateRunsExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.runsExplorer.Update(v)
	m.runsExplorer = next
	m.runsExplorer.Focused = true
	switch key {
	case "t":
		m.navigateTo(ScreenTrust, "asa trust")
		return m, nil
	case input.KeyExplorerGraph: // "g"
		m.navigateTo(ScreenGraph, "asa graph")
		return m, nil
	case input.KeyExplorerRun: // "r" → replay drill-down
		m.navigateTo(ScreenReplay, "asa replay open <replay-id>")
		return m, nil
	case input.KeyExplorerOpen, input.KeyExplorerEnter:
		runID := m.runsExplorer.SelectedRunID(m.snapshot.Runs)
		if runID == "" {
			return m, nil
		}
		m.lastCommandResult = "run: opened " + runID
		return m, nil
	}
	return m, nil
}

// currentRunDetail returns the aggregated detail for the selected run.
func (m model) currentRunDetail() bus.RunDetail {
	runID := m.runsExplorer.SelectedRunID(m.snapshot.Runs)
	if runID == "" {
		return bus.RunDetail{}
	}
	return m.queryRunDetail(runID)
}

func (m model) queryRunDetail(runID string) bus.RunDetail {
	if m.queryBus == nil {
		return bus.RunDetail{ID: runID}
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetRunDetailQuery{RunID: runID})
	if err != nil {
		return bus.RunDetail{ID: runID, Warning: err.Error()}
	}
	typed, ok := res.(bus.RunDetail)
	if !ok {
		return bus.RunDetail{ID: runID}
	}
	return typed
}

func (m *model) updateGraphExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.graphExplorer.Update(v)
	m.graphExplorer = next
	m.graphExplorer.Focused = true

	selected := m.graphExplorer.SelectedNode(m.refreshGraphView(), m.snapshot.GraphExplorer.Nodes)
	switch key {
	case input.KeyExplorerLogs:
		m.navigateTo(ScreenLogs, "asa logs")
		return m, nil
	case input.KeyExplorerExplain:
		if selected == nil {
			m.openExplainForFocus(bus.FocusKindGraphNode, "", m.snapshot.GraphExplorer.FlowID)
			return m, nil
		}
		m.openExplainForFocus(bus.FocusKindGraphNode, selected.ID, m.snapshot.GraphExplorer.FlowID)
		return m, nil
	case input.KeyExplorerRun:
		if selected == nil {
			return m, nil
		}
		graphID := m.snapshot.GraphExplorer.GraphID
		return m, m.dispatchCommand(bus.GraphResumeCommand{GraphID: graphID}, "asa graph resume <graph-id>")
	case input.KeyExplorerExport:
		graphID := m.snapshot.GraphExplorer.GraphID
		return m, m.dispatchCommand(bus.ExportGraphCommand{GraphID: graphID, Format: "mermaid"}, "asa graph visualize <graph-id> --format mermaid")
	case input.KeyExplorerOpen, input.KeyExplorerEnter, input.KeyExplorerDeps:
		if selected == nil {
			return m, nil
		}
		detail := m.queryGraphNodeDetail(m.snapshot.GraphExplorer.GraphID, selected.ID)
		m.graphExplorer.Detail = detail
		m.graphExplorer.ShowDetail = true
		m.state.Nav.Push(state.NavigationFrame{
			Screen:  ScreenGraph,
			Subject: selected.ID,
			Detail:  "node",
		})
		return m, nil
	}
	return m, nil
}

func (m *model) updateFlowExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.flowExplorer.Update(v)
	m.flowExplorer = next
	m.flowExplorer.Focused = true
	stepID := m.flowExplorer.SelectedStepID(m.snapshot.FlowExplorer)
	if stepID == "" {
		return m, nil
	}
	m.flowStep = m.queryFlowStepDetail(m.snapshot.FlowExplorer.FlowID, stepID)
	switch key {
	case input.KeyExplorerExplain:
		m.openExplainForFocus(bus.FocusKindFlowStep, stepID, m.snapshot.FlowExplorer.FlowID)
		return m, nil
	case input.KeyExplorerOpen, input.KeyExplorerEnter:
		m.state.Nav.Push(state.NavigationFrame{
			Screen:  ScreenFlow,
			Subject: stepID,
			Detail:  "step",
		})
	}
	return m, nil
}

func (m *model) updateKnowledgeExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.knowledgeExplorer.Update(v)
	m.knowledgeExplorer = next
	m.knowledgeExplorer.Focused = true

	match := m.knowledgeExplorer.SelectedMatch(m.snapshot.Knowledge)
	switch key {
	case input.KeyExplorerOpen, input.KeyExplorerEnter:
		if match == nil {
			return m, nil
		}
		detail := m.queryKnowledgeMatchDetail(match.ID)
		m.knowledgeExplorer.Detail = detail
		m.knowledgeExplorer.ShowDetail = true
		m.state.Nav.Push(state.NavigationFrame{Screen: ScreenKnowledge, Subject: match.ID, Detail: "match"})
		return m, nil
	case input.KeyExplorerImpact:
		if match == nil {
			return m, nil
		}
		flow, action := knowledgeImpactTargets(match)
		return m, m.dispatchCommand(bus.AnalyzeKnowledgeImpactCommand{Flow: flow, Action: action}, "asa impact analyze")
	case input.KeyExplorerContext:
		if match == nil {
			return m, nil
		}
		return m, m.dispatchCommand(bus.BuildKnowledgeContextCommand{NodeID: match.ID}, "asa knowledge query --start <node>")
	case input.KeyExplorerGraph:
		m.navigateTo(ScreenGraph, "asa graph")
		return m, nil
	case input.KeyExplorerExplain:
		if match == nil {
			return m, nil
		}
		m.openExplainForFocus(bus.FocusKindDecision, match.Name, match.ID)
		return m, nil
	}
	return m, nil
}

func (m *model) updateTrustExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.trustExplorer.Update(v)
	m.trustExplorer = next
	m.trustExplorer.Focused = true
	label := m.trustExplorer.SelectedLabel(m.snapshot.TrustExplorer)
	if label == "" {
		return m, nil
	}
	if key == input.KeyExplorerOpen || key == input.KeyExplorerEnter {
		detail := m.queryTrustDimensionDetail(label)
		m.trustExplorer.Detail = detail
		m.trustExplorer.ShowDetail = true
		m.state.Nav.Push(state.NavigationFrame{Screen: ScreenTrust, Subject: label, Detail: "dimension"})
	}
	if key == input.KeyExplorerExplain {
		m.openExplainForFocus(bus.FocusKindTrustDimension, label, "")
		return m, nil
	}
	return m, nil
}

func (m *model) updateReplayExplorerKey(v tea.KeyMsg, key string) (tea.Model, tea.Cmd) {
	next, _ := m.replayExplorer.Update(v)
	m.replayExplorer = next
	m.replayExplorer.Focused = true

	replayID := m.snapshot.Replay.ReplayID
	idx := m.replayExplorer.SelectedEventIndex(len(m.snapshot.Replay.Timeline))
	switch key {
	case input.KeyExplorerJump, input.KeyExplorerArtifact:
		detail := m.queryReplayEventDetail(replayID, idx)
		m.replayExplorer.Detail = detail
		if key == input.KeyExplorerJump {
			m.replayExplorer.ShowJump = true
		}
		if key == input.KeyExplorerArtifact {
			m.replayExplorer.ShowArtifact = true
			m.replayExplorer.ShowJump = true
		}
		m.state.Nav.Push(state.NavigationFrame{Screen: ScreenReplay, Subject: replayID, Detail: "event"})
		return m, nil
	case input.KeyExplorerCompare:
		other := m.replayExplorer.CompareReplay
		if other == "" {
			other = m.replayID
		}
		if other == "" || replayID == "" {
			return m, nil
		}
		if other == replayID {
			other = replayID + "-baseline"
		}
		cmp := m.queryReplayCompare(replayID, other)
		m.replayExplorer.Compare = cmp
		return m, m.dispatchCommand(bus.CompareReplayCommand{ReplayA: replayID, ReplayB: other}, "asa replay compare")
	case input.KeyExplorerOffline:
		if replayID == "" {
			return m, nil
		}
		return m, m.dispatchCommand(bus.ReplayRunCommand{RunID: replayID, Offline: true}, "asa replay run --offline")
	case input.KeyExplorerDivergence:
		other := m.replayExplorer.CompareReplay
		if other == "" {
			other = m.replayID
		}
		if other == "" || replayID == "" {
			return m, nil
		}
		if other == replayID {
			other = replayID + "-baseline"
		}
		cmp := m.queryReplayCompare(replayID, other)
		m.replayExplorer.Compare = cmp
		return m, m.dispatchCommand(bus.ExplainReplayDivergenceCommand{ReplayA: replayID, ReplayB: other}, "asa replay explain")
	}
	return m, nil
}

func (m model) queryGraphNodeDetail(graphID, nodeID string) *bus.GraphNodeDetail {
	if m.queryBus == nil {
		return nil
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetGraphNodeDetailQuery{GraphID: graphID, NodeID: nodeID})
	if err != nil {
		return nil
	}
	detail, ok := res.(bus.GraphNodeDetail)
	if !ok {
		return nil
	}
	return &detail
}

func (m model) queryFlowStepDetail(flowID, stepID string) bus.FlowStepDetail {
	if m.queryBus == nil {
		return bus.FlowStepDetail{ID: stepID}
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetFlowStepDetailQuery{FlowID: flowID, StepID: stepID})
	if err != nil {
		return bus.FlowStepDetail{ID: stepID}
	}
	typed, ok := res.(bus.FlowStepDetailResult)
	if !ok {
		return bus.FlowStepDetail{ID: stepID}
	}
	return typed.Step
}

func (m model) queryKnowledgeMatchDetail(matchID string) *bus.KnowledgeMatchDetail {
	if m.queryBus == nil {
		return nil
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetKnowledgeMatchDetailQuery{MatchID: matchID})
	if err != nil {
		return nil
	}
	detail, ok := res.(bus.KnowledgeMatchDetail)
	if !ok {
		return nil
	}
	return &detail
}

func (m model) queryTrustDimensionDetail(label string) *bus.TrustDimensionDetail {
	if m.queryBus == nil {
		return nil
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetTrustDimensionDetailQuery{Label: label})
	if err != nil {
		return nil
	}
	detail, ok := res.(bus.TrustDimensionDetail)
	if !ok {
		return nil
	}
	return &detail
}

func (m model) queryReplayEventDetail(replayID string, index int) *bus.ReplayEventDetail {
	if m.queryBus == nil {
		return nil
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetReplayEventDetailQuery{ReplayID: replayID, Index: index})
	if err != nil {
		return nil
	}
	detail, ok := res.(bus.ReplayEventDetail)
	if !ok {
		return nil
	}
	return &detail
}

func (m model) queryReplayCompare(replayA, replayB string) *bus.ReplayCompareResult {
	if m.queryBus == nil {
		return nil
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetReplayCompareQuery{ReplayA: replayA, ReplayB: replayB})
	if err != nil {
		return nil
	}
	cmp, ok := res.(bus.ReplayCompareResult)
	if !ok {
		return nil
	}
	return &cmp
}

func (m model) refreshGraphView() bus.GraphViewResult {
	if m.queryBus == nil {
		return bus.GraphViewResult{View: m.graphExplorer.CurrentView(), Nodes: m.snapshot.GraphExplorer.Nodes}
	}
	res, err := m.queryBus.Query(m.ctx, bus.GetGraphViewQuery{
		FlowID:  m.snapshot.GraphExplorer.FlowID,
		GraphID: m.snapshot.GraphExplorer.GraphID,
		View:    m.graphExplorer.CurrentView(),
	})
	if err != nil {
		return bus.GraphViewResult{View: m.graphExplorer.CurrentView(), Nodes: m.snapshot.GraphExplorer.Nodes}
	}
	typed, ok := res.(bus.GraphViewResult)
	if !ok {
		return bus.GraphViewResult{View: m.graphExplorer.CurrentView(), Nodes: m.snapshot.GraphExplorer.Nodes}
	}
	return typed
}

func knowledgeImpactTargets(match *bus.KnowledgeMatch) (flow, action string) {
	if match == nil {
		return "", ""
	}
	parts := strings.SplitN(match.ID, ":", 2)
	if len(parts) == 2 {
		return parts[1], parts[1]
	}
	return match.Name, match.Name
}
