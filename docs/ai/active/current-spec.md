# Current spec — task-validation-gates

**Phase :** **task-validation-gates** — **Tranche A livrée** (`2026-06-06`, ADR-031)
**Handoff :** [`handoff.md`](handoff.md) — clôturé Tranche A ; prochaine spec à définir.

## Objet

Gate de gouvernance read-only après chaque tâche dev (`mode: per-task`), verdict structuré
`PASS | WARN | FAIL`, pour détecter tôt spec drift et violations d'architecture — **sans**
nouveau rôle métier obligatoire ni nouvelle commande CLI top-level.

## Décisions figées (Tranche A)

| # | Sujet | Décision |
|---|--------|----------|
| D1 | Config | `work.governance` |
| D2 | WARN | Advisory par défaut (`warn_is_advisory: true`) — continue avec trace |
| D3 | Agent default | `reviewer` ; `architect` autorisé si configuré |

## Spec & code

- **Kiro :** `.kiro/specs/task-validation-gates/` (requirements, design, tasks)
- **Code :** `internal/config/governance_config.go`, `internal/workflow/governance*.go`, `dev_task.go`,
  `pkg/asagiri/types.go` (`TaskGovernance`), `internal/report/report.go`, `.asagiri/config.yaml.example`

## Livré (Tranche A)

- Config `work.governance` + defaults (`enabled: false`)
- Modes `off` | `per-task` ; autres modes ignorés (`EnabledButInactive`)
- Gate inline dans `DevFeature` (`devTaskWithGovernanceRetries`)
- Parse YAML/JSON + classify ; dry-run → PASS simulé
- Retry option A : `retries_used < max_retries` ; passages max = `max_retries + 1`
- Trace payload, logs, report minimal
- Tests parse, config, workflow, retry 0/1/2, régression config legacy

## Hors scope (Tranche A — non livré)

- `smart`, `per-step`, `milestone`
- execution graph / trust graph
- Commande `asa governance`
- UI / dashboard
- Refonte `review` / `verify`

## Sémantique retry (canonique)

| | |
|---|---|
| `max_retries` | Relances autorisées après le 1er FAIL governance |
| `governance.retries` | Relances déjà consommées |
| `history[].retry` | Tentative governance (0 = première évaluation) |
| Passages max | `max_retries + 1` |

## Invariants

- Configs sans `work.governance` ou `enabled: false` : pipeline dev inchangé
- Unitary_Command préservées (`asa dev`, `asa verify`, `asa review`, …)
- Pas de panic aux frontières CLI

## Précédent (livré)

- **task-validation-gates Tranche A** — ADR-031 (`2026-06-06`)
- **audit-coherence-consolidation** — ADR-030 (`2026-06-05`)

Branding : **Asagiri** / **`asa`**.
