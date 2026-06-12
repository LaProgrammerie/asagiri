package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// DrawerViewModel configures a side drawer panel.
type DrawerViewModel struct {
	Title    string
	Body     string
	Position string // left|right
}

// RenderDrawer renders a drawer panel.
func RenderDrawer(vm DrawerViewModel, th theme.Theme) string {
	title := vm.Title
	if title == "" {
		title = "Drawer"
	}
	pos := strings.TrimSpace(vm.Position)
	if pos != "" {
		title = title + " (" + pos + ")"
	}
	return Panel(title, vm.Body, th)
}
