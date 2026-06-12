package executiongraph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Repository persists execution graphs under .asagiri/graphs/<id>/ (spec §6).
type Repository struct {
	RepoRoot string
}

// NewRepository creates a graph repository rooted at repoRoot.
func NewRepository(repoRoot string) *Repository {
	return &Repository{RepoRoot: repoRoot}
}

func (r *Repository) graphDir(graphID string) string {
	return filepath.Join(r.RepoRoot, ".asagiri", "graphs", graphID)
}

// Save writes execution-graph.yaml and execution-graph.json after validation.
func (r *Repository) Save(graph ExecutionGraph) (yamlPath, jsonPath string, err error) {
	if err := graph.Validate(); err != nil {
		return "", "", fmt.Errorf("save execution graph: %w", err)
	}
	dir := r.graphDir(graph.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create graph dir: %w", err)
	}

	yamlPath = filepath.Join(dir, "execution-graph.yaml")
	yamlBody, err := yaml.Marshal(graph)
	if err != nil {
		return "", "", fmt.Errorf("marshal execution graph yaml: %w", err)
	}
	if err := os.WriteFile(yamlPath, yamlBody, 0o644); err != nil {
		return "", "", fmt.Errorf("write execution graph yaml: %w", err)
	}

	jsonPath = filepath.Join(dir, "execution-graph.json")
	jsonBody, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal execution graph json: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonBody, 0o644); err != nil {
		return "", "", fmt.Errorf("write execution graph json: %w", err)
	}

	return yamlPath, jsonPath, nil
}

// Load reads an execution graph from YAML, falling back to JSON.
func (r *Repository) Load(graphID string) (ExecutionGraph, error) {
	if err := ValidateGraphID(graphID); err != nil {
		return ExecutionGraph{}, fmt.Errorf("load execution graph: %w", err)
	}
	dir := r.graphDir(graphID)
	yamlPath := filepath.Join(dir, "execution-graph.yaml")
	if body, err := os.ReadFile(yamlPath); err == nil {
		var graph ExecutionGraph
		if err := yaml.Unmarshal(body, &graph); err != nil {
			return ExecutionGraph{}, fmt.Errorf("parse execution graph yaml: %w", err)
		}
		return finalizeLoad(graphID, graph)
	} else if !os.IsNotExist(err) {
		return ExecutionGraph{}, fmt.Errorf("read execution graph yaml: %w", err)
	}

	jsonPath := filepath.Join(dir, "execution-graph.json")
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		return ExecutionGraph{}, fmt.Errorf("read execution graph: %w", err)
	}
	var graph ExecutionGraph
	if err := json.Unmarshal(body, &graph); err != nil {
		return ExecutionGraph{}, fmt.Errorf("parse execution graph json: %w", err)
	}
	return finalizeLoad(graphID, graph)
}

func finalizeLoad(graphID string, graph ExecutionGraph) (ExecutionGraph, error) {
	if graph.ID != graphID {
		return ExecutionGraph{}, fmt.Errorf("load execution graph: id mismatch: requested %q, file has %q", graphID, graph.ID)
	}
	if err := graph.Validate(); err != nil {
		return ExecutionGraph{}, fmt.Errorf("load execution graph: %w", err)
	}
	return graph, nil
}

// ParseYAML unmarshals graph YAML bytes.
func ParseYAML(body []byte) (ExecutionGraph, error) {
	var graph ExecutionGraph
	if err := yaml.Unmarshal(body, &graph); err != nil {
		return ExecutionGraph{}, fmt.Errorf("parse execution graph yaml: %w", err)
	}
	return graph, nil
}

// ParseJSON unmarshals graph JSON bytes.
func ParseJSON(body []byte) (ExecutionGraph, error) {
	var graph ExecutionGraph
	if err := json.Unmarshal(body, &graph); err != nil {
		return ExecutionGraph{}, fmt.Errorf("parse execution graph json: %w", err)
	}
	return graph, nil
}
