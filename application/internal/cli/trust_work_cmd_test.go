package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestTrustTaskTextOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	payload := `{"gates":{"history":[{"gate":"enrich","status":"pass","at":"2026-01-01T00:00:00Z"}]}}`
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-trust-1", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusEnriched, PayloadJSON: payload,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "task", "task-trust-1"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Summary")
	require.Contains(t, out.String(), "task-trust-1")
}

func TestTrustTaskJSONOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-json", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "task", "task-json", "--json"})
	require.NoError(t, root.Execute())

	var report worktrust.WorkTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "task", report.Scope.Kind)
	require.Equal(t, "task-json", report.Scope.TaskID)
	require.Equal(t, worktrust.ReportVersion, report.ReportVersion)
}

func TestTrustFeatureTextOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-a", RunID: "run-1", Feature: "onboarding", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-b", RunID: "run-1", Feature: "onboarding", Status: asagiri.StatusVerified,
		PayloadJSON: `{"gates":{"history":[{"gate":"verify_evidence","status":"pass","at":"2026-01-01T00:00:00Z"}]}}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "feature", "onboarding"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Summary")
	require.Contains(t, out.String(), "Tasks: 2")
}

func TestTrustFeatureJSONOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-f1", RunID: "run-1", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "feature", "billing", "--json"})
	require.NoError(t, root.Execute())

	var report worktrust.FeatureTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "feature", report.Scope.Kind)
	require.Equal(t, "billing", report.Scope.Feature)
	require.Equal(t, 1, report.TaskCount)
}

func TestTrustTaskUnknownTask(t *testing.T) {
	trustWorkTestRepo(t)

	root := newRootCmd()
	root.SetArgs([]string{"trust", "task", "missing-task-id"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing-task-id")
}

func TestTrustFeatureNoTasks(t *testing.T) {
	trustWorkTestRepo(t)

	root := newRootCmd()
	root.SetArgs([]string{"trust", "feature", "empty-feature"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty-feature")
}

func TestTrustRunTextOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-trust-1", Feature: "onboarding", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-r1", RunID: "run-trust-1", Feature: "onboarding", Status: asagiri.StatusEnriched,
		PayloadJSON: `{"gates":{"history":[{"gate":"enrich","status":"pass","at":"2026-01-01T00:00:00Z"}]}}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "run", "run-trust-1"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Summary")
	require.Contains(t, out.String(), "run-trust-1")
}

func TestTrustTaskExplainOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-explain", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "task", "task-explain", "--explain"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Dimensions")
}

func TestTrustRunJSONOutput(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-json", Feature: "billing", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-rj", RunID: "run-json", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"trust", "run", "run-json", "--json"})
	require.NoError(t, root.Execute())

	var report worktrust.RunTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "run", report.Scope.Kind)
	require.Equal(t, "run-json", report.Scope.ID)
	require.Equal(t, worktrust.ReportVersion, report.ReportVersion)
}

func TestTrustRunUnknownRun(t *testing.T) {
	trustWorkTestRepo(t)

	root := newRootCmd()
	root.SetArgs([]string{"trust", "run", "missing-run-id"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing-run-id")
}

func TestTrustRunNoTasks(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-empty", Feature: "solo", Status: sqlite.StatusPending, StepsJSON: `[]`,
	}))
	_ = store.Close()

	root := newRootCmd()
	root.SetArgs([]string{"trust", "run", "run-empty"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "run-empty")
}

func TestTrustTaskSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-save", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "task", "task-save", "--save"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Summary")
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/tasks/task-save.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "tasks", "task-save.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
	var report worktrust.WorkTrustReport
	require.NoError(t, json.Unmarshal(raw, &report))
	require.Equal(t, "task-save", report.Scope.TaskID)
}

func TestTrustTaskJSONSave(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-json-save", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "task", "task-json-save", "--json", "--save"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report worktrust.WorkTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "task-json-save", report.Scope.TaskID)
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/tasks/task-json-save.json")
}

func TestTrustFeatureJSONSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-fjson", RunID: "run-1", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "feature", "billing", "--json", "--save"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report worktrust.FeatureTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "billing", report.Scope.Feature)
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/features/billing.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "features", "billing.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
}

func TestTrustRunJSONSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-json-save", Feature: "billing", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-rjson", RunID: "run-json-save", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "run", "run-json-save", "--json", "--save"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report worktrust.RunTrustReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "run-json-save", report.Scope.ID)
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/runs/run-json-save.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "runs", "run-json-save.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
}

func TestTrustFeatureSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-fsave", RunID: "run-1", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, _, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "feature", "billing", "--save"})
	require.NoError(t, root.Execute())
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/features/billing.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "features", "billing.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
}

func TestTrustRunSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-save", Feature: "billing", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-rsave", RunID: "run-save", Feature: "billing", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, _, errOut := rootWithSplitIO()
	root.SetArgs([]string{"trust", "run", "run-save", "--save"})
	require.NoError(t, root.Execute())
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/trust/runs/run-save.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "runs", "run-save.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
}

func TestTrustTaskSaveOverwriteStable(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-ow", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root := newRootCmd()
	root.SetArgs([]string{"trust", "task", "task-ow", "--save"})
	require.NoError(t, root.Execute())
	require.NoError(t, root.Execute())

	abs := filepath.Join(repo, ".asagiri", "reports", "trust", "tasks", "task-ow.json")
	info, err := os.Stat(abs)
	require.NoError(t, err)
	require.False(t, info.IsDir())
}

func trustWorkTestRepo(t *testing.T) (string, *sqlite.Store) {
	t.Helper()
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	dbPath := filepath.Join(dir, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return dir, store
}
