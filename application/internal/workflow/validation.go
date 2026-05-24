package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	appvalidate "github.com/LaProgrammerie/hyper-fast-builder/application/internal/validation"
)

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

func (s *Service) runVerification(ctx context.Context, dir, payloadJSON string) error {
	if s.dryRun {
		return nil
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
	_, err := appvalidate.NewRunner(s.dryRun).Run(ctx, dir, commands)
	return err
}

