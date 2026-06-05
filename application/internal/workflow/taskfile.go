package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/plan"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

func planToCanonical(feature string, t plan.Task) asagiri.Task {
	now := time.Now().UTC()
	task := asagiri.Task{
		ID:      t.ID,
		Title:   t.Title,
		Feature: feature,
		Status:  asagiri.StatusPending,
		Risk:    "medium",
		Type:    "implementation",
		Source: asagiri.TaskSource{
			Spec: fmt.Sprintf(".kiro/specs/%s/tasks.md", feature),
		},
		Scope: asagiri.TaskScope{
			AllowedPaths: []string{"application/**"},
		},
		Acceptance: t.Checks,
		Agents: asagiri.TaskAgents{
			Implementer: config.DefaultAgentDev,
			Reviewer:      config.DefaultAgentReviewer,
			Enricher:      config.DefaultAgentEnrich,
		},
	}
	task.TouchMetadata(now)
	if t.Status != "" {
		task.Status = t.Status
	}
	return task
}

func payloadToCanonical(payloadJSON string) (asagiri.Task, error) {
	var task asagiri.Task
	if err := json.Unmarshal([]byte(payloadJSON), &task); err != nil {
		return asagiri.Task{}, err
	}
	return task, nil
}

func canonicalToPayload(task asagiri.Task) (string, error) {
	body, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *Service) persistCanonicalTaskFiles(feature string, tasks []asagiri.Task) error {
	dir := filepath.Join(s.repoRoot, ".asagiri", "tasks", feature)
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
		from = asagiri.StatusPending
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
