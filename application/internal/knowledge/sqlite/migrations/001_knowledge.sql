CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL
);

INSERT INTO schema_version (version) SELECT 1
WHERE NOT EXISTS (SELECT 1 FROM schema_version);

CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    source_kind TEXT NOT NULL DEFAULT '',
    source_path TEXT NOT NULL DEFAULT '',
    source_extractor TEXT NOT NULL DEFAULT '',
    source_evidence TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes (type);
CREATE INDEX IF NOT EXISTS idx_nodes_path ON nodes (path);

CREATE TABLE IF NOT EXISTS edges (
    id TEXT PRIMARY KEY,
    from_node_id TEXT NOT NULL,
    to_node_id TEXT NOT NULL,
    type TEXT NOT NULL,
    source_kind TEXT NOT NULL DEFAULT '',
    source_path TEXT NOT NULL DEFAULT '',
    source_extractor TEXT NOT NULL DEFAULT '',
    source_evidence TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (from_node_id) REFERENCES nodes (id) ON DELETE CASCADE,
    FOREIGN KEY (to_node_id) REFERENCES nodes (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_edges_from ON edges (from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_to ON edges (to_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_type ON edges (type);

CREATE TABLE IF NOT EXISTS node_properties (
    node_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value_json TEXT NOT NULL,
    PRIMARY KEY (node_id, key),
    FOREIGN KEY (node_id) REFERENCES nodes (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS edge_properties (
    edge_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value_json TEXT NOT NULL,
    PRIMARY KEY (edge_id, key),
    FOREIGN KEY (edge_id) REFERENCES edges (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS snapshots (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    graph_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS index_metadata (
    key TEXT PRIMARY KEY,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
