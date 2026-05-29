package executiongraph

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
)

// CheckpointState is persisted resumable state after a node completes (spec §15).
type CheckpointState struct {
	AfterNode    string   `json:"after_node"`
	GitRef       string   `json:"git_ref,omitempty"`
	GitDirty     bool     `json:"git_dirty,omitempty"`
	WorktreePath string   `json:"worktree_path,omitempty"`
	Outputs      []string `json:"outputs,omitempty"`
	Validations  []string `json:"validations,omitempty"`
	CostConsumed float64  `json:"cost_consumed"`
	Duration     string   `json:"duration,omitempty"`
	CreatedAt    string   `json:"created_at"`
}

func (r *Repository) checkpointDir(graphID string) string {
	return filepath.Join(r.graphDir(graphID), "checkpoints")
}

// SaveCheckpoint persists resumable state for a graph node.
func (r *Repository) SaveCheckpoint(graphID string, state CheckpointState) (string, error) {
	if err := ValidateGraphID(graphID); err != nil {
		return "", fmt.Errorf("save checkpoint: %w", err)
	}
	if state.AfterNode == "" {
		return "", fmt.Errorf("save checkpoint: after_node required")
	}
	dir := r.checkpointDir(graphID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create checkpoint dir: %w", err)
	}
	if state.CreatedAt == "" {
		state.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	path := filepath.Join(dir, state.AfterNode+".json")
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal checkpoint: %w", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("write checkpoint: %w", err)
	}
	return path, nil
}

// LoadLatestCheckpoint returns the most recently written checkpoint for a graph.
func (r *Repository) LoadLatestCheckpoint(graphID string) (CheckpointState, bool, error) {
	dir := r.checkpointDir(graphID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return CheckpointState{}, false, nil
		}
		return CheckpointState{}, false, fmt.Errorf("read checkpoint dir: %w", err)
	}
	var latest CheckpointState
	var latestTime time.Time
	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return CheckpointState{}, false, fmt.Errorf("read checkpoint %s: %w", e.Name(), err)
		}
		var state CheckpointState
		if err := json.Unmarshal(raw, &state); err != nil {
			return CheckpointState{}, false, fmt.Errorf("parse checkpoint %s: %w", e.Name(), err)
		}
		ts, err := time.Parse(time.RFC3339, state.CreatedAt)
		if err != nil {
			ts = time.Time{}
		}
		if !found || ts.After(latestTime) {
			latest = state
			latestTime = ts
			found = true
		}
	}
	return latest, found, nil
}

// CountCheckpoints returns the number of persisted checkpoint files for a graph.
func (r *Repository) CountCheckpoints(graphID string) (int, error) {
	dir := r.checkpointDir(graphID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read checkpoint dir: %w", err)
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		count++
	}
	return count, nil
}

// AppendGraphEvent appends one JSON line to events.jsonl under the graph directory.
func (r *Repository) AppendGraphEvent(graphID, eventType string, payload map[string]any) error {
	dir := r.graphDir(graphID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure graph dir: %w", err)
	}
	line := map[string]any{
		"type":       eventType,
		"graph_id":   graphID,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range payload {
		line[k] = v
	}
	body, err := json.Marshal(line)
	if err != nil {
		return fmt.Errorf("marshal graph event: %w", err)
	}
	path := filepath.Join(dir, "events.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open events.jsonl: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(body, '\n')); err != nil {
		return fmt.Errorf("write events.jsonl: %w", err)
	}
	return nil
}

// CaptureGitState returns HEAD ref and dirty flag; stubs when git is unavailable.
func CaptureGitState(repoRoot string) (ref string, dirty bool) {
	ref, err := bootstrap.GitHead(repoRoot)
	if err != nil {
		return "unknown", false
	}
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return ref, false
	}
	return ref, len(strings.TrimSpace(string(out))) > 0
}
