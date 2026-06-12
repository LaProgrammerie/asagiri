package doctor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

const (
	findingTaskWithoutGraph       = "task_without_graph_node"
	findingGraphNodeNeverExecuted = "graph_node_never_executed"
	findingTaskWithoutKnowledge   = "task_without_knowledge_context"
	findingAgentRunWithoutTask    = "agent_run_without_task"
	findingTrustGapCriticalFlow   = "trust_gap_critical_flow"
)

type graphIndex struct {
	graphs        []executiongraph.ExecutionGraph
	tasksInGraph  map[string]struct{}
	neverExecuted []archNodeRef
	criticalFlows map[string]criticalFlowRef
}

type archNodeRef struct {
	GraphID string
	NodeID  string
	Flow    string
	Product string
}

type criticalFlowRef struct {
	Flow    string
	Product string
}

type trustReportIndex struct {
	byFlow map[string]trustReportSnap
	count  int
}

type trustReportSnap struct {
	TrustID      string
	Flow         string
	ResidualRisk trust.ResidualRisk
	GateStatus   trust.GateStatus
	GeneratedAt  string
}

// BuildArchitecture synthesizes a read-only cross-artefact report from startDir (typically cwd).
func BuildArchitecture(startDir string) (ArchitectureReport, error) {
	report := ArchitectureReport{ReportVersion: ArchitectureReportVersion}

	repoRoot, err := bootstrap.GitRoot(startDir)
	if err != nil {
		return report, err
	}
	report.Repository.GitRoot = repoRoot

	var cfg *config.Config
	if cfgPath := config.ConfigPath(repoRoot); cfgPath != "" {
		if loaded, loadErr := config.Load(cfgPath, repoRoot); loadErr == nil {
			cfg = loaded
		}
	}

	tasks, sqlitePresent := collectAllTasks(repoRoot, cfg)
	report.Sources.SQLitePresent = sqlitePresent
	report.Summary.Tasks = len(tasks)

	gidx, graphsPresent := collectGraphIndex(repoRoot)
	report.Sources.GraphsPresent = graphsPresent
	report.Summary.ExecutionGraphs = len(gidx.graphs)
	for _, g := range gidx.graphs {
		report.Summary.ExecutionGraphNodes += len(g.Nodes)
	}

	kg, knowledgeStore, knowledgeJSON := loadKnowledgeGraph(repoRoot)
	report.Sources.KnowledgeStore = knowledgeStore
	report.Sources.KnowledgeJSON = knowledgeJSON
	report.Summary.KnowledgeNodes = len(kg.Nodes)

	trustIdx := collectTrustReports(repoRoot)
	report.Sources.TrustReportsPresent = trustIdx.count > 0
	report.Summary.TrustReports = trustIdx.count

	ledger, ledgerPresent := collectAgentLedger(repoRoot)
	report.Sources.AgentLedgerPresent = ledgerPresent
	report.Summary.AgentLedgerEntries = ledger.Count

	var findings []ArchitectureFinding

	for _, task := range tasks {
		if _, ok := gidx.tasksInGraph[task.ID]; !ok {
			findings = append(findings, ArchitectureFinding{
				Kind:     findingTaskWithoutGraph,
				Severity: StatusWarn,
				TaskID:   task.ID,
				Feature:  task.Feature,
				Message:  "task sans nœud dans un execution graph",
			})
		}
		if !taskHasKnowledgeContext(task.ID, kg) {
			findings = append(findings, ArchitectureFinding{
				Kind:     findingTaskWithoutKnowledge,
				Severity: StatusWarn,
				TaskID:   task.ID,
				Feature:  task.Feature,
				Message:  "task sans contexte dans le knowledge graph",
			})
		}
	}

	for _, ref := range gidx.neverExecuted {
		findings = append(findings, ArchitectureFinding{
			Kind:     findingGraphNodeNeverExecuted,
			Severity: StatusWarn,
			GraphID:  ref.GraphID,
			NodeID:   ref.NodeID,
			Flow:     ref.Flow,
			Product:  ref.Product,
			Message:  "nœud d'execution graph jamais exécuté",
		})
	}

	for _, entry := range ledger.Entries {
		if strings.TrimSpace(entry.TaskID) != "" {
			continue
		}
		findings = append(findings, ArchitectureFinding{
			Kind:     findingAgentRunWithoutTask,
			Severity: StatusWarn,
			RunID:    entry.RunID,
			AgentID:  entry.AgentID,
			Feature:  entry.Feature,
			Message:  "entrée agent ledger sans task_id",
		})
	}

	for _, cref := range sortedCriticalFlows(gidx.criticalFlows) {
		if gap, msg := trustGapForFlow(trustIdx, cref.Flow); gap {
			findings = append(findings, ArchitectureFinding{
				Kind:     findingTrustGapCriticalFlow,
				Severity: StatusWarn,
				Flow:     cref.Flow,
				Product:  cref.Product,
				Message:  msg,
			})
		}
	}

	sortArchitectureFindings(findings)
	report.Findings = findings
	fillArchitectureSummaryCounts(&report.Summary, findings)
	report.Recommendations = buildArchitectureRecommendations(report.Summary, gidx, findings)
	return report, nil
}

