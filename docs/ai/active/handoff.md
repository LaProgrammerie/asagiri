# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.
> **Statut :** **task-validation-gates — Tranche A livrée** (`2026-06-06`, ADR-031).
> **Prochaine étape :** définir une nouvelle spec / handoff avant tout travail hors maintenance Tranche A.

## Livraison Tranche A (clôturée)

Gate de gouvernance read-only post-dev (`per-task`), opt-in via `work.governance.enabled`,
verdict `PASS/WARN/FAIL`, retry dev sur FAIL (option A), trace payload/logs/report.

Références : `.kiro/specs/task-validation-gates/`, ADR-031, `docs/ai/02-architecture.md` (flux V1).

## Sémantique retry (invariant — ne pas régresser)

- **`max_retries`** : relances autorisées **après le premier FAIL** governance.
- **`governance.retries`** : relances **consommées** (incrément si relance accordée).
- **`history[].retry`** : tentative governance (0 = première évaluation).
- **Condition** : `retries_used < max_retries` → relance ; sinon `failed`.
- **Passages max** : `max_retries + 1`.

## Definition of Done — Tranche A (toutes cochées)

- [x] `work.governance` config + defaults ; example documenté
- [x] `off` / config legacy : zéro régression pipeline dev
- [x] `per-task` : gate après `implemented` (`devTaskWithGovernanceRetries`)
- [x] Verdict parse/classify ; dry-run → PASS simulé
- [x] FAIL retry option A ; WARN advisory + trace
- [x] Tests parse, config, workflow, retry, round-trip payload, régression legacy
- [x] Quality_Gate vert (`build`, `vet`, `test`)
- [x] ADR-031 ; sync `current-spec.md` / ce handoff

## Hors scope Tranche A (interdit sans nouvelle spec)

- Modes `smart`, `per-step`, `milestone`
- execution graph / trust graph
- Commande `asa governance`
- UI / dashboard
- Refonte `review`, `verify`, factory, routing, coordination
- Statut `governance_failed`

## Quality_Gate (commande reproductible)

```bash
make build && go vet ./... && go test ./...
```

## Références

- `.kiro/specs/task-validation-gates/`
- `application/internal/workflow/governance.go`, `dev_task.go`, `workflow.go`
- `docs/ai/05-decisions.md` — ADR-031
- `docs/ai/03-standards.md`
