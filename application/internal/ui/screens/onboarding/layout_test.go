package onboarding_test

import (
	"strings"
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func stripANSI(v string) string {
	var out []rune
	in := false
	for _, r := range v {
		if r == '\x1b' {
			in = true
			continue
		}
		if in {
			if r == 'm' {
				in = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func TestWizardWelcomeRendersIntactBorders(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{Step: onbdomain.StepWelcome}, false)
	got := stripANSI(onboarding.Render(onboarding.ViewModel{
		Model:      m,
		FullScreen: true,
		Width:      100,
		Height:     30,
		Theme:      theme.Default(),
	}))

	require.Contains(t, got, "╭")
	require.Contains(t, got, "╯")
	require.NotContains(t, got, "╭╭")
	require.Equal(t, 1, strings.Count(got, "ASAGIRI"))

	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Bienvenue") && strings.Contains(line, "│") {
			require.True(t, strings.HasPrefix(strings.TrimLeft(line, " "), "│") || strings.Contains(line, "│ Bienvenue"),
				"unexpected panel line layout: %q", line)
		}
	}
}

func TestWizardPanelAlignsWithOuterFrame(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepReview,
		Answers: onbdomain.Answers{
			ProjectName: "chatbot",
		},
	}, false)
	got := stripANSI(onboarding.Render(onboarding.ViewModel{
		Model:      m,
		FullScreen: true,
		Width:      100,
		Height:     30,
		Theme:      theme.Default(),
	}))

	var widths []int
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "╮") || strings.Contains(line, "╯") {
			widths = append(widths, len([]rune(line)))
		}
	}
	require.NotEmpty(t, widths)
	first := widths[0]
	for _, w := range widths {
		require.InDelta(t, first, w, 2, "border lines should share width")
	}
}
