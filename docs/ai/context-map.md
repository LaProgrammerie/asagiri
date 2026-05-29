# Context map (sources of truth & tools)

Complements `AGENTS.md`: **where things live**, **who consumes what**, and **how to avoid drift** between Kiro, Cursor, Copilot, and the project canon.

## Sources of truth by topic

| Topic | Where it lives |
|-------|----------------|
| **Phase finale — livrée** | [`spec-phase-finale.md`](../archives/specs/spec-phase-finale.md) §1 (registre PF-* **Fermé**) ; [`handoff.md`](active/handoff.md) archive PF-* 100 % ; [`problems.md`](../../problems.md) GAP-* clôturés |
| **Executable Product Layer (spec-my-A, FULL)** | `spec-my-A.md`, `docs/ai/06-spec-my-a.md`, ADR-018/019, **ADR-025/026** ; `internal/memory/embedder/` ; docs-site `en|fr|de|es` : `cli/runtime`, `runtime-serve`, `concepts/runtime`, `reference/typescript-sdk`, `configuration/config-file` (`runtime.memory.embedder`) |
| **Trust & Verification Engine (spec-my-B, livré)** | `spec-my-B.md`, `docs/ai/06-spec-my-b.md`, `docs/ai/active/handoff.md`, ADR-020/021 ; code `application/internal/trust/`, `.asagiri/trust/<id>/` ; docs-site `en|fr|de|es` : `concepts/trust-engine`, `cli/verify-trust`, `cli/trust-gates`, `cli/trust-replay`, `configuration/config-file` (bloc `verification`) ; CLI généré `en/cli/generated/verify-trust`, `trust`, `trust-gates`, `trust-replay` |
| **Execution Graph Planner (spec-my-C, FULL)** | `spec-my-C.md`, `docs/ai/06-spec-my-c.md`, ADR-022 ; `internal/executiongraph/` (checkpoints, gates, `trust.Engine`, inférence V2 PF-C-06) ; docs-site `en|fr|de|es` : `graph-run`, `graph-resume`, `execution_graph` config |
| **Multi-Agent Coordination (spec-my-D, D-FULL)** | `spec-my-D.md`, `docs/ai/06-spec-my-d.md`, ADR-023 ; `internal/coordination/` (`EnsureWorktree`, `NodeExecutor`, `AssignmentHistory`) ; docs-site `en|fr|de|es` : `concepts/multi-agent-coordination`, `coordination` config |
| **Engineering Knowledge Graph (spec-my-E, livré)** | `spec-my-E.md`, `docs/ai/06-spec-my-e.md`, ADR-024 ; `internal/knowledge/` + extractors (**analytics** : `contracts/analytics.yaml`) ; docs-site `en|fr|de|es` : `engineering-knowledge-graph`, `cli/knowledge`, `cli/impact`, `cli/context` |
| **Replay & Deterministic Execution (spec-my-F, livré)** | `spec-my-F.md`, `docs/ai/06-spec-my-f.md`, `docs/ai/active/handoff.md` ; `internal/replay/` (≠ `internal/trust/replay/`) ; `.asagiri/replays/<replay-id>/` ; docs-site `en|fr|de|es` : `concepts/replay-engine`, `cli/replay`, `configuration/config-file` (bloc `replay`) ; CLI généré `en/cli/generated/replay*` |
| **Experience Platform UI (spec-ui, livré)** | `spec-ui.md`, `docs/ai/06-spec-ui.md`, `docs/ai/active/handoff.md`, ADR-027 ; `application/internal/ui/` + `application/internal/cli/*_ui*.go` ; docs-site `en|fr|de|es` : `experience/index`, `mission-control`, `dashboard`, `command-palette`, `keyboard-shortcuts`, `mouse-support`, `themes`, `accessibility` |
| **Asagiri rebrand (spec-rename)** | `spec-rename.md`, `docs/ai/active/handoff.md`, ADR-016, `docs/migration/github-rename-asagiri.md` |
| **Consolidation OSS / fiabilisation** | [`spec-postv123.md`](archives/specs/spec-postv123.md), `docs/consolidation/` |
| **Public documentation site** | [`spec-doc.md`](archives/specs/spec-doc.md), [`spec-deploy-doc.md`](archives/specs/spec-deploy-doc.md), `docs-site/` (Fumadocs → **Cloudflare Pages**, projet **`asagiri-docs`**) ; contenu `content/docs/{en,fr,de,es}/` — **en** défaut/fallback, référence CLI générée sous `en/cli/generated/` ; CI `.github/workflows/docs-cloudflare-pages.yml` |
| **Doc / code / spec drift tracker** | `problems.md` (repo root) |
| **Archives specs longues** | [`docs/ai/archives/specs/`](archives/specs/README.md) — index de toutes les `spec*.md` |
| **Asagiri cost/perf (V3, livré)** | [`specv3.md`](archives/specs/specv3.md), `docs/ai/06-spec-v3.md`, ADR-010 ; `internal/cost/`, `contextopt/`, `pipeline/` ; docs-site `cost-performance/`, `concepts/cost-aware-workflows` |
| **Intent layer** | [`specv2.md`](archives/specs/specv2.md) |
| **V1 spec (historique, noms AgentFlow)** | [`spec.md`](archives/specs/spec.md) |
| Short cross-tool index | `AGENTS.md` (root) |
| Stack locale (Go, Docker, Makefile) | `docs/ai/02-architecture.md`, `docs/ai/03-standards.md`, `Makefile` |
| Decisions / architecture / standards (detail) | `docs/ai/*.md` |
| Native Kiro spec workflow artefacts | `.kiro/specs/<feature>/` (`requirements.md`, `design.md`, `tasks.md`, …) |
| Spec summary **outside Kiro** | `docs/ai/active/current-spec.md` |
| **Execution** contract (Cursor, Copilot, human) | `docs/ai/active/handoff.md` |
| Kiro projection + targeted rules | `.kiro/steering/` |
| **Repo-specific** workflows (e.g. handoff) | `.kiro/skills/` **in this repo** |
| **Universal** workflows (review, release, debug, …) | `~/.kiro/skills/` |
| Short static **Cursor** rules | `.cursor/rules/*.mdc` |
| This repo as **upstream template** for forks | `.cursor/rules/template-is-upstream.mdc`, `.kiro/steering/35-template-downstream.md` |
| Template drift sync hooks/rules for downstream repos | `.cursor/hooks.json`, `.cursor/hooks/template-sync-*.sh`, `.cursor/rules/template-generic-sync.mdc`, `.kiro/steering/35-template-generic-sync.md` |
| GitHub bridge | `.github/copilot-instructions.md` |
| Phase 2 repo + module Go path | `docs/migration/github-rename-asagiri.md` |

