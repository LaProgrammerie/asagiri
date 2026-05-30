package onboarding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestApplyGitignoreAutofixCreatesFile(t *testing.T) {
	repo := t.TempDir()
	applied, report, err := onboarding.ApplyReadinessAutofixes(repo)
	require.NoError(t, err)
	require.Len(t, applied, 1)
	require.Equal(t, ".gitignore", applied[0].Path)
	require.Contains(t, applied[0].AddedLines, ".asagiri/state.sqlite")
	require.Contains(t, applied[0].AddedLines, ".asagiri/worktrees/")

	data, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	require.NoError(t, err)
	require.Contains(t, string(data), ".asagiri/state.sqlite")
	require.Contains(t, string(data), ".asagiri/worktrees/")

	checks := onboarding.RunDoctorChecks(repo, nil, onboarding.DoctorOpts{Full: true, SkipExec: true})
	for _, c := range checks {
		if c.ID == "gitignore._asagiri_state.sqlite" || c.ID == "gitignore.worktrees_" {
			require.Equal(t, onboarding.StatusOK, c.Status, c.ID)
		}
	}
	require.GreaterOrEqual(t, report.Score, 50)
}

func TestApplyGitignoreAutofixAppendsMissingLines(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".asagiri/state.sqlite\n"), 0o644))

	applied, _, err := onboarding.ApplyReadinessAutofixes(repo)
	require.NoError(t, err)
	require.Len(t, applied, 1)
	require.Equal(t, []string{".asagiri/worktrees/"}, applied[0].AddedLines)

	data, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	require.NoError(t, err)
	require.Contains(t, string(data), ".asagiri/worktrees/")
}

func TestApplyGitignoreAutofixIdempotent(t *testing.T) {
	repo := t.TempDir()
	content := ".asagiri/state.sqlite\nworktrees/\n"
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(content), 0o644))

	applied, _, err := onboarding.ApplyReadinessAutofixes(repo)
	require.NoError(t, err)
	require.Empty(t, applied)
}

func TestListAutofixOffersGitignore(t *testing.T) {
	checks := []onboarding.Check{
		{ID: "gitignore._asagiri_state.sqlite", Status: onboarding.StatusFail},
		{ID: "agents.cursor", Status: onboarding.StatusWarn},
	}
	offers := onboarding.ListAutofixOffers(checks)
	require.Len(t, offers, 1)
	require.Equal(t, "gitignore.asagiri", offers[0].ID)
}
