# agent/claudecode

First-class adapter for the [Claude Code](https://docs.anthropic.com/en/docs/claude-code) CLI agent.

## Why a dedicated package

Claude Code requires specific CLI flags for non-interactive use. Without this adapter, those flags would be scattered across configuration examples and documentation. This package is the single place for all Claude-specific behavior.

## Usage

### In config.yaml

```yaml
agents:
  claude-code:
    command: claude
    args: ["--print", "--output-format", "stream-json"]
    timeout: 600

work:
  default_agent: claude-code
```

### In Go code

```go
import "github.com/LaProgrammerie/asagiri/application/internal/agent/claudecode"

// Get the recommended config preset
cfg := claudecode.DefaultConfig()

// Build an adapter (implements agent.Agent)
a, err := claudecode.New(cfg, dryRun)
```

## What it encapsulates

| Concern | Detail |
|---------|--------|
| Non-interactive mode | `--print` flag — exits after producing output |
| Output format | `--output-format stream-json` — one JSON object per line |
| Timeout | 10 min default, applied per invocation via `context.WithTimeout` |
| Output extraction | `extractLastJSON` picks the final result object from the stream |
| Agent contract | Implements `agent.Agent` — drop-in with `exec.Executor` |

## Constants

```go
claudecode.DefaultCommand  // "claude"
claudecode.DefaultTimeout  // 10 * time.Minute
claudecode.AgentName       // "claude-code"
```

## No hardcoding outside this package

All Claude-specific logic lives here. The rest of the codebase treats `claude-code` as a named entry in `config.agents`, identical to `cursor`, `codex`, or `ollama`.
