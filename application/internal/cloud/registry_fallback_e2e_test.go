package cloud_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRegistryFallbackTraceabilityPushPayload(t *testing.T) {
	entry := agentledger.Entry{
		TaskID:               "task-1",
		RunID:                "run-registry",
		Feature:              "feat",
		AgentID:              "cursor-dev",
		Phase:                "dev",
		StartedAt:            "2026-01-01T10:00:00Z",
		EndedAt:              "2026-01-01T10:01:00Z",
		DurationMS:           60000,
		ExitCode:             0,
		PromptHash:           "p",
		OutputHash:           "o",
		Provider:             "openai",
		Model:                "gpt-4o",
		ResolutionSource:     "registry",
		RegistryAgentVersion: 2,
		ConfigReleaseVersion: 3,
		ConfigContentHash:    "hash",
	}

	var ledgerBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		switch r.URL.Path {
		case "/api/v1/runs":
			_, _ = w.Write([]byte(`{"id":"cloud-run-1"}`))
		case "/api/v1/ledger-entries":
			ledgerBody = body
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repo := t.TempDir()
	ledgerPath := filepath.Join(repo, ".asagiri", "logs", "agents", "ledger.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ledgerPath), 0o755))
	raw, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ledgerPath, append(raw, '\n'), 0o644))

	cfg := config.NewTestConfig("test")
	cfg.Cloud.Enabled = true
	cfg.Cloud.BaseURL = srv.URL
	cfg.Cloud.ProjectID = "p1"

	_, err = cloud.Push(context.Background(), cloud.PushOptions{
		RepoRoot: repo,
		Config:   cfg,
		Token:    "tok",
		All:      true,
		DryRun:   false,
	})
	require.NoError(t, err)

	var ledgerReq map[string]any
	require.NoError(t, json.Unmarshal(ledgerBody, &ledgerReq))
	require.Equal(t, "registry", ledgerReq["resolutionSource"])
	require.Equal(t, float64(2), ledgerReq["registryAgentVersion"])
	require.Equal(t, float64(3), ledgerReq["configReleaseVersion"])
}
