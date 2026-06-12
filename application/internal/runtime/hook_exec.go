package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecuteHookCommand runs a whitelisted hook command (asa-only, spec-my-A §24.9).
func ExecuteHookCommand(ctx context.Context, repoRoot, command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("hook: empty command")
	}
	if !strings.HasPrefix(command, "asa ") {
		return fmt.Errorf("hook: only `asa ...` commands are allowed, got %q", command)
	}
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("hook: invalid command")
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook %q: %w: %s", command, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ProcessHookQueue executes pending hooks (worker loop).
func (s *Store) ProcessHookQueue(ctx context.Context, max int) (int, error) {
	jobs, err := s.DequeueHooks(max)
	if err != nil {
		return 0, err
	}
	var done int
	for _, j := range jobs {
		if ctx.Err() != nil {
			return done, ctx.Err()
		}
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		err := ExecuteHookCommand(runCtx, s.repoRoot, j.Command)
		cancel()
		status := "done"
		if err != nil {
			status = "failed"
			_, _ = s.EmitEvent("hook.failed", "hooks", "", "", map[string]any{"command": j.Command, "error": err.Error()})
		} else {
			_, _ = s.EmitEvent("hook.executed", "hooks", "", "", map[string]any{"command": j.Command})
		}
		_ = s.MarkHookDone(j.ID, status)
		done++
	}
	return done, nil
}
