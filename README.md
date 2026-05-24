# Asagiri

Deterministic orchestration for AI coding workflows.

Asagiri turns specs into auditable, cost-aware development runs using local investigation, git worktrees, external coding agents, validation commands, and reproducible reports.

## Why

- **Local-first** — investigate and optimize context before calling cloud models
- **Cost-aware** — estimate tokens and spend before execution
- **Isolated** — one git worktree per task
- **Observable** — SQLite state, reports, and explicit validation steps

## Install

### Homebrew (macOS / Linux)

```bash
brew tap LaProgrammerie/tap
brew install asa
asa version
```

Requires the [homebrew-tap](https://github.com/LaProgrammerie/homebrew-tap) repository and a published release.

### Manual install (released binary)

Download an archive from [GitHub Releases](https://github.com/LaProgrammerie/asagiri/releases) for your OS/arch (`asa_Darwin_arm64.tar.gz`, `asa_Linux_x86_64.tar.gz`, `asa_Windows_arm64.zip`, etc.).

```bash
VERSION=v0.1.0
curl -LO "https://github.com/LaProgrammerie/asagiri/releases/download/${VERSION}/checksums.txt"
curl -LO "https://github.com/LaProgrammerie/asagiri/releases/download/${VERSION}/asa_Darwin_arm64.tar.gz"
sha256sum -c checksums.txt
tar -xzf asa_Darwin_arm64.tar.gz
sudo install -m 755 asa /usr/local/bin/asa
asa version
```

Adjust `VERSION`, archive name, and install path for your platform.

### From source

Requirements: Go 1.25+, `git`, `make`.

```bash
git clone https://github.com/LaProgrammerie/asagiri.git
cd asagiri
go mod download
make build
./bin/asa version
```

### Verify

```bash
asa doctor
```

See [`docs/release-process.md`](docs/release-process.md) for maintainers.

## Quickstart

```bash
export ASA_DRY_RUN=1   # optional: no external agent binaries
./bin/asa init
./bin/asa doctor
./bin/asa work "develop my-feature" --dry-run --plan-only --yes
```

See [`examples/quickstart/`](examples/quickstart/) for a longer walkthrough.

## Core workflow

```bash
./bin/asa work "develop billing-v2" --estimate-only
./bin/asa work "develop billing-v2"
./bin/asa status
./bin/asa report <run-id>
```

Primitive pipeline (Kiro → plan → dev → verify): [`spec.md`](spec.md) · Intent layer: [`specv2.md`](specv2.md) · Cost/perf: [`specv3.md`](specv3.md)

## Documentation

Public docs (Fumadocs, static export on **Cloudflare Pages**) — **en** (default), **fr**, **de**, **es**. Production URL depends on your Pages project / custom domain (configure after first deploy).

Build locally:

```bash
go run ./application/cmd/asa docs generate-cli --output docs-site/content/docs/en/cli/generated
cd docs-site && corepack enable && pnpm install && pnpm run docs:check
# static output: docs-site/out/
```

**CI deploy** (push to `main` or PR preview): [`.github/workflows/docs-cloudflare-pages.yml`](.github/workflows/docs-cloudflare-pages.yml). Repository secrets (no values in repo): `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_PAGES_PROJECT`. See [`docs-site/README.md`](docs-site/README.md).

## Status

- V1 primitives, V2 intent layer, V3 cost/perf: implemented
- Consolidation & OSS readiness: in progress ([`spec-postv123.md`](spec-postv123.md))
- **Experimental:** MCP server, Notion sync, confidence scoring — see docs

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) and the [contributing guide](https://laprogrammerie.github.io/asagiri/docs/contributing/).

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
