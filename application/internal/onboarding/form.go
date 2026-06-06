package onboarding

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding/detect"
)

// AdvancedFields holds optional config overrides from the wizard advanced panel.
type AdvancedFields struct {
	WorkStopAfter           string `json:"work_stop_after,omitempty"`
	BudgetMaxCost           string `json:"budget_max_cost,omitempty"`
	VerificationProfile     string `json:"verification_profile,omitempty"`
	CoordinationMaxParallel string `json:"coordination_max_parallel,omitempty"`
	UITheme                 string `json:"ui_theme,omitempty"`
	MCPEnabled              string `json:"mcp_enabled,omitempty"`
}

// Form is the in-memory wizard form shared by CLI and TUI.
type Form struct {
	Step              WizardStep
	Answers           Answers
	Advanced          AdvancedFields
	ValidationPreview []string
	DetectedStacks    []string
	KnownAgentKeys    []string
	Errors            map[string]string
	SkippedFields     []string
	HasAsagiriConfig  bool // true when a real .asagiri/config.yaml was loaded
}

// TUIStepOrder is the interactive wizard flow (review then apply via command).
var TUIStepOrder = []WizardStep{
	StepWelcome,
	StepProject,
	StepStack,
	StepProviders,
	StepAgents,
	StepDocs,
	StepFeature,
	StepReview,
}

// BuildForm prefills wizard answers from persisted state, repo detection, and existing config.
func BuildForm(repoRoot string, st State, cfg *config.Config) Form {
	form := Form{
		Step:   st.CurrentStep,
		Answers: st.Answers,
		Errors: make(map[string]string),
	}
	if form.Step == "" || tuiStepIndex(form.Step) < 0 {
		form.Step = StepWelcome
	}
	form.Answers = MergeAnswers(State{Answers: form.Answers}, Options{}, repoRoot).Answers

	if cfg != nil {
		form.HasAsagiriConfig = !config.IsTemplateDefaultProjectName(cfg.Project.Name) || cfg.Work.DefaultAgent != ""
		form = overlayConfig(form, cfg)
		form.Advanced = advancedFromConfig(cfg)
	}

	matches, cmds := detect.DetectAll(repoRoot, firstNonEmpty(form.Answers.Stack, "auto"))
	if form.Answers.Stack == "" {
		form.Answers.Stack = detect.PrimaryStack(matches)
	}
	for _, m := range matches {
		form.DetectedStacks = append(form.DetectedStacks, m.ID)
	}
	for _, c := range cmds {
		line := c.Name + ": " + c.Command
		if c.Required {
			line += " (required)"
		}
		form.ValidationPreview = append(form.ValidationPreview, line)
	}
	if form.Answers.FeatureSlug == "" && form.Answers.ProjectName != "" {
		form.Answers.FeatureSlug = SlugFromName(form.Answers.ProjectName) + "-mvp"
	}
	if form.Step == StepDocs {
		prefillProductOneLiner(&form)
	}
	return form
}

func overlayConfig(form Form, cfg *config.Config) Form {
	if cfg.Project.Name != "" && !config.IsTemplateDefaultProjectName(cfg.Project.Name) {
		form.Answers.ProjectName = cfg.Project.Name
	}
	if cfg.Project.DefaultBranch != "" {
		form.Answers.DefaultBranch = cfg.Project.DefaultBranch
	}
	if cfg.Work.DefaultSpecAgent != "" {
		form.Answers.DefaultSpecAgent = cfg.Work.DefaultSpecAgent
	}
	if cfg.Work.DefaultAgent != "" {
		form.Answers.DefaultAgent = cfg.Work.DefaultAgent
	}
	if cfg.Work.DefaultReviewer != "" {
		form.Answers.DefaultReviewer = cfg.Work.DefaultReviewer
	}
	if cfg.Work.DefaultEnricher != "" {
		form.Answers.DefaultEnricher = cfg.Work.DefaultEnricher
	}
	form.KnownAgentKeys = AgentKeys(cfg)
	return form
}

