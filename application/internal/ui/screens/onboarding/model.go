package onboarding

import (
	"fmt"
	"strings"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	FooterPrev     = 0
	FooterNext     = 1
	FooterAdvanced = 2
	FooterApply    = 3
)

// FieldKind controls how a wizard row is edited and rendered.
type FieldKind int

const (
	FieldText FieldKind = iota
	FieldSelect
	FieldManaged // workflow step owned by Asagiri; not editable
)

// FieldDef describes one editable wizard field row.
type FieldDef struct {
	Key      string
	Label    string
	ReadOnly bool
	Kind     FieldKind
	Choices  []string // FieldSelect: keys from config.agents
}

// Model drives the interactive onboarding wizard.
type Model struct {
	Step              onbdomain.WizardStep
	Fields            map[string]string
	Advanced          map[string]string
	FocusField        int
	FocusFooter       int // -1 = field focus
	ShowAdvanced      bool
	Applied           bool
	Readiness         bus.ReadinessResult
	Errors            map[string]string
	ValidationPreview []string
	DetectedStacks    []string
	SkippedFields     []string
	Message           string
	MouseEnabled      bool
	Initialized       bool
	AgentChoices      []string
	fieldRows         []FieldDef
}

// NewModel returns an empty wizard model.
func NewModel() Model {
	return Model{
		Fields:      map[string]string{},
		Advanced:    map[string]string{},
		Errors:      map[string]string{},
		FocusFooter: -1,
		Step:        onbdomain.StepWelcome,
	}
}

// NewModelFromForm builds UI state from an onboarding Form.
func NewModelFromForm(form onbdomain.Form, mouseEnabled bool) Model {
	m := NewModel()
	m.Step = form.Step
	m.Fields = form.FieldsMap()
	m.Advanced = form.AdvancedMap()
	m.ValidationPreview = append([]string(nil), form.ValidationPreview...)
	m.DetectedStacks = append([]string(nil), form.DetectedStacks...)
	m.SkippedFields = append([]string(nil), form.SkippedFields...)
	m.Errors = form.Errors
	m.MouseEnabled = mouseEnabled
	m.Initialized = true
	m.AgentChoices = append([]string(nil), form.KnownAgentKeys...)
	m.RefreshFieldRows()
	return m
}

// SyncForm copies a domain form into the model.
func (m *Model) SyncForm(form onbdomain.Form) {
	m.Step = form.Step
	m.Fields = form.FieldsMap()
	m.Advanced = form.AdvancedMap()
	m.ValidationPreview = append([]string(nil), form.ValidationPreview...)
	m.DetectedStacks = append([]string(nil), form.DetectedStacks...)
	m.SkippedFields = append([]string(nil), form.SkippedFields...)
	if form.Errors != nil {
		m.Errors = form.Errors
	}
	m.AgentChoices = append([]string(nil), form.KnownAgentKeys...)
	m.RefreshFieldRows()
}

// CurrentForm rebuilds a domain form from model state.
func (m Model) CurrentForm() onbdomain.Form {
	form := onbdomain.FormFromMaps(m.Step, m.Fields, m.Advanced)
	form.ValidationPreview = append([]string(nil), m.ValidationPreview...)
	form.DetectedStacks = append([]string(nil), m.DetectedStacks...)
	form.SkippedFields = append([]string(nil), m.SkippedFields...)
	form.Errors = m.Errors
	return form
}

// Update handles keyboard and mouse input for the wizard.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		return m.updateKey(v)
	case tea.MouseMsg:
		if m.MouseEnabled {
			return m.updateMouse(v)
		}
	}
	return m, nil
}