func collectAllTasks(repoRoot string, cfg *config.Config) ([]sqlite.Task, bool) {
	if cfg == nil {
		return nil, false
	}
	dbPath := cfg.StateDBPath(repoRoot)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, false
	}
	store, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, true
	}
	defer func() { _ = store.Close() }()

	runs, err := store.ListRuns(5000)
	if err != nil {
		return nil, true
	}
	seen := map[string]struct{}{}
	var out []sqlite.Task
	for _, run := range runs {
		tasks, err := store.ListTasksByRun(run.ID)
		if err != nil {
			continue
		}
		for _, t := range tasks {
			if _, dup := seen[t.ID]; dup {
				continue
			}
			seen[t.ID] = struct{}{}
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, true
}

func collectGraphIndex(repoRoot string) (graphIndex, bool) {
	idx := graphIndex{
		tasksInGraph:  map[string]struct{}{},
		criticalFlows: map[string]criticalFlowRef{},
	}
	graphsRoot := filepath.Join(repoRoot, ".asagiri", "graphs")
	entries, err := os.ReadDir(graphsRoot)
	if err != nil {
		return idx, false
	}
	repo := executiongraph.NewRepository(repoRoot)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		graph, err := repo.Load(entry.Name())
		if err != nil {
			continue
		}
		idx.graphs = append(idx.graphs, graph)
		events := executedNodesFromEvents(filepath.Join(graphsRoot, entry.Name(), "events.jsonl"))
		for _, n := range graph.Nodes {
			if tid := strings.TrimSpace(n.Task); tid != "" {
				idx.tasksInGraph[tid] = struct{}{}
			}
			if !nodeWasExecuted(n, events) {
				idx.neverExecuted = append(idx.neverExecuted, archNodeRef{
					GraphID: graph.ID,
					NodeID:  n.ID,
					Flow:    graph.Flow,
					Product: graph.Product,
				})
			}
		}
		if flow := strings.TrimSpace(graph.Flow); flow != "" && graphIsCritical(graph) {
			idx.criticalFlows[strings.ToLower(flow)] = criticalFlowRef{Flow: flow, Product: graph.Product}
		}
	}
	sort.Slice(idx.graphs, func(i, j int) bool { return idx.graphs[i].ID < idx.graphs[j].ID })
	sort.Slice(idx.neverExecuted, func(i, j int) bool {
		if idx.neverExecuted[i].GraphID != idx.neverExecuted[j].GraphID {
			return idx.neverExecuted[i].GraphID < idx.neverExecuted[j].GraphID
		}
		return idx.neverExecuted[i].NodeID < idx.neverExecuted[j].NodeID
	})
	return idx, len(idx.graphs) > 0
}

func executedNodesFromEvents(path string) map[string]struct{} {
	out := map[string]struct{}{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		nodeID := firstEventString(row, "node_id", "node")
		if nodeID == "" {
			continue
		}
		eventType := firstEventString(row, "type", "event")
		if isExecutionEvent(eventType) {
			out[nodeID] = struct{}{}
		}
	}
	return out
}

func isExecutionEvent(eventType string) bool {
	switch strings.ToLower(strings.TrimSpace(eventType)) {
	case "graph.node.started", "graph.node.completed", "graph.node.failed",
		"graph.node.succeeded", "graph.node.skipped", "agent.completed", "agent.failed":
		return true
	default:
		return strings.Contains(eventType, "node.") || strings.HasPrefix(eventType, "agent.")
	}
}

func firstEventString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		raw, ok := row[key]
		if !ok || raw == nil {
			continue
		}
		if s, ok := raw.(string); ok {
			if v := strings.TrimSpace(s); v != "" {
				return v
			}
		}
	}
	return ""
}

