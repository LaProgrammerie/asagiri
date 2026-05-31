package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Panel renders a titled container.
func Panel(title, body string, th theme.Theme) string {
	st := th.Styles()
	header := st.PanelTitle.Render(title)
	content := st.Fg.Render(body)
	return st.Panel.Render(header + "\n" + content)
}

// PanelSized renders a titled container clamped to a w×h footprint (border
// included). Derived from screens/onboarding.eosColumn: full border, side
// padding, height clamp so a pane never overflows the shared cockpit layout.
func PanelSized(title, body string, w, h int, th theme.Theme) string {
	st := th.Styles()
	p := st.Theme.Palette
	if w < 6 {
		w = 6
	}
	if h < 3 {
		h = 3
	}
	// Full border consumes 2 cols/rows; horizontal padding consumes 2 cols.
	innerW := w - 4
	if innerW < 1 {
		innerW = 1
	}

	var lines []string
	if strings.TrimSpace(title) != "" {
		lines = append(lines, st.PanelTitle.Render(clampLine(title, innerW)))
	}
	if body != "" {
		lines = append(lines, st.Fg.Render(body))
	}
	content := strings.Join(lines, "\n")

	// Clamp content to the inner height (border rows = 2) so the body never
	// pushes the panel past h rows.
	if maxRows := h - 2; maxRows > 0 {
		rendered := strings.Split(content, "\n")
		if len(rendered) > maxRows {
			content = strings.Join(rendered[:maxRows], "\n")
		}
	}

	border := lipgloss.RoundedBorder()
	if th.IsHighContrast() {
		border = lipgloss.ThickBorder()
	}
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(p.Border)).
		Background(lipgloss.Color(p.Background)).
		Padding(0, 1).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}

// clampLine truncates a single line to at most maxW display cells.
func clampLine(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	r := []rune(s)
	if maxW <= 1 {
		return string(r[:1])
	}
	out := string(r)
	for lipgloss.Width(out) > maxW-1 && len(out) > 0 {
		rr := []rune(out)
		out = string(rr[:len(rr)-1])
	}
	return out + "…"
}
