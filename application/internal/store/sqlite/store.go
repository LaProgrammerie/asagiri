package sqlite

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const driverName = "sqlite"

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusVerified = "verified"
	StatusReviewed = "reviewed"
	StatusFailed   = "failed"
	StatusDone     = "done"
)

// Store wraps a SQLite connection for AgentFlow state.
type Store struct {
	db *sql.DB
}

// Run models one workflow run persisted in SQLite.
type Run struct {
	ID        string
	Feature   string
	Status    string
	StepsJSON string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Task models one unit of work linked to a run.
type Task struct {
	ID           string
	RunID        string
	Feature      string
	Status       string
	PayloadJSON  string
	WorktreePath string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Open opens (or creates) the database at path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open(driverName, path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// SchemaVersion returns the current migration version (0 if unset).
func (s *Store) SchemaVersion() (int, error) {
	var exists int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'`,
	).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("check schema_version table: %w", err)
	}
	if exists == 0 {
		return 0, nil
	}

	var version int
	err = s.db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}
	return version, nil
}

// Migrate applies pending embedded migrations.
func (s *Store) Migrate() error {
	current, err := s.SchemaVersion()
	if err != nil {
		return err
	}

	migrations := []struct {
		version int
		file    string
	}{
		{1, "migrations/001_initial.sql"},
		{2, "migrations/002_runs_tasks_v1.sql"},
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		body, err := migrationsFS.ReadFile(m.file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.file, err)
		}
		if err := s.applyMigration(m.version, string(body)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) applyMigration(version int, sqlBody string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(sqlBody); err != nil {
		return fmt.Errorf("exec migration v%d: %w", version, err)
	}

	if _, err := tx.Exec(`DELETE FROM schema_version`); err != nil {
		return fmt.Errorf("clear schema_version: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, version); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}

	return tx.Commit()
}

// Ping verifies the database is reachable.
func (s *Store) Ping() error {
	return s.db.Ping()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func timeToText(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseDBTime(raw string) (time.Time, error) {
	ts, err := time.Parse(time.RFC3339Nano, raw)
	if err == nil {
		return ts, nil
	}
	ts, err = time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func normalizeRunStatus(status string) string {
	switch status {
	case StatusPending, StatusRunning, StatusVerified, StatusReviewed, StatusFailed, StatusDone, "":
		if status == "" {
			return StatusPending
		}
		return status
	default:
		return StatusPending
	}
}

func normalizeStatus(status string) string {
	if status == "" {
		return agentflow.StatusPending
	}
	switch status {
	case StatusDone:
		return agentflow.StatusImplemented
	case StatusFailed:
		return agentflow.StatusFailed
	default:
		return status
	}
}

// CreateRun inserts one run row.
func (s *Store) CreateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("create run: run nil")
	}
	if run.ID == "" {
		return fmt.Errorf("create run: id requis")
	}
	if run.Feature == "" {
		return fmt.Errorf("create run: feature requis")
	}
	if run.StepsJSON == "" {
		run.StepsJSON = "[]"
	}
	if !json.Valid([]byte(run.StepsJSON)) {
		return fmt.Errorf("create run: steps_json invalide")
	}

	now := nowUTC()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now
	run.Status = normalizeRunStatus(run.Status)

	_, err := s.db.Exec(
		`INSERT INTO runs (id, feature, status, steps_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		run.ID, run.Feature, run.Status, run.StepsJSON, timeToText(run.CreatedAt), timeToText(run.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create run: %w", err)
	}
	return nil
}

// UpdateRun updates status and steps for one run.
func (s *Store) UpdateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("update run: run nil")
	}
	if run.ID == "" {
		return fmt.Errorf("update run: id requis")
	}
	if run.StepsJSON != "" && !json.Valid([]byte(run.StepsJSON)) {
		return fmt.Errorf("update run: steps_json invalide")
	}

	current, err := s.GetRun(run.ID)
	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}

	steps := current.StepsJSON
	if run.StepsJSON != "" {
		steps = run.StepsJSON
	}
	status := current.Status
	if run.Status != "" {
		status = normalizeRunStatus(run.Status)
	}
	updated := nowUTC()

	_, err = s.db.Exec(
		`UPDATE runs
		 SET status = ?, steps_json = ?, updated_at = ?
		 WHERE id = ?`,
		status, steps, timeToText(updated), run.ID,
	)
	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}
	return nil
}

// GetRun loads one run by ID.
func (s *Store) GetRun(id string) (*Run, error) {
	row := s.db.QueryRow(
		`SELECT id, feature, status, steps_json, created_at, updated_at
		 FROM runs WHERE id = ?`,
		id,
	)

	var r Run
	var createdRaw, updatedRaw string
	if err := row.Scan(&r.ID, &r.Feature, &r.Status, &r.StepsJSON, &createdRaw, &updatedRaw); err != nil {
		return nil, err
	}
	createdAt, err := parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse run.created_at: %w", err)
	}
	updatedAt, err := parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse run.updated_at: %w", err)
	}
	r.CreatedAt = createdAt
	r.UpdatedAt = updatedAt
	return &r, nil
}

// ListRuns returns recent runs, newest first.
func (s *Store) ListRuns(limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, feature, status, steps_json, created_at, updated_at
		 FROM runs
		 ORDER BY created_at DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	defer rows.Close()

	out := make([]Run, 0)
	for rows.Next() {
		var r Run
		var createdRaw, updatedRaw string
		if err := rows.Scan(&r.ID, &r.Feature, &r.Status, &r.StepsJSON, &createdRaw, &updatedRaw); err != nil {
			return nil, fmt.Errorf("list runs scan: %w", err)
		}
		r.CreatedAt, err = parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse run.created_at: %w", err)
		}
		r.UpdatedAt, err = parseDBTime(updatedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse run.updated_at: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list runs rows: %w", err)
	}
	return out, nil
}

// FindLatestRunByFeature returns the latest run for one feature.
func (s *Store) FindLatestRunByFeature(feature string) (*Run, error) {
	row := s.db.QueryRow(
		`SELECT id, feature, status, steps_json, created_at, updated_at
		 FROM runs
		 WHERE feature = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		feature,
	)

	var r Run
	var createdRaw, updatedRaw string
	if err := row.Scan(&r.ID, &r.Feature, &r.Status, &r.StepsJSON, &createdRaw, &updatedRaw); err != nil {
		return nil, err
	}
	var err error
	r.CreatedAt, err = parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse run.created_at: %w", err)
	}
	r.UpdatedAt, err = parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse run.updated_at: %w", err)
	}
	return &r, nil
}

// CreateTask inserts one task row.
func (s *Store) CreateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("create task: task nil")
	}
	if task.ID == "" {
		return fmt.Errorf("create task: id requis")
	}
	if task.RunID == "" {
		return fmt.Errorf("create task: run_id requis")
	}
	if task.Feature == "" {
		return fmt.Errorf("create task: feature requis")
	}
	if task.PayloadJSON == "" {
		task.PayloadJSON = "{}"
	}
	if !json.Valid([]byte(task.PayloadJSON)) {
		return fmt.Errorf("create task: payload_json invalide")
	}

	now := nowUTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	task.Status = normalizeStatus(task.Status)

	_, err := s.db.Exec(
		`INSERT INTO tasks (id, run_id, feature, status, payload_json, worktree_path, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.RunID, task.Feature, task.Status, task.PayloadJSON, task.WorktreePath, timeToText(task.CreatedAt), timeToText(task.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

// UpdateTask updates mutable task fields.
func (s *Store) UpdateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("update task: task nil")
	}
	if task.ID == "" {
		return fmt.Errorf("update task: id requis")
	}

	current, err := s.GetTask(task.ID)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	status := current.Status
	if task.Status != "" {
		status = normalizeStatus(task.Status)
	}
	payload := current.PayloadJSON
	if task.PayloadJSON != "" {
		if !json.Valid([]byte(task.PayloadJSON)) {
			return fmt.Errorf("update task: payload_json invalide")
		}
		payload = task.PayloadJSON
	}
	worktreePath := current.WorktreePath
	if task.WorktreePath != "" {
		worktreePath = task.WorktreePath
	}
	updated := nowUTC()

	_, err = s.db.Exec(
		`UPDATE tasks
		 SET status = ?, payload_json = ?, worktree_path = ?, updated_at = ?
		 WHERE id = ?`,
		status, payload, worktreePath, timeToText(updated), task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

// GetTask loads one task by ID.
func (s *Store) GetTask(id string) (*Task, error) {
	row := s.db.QueryRow(
		`SELECT id, run_id, feature, status, payload_json, worktree_path, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		id,
	)

	var t Task
	var createdRaw, updatedRaw string
	if err := row.Scan(&t.ID, &t.RunID, &t.Feature, &t.Status, &t.PayloadJSON, &t.WorktreePath, &createdRaw, &updatedRaw); err != nil {
		return nil, err
	}
	createdAt, err := parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse task.created_at: %w", err)
	}
	updatedAt, err := parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse task.updated_at: %w", err)
	}
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	return &t, nil
}

