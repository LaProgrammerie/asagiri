package onboarding

import (
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

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
	if m.Step == onbdomain.StepDocs {
		rows = append(rows, st.Muted.Render("Phrase pour docs/ai/01-product.md — pas le nom saisi à l’étape Projet."))
		rows = append(rows, "")
	}
	if m.Step == onbdomain.StepFeature {
		rows = append(rows, st.Muted.Render("Première spec à développer — crée .kiro/specs/<slug>/ (complétable ensuite)."))
		rows = append(rows, "")
	}

	for i, row := range m.fieldRows {
		focused := m.FocusFooter < 0 && i == m.FocusField
		label := rowLabel(row)
		icon := fieldIconForKey(row.Key, label)
		managed := row.Kind == FieldManaged
		pill := false
		if row.Kind == FieldSelect {
			pill = focused // badge only while editing; otherwise plain text
		}
		value := m.fieldValue(row.Key)
		isValidation := strings.HasPrefix(row.Key, "validation_")
		if isValidation {
			value = validationName(value)
		}
		if row.Kind == FieldSelect {
			value = formatSelectDisplay(value, row.Choices, focused)
		}
		rows = append(rows, st.RenderEOSFieldGrid(icon, label, value, focused, pill, managed))
		if errMsg := m.Errors[row.Key]; errMsg != "" {
			rows = append(rows, "     "+st.Error.Render("! "+errMsg))
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
	case key == "default_spec_agent", key == "default_enricher", key == "default_agent", key == "default_reviewer":
		return "◎"
	case key == "pipeline_plan", key == "pipeline_verify":
		return "◇"
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
	case "default_spec_agent":
		return "Spec"
	case "default_enricher":
		return "Enrich"
	case "default_agent":
		return "Dev"
	case "default_reviewer":
		return "Review"
	case "pipeline_plan":
		return "Plan"
	case "pipeline_verify":
		return "Verify"
	case "product_one_liner":
		return "Phrase produit"
	case "feature_slug":
		return "Slug"
	default:
		if strings.HasPrefix(row.Key, "validation_") {
			return "Validation"
		}
	}
	return row.Label
}

func formatSelectDisplay(current string, choices []string, focused bool) string {
	current = strings.TrimSpace(current)
	if current == "" && len(choices) > 0 {
		current = choices[0]
	}
	if focused {
		return "‹ " + current + " ›"
	}
	if len(choices) > 1 {
		return current + " ▾"
	}
	return current
}

