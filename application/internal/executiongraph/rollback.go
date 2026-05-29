package executiongraph

import (
	"fmt"
	"sort"
	"strings"
)

// RollbackAssessment summarizes rollback coverage for a graph (spec §16).
type RollbackAssessment struct {
	GraphDefault    RollbackPlan
	NodeStrategies  map[string]RollbackStrategy
	MissingStrategy []string
	Warnings        []string
}

// PlanRollback assigns rollback strategies and detects uncovered high-risk nodes.
func PlanRollback(graph ExecutionGraph, bindings []TaskBinding) RollbackAssessment {
	bindingByNode := bindingIndex(bindings)
	nodeStrategies := make(map[string]RollbackStrategy, len(graph.Nodes))
	missing := make([]string, 0)
	warnings := make([]string, 0)

	for _, n := range graph.Nodes {
		if riskRank(n.Risk) < riskRank(RiskLevelHigh) {
			continue
		}
		var binding *TaskBinding
		if b, ok := bindingByNode[n.ID]; ok {
			binding = &b
		}
		strategy, ok := rollbackStrategyFor(n, binding)
		if !ok {
			missing = append(missing, n.ID)
			warnings = append(warnings, fmt.Sprintf("node %q has no clear rollback strategy", n.ID))
			continue
		}
		nodeStrategies[n.ID] = strategy
	}

	defaultStrategy := defaultGraphRollback(nodeStrategies, missing)
	if len(missing) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d high-risk node(s) missing rollback strategy", len(missing)))
	}

	sort.Strings(missing)

	return RollbackAssessment{
		GraphDefault: RollbackPlan{
			Strategy:        defaultStrategy,
			PreserveReports: true,
		},
		NodeStrategies:  nodeStrategies,
		MissingStrategy: missing,
		Warnings:        warnings,
	}
}

// ApplyRollbackEnrichment sets graph rollback defaults and node-level strategies.
func ApplyRollbackEnrichment(graph *ExecutionGraph, bindings []TaskBinding) RollbackAssessment {
	assessment := PlanRollback(*graph, bindings)
	graph.Rollback = &RollbackPlan{
		Strategy:        assessment.GraphDefault.Strategy,
		PreserveReports: assessment.GraphDefault.PreserveReports,
	}

	for i := range graph.Nodes {
		strategy, ok := assessment.NodeStrategies[graph.Nodes[i].ID]
		if !ok {
			continue
		}
		graph.Nodes[i].RollbackStrategy = strategy
	}
	return assessment
}

func rollbackStrategyFor(node GraphNode, binding *TaskBinding) (RollbackStrategy, bool) {
	if node.RollbackStrategy != "" {
		return node.RollbackStrategy, true
	}
	switch node.Type {
	case NodeTypeManualApproval:
		return RollbackStrategyManual, true
	case NodeTypeRollback:
		return RollbackStrategyPatchRevert, true
	}

	action := ""
	if binding != nil {
		action = strings.ToLower(binding.Action)
	}
	if strings.Contains(action, "migration") || strings.Contains(strings.ToLower(node.ID), "migration") {
		return RollbackStrategyMigrationDown, true
	}
	if binding != nil && binding.ContractRef != "" && !strings.HasPrefix(binding.ContractRef, "TODO:") {
		return RollbackStrategyPatchRevert, true
	}
	if binding != nil && strings.Contains(action, "feature") {
		return RollbackStrategyFeatureFlagDisable, true
	}
	if node.Type == NodeTypeImplementation || node.Type == NodeTypeReview || node.Type == NodeTypeTrustVerification {
		if node.Risk == RiskLevelCritical {
			return RollbackStrategyManual, true
		}
		return RollbackStrategyWorktreeReset, true
	}
	return "", false
}

func defaultGraphRollback(nodeStrategies map[string]RollbackStrategy, missing []string) RollbackStrategy {
	if len(missing) > 0 {
		return RollbackStrategyManual
	}
	priority := []RollbackStrategy{
		RollbackStrategyManual,
		RollbackStrategyMigrationDown,
		RollbackStrategyPatchRevert,
		RollbackStrategyFeatureFlagDisable,
		RollbackStrategyWorktreeReset,
	}
	for _, candidate := range priority {
		for _, strategy := range nodeStrategies {
			if strategy == candidate {
				return candidate
			}
		}
	}
	return RollbackStrategyWorktreeReset
}

// MissingRollbackNodes returns high-risk nodes without an assigned rollback strategy.
func MissingRollbackNodes(graph ExecutionGraph) []string {
	missing := make([]string, 0)
	for _, n := range graph.Nodes {
		if riskRank(n.Risk) < riskRank(RiskLevelHigh) {
			continue
		}
		if n.RollbackStrategy == "" {
			missing = append(missing, n.ID)
		}
	}
	sort.Strings(missing)
	return missing
}
