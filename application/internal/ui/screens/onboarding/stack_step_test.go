package onboarding_test

import (
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestStackStepUsesSelectWithoutReadonlyPills(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepStack,
		Answers: onbdomain.Answers{
			Stack: "castor",
		},
		DetectedStacks:    []string{"castor"},
		ValidationPreview: []string{"static-checks: make lint", "phpunit: vendor/bin/phpunit"},
	}, false)
	m.FocusFooter = onboarding.FooterNext // unfocused: plain value + ▾
	got := onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
	})
	require.Contains(t, got, "castor ▾")
	require.Contains(t, got, "static-checks")
}

func TestModelSelectCyclesStackChoice(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepStack,
		Answers: onbdomain.Answers{
			Stack: "auto",
		},
	}, false)
	m.RefreshFieldRows()
	m.FocusField = 0

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	require.Equal(t, "go", next.Fields["stack"])

	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.Equal(t, "castor", next.Fields["stack"])
}
