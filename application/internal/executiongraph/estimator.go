package executiongraph

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	budgetStatusOK       = "OK"
	budgetStatusExceeded = "EXCEEDED"
)

// NodeEstimate holds cost and duration for one node.
type NodeEstimate struct {
	Cost     float64
	Duration time.Duration
}

// GraphEstimate aggregates planning estimates (spec §13).
type GraphEstimate struct {
	Nodes             int
	ParallelGroups    int
	EstimatedDuration string
	EstimatedCost     float64
	HighestRisk       RiskLevel
	BudgetStatus      string
	ByNode            map[string]NodeEstimate
	ByAgent           map[string]float64
	ByNodeType        map[NodeType]float64
}

type nodeCostProfile struct {
	cost     float64
	duration time.Duration
}

var defaultNodeProfiles = map[NodeType]nodeCostProfile{
	NodeTypeInvestigation:          {cost: 0.00, duration: 30 * time.Second},
	NodeTypeEnrichment:             {cost: 0.02, duration: 2 * time.Minute},
	NodeTypeArchitectureDerivation: {cost: 0.03, duration: 3 * time.Minute},
	NodeTypeContractGeneration:     {cost: 0.04, duration: 3 * time.Minute},
	NodeTypeImplementation:         {cost: 0.08, duration: 4 * time.Minute},
	NodeTypeValidation:             {cost: 0.01, duration: 2 * time.Minute},
	NodeTypeReview:                 {cost: 0.05, duration: 3 * time.Minute},
	NodeTypeTrustVerification:      {cost: 0.02, duration: time.Minute},
	NodeTypeDocumentation:          {cost: 0.02, duration: 2 * time.Minute},
	NodeTypeReleaseCheck:           {cost: 0.01, duration: time.Minute},
	NodeTypeManualApproval:         {cost: 0.00, duration: 5 * time.Minute},
	NodeTypeRollback:               {cost: 0.01, duration: 2 * time.Minute},
}

// EstimateNode returns cost and duration for a node based on type and risk.
func EstimateNode(node GraphNode) NodeEstimate {
	profile, ok := defaultNodeProfiles[node.Type]
	if !ok {
		profile = defaultNodeProfiles[NodeTypeImplementation]
	}
	cost := profile.cost
	duration := profile.duration

	switch node.Type {
	case NodeTypeImplementation:
		switch node.Risk {
		case RiskLevelHigh:
			cost = 0.12
			duration = 6 * time.Minute
		case RiskLevelCritical:
			cost = 0.15
			duration = 8 * time.Minute
		}
	case NodeTypeReview, NodeTypeTrustVerification:
		if riskRank(node.Risk) >= riskRank(RiskLevelHigh) {
			cost = 0.06
			duration = 4 * time.Minute
		}
	}

	return NodeEstimate{Cost: cost, Duration: duration}
}

// ApplyEstimates fills estimated_cost and estimated_duration on nodes.
func ApplyEstimates(nodes []GraphNode) []GraphNode {
	out := make([]GraphNode, len(nodes))
	for i, n := range nodes {
		estimate := EstimateNode(n)
		n.EstimatedCost = estimate.Cost
		n.EstimatedDuration = formatDuration(estimate.Duration)
		out[i] = n
	}
	return out
}

// EstimateGraph aggregates node estimates using parallel groups when available.
func EstimateGraph(graph ExecutionGraph, schedule *ExecutionSchedule) GraphEstimate {
	nodeByID := make(map[string]GraphNode, len(graph.Nodes))
	estimates := make(map[string]NodeEstimate, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeByID[n.ID] = n
		estimates[n.ID] = EstimateNode(n)
	}

	totalCost := 0.0
	byAgent := make(map[string]float64)
	byType := make(map[NodeType]float64)
	for id, est := range estimates {
		totalCost += est.Cost
		node := nodeByID[id]
		byAgent[node.Agent] += est.Cost
		byType[node.Type] += est.Cost
	}

	parallelGroups := 0
	totalDuration := time.Duration(0)
	if schedule != nil && len(schedule.ParallelGroups) > 0 {
		parallelGroups = len(schedule.ParallelGroups)
		for _, group := range schedule.ParallelGroups {
			var groupMax time.Duration
			for _, id := range group {
				if est, ok := estimates[id]; ok && est.Duration > groupMax {
					groupMax = est.Duration
				}
			}
			totalDuration += groupMax
		}
	} else {
		parallelGroups = len(graph.Nodes)
		for _, est := range estimates {
			totalDuration += est.Duration
		}
	}

	budgetStatus := budgetStatusOK
	if graph.Strategy.Budget > 0 && totalCost > graph.Strategy.Budget {
		budgetStatus = budgetStatusExceeded
	}

	return GraphEstimate{
		Nodes:             len(graph.Nodes),
		ParallelGroups:    parallelGroups,
		EstimatedDuration: formatDuration(totalDuration),
		EstimatedCost:     roundCost(totalCost),
		HighestRisk:       HighestRisk(graph.Nodes),
		BudgetStatus:      budgetStatus,
		ByNode:            estimates,
		ByAgent:           sortCostMap(byAgent),
		ByNodeType:        byType,
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		secs := int(d.Round(time.Second) / time.Second)
		return fmt.Sprintf("%ds", secs)
	}
	mins := int(d.Round(time.Minute) / time.Minute)
	if mins == 1 {
		return "1m"
	}
	return fmt.Sprintf("%dm", mins)
}

func roundCost(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func sortCostMap(values map[string]float64) map[string]float64 {
	if len(values) == 0 {
		return values
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make(map[string]float64, len(values))
	for _, k := range keys {
		out[k] = roundCost(values[k])
	}
	return out
}

// FormatEstimateSummary renders a terminal-friendly estimate block (spec §13).
func FormatEstimateSummary(est GraphEstimate) string {
	var b strings.Builder
	b.WriteString("Execution Graph Estimate\n")
	b.WriteString("────────────────────────\n")
	_, _ = fmt.Fprintf(&b, "Nodes:              %d\n", est.Nodes)
	_, _ = fmt.Fprintf(&b, "Parallel groups:    %d\n", est.ParallelGroups)
	_, _ = fmt.Fprintf(&b, "Estimated duration: %s\n", est.EstimatedDuration)
	_, _ = fmt.Fprintf(&b, "Estimated cost:     €%.2f\n", est.EstimatedCost)
	_, _ = fmt.Fprintf(&b, "Highest risk:       %s\n", est.HighestRisk)
	_, _ = fmt.Fprintf(&b, "Budget status:      %s\n", est.BudgetStatus)
	return b.String()
}
