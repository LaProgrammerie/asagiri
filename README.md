# Asagiri

Deterministic orchestration for AI coding workflows.

Asagiri turns a spec into auditable, cost-aware development runs — using local investigation, git worktrees, external coding agents, validation, and reproducible reports.

---

## Quick Start

```bash
# 1. Install
brew tap LaProgrammerie/tap && brew install asa

# 2. Bootstrap your repo (once)
asa onboard

# 3. Describe what you want
asa work "add Stripe payment support"

# 4. Check what to do next
asa next

# 5. See the current state
asa status
```

That's the core loop. Everything else builds on top of it.

---

## Install

### Homebrew (macOS / Linux)

```bash
brew tap LaProgrammerie/tap
brew install asa
asa version
```

### Manual (released binary)

Download from [GitHub Releases](https://github.com/LaProgrammerie/asagiri/releases):

```bash
VERSION=v0.1.0
curl -LO "https://github.com/LaProgrammerie/asagiri/releases/download/${VERSION}/asa_Darwin_arm64.tar.gz"
tar -xzf asa_Darwin_arm64.tar.gz
sudo install -m 755 asa /usr/local/bin/asa
asa doctor
```

### From source

Requires Go 1.25+, `git`, `make`.

```bash
git clone https://github.com/LaProgrammerie/asagiri.git
cd asagiri
make build
./bin/asa version
```

---

## Daily Workflow

```bash
# Describe → plan → implement → validate in one command
asa work "add Stripe support" --estimate-only   # preview cost first
asa work "add Stripe support"                   # execute

# Resume after an interruption
asa continue

# See the recommended next step
asa next --feature stripe-support

# Show all runs
asa status

# Full report for a run
asa report <run-id>
```

### Primitive pipeline (when you want explicit control)

```bash
asa spec stripe-support       # produce a spec
asa plan stripe-support       # break into tasks
asa enrich stripe-support     # enrich with local context
asa dev stripe-support        # implement
asa verify stripe-support     # run validation commands
asa review stripe-support     # independent review
asa pr stripe-support         # prepare PR diff
```

### Try without real agents (CI / exploration)

```bash
export ASA_DRY_RUN=1
asa onboard
asa work "my feature" --plan-only --yes
```

---

## Advanced Tools

```bash
asa tools          # discovery catalog
```

Once you understand the core workflow, advanced tools unlock:

| Command | Purpose |
|---------|---------|
| `asa trust` | Execution gates and verified replay |
| `asa replay` | Capture, replay and diff workflows |
| `asa knowledge` | Engineering knowledge graph |
| `asa graph` | Multi-agent execution graphs |
| `asa investigate` | Structured local investigation |
| `asa estimate` | Token / cost preview without execution |
| `asa sync` | Import specs from Notion or local sources |

---

## Architecture

- **Local-first** — investigate and optimize context before calling cloud models
- **Cost-aware** — estimate tokens and budget before execution
- **Isolated** — one git worktree per task, no global state pollution
- **Observable** — SQLite state, full run reports, explicit validation steps
- **Provider-agnostic** — cursor, codex, kiro, claude-code, ollama or any subprocess

Docs: [Intent layer](docs/ai/archives/specs/specv2.md) · [Cost/perf](docs/ai/archives/specs/specv3.md) · [Architecture](docs/ai/02-architecture.md)

---

## Documentation

Public docs (multilingual, Cloudflare Pages):

```bash
go run ./application/cmd/asa docs generate-cli --output docs-site/content/docs/en/cli/generated
cd docs-site && pnpm install && pnpm run docs:check
```

---

## Status

- V1 primitives, V2 intent layer, V3 cost/perf: implemented
- Progressive command surface: implemented (`asa tools`)
- Claude Code first-class adapter: implemented (`agent/claudecode`)
- **Experimental:** MCP server, Notion sync, confidence scoring

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
