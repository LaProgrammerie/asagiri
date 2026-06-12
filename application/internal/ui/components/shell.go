package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// NavItem is one entry of the persistent navigation rail. Generalised from
// screens/onboarding.navEntry so the shell and the wizard share one rail.
type NavItem struct {
	Icon   string
	Label  string
	Badge  string // optional state badge (active runs, trust alert, queue…)
	Active bool
}

// RenderNavRail renders a vertical navigation rail clamped to a w×h footprint,
// generalised from screens/onboarding.renderEOSNav. The active entry uses the
// NavActive style; optional badges are right-aligned per row.
func RenderNavRail(title string, items []NavItem, w, h int, th theme.Theme) string {
	st := th.Styles()
	p := st.Theme.Palette
	if w < 8 {
		w = 8
	}
	if h < 3 {
		h = 3
	}
	textW := w - 4 // border (1, right side only) + padding (2) + slack
	if textW < 4 {
		textW = 4
	}

	var b strings.Builder
	if strings.TrimSpace(title) != "" {
		b.WriteString(st.SectionHead.Render(clampLine(title, textW)) + "\n\n")
	}
	b.WriteString(renderNavItems(st, items, textW))

	content := b.String()
	if maxRows := h - 2; maxRows > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > maxRows {
			content = strings.Join(lines[:maxRows], "\n")
		}
	}

	border := lipgloss.NormalBorder()
	if th.IsHighContrast() {
		border = lipgloss.ThickBorder()
	}
	return lipgloss.NewStyle().
		Width(w-1).
		Height(h).
		Padding(0, 1).
		Border(border, false, true, false, false).
		BorderForeground(lipgloss.Color(p.Border)).
		Background(lipgloss.Color(p.Background)).
		Render(content)
}

func renderNavItems(st theme.Styles, items []NavItem, w int) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		icon := item.Icon
		if icon == "" {
			icon = "•"
		}
		label := " " + icon + "  " + item.Label
		badge := strings.TrimSpace(item.Badge)
		if badge != "" {
			gap := w - lipgloss.Width(label) - lipgloss.Width(badge) - 1
			if gap < 1 {
				gap = 1
			}
			label = label + strings.Repeat(" ", gap) + badge
		}
		if item.Active {
			lines = append(lines, st.NavActive.Width(w).Render(clampLine(label, w)))
		} else {
			lines = append(lines, st.Muted.Render(clampLine(label, w)))
		}
	}
	return strings.Join(lines, "\n")
}

// RenderTopBar renders a full-width top status strip with a left/right split and
// a bottom border. Generalised from screens/onboarding.renderEOSTopBar.
func RenderTopBar(left, right string, w int, th theme.Theme) string {
	return renderShellBar(left, right, w, th, true)
}

// RenderBottomBar renders a full-width bottom status strip with a top border.
// Generalised from screens/onboarding.renderEOSBottomBar.
func RenderBottomBar(left, right string, w int, th theme.Theme) string {
	return renderShellBar(left, right, w, th, false)
}

func renderShellBar(left, right string, w int, th theme.Theme, top bool) string {
	st := th.Styles()
	p := st.Theme.Palette
	if w < 4 {
		w = 4
	}
	innerW := w - 2
	gap := innerW - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	border := lipgloss.NormalBorder()
	if th.IsHighContrast() {
		border = lipgloss.ThickBorder()
	}
	style := lipgloss.NewStyle().
		Width(w).
		Background(lipgloss.Color(p.Background)).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(0, 1)
	if top {
		style = style.Border(border, false, false, true, false)
	} else {
		style = style.Border(border, true, false, false, false)
	}
	return style.Render(line)
}
