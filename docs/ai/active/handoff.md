# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-doc-v2` : passe éditoriale documentation** (en/fr/de/es) + consolidation (`2026-05-17`).

## Immediate objective

**Passe éditoriale `spec-doc-v2`** : réécriture prose (en/fr/de/es) de `docs-site/content/docs/**` hors `cli/generated/**` — ton ingénierie, WHY+HOW, ~70 % prose ; fond technique inchangé. Build `pnpm build` obligatoire. Pas de commit automatique.

Parallèle / hérité : site Fumadocs, CLI générée, Cloudflare Pages (`spec-postv123`, `spec-doc`).

## Allowed scope (spec-postv123)

- `docs-site/` (Fumadocs, contenu EN, static export)
- `.github/workflows/docs-cloudflare-pages.yml`, `.github/workflows/docs-check.yml`
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
- [x] `docs-site` build static (`out/`) + workflow Cloudflare Pages (`docs-cloudflare-pages.yml`) + validation PR sans secrets (`docs-check.yml`)
- [x] MVP pages EN + CLI ref générée (`docs generate-cli`)
- [x] `TestCLICommandsDocumented` (docgen)
- [x] **Tranche doc-content** : pages MVP `docs-site/content/docs/*` sans placeholder ; pages config/mcp/reference ; `problems.md` ; regen CLI

## Definition of Done — spec-doc-v2 (éditorial)

- [x] ~144 MDX (en/fr/de/es) réécrits hors `cli/generated/**`
- [x] Pages prioritaires : index, getting-started, concepts, architecture, cost-performance, workflows, reliability, security, `cli/index.mdx`
- [x] Cohérence narrative entre langues ; aucune feature inventée
- [x] `cli/generated/*.mdx` non modifiés manuellement (stubs fr/de/es inchangés)
- [x] `cd docs-site && pnpm typecheck && pnpm lint && pnpm build` + `out/` présent

## Hors scope

- Tokenizers provider-exacts
- CI GitHub release complète
- Bubbletea interactif
- Renommage module Go

## References

- [`spec-deploy-doc.md`](../../../spec-deploy-doc.md) (Cloudflare Pages CI)
- [`spec-postv123.md`](../../../spec-postv123.md)
- [`docs/consolidation/README.md`](../../consolidation/README.md)
- [`specv3.md`](../../../specv3.md), [`specv2.md`](../../../specv2.md), [`spec.md`](../../../spec.md)
- ADR-011 dans [`05-decisions.md`](../05-decisions.md)