func advancedFromConfig(cfg *config.Config) AdvancedFields {
	adv := AdvancedFields{
		WorkStopAfter:       cfg.Work.StopAfter,
		VerificationProfile: cfg.Verification.DefaultProfile,
		UITheme:             cfg.UI.Theme,
	}
	if cfg.Budgets.PerRun.MaxEstimatedCost > 0 {
		adv.BudgetMaxCost = fmt.Sprintf("%.2f", cfg.Budgets.PerRun.MaxEstimatedCost)
	}
	if cfg.Coordination.MaxParallelAgents > 0 {
		adv.CoordinationMaxParallel = strconv.Itoa(cfg.Coordination.MaxParallelAgents)
	}
	if cfg.MCP.Enabled {
		adv.MCPEnabled = "true"
	} else {
		adv.MCPEnabled = "false"
	}
	return adv
}

// FieldsMap flattens answers and advanced fields for bus transport.
func (f Form) FieldsMap() map[string]string {
	m := map[string]string{
		"project_name":       f.Answers.ProjectName,
		"default_branch":     f.Answers.DefaultBranch,
		"tagline":            f.Answers.Tagline,
		"stack":              f.Answers.Stack,
		"default_spec_agent": f.Answers.DefaultSpecAgent,
		"default_enricher":   f.Answers.DefaultEnricher,
		"default_agent":      f.Answers.DefaultAgent,
		"default_reviewer":   f.Answers.DefaultReviewer,
		"product_one_liner":  f.Answers.ProductOneLiner,
		"feature_slug":       f.Answers.FeatureSlug,
		"enabled_providers":  f.Answers.EnabledProviders,
	}
	if len(f.KnownAgentKeys) > 0 {
		m["agents_available"] = strings.Join(f.KnownAgentKeys, ", ")
	}
	return m
}

// AdvancedMap flattens advanced panel fields.
func (f Form) AdvancedMap() map[string]string {
	return map[string]string{
		"work_stop_after":            f.Advanced.WorkStopAfter,
		"budget_max_cost":            f.Advanced.BudgetMaxCost,
		"verification_profile":       f.Advanced.VerificationProfile,
		"coordination_max_parallel":  f.Advanced.CoordinationMaxParallel,
		"ui_theme":                   f.Advanced.UITheme,
		"mcp_enabled":                f.Advanced.MCPEnabled,
	}
}

// FormFromMaps rebuilds a Form from bus field maps.
func FormFromMaps(step WizardStep, fields, advanced map[string]string) Form {
	f := Form{Step: step, Errors: map[string]string{}}
	if step == "" {
		f.Step = StepWelcome
	}
	f.Answers = Answers{
		ProjectName:      strings.TrimSpace(fields["project_name"]),
		DefaultBranch:    strings.TrimSpace(fields["default_branch"]),
		Tagline:          strings.TrimSpace(fields["tagline"]),
		Stack:            strings.TrimSpace(fields["stack"]),
		DefaultSpecAgent: strings.TrimSpace(fields["default_spec_agent"]),
		DefaultEnricher:  strings.TrimSpace(fields["default_enricher"]),
		DefaultAgent:     strings.TrimSpace(fields["default_agent"]),
		DefaultReviewer:  strings.TrimSpace(fields["default_reviewer"]),
		ProductOneLiner:  strings.TrimSpace(fields["product_one_liner"]),
		FeatureSlug:      strings.TrimSpace(fields["feature_slug"]),
		EnabledProviders: strings.TrimSpace(fields["enabled_providers"]),
	}
	if v := strings.TrimSpace(fields["agents_available"]); v != "" {
		f.KnownAgentKeys = strings.Split(v, ", ")
	}
	f.Advanced = AdvancedFields{
		WorkStopAfter:           strings.TrimSpace(advanced["work_stop_after"]),
		BudgetMaxCost:           strings.TrimSpace(advanced["budget_max_cost"]),
		VerificationProfile:     strings.TrimSpace(advanced["verification_profile"]),
		CoordinationMaxParallel: strings.TrimSpace(advanced["coordination_max_parallel"]),
		UITheme:                 strings.TrimSpace(advanced["ui_theme"]),
		MCPEnabled:              strings.TrimSpace(advanced["mcp_enabled"]),
	}
	return f
}

