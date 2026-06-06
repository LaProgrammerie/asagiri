package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ViewModel is the data contract for the onboarding screen.
type ViewModel struct {
	Model           Model
	Readiness       bus.ReadinessResult
	Shell           ShellContext
	ShowCLI         bool
	WizardMode      bool
	InAppShell      bool // true when rendered inside app.View() (FR-5.1)
	FullScreen      bool
	Width           int
	Height          int
	Theme           theme.Theme
}

// Render returns the interactive onboarding wizard view. The fullscreen wizard
// is routed through the shared cockpit shell (components.*); narrower terminals
// and non-wizard contexts fall back to the flat layout.
func Render(vm ViewModel) string {
	if vm.InAppShell {
		return renderInAppShell(vm)
	}
	if vm.FullScreen && vm.WizardMode && vm.Width >= 90 && vm.Height >= 24 {
		return renderShell(vm)
	}
	if vm.Model.Applied {
		return renderReadySummary(vm)
	}
	return renderWizard(vm)
}

func innerContentWidth(termW int, st theme.Styles) int {
	outer := st.Theme.BorderStyle().Padding(1, 2)
	w := termW - outer.GetHorizontalBorderSize() - outer.GetHorizontalPadding()
	if w < 40 {
		return 40
	}
	return w
}

func outerBoxWidth(termW int, st theme.Styles) int {
	outer := st.Theme.BorderStyle().Padding(1, 2)
	w := termW - outer.GetHorizontalBorderSize()
	if w < 44 {
		return 44
	}
	return w
}

func renderWizard(vm ViewModel) string {
	m := vm.Model
	st := vm.Theme.Styles()
	innerW := innerContentWidth(vm.Width, st)

	meta := fmt.Sprintf("Étape %s · %s", StepLabel(m.Step), StepProgress(m.Step))
	header := st.RenderPageHeader("Project Onboarding Wizard", meta)
	tabs := st.RenderTabBar(wizardTabLabels(), stepIndex(m.Step))
	content := st.ContentArea.Width(innerW).Render(renderWizardPanelBody(vm, st))
	footer := renderFooterStyled(m, st)

	var hints strings.Builder
	hints.WriteString(st.Hint.Render("Tab/↑↓ champs · ←→ boutons · Ctrl+P/N étapes · Ctrl+A advanced · Ctrl+S apply · Ctrl+Q quitter"))
	if vm.ShowCLI {
		hints.WriteString("\n")
		hints.WriteString(st.Muted.Render("CLI: asa onboard --yes | asa ready --json"))
	}

	body := joinBlocks(header, tabs, content, footer, hints.String())
	if !vm.FullScreen {
		return body + "\n"
	}
	return wrapFullscreen(body, StepLabel(m.Step), st, vm.Width)
}

