package replay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWriteReplay(t *testing.T) {
	repo := t.TempDir()
	trustID := "trust-2026-05-29-replay"

	err := WriteReplay(repo, trustID, Manifest{
		Flow:   "onboarding",
		Branch: "main",
	})
	require.NoError(t, err)

	path := filepath.Join(repo, ".asagiri", "trust", trustID, "replay.yaml")
	require.FileExists(t, path)

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	var decoded Manifest
	require.NoError(t, yaml.Unmarshal(body, &decoded))
	require.Equal(t, trustID, decoded.TrustID)
	require.Equal(t, "onboarding", decoded.Flow)
	require.Equal(t, "main", decoded.Branch)
	require.Contains(t, string(body), "flow: onboarding")
}

func TestWriteReplayFullManifest(t *testing.T) {
	repo := t.TempDir()
	trustID := "trust-2026-05-29-full"
	err := WriteReplay(repo, trustID, Manifest{
		TrustID:    trustID,
		Flow:       "onboarding",
		Branch:     "feat/x",
		RepoCommit: "abc123",
		Checks:     []string{"contracts", "flows"},
		Commands:   []string{"go test ./..."},
	})
	require.NoError(t, err)
	loaded, err := Load(repo, trustID)
	require.NoError(t, err)
	require.Equal(t, "abc123", loaded.RepoCommit)
	require.Equal(t, []string{"contracts", "flows"}, loaded.Checks)
	require.Equal(t, []string{"go test ./..."}, loaded.Commands)
}

func TestWriteReplayFillsEmptyTrustID(t *testing.T) {
	repo := t.TempDir()
	trustID := "trust-2026-05-29-fill"

	err := WriteReplay(repo, trustID, Manifest{Flow: "f"})
	require.NoError(t, err)

	path := filepath.Join(repo, ".asagiri", "trust", trustID, "replay.yaml")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(body), "trust_id: "+trustID)
}
