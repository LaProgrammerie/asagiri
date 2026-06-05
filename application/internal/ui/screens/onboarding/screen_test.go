package onboarding_test

import (
	"strings"
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderWizardScreen(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "chatbot",
			DefaultBranch: "main",
		},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m, ShowCLI: true})
	require.Contains(t, got, "Project Onboarding Wizard")
	require.Contains(t, got, "chatbot")
	require.Contains(t, got, "Next →")
	require.Contains(t, got, "asa onboard")
}

func TestRenderReadySummary(t *testing.T) {
	m := onboarding.NewModel()
	m.Applied = true
	got := onboarding.Render(onboarding.ViewModel{
		Model: m,
		Readiness: bus.ReadinessResult{
			Ready: true,
			Score: 88,
		},
	})
	require.Contains(t, got, "READY")
	require.Contains(t, got, "88/100")
}

func TestRenderReadySummaryWithAutofixOffers(t *testing.T) {
	m := onboarding.NewModel()
	m.Applied = true
	got := onboarding.Render(onboarding.ViewModel{
		Model: m,
		Readiness: bus.ReadinessResult{
			Ready: false,
			Score: 55,
			AutofixOffers: []bus.AutofixOffer{{
				ID:    "gitignore.asagiri",
				Title: ".gitignore — entrées Asagiri",
				Lines: []string{".asagiri/state.sqlite", ".asagiri/worktrees/"},
			}},
		},
		WizardMode: true,
		FullScreen: true,
		Width:      140,
		Height:     40,
		Theme:      theme.Default(),
		Shell:      onboarding.ShellContext{Workspace: "chatbot"},
	})
	require.Contains(t, got, "Corrections auto restantes")
	require.Contains(t, got, "O · appliquer")
}

func TestModelKeyboardAdvanceStep(t *testing.T) {
	form := onbdomain.Form{
		Step: onbdomain.StepWelcome,
		Answers: onbdomain.Answers{
			ProjectName:   "bot",
			DefaultBranch: "main",
		},
	}
	m := onboarding.NewModelFromForm(form, false)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.Nil(t, cmd)
	require.Equal(t, onbdomain.StepWelcome, next.Step)

	next, cmd = next.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	require.NotNil(t, cmd)
	msg := cmd()
	footer, ok := msg.(onboarding.OnboardingFooterMsg)
	require.True(t, ok)
	require.Equal(t, onboarding.FooterNext, footer.Button)
}

func TestModelTabCyclesFields(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "a",
			DefaultBranch: "main",
		},
	}, false)
	require.Equal(t, 0, m.FocusField)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, 1, next.FocusField)
}

func TestModelAdvancedToggle(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{Step: onbdomain.StepProject}, false)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	require.True(t, next.ShowAdvanced)
	rendered := onboarding.Render(onboarding.ViewModel{Model: next})
	require.True(t, strings.Contains(rendered, "work.stop_after"))
}

func TestModelLetterNInFieldDoesNotAdvanceStep(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "",
			DefaultBranch: "main",
		},
	}, false)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.Nil(t, cmd)
	require.Equal(t, "n", next.Fields["project_name"])
}

func TestModelTypingQInFieldDoesNotQuit(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "",
			DefaultBranch: "main",
		},
	}, false)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.Nil(t, cmd)
	require.Equal(t, "q", next.Fields["project_name"])
}

func TestModelCtrlQWhenAppliedQuits(t *testing.T) {
	m := onboarding.NewModel()
	m.Applied = true
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	require.NotNil(t, cmd)
	require.True(t, next.Applied)
}

func TestModelSpaceInTagline(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "chatbot",
			DefaultBranch: "main",
		},
	}, false)
	m.FocusField = 2 // tagline
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	require.Nil(t, cmd)
	require.Equal(t, " ", next.Fields["tagline"])
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeySpace})
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	require.Equal(t, " h i", next.Fields["tagline"])
}

func TestModelTypingInFieldDoesNotToggleAdvanced(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName:   "",
			DefaultBranch: "main",
		},
	}, false)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	require.False(t, next.ShowAdvanced)
	require.Equal(t, "a", next.Fields["project_name"])
}

func TestRenderWizardFullscreen(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepProject,
		Answers: onbdomain.Answers{
			ProjectName: "chatbot",
		},
	}, true)
	got := onboarding.Render(onboarding.ViewModel{
		Model:      m,
		WizardMode: true,
		FullScreen: true,
		Width:      120,
		Height:     40,
		Theme:      theme.Default(),
		Shell:      onboarding.ShellContext{Workspace: "chatbot"},
	})
	require.Contains(t, got, "Project Onboarding")
	require.Contains(t, got, "chatbot")
	require.NotContains(t, got, "Screen: onboarding")
}

