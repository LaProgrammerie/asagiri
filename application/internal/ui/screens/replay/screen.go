package replay

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	tea "github.com/charmbracelet/bubbletea"
)

// Model drives interactive replay explorer state.
type Model struct {
	EventCursor   int
	CompareReplay string
	Focused       bool
	ShowJump      bool
	ShowArtifact  bool
	Detail        *bus.ReplayEventDetail
	Compare       *bus.ReplayCompareResult
}

// NewModel returns default replay explorer interaction state.
func NewModel() Model {
	return Model{Focused: true}
}

// ViewModel contains replay explorer data.
type ViewModel struct {
	Replay  bus.ReplayPackageResult
	Detail  *bus.ReplayEventDetail
	Compare *bus.ReplayCompareResult
	Model   Model
	ShowCLI bool
}

// Update handles replay explorer keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up":
		if m.EventCursor > 0 {
			m.EventCursor--
		}
	case "down":
		m.EventCursor++
	case input.KeyExplorerJump:
		m.ShowJump = true
		m.ShowArtifact = false
	case input.KeyExplorerArtifact:
		m.ShowArtifact = true
		m.ShowJump = true
	case input.KeyExplorerBack:
		m.ShowJump = false
		m.ShowArtifact = false
		m.Detail = nil
		m.Compare = nil
	}
	return m, nil
}

// SelectIndex moves the event cursor within bounds.
func (m *Model) SelectIndex(index, count int) {
	if count <= 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	m.EventCursor = index
	m.Focused = true
}

// SelectedEventIndex returns clamped event cursor.
func (m Model) SelectedEventIndex(timelineLen int) int {
	if timelineLen == 0 {
		return 0
	}
	idx := m.EventCursor
	if idx < 0 {
		return 0
	}
	if idx >= timelineLen {
		return timelineLen - 1
	}
	return idx
}

// Render returns replay timeline content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Replay\n")
	b.WriteString("ID: " + value(vm.Replay.ReplayID, "-") + "\n")
	if !vm.Replay.CreatedAt.IsZero() {
		b.WriteString("Created: " + vm.Replay.CreatedAt.Format(time.RFC3339) + "\n")
	}
	if vm.Replay.RepoBranch != "" || vm.Replay.RepoCommit != "" {
		fmt.Fprintf(&b, "Repo: %s @ %s\n", value(vm.Replay.RepoBranch, "-"), shortCommit(vm.Replay.RepoCommit))
	}
	b.WriteString("Mode: " + value(vm.Replay.Mode, "full") + "\n")
	if vm.Replay.Warning != "" {
		b.WriteString("Warning: " + vm.Replay.Warning + "\n")
	}
	if len(vm.Replay.Artifacts) > 0 {
		b.WriteString("\nArtifacts\n")
		for _, artifact := range vm.Replay.Artifacts {
			b.WriteString("- " + artifact + "\n")
		}
	}

	entries := make([]components.TimelineEntry, 0, len(vm.Replay.Timeline))
	for _, event := range vm.Replay.Timeline {
		entries = append(entries, components.TimelineEntry{
			Time:     event.Time,
			Label:    event.Type,
			Artifact: event.Artifact,
		})
	}
	b.WriteString("\nTimeline\n")
	b.WriteString(components.RenderTimelineView(components.TimelineViewModel{
		Entries: entries,
		Cursor:  vm.Model.SelectedEventIndex(len(entries)),
		Focused: vm.Model.Focused,
	}))

	if vm.Model.ShowJump && vm.Detail != nil {
		b.WriteString("\nJump target\n")
		b.WriteString("───────────────────────\n")
		b.WriteString("Event  " + vm.Detail.Type + "\n")
		if !vm.Detail.Time.IsZero() {
			b.WriteString("Time   " + vm.Detail.Time.Format("15:04:05") + "\n")
		}
	}
	if vm.Model.ShowArtifact && vm.Detail != nil && vm.Detail.Artifact != "" {
		b.WriteString("\nArtifact\n")
		b.WriteString("───────────────────────\n")
		b.WriteString("Path   " + value(vm.Detail.ArtifactPath, vm.Detail.Artifact) + "\n")
	}
	if vm.Compare != nil {
		b.WriteString("\nCompare\n")
		b.WriteString("───────────────────────\n")
		for _, line := range vm.Compare.Summary {
			b.WriteString("- " + line + "\n")
		}
		for _, line := range vm.Compare.Divergences {
			b.WriteString("- " + line + "\n")
		}
	}

	b.WriteString("\nActions\n")
	b.WriteString("- jump to event: j\n")
	b.WriteString("- inspect artifact: a\n")
	b.WriteString("- compare run: m\n")
	b.WriteString("- replay offline: O\n")
	b.WriteString("- explain divergence: D\n")
	if vm.ShowCLI {
		b.WriteString("- open replay: asa replay open " + value(vm.Replay.ReplayID, "<replay-id>") + "\n")
		b.WriteString("- compare: asa replay compare <replay-a> <replay-b>\n")
		b.WriteString("- explain: asa replay explain <replay-a> <replay-b>\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func shortCommit(commit string) string {
	c := strings.TrimSpace(commit)
	if len(c) > 12 {
		return c[:12]
	}
	return c
}
