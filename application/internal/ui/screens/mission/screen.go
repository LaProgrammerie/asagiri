package mission

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// ViewModel is the data contract for Mission Control.
type ViewModel struct {
	Workspace         string
	Branch            string
	SessionStatus     string
	RuntimeStatus     string
	ActiveAgents      []bus.ActiveAgentSummary
	Trust             bus.TrustSummaryResult
	Flow              bus.FlowGraphResult
	Runs              []bus.RunSummary
	Events            []bus.EventSummary
	QueuedEvents      int
	CostTodayEUR      float64
	CostMonthEUR      float64
	Warnings          []string
	Warning           string
	Recommended       []bus.RecommendedAction
	Readiness         bus.ReadinessResult
	Now               time.Time
	DisableAnimations bool
	AnimFrame         int
	EventFeed         components.EventFeedViewModel

	// Rich cockpit geometry (panelised path). Zero values fall back to the
	// flat Render output for plain/json parity.
	Width            int
	Height           int
	CompactThreshold int
	Theme            theme.Theme
}

// Render returns Mission Control textual content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Mission Control\n")
	b.WriteString("=============\n")
	if banner := components.ReadinessBanner(components.ReadinessBannerFromResult(vm.Readiness)); banner != "" {
		b.WriteString(banner)
		b.WriteString("\n\n")
	}
	b.WriteString(RenderHeader(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderRuntimeRuns(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderTrustPane(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderActiveFlowPane(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderAgentTheatrePane(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderRecommendedActionsPane(vm))
	b.WriteString("\n\n")
	b.WriteString(RenderEventsPane(vm))
	return b.String()
}

// RenderHeader renders workspace/runtime header.
func RenderHeader(vm ViewModel) string {
	var b strings.Builder
	workspace := vm.Workspace
	if workspace == "" {
		workspace = "-"
	}
	branch := vm.Branch
	if branch == "" {
		branch = "-"
	}
	session := vm.SessionStatus
	if session == "" {
		session = "inactive"
	}
	fmt.Fprintf(&b, "Workspace: %s  Branch: %s\n", workspace, branch)
	runtimePrefix := ""
	if vm.RuntimeStatus == "running" && !vm.DisableAnimations {
		runtimePrefix = components.ShimmerPrefix(true, vm.AnimFrame)
	}
	fmt.Fprintf(&b, "Runtime: %s%s  Session: %s", runtimePrefix, vm.RuntimeStatus, session)
	return b.String()
}

// RenderRuntimeRuns renders runtime and recent runs section.
func RenderRuntimeRuns(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Runtime\n")
	fmt.Fprintf(&b, "Agents: %d\n", len(vm.ActiveAgents))
	sessionCount := 0
	if vm.SessionStatus == "active" {
		sessionCount = 1
	}
	fmt.Fprintf(&b, "Sessions: %d\n", sessionCount)
	fmt.Fprintf(&b, "Queue: %d\n", vm.QueuedEvents)
	fmt.Fprintf(&b, "Cost today: €%.2f\n", vm.CostTodayEUR)
	fmt.Fprintf(&b, "Cost month: €%.2f\n", vm.CostMonthEUR)
	if vm.Warning != "" || len(vm.Warnings) > 0 {
		b.WriteString("Warnings:\n")
		if vm.Warning != "" {
			fmt.Fprintf(&b, "- %s\n", vm.Warning)
		}
		for _, warning := range vm.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	fmt.Fprintf(&b, "Updated: %s\n\n", vm.Now.Format(time.RFC3339))
	b.WriteString("Recent runs\n")
	if len(vm.Runs) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}
	for i, run := range vm.Runs {
		if i >= 5 {
			break
		}
		fmt.Fprintf(&b, "- %s  %s  %s\n", run.ID, run.Feature, run.Status)
	}
	return b.String()
}

// RenderTrustPane renders trust confidence summary.
func RenderTrustPane(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Trust\n")
	if len(vm.Trust.Dimensions) == 0 {
		b.WriteString("- unavailable\n")
		return b.String()
	}
	for _, dim := range vm.Trust.Dimensions {
		fmt.Fprintf(&b, "- %-13s %s %2.0f%%\n", dim.Label, meter(dim.Score), dim.Score*100)
	}
	fmt.Fprintf(&b, "- Overall       %s %2.0f%%", meter(vm.Trust.Overall), vm.Trust.Overall*100)
	return b.String()
}

// RenderActiveFlowPane renders the active flow projection.
func RenderActiveFlowPane(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Active flow\n")
	flowID := vm.Flow.FlowID
	if flowID == "" {
		flowID = "-"
	}
	b.WriteString(flowID + "\n")
	if len(vm.Flow.Steps) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}
	for _, step := range vm.Flow.Steps {
		fmt.Fprintf(&b, "%s %s  ", flowStatusGlyph(step.Status, !vm.DisableAnimations, vm.AnimFrame), stepLabel(step))
	}
	return strings.TrimRight(b.String(), " ")
}

// RenderAgentTheatrePane renders a compact agent theatre line.
func RenderAgentTheatrePane(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Agent theatre\n")
	if len(vm.ActiveAgents) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}
	for _, ag := range vm.ActiveAgents {
		role := ag.Role
		if role == "" {
			role = "agent"
		}
		agentRef := ag.AgentRef
		if agentRef == "" {
			agentRef = "-"
		}
		fmt.Fprintf(&b, "- %s %s %s\n", role, statusGlyph(ag.Status, !vm.DisableAnimations, vm.AnimFrame), agentRef)
	}
	return strings.TrimRight(b.String(), "\n")
}

// RenderRecommendedActionsPane renders contextual next steps.
func RenderRecommendedActionsPane(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Recommended actions\n")
	if len(vm.Recommended) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}
	for i, action := range vm.Recommended {
		if i >= 6 {
			break
		}
		fmt.Fprintf(&b, "- %s: %s\n", action.Title, action.Description)
		if strings.TrimSpace(action.CLIEquivalent) != "" {
			fmt.Fprintf(&b, "  CLI: %s\n", action.CLIEquivalent)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// RenderEventsPane renders recent events.
func RenderEventsPane(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Recent events\n")
	feed := vm.EventFeed
	feed.Events = vm.Events
	if feed.Limit <= 0 {
		feed.Limit = 5
	}
	b.WriteString(components.RenderEventFeed(feed))
	return b.String()
}

func meter(v float64) string {
	const width = 10
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	filled := int(math.Round(v * float64(width)))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func statusGlyph(status string, animated bool, frame int) string {
	switch status {
	case "running":
		if !animated {
			return "•"
		}
		return components.ShimmerPrefix(true, frame) + "⠋"
	case "done":
		return "✓"
	case "failed":
		return "✕"
	case "blocked":
		return "⊘"
	default:
		return "○"
	}
}

func flowStatusGlyph(status string, animated bool, frame int) string {
	switch status {
	case "succeeded", "completed", "done":
		return "✓"
	case "running":
		if !animated {
			return "•"
		}
		return components.ShimmerPrefix(true, frame) + "⠋"
	case "failed":
		return "✕"
	default:
		return "○"
	}
}

func stepLabel(step bus.FlowGraphStep) string {
	if strings.TrimSpace(step.Label) != "" {
		return step.Label
	}
	return step.ID
}