func TestExistingConfigDetectedInWelcomeStep(t *testing.T) {
	// HasAsagiriConfig comes from the domain Form — simulate a repo with existing config
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:             onbdomain.StepWelcome,
		HasAsagiriConfig: true,
		Answers: onbdomain.Answers{
			ProjectName:   "myapp",
			DefaultBranch: "main",
		},
	}, false)
	require.True(t, m.ExistingConfig, "ExistingConfig should mirror Form.HasAsagiriConfig")
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	require.Contains(t, got, "détectée")
}

func TestProjectStepShowsDetectionBlock(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:           onbdomain.StepProject,
		DetectedStacks: []string{"go", "node"},
		Answers: onbdomain.Answers{
			ProjectName:   "chatbot",
			DefaultBranch: "main",
		},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	require.Contains(t, got, "DÉTECTÉ")
	require.Contains(t, got, "go")
}

func TestStackStepShowsCapabilities(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:              onbdomain.StepStack,
		DetectedStacks:    []string{"go"},
		ValidationPreview: []string{"go test: go test ./... (required)"},
		Answers:           onbdomain.Answers{Stack: "go"},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m, WizardMode: true, FullScreen: true, Width: 120, Height: 40, Theme: theme.Default()})
	require.Contains(t, got, "VALIDATIONS ACTIVÉES")
	require.Contains(t, got, "CAPACITÉS ACTIVÉES")
}

func TestAgentsStepShowsPipeline(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepAgents,
		Answers: onbdomain.Answers{
			DefaultSpecAgent: "kiro",
			DefaultEnricher:  "ollama",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
		},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m, WizardMode: true, FullScreen: true, Width: 120, Height: 40, Theme: theme.Default()})
	require.Contains(t, got, "PIPELINE GÉNÉRÉ")
	require.Contains(t, got, "kiro")
	require.Contains(t, got, "cursor")
}

func TestFeatureStepShowsPathPreview(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:    onbdomain.StepFeature,
		Answers: onbdomain.Answers{FeatureSlug: "stripe-mvp"},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m, WizardMode: true, FullScreen: true, Width: 120, Height: 40, Theme: theme.Default()})
	require.Contains(t, got, ".kiro/specs/stripe-mvp/")
}

func TestReviewStepShowsArtefacts(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:    onbdomain.StepReview,
		Answers: onbdomain.Answers{ProjectName: "chatbot", DefaultBranch: "main", Stack: "go", FeatureSlug: "chatbot-mvp"},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m, WizardMode: true, FullScreen: true, Width: 120, Height: 40, Theme: theme.Default()})
	require.Contains(t, got, "CE QUI VA ÊTRE CRÉÉ")
	require.Contains(t, got, ".asagiri/config.yaml")
	require.Contains(t, got, ".kiro/specs/chatbot-mvp/")
}

func TestAdvancedButtonLabelShowsCount(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{Step: onbdomain.StepProject}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	require.Contains(t, got, "Mode expert")
	require.Contains(t, got, "6")
}

func TestDocsStepShowsDocsAIFiles(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:    onbdomain.StepDocs,
		Answers: onbdomain.Answers{ProductOneLiner: "AI coding assistant"},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	require.Contains(t, got, "docs/ai/01-product.md")
	require.Contains(t, got, "FICHIERS QUI SERONT CRÉÉS")
}

func TestFeatureStepShowsWorkflowSteps(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step:    onbdomain.StepFeature,
		Answers: onbdomain.Answers{FeatureSlug: "stripe-mvp"},
	}, false)
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	require.Contains(t, got, "WORKFLOW SUIVANT")
	require.Contains(t, got, "asa work")
	require.Contains(t, got, "stripe-mvp")
	require.Contains(t, got, ".kiro/specs/stripe-mvp/")
}

func TestNoExistingConfigShowsNeutralWelcome(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepWelcome,
		Answers: onbdomain.Answers{ProjectName: "fresh", DefaultBranch: "main"},
		// HasAsagiriConfig: false (zero value)
	}, false)
	require.False(t, m.ExistingConfig)
	got := onboarding.Render(onboarding.ViewModel{Model: m})
	// Neutral message — no "existante détectée"
	require.NotContains(t, got, "existante détectée")
}
