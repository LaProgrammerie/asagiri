package executiongraph

import (
	"fmt"
	"strings"
)

// FormatTerminalSummary renders the §22 terminal UX for a planned graph.
func FormatTerminalSummary(graph ExecutionGraph, schedule ExecutionSchedule, est GraphEstimate) string {
	var b strings.Builder
	b.WriteString("Asagiri Execution Graph\n")
	b.WriteString("═══════════════════════\n")
	fmt.Fprintf(&b, "Product: %s\n", graph.Product)
	if graph.Flow != "" {
		fmt.Fprintf(&b, "Flow:    %s\n", graph.Flow)
	}
	fmt.Fprintf(&b, "Graph ID: %s\n\n", graph.ID)

	b.WriteString("Graph\n")
	b.WriteString("─────\n")
	fmt.Fprintf(&b, "Nodes:             %d\n", est.Nodes)
	fmt.Fprintf(&b, "Dependencies:      %d\n", len(graph.Edges))
	fmt.Fprintf(&b, "Parallel groups:   %d\n", est.ParallelGroups)
	fmt.Fprintf(&b, "Checkpoints:       %d\n", len(graph.Checkpoints))
	fmt.Fprintf(&b, "Highest risk:      %s\n", est.HighestRisk)
	fmt.Fprintf(&b, "Estimated cost:    €%.2f\n", est.EstimatedCost)
	fmt.Fprintf(&b, "Estimated duration: %s\n", est.EstimatedDuration)

	for i, group := range schedule.ParallelGroups {
		b.WriteString("\n")
		fmt.Fprintf(&b, "Parallel group %d\n", i+1)
		b.WriteString("────────────────\n")
		for _, id := range group {
			node := nodeByID(graph.Nodes, id)
			label := id
			agent := ""
			cost := 0.0
			if node != nil {
				if node.Title != "" {
					label = node.Title
				}
				agent = node.Agent
				cost = node.EstimatedCost
			}
			marker := "○"
			if i == 0 {
				marker = "✓"
			}
			fmt.Fprintf(&b, "%s %-30s %-12s %5.2f€\n", marker, truncateLabel(label, 30), agent, cost)
		}
	}

	if len(schedule.Blocked) > 0 {
		b.WriteString("\nBlocked\n")
		b.WriteString("───────\n")
		for _, blocked := range schedule.Blocked {
			waitFor := strings.Join(blocked.WaitFor, ", ")
			fmt.Fprintf(&b, "%s waits for %s\n", blocked.NodeID, waitFor)
		}
	}

	return b.String()
}

func truncateLabel(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
