package onboarding

import (
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

const shellNavW = 26

// renderShell renders the onboarding wizard through the shared cockpit shell:
// top/bottom status bars + persistent nav rail (components.*), a responsive
// center built from the step card + progression dashboard. No fake telemetry.
func renderShell(vm ViewModel) string {
	st := vm.Theme.Styles()
	p := st.Theme.Palette
	w, h := vm.Width, vm.Height

	top := components.RenderTopBar(shellTopLeft(st), shellTopRight(vm, st), w, vm.Theme)
	bottom := components.RenderBottomBar(shellBottomLeft(st, w), shellBottomRight(vm, st), w, vm.Theme)

	bodyH := h - lipgloss.Height(top) - lipgloss.Height(bottom)
	if bodyH < 12 {
		bodyH = 12
	}

	rail := components.RenderNavRail("NAVIGATION", onboardingNavItems(), shellNavW, bodyH, vm.Theme)
	centerW := w - shellNavW
	if centerW < 36 {
		centerW = 36
	}

	var center string
	if vm.Model.Applied {
		center = renderShellReady(vm, st, centerW, bodyH)
	} else {
		center = renderShellCenter(vm, st, centerW, bodyH)
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, rail, center)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(p.Background)).
		Width(w).
		Render(top + "\n" + body + "\n" + bottom)
}

func onboardingNavItems() []components.NavItem {
	return []components.NavItem{
		{Icon: "◆", Label: "Onboarding", Active: true},
		{Icon: "⌂", Label: "Mission"},
		{Icon: "▤", Label: "Dashboard"},
		{Icon: "▶", Label: "Runs"},
		{Icon: "◎", Label: "Agents"},
		{Icon: "⛉", Label: "Trust"},
		{Icon: "❏", Label: "Knowledge"},
		{Icon: "↺", Label: "Replay"},
		{Icon: "⚙", Label: "Settings"},
	}
}

func shellTopLeft(st theme.Styles) string {
	return st.Brand.Render(" ASAGIRI ") + st.Muted.Render(" • ") + st.PanelTitle.Render("Project Onboarding")
}

func shellTopRight(vm ViewModel, st theme.Styles) string {
	clock := vm.Shell.Clock
	if clock == "" {
		clock = "--:--:--"
	}
	dot := st.Success.Render("●")
	if !vm.Shell.Online {
		dot = st.Muted.Render("●")
	}
	return dot + " " + st.Muted.Render(fallbackStr(vm.Shell.Workspace, "workspace")) +
		"   " + st.Muted.Render("⎇ "+fallbackStr(vm.Shell.Branch, "main")) +
		"   " + st.Muted.Render("◷ "+clock)
}

func shellBottomLeft(st theme.Styles, w int) string {
	if w >= 128 {
		return st.Muted.Render("TIP  Tab/↑↓ champs · Ctrl+N étape suivante · Ctrl+S apply")
	}
	return st.Muted.Render("TIP  Ctrl+N suivant · Ctrl+S apply")
}

func shellBottomRight(vm ViewModel, st theme.Styles) string {
	online := st.Success.Render("● En ligne")
	if !vm.Shell.Online {
		online = st.Error.Render("● Hors ligne")
	}
	return st.PanelTitle.Render("Mode: Wizard") + "  " + online
}

// renderShellCenter builds the wizard center column: header + step card +
// optional progression/readiness dashboard + footer, clamped to w×h.
func renderShellCenter(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	m := vm.Model
	idx := stepIndex(m.Step) + 1
	total := len(onbdomain.TUIStepOrder)
	innerW := w - 4
	if innerW < 24 {
		innerW = 24
	}

	header := renderEOSCenterHeader(st, m.Step, idx, total, innerW)
	card := renderEOSStepContent(vm, st, contentBoxWidth(w))
	footer := renderEOSFooter(m, st, innerW)

	const footerLines = 1
	base := header + "\n\n" + card
	if free := h - 2 - lineCount(base) - footerLines; free >= 9 {
		if dash := renderCenterDashboard(vm, st, innerW, free-2); dash != "" {
			base = base + "\n\n" + dash
		}
	}

	spacer := h - 2 - lineCount(base) - footerLines
	if spacer < 0 {
		spacer = 0
	}
	content := base + strings.Repeat("\n", spacer+1) + footer
	return clampColumn(content, w, h, p)
}

func renderShellReady(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	r := vm.Readiness
	if vm.Model.Readiness.Score > 0 || len(vm.Model.Readiness.Checks) > 0 {
		r = vm.Model.Readiness
	}

	var b strings.Builder
	b.WriteString(st.Brand.Render("ASAGIRI") + "  " + st.PanelTitle.Render("ONBOARDING TERMINÉ") + "\n\n")

	statusLabel := "NOT READY"
	statusStyle := st.Error
	if r.Ready {
		statusLabel = "READY"
		statusStyle = st.Success
	}
	b.WriteString(statusStyle.Render(statusLabel) + "  " + st.RenderProgress(r.Score, 100, 20) + "\n\n")

	if len(r.Checks) > 0 {
		b.WriteString(st.Muted.Render("Checks") + "\n")
		for _, c := range r.Checks {
			b.WriteString(st.RenderCheckStatus(c.Status, c.ID, c.Message) + "\n")
		}
	}
	if len(r.AutofixOffers) > 0 && !r.Ready {
		b.WriteString("\n" + st.AccentBlock.Render("Corrections auto disponibles") + "\n")
		for _, o := range r.AutofixOffers {
			b.WriteString(st.Fg.Render("• "+o.Title) + "\n")
		}
		b.WriteString("\n" + st.RenderButton("O · appliquer", true) + st.RenderButton("N · ignorer", false) + "\n")
	}
	if vm.Model.Message != "" {
		b.WriteString("\n" + st.Fg.Render(vm.Model.Message) + "\n")
	}
	if r.Ready {
		b.WriteString("\n" + st.Hint.Render("Bascule automatique vers Mission Control…"))
	} else {
		b.WriteString("\n" + st.Hint.Render("O · appliquer   N · ignorer   Ctrl+Q · quitter"))
	}

	return clampColumn(b.String(), w, h, p)
}

// clampColumn renders content inside a borderless w×h column (padding 1×2),
// trimming overflowing rows so the shared body layout never breaks.
func clampColumn(content string, w, h int, p theme.Palette) string {
	if maxLines := h - 2; maxLines > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > maxLines {
			content = strings.Join(lines[:maxLines], "\n")
		}
	}
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(1, 2).
		Background(lipgloss.Color(p.Background)).
		Render(content)
}
