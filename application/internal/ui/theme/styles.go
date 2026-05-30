package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles bundles reusable lipgloss styles for the Asagiri TUI.
type Styles struct {
	Theme Theme

	Brand       lipgloss.Style
	AppTitle    lipgloss.Style
	Subtitle    lipgloss.Style
	Fg          lipgloss.Style
	Muted       lipgloss.Style
	Bold        lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Error       lipgloss.Style
	Panel       lipgloss.Style
	ContentArea lipgloss.Style
	PanelTitle  lipgloss.Style
	Tab         lipgloss.Style
	TabActive   lipgloss.Style
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style
	ButtonGhost lipgloss.Style
	FieldLabel  lipgloss.Style
	FieldValue  lipgloss.Style
	FieldFocus  lipgloss.Style
	StatusBar   lipgloss.Style
	StatusLeft  lipgloss.Style
	StatusRight lipgloss.Style
	Hint        lipgloss.Style
	AccentBlock lipgloss.Style
	Divider     lipgloss.Style
	CheckOK     lipgloss.Style
	CheckWarn   lipgloss.Style
	CheckFail   lipgloss.Style
}

// Styles returns lipgloss styles derived from the theme palette.
func (t Theme) Styles() Styles {
	p := t.Palette
	fg := p.Foreground
	if fg == "" {
		fg = "#E5E7EB"
	}
	bg := p.Background
	if bg == "" {
		bg = "#0D0F14"
	}
	border := lipgloss.RoundedBorder()
	if t.IsHighContrast() {
		border = lipgloss.ThickBorder()
	}

	return Styles{
		Theme: t,
		Brand: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 1),
		AppTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 2),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)).
			Italic(true),
		Fg: lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)),
		Bold: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(fg)),
		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.Success)),
		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.Warning)),
		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.Error)),
		Panel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(lipgloss.Color(p.Border)).
			Padding(0, 1),
		ContentArea: lipgloss.NewStyle().
			Background(lipgloss.Color(p.Border)).
			Padding(1, 2),
		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(p.Primary)),
		Tab: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)).
			Padding(0, 1).
			MarginRight(1),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 1).
			MarginRight(1),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Border)).
			Padding(0, 2).
			MarginRight(1),
		ButtonFocus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(bg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 2).
			MarginRight(1),
		ButtonGhost: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)).
			Padding(0, 2).
			MarginRight(1),
		FieldLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)).
			Width(22).
			Align(lipgloss.Right).
			MarginRight(1),
		FieldValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)),
		FieldFocus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Border)).
			Padding(0, 1),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Border)).
			Padding(0, 1),
		StatusLeft: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(bg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(0, 1),
		StatusRight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Muted)),
		Hint: lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color(p.Muted)),
		AccentBlock: lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(p.Primary)).
			Padding(1, 2),
		Divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Border)),
		CheckOK: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Success)),
		CheckWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Warning)),
		CheckFail: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Error)),
	}
}

// RenderBrandStack renders a staggered brand mark (Lip Gloss–style hero).
func (s Styles) RenderBrandStack(word string, lines int) string {
	if lines < 1 {
		lines = 1
	}
	if lines > 5 {
		lines = 5
	}
	var parts []string
	for i := 0; i < lines; i++ {
		pad := strings.Repeat(" ", i*2)
		parts = append(parts, pad+s.Brand.Render(word))
	}
	return strings.Join(parts, "\n")
}

// RenderButton renders a pill button; focused uses primary fill.
func (s Styles) RenderButton(label string, focused bool) string {
	if focused {
		return s.ButtonFocus.Render(label)
	}
	return s.Button.Render(label)
}

// RenderTabBar renders horizontal step tabs.
func (s Styles) RenderTabBar(labels []string, active int) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for i, label := range labels {
		if i == active {
			parts = append(parts, s.TabActive.Render(label))
		} else {
			parts = append(parts, s.Tab.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// RenderField renders one label/value row with optional focus ring.
func (s Styles) RenderField(label, value string, focused bool) string {
	if value == "" {
		value = "—"
	}
	valStyle := s.FieldValue
	prefix := "  "
	if focused {
		valStyle = s.FieldFocus
		prefix = "▸ "
	}
	labelPart := s.FieldLabel.Render(label)
	valuePart := valStyle.Render(" " + value)
	return prefix + lipgloss.JoinHorizontal(lipgloss.Top, labelPart, valuePart)
}

// RenderProgress renders a score bar (0–100) with semantic color.
func (s Styles) RenderProgress(score, max int, width int) string {
	if max <= 0 {
		max = 100
	}
	if width <= 0 {
		width = 24
	}
	if score < 0 {
		score = 0
	}
	if score > max {
		score = max
	}
	ratio := float64(score) / float64(max)
	filled := int(ratio * float64(width))
	barStyle := s.Success
	if ratio < 0.5 {
		barStyle = s.Error
	} else if ratio < 0.8 {
		barStyle = s.Warning
	}
	filledStr := barStyle.Render(strings.Repeat("█", filled))
	emptyStr := s.Divider.Render(strings.Repeat("░", width-filled))
	return filledStr + emptyStr + s.Bold.Render(fmt.Sprintf(" %d/%d", score, max))
}

// RenderCheckStatus formats a readiness check line.
func (s Styles) RenderCheckStatus(status, id, message string) string {
	icon := "○"
	style := s.Muted
	switch status {
	case "ok":
		icon = "✓"
		style = s.CheckOK
	case "warn":
		icon = "⚠"
		style = s.CheckWarn
	case "fail":
		icon = "✕"
		style = s.CheckFail
	}
	line := icon + " " + id
	if message != "" {
		line += ": " + message
	}
	return style.Render(line)
}

// RenderStatusBar renders a full-width footer strip.
func (s Styles) RenderStatusBar(leftLabel, leftValue, right string) string {
	left := s.StatusLeft.Render(leftLabel) + " " + s.Fg.Render(leftValue)
	if right == "" {
		return s.StatusBar.Render(left)
	}
	return s.StatusBar.Render(left + "  " + s.StatusRight.Render(right))
}

// RenderPageHeader renders a compact single-line header (brand + title + meta).
func (s Styles) RenderPageHeader(title, meta string) string {
	brand := s.Brand.Render("ASAGIRI")
	head := lipgloss.JoinHorizontal(lipgloss.Center, brand, "  ", s.Bold.Render(title))
	if meta == "" {
		return head
	}
	return head + "\n" + s.Muted.Render(meta)
}

// RenderStatusBarFull renders a status strip stretched to width (content area only).
func (s Styles) RenderStatusBarFull(width int, leftLabel, leftValue, right string) string {
	if width <= 0 {
		return s.RenderStatusBar(leftLabel, leftValue, right)
	}
	left := s.StatusLeft.Render(leftLabel) + " " + s.Fg.Render(leftValue)
	rightPart := s.StatusRight.Render(right)
	gap := width - lipgloss.Width(left) - lipgloss.Width(rightPart)
	if gap < 1 {
		gap = 1
	}
	content := left + strings.Repeat(" ", gap) + rightPart
	return s.StatusBar.Width(width).Render(content)
}

// RenderHero pairs a brand stack with title and subtitle columns.
func (s Styles) RenderHero(brandWord, title, subtitle string) string {
	return s.RenderPageHeader(title, subtitle)
}
