package runtime

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/embedutil"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func memoryEmbedJSON(summary string) (string, error) {
	v, err := embedder.EmbedText(context.Background(), summary)
	if err != nil {
		return "", err
	}
	return embedutil.ToJSON(v), nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

const driverName = "sqlite"

// DefaultRelDir is the runtime root under the repo.
const DefaultRelDir = ".asagiri/runtime"

// Store persists sessions, branches, events, and memory (spec-my-A §24.5).
type Store struct {
	repoRoot string
	db       *sql.DB
}

// Open opens or creates the runtime SQLite database.
func Open(repoRoot string) (*Store, error) {
	dir := filepath.Join(repoRoot, DefaultRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dir, "runtime.db")
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("open runtime db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping runtime db: %w", err)
	}
	s := &Store{repoRoot: repoRoot, db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(string(raw)); err != nil {
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("runtime migrate %s: %w", name, err)
		}
	}
	return nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// RepoRoot returns the repository root.
func (s *Store) RepoRoot() string { return s.repoRoot }

// DBPath returns the SQLite file path.
func (s *Store) DBPath() string {
	return filepath.Join(s.repoRoot, DefaultRelDir, "runtime.db")
}

func (s *Store) setMeta(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO runtime_meta(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (s *Store) getMeta(key string) (string, bool, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM runtime_meta WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

// CreateSession inserts a new session and emits session.created.
func (s *Store) CreateSession(name, productID, flowID string) (Session, error) {
	now := time.Now().UTC()
	sess := Session{
		ID:        uuid.NewString(),
		Name:      name,
		ProductID: productID,
		FlowID:    flowID,
		Status:    SessionActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	ctxJSON, _ := json.Marshal(sess.RuntimeContext)
	_, err := s.db.Exec(`INSERT INTO sessions(id, name, product_id, flow_id, branch_id, status, context_json, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		sess.ID, sess.Name, nullStr(sess.ProductID), nullStr(sess.FlowID), nil,
		string(sess.Status), string(ctxJSON), formatTime(now), formatTime(now))
	if err != nil {
		return Session{}, err
	}
	_, _ = s.EmitEvent("session.created", "runtime", sess.ID, sess.FlowID, map[string]any{"name": name})
	_ = s.DispatchHooks("session.created", false)
	return sess, nil
}

// ListSessions returns active and paused sessions.
func (s *Store) ListSessions() ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, name, product_id, flow_id, branch_id, status, context_json, created_at, updated_at
		FROM sessions WHERE status != 'archived' ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var sess Session
		var productID, flowID, branchID, ctxJSON sql.NullString
		var status, createdAt, updatedAt string
		if err := rows.Scan(&sess.ID, &sess.Name, &productID, &flowID, &branchID, &status, &ctxJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		sess.ProductID = productID.String
		sess.FlowID = flowID.String
		sess.BranchID = branchID.String
		sess.Status = SessionStatus(status)
		sess.CreatedAt = parseTime(createdAt)
		sess.UpdatedAt = parseTime(updatedAt)
		if ctxJSON.Valid && ctxJSON.String != "" {
			_ = json.Unmarshal([]byte(ctxJSON.String), &sess.RuntimeContext)
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// GetSession loads one session by id or name.
func (s *Store) GetSession(idOrName string) (Session, error) {
	row := s.db.QueryRow(`SELECT id, name, product_id, flow_id, branch_id, status, context_json, created_at, updated_at
		FROM sessions WHERE id=? OR name=? LIMIT 1`, idOrName, idOrName)
	var sess Session
	var productID, flowID, branchID, ctxJSON sql.NullString
	var status, createdAt, updatedAt string
	err := row.Scan(&sess.ID, &sess.Name, &productID, &flowID, &branchID, &status, &ctxJSON, &createdAt, &updatedAt)
	if err != nil {
		return Session{}, err
	}
	sess.ProductID = productID.String
	sess.FlowID = flowID.String
	sess.BranchID = branchID.String
	sess.Status = SessionStatus(status)
	sess.CreatedAt = parseTime(createdAt)
	sess.UpdatedAt = parseTime(updatedAt)
	if ctxJSON.Valid {
		_ = json.Unmarshal([]byte(ctxJSON.String), &sess.RuntimeContext)
	}
	return sess, nil
}

// CreateBranch adds a branch under a session.
func (s *Store) CreateBranch(sessionID, name string, branchType BranchType, parentBranchID string) (Branch, error) {
	if _, err := s.GetSession(sessionID); err != nil {
		return Branch{}, fmt.Errorf("session not found: %w", err)
	}
	now := time.Now().UTC()
	if branchType == "" {
		branchType = BranchFlow
	}
	b := Branch{
		ID:             uuid.NewString(),
		ParentBranchID: parentBranchID,
		SessionID:      sessionID,
		Name:           name,
		Type:           branchType,
		CreatedAt:      now,
	}
	_, err := s.db.Exec(`INSERT INTO branches(id, parent_branch_id, session_id, name, branch_type, description, divergence_json, created_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		b.ID, nullStr(b.ParentBranchID), b.SessionID, b.Name, string(b.Type), "", "{}", formatTime(now))
	if err != nil {
		return Branch{}, err
	}
	_, _ = s.EmitEvent("branch.created", "runtime", sessionID, "", map[string]any{"branch_id": b.ID, "name": name})
	return b, nil
}

// ListBranches returns branches for a session.
func (s *Store) ListBranches(sessionID string) ([]Branch, error) {
	sess, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT id, parent_branch_id, session_id, name, branch_type, description, created_at
		FROM branches WHERE session_id=? ORDER BY created_at`, sess.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Branch
	for rows.Next() {
		var b Branch
		var parentID, desc sql.NullString
		var branchType, createdAt string
		if err := rows.Scan(&b.ID, &parentID, &b.SessionID, &b.Name, &branchType, &desc, &createdAt); err != nil {
			return nil, err
		}
		b.ParentBranchID = parentID.String
		b.Type = BranchType(branchType)
		b.Description = desc.String
		b.CreatedAt = parseTime(createdAt)
		out = append(out, b)
	}
	return out, rows.Err()
}

// EmitEvent persists a runtime bus event.
func (s *Store) EmitEvent(eventType, source, sessionID, flowID string, payload map[string]any) (RuntimeEvent, error) {
	now := time.Now().UTC()
	ev := RuntimeEvent{
		ID:        uuid.NewString(),
		Type:      eventType,
		Source:    source,
		SessionID: sessionID,
		FlowID:    flowID,
		Payload:   payload,
		CreatedAt: now,
	}
	raw, _ := json.Marshal(payload)
	_, err := s.db.Exec(`INSERT INTO runtime_events(id, event_type, source, session_id, flow_id, payload_json, created_at)
		VALUES(?,?,?,?,?,?,?)`,
		ev.ID, ev.Type, source, nullStr(sessionID), nullStr(flowID), string(raw), formatTime(now))
	return ev, err
}

// ListEvents returns recent events newest-first.
func (s *Store) ListEvents(limit int) ([]RuntimeEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`SELECT id, event_type, source, session_id, flow_id, payload_json, created_at
		FROM runtime_events ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

// ListEventsSince returns events after a timestamp (for --follow polling).
func (s *Store) ListEventsSince(since time.Time, limit int) ([]RuntimeEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id, event_type, source, session_id, flow_id, payload_json, created_at
		FROM runtime_events WHERE created_at > ? ORDER BY created_at ASC LIMIT ?`,
		formatTime(since), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func scanEvents(rows *sql.Rows) ([]RuntimeEvent, error) {
	var out []RuntimeEvent
	for rows.Next() {
		var ev RuntimeEvent
		var sessionID, flowID, payloadJSON sql.NullString
		var createdAt string
		if err := rows.Scan(&ev.ID, &ev.Type, &ev.Source, &sessionID, &flowID, &payloadJSON, &createdAt); err != nil {
			return nil, err
		}
		ev.SessionID = sessionID.String
		ev.FlowID = flowID.String
		ev.CreatedAt = parseTime(createdAt)
		if payloadJSON.Valid && payloadJSON.String != "" {
			_ = json.Unmarshal([]byte(payloadJSON.String), &ev.Payload)
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// UpsertMemory stores or updates a memory entry.
func (s *Store) UpsertMemory(e MemoryEntry) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.LastUsedAt.IsZero() {
		e.LastUsedAt = now
	}
	tags, _ := json.Marshal(e.Tags)
	flows, _ := json.Marshal(e.LinkedFlows)
	emb := e.EmbeddingJSON
	if emb == "" && e.Summary != "" {
		var err error
		emb, err = memoryEmbedJSON(e.Summary)
		if err != nil {
			return err
		}
	}
	_, err := s.db.Exec(`INSERT INTO memory_entries(id, scope, entry_type, summary, source, relevance, tags_json, linked_flows_json, embedding_json, created_at, last_used_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET summary=excluded.summary, relevance=excluded.relevance, embedding_json=excluded.embedding_json, last_used_at=excluded.last_used_at`,
		e.ID, string(e.Scope), e.Type, e.Summary, e.Source, e.Relevance, string(tags), string(flows), nullStr(emb),
		formatTime(e.CreatedAt), formatTime(e.LastUsedAt))
	return err
}

// CountMemory returns memory entry count.
func (s *Store) CountMemory() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM memory_entries`).Scan(&n)
	return n, err
}

// CountSessions returns non-archived session count.
func (s *Store) CountSessions() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status != 'archived'`).Scan(&n)
	return n, err
}

// CountQueuedEvents returns pending hook queue items.
func (s *Store) CountQueuedEvents() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM hook_queue WHERE status='pending'`).Scan(&n)
	if err != nil {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM runtime_events WHERE created_at > datetime('now', '-1 hour')`).Scan(&n)
	}
	return n, err
}

// ListMemory returns entries optionally filtered by scope (empty = all).
func (s *Store) ListMemory(scope MemoryScope, limit int) ([]MemoryEntry, error) {
	q := `SELECT id, scope, entry_type, summary, source, relevance, tags_json, linked_flows_json, COALESCE(embedding_json,''), created_at, last_used_at
		FROM memory_entries`
	var args []any
	if scope != "" {
		q += ` WHERE scope=?`
		args = append(args, string(scope))
	}
	q += ` ORDER BY relevance DESC, last_used_at DESC`
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var scopeStr, tagsJSON, flowsJSON, embJSON, createdAt, lastUsed string
		if err := rows.Scan(&e.ID, &scopeStr, &e.Type, &e.Summary, &e.Source, &e.Relevance, &tagsJSON, &flowsJSON, &embJSON, &createdAt, &lastUsed); err != nil {
			return nil, err
		}
		e.Scope = MemoryScope(scopeStr)
		e.CreatedAt = parseTime(createdAt)
		e.LastUsedAt = parseTime(lastUsed)
		if tagsJSON != "" {
			_ = json.Unmarshal([]byte(tagsJSON), &e.Tags)
		}
		if flowsJSON != "" {
			_ = json.Unmarshal([]byte(flowsJSON), &e.LinkedFlows)
		}
		e.EmbeddingJSON = embJSON
		out = append(out, e)
	}
	return out, rows.Err()
}

// EnqueueHook schedules a hook command for the worker.
func (s *Store) EnqueueHook(eventType, command string) error {
	_, err := s.db.Exec(`INSERT INTO hook_queue(id, event_type, command, status, created_at)
		VALUES(?,?,?,?,?)`, uuid.NewString(), eventType, command, "pending", formatTime(time.Now().UTC()))
	return err
}

// DequeueHooks returns pending hooks up to limit.
func (s *Store) DequeueHooks(limit int) ([]HookJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`SELECT id, event_type, command, created_at FROM hook_queue
		WHERE status='pending' ORDER BY created_at ASC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HookJob
	for rows.Next() {
		var j HookJob
		var created string
		if err := rows.Scan(&j.ID, &j.EventType, &j.Command, &created); err != nil {
			return nil, err
		}
		j.CreatedAt = parseTime(created)
		out = append(out, j)
	}
	return out, rows.Err()
}

// MarkHookDone updates hook status after execution.
func (s *Store) MarkHookDone(id, status string) error {
	_, err := s.db.Exec(`UPDATE hook_queue SET status=?, executed_at=? WHERE id=?`,
		status, formatTime(time.Now().UTC()), id)
	return err
}

// TouchWorkerHeartbeat records worker liveness.
func (s *Store) TouchWorkerHeartbeat(workerID string) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(`INSERT INTO workers(id, status, last_heartbeat, created_at)
		VALUES(?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET status=excluded.status, last_heartbeat=excluded.last_heartbeat`,
		workerID, "active", now, now)
	return err
}

// BuildStateGraph projects sessions, branches, and recent events.
func (s *Store) BuildStateGraph() (StateGraph, error) {
	sessions, err := s.ListSessions()
	if err != nil {
		return StateGraph{}, err
	}
	var branches []Branch
	for _, sess := range sessions {
		bs, err := s.ListBranches(sess.ID)
		if err != nil {
			continue
		}
		branches = append(branches, bs...)
	}
	events, err := s.ListEvents(20)
	if err != nil {
		return StateGraph{}, err
	}
	return StateGraph{Sessions: sessions, Branches: branches, Events: events}, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	if t.IsZero() {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}