func nodeWasExecuted(n executiongraph.GraphNode, events map[string]struct{}) bool {
	switch n.Status {
	case executiongraph.NodeStatusSucceeded, executiongraph.NodeStatusFailed,
		executiongraph.NodeStatusSkipped, executiongraph.NodeStatusRolledBack:
		return true
	case executiongraph.NodeStatusRunning:
		return true
	}
	if _, ok := events[n.ID]; ok {
		return true
	}
	return false
}

func graphIsCritical(graph executiongraph.ExecutionGraph) bool {
	switch graph.Status {
	case executiongraph.GraphStatusFailed, executiongraph.GraphStatusBlocked, executiongraph.GraphStatusRunning:
		return true
	}
	for _, n := range graph.Nodes {
		if n.Risk == executiongraph.RiskLevelHigh || n.Risk == executiongraph.RiskLevelCritical {
			return true
		}
		if n.Type == executiongraph.NodeTypeTrustVerification {
			return true
		}
	}
	return false
}

func loadKnowledgeGraph(repoRoot string) (knowledge.KnowledgeGraph, bool, bool) {
	if store, err := knowledge.OpenStoreIfExists(repoRoot); err == nil {
		defer func() { _ = store.Close() }()
		if g, err := store.LoadGraph(context.Background()); err == nil {
			return g, true, fileExists(knowledge.JSONPath(repoRoot))
		}
	}
	if body, err := os.ReadFile(knowledge.JSONPath(repoRoot)); err == nil {
		if g, err := knowledge.ParseJSON(body); err == nil {
			return g, false, true
		}
	}
	return knowledge.KnowledgeGraph{}, false, false
}

func taskHasKnowledgeContext(taskID string, kg knowledge.KnowledgeGraph) bool {
	if len(kg.Nodes) == 0 {
		return false
	}
	taskID = strings.TrimSpace(taskID)
	for _, n := range kg.Nodes {
		if n.Name == taskID {
			return true
		}
		if strings.Contains(n.Path, taskID) {
			return true
		}
		if strings.Contains(n.ID, taskID) {
			return true
		}
		if art, _ := n.Properties["artefact"].(string); art == "task" && strings.Contains(n.Name, taskID) {
			return true
		}
	}
	return false
}

