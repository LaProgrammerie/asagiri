// Package claudecode provides a first-class adapter for the Claude Code CLI agent.
// It encapsulates Claude-specific CLI flags, non-interactive mode, output parsing,
// and timeout handling so that Claude-specific behavior stays in one place.
package claudecode

import (
	"context"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/exec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const (
	// DefaultCommand is the Claude Code CLI binary name.
	DefaultCommand = "claude"
	// DefaultTimeout is the per-invocation timeout.
	DefaultTimeout = 10 * time.Minute
	// AgentName is the canonical name used in config and routing.
	AgentName = "claude-code"
)

// DefaultConfig returns a config.Agent preset for Claude Code that callers can
// embed into their config.yaml or pass directly to New().
func DefaultConfig() config.Agent {
	return config.Agent{
		Command: DefaultCommand,
		// --print: non-interactive, outputs result to stdout and exits.
		// --output-format stream-json: structured output for parsing.
		Args: []string{"--print", "--output-format", "stream-json"},
	}
}

// Adapter wraps exec.Executor with Claude-specific output handling.
type Adapter struct {
	inner   *exec.Executor
	timeout time.Duration
}

// New returns an Adapter for the given config.Agent.
// Pass DefaultConfig() to use Claude Code with recommended defaults.
func New(cfg config.Agent, dryRun bool) (*Adapter, error) {
	if cfg.Command == "" {
		cfg.Command = DefaultCommand
	}
	e, err := exec.New(AgentName, cfg, dryRun)
	if err != nil {
		return nil, err
	}
	return &Adapter{inner: e, timeout: DefaultTimeout}, nil
}

// Name implements agent.Agent.
func (a *Adapter) Name() string { return AgentName }

// Capabilities reports Claude Code's supported features.
func (a *Adapter) Capabilities() agent.Capabilities {
	return agent.Capabilities{
		SupportsStreaming: false, // --print mode is batch
		SupportsFiles:    true,
		SupportsJSON:     true, // stream-json output
	}
}

// Run executes Claude Code non-interactively and returns the result.
// A per-invocation timeout is applied; cancellation propagates from ctx.
func (a *Adapter) Run(ctx context.Context, req agent.RunRequest) (agent.RunResult, error) {
	tctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	result, err := a.inner.Run(tctx, req)
	if err != nil {
		return result, err
	}

	result.Stdout = extractLastJSON(result.Stdout)
	return result, nil
}

// extractLastJSON returns the last JSON object from a stream-json output.
// Claude Code emits one JSON object per line; we want the final result object.
func extractLastJSON(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "{") {
			return line
		}
	}
	return output
}
