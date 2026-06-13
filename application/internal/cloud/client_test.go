package cloud_test

import (
	"context"
	"encoding/json"
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

func TestClientMeAndProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			_, _ = w.Write([]byte(`{"id":"u1","email":"dev@test","displayName":"Dev"}`))
		case "/api/v1/projects":
			_, _ = w.Write([]byte(`[{"id":"p1","name":"Demo","slug":"demo"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(cloud.ClientOptions{BaseURL: srv.URL, Token: "tok"})
	me, err := client.Me(context.Background())
	require.NoError(t, err)
	require.Equal(t, "dev@test", me.Email)

	projects, err := client.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "p1", projects[0].ID)
}

func TestPushDryRunGroupsLedger(t *testing.T) {
	repo := t.TempDir()
	ledgerPath := filepath.Join(repo, ".asagiri", "logs", "agents", "ledger.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ledgerPath), 0o755))
	entry := agentledger.Entry{
		RunID: "run-a", Feature: "feat", AgentID: "dev", ExitCode: 0,
		StartedAt: "2026-01-01T10:00:00Z", EndedAt: "2026-01-01T10:01:00Z", DurationMS: 60000,
	}
	raw, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ledgerPath, append(raw, '\n'), 0o644))

	cfg := config.NewTestConfig("test")
	cfg.Cloud.Enabled = true
	cfg.Cloud.BaseURL = "http://example"
	cfg.Cloud.ProjectID = "p1"

	report, err := cloud.Push(context.Background(), cloud.PushOptions{
		RepoRoot: repo,
		Config:   cfg,
		Token:    "tok",
		All:      true,
		DryRun:   true,
	})
	require.NoError(t, err)
	require.Equal(t, "dry-run", report.Mode)
	require.Len(t, report.Items, 1)
	require.Equal(t, "run-a", report.Items[0].LocalRunID)
	require.True(t, report.Items[0].DryRun)
}
