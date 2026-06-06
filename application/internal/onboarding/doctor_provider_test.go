package onboarding_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestCheckAgentRefProviderNotInPATH(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"kiro-cli": {
				Type:    config.ProviderTypeKiroCLI,
				Command: "definitely-not-a-real-binary-asa-test",
			},
		},
		Agents: map[string]config.Agent{
			"laprogrammerie": {Provider: "kiro-cli"},
		},
		Work: config.WorkConfig{DefaultSpecAgent: "laprogrammerie"},
	}

	checks := onboarding.RunDoctorChecks(t.TempDir(), cfg, onboarding.DoctorOpts{Full: true, SkipExec: false})
	var found bool
	for _, c := range checks {
		if c.ID == "providers.kiro-cli" {
			found = true
			require.Equal(t, onboarding.StatusWarn, c.Status)
			require.Contains(t, c.Message, "Provider \"kiro-cli\" introuvable dans PATH")
			require.Contains(t, c.Message, "laprogrammerie")
		}
	}
	require.True(t, found, "expected providers.kiro-cli check")
}

func TestCheckAgentRefMissingProviderSection(t *testing.T) {
	cfg := config.NewTestConfig("demo")
	cfg.Agents["dev"] = config.Agent{Provider: "missing-provider"}
	cfg.Work.DefaultAgent = "dev"

	checks := onboarding.RunDoctorChecks(t.TempDir(), cfg, onboarding.DoctorOpts{Full: true, SkipExec: true})
	var found bool
	for _, c := range checks {
		if c.ID == "providers.missing-provider" {
			found = true
			require.Contains(t, c.Message, "provider \"missing-provider\" introuvable")
		}
	}
	require.True(t, found)
}

func TestValidateStepProvidersRequiresSelection(t *testing.T) {
	errs := onboarding.ValidateStep(onboarding.StepProviders, onboarding.Form{
		Step: onboarding.StepProviders,
	})
	require.Contains(t, errs, "enabled_providers")

	errs = onboarding.ValidateStep(onboarding.StepProviders, onboarding.Form{
		Step:    onboarding.StepProviders,
		Answers: onboarding.Answers{EnabledProviders: "kiro-cli, ollama"},
	})
	require.Empty(t, errs)
}