func (m Model) updateKey(v tea.KeyMsg) (Model, tea.Cmd) {
	key := v.String()

	if m.Applied {
		switch key {
		case "ctrl+q", "ctrl+c":
			return m, tea.Quit
		case "ctrl+f":
			if m.hasAutofixOffers() {
				return m, m.AutofixActionCmd()
			}
		case "y", "Y", "o", "O", "enter":
			if m.hasAutofixOffers() {
				return m, m.AutofixActionCmd()
			}
		case "n", "N":
			m.Message = "Corrections ignorées — Ctrl+Q pour quitter"
		}
		return m, nil
	}

	if cmd, handled := (&m).handleCtrlAction(key); handled {
		return m, cmd
	}

	if m.FocusFooter >= 0 {
		switch key {
		case "tab", "right":
			m.FocusFooter = (m.FocusFooter + 1) % m.footerCount()
		case "shift+tab", "left":
			m.FocusFooter = (m.FocusFooter - 1 + m.footerCount()) % m.footerCount()
		case "enter", " ":
			return m, m.FooterActionCmd(m.FocusFooter)
		case "esc", "up":
			m.FocusFooter = -1
			if len(m.fieldRows) > 0 {
				m.FocusField = 0
			}
		}
		return m, nil
	}

	if m.editingField() {
		if m.focusedSelect() {
			switch key {
			case "tab", "down":
				m.cycleFocus(1)
			case "shift+tab", "up":
				m.cycleFocus(-1)
			case "left", "h":
				m.cycleSelectField(-1)
			case "right", "l":
				m.cycleSelectField(1)
			case "esc":
				m.FocusFooter = FooterNext
			case "enter", " ":
				m.cycleSelectField(1)
			default:
				if v.Type == tea.KeyRunes && len(v.Runes) > 0 {
					m.jumpSelectToPrefix(string(v.Runes))
				}
			}
			return m, nil
		}
		switch key {
		case "tab", "down":
			m.cycleFocus(1)
		case "shift+tab", "up":
			m.cycleFocus(-1)
		case "left":
			m.cycleFocus(-1)
		case "right":
			m.cycleFocus(1)
		case "esc":
			m.FocusFooter = FooterNext
		case "backspace":
			m.editFocusedField(func(val string) string {
				if len(val) == 0 {
					return val
				}
				return val[:len(val)-1]
			})
		case " ":
			m.editFocusedField(func(val string) string { return val + " " })
		default:
			if v.Type == tea.KeyRunes && len(v.Runes) > 0 {
				ch := string(v.Runes)
				m.editFocusedField(func(val string) string { return val + ch })
			}
		}
		return m, nil
	}

	switch key {
	case "tab", "down":
		m.cycleFocus(1)
	case "shift+tab", "up":
		m.cycleFocus(-1)
	case "left":
		m.cycleFocus(-1)
	case "right":
		m.cycleFocus(1)
	case "esc":
		m.FocusFooter = FooterNext
	case "enter", " ":
		return m, m.FooterActionCmd(FooterNext)
	default:
		if v.Type == tea.KeyRunes && len(v.Runes) > 0 && !m.focusedReadOnly() {
			ch := string(v.Runes)
			m.editFocusedField(func(val string) string { return val + ch })
		}
	}
	return m, nil
}

// handleCtrlAction maps wizard actions to Ctrl+ shortcuts (safe while typing in fields).
func (m *Model) handleCtrlAction(key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+p":
		return m.FooterActionCmd(FooterPrev), true
	case "ctrl+n":
		if m.canApply() && m.onReviewStep() {
			return m.FooterActionCmd(FooterApply), true
		}
		return m.FooterActionCmd(FooterNext), true
	case "ctrl+a":
		m.ShowAdvanced = !m.ShowAdvanced
		m.RefreshFieldRows()
		return nil, true
	case "ctrl+s":
		if m.canApply() {
			return m.FooterActionCmd(FooterApply), true
		}
		return nil, true
	case "ctrl+q":
		return tea.Quit, true
	default:
		return nil, false
	}
}

func (m Model) editingField() bool {
	if m.FocusFooter >= 0 {
		return false
	}
	rows := m.fieldRows
	if len(rows) == 0 {
		return false
	}
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return false
	}
	row := rows[m.FocusField]
	return row.Kind == FieldSelect || (!row.ReadOnly && row.Kind == FieldText)
}

func (m Model) focusedSelect() bool {
	rows := m.fieldRows
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return false
	}
	return rows[m.FocusField].Kind == FieldSelect
}

func (m *Model) cycleSelectField(dir int) {
	rows := m.fieldRows
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return
	}
	row := rows[m.FocusField]
	choices := row.Choices
	if len(choices) == 0 {
		if row.Key == "stack" {
			choices = defaultStackChoices()
		} else {
			choices = defaultAgentChoices()
		}
	}
	key := row.Key
	current := strings.TrimSpace(m.Fields[key])
	idx := 0
	for i, c := range choices {
		if c == current {
			idx = i
			break
		}
	}
	idx += dir
	for idx < 0 {
		idx += len(choices)
	}
	idx %= len(choices)
	m.Fields[key] = choices[idx]
}

func (m *Model) jumpSelectToPrefix(prefix string) {
	rows := m.fieldRows
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return
	}
	row := rows[m.FocusField]
	choices := row.Choices
	if len(choices) == 0 {
		return
	}
	prefix = strings.ToLower(prefix)
	for _, c := range choices {
		if strings.HasPrefix(strings.ToLower(c), prefix) {
			m.Fields[row.Key] = c
			return
		}
	}
}