// ListTasksByRun returns tasks for one run in creation order.
func (s *Store) ListTasksByRun(runID string) ([]Task, error) {
	rows, err := s.db.Query(
		`SELECT id, run_id, feature, status, payload_json, worktree_path, created_at, updated_at
		 FROM tasks
		 WHERE run_id = ?
		 ORDER BY created_at ASC`,
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks by run: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// ListTasksByFeature returns tasks for one feature.
func (s *Store) ListTasksByFeature(feature string) ([]Task, error) {
	rows, err := s.db.Query(
		`SELECT id, run_id, feature, status, payload_json, worktree_path, created_at, updated_at
		 FROM tasks
		 WHERE feature = ?
		 ORDER BY created_at ASC`,
		feature,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks by feature: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	out := make([]Task, 0)
	for rows.Next() {
		var t Task
		var createdRaw, updatedRaw string
		if err := rows.Scan(&t.ID, &t.RunID, &t.Feature, &t.Status, &t.PayloadJSON, &t.WorktreePath, &createdRaw, &updatedRaw); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		var err error
		t.CreatedAt, err = parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse task.created_at: %w", err)
		}
		t.UpdatedAt, err = parseDBTime(updatedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse task.updated_at: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tasks rows: %w", err)
	}
	return out, nil
}