func renderWizardPanelBody(vm ViewModel, st theme.Styles) string {
	m := vm.Model
	var body strings.Builder

	switch m.Step {
	case onbdomain.StepWelcome:
		body.WriteString(st.PanelTitle.Render(StepLabel(m.Step)) + "\n\n")
		if m.ExistingConfig {
			body.WriteString(st.Success.Render("✓ Configuration existante détectée — mode vérification") + "\n")
		} else {
			body.WriteString(st.Fg.Render("Asagiri prépare config, validation, docs et première spec Kiro.") + "\n")
			body.WriteString(st.Muted.Render("L'onboarding ne remplace pas la spec produit ni le handoff.") + "\n")
		}
	case onbdomain.StepProject:
		for _, line := range renderProjectDetection(m, st) {
			body.WriteString(line + "\n")
		}
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
			if errMsg := m.Errors[row.Key]; errMsg != "" {
				body.WriteString(st.Error.Render("    ! "+errMsg) + "\n")
			}
		}
	case onbdomain.StepStack:
		for _, line := range renderStackDetection(m, st) {
			body.WriteString(line + "\n")
		}
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
		}
	case onbdomain.StepProviders:
		body.WriteString(st.Muted.Render("Runtimes externes — l'adapter est choisi par providers.<id>.type, pas par le nom.") + "\n\n")
		for _, line := range renderProviderCatalog(m, st) {
			body.WriteString(line + "\n")
		}
		body.WriteString("\n")
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
		}
	case onbdomain.StepAgents:
		for _, line := range renderAgentPipeline(m, st) {
			body.WriteString(line + "\n")
		}
		body.WriteString("\n")
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
		}
	case onbdomain.StepDocs:
		body.WriteString(st.Muted.Render("Cette phrase sera écrite dans docs/ai/01-product.md — pas le nom saisi à l'étape Projet.") + "\n\n")
		body.WriteString(st.SectionHead.Render("FICHIERS QUI SERONT CRÉÉS") + "\n")
		for _, f := range docsAIFiles() {
			body.WriteString(st.Success.Render("✓") + " " + st.Muted.Render(f) + "\n")
		}
		body.WriteString("\n")
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
		}
	case onbdomain.StepFeature:
		slug := strings.TrimSpace(m.Fields["feature_slug"])
		if slug != "" {
			body.WriteString(st.SectionHead.Render("SPEC QUI SERA CRÉÉE") + "\n")
			body.WriteString(st.Success.Render("✓") + " " + st.Muted.Render(".kiro/specs/"+slug+"/") + "\n")
			body.WriteString(st.Success.Render("✓") + " " + st.Muted.Render(".kiro/specs/"+slug+"/requirements.md") + "\n\n")
			body.WriteString(st.SectionHead.Render("WORKFLOW SUIVANT") + "\n")
			for _, step := range nextWorkflowSteps(slug) {
				body.WriteString(st.Muted.Render("  $ "+step) + "\n")
			}
			body.WriteString("\n")
		}
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
			if errMsg := m.Errors[row.Key]; errMsg != "" {
				body.WriteString(st.Error.Render("    ! "+errMsg) + "\n")
			}
		}
		if slug != "" {
			body.WriteString(st.Muted.Render("  → .kiro/specs/"+slug+"/") + "\n")
		}
	default:
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
			if errMsg := m.Errors[row.Key]; errMsg != "" {
				body.WriteString(st.Error.Render("    ! "+errMsg) + "\n")
			}
		}
		if m.Step == onbdomain.StepReview {
			for _, line := range renderArtefactsSection(m, st) {
				body.WriteString(line + "\n")
			}
		}
	}

	out := body.String()
	if len(m.SkippedFields) > 0 {
		out += "\n" + st.Warning.Render("Conservé (config custom): "+strings.Join(m.SkippedFields, ", "))
	}
	if m.Message != "" {
		out += "\n" + st.Success.Render(m.Message)
	}
	return out
}

