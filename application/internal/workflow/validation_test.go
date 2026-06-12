package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	appvalidate "github.com/LaProgrammerie/asagiri/application/internal/validation"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestDefaultValidationCommands(t *testing.T) {
	dir := t.TempDir()
	cmds := validationLinesForRepo(dir)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 cmds, got %v", cmds)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds = validationLinesForRepo(dir)
	if len(cmds) < 2 {
		t.Fatalf("expected make lint, got %v", cmds)
	}
}

func TestParseValidationCommands(t *testing.T) {
	payload := `{"validation_commands":["go test ./...","go vet ./..."]}`
	got := parseValidationCommands(payload)
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestRunVerificationDryRunReturnsNilResults(t *testing.T) {
	svc := &Service{dryRun: true}
	results, err := svc.runVerification(context.Background(), t.TempDir(), `{}`)
	require.NoError(t, err)
	require.Nil(t, results)
}

func TestRunVerificationSuccessReturnsResults(t *testing.T) {
	dir := t.TempDir()
	svc := &Service{repoRoot: dir, dryRun: false}
	payload := `{"validation_commands":["echo ok"]}`
	results, err := svc.runVerification(context.Background(), dir, payload)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "task-0", results[0].Name)
	require.Equal(t, "echo ok", results[0].Command)
	require.Equal(t, 0, results[0].ExitCode)
	require.NoError(t, results[0].Err)
}

func TestRunVerificationFailureReturnsPartialResults(t *testing.T) {
	dir := t.TempDir()
	svc := &Service{repoRoot: dir, dryRun: false}
	payload := `{"validation_commands":["echo ok","__asa_missing_validation_cmd__"]}`
	results, err := svc.runVerification(context.Background(), dir, payload)
	require.Error(t, err)
	require.Len(t, results, 2)
	require.Equal(t, 0, results[0].ExitCode)
	require.NoError(t, results[0].Err)
	require.Error(t, results[1].Err)
}

func TestPersistValidationEvidenceSuccess(t *testing.T) {
	repo := t.TempDir()
	svc := &Service{repoRoot: repo, dryRun: false}
	results := []appvalidate.Result{{
		Name: "task-0", Command: "echo ok", ExitCode: 0, Output: "ok",
	}}
	require.NoError(t, svc.persistValidationEvidence("task-1", "", results))

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(repo, "task-1"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "task-1", doc.TaskID)
	require.False(t, doc.DryRun)
	require.NotEmpty(t, doc.At)
	require.Len(t, doc.Commands, 1)
	require.Equal(t, "task-0", doc.Commands[0].Name)
	require.Equal(t, 0, doc.Commands[0].ExitCode)
}

func TestPersistValidationEvidencePartialFailure(t *testing.T) {
	repo := t.TempDir()
	svc := &Service{repoRoot: repo, dryRun: false}
	results := []appvalidate.Result{
		{Name: "task-0", Command: "echo ok", ExitCode: 0, Output: "ok"},
		{Name: "task-1", Command: "__missing__", ExitCode: 1, Err: fmt.Errorf("failed")},
	}
	require.NoError(t, svc.persistValidationEvidence("task-partial", "/wt", results))

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(repo, "task-partial"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "/wt", doc.Worktree)
	require.Len(t, doc.Commands, 2)
	require.Equal(t, 1, doc.Commands[1].ExitCode)
}

func TestPersistValidationEvidenceDryRunEmptyBundle(t *testing.T) {
	repo := t.TempDir()
	svc := &Service{repoRoot: repo, dryRun: true}
	require.NoError(t, svc.persistValidationEvidence("task-dry", "", nil))

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(repo, "task-dry"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.True(t, doc.DryRun)
	require.Empty(t, doc.Commands)
}

func seedTaskWithValidationCommands(t *testing.T, store *sqlite.Store, feature, id, status string, cmds []string) sqlite.Task {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"validation_commands": cmds})
	require.NoError(t, err)
	task := seedTask(t, store, feature, id, status)
	require.NoError(t, store.UpdateTask(&sqlite.Task{ID: id, PayloadJSON: string(payload)}))
	task.PayloadJSON = string(payload)
	return task
}

func TestVerifyFeaturePersistsValidationOnSuccess(t *testing.T) {
	svc, store := humanReviewTestService(t, false, false)
	task := seedTaskWithValidationCommands(t, store, "feat", "task-verify-ok", asagiri.StatusImplemented, []string{"echo ok"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(svc.repoRoot, task.ID))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Len(t, doc.Commands, 1)
	require.Equal(t, 0, doc.Commands[0].ExitCode)
}

func TestVerifyFeaturePersistsValidationOnFailure(t *testing.T) {
	svc, store := humanReviewTestService(t, false, false)
	task := seedTaskWithValidationCommands(t, store, "feat", "task-verify-ko", asagiri.StatusImplemented,
		[]string{"echo ok", "__asa_missing_validation_cmd__"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.Error(t, err)

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(svc.repoRoot, task.ID))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Len(t, doc.Commands, 2)
	require.Equal(t, 1, doc.Commands[1].ExitCode)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerifyFailed, fresh.Status)
}

func TestVerifyFeaturePersistsValidationDryRun(t *testing.T) {
	svc, store := humanReviewTestService(t, false, false)
	svc.dryRun = true
	task := seedTask(t, store, "feat", "task-verify-dry", asagiri.StatusImplemented)

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	var doc validationEvidenceDocument
	body, err := os.ReadFile(validationResultsPath(svc.repoRoot, task.ID))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &doc))
	require.True(t, doc.DryRun)
	require.Empty(t, doc.Commands)
}
