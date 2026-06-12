# Architecture decisions (public)

Public ADR excerpts for contributors. The canonical log lives in [`docs/ai/05-decisions.md`](../ai/05-decisions.md) (agent/engineering canon).

| ID | Summary |
|----|---------|
| ADR-001 | Go stack, `application/` layout, Makefile orchestration |
| ADR-009 | Intent layer; Notion never executed without local snapshot |
| ADR-010 | V3 cost/perf; pricing from config only; MCP off by default |
| ADR-011 | Consolidation post-V3; Apache 2.0; `redact` package |
| ADR-012 | Public docs: Fumadocs static site under `docs-site/` |
| ADR-014 | Docs hosting: Cloudflare Pages (Wrangler CI), pnpm, `docs-check` for forks |
| ADR-022 | Execution Graph Planner local-first (`.asagiri/graphs/`, conservative parallelism) |
| ADR-023 | Multi-agent coordination foundation (`internal/coordination/`, `coordination:`, `agent.*` events) |
| ADR-024 | Engineering knowledge graph foundation (`internal/knowledge/`, `.asagiri/knowledge/graph.sqlite`) |
| ADR-025 | Pluggable memory embeddings (`hash` / Ollama / opt-in cloud) — PF-A-01 |
| ADR-026 | npm TypeScript SDK (`@laprogrammerie/asagiri`, tag `sdk-v*`) — PF-A-02 |
| ADR-033 | Trust & Diagnostic Architecture — worktrust, daily UX, reportsink `--save`, Trust Gate future |
| ADR-034 | Report history & diff — `history/` snapshots, `reportdiff`, `asa doctor|trust diff` |
| ADR-037 | Monetization & Distribution V1 — OSS / Pro Local / Team Cloud / Enterprise ; monétisation additive |
| ADR-038 | Pro Local packs — `asagiri-pack-v1`, `asagiri-packs` CLI séparé |
| ADR-039 | Team Cloud — control plane metadata, sync opt-in, pas de proxy LLM |
| ADR-040 | Enterprise — SSO, RBAC, audit WORM, on-prem, policies client-side |
| ADR-023 (D-FULL) | Coordination worktrees + `NodeExecutor` + assignment history scoring |

Full table: [`docs/ai/05-decisions.md`](../ai/05-decisions.md).
