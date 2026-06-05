package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// renderWelcomePanel renders the welcome step with detection context.
func renderWelcomePanel(vm ViewModel, st theme.Styles, boxW int) string {
	innerW := boxW - 6
	if innerW < 20 {
		innerW = 20
	}
	center := func(s string) string { return lipgloss.PlaceHorizontal(innerW, lipgloss.Center, s) }
	m := vm.Model

	var b strings.Builder
	b.WriteString(center(st.HeroTitle.Render("Bienvenue dans Asagiri")) + "\n")

	if m.ExistingConfig {
		b.WriteString(center(st.Success.Render("✓ Configuration existante détectée — mode vérification et complétion")) + "\n\n")
	} else {
		b.WriteString(center(st.Muted.Render("Configurez votre projet et générez votre première spec Kiro.")) + "\n\n")
	}

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

	switch m.Step {
	case onbdomain.StepProject:
		rows = append(rows, renderProjectDetection(m, st)...)

	case onbdomain.StepStack:
		rows = append(rows, renderStackDetection(m, st)...)

	case onbdomain.StepAgents:
		rows = append(rows, renderAgentPipeline(m, st)...)
		rows = append(rows, "")

	case onbdomain.StepDocs:
		rows = append(rows, st.Muted.Render("Cette phrase sera écrite dans docs/ai/01-product.md — pas le nom saisi à l'étape Projet."))
		rows = append(rows, "")
		rows = append(rows, st.SectionHead.Render("FICHIERS QUI SERONT CRÉÉS"))
		for _, f := range docsAIFiles() {
			rows = append(rows, st.Success.Render("✓")+" "+st.Muted.Render(f))
		}
		rows = append(rows, "")

	case onbdomain.StepFeature:
		slug := strings.TrimSpace(m.Fields["feature_slug"])
		if slug != "" {
			rows = append(rows, st.SectionHead.Render("SPEC QUI SERA CRÉÉE"))
			rows = append(rows, st.Success.Render("✓")+" "+st.Muted.Render(".kiro/specs/"+slug+"/"))
			rows = append(rows, st.Success.Render("✓")+" "+st.Muted.Render(".kiro/specs/"+slug+"/requirements.md"))
			rows = append(rows, "")
			rows = append(rows, st.SectionHead.Render("WORKFLOW SUIVANT"))
			for _, step := range nextWorkflowSteps(slug) {
				rows = append(rows, st.Muted.Render("  $ "+step))
			}
		} else {
			rows = append(rows, st.Muted.Render("Première spec à développer — crée .kiro/specs/<slug>/"))
		}
		rows = append(rows, "")
	}

	// Field rows
	for i, row := range m.fieldRows {
		focused := m.FocusFooter < 0 && i == m.FocusField
		label := rowLabel(row)
		icon := fieldIconForKey(row.Key, label)
		managed := row.Kind == FieldManaged
		pill := row.Kind == FieldSelect && focused
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
		if !isValidation {
			rows = append(rows, "")
		}
	}

	// Feature slug: live path preview
	if m.Step == onbdomain.StepFeature {
		if slug := strings.TrimSpace(m.Fields["feature_slug"]); slug != "" {
			rows = append(rows, st.Muted.Render("  → .kiro/specs/"+slug+"/"))
		}
	}

	// Review step: artefacts preview
	if m.Step == onbdomain.StepReview {
		rows = append(rows, renderArtefactsSection(m, st)...)
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

// renderProjectDetection renders the "Détecté" block above project fields.
func renderProjectDetection(m Model, st theme.Styles) []string {
	var rows []string
	rows = append(rows, st.SectionHead.Render("DÉTECTÉ"))

	branch := fallbackStr(m.Fields["default_branch"], "main")
	rows = append(rows, st.Success.Render("✓")+" "+st.Fg.Render("Dépôt Git"))
	rows = append(rows, st.Success.Render("✓")+" "+st.Fg.Render("Branche : ")+st.Muted.Render(branch))

	if len(m.DetectedStacks) > 0 {
		rows = append(rows, st.Success.Render("✓")+" "+st.Fg.Render("Langages : ")+st.Muted.Render(strings.Join(m.DetectedStacks, ", ")))
	}
	if m.ExistingConfig {
		rows = append(rows, st.Success.Render("✓")+" "+st.Fg.Render("Configuration Asagiri existante"))
	}
	rows = append(rows, "")
	return rows
}

// renderStackDetection renders validation commands and activated capabilities.
func renderStackDetection(m Model, st theme.Styles) []string {
	var rows []string

	if len(m.ValidationPreview) > 0 {
		rows = append(rows, st.SectionHead.Render("VALIDATIONS ACTIVÉES"))
		for _, v := range m.ValidationPreview {
			rows = append(rows, st.Success.Render("✓")+" "+st.Muted.Render(validationName(v)))
		}
		rows = append(rows, "")
	}

	capabilities := stackCapabilities(m.Fields["stack"])
	if len(capabilities) > 0 {
		rows = append(rows, st.SectionHead.Render("CAPACITÉS ACTIVÉES"))
		for _, cap := range capabilities {
			rows = append(rows, st.RenderBadgeMuted(cap))
		}
		rows = append(rows, "")
	}

	return rows
}

// renderAgentPipeline renders the full pipeline with arrows and local/cloud indicators.
func renderAgentPipeline(m Model, st theme.Styles) []string {
	type step struct {
		label string
		key   string
		fixed bool // managed by Asagiri (not configurable)
	}
	pipeline := []step{
		{"Spec", "default_spec_agent", false},
		{"Plan", "", true},
		{"Enrich", "default_enricher", false},
		{"Dev", "default_agent", false},
		{"Verify", "", true},
		{"Review", "default_reviewer", false},
	}

	var rows []string
	rows = append(rows, st.SectionHead.Render("PIPELINE GÉNÉRÉ"))

	activeAgents := 0
	for i, s := range pipeline {
		var agent, indicator string
		if s.fixed {
			agent = "Asagiri"
			indicator = st.Muted.Render("(local)")
		} else {
			agent = fallbackStr(m.Fields[s.key], s.key)
			if isLocalAgentName(agent) {
				indicator = st.Muted.Render("(local)")
			} else {
				indicator = st.Muted.Render("(cloud)")
				activeAgents++
			}
		}
		arrow := "  →  "
		if i == 0 {
			arrow = "     "
		}
		rows = append(rows,
			st.Muted.Render(arrow)+
				st.Fg.Render(fmt.Sprintf("%-8s", s.label))+"  "+
				st.Success.Render(agent)+"  "+indicator,
		)
	}

	if activeAgents > 0 {
		rows = append(rows, "")
		rows = append(rows, st.Muted.Render(fmt.Sprintf("  %d agent(s) cloud actif(s)", activeAgents)))
	}

	return rows
}

// renderArtefactsSection renders the "Ce qui va être créé" block on the Review step.
func renderArtefactsSection(m Model, st theme.Styles) []string {
	var rows []string
	rows = append(rows, "")
	rows = append(rows, st.SectionHead.Render("CE QUI VA ÊTRE CRÉÉ"))

	artefacts := m.ArtefactsPreview
	if len(artefacts) == 0 {
		artefacts = defaultArtefactsPreview(m.Fields["feature_slug"])
	}
	for _, a := range artefacts {
		rows = append(rows, st.Success.Render("✓")+" "+st.Muted.Render(a))
	}
	rows = append(rows, "")
	rows = append(rows, st.Muted.Render("Aucune modification destructive — fichiers existants préservés."))
	return rows
}

// isLocalAgentName reports whether an agent name suggests a local backend.
func isLocalAgentName(name string) bool {
	switch strings.ToLower(name) {
	case config.DefaultAgentEnrich, "local":  // DefaultAgentEnrich == "ollama"
		return true
	}
	return false
}

// stackCapabilities returns the workflow capabilities unlocked by a stack.
func stackCapabilities(stack string) []string {
	switch strings.ToLower(strings.TrimSpace(stack)) {
	case "go":
		return []string{"spec", "plan", "enrich", "dev", "verify", "review", "pr"}
	case "castor", "php":
		return []string{"spec", "plan", "enrich", "dev", "verify", "review"}
	case "node", "typescript":
		return []string{"spec", "plan", "enrich", "dev", "verify", "review"}
	default:
		return []string{"spec", "plan", "dev", "verify"}
	}
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

// docsAIFiles returns the list of docs/ai/ files that will be created on Apply.
func docsAIFiles() []string {
	return []string{
		"docs/ai/01-product.md",
		"docs/ai/02-architecture.md",
		"docs/ai/03-standards.md",
		"docs/ai/04-workflows.md",
		"docs/ai/context-map.md",
	}
}

// nextWorkflowSteps returns the CLI commands the user will run after onboarding.
func nextWorkflowSteps(slug string) []string {
	return []string{
		"asa work \"develop " + slug + "\"",
		"asa next",
		"asa status",
	}
}
