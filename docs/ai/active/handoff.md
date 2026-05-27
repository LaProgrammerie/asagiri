# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche :** `spec-my-A` **livrée** (`2026-05-27`) — audit doc `2026-05-27`.

## Objectif

Maintenir `spec-my-A.md` (§1–26) et la documentation alignée (canon `docs/ai/`, site `docs-site/` EN/FR/DE/ES).

**Suite hors ce handoff :** [`spec-phase-finale.md`](../../../spec-phase-finale.md).

---

## Périmètre autorisé (maintenance)

- `application/internal/product/**`, `application/internal/product/derivation/**`
- `application/internal/runtime/**`, `application/internal/runtime/api/**`
- `application/internal/memory/**`, `application/internal/skills/**`, `application/internal/embedutil/**`
- `application/internal/analysis/**`, `application/internal/investigation/**`
- `application/pkg/asagiri/**`, `sdk/typescript/**`
- `application/internal/cli/**`, `application/internal/cli/docgen/**`
- `docs/ai/**`, `docs-site/content/docs/**`
- `.asagiri/config.yaml.example`, fixtures `.asagiri/products/**`
- `spec-my-A.md`, `spec-phase-finale.md`

---

## Definition of Done — spec-my-A

### Code

- [x] Bloc A — Product layer (§19)
- [x] Bloc B — Business intent (§23)
- [x] Bloc C — Runtime (§24, modes, API, metrics, rich status, memory, skills, hooks)
- [x] Bloc C2 — Analysis (§24.16, `asa analysis build`)
- [x] Bloc D — Investigation (§25, context-pack.md, impact, work/verify hooks)
- [x] `go test ./...` + `sdk/typescript` npm test

### Canon `docs/ai/`

- [x] [`06-spec-my-a.md`](../06-spec-my-a.md)
- [x] [`02-architecture.md`](../02-architecture.md) — packages spec-my-A
- [x] [`active/current-spec.md`](current-spec.md)
- [x] [`context-map.md`](../context-map.md)
- [x] ADR-018, ADR-019 dans [`05-decisions.md`](../05-decisions.md)

### Site docs (`docs-site`)

- [x] EN — pages dédiées (pas de stubs)
- [x] FR — traductions complètes
- [x] DE — traductions complètes
- [x] ES — traductions complètes
- [x] `configuration/config-file` — bloc `runtime` (EN/FR/DE/ES)
- [x] CLI `generated/` — `make build && ./bin/asa docs generate-cli` (`2026-05-27`)

---

## Hors scope

- `spec-phase-finale` (embeddings Ollama, publish npm) sauf mise à jour doc seule
- Commit / push par l'agent
- Investigation cloud par défaut

---

## Audit handoff (2026-05-27)

| Vérification | Résultat |
|--------------|----------|
| Commandes §22 spec-my-A | Couvertes par tests integration + CLI |
| Critères §24.21 runtime | daemon, sessions, hooks, memory, skills, graphs |
| Critères §25.24 investigation | rapport, scope, graph, investigate-first, verify-on-failure |
| Doc EN/FR/DE/ES spec-my-A | Pages listées dans `06-spec-my-a.md` §5 |
| `current-spec` vs handoff | Alignés |
| Écarts documentés | Uniquement `spec-phase-finale` |

---

## Validation

```bash
cd application && go test ./...
cd sdk/typescript && npm test
asa analysis build --product workspace-saas
asa daemon status --rich
asa runtime serve --port 8765
asa investigate impact --flow onboarding --change "async invitations"
```

## Références

- [`spec-my-A.md`](../../../spec-my-A.md)
- [`06-spec-my-a.md`](../06-spec-my-a.md)
- ADR-018, ADR-019
