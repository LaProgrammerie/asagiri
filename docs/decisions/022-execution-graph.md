# ADR-022 — Execution Graph Planner (local-first)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-my-C.md`](../../spec-my-C.md)

## Context

Asagiri executed tasks sequentially (`spec → tasks → run one after another`). Complex features need dependency-aware ordering, safe parallelism, cost/risk visibility, checkpoints, and trust gates before delivery.

## Decision

1. Add package **`internal/executiongraph/`** with planner, dependency inference, scheduler, estimator, checkpoints, rollback, repository, and runner.
2. Persist plans and run artefacts under **`.asagiri/graphs/<graph-id>/`** (`execution-graph.yaml`, `plan.md`, `metrics.json`, `events.jsonl`, …).
3. Expose CLI: **`asa plan graph`**, **`asa plan explain`**, **`asa graph run|status|resume|visualize`**.
4. Configure defaults via **`execution_graph:`** in `.asagiri/config.yaml`.
5. Integrate with **trust** (gates on high-risk nodes) and **investigation** (auto-insert investigation nodes) without a remote orchestrator.
6. Emit runtime events **`graph.*`** for observability (§19 spec-my-C).
7. **Conservative parallelism** by default (`max_parallel: 2`); never parallelize tasks that touch the same file; CI may force `--max-parallel 1`.

## Consequences

- Execution plans are versionable, inspectable, and exportable (Mermaid/JSON/DOT/markdown).
- Linear `work` / `plan` flows remain available; graph mode targets multi-task features.
- Resume and dry-run are supported locally; no distributed queue in V1.
- Docs and UX must explain plans (`plan explain`, visualize) to avoid false sense of control.

## Related

- Canon: [`docs/ai/06-spec-my-c.md`](../ai/06-spec-my-c.md)
- Prerequisite ADRs: ADR-018 (runtime), ADR-020/021 (trust)
- Log: [`docs/ai/05-decisions.md`](../ai/05-decisions.md)
