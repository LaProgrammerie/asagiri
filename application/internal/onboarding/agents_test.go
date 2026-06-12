package onboarding_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestValidateStepAgentsRequiresKnownKeys(t *testing.T) {
	form := onboarding.Form{
		Step: onboarding.StepAgents,
		Answers: onboarding.Answers{
			DefaultSpecAgent: config.DefaultAgentSpec,
			DefaultEnricher:  config.DefaultAgentEnrich,
			DefaultAgent:     config.DefaultAgentDev,
			DefaultReviewer:  config.DefaultAgentReviewer,
		},
		KnownAgentKeys: config.DefaultLogicalAgentIDs(),
	}
	require.Empty(t, onboarding.ValidateStep(onboarding.StepAgents, form))

	form.Answers.DefaultAgent = "unknown"
	errs := onboarding.ValidateStep(onboarding.StepAgents, form)
	require.Contains(t, errs, "default_agent")
}

func TestAgentKeysFromConfig(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Agents = map[string]config.Agent{
		"cursor": {Command: "cursor"},
		"codex":  {Command: "codex"},
	}
	keys := onboarding.AgentKeys(cfg)
	require.Equal(t, []string{"codex", "cursor"}, keys)
}
