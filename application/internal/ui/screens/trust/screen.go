package trust

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	tea "github.com/charmbracelet/bubbletea"
)

// Model drives interactive trust explorer state.
type Model struct {
	DimensionCursor int
	Focused         bool
	ShowDetail      bool
	Detail          *bus.TrustDimensionDetail
}

// NewModel returns default trust explorer interaction state.
func NewModel() Model {
	return Model{Focused: true}
}

// ViewModel contains trust explorer data.
type ViewModel struct {
	Trust   bus.TrustExplorerResult
	Detail  *bus.TrustDimensionDetail
	Model   Model
	ShowCLI bool
}

// Update handles trust explorer keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up":
		if m.DimensionCursor > 0 {
			m.DimensionCursor--
		}
	case "down":
		m.DimensionCursor++
	case input.KeyExplorerBack:
		m.ShowDetail = false
		m.Detail = nil
	case input.KeyExplorerOpen, input.KeyExplorerEnter:
		m.ShowDetail = true
	}
	return m, nil
}

// SelectIndex moves the dimension cursor within bounds.
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
	m.DimensionCursor = index
	m.Focused = true
}

// SelectedLabel returns the selected dimension label.
func (m Model) SelectedLabel(trust bus.TrustExplorerResult) string {
	if len(trust.Dimensions) == 0 {
		return ""
	}
	idx := m.DimensionCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(trust.Dimensions) {
		idx = len(trust.Dimensions) - 1
	}
	return trust.Dimensions[idx].Label
}

// Render returns trust explorer textual content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Trust Summary\n")
	fmt.Fprintf(&b, "Overall: %.0f%%  Residual risk: %s\n", vm.Trust.Overall*100, value(vm.Trust.ResidualRisk, "unknown"))
	fmt.Fprintf(&b, "Gate: %s\n", value(vm.Trust.GateStatus, "unknown"))
	if vm.Trust.GateReason != "" {
		b.WriteString("Gate reason: " + vm.Trust.GateReason + "\n")
	}
	if vm.Trust.Warning != "" {
		b.WriteString("Warning: " + vm.Trust.Warning + "\n")
	}
	if len(vm.Trust.Warnings) > 0 {
		b.WriteString("\nWarnings\n")
		for _, warning := range vm.Trust.Warnings {
			b.WriteString("- " + warning + "\n")
		}
	}
	if len(vm.Trust.Dimensions) == 0 {
		b.WriteString("- unavailable")
		return b.String()
	}

	rows := make([]components.TableRow, 0, len(vm.Trust.Dimensions))
	for _, dim := range vm.Trust.Dimensions {
		rows = append(rows, components.TableRow{
			Cells: []string{dim.Label, fmt.Sprintf("%.0f%%", dim.Score*100)},
		})
	}
	b.WriteString("\n")
	b.WriteString(components.RenderTableView(components.TableViewModel{
		Title:   "Dimensions",
		Headers: []string{"Dimension", "Score"},
		Rows:    rows,
		Cursor:  vm.Model.DimensionCursor,
		Focused: vm.Model.Focused,
	}))

	if vm.Model.ShowDetail && vm.Detail != nil {
		b.WriteString("\nSelected: " + vm.Detail.Label + "\n")
		b.WriteString("───────────────────────\n")
		fmt.Fprintf(&b, "Score          %.0f%%\n", vm.Detail.Score*100)
		b.WriteString("Gate           " + value(vm.Detail.GateStatus, "unknown") + "\n")
		if vm.Detail.GateReason != "" {
			b.WriteString("Gate reason    " + vm.Detail.GateReason + "\n")
		}
		b.WriteString("Residual risk  " + value(vm.Detail.ResidualRisk, "unknown") + "\n")
		if len(vm.Detail.Findings) > 0 {
			b.WriteString("\nFindings\n")
			for _, finding := range vm.Detail.Findings {
				b.WriteString("- " + finding + "\n")
			}
		}
		if len(vm.Detail.Evidence) > 0 {
			b.WriteString("\nEvidence\n")
			for _, evidence := range vm.Detail.Evidence {
				b.WriteString("- " + evidence + "\n")
			}
		}
		if len(vm.Detail.Checks) > 0 {
			b.WriteString("\nChecks\n")
			for _, check := range vm.Detail.Checks {
				b.WriteString("- " + check + "\n")
			}
		}
	}
	b.WriteString("\nKeys: ↑↓ select  Enter drill-down  b back\n")
	if vm.ShowCLI {
		b.WriteString("CLI: asa verify trust <flow>\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
