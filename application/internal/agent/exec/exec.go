package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	osExec "os/exec"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/redact"
)

// Executor runs one configured agent command as a subprocess.
type Executor struct {
	name    string
	command string
	args    []string
	dryRun  bool
}

func New(name string, cfg config.Agent, dryRun bool) (*Executor, error) {
	if cfg.Command == "" {
		return nil, fmt.Errorf("agent %q: commande manquante", name)
	}
	return &Executor{
		name:    name,
		command: cfg.Command,
		args:    append([]string{}, cfg.Args...),
		dryRun:  dryRun || env.DryRunEnabled(),
	}, nil
}

func (e *Executor) Name() string {
	return e.name
}

func (e *Executor) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		SupportsStreaming: false,
		SupportsFiles:     true,
		SupportsJSON:      false,
	}
}

func (e *Executor) Run(ctx context.Context, req agent.RunRequest) (agent.RunResult, error) {
	started := time.Now().UTC()
	finalArgs := append([]string{}, e.args...)
	finalArgs = append(finalArgs, req.Args...)

	cmdLine := strings.TrimSpace(e.command + " " + strings.Join(finalArgs, " "))
	if e.dryRun {
		ended := time.Now().UTC()
		return agent.RunResult{
			Command:   cmdLine,
			ExitCode:  0,
			Stdout:    "dry-run: commande agent non exécutée",
			Stderr:    "",
			DryRun:    true,
			StartedAt: started.Format(time.RFC3339Nano),
			EndedAt:   ended.Format(time.RFC3339Nano),
		}, nil
	}

	cmd := osExec.CommandContext(ctx, e.command, finalArgs...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	env := os.Environ()
	for k, v := range req.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	if req.Prompt != "" {
		cmd.Stdin = strings.NewReader(req.Prompt)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	ended := time.Now().UTC()
	result := agent.RunResult{
		Command:   cmdLine,
		ExitCode:  0,
		Stdout:    redact.String(stdoutBuf.String()),
		Stderr:    redact.String(stderrBuf.String()),
		DryRun:    false,
		StartedAt: started.Format(time.RFC3339Nano),
		EndedAt:   ended.Format(time.RFC3339Nano),
	}

	if runErr == nil {
		return result, nil
	}

	var exitErr *osExec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, fmt.Errorf("agent %q: sortie non nulle (%d)", e.name, result.ExitCode)
	}
	return result, fmt.Errorf("agent %q: %w", e.name, runErr)
}
