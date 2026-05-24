package sqlite

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
	"github.com/stretchr/testify/require"
)

func TestMigrateFromEmpty(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state.sqlite")

	store, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.Migrate())

	v, err := store.SchemaVersion()
	require.NoError(t, err)
	require.Equal(t, 2, v)

	require.NoError(t, store.Migrate())
	v2, err := store.SchemaVersion()
	require.NoError(t, err)
	require.Equal(t, 2, v2)
}

func TestTablesAndColumnsExist(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "state.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	for _, table := range []string{"schema_version", "runs", "tasks"} {
		var name string
		err := store.db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		require.NoError(t, err)
	}

	for _, column := range []struct {
		table string
		name  string
	}{
		{"runs", "steps_json"},
		{"tasks", "worktree_path"},
	} {
		var found int
		query := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?`, column.table)
		err := store.db.QueryRow(
			query, column.name,
		).Scan(&found)
		require.NoError(t, err)
		require.Equal(t, 1, found)
	}
}

func TestRunAndTaskCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "state.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	run := &Run{
		ID:        "run-1",
		Feature:   "feature-x",
		Status:    StatusPending,
		StepsJSON: `[{"name":"plan","status":"pending"}]`,
	}
	require.NoError(t, store.CreateRun(run))

	run.Status = StatusRunning
	require.NoError(t, store.UpdateRun(run))

	gotRun, err := store.GetRun("run-1")
	require.NoError(t, err)
	require.Equal(t, StatusRunning, gotRun.Status)

	task := &Task{
		ID:          "task-1",
		RunID:       run.ID,
		Feature:     run.Feature,
		Status:      StatusPending,
		PayloadJSON: `{"title":"Task 1"}`,
	}
	require.NoError(t, store.CreateTask(task))
	require.NoError(t, store.UpdateTask(&Task{ID: task.ID, Status: StatusDone, WorktreePath: ".agentflow/worktrees/x"}))

	gotTask, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, agentflow.StatusImplemented, gotTask.Status)
	require.Equal(t, ".agentflow/worktrees/x", gotTask.WorktreePath)

	runs, err := store.ListRuns(10)
	require.NoError(t, err)
	require.Len(t, runs, 1)

	tasks, err := store.ListTasksByRun(run.ID)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
}
