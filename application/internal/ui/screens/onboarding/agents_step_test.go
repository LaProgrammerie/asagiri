package onboarding_test

import (
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAgentsStepShowsManagedPipelineRows(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepAgents,
		Answers: onbdomain.Answers{
			DefaultSpecAgent: "kiro",
			DefaultEnricher:  "ollama",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
		},
	}, false)
	m.AgentChoices = []string{"kiro", "cursor", "codex", "ollama"}
	got := onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
	})
	require.Contains(t, got, "Géré par Asagiri")
	require.Contains(t, got, "Plan")
	require.Contains(t, got, "Verify")
	require.Contains(t, got, "▾")
}

func TestModelSelectCyclesAgentChoice(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepAgents,
		Answers: onbdomain.Answers{
			DefaultSpecAgent: "kiro",
			DefaultEnricher:  "ollama",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
		},
	}, false)
	m.AgentChoices = []string{"kiro", "cursor", "codex", "ollama"}
	m.RefreshFieldRows()
	m.FocusField = 0 // default_spec_agent

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	require.Equal(t, "cursor", next.Fields["default_spec_agent"])

	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	require.Equal(t, "ollama", next.Fields["default_spec_agent"])
}

