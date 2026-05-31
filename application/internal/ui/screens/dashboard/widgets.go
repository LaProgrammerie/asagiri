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

// Widget defines the composable dashboard widget contract (spec-ui §23.1).
type Widget interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Title() string
	MinSize() Size
}

type runtimeWidget struct {
	snapshot  bus.MissionControlSnapshotResult
	animated  bool
	animFrame int
}
type agentWidget struct {
	snapshot  bus.MissionControlSnapshotResult
	animated  bool
	animFrame int
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
	snapshot  bus.MissionControlSnapshotResult
	animated  bool
	animFrame int
}
type eventWidget struct {
	snapshot  bus.MissionControlSnapshotResult
	eventFeed components.EventFeedViewModel
	animated  bool
}
type progressWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type riskWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type knowledgeWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type replayWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type performanceWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type sessionsWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type queueWidget struct {
	snapshot bus.MissionControlSnapshotResult
	animated bool
}
type runsSummaryWidget struct {
	snapshot bus.MissionControlSnapshotResult
}

func RuntimeWidget(snapshot bus.MissionControlSnapshotResult, animated bool, animFrame int) Widget {
	return runtimeWidget{snapshot: snapshot, animated: animated, animFrame: animFrame}
}
func AgentWidget(snapshot bus.MissionControlSnapshotResult, animated bool, animFrame int) Widget {
	return agentWidget{snapshot: snapshot, animated: animated, animFrame: animFrame}
}
func TrustWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return trustWidget{snapshot: snapshot, animated: animated}
}
func CostWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return costWidget{snapshot: snapshot, animated: animated}
}
func FlowWidget(snapshot bus.MissionControlSnapshotResult, animated bool, animFrame int) Widget {
	return flowWidget{snapshot: snapshot, animated: animated, animFrame: animFrame}
}
func EventWidget(snapshot bus.MissionControlSnapshotResult, feed components.EventFeedViewModel, animated bool) Widget {
	return eventWidget{snapshot: snapshot, eventFeed: feed, animated: animated}
}
func ProgressWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return progressWidget{snapshot: snapshot, animated: animated}
}
func RiskWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return riskWidget{snapshot: snapshot, animated: animated}
}
func KnowledgeWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return knowledgeWidget{snapshot: snapshot, animated: animated}
}
func ReplayWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return replayWidget{snapshot: snapshot, animated: animated}
}
func PerformanceWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return performanceWidget{snapshot: snapshot, animated: animated}
}
func SessionsWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return sessionsWidget{snapshot: snapshot, animated: animated}
}
func QueueWidget(snapshot bus.MissionControlSnapshotResult, animated bool) Widget {
	return queueWidget{snapshot: snapshot, animated: animated}
}
func RunsSummaryWidget(snapshot bus.MissionControlSnapshotResult) Widget {
	return runsSummaryWidget{snapshot: snapshot}
}

func (w runtimeWidget) Init() tea.Cmd  { return nil }
func (w agentWidget) Init() tea.Cmd    { return nil }
func (w trustWidget) Init() tea.Cmd    { return nil }
func (w costWidget) Init() tea.Cmd     { return nil }
func (w flowWidget) Init() tea.Cmd     { return nil }
func (w eventWidget) Init() tea.Cmd    { return nil }
func (w progressWidget) Init() tea.Cmd { return nil }
func (w riskWidget) Init() tea.Cmd     { return nil }
func (w knowledgeWidget) Init() tea.Cmd { return nil }
func (w replayWidget) Init() tea.Cmd   { return nil }
func (w performanceWidget) Init() tea.Cmd { return nil }
func (w sessionsWidget) Init() tea.Cmd { return nil }
func (w queueWidget) Init() tea.Cmd        { return nil }
func (w runsSummaryWidget) Init() tea.Cmd { return nil }

func (w runtimeWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }
func (w agentWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w trustWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w costWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)      { return w, nil }
func (w flowWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)      { return w, nil }
func (w eventWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w progressWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)  { return w, nil }
func (w riskWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)     { return w, nil }
func (w knowledgeWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }
func (w replayWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)    { return w, nil }
func (w performanceWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }
func (w sessionsWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)  { return w, nil }
func (w queueWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd)       { return w, nil }
func (w runsSummaryWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }

func (w runtimeWidget) Title() string     { return "Runtime" }
func (w agentWidget) Title() string       { return "Agents" }
func (w trustWidget) Title() string       { return "Trust" }
func (w costWidget) Title() string        { return "Costs" }
func (w flowWidget) Title() string        { return "Flow" }
func (w eventWidget) Title() string       { return "Events" }
func (w progressWidget) Title() string    { return "Progress" }
func (w riskWidget) Title() string        { return "Risk" }
func (w knowledgeWidget) Title() string   { return "Knowledge" }
func (w replayWidget) Title() string      { return "Replay" }
func (w performanceWidget) Title() string { return "Performance" }
func (w sessionsWidget) Title() string    { return "Sessions" }
func (w queueWidget) Title() string          { return "Queue" }
func (w runsSummaryWidget) Title() string     { return "Runs" }

