package settings

import (
	"fmt"
	"strings"
)

// ViewModel is the data contract for settings screen.
type ViewModel struct {
	ThemeName       string
	MouseEnabled    bool
	Animations      bool
	AvailableThemes []string
}

// Render returns a compact settings screen.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Settings\n")
	b.WriteString("========\n")
	fmt.Fprintf(&b, "Theme: %s\n", vm.ThemeName)
	fmt.Fprintf(&b, "Mouse: %t\n", vm.MouseEnabled)
	fmt.Fprintf(&b, "Animations: %t\n", vm.Animations)
	b.WriteString("\nThemes:\n")
	for _, name := range vm.AvailableThemes {
		prefix := "- "
		if name == vm.ThemeName {
			prefix = "* "
		}
		b.WriteString(prefix + name + "\n")
	}
	return b.String()
}
