package onboarding_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestMergeAnswersUsesLogicalDefaults(t *testing.T) {
	st := onboarding.MergeAnswers(onboarding.State{}, onboarding.Options{}, "/tmp/repo")
	require.Equal(t, config.DefaultAgentSpec, st.Answers.DefaultSpecAgent)
	require.Equal(t, config.DefaultAgentDev, st.Answers.DefaultAgent)
	require.Equal(t, config.DefaultAgentReviewer, st.Answers.DefaultReviewer)
	require.Equal(t, config.DefaultAgentEnrich, st.Answers.DefaultEnricher)
}

func TestCheckAgentRefMissingLogicalAgent(t *testing.T) {
	cfg := config.NewTestConfig("demo")
	delete(cfg.Agents, config.DefaultAgentDev)
	cfg.Work.DefaultAgent = config.DefaultAgentDev

	checks := onboarding.RunDoctorChecks(t.TempDir(), cfg, onboarding.DoctorOpts{Full: true, SkipExec: true})
	var found bool
	for _, c := range checks {
		if c.ID == "agents.dev" {
			found = true
			require.Contains(t, c.Message, `Agent logique "dev" absent de config.agents`)
			require.Contains(t, c.Message, "work.default_agent")
		}
	}
	require.True(t, found)
}
