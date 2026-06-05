package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func (s *Store) GetNode(ctx context.Context, id string) (knowledge.GraphNode, error) {
	if err := knowledge.ValidateNodeID(id); err != nil {
		return knowledge.GraphNode{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, type, name, path,
       source_kind, source_path, source_extractor, source_evidence,
       confidence, created_at, updated_at
FROM nodes WHERE id = ?`, id)
	node, err := scanNode(row)
	if err == sql.ErrNoRows {
		return knowledge.GraphNode{}, fmt.Errorf("%w: node %q", knowledge.ErrNotFound, id)
	}
	if err != nil {
		return knowledge.GraphNode{}, err
	}
	props, err := s.loadNodeProperties(ctx, id)
	if err != nil {
		return knowledge.GraphNode{}, err
	}
	node.Properties = props
	return node, nil
}

func (s *Store) GetEdge(ctx context.Context, id string) (knowledge.GraphEdge, error) {
	if err := knowledge.ValidateEdgeID(id); err != nil {
		return knowledge.GraphEdge{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, from_node_id, to_node_id, type,
       source_kind, source_path, source_extractor, source_evidence,
       confidence, created_at, updated_at
FROM edges WHERE id = ?`, id)
	edge, err := scanEdge(row)
	if err == sql.ErrNoRows {
		return knowledge.GraphEdge{}, fmt.Errorf("%w: edge %q", knowledge.ErrNotFound, id)
	}
	if err != nil {
		return knowledge.GraphEdge{}, err
	}
	props, err := s.loadEdgeProperties(ctx, id)
	if err != nil {
		return knowledge.GraphEdge{}, err
	}
	edge.Properties = props
	return edge, nil
}

func (s *Store) ListNodes(ctx context.Context, filter knowledge.NodeFilter) ([]knowledge.GraphNode, error) {
	if filter.ID != "" {
		if err := knowledge.ValidateNodeID(filter.ID); err != nil {
			return nil, err
		}
	}
	var clauses []string
	var args []any
	if filter.ID != "" {
		clauses = append(clauses, "id = ?")
		args = append(args, filter.ID)
	}
	if filter.Type != "" {
		clauses = append(clauses, "type = ?")
		args = append(args, string(filter.Type))
	}
	if filter.Path != "" {
		clauses = append(clauses, "path = ?")
		args = append(args, filter.Path)
	}
	if filter.PathLike != "" {
		clauses = append(clauses, "path LIKE ?")
		args = append(args, filter.PathLike+"%")
	}
	query := `SELECT id, type, name, path,
       source_kind, source_path, source_extractor, source_evidence,
       confidence, created_at, updated_at
FROM nodes`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var nodes []knowledge.GraphNode
	for rows.Next() {
		node, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		props, err := s.loadNodeProperties(ctx, node.ID)
		if err != nil {
			return nil, err
		}
		node.Properties = props
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Store) ListEdges(ctx context.Context, filter knowledge.EdgeFilter) ([]knowledge.GraphEdge, error) {
	if filter.ID != "" {
		if err := knowledge.ValidateEdgeID(filter.ID); err != nil {
			return nil, err
		}
	}
	if filter.FromNodeID != "" {
		if err := knowledge.ValidateNodeID(filter.FromNodeID); err != nil {
			return nil, err
		}
	}
	if filter.ToNodeID != "" {
		if err := knowledge.ValidateNodeID(filter.ToNodeID); err != nil {
			return nil, err
		}
	}
	var clauses []string
	var args []any
	if filter.ID != "" {
		clauses = append(clauses, "id = ?")
		args = append(args, filter.ID)
	}
	if filter.Type != "" {
		clauses = append(clauses, "type = ?")
		args = append(args, string(filter.Type))
	}
	if filter.FromNodeID != "" {
		clauses = append(clauses, "from_node_id = ?")
		args = append(args, filter.FromNodeID)
	}
	if filter.ToNodeID != "" {
		clauses = append(clauses, "to_node_id = ?")
		args = append(args, filter.ToNodeID)
	}
	query := `SELECT id, from_node_id, to_node_id, type,
       source_kind, source_path, source_extractor, source_evidence,
       confidence, created_at, updated_at
FROM edges`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list edges: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var edges []knowledge.GraphEdge
	for rows.Next() {
		edge, err := scanEdge(rows)
		if err != nil {
			return nil, err
		}
		props, err := s.loadEdgeProperties(ctx, edge.ID)
		if err != nil {
			return nil, err
		}
		edge.Properties = props
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}

func (s *Store) LoadGraph(ctx context.Context) (knowledge.KnowledgeGraph, error) {
	nodes, err := s.ListNodes(ctx, knowledge.NodeFilter{})
	if err != nil {
		return knowledge.KnowledgeGraph{}, err
	}
	edges, err := s.ListEdges(ctx, knowledge.EdgeFilter{})
	if err != nil {
		return knowledge.KnowledgeGraph{}, err
	}
	return knowledge.KnowledgeGraph{Nodes: nodes, Edges: edges}, nil
}

func (s *Store) SetIndexMetadata(ctx context.Context, key string, value any) error {
	if err := validateMetadataKey(key); err != nil {
		return err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("index metadata marshal: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `
INSERT INTO index_metadata (key, value_json, updated_at) VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value_json = excluded.value_json, updated_at = excluded.updated_at`,
		key, string(body), now,
	)
	if err != nil {
		return fmt.Errorf("set index metadata: %w", err)
	}
	return nil
}

func (s *Store) GetIndexMetadata(ctx context.Context, key string) (map[string]any, error) {
	if err := validateMetadataKey(key); err != nil {
		return nil, err
	}
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT value_json FROM index_metadata WHERE key = ?`, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: metadata %q", knowledge.ErrNotFound, key)
	}
	if err != nil {
		return nil, fmt.Errorf("get index metadata: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse index metadata: %w", err)
	}
	return out, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNode(row rowScanner) (knowledge.GraphNode, error) {
	var node knowledge.GraphNode
	var nodeType string
	var createdAt, updatedAt string
	if err := row.Scan(
		&node.ID, &nodeType, &node.Name, &node.Path,
		&node.Source.Kind, &node.Source.Path, &node.Source.Extractor, &node.Source.Evidence,
		&node.Confidence, &createdAt, &updatedAt,
	); err != nil {
		return knowledge.GraphNode{}, fmt.Errorf("scan node: %w", err)
	}
	node.Type = knowledge.NodeType(nodeType)
	node.CreatedAt = parseTime(createdAt)
	node.UpdatedAt = parseTime(updatedAt)
	return node, nil
}

func scanEdge(row rowScanner) (knowledge.GraphEdge, error) {
	var edge knowledge.GraphEdge
	var edgeType string
	var createdAt, updatedAt string
	if err := row.Scan(
		&edge.ID, &edge.From, &edge.To, &edgeType,
		&edge.Source.Kind, &edge.Source.Path, &edge.Source.Extractor, &edge.Source.Evidence,
		&edge.Confidence, &createdAt, &updatedAt,
	); err != nil {
		return knowledge.GraphEdge{}, fmt.Errorf("scan edge: %w", err)
	}
	edge.Type = knowledge.EdgeType(edgeType)
	edge.CreatedAt = parseTime(createdAt)
	edge.UpdatedAt = parseTime(updatedAt)
	return edge, nil
}

func (s *Store) loadNodeProperties(ctx context.Context, nodeID string) (map[string]any, error) {
	return loadProperties(ctx, s.db, `SELECT key, value_json FROM node_properties WHERE node_id = ? ORDER BY key`, nodeID)
}

func (s *Store) loadEdgeProperties(ctx context.Context, edgeID string) (map[string]any, error) {
	return loadProperties(ctx, s.db, `SELECT key, value_json FROM edge_properties WHERE edge_id = ? ORDER BY key`, edgeID)
}

func loadProperties(ctx context.Context, db *sql.DB, query, id string) (map[string]any, error) {
	rows, err := db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("load properties: %w", err)
	}
	defer func() { _ = rows.Close() }()
	props := make(map[string]any)
	for rows.Next() {
		var key, raw string
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			return nil, fmt.Errorf("parse property %q: %w", key, err)
		}
		props[key] = value
	}
	if len(props) == 0 {
		return nil, rows.Err()
	}
	return props, rows.Err()
}

func validateMetadataKey(key string) error {
	if key == "" {
		return fmt.Errorf("index metadata: key required")
	}
	if strings.Contains(key, "..") || strings.ContainsAny(key, `/\`) {
		return fmt.Errorf("index metadata: invalid key %q", key)
	}
	return nil
}

func parseTime(raw string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, raw)
	}
	return t
}