// ValidateStep returns field-level errors for one wizard step.
func ValidateStep(step WizardStep, f Form) map[string]string {
	errors := map[string]string{}
	switch step {
	case StepProject:
		if strings.TrimSpace(f.Answers.ProjectName) == "" {
			errors["project_name"] = "nom requis"
		}
		if strings.TrimSpace(f.Answers.DefaultBranch) == "" {
			errors["default_branch"] = "branche requise"
		}
	case StepStack:
		if strings.TrimSpace(f.Answers.Stack) == "" {
			errors["stack"] = "stack requise (go, castor, node…)"
		}
	case StepProviders:
		if len(parseEnabledProvidersCSV(f.Answers.EnabledProviders)) == 0 {
			errors["enabled_providers"] = "au moins un provider (ex. kiro-cli, ollama)"
		}
	case StepAgents:
		known := f.KnownAgentKeys
		for _, spec := range []struct {
			key string
			val string
		}{
			{"default_spec_agent", f.Answers.DefaultSpecAgent},
			{"default_enricher", f.Answers.DefaultEnricher},
			{"default_agent", f.Answers.DefaultAgent},
			{"default_reviewer", f.Answers.DefaultReviewer},
		} {
			if msg := validateAgentRef(spec.val, known); msg != "" {
				errors[spec.key] = msg
			}
		}
	case StepFeature:
		slug := SlugFromName(f.Answers.FeatureSlug)
		if slug == "" {
			errors["feature_slug"] = "slug invalide"
		}
	}
	if advErrs := validateAdvanced(f.Advanced); len(advErrs) > 0 {
		for k, v := range advErrs {
			errors[k] = v
		}
	}
	return errors
}

func validateAdvanced(adv AdvancedFields) map[string]string {
	errors := map[string]string{}
	if v := strings.TrimSpace(adv.BudgetMaxCost); v != "" {
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			errors["budget_max_cost"] = "nombre attendu"
		}
	}
	if v := strings.TrimSpace(adv.CoordinationMaxParallel); v != "" {
		if n, err := strconv.Atoi(v); err != nil || n < 1 {
			errors["coordination_max_parallel"] = "entier >= 1"
		}
	}
	if v := strings.TrimSpace(adv.MCPEnabled); v != "" && v != "true" && v != "false" {
		errors["mcp_enabled"] = "true ou false"
	}
	return errors
}

// AdvanceTUIStep moves the wizard forward/back in the TUI flow with optional validation.
func AdvanceTUIStep(f Form, direction string, validate bool) (Form, error) {
	idx := tuiStepIndex(f.Step)
	if idx < 0 {
		f.Step = StepWelcome
		idx = 0
	}
	if validate && direction == "next" {
		if errs := ValidateStep(f.Step, f); len(errs) > 0 {
			f.Errors = errs
			return f, fmt.Errorf("validation step %s", f.Step)
		}
	}
	switch direction {
	case "next":
		if idx >= len(TUIStepOrder)-1 {
			return f, nil
		}
		f.Step = TUIStepOrder[idx+1]
	case "prev":
		if idx <= 0 {
			return f, nil
		}
		f.Step = TUIStepOrder[idx-1]
	default:
		return f, fmt.Errorf("direction inconnue %q", direction)
	}
	f.Errors = map[string]string{}
	if f.Step == StepDocs {
		prefillProductOneLiner(&f)
	}
	return f, nil
}

// prefillProductOneLiner copies tagline into the product pitch when docs step is empty.
func prefillProductOneLiner(f *Form) {
	if strings.TrimSpace(f.Answers.ProductOneLiner) != "" {
		return
	}
	if t := strings.TrimSpace(f.Answers.Tagline); t != "" {
		f.Answers.ProductOneLiner = t
	}
}

// StateFromForm converts a form to persisted wizard state.
func StateFromForm(f Form) State {
	return State{
		CurrentStep: f.Step,
		Answers:     f.Answers,
	}
}

// OptionsFromForm builds onboard options from a form snapshot.
func OptionsFromForm(f Form) Options {
	return Options{
		Yes:             true,
		NonInteractive:  true,
		Stack:           f.Answers.Stack,
		ProjectName:     f.Answers.ProjectName,
		DefaultBranch:   f.Answers.DefaultBranch,
		Tagline:         f.Answers.Tagline,
		DefaultSpecAgent: f.Answers.DefaultSpecAgent,
		DefaultAgent:     f.Answers.DefaultAgent,
		DefaultReviewer:  f.Answers.DefaultReviewer,
		DefaultEnricher:  f.Answers.DefaultEnricher,
		FeatureSlug:      f.Answers.FeatureSlug,
		ProductOneLiner: f.Answers.ProductOneLiner,
	}
}

