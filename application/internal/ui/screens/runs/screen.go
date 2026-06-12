package runs

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model drives Runs screen interaction (list selection).
type Model struct {
	Cursor  int
	Focused bool
}

// NewModel returns default Runs interaction state.
func NewModel() Model { return Model{Focused: true} }

// Update handles keyboard navigation in the runs list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		m.Cursor++
	}
	return m, nil
}

// SelectIndex clamps the cursor within [0,count).
func (m *Model) SelectIndex(index, count int) {
	if count <= 0 {
		m.Cursor = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	m.Cursor = index
	m.Focused = true
}

// Selected returns the clamped selected index for a list of n runs.
func (m Model) Selected(n int) int {
	if n <= 0 {
		return 0
	}
	idx := m.Cursor
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

// SelectedRunID returns the id of the currently selected run, or "".
func (m Model) SelectedRunID(runs []bus.RunSummary) string {
	if len(runs) == 0 {
		return ""
	}
	return runs[m.Selected(len(runs))].ID
}

// ViewModel contains Runs screen data.
type ViewModel struct {
	Runs      []bus.RunSummary
	Detail    bus.RunDetail
	Model     Model
	Readiness bus.ReadinessResult
	ShowCLI   bool
	Width     int
	Height    int
	Theme     theme.Theme
}

// Render returns the Runs screen content: list pane + detail pane. When the
// repository has no runs or is not onboarded, it renders an empty state that
// explicitly invites onboarding (R7.6).
func Render(vm ViewModel) string {
	if !vm.Readiness.Ready || len(vm.Runs) == 0 {
		return renderEmptyState(vm)
	}
	selected := vm.Model.Selected(len(vm.Runs))
	list := renderRunList(vm, selected)
	detail := renderRunDetail(vm)

	if vm.Width >= 90 && vm.Height > 0 {
		th := vm.Theme
		listW := 34
		detailW := vm.Width - listW - 1
		if detailW < 40 {
			detailW = 40
		}
		h := vm.Height
		if h < 8 {
			h = 8
		}
		left := components.PanelSized("Runs", list, listW, h, th)
		right := components.PanelSized("Run detail", detail, detailW, h, th)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	}
	return "Runs\n" + list + "\n\nRun detail\n" + detail
}

func renderRunList(vm ViewModel, selected int) string {
	var b strings.Builder
	for i, run := range vm.Runs {
		marker := "  "
		if i == selected {
			marker = "▸ "
		}
		feature := run.Feature
		if feature == "" {
			feature = run.ID
		}
		fmt.Fprintf(&b, "%s%s %s  %s\n", marker, statusGlyph(run.Status), feature, run.Status)
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderRunDetail(vm ViewModel) string {
	d := vm.Detail
	var b strings.Builder
	if d.Warning != "" {
		b.WriteString("Warning: " + d.Warning + "\n\n")
	}
	feature := d.Feature
	if feature == "" {
		feature = "-"
	}
	b.WriteString("Run: " + value(d.ID, "-") + "\n")
	b.WriteString("Feature: " + feature + "  Status: " + value(d.Status, "-") + "\n")
	if d.Worktree != "" {
		b.WriteString("Worktree: " + d.Worktree + "\n")
	}

	b.WriteString("\nPipeline\n")
	b.WriteString(renderPipeline(d.Pipeline) + "\n")

	b.WriteString("\nValidation: " + value(d.Validation, "—") + "\n")

	b.WriteString("\nTrust gate\n")
	b.WriteString(renderTrustGate(vm) + "\n")

	fmt.Fprintf(&b, "\nCost: €%.2f\n", d.CostEUR)

	if len(d.Agents) > 0 {
		b.WriteString("\nAgents\n")
		for i, ag := range d.Agents {
			if i >= 4 {
				break
			}
			role := value(ag.Role, "agent")
			ref := value(ag.AgentRef, "-")
			fmt.Fprintf(&b, "- %s %s %s\n", role, statusGlyph(ag.Status), ref)
		}
	}

	if len(d.Events) > 0 {
		b.WriteString("\nRecent events\n")
		for i, ev := range d.Events {
			if i >= 5 {
				break
			}
			b.WriteString("- " + ev.Type + "\n")
		}
	}

	b.WriteString("\nActions\n")
	b.WriteString("- enter open   t trust   g graph   r replay\n")
	if vm.ShowCLI {
		b.WriteString("- CLI: asa runs\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderPipeline(steps []bus.RunPipelineStep) string {
	if len(steps) == 0 {
		return "- none"
	}
	parts := make([]string, 0, len(steps))
	for _, s := range steps {
		label := s.Label
		if label == "" {
			label = s.ID
		}
		parts = append(parts, statusGlyph(s.Status)+" "+label)
	}
	return strings.Join(parts, "  →  ")
}

func renderTrustGate(vm ViewModel) string {
	st := vm.Theme.Styles()
	overall := vm.Detail.TrustGate.Overall
	if overall <= 0 && len(vm.Detail.TrustGate.Dimensions) == 0 {
		return "- unavailable"
	}
	return st.RenderBarGauge("Overall", int(overall*100), 10, 16)
}

func renderEmptyState(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("No runs yet.\n\n")
	if !vm.Readiness.Ready {
		// R7.6: the repository is not onboarded — invite onboarding explicitly.
		b.WriteString("This repository is not onboarded yet.\n")
	} else {
		b.WriteString("This repository has no recorded runs.\n")
	}
	b.WriteString("Onboard the project to get started:\n\n")
	b.WriteString("  asa onboard --ui\n")
	return b.String()
}

func statusGlyph(status string) string {
	switch status {
	case "completed", "done", "success", "succeeded":
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

func value(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
