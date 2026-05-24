# AgentFlow

Deterministic orchestration for AI coding workflows.

AgentFlow turns specs into auditable, cost-aware development runs using local investigation, git worktrees, external coding agents, validation commands, and reproducible reports.

## Why

- **Local-first** — investigate and optimize context before calling cloud models
- **Cost-aware** — estimate tokens and spend before execution
- **Isolated** — one git worktree per task
- **Observable** — SQLite state, reports, and explicit validation steps

## Install

### Homebrew (macOS / Linux)

```bash
brew tap LaProgrammerie/tap
brew install agentflow
agentflow version
```

Requires the [homebrew-tap](https://github.com/LaProgrammerie/homebrew-tap) repository and a published release.

### Manual install (released binary)

Download an archive from [GitHub Releases](https://github.com/LaProgrammerie/hyper-fast-builder/releases) for your OS/arch (`agentflow_Darwin_arm64.tar.gz`, `agentflow_Linux_x86_64.tar.gz`, `agentflow_Windows_arm64.zip`, etc.).

```bash
VERSION=v0.1.0
curl -LO "https://github.com/LaProgrammerie/hyper-fast-builder/releases/download/${VERSION}/checksums.txt"
curl -LO "https://github.com/LaProgrammerie/hyper-fast-builder/releases/download/${VERSION}/agentflow_Darwin_arm64.tar.gz"
sha256sum -c checksums.txt
tar -xzf agentflow_Darwin_arm64.tar.gz
sudo install -m 755 agentflow /usr/local/bin/agentflow
agentflow version
```

Adjust `VERSION`, archive name, and install path for your platform.

### From source

Requirements: Go 1.25+, `git`, `make`.

```bash
git clone https://github.com/LaProgrammerie/hyper-fast-builder.git
cd hyper-fast-builder
go mod download
make build
./bin/agentflow version
```

### Verify

```bash
agentflow doctor
```

See [`docs/release-process.md`](docs/release-process.md) for maintainers.

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

Public docs (Fumadocs, static export on **Cloudflare Pages**) — **en** (default), **fr**, **de**, **es**. Production URL depends on your Pages project / custom domain (configure after first deploy).

Build locally:

```bash
go run ./application/cmd/agentflow docs generate-cli --output docs-site/content/docs/en/cli/generated
cd docs-site && corepack enable && pnpm install && pnpm run docs:check
# static output: docs-site/out/
```

**CI deploy** (push to `main` or PR preview): [`.github/workflows/docs-cloudflare-pages.yml`](.github/workflows/docs-cloudflare-pages.yml). Repository secrets (no values in repo): `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_PAGES_PROJECT`. See [`docs-site/README.md`](docs-site/README.md).

## Status

- V1 primitives, V2 intent layer, V3 cost/perf: implemented
- Consolidation & OSS readiness: in progress ([`spec-postv123.md`](spec-postv123.md))
- **Experimental:** MCP server, Notion sync, confidence scoring — see docs

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) and the [contributing guide](https://laprogrammerie.github.io/hyper-fast-builder/docs/contributing/).

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