// ApplyAdvancedPatch merges advanced wizard fields into config when set.
func ApplyAdvancedPatch(cfg *config.Config, adv AdvancedFields) {
	if cfg == nil {
		return
	}
	if v := strings.TrimSpace(adv.WorkStopAfter); v != "" {
		cfg.Work.StopAfter = v
	}
	if v := strings.TrimSpace(adv.VerificationProfile); v != "" {
		cfg.Verification.DefaultProfile = v
	}
	if v := strings.TrimSpace(adv.UITheme); v != "" {
		cfg.UI.Theme = v
	}
	if v := strings.TrimSpace(adv.BudgetMaxCost); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Budgets.PerRun.MaxEstimatedCost = n
		}
	}
	if v := strings.TrimSpace(adv.CoordinationMaxParallel); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Coordination.MaxParallelAgents = n
		}
	}
	if v := strings.TrimSpace(adv.MCPEnabled); v == "true" {
		cfg.MCP.Enabled = true
	} else if v == "false" {
		cfg.MCP.Enabled = false
	}
}

// ApplyForm runs onboarding using form answers (config + docs + readiness).
func ApplyForm(repoRoot string, f Form) (Result, error) {
	opts := OptionsFromForm(f)
	st := StateFromForm(f)

	matches, validation := detect.DetectAll(repoRoot, firstNonEmpty(opts.Stack, "auto"))
	if len(validation) == 0 {
		validation = config.DefaultGoValidationCommandsForRepo(repoRoot)
	}
	if st.Answers.Stack == "" {
		st.Answers.Stack = detect.PrimaryStack(matches)
	}

	patch := ConfigPatch{
		ProjectName:     st.Answers.ProjectName,
		DefaultBranch:   st.Answers.DefaultBranch,
		BranchPrefix:    SlugFromName(st.Answers.ProjectName),
		DefaultSpecAgent: st.Answers.DefaultSpecAgent,
		DefaultAgent:     st.Answers.DefaultAgent,
		DefaultReviewer:  st.Answers.DefaultReviewer,
		DefaultEnricher:  st.Answers.DefaultEnricher,
		Validation:       validation,
	}

	cfgPath := config.ConfigPath(repoRoot)
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	if cfgErr != nil {
		cfg = config.NewTestConfig(filepath.Base(repoRoot))
	}

	merged, skipped := MergeConfig(cfg, patch, filepath.Base(repoRoot))
	ApplyAdvancedPatch(merged, f.Advanced)

	planned := []PlannedChange{{Path: config.DefaultConfigRel, Action: "update", Summary: "merge config"}}
	docPlanned, docErr := BootstrapDocs(repoRoot, st.Answers, false, false)
	if docErr != nil {
		return Result{}, docErr
	}
	planned = append(planned, docPlanned...)

	backupPath, err := WriteConfig(repoRoot, cfgPath, merged, false)
	if err != nil {
		return Result{}, err
	}
	st.CurrentStep = StepReview
	if err := SaveState(repoRoot, st); err != nil {
		return Result{}, err
	}

	report, err := AssessReadiness(repoRoot, merged, false)
	if err != nil {
		return Result{}, err
	}
	var applied []AppliedAutofix
	if !report.Ready && len(ListAutofixOffers(report.Checks)) > 0 {
		applied, report, err = ApplyReadinessAutofixes(repoRoot)
		if err != nil {
			return Result{}, err
		}
	}
	_ = PersistReport(repoRoot, report)

	return Result{
		Report:           report,
		PlannedChanges:   planned,
		SkippedFields:    skipped,
		AppliedAutofixes: applied,
		ConfigPath:       cfgPath,
		BackupPath:       backupPath,
	}, nil
}

func tuiStepIndex(step WizardStep) int {
	for i, s := range TUIStepOrder {
		if s == step {
			return i
		}
	}
	return -1
}
