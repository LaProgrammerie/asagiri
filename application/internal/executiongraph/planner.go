package executiongraph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const (
	defaultMaxParallel = 2
	defaultBudget      = 0.50
)

// Planner builds execution graphs from product inputs (spec §10–11 entry point).
type Planner struct {
	RepoRoot string
	Inferer  DependencyInferer
	Now      func() time.Time
}

// NewPlanner returns a planner wired with default dependency inference.
func NewPlanner(repoRoot string) *Planner {
	return &Planner{
		RepoRoot: repoRoot,
		Inferer:  DefaultDependencyInferer{},
		Now:      time.Now,
	}
}

func (p *Planner) Build(ctx context.Context, req GraphPlanRequest) (ExecutionGraph, error) {
	if req.Product == "" {
		return ExecutionGraph{}, fmt.Errorf("planner: product required")
	}
	if p.Inferer == nil {
		p.Inferer = DefaultDependencyInferer{}
	}
	nowFn := p.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	flow, flowID, err := p.loadFlow(req)
	if err != nil {
		return ExecutionGraph{}, err
	}
	tasks, err := loadProductTasks(p.RepoRoot, req.Product)
	if err != nil {
		return ExecutionGraph{}, err
	}
	if len(tasks) == 0 {
		tasks = tasksFromFlow(req.Product, flowID, flow)
	}

	bindings := buildTaskBindings(tasks, flow)
	nodes := buildBaseNodes(req, flowID, bindings)
	nodes = appendStubNodes(nodes, flow, bindings)

	edges, err := p.Inferer.Infer(ctx, DependencyInput{
		Product:      req.Product,
		Flow:         flowID,
		RepoRoot:     p.RepoRoot,
		Nodes:        nodes,
		TaskBindings: bindings,
	})
	if err != nil {
		return ExecutionGraph{}, err
	}
	edges = append(edges, baseContextEdges(nodes)...)
	if extra, err := AppendKnowledgeDependencyEdges(ctx, p.RepoRoot, flowID, nodes, edges); err != nil {
		return ExecutionGraph{}, err
	} else if len(extra) > 0 {
		edges = append(edges, extra...)
	}

	if err := DetectCycles(nodes, edges); err != nil {
		return ExecutionGraph{}, err
	}

	graph := ExecutionGraph{
		ID:        NewGraphID(),
		Product:   req.Product,
		Flow:      flowID,
		Status:    GraphStatusPlanned,
		CreatedAt: nowFn().UTC().Format(time.RFC3339),
		Strategy: Strategy{
			MaxParallel: defaultMaxParallel,
			StopOnRisk:  RiskLevelHigh,
			StrictTrust: true,
			Budget:      defaultBudget,
		},
		Nodes: nodes,
		Edges: dedupeEdges(edges),
	}
	if err := graph.Validate(); err != nil {
		return ExecutionGraph{}, err
	}
	trustInput := TrustEnrichmentInput{
		Gates:                    req.Gates,
		TrustRequiredForHighRisk: req.Gates.TrustRequiredForHighRisk,
	}
	return p.EnrichGraph(ctx, graph, bindings, flow, trustInput)
}

// EnrichGraph applies agent assignment, risk, trust, investigation, estimates, checkpoints, and rollback (spec §12–18).
func (p *Planner) EnrichGraph(ctx context.Context, graph ExecutionGraph, bindings []TaskBinding, flow product.Flow, trustInput TrustEnrichmentInput) (ExecutionGraph, error) {
	trustInput.Flow = flow
	graph.Nodes = AssignAgents(graph.Nodes, bindings)
	var extraEdges []GraphEdge
	graph.Nodes, extraEdges = ApplyRiskEnrichment(graph.Nodes, bindings, graph.Edges, trustInput.Gates.HumanApprovalFor)
	if len(extraEdges) > 0 {
		graph.Edges = dedupeEdges(append(graph.Edges, extraEdges...))
	}
	graph.Nodes, extraEdges = ApplyTrustEnrichment(graph.Nodes, bindings, graph.Edges, trustInput)
	if len(extraEdges) > 0 {
		graph.Edges = dedupeEdges(append(graph.Edges, extraEdges...))
	}
	graph.Nodes, extraEdges = ApplyInvestigationEnrichment(graph.Nodes, bindings, graph.Edges, flow)
	if len(extraEdges) > 0 {
		graph.Edges = dedupeEdges(append(graph.Edges, extraEdges...))
	}
	graph.Nodes = AssignAgents(graph.Nodes, bindings)
	graph.Nodes = ApplyEstimates(graph.Nodes)
	graph.Checkpoints = GenerateCheckpoints(graph)
	ApplyRollbackEnrichment(&graph, bindings)
	if err := ApplyKnowledgeGraphEnrichment(ctx, p.RepoRoot, &graph); err != nil {
		return ExecutionGraph{}, err
	}
	if err := graph.Validate(); err != nil {
		return ExecutionGraph{}, err
	}
	return graph, nil
}

