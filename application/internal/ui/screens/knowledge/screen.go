package knowledge

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/input"
	tea "github.com/charmbracelet/bubbletea"
)

// Model drives interactive knowledge explorer state.
type Model struct {
	Query       string
	SearchMode  bool
	MatchCursor int
	Focused     bool
	Detail      *bus.KnowledgeMatchDetail
	ShowDetail  bool
}

// NewModel returns default knowledge explorer interaction state.
func NewModel() Model {
	return Model{Focused: true}
}

// ViewModel contains knowledge explorer data.
type ViewModel struct {
	Search  bus.KnowledgeSearchResult
	Detail  *bus.KnowledgeMatchDetail
	Model   Model
	ShowCLI bool
}

// Update handles knowledge explorer keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case input.KeyExplorerSearch:
		m.SearchMode = true
	case "esc":
		m.SearchMode = false
		m.Query = ""
		m.MatchCursor = 0
	case "up":
		if m.MatchCursor > 0 {
			m.MatchCursor--
		}
	case "down":
		m.MatchCursor++
	case input.KeyExplorerBack:
		m.ShowDetail = false
		m.Detail = nil
	case input.KeyExplorerOpen, input.KeyExplorerEnter:
		m.ShowDetail = true
	case "backspace":
		if m.SearchMode && len(m.Query) > 0 {
			m.Query = m.Query[:len(m.Query)-1]
		}
		m.MatchCursor = 0
	default:
		if m.SearchMode && key.Type == tea.KeyRunes && len(key.Runes) > 0 {
			m.Query += string(key.Runes)
			m.MatchCursor = 0
		}
	}
	return m, nil
}

// SelectIndex moves the match cursor within bounds.
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
	m.MatchCursor = index
	m.Focused = true
}

// SelectedMatch returns the match at cursor.
func (m Model) SelectedMatch(search bus.KnowledgeSearchResult) *bus.KnowledgeMatch {
	if len(search.Matches) == 0 {
		return nil
	}
	idx := m.MatchCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(search.Matches) {
		idx = len(search.Matches) - 1
	}
	match := search.Matches[idx]
	return &match
}

// Render returns knowledge explorer content.
func Render(vm ViewModel) string {
	query := strings.TrimSpace(vm.Model.Query)
	if query == "" {
		query = strings.TrimSpace(vm.Search.Query)
	}
	displayQuery := query
	if displayQuery == "" {
		displayQuery = "(type / to search)"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Knowledge search: %s\n", displayQuery))
	if vm.Search.Warning != "" {
		b.WriteString("Warning: " + vm.Search.Warning + "\n")
	}
	if vm.Model.SearchMode {
		b.WriteString("Search mode: " + vm.Model.Query + "█\n")
	}
	if len(vm.Search.Matches) == 0 {
		b.WriteString("- no match\n")
	} else {
		rows := make([]components.TableRow, 0, len(vm.Search.Matches))
		for _, match := range vm.Search.Matches {
			rows = append(rows, components.TableRow{
				Cells: []string{
					match.ID,
					match.Type,
					match.Name,
					fmt.Sprintf("%.0f%%", match.Score*100),
				},
			})
		}
		b.WriteString(components.RenderTableView(components.TableViewModel{
			Headers: []string{"ID", "Type", "Name", "Score"},
			Rows:    rows,
			Cursor:  vm.Model.MatchCursor,
			Focused: vm.Model.Focused,
		}))
	}
	if vm.Model.ShowDetail && vm.Detail != nil {
		b.WriteString("\nSelected: " + vm.Detail.Name + "\n")
		b.WriteString("───────────────────────\n")
		b.WriteString("Path   " + vm.Detail.Path + "\n")
		if len(vm.Detail.RelatedFlows) > 0 {
			b.WriteString("Flows  " + strings.Join(vm.Detail.RelatedFlows, ", ") + "\n")
		}
		if len(vm.Detail.RelatedAPIs) > 0 {
			b.WriteString("APIs   " + strings.Join(vm.Detail.RelatedAPIs, ", ") + "\n")
		}
		if len(vm.Detail.RelatedTests) > 0 {
			b.WriteString("Tests  " + strings.Join(vm.Detail.RelatedTests, ", ") + "\n")
		}
		if len(vm.Detail.RelatedEvents) > 0 {
			b.WriteString("Events " + strings.Join(vm.Detail.RelatedEvents, ", ") + "\n")
		}
	}
	b.WriteString("\nActions\n")
	b.WriteString("- impact analyze: i\n")
	b.WriteString("- build context: c\n")
	b.WriteString("- open graph: g\n")
	b.WriteString("- explain relationship: e\n")
	b.WriteString("- search: /\n")
	if vm.ShowCLI {
		if query != "" {
			b.WriteString(`CLI: asa knowledge query "` + query + `"` + "\n")
		} else {
			b.WriteString("CLI: asa knowledge\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
