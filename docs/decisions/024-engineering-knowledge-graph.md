# ADR-024 — Engineering Knowledge Graph (spec-my-E)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-my-E.md`](../ai/archives/specs/spec-my-E.md)

## Context

Agents and planners reason over isolated files. Product flows, API contracts, code symbols, and tests are not linked in a queryable local structure. Spec-my-E introduces an engineering knowledge graph for investigation, trust, execution planning, and coordination.

## Decision

1. Add **`application/internal/knowledge/`** with graph models, SQLite store, extractors, query/BFS, impact analysis, staleness detection, snapshots, and terminal UX.
2. Persist under **`.asagiri/knowledge/graph.sqlite`** and export **`graph.json`**; named snapshots under **`snapshots/<name>/`**.
3. **Hybrid IDs** `type:stable_key`; provenance and confidence on every node/edge.
4. **Incremental build** via per-category `source_mtimes` in `index_metadata.build`; skip unchanged extractors.
5. **CLI** : `asa knowledge build|query|explain|snapshot`, `asa impact analyze` ; `--json` on all knowledge/impact commands.
6. **Bridges** (read-only) in investigation, trust, executiongraph, coordination ; `asa context build --from-graph`.
7. **Config** : optional `knowledge:` block in `.asagiri/config.yaml` (defaults documented on site).

## Consequences

- Local-first graph; no cloud KB in V1.
- Extractors are heuristic (Go/PHP/TS simple scans); perfect multi-language AST is out of scope.
- Docs-site concept + CLI pages in EN/FR/DE/ES; regenerate English CLI reference with `asa docs generate-cli`.

## Related

- Prerequisites: ADR-022 (execution graph), ADR-023 (multi-agent coordination)
- Canon: [`docs/ai/06-spec-my-e.md`](../ai/06-spec-my-e.md)
- Handoff: [`docs/ai/active/handoff.md`](../ai/active/handoff.md) — spec-my-E livrée `2026-05-29`
