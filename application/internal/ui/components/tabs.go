package components

import (
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// TabsViewModel configures tab rendering.
type TabsViewModel struct {
	Labels  []string
	Active  int
	Focused bool
	Theme   theme.Theme
}

// RenderTabs renders a horizontal tab bar with lipgloss pill styling.
func RenderTabs(vm TabsViewModel) string {
	if len(vm.Labels) == 0 {
		return ""
	}
	th := vm.Theme
	if th.Name == "" {
		th = theme.Default()
	}
	st := th.Styles()
	return st.RenderTabBar(vm.Labels, vm.Active)
}