## Recommended reading order

1. `AGENTS.md`
2. `docs/ai/context-map.md` (this file)
3. `docs/ai/00-overview.md` … `05-decisions.md` as needed
4. **Implement:** `docs/ai/active/handoff.md` (prescriptive contract) **then** `docs/ai/03-standards.md` and useful sections of `02-architecture.md`

## `.kiro/specs/*` vs `docs/ai/active/*`

| Location | Role |
|----------|------|
| `.kiro/specs/...` | **Native Kiro artefacts**: operational truth for specifying, splitting, and tracking tasks in the tool. |
| `docs/ai/active/current-spec.md` | **Cross-tool summary / projection**: what another agent must know without opening the whole spec folder. **Refresh** when the Kiro spec changes materially. |
| `docs/ai/active/handoff.md` | **Prescriptive execution contract**: allowed scope, allowed/denied files, plan, tests, DoD; derived from Kiro tasks / design. |

**Anti-drift rule:** if you change requirements/design/tasks under `.kiro/specs/`, update **at least** `current-spec.md` and, if implementation scope moves, `handoff.md`. Do not let the three diverge.

## Responsibilities: Kiro vs Cursor vs Copilot

| Tool | Primary role |
|------|----------------|
| **Kiro** | Specify, structure, produce artefacts under `.kiro/specs/`; project steering in `.kiro/steering/`. |
| **Cursor** | Implement and iterate; use `handoff.md` + `docs/ai` + short rules; rich context **on demand** (files, skills, tools). |
| **Copilot** | Adapter in the GitHub ecosystem: same sources of truth via `copilot-instructions.md` + repo files. |

## Update conventions

- **Durable product / architecture change** → `docs/ai/` (+ `05-decisions.md` if a structural tradeoff).
- **Active spec change** → `.kiro/specs/...` then `active/current-spec.md` / `active/handoff.md`.
- **Editor policy or tool-specific convention** → `.cursor/rules/` or `.kiro/steering/` depending on whether it is mainly Cursor or Kiro.
- **Reusable workflow across all repos** → `~/.kiro/skills/`, not this project template.
