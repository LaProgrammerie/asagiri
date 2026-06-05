package bus

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
)

func (b *queryBus) handleGetReadiness(_ context.Context, q GetReadinessQuery) (QueryResult, error) {
	cfg := b.deps.Config
	if cfg == nil {
		var err error
		cfg, err = config.Load(config.ConfigPath(b.deps.RepoRoot), b.deps.RepoRoot)
		if err != nil {
			cfg = nil
		}
	}
	report, err := onboarding.AssessReadiness(b.deps.RepoRoot, cfg, q.Strict)
	if err != nil {
		return ReadinessResult{}, err
	}
	return toReadinessResult(report), nil
}

func (b *queryBus) handleGetOnboardingState(_ context.Context, _ GetOnboardingStateQuery) (QueryResult, error) {
	st, err := onboarding.LoadState(b.deps.RepoRoot)
	if err != nil {
		return OnboardingStateResult{}, err
	}
	return OnboardingStateResult{
		CurrentStep: string(st.CurrentStep),
		Answers:     onboardingStateAnswers(st),
		Completed:   st.Completed,
	}, nil
}

func (b *queryBus) handleGetOnboardingWizard(_ context.Context, _ GetOnboardingWizardQuery) (QueryResult, error) {
	st, err := onboarding.LoadState(b.deps.RepoRoot)
	if err != nil {
		return OnboardingWizardResult{}, err
	}
	cfg := b.deps.Config
	if cfg == nil {
		cfg, err = config.Load(config.ConfigPath(b.deps.RepoRoot), b.deps.RepoRoot)
		if err != nil {
			cfg = nil
		}
	}
	form := onboarding.BuildForm(b.deps.RepoRoot, st, cfg)
	steps := make([]string, 0, len(onboarding.TUIStepOrder))
	for _, s := range onboarding.TUIStepOrder {
		steps = append(steps, string(s))
	}
	return OnboardingWizardResult{
		CurrentStep:       string(form.Step),
		Steps:             steps,
		Fields:            form.FieldsMap(),
		Advanced:          form.AdvancedMap(),
		ValidationPreview: form.ValidationPreview,
		DetectedStacks:    form.DetectedStacks,
		Errors:            form.Errors,
		SkippedFields:     form.SkippedFields,
	}, nil
}

func (b *queryBus) handleValidateOnboardingStep(_ context.Context, q ValidateOnboardingStepQuery) (QueryResult, error) {
	step, err := onboarding.ParseStep(strings.TrimSpace(q.Step))
	if err != nil {
		step = onboarding.WizardStep(q.Step)
	}
	form := onboarding.FormFromMaps(step, q.Fields, q.Advanced)
	errs := onboarding.ValidateStep(step, form)
	return ValidateOnboardingStepResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}, nil
}

func onboardingStateAnswers(st onboarding.State) map[string]string {
	return map[string]string{
		"project_name":      st.Answers.ProjectName,
		"default_branch":    st.Answers.DefaultBranch,
		"tagline":           st.Answers.Tagline,
		"stack":             st.Answers.Stack,
		"default_spec_agent": st.Answers.DefaultSpecAgent,
		"default_enricher":   st.Answers.DefaultEnricher,
		"default_agent":      st.Answers.DefaultAgent,
		"default_reviewer":   st.Answers.DefaultReviewer,
		"feature_slug":      st.Answers.FeatureSlug,
		"product_one_liner": st.Answers.ProductOneLiner,
	}
}

func toReadinessResult(report onboarding.Report) ReadinessResult {
	checks := make([]ReadinessCheck, 0, len(report.Checks))
	for _, c := range report.Checks {
		checks = append(checks, ReadinessCheck{
			ID:      c.ID,
			Status:  string(c.Status),
			Message: c.Message,
			FixCLI:  c.FixCLI,
		})
	}
	actions := make([]ReadinessAction, 0, len(report.NextActions))
	for _, a := range report.NextActions {
		actions = append(actions, ReadinessAction{Title: a.Title, CLI: a.CLI})
	}
	offers := make([]AutofixOffer, 0, len(onboarding.ListAutofixOffers(report.Checks)))
	for _, o := range onboarding.ListAutofixOffers(report.Checks) {
		offers = append(offers, AutofixOffer{
			ID:          o.ID,
			Title:       o.Title,
			Description: o.Description,
			Lines:       o.Lines,
		})
	}
	return ReadinessResult{
		Ready:         report.Ready,
		Score:         report.Score,
		Checks:        checks,
		NextActions:   actions,
		AutofixOffers: offers,
	}
}

