package components

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// RuntimeCard renders runtime status lines for widgets.
func RuntimeCard(st bus.RuntimeStatusResult, animated bool, frame int) string {
	status := "stopped"
	if st.Status.Running {
		status = "running"
	}
	visual := ParseVisualState(status)
	shimmer := LoadingShimmer(animated, frame, visual)
	glyph := StatusGlyph(visual, animated)
	return strings.Join([]string{
		fmt.Sprintf("%s%s Status: %s", shimmer, glyph, status),
		fmt.Sprintf("Sessions: %d", st.Status.Sessions),
		fmt.Sprintf("Flows: %d", st.Status.FlowsActive),
		fmt.Sprintf("Queue: %d", st.Status.QueuedEvents),
	}, "\n")
}

// AgentCardLine renders one agent row.
func AgentCardLine(ag bus.ActiveAgentSummary, animated bool, frame int) string {
	role := ag.Role
	if role == "" {
		role = "agent"
	}
	ref := ag.AgentRef
	if strings.TrimSpace(ref) == "" {
		ref = "-"
	}
	visual := ParseVisualState(ag.Status)
	shimmer := LoadingShimmer(animated, frame, visual)
	return fmt.Sprintf("%s%s %s %s", shimmer, StatusGlyph(visual, animated), role, ref)
}

// TrustCard renders trust dimension scores.
func TrustCard(tr bus.TrustSummaryResult) string {
	if len(tr.Dimensions) == 0 {
		return "No trust report"
	}
	lines := make([]string, 0, len(tr.Dimensions)+1)
	for _, dim := range tr.Dimensions {
		lines = append(lines, fmt.Sprintf("%-13s %s %2.0f%%", dim.Label, ProgressBar(dim.Score, 10), dim.Score*100))
	}
	lines = append(lines, fmt.Sprintf("%-13s %s %2.0f%%", "Overall", ProgressBar(tr.Overall, 10), tr.Overall*100))
	return strings.Join(lines, "\n")
}

// FlowCard renders compact flow step glyphs.
func FlowCard(flow bus.FlowGraphResult, animated bool, frame int) string {
	if len(flow.Steps) == 0 {
		return "No active flow"
	}
	labels := make([]string, 0, min(4, len(flow.Steps)))
	for i, step := range flow.Steps {
		if i >= 4 {
			break
		}
		label := step.Label
		if label == "" {
			label = "-"
		}
		visual := ParseVisualState(step.Status)
		shimmer := LoadingShimmer(animated, frame, visual)
		labels = append(labels, fmt.Sprintf("%s%s %s", shimmer, StatusGlyph(visual, animated), label))
	}
	return strings.Join(labels, "   ")
}

// RiskCard renders residual risk and high-risk graph nodes.
func RiskCard(trust bus.TrustExplorerResult, graph bus.GraphExplorerResult) string {
	var lines []string
	risk := strings.TrimSpace(trust.ResidualRisk)
	if risk == "" {
		risk = "unknown"
	}
	lines = append(lines, fmt.Sprintf("Residual: %s", risk))
	count := 0
	for _, node := range graph.Nodes {
		r := strings.ToLower(strings.TrimSpace(node.Risk))
		if r != "high" && r != "critical" {
			continue
		}
		title := node.Title
		if title == "" {
			title = node.ID
		}
		lines = append(lines, fmt.Sprintf("• %s (%s)", title, node.Risk))
		count++
		if count >= 3 {
			break
		}
	}
	if count == 0 {
		lines = append(lines, "No high-risk nodes")
	}
	return strings.Join(lines, "\n")
}

// GraphNodeCard renders one graph node summary.
func GraphNodeCard(node bus.GraphNodeSummary, animated bool) string {
	title := node.Title
	if title == "" {
		title = node.ID
	}
	return fmt.Sprintf("%s %s [%s] risk=%s",
		StatusGlyph(ParseVisualState(node.Status), animated),
		title,
		node.Type,
		emptyFallback(node.Risk, "unknown"),
	)
}

// KnowledgeCard renders top knowledge matches.
func KnowledgeCard(k bus.KnowledgeSearchResult) string {
	if len(k.Matches) == 0 {
		if strings.TrimSpace(k.Warning) != "" {
			return k.Warning
		}
		return "No knowledge hits"
	}
	lines := make([]string, 0, min(4, len(k.Matches)))
	for i, m := range k.Matches {
		if i >= 4 {
			break
		}
		name := m.Name
		if name == "" {
			name = m.ID
		}
		lines = append(lines, fmt.Sprintf("• %s (%s) %.0f%%", name, m.Type, m.Score*100))
	}
	if q := strings.TrimSpace(k.Query); q != "" {
		lines = append([]string{fmt.Sprintf("Query: %s", q)}, lines...)
	}
	return strings.Join(lines, "\n")
}

// ReplayCard renders latest replay summary.
func ReplayCard(r bus.ReplayPackageResult) string {
	if strings.TrimSpace(r.ReplayID) == "" {
		if strings.TrimSpace(r.Warning) != "" {
			return r.Warning
		}
		return "No replay package"
	}
	lines := []string{
		fmt.Sprintf("ID: %s", r.ReplayID),
		fmt.Sprintf("Mode: %s", emptyFallback(r.Mode, "-")),
		fmt.Sprintf("Events: %d", len(r.Timeline)),
	}
	for i, ev := range r.Timeline {
		if i >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf(" %s %s", ev.Time.Format("15:04:05"), ev.Type))
	}
	return strings.Join(lines, "\n")
}

// CostCard renders today/month costs.
func CostCard(today, month float64) string {
	return strings.Join([]string{
		fmt.Sprintf("Today: €%.2f", today),
		fmt.Sprintf("Month: €%.2f", month),
	}, "\n")
}

func emptyFallback(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
