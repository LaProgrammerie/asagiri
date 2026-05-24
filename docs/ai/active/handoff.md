# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-doc` : documentation publique** + consolidation (`2026-05-17`).

## Immediate objective

Site docs public Fumadocs (`docs-site/`), référence CLI générée, CI GitHub Pages — en parallèle de la consolidation OSS (`spec-postv123`).

## Allowed scope (spec-postv123)

- `docs-site/` (Fumadocs, contenu EN, static export)
- `.github/workflows/docs.yml`
- `application/internal/cli/docgen/` + `docs` subcommand
- `docs/decisions/`, `docs/contributing/`, `docs/specs/`
- `CODE_OF_CONDUCT.md`, `SECURITY.md`
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
- [x] `docs-site` build static (`out/`) + workflow Pages
- [x] MVP pages EN + CLI ref générée (`docs generate-cli`)
- [x] `TestCLICommandsDocumented` (docgen)

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
