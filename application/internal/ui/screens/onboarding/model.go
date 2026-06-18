package onboarding

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
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
	FieldAction  // explicit button (e.g. test connection)
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
	CloudProjectChoices []string
	ExistingConfig    bool     // true when .asagiri/config.yaml already exists
	ArtefactsPreview  []string // files that will be created on Apply
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
	m.CloudProjectChoices = append([]string(nil), form.CloudProjectChoices...)
	m.ExistingConfig = form.HasAsagiriConfig
	m.ArtefactsPreview = defaultArtefactsPreview(form.Answers.FeatureSlug)
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
	m.CloudProjectChoices = append([]string(nil), form.CloudProjectChoices...)
	m.ExistingConfig = form.HasAsagiriConfig
	m.ArtefactsPreview = defaultArtefactsPreview(form.Answers.FeatureSlug)
	m.RefreshFieldRows()
}

// CurrentForm rebuilds a domain form from model state.
func (m Model) CurrentForm() onbdomain.Form {
	form := onbdomain.FormFromMaps(m.Step, m.Fields, m.Advanced)
	form.ValidationPreview = append([]string(nil), m.ValidationPreview...)
	form.DetectedStacks = append([]string(nil), m.DetectedStacks...)
	form.SkippedFields = append([]string(nil), m.SkippedFields...)
	form.Errors = m.Errors
	form.CloudProjectChoices = append([]string(nil), m.CloudProjectChoices...)
	if strings.TrimSpace(m.Fields["cloud_probed"]) == "true" {
		form.CloudProbed = true
	}
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
		row := m.fieldRows[m.FocusField]
		if row.Kind == FieldAction {
			switch key {
			case "enter", " ":
				m.Message = "Test de connexion en cours…"
				return m, m.TestCloudActionCmd()
			case "tab", "down":
				m.cycleFocus(1)
			case "shift+tab", "up":
				m.cycleFocus(-1)
			case "left", "h", "esc":
				m.cycleFocus(-1)
			case "right", "l":
				m.cycleFocus(1)
			}
			return m, nil
		}
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
		if len(m.fieldRows) > 0 && m.FocusField >= 0 && m.FocusField < len(m.fieldRows) {
			row := m.fieldRows[m.FocusField]
			if row.Kind == FieldAction {
				m.Message = "Test de connexion en cours…"
				return m, m.TestCloudActionCmd()
			}
		}
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
	return row.Kind == FieldSelect || row.Kind == FieldAction || (!row.ReadOnly && row.Kind == FieldText)
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
	if key == "cloud_mode" {
		m.Fields[key] = choices[idx]
		if onbdomain.CloudConnect(choices[idx]) && strings.TrimSpace(m.Fields["cloud_base_url"]) == "" {
			m.Fields["cloud_base_url"] = onbdomain.DefaultCloudOnboardBaseURL()
		}
		m.RefreshFieldRows()
		return
	}
	if key == "cloud_project_id" {
		m.Fields[key] = onbdomain.ProjectIDForLabel(m.Fields["cloud_project_options"], choices[idx])
		m.Fields["cloud_project_name"] = choices[idx]
		return
	}
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
			if rows[i].Kind == FieldAction {
				m.Message = "Test de connexion en cours…"
				return m, m.TestCloudActionCmd()
			}
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

// FocusFieldByKey moves focus to the first row with the given key.
func (m *Model) FocusFieldByKey(key string) bool {
	for i, row := range m.fieldRows {
		if row.Key == key {
			m.FocusField = i
			m.FocusFooter = -1
			return true
		}
	}
	return false
}

// RefreshFieldRows rebuilds visible field rows for the current step.
func (m *Model) RefreshFieldRows() {
	m.fieldRows = stepFields(m.Step, m.ShowAdvanced, m.ValidationPreview, m.DetectedStacks, m.AgentChoices, m.Fields["cloud_mode"], m.CloudProjectChoices)
	if m.FocusField >= len(m.fieldRows) {
		m.FocusField = max(0, len(m.fieldRows)-1)
	}
}

func stepFields(step onbdomain.WizardStep, advanced bool, preview, stacks, agentChoices []string, cloudMode string, cloudProjectChoices []string) []FieldDef {
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
			{Key: "stack", Label: "Technologie", Kind: FieldSelect, Choices: defaultStackChoices()},
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
	case onbdomain.StepProviders:
		rows = []FieldDef{
			{Key: "enabled_providers", Label: "Outils IA installés (séparés par des virgules)", Kind: FieldText},
			{Key: "provider_hint", Label: "Exemples", Kind: FieldManaged, ReadOnly: true},
		}
	case onbdomain.StepAgents:
		rows = []FieldDef{
			{Key: "default_spec_agent", Label: "Rédaction de la spec", Kind: FieldSelect, Choices: choices},
			{Key: "pipeline_plan", Label: "Planification", Kind: FieldManaged, ReadOnly: true},
			{Key: "default_enricher", Label: "Enrichissement", Kind: FieldSelect, Choices: choices},
			{Key: "default_agent", Label: "Développement", Kind: FieldSelect, Choices: choices},
			{Key: "pipeline_verify", Label: "Vérification", Kind: FieldManaged, ReadOnly: true},
			{Key: "default_reviewer", Label: "Relecture", Kind: FieldSelect, Choices: choices},
		}
	case onbdomain.StepCloud:
		rows = []FieldDef{
			{Key: "cloud_mode", Label: "Cloud (optionnel)", Kind: FieldSelect, Choices: []string{"local", "connect"}},
		}
		if onbdomain.CloudConnect(cloudMode) {
			rows = append(rows,
				FieldDef{Key: "cloud_base_url", Label: "URL"},
				FieldDef{Key: "cloud_token", Label: "Token"},
				FieldDef{Key: "cloud_test", Label: "Tester la connexion", Kind: FieldAction},
				FieldDef{Key: "cloud_status", Label: "Statut", Kind: FieldManaged, ReadOnly: true},
			)
			projectChoices := cloudProjectChoices
			if len(projectChoices) == 0 {
				projectChoices = []string{"(testez la connexion d'abord)"}
			}
			rows = append(rows, FieldDef{
				Key: "cloud_project_id", Label: "Projet", Kind: FieldSelect, Choices: projectChoices,
			})
		}
	case onbdomain.StepDocs:
		rows = []FieldDef{{Key: "product_one_liner", Label: productOneLinerFieldLabel()}}
	case onbdomain.StepFeature:
		rows = []FieldDef{{Key: "feature_slug", Label: "Nom du premier chantier (ex. auth-mvp)"}}
	case onbdomain.StepReview:
		rows = []FieldDef{
			{Key: "project_name", Label: "Projet", ReadOnly: true},
			{Key: "default_branch", Label: "Branche", ReadOnly: true},
			{Key: "stack", Label: "Technologie", ReadOnly: true},
			{Key: "default_spec_agent", Label: "Rédaction spec", ReadOnly: true},
			{Key: "default_enricher", Label: "Enrichissement", ReadOnly: true},
			{Key: "default_agent", Label: "Développement", ReadOnly: true},
			{Key: "default_reviewer", Label: "Relecture", ReadOnly: true},
			{Key: "product_one_liner", Label: productOneLinerFieldLabel(), ReadOnly: true},
			{Key: "feature_slug", Label: "Premier chantier", ReadOnly: true},
			{Key: "cloud_mode", Label: "Cloud", ReadOnly: true},
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
	case "provider_hint":
		return "kiro-cli, cursor-cli, claude-code, ollama"
	case "cloud_mode":
		mode := strings.TrimSpace(m.Fields["cloud_mode"])
		if onbdomain.CloudSkipped(mode) {
			return "Local uniquement"
		}
		return "Connecter Asagiri Cloud"
	case "cloud_status":
		if onbdomain.CloudSkipped(m.Fields["cloud_mode"]) {
			return ""
		}
		if strings.TrimSpace(m.Fields["cloud_probed"]) == "true" {
			email := strings.TrimSpace(m.Fields["cloud_connected_email"])
			if email != "" {
				return "✓ Connecté : " + email
			}
			return "✓ Connecté"
		}
		return "Non connecté"
	case "cloud_test":
		return "[ Tester la connexion ]"
	case "cloud_project_id":
		if name := strings.TrimSpace(m.Fields["cloud_project_name"]); name != "" {
			return name
		}
		id := strings.TrimSpace(m.Fields["cloud_project_id"])
		if id == "" {
			return ""
		}
		for label, pid := range onbdomain.DecodeCloudProjectOptions(m.Fields["cloud_project_options"]) {
			if pid == id {
				return label
			}
		}
		return id
	}
	if isAdvancedKey(key) {
		return m.Advanced[key]
	}
	if key == "cloud_token" {
		val := m.Fields[key]
		if val != "" && (m.FocusFooter >= 0 || !m.editingField() || m.fieldRows[m.FocusField].Key != "cloud_token") {
			n := len(val)
			if n > 16 {
				n = 16
			}
			return strings.Repeat("*", n)
		}
	}
	return m.Fields[key]
}

func (m Model) onReviewStep() bool { return m.Step == onbdomain.StepReview }
func (m Model) canApply() bool     { return m.onReviewStep() && !m.Applied }
func (m Model) footerCount() int   { return 4 }

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

// OnboardingTestCloudRequestMsg asks the app shell to probe cloud connectivity.
type OnboardingTestCloudRequestMsg struct {
	Form onbdomain.Form
}

// TestCloudActionCmd returns a message for the app shell to probe cloud API.
func (m Model) TestCloudActionCmd() tea.Cmd {
	return func() tea.Msg {
		return OnboardingTestCloudRequestMsg{Form: m.CurrentForm()}
	}
}

// OnboardingFooterMsg is handled by the app shell to dispatch bus commands.
type OnboardingFooterMsg struct {
	Button int
	Form   onbdomain.Form
}

// OnboardingTestCloudMsg delivers cloud probe results to the wizard.
type OnboardingTestCloudMsg struct {
	Form onbdomain.Form
	Err  error
	Msg  string
}

func defaultAgentChoices() []string {
	return config.DefaultLogicalAgentIDs()
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
	return "Nom du premier chantier (ex. auth-mvp)"
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
		return "Technologie"
	case onbdomain.StepProviders:
		return "Outils IA"
	case onbdomain.StepAgents:
		return "Répartition"
	case onbdomain.StepCloud:
		return "Cloud (optionnel)"
	case onbdomain.StepDocs:
		return "Description"
	case onbdomain.StepFeature:
		return "Premier chantier"
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

// defaultArtefactsPreview returns the list of files that will be created on Apply.
func defaultArtefactsPreview(featureSlug string) []string {
	artefacts := []string{
		".asagiri/config.yaml",
		"docs/ai/",
		"docs/ai/01-product.md",
		"docs/ai/02-architecture.md",
	}
	if slug := strings.TrimSpace(featureSlug); slug != "" {
		artefacts = append(artefacts, ".kiro/specs/"+slug+"/")
		artefacts = append(artefacts, ".kiro/specs/"+slug+"/requirements.md")
	}
	return artefacts
}

// advancedButtonLabel returns the footer label for the Advanced toggle,
// showing the parameter count and open/closed state.
func advancedButtonLabel(m Model) string {
	const nParams = 6 // work_stop_after, budget, verification, coordination, theme, mcp
	if m.ShowAdvanced {
		return fmt.Sprintf("Mode expert (%d) ▼", nParams)
	}
	return fmt.Sprintf("Mode expert (%d)", nParams)
}
