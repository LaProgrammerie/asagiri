CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    product_id TEXT,
    flow_id TEXT,
    branch_id TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    context_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS branches (
    id TEXT PRIMARY KEY,
    parent_branch_id TEXT,
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    branch_type TEXT NOT NULL DEFAULT 'flow',
    description TEXT,
    divergence_json TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS runtime_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    source TEXT,
    session_id TEXT,
    flow_id TEXT,
    payload_json TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS memory_entries (
    id TEXT PRIMARY KEY,
    scope TEXT NOT NULL,
    entry_type TEXT NOT NULL,
    summary TEXT NOT NULL,
    source TEXT,
    relevance REAL NOT NULL DEFAULT 0.5,
    tags_json TEXT,
    linked_flows_json TEXT,
    created_at TEXT NOT NULL,
    last_used_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS flow_states (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    flow_id TEXT NOT NULL,
    state_json TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runtime_metrics (
    key TEXT PRIMARY KEY,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runtime_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runtime_events_created ON runtime_events(created_at);
CREATE INDEX IF NOT EXISTS idx_runtime_events_session ON runtime_events(session_id);
