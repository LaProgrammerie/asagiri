package executiongraph

import (
	"fmt"
	"sort"
	"strings"
)

// FormatExplain renders dependency, parallelism, and risk explanations (spec §5.2).
func FormatExplain(graph ExecutionGraph, schedule ExecutionSchedule) string {
	var lines []string

	nodeTitle := func(id string) string {
		if n := nodeByID(graph.Nodes, id); n != nil && n.Title != "" {
			return n.Title
		}
		return id
	}

	for _, e := range graph.SortedEdges() {
		switch e.Type {
		case EdgeTypeRequires, EdgeTypeMustRunAfter, EdgeTypeBlocks:
			reason := e.Reason
			if reason == "" {
				reason = "dependency between tasks"
			}
			lines = append(lines, fmt.Sprintf("Task %s depends on %s because %s.",
				e.To, e.From, reason))
		case EdgeTypeValidates:
			lines = append(lines, fmt.Sprintf("%s is required because %s validates %s.",
				nodeTitle(e.From), nodeTitle(e.From), nodeTitle(e.To)))
		case EdgeTypeProducesContextFor:
			lines = append(lines, fmt.Sprintf("Task %s depends on %s because investigation provides context.",
				e.To, e.From))
		case EdgeTypeParallelWith:
			lines = append(lines, fmt.Sprintf("Task %s can run in parallel with %s because they touch independent paths.",
				e.From, e.To))
		}
	}

	for _, group := range schedule.ParallelGroups {
		if len(group) < 2 {
			continue
		}
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				lines = append(lines, fmt.Sprintf("Task %s can run in parallel with %s because they touch independent paths.",
					group[i], group[j]))
			}
		}
	}

	for _, n := range graph.Nodes {
		if n.Type == NodeTypeReview && n.Risk == RiskLevelHigh {
			lines = append(lines, "Security review is required because a sensitive action was detected in the flow.")
		}
		if n.Type == NodeTypeTrustVerification {
			lines = append(lines, "Trust verification is required because a sensitive action requires gate checks.")
		}
		if n.Type == NodeTypeValidation && strings.Contains(n.ID, "contracts") {
			lines = append(lines, "Contract validation is required because public API contracts are referenced.")
		}
	}

	for _, n := range graph.Nodes {
		if n.Type != NodeTypeInvestigation {
			continue
		}
		if len(graph.Flow) > 0 {
			lines = append(lines, fmt.Sprintf("Investigation node %s runs first to gather context for flow %s.",
				n.ID, graph.Flow))
		}
	}

	lines = dedupeStrings(lines)
	sort.Strings(lines)

	var b strings.Builder
	b.WriteString("Execution Plan Explanation\n")
	b.WriteString("──────────────────────────\n")
	if len(lines) == 0 {
		b.WriteString("No explicit planning decisions recorded.\n")
		return b.String()
	}
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
