package executiongraph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GraphArtifacts lists persisted files under .asagiri/graphs/<id>/ (spec §6–7, §23).
type GraphArtifacts struct {
	Dir           string `json:"dir"`
	YAML          string `json:"yaml"`
	JSON          string `json:"json"`
	PlanMD        string `json:"plan_md"`
	MetricsJSON   string `json:"metrics_json"`
	TimelineJSONL string `json:"timeline_jsonl"`
	EventsJSONL   string `json:"events_jsonl"`
}

// MetricsSnapshot is the metrics.json skeleton written at plan time (spec §23).
type MetricsSnapshot struct {
	GraphID           string  `json:"graph_id"`
	Product           string  `json:"product"`
	Flow              string  `json:"flow,omitempty"`
	Nodes             int     `json:"nodes"`
	Edges             int     `json:"edges"`
	ParallelGroups    int     `json:"parallel_groups"`
	Checkpoints       int     `json:"checkpoints"`
	EstimatedCost     float64 `json:"estimated_cost"`
	EstimatedDuration string  `json:"estimated_duration"`
	HighestRisk       string  `json:"highest_risk"`
	BudgetStatus      string  `json:"budget_status"`
}

// SaveAll persists the graph and companion artefacts.
func (r *Repository) SaveAll(graph ExecutionGraph, schedule *ExecutionSchedule) (GraphArtifacts, error) {
	yamlPath, jsonPath, err := r.Save(graph)
	if err != nil {
		return GraphArtifacts{}, err
	}

	dir := r.graphDir(graph.ID)
	est := EstimateGraph(graph, schedule)

	planPath := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(planPath, []byte(RenderPlanMD(graph)), 0o644); err != nil {
		return GraphArtifacts{}, fmt.Errorf("write plan.md: %w", err)
	}

	metricsPath := filepath.Join(dir, "metrics.json")
	metricsBody, err := json.MarshalIndent(MetricsSnapshot{
		GraphID:           graph.ID,
		Product:           graph.Product,
		Flow:              graph.Flow,
		Nodes:             est.Nodes,
		Edges:             len(graph.Edges),
		ParallelGroups:    est.ParallelGroups,
		Checkpoints:       len(graph.Checkpoints),
		EstimatedCost:     est.EstimatedCost,
		EstimatedDuration: est.EstimatedDuration,
		HighestRisk:       string(est.HighestRisk),
		BudgetStatus:      est.BudgetStatus,
	}, "", "  ")
	if err != nil {
		return GraphArtifacts{}, fmt.Errorf("marshal metrics.json: %w", err)
	}
	if err := os.WriteFile(metricsPath, metricsBody, 0o644); err != nil {
		return GraphArtifacts{}, fmt.Errorf("write metrics.json: %w", err)
	}

	timelinePath := filepath.Join(dir, "timeline.jsonl")
	if err := initEmptyFile(timelinePath); err != nil {
		return GraphArtifacts{}, err
	}

	eventsPath := filepath.Join(dir, "events.jsonl")
	if err := initEmptyFile(eventsPath); err != nil {
		return GraphArtifacts{}, err
	}

	return GraphArtifacts{
		Dir:           dir,
		YAML:          yamlPath,
		JSON:          jsonPath,
		PlanMD:        planPath,
		MetricsJSON:   metricsPath,
		TimelineJSONL: timelinePath,
		EventsJSONL:   eventsPath,
	}, nil
}

func initEmptyFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	return os.WriteFile(path, nil, 0o644)
}
