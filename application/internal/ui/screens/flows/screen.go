package flows

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	tea "github.com/charmbracelet/bubbletea"
)

// Model drives interactive flow explorer state.
type Model struct {
	StepCursor int
	Focused    bool
}

// NewModel returns default flow explorer interaction state.
func NewModel() Model {
	return Model{}
}

// ViewModel contains flow explorer data.
type ViewModel struct {
	Flow    bus.FlowExplorerResult
	Step    bus.FlowStepDetail
	Model   Model
	ShowCLI bool
}

// Update handles flow explorer keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up":
		if m.StepCursor > 0 {
			m.StepCursor--
		}
	case "down":
		m.StepCursor++
	case input.KeyExplorerEnter, input.KeyExplorerOpen:
		return m, nil
	}
	return m, nil
}

// SelectIndex moves the step cursor within bounds.
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
	m.StepCursor = index
	m.Focused = true
}

// SelectedStepID returns the selected step id.
func (m Model) SelectedStepID(flow bus.FlowExplorerResult) string {
	if len(flow.Steps) == 0 {
		return ""
	}
	idx := m.StepCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(flow.Steps) {
		idx = len(flow.Steps) - 1
	}
	return flow.Steps[idx].ID
}

// Render returns flow explorer content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Flow: " + value(vm.Flow.FlowID, "-") + "\n")
	if vm.Flow.Warning != "" {
		b.WriteString("Warning: " + vm.Flow.Warning + "\n")
	}
	if len(vm.Flow.Steps) == 0 {
		b.WriteString("- none\n")
		return strings.TrimRight(b.String(), "\n")
	}

	type stepLine struct {
		label string
		meta  string
	}
	items := make([]stepLine, 0, len(vm.Flow.Steps))
	for _, step := range vm.Flow.Steps {
		items = append(items, stepLine{
			label: statusGlyph(step.Status) + " " + value(step.Label, step.ID),
			meta:  "(" + value(step.Status, "unknown") + ")",
		})
	}
	for i, item := range items {
		if i > 0 {
			b.WriteString("  ↓\n")
		}
		prefix := " "
		if vm.Model.Focused && i == vm.Model.StepCursor {
			prefix = ">"
		}
		b.WriteString(prefix + " " + item.label + " " + item.meta + "\n")
	}

	selected := vm.Step
	if selected.ID == "" {
		selected = vm.Flow.Steps[clampIndex(vm.Model.StepCursor, len(vm.Flow.Steps))]
	}
	b.WriteString("\nSelected: " + value(selected.Label, selected.ID) + "\n")
	b.WriteString("───────────────────────\n")
	b.WriteString("API        " + value(selected.API, "n/a") + "\n")
	b.WriteString("Service    " + value(selected.Service, "n/a") + "\n")
	b.WriteString("Event      " + value(selected.Event, "n/a") + "\n")
	if len(selected.Tests) > 0 {
		b.WriteString("Tests      " + strings.Join(selected.Tests, ", ") + "\n")
	} else {
		b.WriteString("Tests      n/a\n")
	}
	if len(selected.Metrics) > 0 {
		b.WriteString("Metrics    " + strings.Join(selected.Metrics, ", ") + "\n")
	} else {
		b.WriteString("Metrics    n/a\n")
	}
	b.WriteString(fmt.Sprintf("Trust      %.0f%%\n", selected.TrustScore*100))
	b.WriteString("Risk       " + value(selected.Risk, "unknown") + "\n")
	b.WriteString("\nKeys: ↑↓ select step  Enter open\n")
	if vm.ShowCLI {
		b.WriteString("CLI: asa flow open " + value(vm.Flow.FlowID, "<flow>") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func clampIndex(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func statusGlyph(status string) string {
	switch status {
	case "succeeded", "completed", "done":
		return "✓"
	case "running":
		return "⠋"
	case "failed":
		return "✕"
	default:
		return "○"
	}
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
