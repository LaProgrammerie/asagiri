package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

const eosLabelColW = 16

// renderStepperVisual draws numbered circular nodes connected by lines, with a
// label row underneath. Falls back to a compact dotted line when too narrow.
func renderStepperVisual(step onbdomain.WizardStep, st theme.Styles, w int) string {
	active := stepIndex(step)
	labels := wizardStepLabels()
	n := len(labels)
	if n == 0 {
		return ""
	}
	cell := w / n
	if cell < 5 {
		return renderStepperCompact(active, st, w)
	}

	var nodeRow, labelRow strings.Builder
	for i, label := range labels {
		nodeRow.WriteString(stepCell(i, active, n, cell, st))

		lb := label
		if lipglossWidth(lb) > cell-1 {
			lb = truncateRunes(lb, cell-1)
		}
		var styled string
		switch {
		case i == active:
			styled = st.PanelTitle.Render(lb)
		case i < active:
			styled = st.Success.Render(lb)
		default:
			styled = st.Muted.Render(lb)
		}
		labelRow.WriteString(padCenter(styled, cell))
	}
	return strings.TrimRight(nodeRow.String(), " ") + "\n" + strings.TrimRight(labelRow.String(), " ")
}

func stepCell(i, active, n, cell int, st theme.Styles) string {
	node := stepNode(i, active, st)
	nodeW := lipglossWidth(node)
	rem := cell - nodeW
	if rem < 0 {
		rem = 0
	}
	leftN := rem / 2
	rightN := rem - leftN

	leftSeg := strings.Repeat("─", leftN)
	rightSeg := strings.Repeat("─", rightN)
	if i == 0 {
		leftSeg = strings.Repeat(" ", leftN)
	}
	if i == n-1 {
		rightSeg = strings.Repeat(" ", rightN)
	}
	left := colorConnector(leftSeg, i <= active, st)
	right := colorConnector(rightSeg, i < active, st)
	return left + node + right
}

func colorConnector(seg string, done bool, st theme.Styles) string {
	if strings.TrimSpace(seg) == "" {
		return seg
	}
	if done {
		return st.Success.Render(seg)
	}
	return st.Muted.Render(seg)
}

var (
	stepOutlineGlyphs = []rune("①②③④⑤⑥⑦⑧⑨")
	stepFilledGlyphs  = []rune("❶❷❸❹❺❻❼❽❾")
)

// stepNode renders one circular step marker using circled-digit glyphs.
func stepNode(i, active int, st theme.Styles) string {
	switch {
	case i < active:
		return st.StepDone.Render("✓")
	case i == active:
		if i < len(stepFilledGlyphs) {
			return st.StepActiveGlyph.Render(string(stepFilledGlyphs[i]))
		}
		return st.StepActiveGlyph.Render(fmt.Sprintf("%d", i+1))
	default:
		if i < len(stepOutlineGlyphs) {
			return st.StepPending.Render(string(stepOutlineGlyphs[i]))
		}
		return st.StepPending.Render(fmt.Sprintf("%d", i+1))
	}
}

func padCenter(s string, w int) string {
	sw := lipglossWidth(s)
	if sw >= w {
		return s
	}
	pad := w - sw
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

func lipglossWidth(s string) int { return lipgloss.Width(s) }

func truncateRunes(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(r[:maxLen])
	}
	return string(r[:maxLen-1]) + "…"
}

// renderWelcomePanel renders the welcome step as a single flat card with
// borderless sections (no nested frames).
func renderWelcomePanel(vm ViewModel, st theme.Styles, boxW int) string {
	// Card border (2) + horizontal padding (2*2) = 6 cells consumed.
	innerW := boxW - 6
	if innerW < 20 {
		innerW = 20
	}
	center := func(s string) string { return lipgloss.PlaceHorizontal(innerW, lipgloss.Center, s) }

	var b strings.Builder
	b.WriteString(center(st.HeroTitle.Render("Bienvenue dans Asagiri")) + "\n")
	b.WriteString(center(st.Muted.Render("Configurez votre projet et générez votre première spec Kiro.")) + "\n\n")

	b.WriteString(st.SectionHead.Render("CE QUE NOUS ALLONS FAIRE") + "\n")
	for _, item := range welcomeChecklist() {
		b.WriteString(st.Success.Render("✓") + "  " + st.Fg.Render(item) + "\n")
	}
	b.WriteString("\n" + st.PanelTitle.Render("◆ Conseil") + "\n")
	b.WriteString(st.Muted.Render("Naviguez avec les flèches ou ouvrez la palette avec « a »."))

	return st.Card.Width(boxW).Render(b.String())
}

func renderStepPanel(vm ViewModel, st theme.Styles, boxW int) string {
	m := vm.Model
	var rows []string
	rows = append(rows, st.HeroTitle.Render(StepLabel(m.Step)), "")

	for i, row := range m.fieldRows {
		focused := m.FocusFooter < 0 && i == m.FocusField
		label := rowLabel(row)
		icon := fieldIconForKey(row.Key, label)
		pill := row.Key == "stack" || row.Key == "detected_stacks" || row.ReadOnly
		value := m.fieldValue(row.Key)
		isValidation := strings.HasPrefix(row.Key, "validation_")
		if isValidation {
			value = validationName(value)
		}
		rows = append(rows, st.RenderEOSFieldGrid(icon, label, value, focused, pill, eosLabelColW))
		if errMsg := m.Errors[row.Key]; errMsg != "" {
			rows = append(rows, "     "+st.Error.Render("! "+errMsg))
		}
		if row.Key == "stack" && !row.ReadOnly {
			rows = append(rows, "     "+st.Muted.Render("(go | castor | node | auto)"))
		}
		// Group consecutive read-only validation rows tightly; blank line otherwise.
		if !isValidation {
			rows = append(rows, "")
		}
	}
	if len(m.SkippedFields) > 0 {
		rows = append(rows, st.Warning.Render("Conservé: "+strings.Join(m.SkippedFields, ", ")))
	}
	if m.Message != "" {
		rows = append(rows, st.Success.Render(m.Message))
	}
	body := strings.TrimRight(strings.Join(rows, "\n"), "\n")
	return st.Card.Width(boxW).Render(body)
}

func fieldIconForKey(key, _ string) string {
	switch {
	case key == "stack" || key == "detected_stacks":
		return "⚙"
	case strings.HasPrefix(key, "validation_"):
		return "✓"
	case key == "project_name":
		return "◫"
	case key == "default_branch":
		return "⎇"
	case key == "tagline":
		return "✎"
	case key == "default_agent", key == "default_reviewer":
		return "◎"
	default:
		return "·"
	}
}

func rowLabel(row FieldDef) string {
	switch row.Key {
	case "stack":
		return "Stack"
	case "detected_stacks":
		return "Détecté"
	case "project_name":
		return "Nom"
	case "default_branch":
		return "Branche"
	case "tagline":
		return "Tagline"
	case "default_agent":
		return "Agent"
	case "default_reviewer":
		return "Reviewer"
	case "product_one_liner":
		return "Produit"
	case "feature_slug":
		return "Feature"
	default:
		if strings.HasPrefix(row.Key, "validation_") {
			return "Validation"
		}
	}
	return row.Label
}

