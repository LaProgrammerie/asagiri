package onboarding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestApplyFormAutofixGitignore(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module demo\n\ngo 1.25.0\n")

	form := onboarding.Form{
		Step: onboarding.StepReview,
		Answers: onboarding.Answers{
			ProjectName:      "demo",
			DefaultBranch:    "main",
			Stack:            "go",
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
			FeatureSlug:      "demo-mvp",
			ProductOneLiner:  "Demo",
		},
	}
	res, err := onboarding.ApplyForm(repo, form)
	require.NoError(t, err)
	require.NotEmpty(t, res.AppliedAutofixes)
	data, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	require.NoError(t, err)
	require.Contains(t, string(data), ".asagiri/state.sqlite")
}
