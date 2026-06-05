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
			DefaultSpecAgent: "kiro",
			DefaultEnricher:  "ollama",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
		},
		KnownAgentKeys: []string{"kiro", "cursor", "codex", "ollama"},
	}
	require.Empty(t, onboarding.ValidateStep(onboarding.StepAgents, form))

	form.Answers.DefaultAgent = "unknown"
	errs := onboarding.ValidateStep(onboarding.StepAgents, form)
	require.Contains(t, errs, "default_agent")
}

func TestAgentKeysFromConfig(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Agents["cursor"] = config.Agent{Command: "cursor"}
	cfg.Agents["codex"] = config.Agent{Command: "codex"}
	keys := onboarding.AgentKeys(cfg)
	require.Equal(t, []string{"codex", "cursor"}, keys)
}