func (p *Planner) loadFlow(req GraphPlanRequest) (product.Flow, string, error) {
	flowID := strings.TrimSpace(req.Flow)
	if flowID == "" {
		return product.Flow{}, "", fmt.Errorf("planner: flow required")
	}
	path := resolveFlowPath(productDir(p.RepoRoot, req.Product), flowID)
	raw, err := os.ReadFile(path)
	if err != nil {
		return product.Flow{}, "", fmt.Errorf("planner: read flow %q: %w", flowID, err)
	}
	flow, err := product.ParseFlowYAML(raw)
	if err != nil {
		return product.Flow{}, "", fmt.Errorf("planner: parse flow %q: %w", flowID, err)
	}
	if flow.ID != "" {
		flowID = flow.ID
	}
	return flow, flowID, nil
}

func loadProductTasks(repoRoot, productName string) ([]asagiri.Task, error) {
	dir := filepath.Join(repoRoot, ".asagiri", "tasks", product.Slug(productName))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}
	tasks := make([]asagiri.Task, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var task asagiri.Task
		if err := yaml.Unmarshal(raw, &task); err != nil {
			continue
		}
		if task.ID == "" {
			task.ID = strings.TrimSuffix(e.Name(), ".yaml")
		}
		tasks = append(tasks, task)
	}
	sortTasksByFlowStep(tasks)
	return tasks, nil
}

func tasksFromFlow(productID, flowID string, flow product.Flow) []asagiri.Task {
	tasks := make([]asagiri.Task, 0, len(flow.Steps))
	for i, step := range flow.Steps {
		taskID := fmt.Sprintf("%s-%03d", product.Slug(productID), i+1)
		tasks = append(tasks, asagiri.Task{
			ID:      taskID,
			Title:   fmt.Sprintf("Implement %s/%s", flowID, step.Action),
			Feature: productID,
			Status:  asagiri.StatusPending,
			Type:    "implementation",
			Source: asagiri.TaskSource{
				Product: productID,
				Flow:    flowID,
				Step:    step.ID,
				Action:  step.Action,
			},
			Scope: asagiri.TaskScope{
				AllowedPaths: []string{"application/**"},
			},
		})
	}
	return tasks
}

func buildTaskBindings(tasks []asagiri.Task, flow product.Flow) []TaskBinding {
	stepByID := make(map[string]product.FlowStep, len(flow.Steps))
	stepIndex := make(map[string]int, len(flow.Steps))
	for i, step := range flow.Steps {
		stepByID[step.ID] = step
		stepIndex[step.ID] = i
	}

	bindings := make([]TaskBinding, 0, len(tasks))
	for _, task := range tasks {
		step, ok := stepByID[task.Source.Step]
		idx := -1
		if ok {
			idx = stepIndex[task.Source.Step]
		} else if task.Source.Action != "" {
			for i, s := range flow.Steps {
				if s.Action == task.Source.Action {
					step = s
					idx = i
					ok = true
					break
				}
			}
		}
		action := task.Source.Action
		contractRef := ""
		sensitive := false
		if ok {
			action = step.Action
			contractRef = step.ContractRef
			sensitive = step.Sensitive
		}
		bindings = append(bindings, TaskBinding{
			NodeID:      implementationNodeID(action),
			TaskID:      task.ID,
			FlowStepID:  task.Source.Step,
			StepIndex:   idx,
			Action:      action,
			ContractRef: contractRef,
			Sensitive:   sensitive,
			ScopePaths:  append([]string(nil), task.Scope.AllowedPaths...),
		})
	}
	return bindings
}

func buildBaseNodes(req GraphPlanRequest, flowID string, bindings []TaskBinding) []GraphNode {
	flowSlug := flowSlug(flowID)
	nodes := []GraphNode{
		{
			ID:    "investigate-" + flowSlug,
			Type:  NodeTypeInvestigation,
			Title: "Investigate " + flowSlug + " flow",
			Agent: "local",
			Risk:  RiskLevelLow,
		},
	}
	for _, b := range bindings {
		if b.NodeID == "" {
			continue
		}
		nodes = append(nodes, GraphNode{
			ID:    b.NodeID,
			Type:  NodeTypeImplementation,
			Title: fmt.Sprintf("Implement %s", strings.ReplaceAll(b.Action, "_", " ")),
			Task:  b.TaskID,
			Agent: "cursor",
			Risk:  riskForBinding(b),
			RequiredChecks: []string{
				"tests",
			},
		})
	}
	if req.IncludeReviews {
		nodes = append(nodes, GraphNode{
			ID:    "verify-" + flowSlug + "-flow",
			Type:  NodeTypeValidation,
			Title: "Verify " + flowSlug + " flow integrity",
			Agent: "local",
			Risk:  RiskLevelMedium,
			RequiredChecks: []string{
				"flows",
			},
		})
	}
	return nodes
}

