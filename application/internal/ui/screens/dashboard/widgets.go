package dashboard

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// Size is the minimum widget footprint.
type Size struct {
	Width  int
	Height int
}

// Widget defines the lot-2 dashboard widget contract.
type Widget interface {
	tea.Model
	Title() string
	MinSize() Size
}

type runtimeWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type agentWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type trustWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type costWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type flowWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type eventWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type progressWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}

func RuntimeWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return runtimeWidget{snapshot: snapshot, animated: animated}
}
func AgentWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return agentWidget{snapshot: snapshot, animated: animated}
}
func TrustWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return trustWidget{snapshot: snapshot, animated: animated}
}
func CostWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return costWidget{snapshot: snapshot, animated: animated}
}
func FlowWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return flowWidget{snapshot: snapshot, animated: animated}
}
func EventWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return eventWidget{snapshot: snapshot, animated: animated}
}
func ProgressWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return progressWidget{snapshot: snapshot, animated: animated}
}

func (w runtimeWidget) Init() tea.Cmd  { return nil }
func (w agentWidget) Init() tea.Cmd    { return nil }
func (w trustWidget) Init() tea.Cmd    { return nil }
func (w costWidget) Init() tea.Cmd     { return nil }
func (w flowWidget) Init() tea.Cmd     { return nil }
func (w eventWidget) Init() tea.Cmd    { return nil }
func (w progressWidget) Init() tea.Cmd { return nil }

func (w runtimeWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)  { return w, nil }
func (w agentWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)    { return w, nil }
func (w trustWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)    { return w, nil }
func (w costWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w flowWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w eventWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)    { return w, nil }
func (w progressWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }

func (w runtimeWidget) Title() string  { return "Runtime" }
func (w agentWidget) Title() string    { return "Agents" }
func (w trustWidget) Title() string    { return "Trust" }
func (w costWidget) Title() string     { return "Costs" }
func (w flowWidget) Title() string     { return "Flow" }
func (w eventWidget) Title() string    { return "Events" }
func (w progressWidget) Title() string { return "Progress" }

func (w runtimeWidget) MinSize() Size  { return Size{Width: 32, Height: 5} }
func (w agentWidget) MinSize() Size    { return Size{Width: 32, Height: 5} }
func (w trustWidget) MinSize() Size    { return Size{Width: 32, Height: 5} }
func (w costWidget) MinSize() Size     { return Size{Width: 32, Height: 4} }
func (w flowWidget) MinSize() Size     { return Size{Width: 40, Height: 4} }
func (w eventWidget) MinSize() Size    { return Size{Width: 40, Height: 5} }
func (w progressWidget) MinSize() Size { return Size{Width: 32, Height: 4} }

func (w runtimeWidget) View() string {
	st := w.snapshot.Runtime.Status
	return strings.Join([]string{
		fmt.Sprintf("Status: %s", runtimeLabel(st.Running)),
		fmt.Sprintf("Sessions: %d", st.Sessions),
		fmt.Sprintf("Flows: %d", st.FlowsActive),
		fmt.Sprintf("Queue: %d", st.QueuedEvents),
	}, "\n")
}

func (w agentWidget) View() string {
	if len(w.snapshot.ActiveAgents) == 0 {
		return "No active agents"
	}
	lines := make([]string, 0, minInt(4, len(w.snapshot.ActiveAgents)))
	for i, ag := range w.snapshot.ActiveAgents {
		if i >= 4 {
			break
		}
		role := ag.Role
		if role == "" {
			role = "agent"
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", statusGlyph(ag.Status, w.animated), role, emptyDash(ag.AgentRef)))
	}
	return strings.Join(lines, "\n")
}

func (w trustWidget) View() string {
	if len(w.snapshot.Trust.Dimensions) == 0 {
		return "No trust report"
	}
	lines := make([]string, 0, len(w.snapshot.Trust.Dimensions)+1)
	for _, dim := range w.snapshot.Trust.Dimensions {
		lines = append(lines, fmt.Sprintf("%-13s %2.0f%%", dim.Label, dim.Score*100))
	}
	lines = append(lines, fmt.Sprintf("%-13s %2.0f%%", "Overall", w.snapshot.Trust.Overall*100))
	return strings.Join(lines, "\n")
}

func (w costWidget) View() string {
	return strings.Join([]string{
		fmt.Sprintf("Today: €%.2f", w.snapshot.CostTodayEUR),
		fmt.Sprintf("Month: €%.2f", w.snapshot.CostMonthEUR),
	}, "\n")
}

func (w flowWidget) View() string {
	if len(w.snapshot.Flow.Steps) == 0 {
		return "No active flow"
	}
	labels := make([]string, 0, minInt(4, len(w.snapshot.Flow.Steps)))
	for i, step := range w.snapshot.Flow.Steps {
		if i >= 4 {
			break
		}
		labels = append(labels, fmt.Sprintf("%s %s", flowStatusGlyph(step.Status, w.animated), emptyDash(step.Label)))
	}
	return strings.Join(labels, "   ")
}

func (w eventWidget) View() string {
	out := components.RenderEventFeed(components.EventFeedViewModel{
		Events:       w.snapshot.Events,
		Limit:        4,
		ShowCLIHints: false,
	})
	if strings.Contains(out, "- none") {
		return "No events"
	}
	return out
}

func (w progressWidget) View() string {
	if len(w.snapshot.Runs) == 0 {
		return "No runs"
	}
	var done int
	for _, run := range w.snapshot.Runs {
		switch run.Status {
		case "completed", "done", "success":
			done++
		}
	}
	total := len(w.snapshot.Runs)
	ratio := float64(done) / float64(total)
	return strings.Join([]string{
		fmt.Sprintf("Completed: %d/%d", done, total),
		fmt.Sprintf("Progress: %s %2.0f%%", meter(ratio), ratio*100),
	}, "\n")
}

func runtimeLabel(running bool) string {
	if running {
		return "running"
	}
	return "stopped"
}

func emptyDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func statusGlyph(status string, animated bool) string {
	switch status {
	case "running":
		if !animated {
			return "•"
		}
		return "⠋"
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

func flowStatusGlyph(status string, animated bool) string {
	switch status {
	case "succeeded", "completed", "done":
		return "✓"
	case "running":
		if !animated {
			return "•"
		}
		return "⠋"
	case "failed":
		return "✕"
	default:
		return "○"
	}
}

func meter(ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	width := 10
	filled := int(ratio * float64(width))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