func dispatchRunOnboardingStep(ctx context.Context, deps Deps, cmd RunOnboardingStepCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	opts := onboarding.Options{
		Yes:            true,
		NonInteractive: true,
		Step:           cmd.Step,
	}
	_, err := onboarding.Onboard(deps.RepoRoot, opts, strings.NewReader(""), io.Discard)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       "Onboarding step exécuté",
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchAdvanceOnboardingStep(ctx context.Context, deps Deps, cmd AdvanceOnboardingStepCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	st, err := onboarding.LoadState(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, err
	}
	form := onboarding.FormFromMaps(st.CurrentStep, cmd.Fields, cmd.Advanced)
	form.Step = st.CurrentStep
	validate := strings.EqualFold(cmd.Direction, "next")
	next, err := onboarding.AdvanceTUIStep(form, strings.ToLower(strings.TrimSpace(cmd.Direction)), validate)
	if err != nil {
		return CommandResult{
			Accepted:      false,
			Message:       "Validation échouée",
			CLIEquivalent: cmd.CLIEquivalent(),
		}, err
	}
	st = onboarding.StateFromForm(next)
	if err := onboarding.SaveState(deps.RepoRoot, st); err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       "Étape: " + string(next.Step),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchSetOnboardingField(ctx context.Context, deps Deps, cmd SetOnboardingFieldCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	st, err := onboarding.LoadState(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, err
	}
	field := strings.TrimSpace(cmd.Field)
	value := cmd.Value
	switch field {
	case "project_name":
		st.Answers.ProjectName = value
	case "default_branch":
		st.Answers.DefaultBranch = value
	case "tagline":
		st.Answers.Tagline = value
	case "stack":
		st.Answers.Stack = value
	case "default_spec_agent":
		st.Answers.DefaultSpecAgent = value
	case "default_enricher":
		st.Answers.DefaultEnricher = value
	case "default_agent":
		st.Answers.DefaultAgent = value
	case "default_reviewer":
		st.Answers.DefaultReviewer = value
	case "feature_slug":
		st.Answers.FeatureSlug = value
	case "product_one_liner":
		st.Answers.ProductOneLiner = value
	default:
		return CommandResult{Accepted: false, Message: "champ inconnu: " + field}, nil
	}
	if err := onboarding.SaveState(deps.RepoRoot, st); err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       "Champ mis à jour",
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchApplyOnboardingConfig(ctx context.Context, deps Deps, cmd ApplyOnboardingConfigCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	if len(cmd.Fields) > 0 {
		step := onboarding.StepReview
		form := onboarding.FormFromMaps(step, cmd.Fields, cmd.Advanced)
		res, err := onboarding.ApplyForm(deps.RepoRoot, form)
		if err != nil {
			return CommandResult{}, err
		}
		msg := "Configuration appliquée"
		if len(res.AppliedAutofixes) > 0 {
			msg += fmt.Sprintf(" — %d correction(s) auto (.gitignore…)", len(res.AppliedAutofixes))
		}
		if res.Report.Ready {
			msg += " — READY"
		} else {
			msg += " — NOT READY"
			if len(onboarding.ListAutofixOffers(res.Report.Checks)) > 0 {
				msg += " — corrections restantes (O ou Ctrl+F)"
			}
		}
		return CommandResult{
			Accepted:      true,
			Message:       msg,
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	stack := strings.TrimSpace(cmd.Stack)
	if stack == "" {
		stack = "auto"
	}
	opts := onboarding.Options{
		Yes:            true,
		NonInteractive: true,
		Stack:          stack,
	}
	_, err := onboarding.Onboard(deps.RepoRoot, opts, strings.NewReader(""), os.Stdout)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       "Configuration onboarding appliquée",
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchSkipOnboardingCheck(ctx context.Context, _ Deps, cmd SkipOnboardingCheckCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	_ = strings.TrimSpace(cmd.CheckID)
	return CommandResult{
		Accepted:      true,
		Message:       "Check ignoré (UI)",
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchApplyReadinessAutofix(ctx context.Context, deps Deps, cmd ApplyReadinessAutofixCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	applied, report, err := onboarding.ApplyReadinessAutofixes(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, err
	}
	if len(applied) == 0 {
		return CommandResult{
			Accepted:      true,
			Message:       "Aucune correction automatique à appliquer",
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	status := "NOT READY"
	if report.Ready {
		status = "READY"
	}
	var parts []string
	for _, fix := range applied {
		if fix.Description != "" {
			parts = append(parts, fix.Description)
		}
	}
	msg := fmt.Sprintf("Corrections appliquées — %s (%d/100)", status, report.Score)
	if len(parts) > 0 {
		msg += " — " + strings.Join(parts, "; ")
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}
