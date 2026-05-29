# Context map (sources of truth & tools)

Complements `AGENTS.md`: **where things live**, **who consumes what**, and **how to avoid drift** between Kiro, Cursor, Copilot, and the project canon.

## Sources of truth by topic

| Topic | Where it lives |
|-------|----------------|
| **Phase finale — reliquats A/B/C + transverse** | [`spec-phase-finale.md`](../../spec-phase-finale.md) §1 (registre **PF-***) ; tracker historique [`problems.md`](../../problems.md) (GAP-* ↔ PF-X-*) ; handoff dérivé quand phase active |
| **Executable Product Layer (spec-my-A, livré)** | `spec-my-A.md`, `docs/ai/06-spec-my-a.md`, `docs/ai/active/handoff.md`, ADR-018/019 ; docs-site `en|fr|de|es` : `cli/runtime`, `runtime-serve`, `investigate`, `analysis`, `concepts/runtime`, `investigation`, `analysis-layer`, `reference/typescript-sdk`, `configuration/config-file` (bloc `runtime`) ; reliquats **PF-A-*** → `spec-phase-finale.md` |
| **Trust & Verification Engine (spec-my-B, livré)** | `spec-my-B.md`, `docs/ai/06-spec-my-b.md`, `docs/ai/active/handoff.md`, ADR-020/021 ; code `application/internal/trust/`, `.asagiri/trust/<id>/` ; docs-site `en|fr|de|es` : `concepts/trust-engine`, `cli/verify-trust`, `cli/trust-gates`, `cli/trust-replay`, `configuration/config-file` (bloc `verification`) ; CLI généré `en/cli/generated/verify-trust`, `trust`, `trust-gates`, `trust-replay` |
| **Execution Graph Planner (spec-my-C, livré)** | `spec-my-C.md`, `docs/ai/06-spec-my-c.md`, `docs/ai/active/handoff.md`, ADR-022 ; code `application/internal/executiongraph/`, `.asagiri/graphs/<id>/` ; docs-site `en|fr|de|es` : `concepts/execution-graph`, `cli/plan-graph`, `plan-explain`, `graph-run`, `graph-status`, `graph-resume`, `graph-visualize`, `configuration/config-file` (bloc `execution_graph`) ; CLI généré `en/cli/generated/plan`, `graph` ; durcissements **PF-C-*** → `spec-phase-finale.md` |
| **Asagiri rebrand (spec-rename)** | `spec-rename.md`, `docs/ai/active/handoff.md`, ADR-016, `docs/migration/github-rename-asagiri.md` |
| **Consolidation OSS / fiabilisation** | `spec-consolidation-now.md` (actif), `spec-postv123.md` (historique), `docs/consolidation/` |
| **Public documentation site** | `spec-doc.md`, `spec-deploy-doc.md`, `docs-site/` (Fumadocs → **Cloudflare Pages**, projet **`asagiri-docs`**) ; contenu `content/docs/{en,fr,de,es}/` — **en** défaut/fallback, référence CLI générée sous `en/cli/generated/` ; CI `.github/workflows/docs-cloudflare-pages.yml` |
| **Doc / code / spec drift tracker** | `problems.md` (repo root) |
| **Asagiri cost/perf (V3)** | `specv3.md` (repo root) |
| **Intent layer** | `specv2.md` (repo root) |
| **V1 spec (historique, noms AgentFlow)** | `spec.md` (repo root) |
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
