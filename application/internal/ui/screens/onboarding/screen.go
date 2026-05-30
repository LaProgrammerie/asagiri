package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

const (
	outerBorderCols = 2
	outerPaddingCols = 4
	panelBorderCols = 2
	panelPaddingCols = 2
)

// frameLayout holds lipgloss-safe widths (content width inside each box).
type frameLayout struct {
	InnerW int // inside outer frame, before its border+padding
	PanelW int // inside inner panel, before its border+padding
}

func computeFrameLayout(width int) frameLayout {
	w := width - outerBorderCols - outerPaddingCols
	if w < 40 {
		w = 40
	}
	pw := w - panelBorderCols - panelPaddingCols
	if pw < 36 {
		pw = 36
	}
	return frameLayout{InnerW: w, PanelW: pw}
}

// ViewModel is the data contract for the onboarding screen.
type ViewModel struct {
	Model           Model
	Readiness       bus.ReadinessResult
	ShowCLI         bool
	WizardMode      bool
	FullScreen      bool
	Width           int
	Height          int
	Theme           theme.Theme
}

// Render returns the interactive onboarding wizard view.
func Render(vm ViewModel) string {
	if vm.Model.Applied {
		return renderReadySummary(vm)
	}
	return renderWizard(vm)
}

func renderWizard(vm ViewModel) string {
	m := vm.Model
	st := vm.Theme.Styles()
	layout := computeFrameLayout(vm.Width)

	meta := fmt.Sprintf("Étape %s · %s", StepLabel(m.Step), StepProgress(m.Step))
	header := st.RenderPageHeader("Project Onboarding Wizard", meta)
	tabs := st.RenderTabBar(wizardTabLabels(), stepIndex(m.Step))
	contentPanel := st.Panel.Width(layout.PanelW).Render(renderWizardPanelBody(vm, st))
	footer := renderFooterStyled(m, st)

	var hints strings.Builder
	hints.WriteString(st.Hint.Render("Tab/↑↓ champs · ←→ boutons · Ctrl+P/N étapes · Ctrl+A advanced · Ctrl+S apply"))
	if vm.ShowCLI {
		hints.WriteString("\n")
		hints.WriteString(st.Muted.Render("CLI: asa onboard --yes | asa ready --json"))
	}

	body := joinBlocks(
		header,
		tabs,
		contentPanel,
		footer,
		hints.String(),
	)

	if !vm.FullScreen {
		return body + "\n"
	}
	return wrapFullscreen(body, StepLabel(m.Step), "Ctrl+P/N · étapes · Ctrl+S · apply · Ctrl+Q · quitter", st, layout)
}

func renderWizardPanelBody(vm ViewModel, st theme.Styles) string {
	m := vm.Model
	var body strings.Builder
	body.WriteString(st.PanelTitle.Render(StepLabel(m.Step)) + "\n\n")

	switch m.Step {
	case onbdomain.StepWelcome:
		body.WriteString(st.Fg.Render("Asagiri prépare config, validation, docs et première spec Kiro.") + "\n")
		body.WriteString(st.Muted.Render("L'onboarding ne remplace pas la spec produit ni le handoff.") + "\n")
	default:
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			body.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
			if errMsg := m.Errors[row.Key]; errMsg != "" {
				body.WriteString(st.Error.Render("    ! "+errMsg) + "\n")
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
		if b.id == FooterAdvanced && m.ShowAdvanced {
			label = "Advanced ▼"
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
	layout := computeFrameLayout(vm.Width)

	statusLabel := "NOT READY"
	statusStyle := st.Error
	if r.Ready {
		statusLabel = "READY"
		statusStyle = st.Success
	}

	var body strings.Builder
	body.WriteString(statusStyle.Render(statusLabel) + "  " + st.RenderProgress(r.Score, 100, progressBarWidth(layout.InnerW)) + "\n\n")

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
		st.Panel.Width(layout.PanelW).Render(body.String()),
	)

	footerHint := "Ctrl+Q · quitter"
	if len(r.AutofixOffers) > 0 && !r.Ready {
		footerHint = "O · appliquer   N · ignorer   Ctrl+F · appliquer   Ctrl+Q · quitter"
	}

	if !vm.FullScreen {
		return content + "\n"
	}
	return wrapFullscreen(content, "Récap", footerHint, st, layout)
}

func wrapFullscreen(body, stepLabel, footerHint string, st theme.Styles, layout frameLayout) string {
	statusBar := st.RenderStatusBar("WIZARD", stepLabel, footerHint)
	column := joinBlocks(body, statusBar)
	return st.Theme.BorderStyle().
		Width(layout.InnerW).
		Padding(1, 2).
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
	case onbdomain.StepAgents:
		return "Agents"
	case onbdomain.StepDocs:
		return "Docs"
	case onbdomain.StepFeature:
		return "Feature"
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
