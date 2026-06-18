package cloud_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCloudPayloadRunTaskIDAndLedgerLocalEntryID(t *testing.T) {
	entry := agentledger.Entry{
		TaskID:     "chatbot-001",
		RunID:      "run-1",
		Feature:    "chatbot",
		AgentID:    "dev",
		Phase:      "dev",
		StartedAt:  "2026-01-01T10:00:00Z",
		EndedAt:    "2026-01-01T10:01:00Z",
		DurationMS: 60000,
		ExitCode:   0,
		PromptHash: "prompt-hash",
		OutputHash: "output-hash",
	}
	localID := agentledger.LocalEntryID(entry)

	var runBody, ledgerBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		switch r.URL.Path {
		case "/api/v1/runs":
			runBody = body
			_, _ = w.Write([]byte(`{"id":"cloud-run-1","localRunId":"run-1"}`))
		case "/api/v1/ledger-entries":
			ledgerBody = body
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repo, cfg := cloudPushFixture(t, srv.URL, entry)
	report, err := cloud.Push(context.Background(), cloud.PushOptions{
		RepoRoot: repo,
		Config:   cfg,
		Token:    "tok",
		All:      true,
		DryRun:   false,
	})
	require.NoError(t, err)
	require.Len(t, report.Items, 1)

	var runReq map[string]any
	require.NoError(t, json.Unmarshal(runBody, &runReq))
	require.Equal(t, "chatbot-001", runReq["taskId"])
	require.Equal(t, "run-1", runReq["localRunId"])

	var ledgerReq map[string]any
	require.NoError(t, json.Unmarshal(ledgerBody, &ledgerReq))
	require.Equal(t, localID, ledgerReq["localEntryId"])
	require.Equal(t, "dev", ledgerReq["agentId"])
}

func TestCloudPayloadIdenticalOnRepeatedPush(t *testing.T) {
	entry := agentledger.Entry{
		TaskID: "task-1", RunID: "run-a", Feature: "feat", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", EndedAt: "2026-01-01T10:01:00Z", DurationMS: 60000,
		ExitCode: 0, PromptHash: "p", OutputHash: "o",
	}

	var mu sync.Mutex
	var runBodies [][]byte
	var ledgerBodies [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/api/v1/runs":
			runBodies = append(runBodies, append([]byte(nil), body...))
			_, _ = w.Write([]byte(`{"id":"cloud-run-1"}`))
		case "/api/v1/ledger-entries":
			ledgerBodies = append(ledgerBodies, append([]byte(nil), body...))
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repo, cfg := cloudPushFixture(t, srv.URL, entry)

	for i := 0; i < 3; i++ {
		_, err := cloud.Push(context.Background(), cloud.PushOptions{
			RepoRoot: repo,
			Config:   cfg,
			Token:    "tok",
			All:      true,
			DryRun:   false,
		})
		require.NoError(t, err)
	}

	require.Len(t, runBodies, 3)
	require.Len(t, ledgerBodies, 3)
	require.Equal(t, string(runBodies[0]), string(runBodies[1]))
	require.Equal(t, string(runBodies[1]), string(runBodies[2]))
	require.Equal(t, string(ledgerBodies[0]), string(ledgerBodies[1]))
	require.Equal(t, string(ledgerBodies[1]), string(ledgerBodies[2]))
}

func cloudPushFixture(t *testing.T, baseURL string, entry agentledger.Entry) (repo string, cfg *config.Config) {
	t.Helper()
	repo = t.TempDir()
	ledgerPath := filepath.Join(repo, ".asagiri", "logs", "agents", "ledger.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ledgerPath), 0o755))
	raw, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ledgerPath, append(raw, '\n'), 0o644))

	cfg = config.NewTestConfig("test")
	cfg.Cloud.Enabled = true
	cfg.Cloud.BaseURL = baseURL
	cfg.Cloud.ProjectID = "p1"
	return repo, cfg
}
