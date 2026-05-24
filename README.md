# AgentFlow

Deterministic orchestration for AI coding workflows.

AgentFlow turns specs into auditable, cost-aware development runs using local investigation, git worktrees, external coding agents, validation commands, and reproducible reports.

## Why

- **Local-first** — investigate and optimize context before calling cloud models
- **Cost-aware** — estimate tokens and spend before execution
- **Isolated** — one git worktree per task
- **Observable** — SQLite state, reports, and explicit validation steps

## Install

Requirements: Go 1.25+, `git`, `make`.

```bash
git clone https://github.com/LaProgrammerie/hyper-fast-builder.git
cd hyper-fast-builder
go mod download
make build
```

## Quickstart

```bash
export AGENTFLOW_DRY_RUN=1   # optional: no external agent binaries
./bin/agentflow init
./bin/agentflow doctor
./bin/agentflow work "develop my-feature" --dry-run --plan-only --yes
```

See [`examples/quickstart/`](examples/quickstart/) for a longer walkthrough.

## Core workflow

```bash
./bin/agentflow work "develop billing-v2" --estimate-only
./bin/agentflow work "develop billing-v2"
./bin/agentflow status
./bin/agentflow report <run-id>
```

Primitive pipeline (Kiro → plan → dev → verify): [`spec.md`](spec.md) · Intent layer: [`specv2.md`](specv2.md) · Cost/perf: [`specv3.md`](specv3.md)

## Documentation

Full docs (English): **https://laprogrammerie.github.io/hyper-fast-builder/docs/**

Build locally:

```bash
go run ./application/cmd/agentflow docs generate-cli --output docs-site/content/docs/cli/generated
cd docs-site && npm install && npm run build
# static output: docs-site/out/
```

## Status

- V1 primitives, V2 intent layer, V3 cost/perf: implemented
- Consolidation & OSS readiness: in progress ([`spec-postv123.md`](spec-postv123.md))
- **Experimental:** MCP server, Notion sync, confidence scoring — see docs

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) and the [contributing guide](https://laprogrammerie.github.io/hyper-fast-builder/docs/contributing/).

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