func collectTrustReports(repoRoot string) trustReportIndex {
	idx := trustReportIndex{byFlow: map[string]trustReportSnap{}}
	trustRoot := filepath.Join(repoRoot, ".asagiri", "trust")
	entries, err := os.ReadDir(trustRoot)
	if err != nil {
		return idx
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(trustRoot, entry.Name(), "report.json")
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var snap trustReportSnap
		var raw struct {
			TrustID      string               `json:"trust_id"`
			Flow         string               `json:"flow"`
			GeneratedAt  string               `json:"generated_at"`
			ResidualRisk trust.ResidualRisk   `json:"residual_risk"`
			Gate         trust.GateEvaluation `json:"gate"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			continue
		}
		snap = trustReportSnap{
			TrustID:      raw.TrustID,
			Flow:         strings.TrimSpace(raw.Flow),
			ResidualRisk: raw.ResidualRisk,
			GateStatus:   raw.Gate.Status,
			GeneratedAt:  raw.GeneratedAt,
		}
		idx.count++
		if snap.Flow == "" {
			continue
		}
		key := strings.ToLower(snap.Flow)
		prev, ok := idx.byFlow[key]
		if !ok || snap.GeneratedAt > prev.GeneratedAt {
			idx.byFlow[key] = snap
		}
	}
	return idx
}

func trustGapForFlow(idx trustReportIndex, flow string) (bool, string) {
	flow = strings.TrimSpace(flow)
	if flow == "" {
		return false, ""
	}
	snap, ok := idx.byFlow[strings.ToLower(flow)]
	if !ok {
		return true, "flow critique sans trust report produit"
	}
	if snap.GateStatus == trust.GateStatusBlocked {
		return true, "trust report avec gate bloqué"
	}
	if snap.ResidualRisk == trust.ResidualRiskHigh {
		return true, "trust report avec risque résiduel élevé"
	}
	return false, ""
}

func collectAgentLedger(repoRoot string) (agentledger.Report, bool) {
	report, err := agentledger.List(repoRoot, agentledger.ListOptions{})
	if err != nil {
		return agentledger.Report{Entries: []agentledger.Entry{}}, false
	}
	if _, err := os.Stat(agentledger.Path(repoRoot)); err != nil {
		return report, false
	}
	return report, true
}

func sortedCriticalFlows(m map[string]criticalFlowRef) []criticalFlowRef {
	out := make([]criticalFlowRef, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Flow < out[j].Flow })
	return out
}

func sortArchitectureFindings(findings []ArchitectureFinding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Kind != findings[j].Kind {
			return findings[i].Kind < findings[j].Kind
		}
		if findings[i].TaskID != findings[j].TaskID {
			return findings[i].TaskID < findings[j].TaskID
		}
		if findings[i].GraphID != findings[j].GraphID {
			return findings[i].GraphID < findings[j].GraphID
		}
		if findings[i].NodeID != findings[j].NodeID {
			return findings[i].NodeID < findings[j].NodeID
		}
		return findings[i].Flow < findings[j].Flow
	})
}

func fillArchitectureSummaryCounts(summary *ArchitectureSummary, findings []ArchitectureFinding) {
	for _, f := range findings {
		switch f.Kind {
		case findingTaskWithoutGraph:
			summary.TasksWithoutGraphNode++
		case findingGraphNodeNeverExecuted:
			summary.GraphNodesNeverExecuted++
		case findingTaskWithoutKnowledge:
			summary.TasksWithoutKnowledge++
		case findingAgentRunWithoutTask:
			summary.AgentRunsWithoutTask++
		case findingTrustGapCriticalFlow:
			summary.TrustGapsCriticalFlows++
		}
	}
}

func buildArchitectureRecommendations(summary ArchitectureSummary, gidx graphIndex, findings []ArchitectureFinding) []Action {
	var actions []Action
	if summary.TasksWithoutGraphNode > 0 || summary.GraphNodesNeverExecuted > 0 {
		cli := planGraphCLI(gidx)
		actions = appendUniqueAction(actions, Action{
			Title: "Planifier ou rafraîchir les execution graphs",
			CLI:   cli,
		})
	}
	if summary.TasksWithoutKnowledge > 0 || summary.KnowledgeNodes == 0 {
		if summary.KnowledgeNodes == 0 {
			actions = appendUniqueAction(actions, Action{
				Title: "Construire le knowledge graph",
				CLI:   "asa knowledge build",
			})
		} else {
			actions = appendUniqueAction(actions, Action{
				Title: "Rafraîchir le knowledge graph",
				CLI:   "asa knowledge build --incremental",
			})
		}
	}
	if summary.AgentRunsWithoutTask > 0 || summary.AgentLedgerEntries > 0 {
		actions = appendUniqueAction(actions, Action{
			Title: "Inspecter les exécutions agents",
			CLI:   "asa agents runs --json",
		})
	}
	if summary.TrustGapsCriticalFlows > 0 {
		cli := verifyTrustCLI(gidx, findings)
		actions = appendUniqueAction(actions, Action{
			Title: "Vérifier la confiance produit sur les flows critiques",
			CLI:   cli,
		})
	}
	return actions
}

func planGraphCLI(gidx graphIndex) string {
	for _, g := range gidx.graphs {
		product := strings.TrimSpace(g.Product)
		flow := strings.TrimSpace(g.Flow)
		if product != "" && flow != "" {
			return fmt.Sprintf("asa plan graph %s --flow %s", product, flow)
		}
	}
	return "asa plan graph <product> --flow <flow>"
}

func verifyTrustCLI(gidx graphIndex, findings []ArchitectureFinding) string {
	for _, f := range findings {
		if f.Kind != findingTrustGapCriticalFlow {
			continue
		}
		flow := strings.TrimSpace(f.Flow)
		product := strings.TrimSpace(f.Product)
		if flow == "" {
			continue
		}
		if product != "" {
			return fmt.Sprintf("asa verify trust %s --product %s", flow, product)
		}
		return fmt.Sprintf("asa verify trust %s", flow)
	}
	if flows := sortedCriticalFlows(gidx.criticalFlows); len(flows) > 0 {
		cref := flows[0]
		if cref.Product != "" {
			return fmt.Sprintf("asa verify trust %s --product %s", cref.Flow, cref.Product)
		}
		return fmt.Sprintf("asa verify trust %s", cref.Flow)
	}
	return "asa verify trust <flow>"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