func appendStubNodes(nodes []GraphNode, flow product.Flow, bindings []TaskBinding) []GraphNode {
	has := nodeIDSet(nodes)
	add := func(n GraphNode) {
		if _, ok := has[n.ID]; ok {
			return
		}
		has[n.ID] = struct{}{}
		nodes = append(nodes, n)
	}

	needsContracts := false
	hasSensitive := false
	hasPublicAPI := false
	for _, b := range bindings {
		ref := strings.TrimSpace(b.ContractRef)
		if strings.HasPrefix(ref, "TODO:") {
			needsContracts = true
		}
		if b.Sensitive {
			hasSensitive = true
		}
		if ref != "" && !strings.HasPrefix(ref, "TODO:") {
			hasPublicAPI = true
		}
	}
	for _, action := range flow.Security.SensitiveActions {
		for _, b := range bindings {
			if b.Action == action {
				hasSensitive = true
			}
		}
	}

	if needsContracts {
		add(GraphNode{
			ID:    "derive-contracts",
			Type:  NodeTypeContractGeneration,
			Title: "Generate flow API contracts",
			Agent: "local",
			Risk:  RiskLevelMedium,
		})
	}
	if hasPublicAPI {
		add(GraphNode{
			ID:    "verify-contracts",
			Type:  NodeTypeValidation,
			Title: "Validate public contract compatibility",
			Agent: "local",
			Risk:  RiskLevelMedium,
			RequiredChecks: []string{
				"backward_compatibility",
			},
		})
	}
	if hasSensitive {
		add(GraphNode{
			ID:    "security-review",
			Type:  NodeTypeReview,
			Title: "Security review for sensitive actions",
			Agent: "codex",
			Risk:  RiskLevelHigh,
			RequiredChecks: []string{
				"security",
			},
		})
		add(GraphNode{
			ID:    "trust-gate",
			Type:  NodeTypeTrustVerification,
			Title: "Trust verification for sensitive action",
			Agent: "local",
			Risk:  RiskLevelHigh,
			RequiredChecks: []string{
				"trust",
				"flows",
			},
		})
	}
	return nodes
}

func baseContextEdges(nodes []GraphNode) []GraphEdge {
	has := nodeIDSet(nodes)
	var investigateID string
	var verifyID string
	for _, n := range nodes {
		switch n.Type {
		case NodeTypeInvestigation:
			investigateID = n.ID
		case NodeTypeValidation:
			if strings.HasPrefix(n.ID, "verify-") && !strings.Contains(n.ID, "contracts") {
				verifyID = n.ID
			}
		}
	}
	if investigateID == "" {
		return nil
	}
	edges := make([]GraphEdge, 0)
	for _, n := range nodes {
		if n.Type != NodeTypeImplementation {
			continue
		}
		edges = append(edges, GraphEdge{
			From:   investigateID,
			To:     n.ID,
			Type:   EdgeTypeProducesContextFor,
			Reason: "investigation provides context for implementation",
		})
		if verifyID != "" {
			edges = append(edges, GraphEdge{
				From:   n.ID,
				To:     verifyID,
				Type:   EdgeTypeMustRunAfter,
				Reason: "validation runs after implementation",
			})
		}
	}
	if verifyID != "" {
		if _, ok := has[verifyID]; ok {
			_ = ok
		}
	}
	return edges
}

func implementationNodeID(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return ""
	}
	return "implement-" + strings.ReplaceAll(strings.ToLower(action), "_", "-")
}

func flowSlug(flowID string) string {
	flowID = strings.TrimSpace(flowID)
	if idx := strings.LastIndex(flowID, "-"); idx >= 0 && idx < len(flowID)-1 {
		return flowID[idx+1:]
	}
	return strings.ReplaceAll(flowID, "_", "-")
}

func riskForBinding(b TaskBinding) RiskLevel {
	if b.Sensitive {
		return RiskLevelHigh
	}
	return RiskLevelMedium
}

func sortTasksByFlowStep(tasks []asagiri.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Source.Step == tasks[j].Source.Step {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].Source.Step < tasks[j].Source.Step
	})
}
