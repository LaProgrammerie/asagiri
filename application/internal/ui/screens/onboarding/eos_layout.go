package onboarding

import (
	"fmt"
	"strings"
	"time"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

const (
	eosLeftW  = 24
	eosRightW = 28
	// Vertical borders between the three body columns (left|center|right).
	eosBodyBorderCols = 4
)

// ShellContext carries workspace metadata for the EOS shell chrome.
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

type eosMetrics struct {
	termW, termH int
	bodyH        int
	leftW        int
	centerW      int
	rightW       int
}

func computeEOSMetrics(termW, termH int) eosMetrics {
	centerW := termW - eosLeftW - eosRightW - eosBodyBorderCols
	if centerW < 36 {
		centerW = 36
	}
	bodyH := termH - 2
	if bodyH < 12 {
		bodyH = 12
	}
	return eosMetrics{
		termW:   termW,
		termH:   termH,
		bodyH:   bodyH,
		leftW:   eosLeftW,
		centerW: centerW,
		rightW:  eosRightW,
	}
}

func renderEOS(vm ViewModel) string {
	if vm.Width < 90 || vm.Height < 24 {
		return renderWizardCompact(vm)
	}
	if vm.Model.Applied {
		return renderEOSReady(vm)
	}
	return renderEOSWizard(vm)
}

func renderEOSWizard(vm ViewModel) string {
	st := vm.Theme.Styles()
	p := st.Theme.Palette
	m := computeEOSMetrics(vm.Width, vm.Height)

	top := renderEOSTopBar(vm, st, m.termW)
	bottom := renderEOSBottomBar(vm, st, m.termW)
	left := renderEOSNav(vm, st, m.leftW, m.bodyH)
	center := renderEOSCenter(vm, st, m.centerW, m.bodyH)
	right := renderEOSContext(vm, st, m.rightW, m.bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(p.Background)).
		Width(m.termW).
		Render(top + "\n" + body + "\n" + bottom)
}

func renderEOSReady(vm ViewModel) string {
	st := vm.Theme.Styles()
	p := st.Theme.Palette
	m := computeEOSMetrics(vm.Width, vm.Height)

	top := renderEOSTopBar(vm, st, m.termW)
	bottom := renderEOSBottomBar(vm, st, m.termW)
	left := renderEOSNav(vm, st, m.leftW, m.bodyH)
	right := renderEOSContext(vm, st, m.rightW, m.bodyH)
	center := renderEOSCenterReady(vm, st, m.centerW, m.bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
	return lipgloss.NewStyle().
		Background(lipgloss.Color(p.Background)).
		Width(m.termW).
		Render(top + "\n" + body + "\n" + bottom)
}

func renderEOSTopBar(vm ViewModel, st theme.Styles, w int) string {
	p := st.Theme.Palette
	title := st.Brand.Render(" ASAGIRI ") + st.Muted.Render(" • ") + st.PanelTitle.Render("Engineering Operating System")
	clock := vm.Shell.Clock
	if clock == "" {
		clock = time.Now().Format("15:04:05")
	}
	right := st.Muted.Render(fmt.Sprintf("%s  %s  %s",
		fallbackStr(vm.Shell.Workspace, "workspace"),
		fallbackStr(vm.Shell.Branch, "main"),
		clock,
	))
	innerW := w - 2
	gap := innerW - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := title + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().
		Width(w).
		Background(lipgloss.Color(p.Background)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(0, 1).
		Render(line)
}

func renderEOSBottomBar(vm ViewModel, st theme.Styles, w int) string {
	p := st.Theme.Palette
	left := st.Muted.Render("TIP  Tab/↑↓ champs · Ctrl+N étape suivante · Ctrl+S apply")
	cost := fmt.Sprintf("€%.2f", vm.Shell.CostTodayEUR)
	online := st.Success.Render("● En ligne")
	if !vm.Shell.Online {
		online = st.Error.Render("● Hors ligne")
	}
	right := st.PanelTitle.Render("Mode: Wizard") + "  " + online +
		"  " + st.Muted.Render("API: "+fallbackStr(vm.Shell.APIProvider, "OpenRouter")) +
		"  " + st.Muted.Render("Cost: "+cost)
	innerW := w - 2
	gap := innerW - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().
		Width(w).
		Background(lipgloss.Color(p.Background)).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(0, 1).
		Render(line)
}

func renderEOSNav(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	var b strings.Builder
	b.WriteString(st.Muted.Render("NAVIGATION") + "\n\n")

	navItems := []struct {
		label  string
		active bool
	}{
		{"Onboarding Wizard", true},
		{"Dashboard", false},
		{"Flows", false},
		{"Runs", false},
		{"Agents", false},
		{"Trust", false},
		{"Knowledge", false},
		{"Replay", false},
		{"Settings", false},
	}
	for _, item := range navItems {
		if item.active {
			b.WriteString(st.PanelTitle.Render("▌ "+item.label) + "\n")
		} else {
			b.WriteString(st.Muted.Render("  "+item.label) + "\n")
		}
	}

	b.WriteString("\n" + st.Muted.Render("RACCOURCIS") + "\n\n")
	shortcuts := []string{
		"↑ ↓   Naviguer",
		"→     Étape suivante",
		"Enter Sélectionner",
		"Ctrl+P Précédent",
		"Ctrl+N Suivant",
	}
	for _, s := range shortcuts {
		b.WriteString(st.Hint.Render(s) + "\n")
	}

	ver := vm.Shell.Version
	if ver == "" {
		ver = "dev"
	}
	b.WriteString("\n" + st.Muted.Render("asa v"+ver))

	return eosColumn(b.String(), w, h, p, "left")
}

func renderEOSContext(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	m := vm.Model
	idx := stepIndex(m.Step) + 1
	total := len(onbdomain.TUIStepOrder)
	pct := idx * 100 / total

	var b strings.Builder
	b.WriteString(st.Muted.Render("CONTEXTE PROJET") + "\n\n")
	b.WriteString(renderKV(st, "Workspace", fallbackStr(vm.Shell.Workspace, fieldOr(m, "project_name", "—"))) + "\n")
	b.WriteString(renderKV(st, "Branche", fallbackStr(vm.Shell.Branch, fieldOr(m, "default_branch", "main"))) + "\n")
	b.WriteString(renderKV(st, "Répertoire", fallbackStr(vm.Shell.Directory, ".")) + "\n")
	status := st.Success.Render("● Initialisation")
	if m.Applied {
		if vm.Readiness.Ready {
			status = st.Success.Render("● READY")
		} else {
			status = st.Error.Render("● NOT READY")
		}
	}
	b.WriteString(renderKV(st, "Statut", status) + "\n")

	b.WriteString("\n" + st.Muted.Render("APERÇU PROCESSUS") + "\n\n")
	b.WriteString(renderProgressRing(st, pct) + "\n")
	b.WriteString(st.Muted.Render(fmt.Sprintf("%d / %d étapes", idx, total)) + "\n")

	b.WriteString("\n" + st.Muted.Render("PROCHAINE ÉTAPE") + "\n\n")
	b.WriteString(st.Fg.Render(nextStepTitle(m.Step)) + "\n")
	b.WriteString(st.Muted.Render(nextStepDesc(m.Step)) + "\n")

	return eosColumn(b.String(), w, h, p, "right")
}

func renderEOSCenter(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	m := vm.Model
	idx := stepIndex(m.Step) + 1
	innerW := columnTextWidth(w)

	var parts []string
	header := st.PanelTitle.Render("ASAGIRI PROJECT ONBOARDING WIZARD")
	stepTag := st.Muted.Render(fmt.Sprintf("Étape %d / %d", idx, len(onbdomain.TUIStepOrder)))
	gap := innerW - lipgloss.Width(header) - lipgloss.Width(stepTag)
	if gap < 1 {
		gap = 1
	}
	parts = append(parts, header+strings.Repeat(" ", gap)+stepTag)
	parts = append(parts, renderStepper(m.Step, st, innerW))
	content := renderEOSStepContent(vm, st, contentBoxWidth(w))
	parts = append(parts, content)
	footer := renderEOSFooter(m, st, innerW)

	used := 0
	for _, part := range parts {
		used += lineCount(part)
	}
	used += len(parts) - 1
	used++ // footer line
	spacerLines := h - used - 2
	if spacerLines < 0 {
		spacerLines = 0
	}

	var b strings.Builder
	b.WriteString(strings.Join(parts, "\n\n"))
	if spacerLines > 0 {
		b.WriteString(strings.Repeat("\n", spacerLines))
	}
	b.WriteString("\n")
	b.WriteString(footer)

	return eosColumn(b.String(), w, h, p, "center")
}

func renderEOSCenterReady(vm ViewModel, st theme.Styles, w, h int) string {
	p := st.Theme.Palette
	r := vm.Readiness
	if vm.Model.Readiness.Score > 0 || len(vm.Model.Readiness.Checks) > 0 {
		r = vm.Model.Readiness
	}

	var b strings.Builder
	b.WriteString(st.PanelTitle.Render("ONBOARDING TERMINÉ") + "\n\n")

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
	b.WriteString("\n" + st.Hint.Render("O · appliquer   N · ignorer   Ctrl+Q · quitter"))

	return eosColumn(b.String(), w, h, p, "center")
}

func renderEOSStepContent(vm ViewModel, st theme.Styles, boxW int) string {
	m := vm.Model
	if boxW < 20 {
		boxW = 20
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(st.Theme.Palette.Border)).
		Padding(1, 2).
		Width(boxW)

	switch m.Step {
	case onbdomain.StepWelcome:
		var b strings.Builder
		b.WriteString(st.PanelTitle.Render("Bienvenue dans Asagiri") + "\n\n")
		b.WriteString(st.Fg.Render("Je vais vous guider pour configurer votre projet.") + "\n")
		b.WriteString(st.Muted.Render("L'onboarding prépare config, validation, docs et spec Kiro.") + "\n\n")
		b.WriteString(st.Muted.Render("Ce que nous allons faire") + "\n")
		for _, item := range welcomeChecklist() {
			b.WriteString(st.Success.Render("✓ ") + st.Fg.Render(item) + "\n")
		}
		tipInner := boxW - 6
		if tipInner < 16 {
			tipInner = 16
		}
		tipBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(st.Theme.Palette.Border)).
			Padding(0, 1).
			Width(tipInner).
			Render(st.PanelTitle.Render("💡 Conseil") + "\n" +
				st.Muted.Render("Utilisez Tab pour naviguer entre les champs, Ctrl+N pour avancer."))
		b.WriteString("\n" + tipBox)
		return box.Render(b.String())

	default:
		var b strings.Builder
		b.WriteString(st.PanelTitle.Render(StepLabel(m.Step)) + "\n\n")
		for i, row := range m.fieldRows {
			focused := m.FocusFooter < 0 && i == m.FocusField
			b.WriteString(st.RenderField(row.Label, m.fieldValue(row.Key), focused) + "\n")
			if errMsg := m.Errors[row.Key]; errMsg != "" {
				b.WriteString(st.Error.Render("    ! "+errMsg) + "\n")
			}
		}
		if len(m.SkippedFields) > 0 {
			b.WriteString("\n" + st.Warning.Render("Conservé: "+strings.Join(m.SkippedFields, ", ")))
		}
		if m.Message != "" {
			b.WriteString("\n" + st.Success.Render(m.Message))
		}
		return box.Render(b.String())
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

func renderStepper(step onbdomain.WizardStep, st theme.Styles, w int) string {
	active := stepIndex(step)

	if lipgloss.Width(renderStepperFull(active, st)) <= w {
		return renderStepperFull(active, st)
	}
	return renderStepperCompact(active, st, w)
}

func wizardStepLabels() []string {
	out := make([]string, len(onbdomain.TUIStepOrder))
	for i, s := range onbdomain.TUIStepOrder {
		out[i] = StepLabel(s)
	}
	return out
}

func renderStepperFull(active int, st theme.Styles) string {
	labels := wizardStepLabels()
	var nodes []string
	for i, label := range labels {
		switch {
		case i < active:
			nodes = append(nodes, st.Success.Render("✓"))
		case i == active:
			nodes = append(nodes, st.PanelTitle.Render("●"))
		default:
			nodes = append(nodes, st.Muted.Render("○"))
		}
		nodes = append(nodes, st.Muted.Render(" "))
		nodes = append(nodes, labelStyle(i, active, st, label))
		if i < len(labels)-1 {
			nodes = append(nodes, st.Muted.Render(" ─ "))
		}
	}
	return strings.Join(nodes, "")
}

func labelStyle(i, active int, st theme.Styles, label string) string {
	switch {
	case i == active:
		return st.PanelTitle.Render(label)
	case i < active:
		return st.Success.Render(label)
	default:
		return st.Muted.Render(label)
	}
}

func renderStepperCompact(active int, st theme.Styles, w int) string {
	var parts []string
	for i := range onbdomain.TUIStepOrder {
		n := fmt.Sprintf("%d", i+1)
		switch {
		case i == active:
			parts = append(parts, st.PanelTitle.Render(n))
		case i < active:
			parts = append(parts, st.Success.Render(n))
		default:
			parts = append(parts, st.Muted.Render(n))
		}
	}
	sep := st.Muted.Render("─")
	line := strings.Join(parts, sep)
	if lipgloss.Width(line) > w {
		line = st.Muted.Render(fmt.Sprintf("Étape %d / %d", active+1, len(onbdomain.TUIStepOrder)))
	}
	return line
}

func renderEOSFooter(m Model, st theme.Styles, w int) string {
	prev := st.RenderButton("← Précédent", m.FocusFooter == FooterPrev)
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

func renderProgressRing(st theme.Styles, pct int) string {
	bar := st.RenderProgress(pct, 100, 10)
	return bar + "\n" + st.PanelTitle.Render(fmt.Sprintf("%d%%", pct))
}

func renderKV(st theme.Styles, key, val string) string {
	return st.Muted.Render(key+":") + " " + st.Fg.Render(val)
}

func columnTextWidth(colW int) int {
	if colW <= 2 {
		return 1
	}
	return colW - 2
}

func contentBoxWidth(colW int) int {
	w := columnTextWidth(colW) - 6
	if w < 20 {
		return 20
	}
	return w
}

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(s, "\n"), "\n"))
}

func eosColumn(content string, w, h int, p theme.Palette, side string) string {
	border := lipgloss.NormalBorder()
	style := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(0, 1).
		BorderForeground(lipgloss.Color(p.Border)).
		Background(lipgloss.Color(p.Background))
	switch side {
	case "left":
		style = style.Border(border, false, true, false, false)
	case "right":
		style = style.Border(border, false, false, false, true)
	default:
		style = style.Border(border, false, true, false, true)
	}
	return style.Render(content)
}

func nextStepTitle(step onbdomain.WizardStep) string {
	idx := stepIndex(step)
	order := onbdomain.TUIStepOrder
	if idx+1 >= len(order) {
		return "Application finale"
	}
	return StepLabel(order[idx+1])
}

func nextStepDesc(step onbdomain.WizardStep) string {
	switch step {
	case onbdomain.StepWelcome:
		return "Configuration du projet (nom, branche, tagline)."
	case onbdomain.StepProject:
		return "Détection et validation de la stack."
	case onbdomain.StepStack:
		return "Agents par défaut et reviewer."
	case onbdomain.StepAgents:
		return "Bootstrap docs/ai/ canon."
	case onbdomain.StepDocs:
		return "Première feature Kiro."
	case onbdomain.StepFeature:
		return "Récapitulatif et application."
	case onbdomain.StepReview:
		return "Score readiness et corrections auto."
	default:
		return ""
	}
}

func fieldOr(m Model, key, fallback string) string {
	if v := strings.TrimSpace(m.Fields[key]); v != "" {
		return v
	}
	return fallback
}

func fallbackStr(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "—"
}

func renderWizardCompact(vm ViewModel) string {
	st := vm.Theme.Styles()
	m := computeEOSMetrics(vm.Width, vm.Height)
	innerW := contentBoxWidth(m.centerW)
	model := vm.Model
	content := st.ContentArea.Width(innerW).Render(renderWizardPanelBody(vm, st))
	body := joinBlocks(
		st.RenderPageHeader("Project Onboarding Wizard", StepProgress(model.Step)),
		renderStepper(model.Step, st, innerW),
		content,
		renderEOSFooter(model, st, innerW),
	)
	return wrapFullscreen(body, StepLabel(model.Step), st, vm.Width)
}
