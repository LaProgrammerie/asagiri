package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

func fracString(n, total int) string { return fmt.Sprintf("%d/%d", n, total) }

func pctString(p int) string { return fmt.Sprintf("%d%%", p) }

// validationName extracts a short validation identifier from a preview string.
func validationName(raw string) string {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexAny(raw, ":("); i > 0 {
		raw = strings.TrimSpace(raw[:i])
	}
	if raw == "" {
		return "validation"
	}
	return raw
}

// scoreComponent is one explainable contribution to the readiness score.
// Every component maps to a concrete, observable signal so the score can be
// decomposed (future "Explain") and never relies on magic numbers.
type scoreComponent struct {
	Label  string
	Weight int
	Done   bool
}

// readinessComponents returns the readiness breakdown derived from real signals
// present in the current wizard state.
func readinessComponents(vm ViewModel) []scoreComponent {
	m := vm.Model
	has := func(key string) bool { return strings.TrimSpace(m.fieldValue(key)) != "" }
	return []scoreComponent{
		{"Config projet", 20, has("project_name") && has("default_branch")},
		{"Stack détectée", 20, len(m.DetectedStacks) > 0 || has("stack")},
		{"Validations QA", 20, len(m.ValidationPreview) > 0},
		{"Agents définis", 15, has("default_agent") || has("default_reviewer")},
		{"Docs produit", 15, has("product_one_liner")},
		{"Feature Kiro", 10, has("feature_slug")},
	}
}

// readinessScore sums the weights of completed components (0–100).
func readinessScore(comps []scoreComponent) int {
	score := 0
	for _, c := range comps {
		if c.Done {
			score += c.Weight
		}
	}
	return score
}

// stepGlyph returns the state marker for a wizard step in the progression list.
func stepGlyph(i, active int, st theme.Styles) string {
	switch {
	case i < active:
		return st.Success.Render("✓")
	case i == active:
		return st.PanelTitle.Render("▶")
	default:
		return st.Muted.Render("○")
	}
}

// renderProgressBar draws a clean filled bar with a fraction suffix.
func renderProgressBar(st theme.Styles, done, total, barW int) string {
	if total <= 0 {
		total = 1
	}
	if barW < 4 {
		barW = 4
	}
	filled := done * barW / total
	if filled > barW {
		filled = barW
	}
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(st.Theme.Palette.Primary)).Render(strings.Repeat("█", filled)) +
		st.Divider.Render(strings.Repeat("░", barW-filled))
	return bar
}

// renderProgressionPanel replaces the ring gauge: progress bar + vertical step
// list + next step. It answers "where am I / what's left / what's next".
func renderProgressionPanel(st theme.Styles, vm ViewModel, w int) string {
	order := onbdomain.TUIStepOrder
	active := stepIndex(vm.Model.Step)
	total := len(order)

	barW := w - 6
	if barW < 6 {
		barW = 6
	}
	var b strings.Builder
	b.WriteString(renderProgressBar(st, active+1, total, barW))
	b.WriteString("  " + st.Bold.Render(fracString(active+1, total)) + "\n\n")

	for i, s := range order {
		label := StepLabel(s)
		styled := st.Muted.Render(label)
		switch {
		case i < active:
			styled = st.Fg.Render(label)
		case i == active:
			styled = st.PanelTitle.Render(label)
		}
		b.WriteString(stepGlyph(i, active, st) + "  " + styled + "\n")
	}
	b.WriteString("\n" + st.Muted.Render("Prochaine : ") + st.Fg.Render(nextStepTitle(vm.Model.Step)))
	return b.String()
}

// renderReadinessSection shows the explainable readiness score and breakdown.
func renderReadinessSection(st theme.Styles, vm ViewModel, w int) string {
	comps := readinessComponents(vm)
	score := readinessScore(comps)
	if vm.Model.Applied && vm.Model.Readiness.Score > 0 {
		score = vm.Model.Readiness.Score
	}

	barW := w - 8
	if barW < 6 {
		barW = 6
	}
	var b strings.Builder
	b.WriteString(renderProgressBar(st, score, 100, barW) + " " + st.Bold.Render(pctString(score)) + "\n\n")
	for _, c := range comps {
		mark := st.Muted.Render("○")
		label := st.Muted.Render(c.Label)
		if c.Done {
			mark = st.Success.Render("✓")
			label = st.Fg.Render(c.Label)
		}
		b.WriteString(mark + " " + label + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// activityEvent is one entry in the Activity feed, derived from real signals.
type activityEvent struct {
	done bool
	text string
}

func activityEvents(vm ViewModel) []activityEvent {
	m := vm.Model
	has := func(key string) bool { return strings.TrimSpace(m.fieldValue(key)) != "" }
	var ev []activityEvent
	ev = append(ev, activityEvent{true, "Dépôt détecté"})
	if len(m.DetectedStacks) > 0 {
		ev = append(ev, activityEvent{true, "Stack détectée : " + strings.Join(m.DetectedStacks, ", ")})
	}
	if len(m.ValidationPreview) > 0 {
		names := make([]string, 0, len(m.ValidationPreview))
		for _, v := range m.ValidationPreview {
			names = append(names, validationName(v))
		}
		ev = append(ev, activityEvent{true, "Validations QA : " + strings.Join(names, ", ")})
	}
	if has("product_one_liner") {
		ev = append(ev, activityEvent{true, "Documentation produit renseignée"})
	}
	if has("default_agent") || has("default_reviewer") {
		ev = append(ev, activityEvent{true, "Agents configurés"})
	}
	// Current in-progress action: the next thing Asagiri will do.
	ev = append(ev, activityEvent{false, strings.TrimSuffix(nextStepDesc(m.Step), ".")})
	return ev
}

// renderActivitySection renders a borderless Activity list (no nested frame),
// truncated to maxRows and maxW.
func renderActivitySection(st theme.Styles, vm ViewModel, maxRows, maxW int) string {
	events := activityEvents(vm)
	if maxRows > 1 && len(events) > maxRows {
		// Keep the most recent done events plus the in-progress one.
		events = append(events[:maxRows-1], events[len(events)-1])
	}
	textW := maxW - 2
	if textW < 8 {
		textW = 8
	}
	var lines []string
	for _, e := range events {
		text := e.text
		if !e.done {
			text += "…"
		}
		text = truncateRunes(text, textW)
		if e.done {
			lines = append(lines, st.Success.Render("✓")+" "+st.Fg.Render(text))
		} else {
			lines = append(lines, st.PanelTitle.Render("⠿")+" "+st.Muted.Render(text))
		}
	}
	return strings.Join(lines, "\n")
}

