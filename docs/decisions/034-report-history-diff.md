# ADR-034 — Report history & diff

**Date :** 2026-06-08  
**Status :** accepted  
**Spec :** [`.kiro/specs/trust-engine-v1/`](../../.kiro/specs/trust-engine-v1/) (T11)  
**Related :** [ADR-033](./033-trust-diagnostic-architecture.md) §3.4

## Context

Trust Engine V1 (T8) a introduit `reportsink` avec snapshots opt-in (`--save`) et chemins stables `latest` par scope. Les opérateurs et la CI n'avaient pas de moyen natif de voir **ce qui a changé** entre deux diagnostics ou deux synthèses trust sans diff externe.

## Decision

### 1. History on save (optional, default on)

Lors d'un `--save`, **avant** d'écraser le snapshot stable :

1. Copier le fichier existant vers `history/` (si présent).
2. Écrire le nouveau contenu sur le chemin stable (atomic write inchangé).

Chemins :

```text
.asagiri/reports/doctor/latest.json              # contrat stable inchangé
.asagiri/reports/doctor/history/<timestamp>.json

.asagiri/reports/trust/tasks/<task-id>.json
.asagiri/reports/trust/tasks/history/<task-id>_<timestamp>.json

.asagiri/reports/trust/features/<feature>.json
.asagiri/reports/trust/features/history/<feature>_<timestamp>.json

.asagiri/reports/trust/runs/<run-id>.json
.asagiri/reports/trust/runs/history/<run-id>_<timestamp>.json
```

- Timestamp UTC : `20060102T150405Z`.
- `SaveOptions.KeepHistory` permet de désactiver l'archivage (tests, outils).
- **Aucune relecture** des snapshots pour scorer ou planner (invariant ADR-033).

### 2. Package `internal/reportdiff`

Diff read-only entre deux snapshots JSON déjà enregistrés :

| Delta | Trust task | Trust feature/run | Doctor |
|-------|------------|-------------------|--------|
| Score | overall | overall moyen | — |
| Verdict | oui | oui | trust.verdict optionnel |
| Dimensions | 6 axes | — | — |
| Next action | recommendation.command | next_actions[0] | next_actions[0].cli |
| Autres | — | task_count | ready, warnings, failures |

Schéma diff : `report_version: "report-diff-v1"`.

### 3. CLI

| Commande | Comparaison par défaut |
|----------|------------------------|
| `asa doctor diff` | `history/` (dernier) vs `latest.json` |
| `asa trust diff task <id>` | idem par scope |
| `asa trust diff feature <feature>` | idem |
| `asa trust diff run <id>` | idem |

- `--from` / `--to` pour chemins explicites.
- `--json` pour sortie structurée.
- Erreur claire si moins de deux `--save` sur le scope.

### 4. Non-goals T11

- Pas de GC / retention automatique.
- Pas de diff live vs saved (backlog).
- Pas de modification scoring, gates, workflow.

## Consequences

- **Positif :** traçabilité opérateur et CI sans DB ; contrat `latest.json` préservé.
- **Négatif :** croissance disque locale — à ignorer en git ou politique manuelle.
- **Writer unique :** seul `reportsink` écrit sous `.asagiri/reports/`.
