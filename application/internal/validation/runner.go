package validation

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Runner executes validation commands without a shell (spec §11.2 Validator).
type Runner struct {
	dryRun bool
}

func NewRunner(dryRun bool) *Runner {
	return &Runner{dryRun: dryRun}
}

// FromConfig builds validation commands from config.
func FromConfig(cfg *config.Config) []Command {
	out := make([]Command, 0, len(cfg.Validation.Commands))
	for _, c := range cfg.Validation.Commands {
		out = append(out, Command{Name: c.Name, Line: c.Command, Required: c.Required})
	}
	return out
}

// Command is one validation step.
type Command struct {
	Name     string
	Line     string
	Required bool
}

// Result holds one command outcome.
type Result struct {
	Name     string
	Command  string
	ExitCode int
	Output   string
	Err      error
}

// Run executes commands in dir.
func (r *Runner) Run(ctx context.Context, dir string, commands []Command) ([]Result, error) {
	if r.dryRun {
		return nil, nil
	}
	results := make([]Result, 0, len(commands))
	for _, c := range commands {
		res := r.runOne(ctx, dir, c)
		results = append(results, res)
		if res.Err != nil && c.Required {
			return results, res.Err
		}
	}
	return results, nil
}

func (r *Runner) runOne(ctx context.Context, dir string, c Command) Result {
	line := strings.TrimSpace(c.Line)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Result{Name: c.Name, Command: line}
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	res := Result{
		Name:     c.Name,
		Command:  line,
		Output:   strings.TrimSpace(string(out)),
	}
	if err != nil {
		res.Err = fmt.Errorf("%s: %w", line, err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
	}
	return res
}
