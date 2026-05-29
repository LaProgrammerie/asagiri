package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// ModalViewModel configures a centered modal dialog.
type ModalViewModel struct {
	Title   string
	Body    string
	Actions []string
}

// RenderModal renders a modal overlay block.
func RenderModal(vm ModalViewModel, th theme.Theme) string {
	var b strings.Builder
	b.WriteString("╭─ " + vm.Title + " ─╮\n")
	b.WriteString(strings.TrimRight(vm.Body, "\n"))
	if len(vm.Actions) > 0 {
		b.WriteString("\n\n")
		b.WriteString(strings.Join(vm.Actions, "  |  "))
	}
	return Panel("Modal", b.String(), th)
}
