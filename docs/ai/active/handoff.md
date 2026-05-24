# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-postv123` : consolidation & OSS readiness** (`2026-05-17`).

## Immediate objective

Consolidation technique AgentFlow : audits, garde-fous, tests packages critiques, explainability CLI, préparation open source — **sans nouvelles features marketing**.

## Allowed scope (spec-postv123)

- `application/internal/*` (refactors ciblés : pipeline, contextopt, config, cost, cli, redact, routing)
- `docs/consolidation/`
- `LICENSE`, `CONTRIBUTING.md`, `ROADMAP.md`, `examples/`, `scripts/benchmark-workflow.sh`
- `Makefile` (cible `benchmark`)
- `README.md`
- `docs/ai/` (handoff, current-spec, context-map, 05-decisions)

## Definition of Done — spec-postv123

- [x] Audits §1–§4, §8–§9 documentés sous `docs/consolidation/`
- [x] Écart critique double-scan corrigé (`CollectForPipeline`)
- [x] Validations MCP si `enabled`
- [x] Explainability estimate/work (steps + résumé)
- [x] Tests : pipeline, routing, executor, redact, golden estimate shape
- [x] `go test -race ./...` vert
- [x] LICENSE Apache 2.0, CONTRIBUTING, ROADMAP, examples/quickstart
- [x] `make benchmark` / script dry-run
- [ ] Couverture `workflow` / `intent` > 50 % (roadmap §6)

## Hors scope

- Tokenizers provider-exacts
- CI GitHub release complète
- Bubbletea interactif
- Renommage module Go

## References

- [`spec-postv123.md`](../../../spec-postv123.md)
- [`docs/consolidation/README.md`](../../consolidation/README.md)
- [`specv3.md`](../../../specv3.md), [`specv2.md`](../../../specv2.md), [`spec.md`](../../../spec.md)
- ADR-011 dans [`05-decisions.md`](../05-decisions.md)
