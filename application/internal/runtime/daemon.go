package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	metaDaemonStatus = "daemon.status"
	metaDaemonPID    = "daemon.pid"
	metaDaemonStart  = "daemon.started_at"
	pidFileName      = "daemon.pid"
	stateFileName    = "daemon.state.json"
)

// DaemonState is written beside the runtime DB for CLI status.
type DaemonState struct {
	Running   bool      `json:"running"`
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	DBPath    string    `json:"db_path"`
}

// StartDaemon initializes runtime storage and marks the daemon running (local-first V1).
func StartDaemon(repoRoot string) (DaemonStatus, error) {
	store, err := Open(repoRoot)
	if err != nil {
		return DaemonStatus{}, err
	}
	defer func() { _ = store.Close() }()

	pid := os.Getpid()
	now := time.Now().UTC()
	if err := store.setMeta(metaDaemonStatus, "running"); err != nil {
		return DaemonStatus{}, err
	}
	if err := store.setMeta(metaDaemonPID, strconv.Itoa(pid)); err != nil {
		return DaemonStatus{}, err
	}
	if err := store.setMeta(metaDaemonStart, formatTime(now)); err != nil {
		return DaemonStatus{}, err
	}
	_, _ = store.EmitEvent("runtime.started", "daemon", "", "", map[string]any{"pid": pid})

	dir := filepath.Join(repoRoot, DefaultRelDir)
	state := DaemonState{Running: true, PID: pid, StartedAt: now, DBPath: store.DBPath()}
	raw, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, stateFileName), raw, 0o644)
	_ = os.WriteFile(filepath.Join(dir, pidFileName), []byte(strconv.Itoa(pid)), 0o644)

	return store.Status()
}

// StopDaemon marks the runtime as stopped without deleting data.
func StopDaemon(repoRoot string) error {
	store, err := Open(repoRoot)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()
	_ = store.setMeta(metaDaemonStatus, "stopped")
	_ = store.setMeta(metaDaemonPID, "")
	_, _ = store.EmitEvent("runtime.stopped", "daemon", "", "", nil)

	dir := filepath.Join(repoRoot, DefaultRelDir)
	_ = os.Remove(filepath.Join(dir, pidFileName))
	state := DaemonState{Running: false, DBPath: store.DBPath()}
	raw, _ := json.MarshalIndent(state, "", "  ")
	return os.WriteFile(filepath.Join(dir, stateFileName), raw, 0o644)
}

// Status returns daemon and runtime counters for `asa daemon status`.
func (s *Store) Status() (DaemonStatus, error) {
	st := DaemonStatus{DBPath: s.DBPath()}
	if info, err := os.Stat(s.DBPath()); err == nil {
		st.DBSizeBytes = info.Size()
	}
	status, ok, err := s.getMeta(metaDaemonStatus)
	if err != nil {
		return st, err
	}
	st.Running = ok && status == "running"
	if pidStr, ok, _ := s.getMeta(metaDaemonPID); ok && pidStr != "" {
		st.PID, _ = strconv.Atoi(pidStr)
	}
	st.Sessions, _ = s.CountSessions()
	st.MemorySize, _ = s.CountMemory()
	st.QueuedEvents, _ = s.CountQueuedEvents()
	rows, err := s.db.Query(`SELECT COUNT(DISTINCT flow_id) FROM runtime_events WHERE flow_id IS NOT NULL AND flow_id != ''`)
	if err == nil {
		defer func() { _ = rows.Close() }()
		if rows.Next() {
			_ = rows.Scan(&st.FlowsActive)
		}
	}
	return st, nil
}

// FormatStatusPlain renders terminal output (spec-my-A §24.3).
func FormatStatusPlain(st DaemonStatus) string {
	status := "stopped"
	if st.Running {
		status = "running"
	}
	return fmt.Sprintf(`Asagiri Runtime
────────────────
Status: %s
Sessions: %d
Flows active: %d
Queued events: %d
Memory size: %d entries
DB size: %d bytes
DB path: %s
`, status, st.Sessions, st.FlowsActive, st.QueuedEvents, st.MemorySize, st.DBSizeBytes, st.DBPath)
}
