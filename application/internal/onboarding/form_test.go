package onboarding_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestValidateStepProject(t *testing.T) {
	form := onboarding.Form{Step: onboarding.StepProject}
	errs := onboarding.ValidateStep(onboarding.StepProject, form)
	require.Contains(t, errs, "project_name")
	require.Contains(t, errs, "default_branch")

	form.Answers.ProjectName = "chatbot"
	form.Answers.DefaultBranch = "main"
	errs = onboarding.ValidateStep(onboarding.StepProject, form)
	require.Empty(t, errs)
}

func TestAdvanceTUIStepValidation(t *testing.T) {
	form := onboarding.Form{Step: onboarding.StepProject, Errors: map[string]string{}}
	next, err := onboarding.AdvanceTUIStep(form, "next", true)
	require.Error(t, err)
	require.Equal(t, onboarding.StepProject, next.Step)
	require.NotEmpty(t, next.Errors)

	form.Answers.ProjectName = "chatbot"
	form.Answers.DefaultBranch = "main"
	next, err = onboarding.AdvanceTUIStep(form, "next", true)
	require.NoError(t, err)
	require.Equal(t, onboarding.StepStack, next.Step)
}

func TestAdvanceTUIStepPrev(t *testing.T) {
	form := onboarding.Form{Step: onboarding.StepStack}
	prev, err := onboarding.AdvanceTUIStep(form, "prev", false)
	require.NoError(t, err)
	require.Equal(t, onboarding.StepProject, prev.Step)
}

func TestAdvanceTUIStepPrefillsProductFromTagline(t *testing.T) {
	form := onboarding.Form{
		Step: onboarding.StepAgents,
		Answers: onboarding.Answers{
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
			Tagline:          "Assistant conversationnel pour le support",
		},
	}
	next, err := onboarding.AdvanceTUIStep(form, "next", false)
	require.NoError(t, err)
	require.Equal(t, onboarding.StepDocs, next.Step)
	require.Equal(t, "Assistant conversationnel pour le support", next.Answers.ProductOneLiner)
}

func TestBuildFormPrefillFromConfig(t *testing.T) {
	repo := t.TempDir()
	cfg := config.NewTestConfig("my-project")
	cfg.Project.Name = "chatbot-php"
	cfg.Project.DefaultBranch = "develop"
	cfg.Work.DefaultSpecAgent = "kiro"
	cfg.Work.DefaultAgent = "cursor"
	cfg.Work.DefaultReviewer = "codex"
	cfg.Work.DefaultEnricher = "ollama"
	cfg.Agents["kiro"] = config.Agent{Command: "kiro"}
	cfg.Agents["cursor"] = config.Agent{Command: "cursor"}
	cfg.Agents["codex"] = config.Agent{Command: "codex"}
	cfg.Agents["ollama"] = config.Agent{Command: "ollama"}
	cfg.Verification.DefaultProfile = "staging"
	cfg.UI.Theme = "asagiri-dark"
	cfg.Budgets.PerRun.MaxEstimatedCost = 2.5
	cfg.Coordination.MaxParallelAgents = 3
	cfg.MCP.Enabled = true

	form := onboarding.BuildForm(repo, onboarding.State{}, cfg)
	require.Equal(t, "chatbot-php", form.Answers.ProjectName)
	require.Equal(t, "develop", form.Answers.DefaultBranch)
	require.Equal(t, "kiro", form.Answers.DefaultSpecAgent)
	require.Equal(t, "cursor", form.Answers.DefaultAgent)
	require.Equal(t, "ollama", form.Answers.DefaultEnricher)
	require.Contains(t, form.FieldsMap()["agents_available"], "cursor")
	require.Equal(t, "staging", form.Advanced.VerificationProfile)
	require.Equal(t, "2.50", form.Advanced.BudgetMaxCost)
	require.Equal(t, "3", form.Advanced.CoordinationMaxParallel)
	require.Equal(t, "true", form.Advanced.MCPEnabled)
}

func TestFormFromMapsRoundTrip(t *testing.T) {
	fields := map[string]string{
		"project_name":   "bot",
		"default_branch": "main",
		"stack":          "castor",
	}
	advanced := map[string]string{"ui_theme": "asagiri-dark"}
	form := onboarding.FormFromMaps(onboarding.StepAgents, fields, advanced)
	require.Equal(t, "bot", form.Answers.ProjectName)
	require.Equal(t, "castor", form.Answers.Stack)
	require.Equal(t, "asagiri-dark", form.Advanced.UITheme)
}

func TestValidateAdvancedBudget(t *testing.T) {
	form := onboarding.Form{
		Advanced: onboarding.AdvancedFields{BudgetMaxCost: "not-a-number"},
	}
	errs := onboarding.ValidateStep(onboarding.StepReview, form)
	require.Contains(t, errs, "budget_max_cost")
}
