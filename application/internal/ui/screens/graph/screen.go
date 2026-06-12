package graph

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	tea "github.com/charmbracelet/bubbletea"
)

// Model drives interactive graph explorer state.
type Model struct {
	ViewIndex  int
	NodeCursor int
	Focused    bool
	Detail     *bus.GraphNodeDetail
	ShowDetail bool
}

// NewModel returns default graph explorer interaction state.
func NewModel() Model {
	return Model{}
}

// ViewModel configures graph explorer rendering.
type ViewModel struct {
	Graph      bus.GraphExplorerResult
	View       bus.GraphViewResult
	Events     []bus.EventSummary
	NodeDetail *bus.GraphNodeDetail
	Model      Model
	ShowCLI    bool
}

// Update handles graph explorer keys when the screen is active.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case input.KeyExplorerViewCycle:
		m.ViewIndex = (m.ViewIndex + 1) % len(bus.GraphViewModes)
		m.NodeCursor = 0
	case "up":
		if m.NodeCursor > 0 {
			m.NodeCursor--
		}
	case "down":
		m.NodeCursor++
	case input.KeyExplorerBack:
		m.ShowDetail = false
		m.Detail = nil
	case input.KeyExplorerOpen, input.KeyExplorerEnter, input.KeyExplorerDeps:
		m.ShowDetail = true
	}
	return m, nil
}

// CurrentView returns the active graph view mode.
func (m Model) CurrentView() bus.GraphViewMode {
	if m.ViewIndex < 0 || m.ViewIndex >= len(bus.GraphViewModes) {
		return bus.GraphViewTimeline
	}
	return bus.GraphViewModes[m.ViewIndex]
}

// SelectIndex moves the node cursor within bounds.
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
	m.NodeCursor = index
	m.Focused = true
}

// SelectedNode returns the node at the cursor for the current view.
func (m Model) SelectedNode(view bus.GraphViewResult, fallback []bus.GraphNodeSummary) *bus.GraphNodeSummary {
	nodes := view.Nodes
	if len(nodes) == 0 {
		nodes = fallback
	}
	if len(nodes) == 0 {
		return nil
	}
	idx := m.NodeCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(nodes) {
		idx = len(nodes) - 1
	}
	node := nodes[idx]
	return &node
}

// Render returns the graph explorer textual view.
func Render(vm ViewModel) string {
	view := vm.View.View
	if view == "" {
		view = vm.Model.CurrentView()
	}
	nodes := vm.View.Nodes
	if len(nodes) == 0 {
		nodes = vm.Graph.Nodes
	}

	var b strings.Builder
	b.WriteString("Graph: " + empty(vm.Graph.GraphID, "-") + "\n")
	b.WriteString("Flow: " + empty(vm.Graph.FlowID, "-") + "\n")
	b.WriteString("Status: " + empty(vm.Graph.Status, "unknown") + "\n")
	b.WriteString("View: " + string(view) + "\n")
	if len(vm.View.Groups) > 0 {
		for _, group := range vm.View.Groups {
			b.WriteString("  " + group + "\n")
		}
	}
	if vm.Graph.Warning != "" {
		b.WriteString("Warning: " + vm.Graph.Warning + "\n")
	}
	b.WriteString("\nNodes\n")
	if len(nodes) == 0 {
		b.WriteString("- none\n")
	} else {
		items := make([]components.TreeItem, 0, len(nodes))
		for _, node := range nodes {
			meta := "[" + empty(node.Status, "unknown") + "] risk=" + empty(node.Risk, "unknown")
			if len(node.BlockedBy) > 0 {
				meta += " blocked_by=" + strings.Join(node.BlockedBy, ",")
			}
			items = append(items, components.TreeItem{
				ID:    node.ID,
				Label: empty(node.Title, node.ID),
				Meta:  meta,
			})
		}
		b.WriteString(components.RenderTreeView(components.TreeViewModel{
			Items:   items,
			Cursor:  vm.Model.NodeCursor,
			Focused: vm.Model.Focused,
		}))
	}
	if vm.Model.ShowDetail && vm.NodeDetail != nil {
		b.WriteString("\nSelected: " + empty(vm.NodeDetail.Title, vm.NodeDetail.NodeID) + "\n")
		b.WriteString("───────────────────────\n")
		b.WriteString("Status   " + empty(vm.NodeDetail.Status, "unknown") + "\n")
		b.WriteString("Risk     " + empty(vm.NodeDetail.Risk, "unknown") + "\n")
		b.WriteString("Type     " + empty(vm.NodeDetail.Type, "unknown") + "\n")
		if len(vm.NodeDetail.Dependencies) > 0 {
			b.WriteString("Depends  " + strings.Join(vm.NodeDetail.Dependencies, ", ") + "\n")
		}
		if len(vm.NodeDetail.Dependents) > 0 {
			b.WriteString("Blocks   " + strings.Join(vm.NodeDetail.Dependents, ", ") + "\n")
		}
		if len(vm.NodeDetail.BlockedBy) > 0 {
			b.WriteString("Blocked  " + strings.Join(vm.NodeDetail.BlockedBy, ", ") + "\n")
		}
	}
	b.WriteString("\nActions\n")
	b.WriteString("- open node: Enter/o\n")
	b.WriteString("- view logs: l\n")
	b.WriteString("- view dependencies: d\n")
	b.WriteString("- explain decision: e\n")
	b.WriteString("- launch/resume: r\n")
	b.WriteString("- export mermaid/json: x\n")
	b.WriteString("- switch view: v\n")
	if vm.ShowCLI {
		b.WriteString("CLI: asa graph visualize " + empty(vm.Graph.GraphID, "<graph-id>") + " --format mermaid\n")
	}
	b.WriteString("\nEvents\n")
	b.WriteString(components.RenderEventFeed(components.EventFeedViewModel{
		Events:       vm.Events,
		Limit:        5,
		ShowCLIHints: vm.ShowCLI,
	}))
	return strings.TrimRight(b.String(), "\n")
}

func empty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
