package onboarding_test

import (
	"strings"
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
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

func TestShellLayoutRendersThroughSharedShell(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{Step: onbdomain.StepWelcome}, false)
	got := stripANSI(onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
		Shell: onboarding.ShellContext{
			Workspace:   "chatbot",
			Branch:      "main",
			Version:     "0.4.0",
			APIProvider: "OpenRouter",
			Online:      true,
		},
	}))

	require.Contains(t, got, "Project Onboarding")
	require.Contains(t, got, "NAVIGATION")
	require.Contains(t, got, "Onboarding")
	require.Contains(t, got, "Bienvenue dans Asagiri")
	require.Contains(t, got, "CE QUE NOUS ALLONS FAIRE")
	require.Contains(t, got, "Précédent")
	require.Contains(t, got, "Suivant")
	require.Equal(t, 2, strings.Count(got, "ASAGIRI"))

	// CK-4.2: no fake telemetry leaks into the shared shell.
	require.NotContains(t, got, "Onboarding Analyzer")
	require.NotContains(t, got, "Analyzing")
	require.NotContains(t, got, "Confiance")
	require.NotContains(t, got, "API:")
}

func TestShellLayoutWidthMatchesTerminal(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{Step: onbdomain.StepWelcome}, false)
	const termW = 120
	got := onboarding.Render(onboarding.ViewModel{
		Model: m, WizardMode: true, FullScreen: true,
		Width: termW, Height: 40, Theme: theme.Default(),
		Shell: onboarding.ShellContext{Workspace: "chatbot", Branch: "main"},
	})
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	require.NotEmpty(t, lines)
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w > termW {
			t.Fatalf("line %d width %d exceeds terminal %d: %q", i, w, termW, line)
		}
	}
}

func TestShellLayoutStepProject(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "chatbot",
			DefaultBranch: "main",
		},
	}, false)
	got := stripANSI(onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
		Shell:      onboarding.ShellContext{Workspace: "chatbot", Branch: "main"},
	}))
	require.Contains(t, got, "chatbot")
	require.Contains(t, got, "Étape 2 / 7")
	require.Contains(t, got, "Nom")
	require.NotContains(t, got, "PROJET")
}

func TestShellLayoutStepDocsNoDuplicateStepTitle(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepDocs,
		Answers: onbdomain.Answers{
			ProductOneLiner: "Chatbot",
		},
	}, false)
	got := stripANSI(onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
		Shell:      onboarding.ShellContext{Workspace: "chatbot", Branch: "main"},
	}))
	require.Contains(t, got, "Phrase produit")
	require.Contains(t, got, "Chatbot")
	require.Contains(t, got, "étape Projet")
	require.NotContains(t, got, "DOCS")
}
