package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ModalViewModel configures a centered modal dialog.
type ModalViewModel struct {
	Title   string
	Body    string
	Actions []string
	Active  int
}

// RenderModal renders a modal overlay block with styled action buttons.
func RenderModal(vm ModalViewModel, th theme.Theme) string {
	st := th.Styles()
	var b strings.Builder
	b.WriteString(st.Fg.Render(strings.TrimRight(vm.Body, "\n")))
	if len(vm.Actions) > 0 {
		b.WriteString("\n\n")
		parts := make([]string, 0, len(vm.Actions))
		for i, action := range vm.Actions {
			parts = append(parts, st.RenderButton(action, i == vm.Active))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, parts...))
	}
	return Panel(vm.Title, b.String(), th)
}
