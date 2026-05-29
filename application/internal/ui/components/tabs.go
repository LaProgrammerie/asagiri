package components

import (
	"fmt"
	"strings"
)

// TabsViewModel configures tab rendering.
type TabsViewModel struct {
	Labels  []string
	Active  int
	Focused bool
}

// RenderTabs renders a horizontal tab bar.
func RenderTabs(vm TabsViewModel) string {
	if len(vm.Labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(vm.Labels))
	for i, label := range vm.Labels {
		prefix := " "
		if vm.Focused && i == vm.Active {
			prefix = "▸"
		}
		style := "["
		if i == vm.Active {
			style = "("
		}
		end := "]"
		if i == vm.Active {
			end = ")"
		}
		parts = append(parts, fmt.Sprintf("%s%s%s%s%s", prefix, style, label, end, ""))
	}
	return strings.Join(parts, "  ")
}
