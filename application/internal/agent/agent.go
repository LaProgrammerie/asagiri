package agent

import "context"

// Capabilities describes what one agent backend can do.
type Capabilities struct {
	SupportsStreaming bool
	SupportsFiles     bool
	SupportsJSON      bool
}

// RunRequest contains one agent invocation payload.
type RunRequest struct {
	Feature    string
	TaskID     string
	Prompt     string
	WorkingDir string
	Args       []string
	Env        map[string]string
}

// RunResult captures execution output for persistence/reporting.
type RunResult struct {
	Command   string
	ExitCode  int
	Stdout    string
	Stderr    string
	DryRun    bool
	StartedAt string
	EndedAt   string
}

// Agent is the stable contract used by AgentFlow.
type Agent interface {
	Name() string
	Capabilities() Capabilities
	Run(ctx context.Context, req RunRequest) (RunResult, error)
}
