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

### Providers vs Agents

Asagiri separates **external runtimes** from **logical profiles**:

| Concept | Config block | Meaning |
| --- | --- | --- |
| **Provider** | `providers:` | External runtime (Kiro CLI, Claude Code, Ollama, …). Adapter selection uses **`provider.type` only** — never the map key. |
| **Agent** | `agents:` | Logical profile referenced by `work.default_*`, `--agent`, and coordination. May point at a provider via `agents.<id>.provider`. |
| **Work role** | `work:` | Which logical agent runs spec, dev, review, enrich by default. |
| **Workflow** | CLI commands | `asa spec`, `asa dev`, `asa review`, … orchestrate the pipeline. |

```
Provider  →  Agent  →  Work role  →  Workflow
kiro-cli  →  dev     →  default_agent  →  asa dev
```

Several logical agents can share one provider:

```yaml
providers:
  kiro-cli:
    type: kiro-cli
    command: kiro
    args: ["--cli"]

agents:
  dev:
    provider: kiro-cli
  architect:
    provider: kiro-cli
  reviewer:
    provider: claude-code

work:
  default_agent: dev
  default_reviewer: reviewer
```

Legacy inline agents remain supported (no `provider` field → implicit `exec` adapter):

```yaml
agents:
  claude:
    command: claude
```

See `.asagiri/config.yaml.example` and [Agents configuration](/docs/configuration/agents) on the docs site.

### `asa work` — three scope paths (V1)

Every instruction is classified deterministically (no LLM). Priority: **technical_task** > **feature_work** > **product_level_intent**. When uncertain, Asagiri stays on the technical path and does **not** trigger the Product Layer.

| Scope | Example | What happens |
| --- | --- | --- |
| `technical_task` | `corrige le bug login` | Normal technical workflow (plan → dev → verify …) |
| `feature_work` | `ajoute export CSV` | Normal technical workflow |
| `product_level_intent` | `Créer un CRM pour artisans` | Product Layer preparation, then **controlled stop** (no auto-chain into dev in the same run) |

### Product intentions — two-step workflow (V1)

Broad product intents use a **deliberate two-invocation** flow. Asagiri does **not** auto-chain preparation and implementation in one command.

**Step 1 — prepare product artifacts**

```bash
asa work "Créer un CRM pour artisans" --dry-run   # preview only
asa work "Créer un CRM pour artisans" --yes       # writes product model → prototype → flows → contracts → specs → tasks
```

Product slugs are derived heuristically from the intent (e.g. `crm-artisans`); an existing slug on disk is preserved. `--plan-only` on a product-level intent shows only the Product Layer plan (not the technical workflow plan).

**Step 2 — continue implementation** (separate invocation)

```bash
asa work crm-artisans --yes
asa next --feature crm-artisans
```

Expert commands (`asa prototype`, `asa flows`, `asa contracts`, `asa spec generate-from-product`) remain available but are optional on the happy path.

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

## Local-First AI Orchestration

Asagiri routes tasks to the cheapest capable model and escalates to premium only when necessary.

### Cost intelligence

```bash
asa cost report          # actual spend + local/cloud ratio + savings (if baseline configured)
asa cost trends          # evolution over two 15-day windows
```

**Without baseline:**
```
Local / cheap:   78% (82,000 tokens) — no LLM cost
Cloud / premium: 22% (23,000 tokens)
Strategy score:  A — strong local-first routing
```

**With `pricing.premium_reference_model: "gpt-4o"` in config.yaml:**
```
Savings (vs gpt-4o)
Premium equivalent: €4.90
Savings:            €4.48
Savings rate:       91.4%
```

### Agent Strategy Score

| Grade | Local token % | Meaning |
|-------|--------------|---------|
| A | ≥ 70% | Strong local-first |
| B | 50–70% | Balanced, room to improve |
| C | 30–50% | Cloud-heavy |
| D | < 30% | Almost all premium |

Score is computed from `step_metrics.local` — no heuristics, no ML.

### Escalation metrics

```
Steps total:         42
Local (no cost):     39
Premium escalations: 3   (7%)
```

"Premium escalation" = any step that ran on a cloud model. Asagiri aims to minimize this.

### Configuration

```yaml
pricing:
  premium_reference_model: "gpt-4o"  # optional — enables savings computation
  models:
    gpt-4o:
      input_per_1m_tokens: 5.0
      output_per_1m_tokens: 15.0
```

Without `premium_reference_model`, only local/cloud ratio and strategy score are shown.
No baseline is invented by default.

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
