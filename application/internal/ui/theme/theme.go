package theme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette stores terminal color tokens.
type Palette struct {
	Primary    string
	Foreground string
	Muted      string
	Success    string
	Warning    string
	Error      string
	Border     string
	Background string
	Surface    string // card/panel fill; defaults to Border tone when empty
}

// Theme defines one named UI palette.
type Theme struct {
	Name    string
	Palette Palette
}

// Names returns known theme ids.
func Names() []string {
	out := make([]string, 0, len(themes))
	for name := range themes {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Default returns the default lot-1 theme.
func Default() Theme {
	return themes["asagiri-dark"]
}

// Resolve returns one theme by id.
func Resolve(name string) (Theme, error) {
	n := strings.TrimSpace(strings.ToLower(name))
	if n == "" {
		return Default(), nil
	}
	th, ok := themes[n]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme %q", name)
	}
	return th, nil
}

// MustResolve resolves a theme and falls back to default when unknown.
func MustResolve(name string) Theme {
	th, err := Resolve(name)
	if err != nil {
		return Default()
	}
	return th
}

// BorderStyle returns a reusable border style for panel containers.
func (t Theme) BorderStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	if t.IsHighContrast() {
		border = lipgloss.ThickBorder()
	}
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(t.Palette.Border)).
		Foreground(lipgloss.Color(t.Palette.Foreground))
}

// IsHighContrast reports whether the current theme targets accessibility contrast.
func (t Theme) IsHighContrast() bool {
	return strings.EqualFold(strings.TrimSpace(t.Name), "high-contrast")
}