func (m Model) updateMouse(v tea.MouseMsg) (Model, tea.Cmd) {
	if v.Button != tea.MouseButtonLeft || v.Action != tea.MouseActionPress {
		return m, nil
	}
	rows := m.fieldRows
	baseY := 6
	for i := range rows {
		if v.Y >= baseY+i && v.Y < baseY+i+1 {
			m.FocusField = i
			m.FocusFooter = -1
			return m, nil
		}
	}
	footerY := baseY + len(rows) + 2
	if v.Y >= footerY && v.Y <= footerY+1 {
		btnWidth := 12
		for i := 0; i < m.footerCount(); i++ {
			xStart := 2 + i*(btnWidth+2)
			if v.X >= xStart && v.X < xStart+btnWidth {
				m.FocusFooter = i
				return m, m.FooterActionCmd(i)
			}
		}
	}
	return m, nil
}

func (m *Model) cycleFocus(dir int) {
	rows := m.fieldRows
	if len(rows) == 0 {
		m.FocusFooter = FooterNext
		return
	}
	next := m.FocusField + dir
	if next < 0 {
		m.FocusFooter = FooterPrev
		return
	}
	if next >= len(rows) {
		m.FocusFooter = FooterNext
		return
	}
	m.FocusField = next
	m.FocusFooter = -1
}

func (m *Model) editFocusedField(edit func(string) string) {
	rows := m.fieldRows
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return
	}
	key := rows[m.FocusField].Key
	if isAdvancedKey(key) {
		m.Advanced[key] = edit(m.Advanced[key])
	} else {
		m.Fields[key] = edit(m.Fields[key])
	}
}

func (m Model) focusedReadOnly() bool {
	rows := m.fieldRows
	if m.FocusField < 0 || m.FocusField >= len(rows) {
		return true
	}
	return rows[m.FocusField].ReadOnly
}

func isAdvancedKey(key string) bool {
	switch key {
	case "work_stop_after", "budget_max_cost", "verification_profile",
		"coordination_max_parallel", "ui_theme", "mcp_enabled":
		return true
	default:
		return false
	}
}

// RefreshFieldRows rebuilds visible field rows for the current step.
func (m *Model) RefreshFieldRows() {
	m.fieldRows = stepFields(m.Step, m.ShowAdvanced, m.ValidationPreview, m.DetectedStacks, m.AgentChoices)
	if m.FocusField >= len(m.fieldRows) {
		m.FocusField = max(0, len(m.fieldRows)-1)
	}
}

func stepFields(step onbdomain.WizardStep, advanced bool, preview, stacks, agentChoices []string) []FieldDef {
	choices := agentChoices
	if len(choices) == 0 {
		choices = defaultAgentChoices()
	}
	var rows []FieldDef
	switch step {
	case onbdomain.StepWelcome:
		return rows
	case onbdomain.StepProject:
		rows = []FieldDef{
			{Key: "project_name", Label: "Nom du projet"},
			{Key: "default_branch", Label: "Branche par défaut"},
			{Key: "tagline", Label: "Tagline (optionnel)"},
		}
	case onbdomain.StepStack:
		rows = []FieldDef{
			{Key: "stack", Label: "Stack", Kind: FieldSelect, Choices: defaultStackChoices()},
		}
		if len(stacks) > 0 {
			rows = append(rows, FieldDef{
				Key: "detected_stacks", Label: "Détecté", ReadOnly: true,
			})
		}
		for i := range preview {
			rows = append(rows, FieldDef{
				Key: fmt.Sprintf("validation_%d", i), Label: "Validation", ReadOnly: true,
			})
		}
	case onbdomain.StepAgents:
		rows = []FieldDef{
			{Key: "default_spec_agent", Label: "Spec", Kind: FieldSelect, Choices: choices},
			{Key: "pipeline_plan", Label: "Plan", Kind: FieldManaged, ReadOnly: true},
			{Key: "default_enricher", Label: "Enrich", Kind: FieldSelect, Choices: choices},
			{Key: "default_agent", Label: "Dev", Kind: FieldSelect, Choices: choices},
			{Key: "pipeline_verify", Label: "Verify", Kind: FieldManaged, ReadOnly: true},
			{Key: "default_reviewer", Label: "Review", Kind: FieldSelect, Choices: choices},
		}
	case onbdomain.StepDocs:
		rows = []FieldDef{{Key: "product_one_liner", Label: productOneLinerFieldLabel()}}
	case onbdomain.StepFeature:
		rows = []FieldDef{{Key: "feature_slug", Label: firstFeatureSlugFieldLabel()}}
	case onbdomain.StepReview:
		rows = []FieldDef{
			{Key: "project_name", Label: "Projet", ReadOnly: true},
			{Key: "default_branch", Label: "Branche", ReadOnly: true},
			{Key: "stack", Label: "Stack", ReadOnly: true},
			{Key: "default_spec_agent", Label: "Spec", ReadOnly: true},
			{Key: "default_enricher", Label: "Enrich", ReadOnly: true},
			{Key: "default_agent", Label: "Dev", ReadOnly: true},
			{Key: "default_reviewer", Label: "Review", ReadOnly: true},
			{Key: "product_one_liner", Label: productOneLinerFieldLabel(), ReadOnly: true},
			{Key: "feature_slug", Label: firstFeatureStepLabel(), ReadOnly: true},
		}
	}
	if advanced {
		rows = append(rows,
			FieldDef{Key: "work_stop_after", Label: "work.stop_after"},
			FieldDef{Key: "budget_max_cost", Label: "budgets.per_run.max_estimated_cost"},
			FieldDef{Key: "verification_profile", Label: "verification.default_profile"},
			FieldDef{Key: "coordination_max_parallel", Label: "coordination.max_parallel_agents"},
			FieldDef{Key: "ui_theme", Label: "ui.theme"},
			FieldDef{Key: "mcp_enabled", Label: "mcp.enabled (true/false)"},
		)
	}
	return rows
}

