package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func (s *Store) UpsertNode(ctx context.Context, node knowledge.GraphNode) error {
	if err := node.Validate(); err != nil {
		return fmt.Errorf("upsert node: %w", err)
	}
	now := time.Now().UTC()
	if node.CreatedAt.IsZero() {
		node.CreatedAt = now
	}
	node.UpdatedAt = now

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("upsert node tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existingCreated string
	err = tx.QueryRowContext(ctx, `SELECT created_at FROM nodes WHERE id = ?`, node.ID).Scan(&existingCreated)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		return fmt.Errorf("upsert node lookup: %w", err)
	default:
		if t, parseErr := time.Parse(time.RFC3339Nano, existingCreated); parseErr == nil {
			node.CreatedAt = t
		}
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO nodes (
    id, type, name, path,
    source_kind, source_path, source_extractor, source_evidence,
    confidence, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    type = excluded.type,
    name = excluded.name,
    path = excluded.path,
    source_kind = excluded.source_kind,
    source_path = excluded.source_path,
    source_extractor = excluded.source_extractor,
    source_evidence = excluded.source_evidence,
    confidence = excluded.confidence,
    updated_at = excluded.updated_at`,
		node.ID, string(node.Type), node.Name, node.Path,
		node.Source.Kind, node.Source.Path, node.Source.Extractor, node.Source.Evidence,
		node.Confidence, formatTime(node.CreatedAt), formatTime(node.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert node row: %w", err)
	}
	if err := replaceNodeProperties(ctx, tx, node.ID, node.Properties); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpsertEdge(ctx context.Context, edge knowledge.GraphEdge) error {
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("upsert edge: %w", err)
	}

	var fromExists, toExists int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE id = ?`, edge.From).Scan(&fromExists); err != nil {
		return fmt.Errorf("upsert edge from check: %w", err)
	}
	if fromExists == 0 {
		return fmt.Errorf("upsert edge: %w: from node %q", knowledge.ErrNotFound, edge.From)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes WHERE id = ?`, edge.To).Scan(&toExists); err != nil {
		return fmt.Errorf("upsert edge to check: %w", err)
	}
	if toExists == 0 {
		return fmt.Errorf("upsert edge: %w: to node %q", knowledge.ErrNotFound, edge.To)
	}

	now := time.Now().UTC()
	if edge.CreatedAt.IsZero() {
		edge.CreatedAt = now
	}
	edge.UpdatedAt = now

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("upsert edge tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var existingCreated string
	err = tx.QueryRowContext(ctx, `SELECT created_at FROM edges WHERE id = ?`, edge.ID).Scan(&existingCreated)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		return fmt.Errorf("upsert edge lookup: %w", err)
	default:
		if t, parseErr := time.Parse(time.RFC3339Nano, existingCreated); parseErr == nil {
			edge.CreatedAt = t
		}
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO edges (
    id, from_node_id, to_node_id, type,
    source_kind, source_path, source_extractor, source_evidence,
    confidence, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    from_node_id = excluded.from_node_id,
    to_node_id = excluded.to_node_id,
    type = excluded.type,
    source_kind = excluded.source_kind,
    source_path = excluded.source_path,
    source_extractor = excluded.source_extractor,
    source_evidence = excluded.source_evidence,
    confidence = excluded.confidence,
    updated_at = excluded.updated_at`,
		edge.ID, edge.From, edge.To, string(edge.Type),
		edge.Source.Kind, edge.Source.Path, edge.Source.Extractor, edge.Source.Evidence,
		edge.Confidence, formatTime(edge.CreatedAt), formatTime(edge.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert edge row: %w", err)
	}
	if err := replaceEdgeProperties(ctx, tx, edge.ID, edge.Properties); err != nil {
		return err
	}
	return tx.Commit()
}

func replaceNodeProperties(ctx context.Context, tx *sql.Tx, nodeID string, props map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM node_properties WHERE node_id = ?`, nodeID); err != nil {
		return fmt.Errorf("clear node properties: %w", err)
	}
	return insertProperties(ctx, tx, "node_properties", nodeID, props)
}

func replaceEdgeProperties(ctx context.Context, tx *sql.Tx, edgeID string, props map[string]any) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM edge_properties WHERE edge_id = ?`, edgeID); err != nil {
		return fmt.Errorf("clear edge properties: %w", err)
	}
	return insertProperties(ctx, tx, "edge_properties", edgeID, props)
}

func insertProperties(ctx context.Context, tx *sql.Tx, table, id string, props map[string]any) error {
	if len(props) == 0 {
		return nil
	}
	for key, value := range props {
		body, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("marshal property %q: %w", key, err)
		}
		var stmt string
		switch table {
		case "node_properties":
			stmt = `INSERT INTO node_properties (node_id, key, value_json) VALUES (?, ?, ?)`
		case "edge_properties":
			stmt = `INSERT INTO edge_properties (edge_id, key, value_json) VALUES (?, ?, ?)`
		default:
			return fmt.Errorf("unknown properties table %q", table)
		}
		if _, err := tx.ExecContext(ctx, stmt, id, key, string(body)); err != nil {
			return fmt.Errorf("insert property %q: %w", key, err)
		}
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