func renderFooterStyled(m Model, st theme.Styles) string {
	type btn struct {
		id    int
		label string
	}
	buttons := []btn{
		{FooterPrev, "← Prev"},
		{FooterNext, "Next →"},
		{FooterAdvanced, "Advanced"},
	}
	if m.canApply() {
		buttons = append(buttons, btn{FooterApply, "Apply"})
	}
	parts := make([]string, 0, len(buttons))
	for _, b := range buttons {
		label := b.label
		if b.id == FooterAdvanced {
			label = advancedButtonLabel(m)
		}
		parts = append(parts, st.RenderButton(label, m.FocusFooter == b.id))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderReadySummary(vm ViewModel) string {
	r := vm.Readiness
	if vm.Model.Readiness.Score > 0 || len(vm.Model.Readiness.Checks) > 0 {
		r = vm.Model.Readiness
	}
	st := vm.Theme.Styles()
	innerW := innerContentWidth(vm.Width, st)

	statusLabel := "NOT READY"
	statusStyle := st.Error
	if r.Ready {
		statusLabel = "READY"
		statusStyle = st.Success
	}

	var body strings.Builder
	body.WriteString(statusStyle.Render(statusLabel) + "  " + st.RenderProgress(r.Score, 100, progressBarWidth(innerW)) + "\n\n")

	if len(r.Checks) > 0 {
		body.WriteString(st.PanelTitle.Render("Checks") + "\n")
		for _, c := range r.Checks {
			body.WriteString(st.RenderCheckStatus(c.Status, c.ID, c.Message) + "\n")
		}
	}
	if len(r.NextActions) > 0 {
		body.WriteString("\n" + st.PanelTitle.Render("Next actions") + "\n")
		for _, a := range r.NextActions {
			body.WriteString(st.Muted.Render("• "+a.Title) + "\n")
			if vm.ShowCLI {
				body.WriteString(st.Hint.Render("  CLI: "+a.CLI) + "\n")
			}
		}
	}
	if len(r.AutofixOffers) > 0 && !r.Ready {
		body.WriteString("\n" + st.AccentBlock.Render("Corrections automatiques disponibles") + "\n")
		for _, o := range r.AutofixOffers {
			line := "• " + o.Title
			if len(o.Lines) > 0 {
				line += " (" + strings.Join(o.Lines, ", ") + ")"
			}
			body.WriteString(st.Fg.Render(line) + "\n")
		}
		body.WriteString("\n" + st.RenderButton("O · appliquer", true) + st.RenderButton("N · ignorer", false) + "\n")
	}
	if vm.Model.Message != "" {
		body.WriteString("\n" + st.Fg.Render(vm.Model.Message) + "\n")
	}

	content := joinBlocks(
		st.RenderPageHeader("Onboarding terminé", "Readiness du dépôt"),
		st.ContentArea.Width(innerW).Render(body.String()),
	)

	if !vm.FullScreen {
		return content + "\n"
	}
	return wrapFullscreen(content, "Récap", st, vm.Width)
}

func wrapFullscreen(body, stepLabel string, st theme.Styles, termW int) string {
	statusBar := st.RenderStatusBar("WIZARD", stepLabel, "")
	column := joinBlocks(body, statusBar)
	boxW := outerBoxWidth(termW, st)
	return st.Theme.BorderStyle().
		Padding(1, 2).
		Width(boxW).
		Background(lipgloss.Color(st.Theme.Palette.Background)).
		Render(column)
}

func joinBlocks(blocks ...string) string {
	parts := make([]string, 0, len(blocks)*2)
	for _, b := range blocks {
		b = strings.TrimRight(b, "\n")
		if b == "" {
			continue
		}
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, b)
	}
	return strings.Join(parts, "\n")
}

func progressBarWidth(innerW int) int {
	w := innerW / 3
	if w < 16 {
		return 16
	}
	if w > 40 {
		return 40
	}
	return w
}

func wizardTabLabels() []string {
	labels := make([]string, 0, len(onbdomain.TUIStepOrder))
	for _, step := range onbdomain.TUIStepOrder {
		labels = append(labels, tabShortLabel(step))
	}
	return labels
}

func tabShortLabel(step onbdomain.WizardStep) string {
	switch step {
	case onbdomain.StepWelcome:
		return "Start"
	case onbdomain.StepProject:
		return "Project"
	case onbdomain.StepStack:
		return "Stack"
	case onbdomain.StepProviders:
		return "Providers"
	case onbdomain.StepAgents:
		return "Agents"
	case onbdomain.StepDocs:
		return "Docs"
	case onbdomain.StepFeature:
		return firstFeatureStepLabel()
	case onbdomain.StepReview:
		return "Review"
	default:
		return string(step)
	}
}

func stepIndex(step onbdomain.WizardStep) int {
	for i, s := range onbdomain.TUIStepOrder {
		if s == step {
			return i
		}
	}
	return 0
}

// RenderLegacyReadiness keeps the read-only readiness panel for non-wizard contexts.
func RenderLegacyReadiness(vm ViewModel) string {
	st := vm.Theme.Styles()
	var b strings.Builder
	b.WriteString(st.RenderPageHeader("Project Onboarding", "") + "\n\n")
	status := "NOT READY"
	statusStyle := st.Error
	if vm.Readiness.Ready {
		status = "READY"
		statusStyle = st.Success
	}
	b.WriteString(statusStyle.Render(status) + "  " + st.RenderProgress(vm.Readiness.Score, 100, 20) + "\n\n")
	if len(vm.Readiness.Checks) > 0 {
		for _, c := range vm.Readiness.Checks {
			b.WriteString(st.RenderCheckStatus(c.Status, c.ID, c.Message) + "\n")
		}
	}
	if len(vm.Readiness.NextActions) > 0 {
		b.WriteString("\n" + st.PanelTitle.Render("Next actions") + "\n")
		for _, a := range vm.Readiness.NextActions {
			b.WriteString(st.Muted.Render("• "+a.Title) + "\n")
			if vm.ShowCLI {
				b.WriteString(st.Hint.Render("  CLI: "+a.CLI) + "\n")
			}
		}
	}
	b.WriteString("\n" + st.Hint.Render("Lancer le wizard: asa onboard --ui"))
	return strings.TrimRight(b.String(), "\n") + "\n"
}
