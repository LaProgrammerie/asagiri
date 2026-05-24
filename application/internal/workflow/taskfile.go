package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/plan"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
	"gopkg.in/yaml.v3"
)

func planToCanonical(feature string, t plan.Task) agentflow.Task {
	now := time.Now().UTC()
	task := agentflow.Task{
		ID:      t.ID,
		Title:   t.Title,
		Feature: feature,
		Status:  agentflow.StatusPending,
		Risk:    "medium",
		Type:    "implementation",
		Source: agentflow.TaskSource{
			Spec: fmt.Sprintf(".kiro/specs/%s/tasks.md", feature),
		},
		Scope: agentflow.TaskScope{
			AllowedPaths: []string{"application/**"},
		},
		Acceptance: t.Checks,
		Agents: agentflow.TaskAgents{
			Implementer: "cursor",
			Reviewer:      "codex",
			Enricher:      "ollama",
		},
	}
	task.TouchMetadata(now)
	if t.Status != "" {
		task.Status = t.Status
	}
	return task
}

func payloadToCanonical(payloadJSON string) (agentflow.Task, error) {
	var task agentflow.Task
	if err := json.Unmarshal([]byte(payloadJSON), &task); err != nil {
		return agentflow.Task{}, err
	}
	return task, nil
}

func canonicalToPayload(task agentflow.Task) (string, error) {
	body, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *Service) persistCanonicalTaskFiles(feature string, tasks []agentflow.Task) error {
	dir := filepath.Join(s.repoRoot, ".agentflow", "tasks", feature)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create tasks dir: %w", err)
	}
	for _, task := range tasks {
		path := filepath.Join(dir, task.ID+".yaml")
		body, err := yaml.Marshal(task)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return fmt.Errorf("write task yaml %s: %w", path, err)
		}
		payload, err := canonicalToPayload(task)
		if err != nil {
			return err
		}
		jsonPath := filepath.Join(dir, task.ID+".json")
		if err := os.WriteFile(jsonPath, []byte(payload+"\n"), 0o644); err != nil {
			return fmt.Errorf("write task json %s: %w", path, err)
		}
	}
	return nil
}

func (s *Service) transitionTask(task sqlite.Task, to string, force bool) error {
	from := task.Status
	if from == "" {
		from = agentflow.StatusPending
	}
	if err := TransitionTask(from, to, force); err != nil {
		return err
	}
	canonical, _ := payloadToCanonical(task.PayloadJSON)
	canonical.Status = to
	canonical.TouchMetadata(time.Now().UTC())
	payload, err := canonicalToPayload(canonical)
	if err != nil {
		return err
	}
	return s.store.UpdateTask(&sqlite.Task{ID: task.ID, Status: to, PayloadJSON: payload})
}
