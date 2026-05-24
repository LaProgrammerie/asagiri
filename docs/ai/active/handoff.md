# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `agentflow-specv2` : livrée** (`2026-05-17`).

## Immediate objective

Couche intention AgentFlow (specv2) : `work`, `continue`, `next`, `inbox`, `sync` ; resolver + planner ; sources Local + Notion ; sans régression des primitives V1.

## Allowed scope (agentflow-specv2)

- `application/internal/intent/`
- `application/internal/source/` (+ `notion/`)
- `application/internal/cli/` (work, continue, next, inbox, sync)
- `application/internal/config/`
- `.agentflow/config.yaml.example`
- `docs/ai/` (handoff, current-spec, 02, 03, 04, 05, context-map)
- `README.md`

## Definition of Done — agentflow-specv2

- [x] Commandes §4 : `work`, `continue`, `next`, `inbox`, `sync` + options spec
- [x] `IntentResolver` hybride (déterministe, fuzzy, Ollama fallback, ambiguïté CI)
- [x] `HighLevelPlanner` + conditions §6.2
- [x] Sources : `LocalSource`, `NotionSource` (httptest + intégration opt-in `NOTION_TOKEN`)
- [x] Config `intent` / `work` / `sources`
- [x] Modes guided / auto (`--yes`) / plan-only
- [x] UX terminal boxed §12 ; `help.go` mis à jour
- [x] Confirmations sync overwrite ; erreur structurée si ambigu non interactif
- [x] Dette PHP `application/` supprimée (tests E2E, src)
- [x] `go test -race ./...` vert
- [ ] `golangci-lint` (selon toolchain locale)

## Hors scope

- Embeddings vectoriels RAG
- Reprise `resume` automatique hors dry-run (inchangé V1)
- GitHub Issues / Linear comme sources

## References

- [`specv2.md`](../../../specv2.md) — spec produit évolution
- [`spec.md`](../../../spec.md) — historique V1
- [`current-spec.md`](current-spec.md)
- ADR-009 dans [`05-decisions.md`](../05-decisions.md)
