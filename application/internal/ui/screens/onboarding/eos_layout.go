package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ShellContext carries workspace metadata for the shared shell chrome.
type ShellContext struct {
	Workspace    string
	Branch       string
	Directory    string
	Clock        string
	Version      string
	CostTodayEUR float64
	APIProvider  string
	Online       bool
}

// renderEOSCenterHeader builds the wizard header: brand + onboarding subtitle and
// the step pill. The step name is not repeated here (FR-5.3: pill only).
func renderEOSCenterHeader(st theme.Styles, _ onbdomain.WizardStep, idx, total, innerW int) string {
	brand := st.Brand.Render("ASAGIRI")
	subtitle := st.Muted.Render("Project Onboarding")
	left := brand + "  " + subtitle
	pill := st.RenderBadge(fmt.Sprintf("Étape %d / %d", idx, total))
	gap := innerW - lipgloss.Width(left) - lipgloss.Width(pill)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + pill
}

func renderEOSStepContent(vm ViewModel, st theme.Styles, boxW int) string {
	if boxW < 20 {
		boxW = 20
	}
	switch vm.Model.Step {
	case onbdomain.StepWelcome:
		return renderWelcomePanel(vm, st, boxW)
	default:
		return renderStepPanel(vm, st, boxW)
	}
}

func welcomeChecklist() []string {
	return []string{
		"Configurer la stack détectée",
		"Créer config.yaml et validation",
		"Initialiser docs/ai/",
		"Configurer les agents IA",
		"Créer spec Kiro initiale",
		"Évaluer la readiness",
	}
}

func renderEOSFooter(m Model, st theme.Styles, w int) string {
	var prev string
	if m.Step == onbdomain.StepWelcome {
		prev = st.ButtonGhost.Render("← Précédent")
	} else {
		prev = st.RenderButton("← Précédent", m.FocusFooter == FooterPrev)
	}
	nextLabel := "Suivant →"
	nextFocused := m.FocusFooter == FooterNext
	if m.canApply() {
		nextLabel = "Appliquer"
		nextFocused = m.FocusFooter == FooterApply
	}
	next := st.RenderButton(nextLabel, nextFocused)
	buttons := prev + "  " + next
	if m.Step != onbdomain.StepWelcome {
		adv := st.RenderButton("Advanced", m.FocusFooter == FooterAdvanced && m.ShowAdvanced)
		buttons += "  " + adv
	}
	gap := w - lipgloss.Width(buttons)
	if gap < 0 {
		gap = 0
	}
	return strings.Repeat(" ", gap) + buttons
}

// columnTextWidth returns the usable text width inside a column, accounting
// for the horizontal padding (2 cells each side).
func columnTextWidth(colW int) int {
	w := colW - 4
	if w < 1 {
		return 1
	}
	return w
}

// contentBoxWidth returns the width for the bordered card inside the center
// column (its border adds 2 cells, so it must be 2 narrower than the text area).
func contentBoxWidth(colW int) int {
	w := columnTextWidth(colW) - 2
	if w < 24 {
		return 24
	}
	return w
}

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(s, "\n"), "\n"))
}

func fallbackStr(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "—"
}
