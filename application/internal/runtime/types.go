package runtime

import "time"

// SessionStatus is the lifecycle of an engineering session.
type SessionStatus string

const (
	SessionActive   SessionStatus = "active"
	SessionPaused   SessionStatus = "paused"
	SessionArchived SessionStatus = "archived"
)

// BranchType categorizes exploratory branches.
type BranchType string

const (
	BranchProduct         BranchType = "product"
	BranchFlow            BranchType = "flow"
	BranchArchitecture    BranchType = "architecture"
	BranchImplementation  BranchType = "implementation"
	BranchReview          BranchType = "review"
	BranchPrototype       BranchType = "prototype"
)

// Session is an active engineering context (spec-my-A §24.6).
type Session struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	ProductID      string            `json:"product_id,omitempty"`
	FlowID         string            `json:"flow_id,omitempty"`
	BranchID       string            `json:"branch_id,omitempty"`
	Status         SessionStatus     `json:"status"`
	ActiveTasks    []string          `json:"active_tasks,omitempty"`
	RuntimeContext map[string]string `json:"runtime_context,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Branch is an exploratory variation (spec-my-A §24.7).
type Branch struct {
	ID             string     `json:"id"`
	ParentBranchID string     `json:"parent_branch_id,omitempty"`
	SessionID      string     `json:"session_id"`
	Name           string     `json:"name"`
	Type           BranchType `json:"type"`
	Description    string     `json:"description,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RuntimeEvent is a persisted bus event (spec-my-A §24.8).
type RuntimeEvent struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Source    string         `json:"source,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	FlowID    string         `json:"flow_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// MemoryScope scopes persistent memory (spec-my-A §24.11).
type MemoryScope string

const (
	ScopeGlobal    MemoryScope = "global"
	ScopeProject   MemoryScope = "project"
	ScopeProduct   MemoryScope = "product"
	ScopeFlow      MemoryScope = "flow"
	ScopeFeature   MemoryScope = "feature"
	ScopeTask      MemoryScope = "task"
	ScopeBranch    MemoryScope = "branch"
	ScopeAgent     MemoryScope = "agent"
)

// MemoryEntry is structured runtime memory (spec-my-A §24.12).
type MemoryEntry struct {
	ID          string      `json:"id"`
	Scope       MemoryScope `json:"scope"`
	Type        string      `json:"type"`
	Summary     string      `json:"summary"`
	Source      string      `json:"source,omitempty"`
	Relevance   float64     `json:"relevance"`
	Tags        []string    `json:"tags,omitempty"`
	LinkedFlows    []string    `json:"linked_flows,omitempty"`
	EmbeddingJSON  string      `json:"embedding_json,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	LastUsedAt     time.Time   `json:"last_used_at"`
}

// DaemonStatus summarizes local runtime state.
type DaemonStatus struct {
	Running      bool   `json:"running"`
	PID          int    `json:"pid,omitempty"`
	Sessions     int    `json:"sessions"`
	FlowsActive  int    `json:"flows_active"`
	QueuedEvents int    `json:"queued_events"`
	MemorySize   int    `json:"memory_size"`
	DBPath       string `json:"db_path"`
	DBSizeBytes  int64  `json:"db_size_bytes"`
}

// StateGraph is a minimal session/branch/event projection.
type StateGraph struct {
	Sessions []Session      `json:"sessions"`
	Branches []Branch       `json:"branches"`
	Events   []RuntimeEvent `json:"recent_events"`
}

// HookJob is a queued hook execution unit.
type HookJob struct {
	ID        string    `json:"id"`
	EventType string    `json:"event_type"`
	Command   string    `json:"command"`
	CreatedAt time.Time `json:"created_at"`
}