func (m Model) fieldValue(key string) string {
	if strings.HasPrefix(key, "validation_") {
		idx := 0
		_, _ = fmt.Sscanf(key, "validation_%d", &idx)
		if idx >= 0 && idx < len(m.ValidationPreview) {
			return m.ValidationPreview[idx]
		}
		return ""
	}
	if key == "detected_stacks" {
		return strings.Join(m.DetectedStacks, ", ")
	}
	switch key {
	case "pipeline_plan":
		return "Géré par Asagiri"
	case "pipeline_verify":
		return "Géré par Asagiri"
	}
	if isAdvancedKey(key) {
		return m.Advanced[key]
	}
	return m.Fields[key]
}

func (m Model) onReviewStep() bool { return m.Step == onbdomain.StepReview }
func (m Model) canApply() bool      { return m.onReviewStep() && !m.Applied }
func (m Model) footerCount() int    { return 4 }

// FooterActionCmd returns a message for the app shell to dispatch bus commands.
func (m Model) FooterActionCmd(btn int) tea.Cmd {
	return func() tea.Msg {
		return OnboardingFooterMsg{Button: btn, Form: m.CurrentForm()}
	}
}

// OnboardingAutofixMsg requests automatic readiness fixes from the app shell.
type OnboardingAutofixMsg struct{}

// AutofixActionCmd returns a message for the app shell to dispatch autofix.
func (m Model) AutofixActionCmd() tea.Cmd {
	return func() tea.Msg {
		return OnboardingAutofixMsg{}
	}
}

func (m Model) hasAutofixOffers() bool {
	return len(m.Readiness.AutofixOffers) > 0 && !m.Readiness.Ready
}

// OnboardingFooterMsg is handled by the app shell to dispatch bus commands.
type OnboardingFooterMsg struct {
	Button int
	Form   onbdomain.Form
}

func defaultAgentChoices() []string {
	return []string{"cursor", "codex", "kiro", "ollama", "claude"}
}

func defaultStackChoices() []string {
	return []string{"auto", "go", "castor", "node"}
}

func productOneLinerFieldLabel() string {
	return "Que fait le produit ? (une phrase, ≠ nom du projet)"
}

func firstFeatureStepLabel() string {
	return "1ère feature"
}

func firstFeatureSlugFieldLabel() string {
	return "Slug Kiro (dossier .kiro/specs/<slug>/)"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// StepLabel returns a human-readable step title.
func StepLabel(step onbdomain.WizardStep) string {
	switch step {
	case onbdomain.StepWelcome:
		return "Bienvenue"
	case onbdomain.StepProject:
		return "Projet"
	case onbdomain.StepStack:
		return "Stack"
	case onbdomain.StepAgents:
		return "Agents"
	case onbdomain.StepDocs:
		return "Docs"
	case onbdomain.StepFeature:
		return firstFeatureStepLabel()
	case onbdomain.StepReview:
		return "Récap"
	default:
		return string(step)
	}
}

// StepProgress returns "2/7" style progress for the TUI header.
func StepProgress(step onbdomain.WizardStep) string {
	idx := -1
	for i, s := range onbdomain.TUIStepOrder {
		if s == step {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", idx+1, len(onbdomain.TUIStepOrder))
}
