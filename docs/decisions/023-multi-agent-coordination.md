# ADR-023 — Multi-Agent Coordination (spec-my-D foundation)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-my-D.md`](../../spec-my-D.md)

## Context

The execution graph (ADR-022) assigns tasks and parallelism but does not govern specialized agent roles, isolation, structured handoffs, cross-validation, or coordination policies. Running multiple LLM agents without explicit roles increases cost, conflicts, and review gaps.

## Decision

1. Add package **`internal/coordination/`** with coordinator, assignment, handoffs, policies, runtime emission, and Lot-4 stubs (conflict, budget, escalation).
2. Define explicit **agent roles** and **isolation modes** (`shared`, `isolated_worktree`, `readonly`, `sandbox`); default isolation **`isolated_worktree`**, **`max_parallel_agents: 2`**.
3. **`DefaultAssigner`** delegates agent ref selection to **`executiongraph.DefaultAgentFor`**, enriched with `coordination.assignment` overrides and `coordination.profiles` (profiles reference existing `agents:` keys).
4. Persist handoffs under **`.asagiri/handoffs/<id>/handoff.yaml`** via `DefaultHandoffBuilder`.
5. Configure via **`coordination:`** in `.asagiri/config.yaml`; validate `profiles.*.agent` ∈ `agents` keys.
6. Emit runtime events **`agent.*`** (§10) from `internal/runtime/agent_events.go` and `coordination.CoordinationEmitter`.
7. **Lot 1 scope only:** no runner integration, no CLI, no docs-site; conflict/budget/escalation remain stubs until Lots 4–5.

## Consequences

- Execution graph nodes gain coordinated role/isolation metadata without replacing the graph planner.
- Cross-validation and merge policies are configured early but enforced in later lots.
- Runner and CLI must call `AgentCoordinator` in Lot 5 without duplicating assignment logic.
- Docs and UX (§19) follow in Lot 6.

## Related

- Prerequisite: ADR-022 (execution graph)
- Handoff: [`docs/ai/active/handoff.md`](../ai/active/handoff.md) — spec-my-D lots 1–6
