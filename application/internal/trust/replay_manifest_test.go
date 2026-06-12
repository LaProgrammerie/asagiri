package trust

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestBuildReplayManifest(t *testing.T) {
	repo := t.TempDir()
	runGitInitCommit(t, repo)

	m := buildReplayManifest(repo, VerificationScope{
		TrustID: "trust-2026-05-29-abc",
		Flow:    "onboarding",
		Branch:  "main",
	}, []VerificationCheck{
		{Type: CheckContracts},
		{Type: CheckFlows},
	}, &config.Config{
		Validation: config.ValidationConfig{
			Commands: []config.ValidationCommand{{Command: "go test ./..."}},
		},
	})
	require.Equal(t, "onboarding", m.Flow)
	require.Equal(t, "main", m.Branch)
	require.Equal(t, []string{"contracts", "flows"}, m.Checks)
	require.Equal(t, []string{"go test ./..."}, m.Commands)
	head, err := bootstrap.GitHead(repo)
	require.NoError(t, err)
	require.Equal(t, head, m.RepoCommit)
}

func runGitInitCommit(t *testing.T, repo string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "T"},
	} {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		require.NoError(t, cmd.Run())
	}
	require.NoError(t, os.WriteFile(filepath.Join(repo, "f"), []byte("1"), 0o644))
	cmd := exec.Command("git", "-C", repo, "add", "f")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", repo, "commit", "-m", "init")
	require.NoError(t, cmd.Run())
}
