package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	appvalidate "github.com/LaProgrammerie/asagiri/application/internal/validation"
)

// validationEvidenceDocument is persisted under .asagiri/logs/<task-id>/validation/results.json.
type validationEvidenceDocument struct {
	TaskID   string                      `json:"task_id"`
	Worktree string                      `json:"worktree,omitempty"`
	DryRun   bool                        `json:"dry_run,omitempty"`
	At       string                      `json:"at"`
	Commands []validationEvidenceCommand `json:"commands"`
}

type validationEvidenceCommand struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output,omitempty"`
}

func validationResultsPath(repoRoot, taskID string) string {
	return filepath.Join(repoRoot, ".asagiri", "logs", taskID, "validation", "results.json")
}

func (s *Service) persistValidationEvidence(taskID, worktreePath string, results []appvalidate.Result) error {
	doc := validationEvidenceDocument{
		TaskID:   taskID,
		Worktree: worktreePath,
		DryRun:   s.dryRun,
		At:       time.Now().UTC().Format(time.RFC3339),
		Commands: make([]validationEvidenceCommand, 0, len(results)),
	}
	for _, r := range results {
		doc.Commands = append(doc.Commands, validationEvidenceCommand{
			Name:     r.Name,
			Command:  r.Command,
			ExitCode: r.ExitCode,
			Output:   r.Output,
		})
	}
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal validation evidence: %w", err)
	}
	body = append(body, '\n')
	dir := filepath.Dir(validationResultsPath(s.repoRoot, taskID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create validation log dir: %w", err)
	}
	path := validationResultsPath(s.repoRoot, taskID)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write validation evidence: %w", err)
	}
	return nil
}

type taskPayload struct {
	ValidationCommands []string `json:"validation_commands"`
}

func parseValidationCommands(payloadJSON string) []string {
	if strings.TrimSpace(payloadJSON) == "" {
		return nil
	}
	var p taskPayload
	if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
		return nil
	}
	return p.ValidationCommands
}

func validationLinesForRepo(repoRoot string) []string {
	if cmds := config.DefaultGoValidationCommandsForRepo(repoRoot); len(cmds) > 0 {
		lines := make([]string, 0, len(cmds))
		for _, c := range cmds {
			lines = append(lines, c.Command)
		}
		return lines
	}
	return []string{"go test ./...", "go vet ./..."}
}

func (s *Service) runVerification(ctx context.Context, dir, payloadJSON string) ([]appvalidate.Result, error) {
	if s.dryRun {
		return nil, nil
	}
	cmds := parseValidationCommands(payloadJSON)
	var commands []appvalidate.Command
	if len(cmds) > 0 {
		for i, line := range cmds {
			commands = append(commands, appvalidate.Command{Name: fmt.Sprintf("task-%d", i), Line: line, Required: true})
		}
	} else {
		commands = appvalidate.FromConfig(s.cfg)
	}
	return appvalidate.NewRunner(s.dryRun).Run(ctx, dir, commands)
}
