package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestNextWithTrustSummary(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-next", Feature: "myfeat", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-next-trust", RunID: "run-next", Feature: "myfeat", Status: asagiri.StatusEnriched,
		PayloadJSON: `{"gates":{"history":[{"gate":"enrich","status":"pass","at":"2026-01-01T00:00:00Z"}]}}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"next", "--feature", "myfeat"})
	require.NoError(t, root.Execute())
	body := out.String()
	require.Contains(t, body, "Next action:")
	require.Contains(t, body, "Trust")
	require.Contains(t, body, "Verdict:")
}

func TestNextNoTrust(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-nt", Feature: "myfeat", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-no-trust", RunID: "run-nt", Feature: "myfeat", Status: asagiri.StatusEnriched,
		PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"next", "--feature", "myfeat", "--no-trust"})
	require.NoError(t, root.Execute())
	require.NotContains(t, out.String(), "Trust\n")
}

func TestStatusWithTrustSection(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-st", Feature: "onboarding", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-st", RunID: "run-st", Feature: "onboarding", Status: asagiri.StatusPlanned,
		PayloadJSON: payloadEnrichFail(),
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "run-st")
	require.Contains(t, out.String(), "Trust")
	require.Contains(t, out.String(), "Feature: onboarding")
}

func TestStatusNoTrust(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-nt", Feature: "solo", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-nt", RunID: "run-nt", Feature: "solo", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"status", "--no-trust"})
	require.NoError(t, root.Execute())
	require.NotContains(t, out.String(), "Trust\n")
}

func TestTrustTaskJSONUnchangedWithDailyUX(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-json-daily", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "task", "task-json-daily", "--json"})
	require.NoError(t, root.Execute())

	var report worktrust.WorkTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, worktrust.ReportVersion, report.ReportVersion)
	require.Equal(t, "task", report.Scope.Kind)
}

func payloadEnrichFail() string {
	return `{"gates":{"history":[{"gate":"enrich","status":"fail","confidence":0.1}]}}`
}