func (w runtimeWidget) MinSize() Size      { return Size{Width: 32, Height: 5} }
func (w agentWidget) MinSize() Size        { return Size{Width: 32, Height: 5} }
func (w trustWidget) MinSize() Size        { return Size{Width: 32, Height: 5} }
func (w costWidget) MinSize() Size         { return Size{Width: 32, Height: 4} }
func (w flowWidget) MinSize() Size         { return Size{Width: 40, Height: 4} }
func (w eventWidget) MinSize() Size        { return Size{Width: 40, Height: 5} }
func (w progressWidget) MinSize() Size     { return Size{Width: 32, Height: 4} }
func (w riskWidget) MinSize() Size         { return Size{Width: 32, Height: 4} }
func (w knowledgeWidget) MinSize() Size   { return Size{Width: 36, Height: 4} }
func (w replayWidget) MinSize() Size        { return Size{Width: 36, Height: 4} }
func (w performanceWidget) MinSize() Size  { return Size{Width: 36, Height: 4} }
func (w sessionsWidget) MinSize() Size      { return Size{Width: 28, Height: 3} }
func (w queueWidget) MinSize() Size           { return Size{Width: 28, Height: 3} }
func (w runsSummaryWidget) MinSize() Size      { return Size{Width: 32, Height: 4} }

func (w runtimeWidget) View() string {
	return components.RuntimeCard(w.snapshot.Runtime, w.animated, w.animFrame)
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
		lines = append(lines, components.AgentCardLine(ag, w.animated, w.animFrame))
	}
	return strings.Join(lines, "\n")
}

func (w trustWidget) View() string {
	return components.TrustCard(w.snapshot.Trust)
}

func (w costWidget) View() string {
	return components.CostCard(w.snapshot.CostTodayEUR, w.snapshot.CostMonthEUR)
}

func (w flowWidget) View() string {
	return components.FlowCard(w.snapshot.Flow, w.animated, w.animFrame)
}

func (w eventWidget) View() string {
	feed := w.eventFeed
	feed.Events = w.snapshot.Events
	if feed.Limit <= 0 {
		feed.Limit = 4
	}
	out := components.RenderEventFeed(feed)
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
		fmt.Sprintf("Progress: %s %2.0f%%", components.ProgressBar(ratio, 10), ratio*100),
	}, "\n")
}

func (w riskWidget) View() string {
	return components.RiskCard(w.snapshot.TrustExplorer, w.snapshot.GraphExplorer)
}

func (w knowledgeWidget) View() string {
	return components.KnowledgeCard(w.snapshot.Knowledge)
}

func (w replayWidget) View() string {
	return components.ReplayCard(w.snapshot.Replay)
}

func (w performanceWidget) View() string {
	samples := performanceSamples(w.snapshot)
	return strings.Join([]string{
		components.LiveCounter("Events", len(w.snapshot.Events), 0),
		fmt.Sprintf("Throughput: %s", components.Sparkline(samples, 10)),
		fmt.Sprintf("Runs done: %s", components.ProgressBar(runCompletionRatio(w.snapshot.Runs), 10)),
	}, "\n")
}

func (w sessionsWidget) View() string {
	n := w.snapshot.Runtime.Status.Sessions
	status := "inactive"
	if n > 0 {
		status = "active"
	}
	return strings.Join([]string{
		components.LiveCounter("Active", n, 0),
		fmt.Sprintf("State: %s", status),
	}, "\n")
}

func (w queueWidget) View() string {
	q := w.snapshot.Runtime.Status.QueuedEvents
	samples := queueSparkline(w.snapshot.Events, 6)
	return strings.Join([]string{
		components.LiveCounter("Depth", q, 0),
		fmt.Sprintf("Trend: %s", components.Sparkline(samples, 8)),
	}, "\n")
}

func (w runsSummaryWidget) View() string {
	if len(w.snapshot.Runs) == 0 {
		return "No recent runs"
	}
	var b strings.Builder
	for i, run := range w.snapshot.Runs {
		if i >= 5 {
			break
		}
		feature := run.Feature
		if feature == "" {
			feature = run.ID
		}
		b.WriteString(fmt.Sprintf("%s  %s  %s\n", runStatusGlyph(run.Status), feature, run.Status))
	}
	return strings.TrimRight(b.String(), "\n")
}

func runStatusGlyph(status string) string {
	switch status {
	case "completed", "done", "success":
		return "✓"
	case "running":
		return "•"
	case "failed", "error":
		return "✕"
	case "blocked":
		return "⊘"
	default:
		return "○"
	}
}

func performanceSamples(snapshot bus.MissionControlSnapshotResult) []float64 {
	if len(snapshot.Events) == 0 {
		return []float64{0}
	}
	max := float64(len(snapshot.Events))
	out := make([]float64, 0, minInt(8, len(snapshot.Events)))
	for i, ev := range snapshot.Events {
		if i >= 8 {
			break
		}
		v := float64(i+1) / max
		if ev.Type != "" {
			v = 0.3 + 0.7*v
		}
		out = append(out, v)
	}
	return out
}

func queueSparkline(events []bus.EventSummary, n int) []float64 {
	if len(events) == 0 {
		return []float64{0}
	}
	out := make([]float64, 0, n)
	step := len(events) / n
	if step < 1 {
		step = 1
	}
	for i := 0; i < len(events) && len(out) < n; i += step {
		out = append(out, float64(i+1)/float64(len(events)))
	}
	return out
}

func runCompletionRatio(runs []bus.RunSummary) float64 {
	if len(runs) == 0 {
		return 0
	}
	var done int
	for _, run := range runs {
		switch run.Status {
		case "completed", "done", "success":
			done++
		}
	}
	return float64(done) / float64(len(runs))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
