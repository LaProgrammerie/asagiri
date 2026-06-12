# ADR-035 — Trust Gate work (`work.gates.trust`)

**Date :** 2026-06-08  
**Statut :** Accepté  
**Phase :** Trust Engine T12

## Contexte

Trust Engine V1 (`internal/worktrust`) synthétise un score advisory via `asa trust *`. Les work gates (ADR-031/032) bloquent le workflow via `gates.history` et des guards par étape. ADR-033 spécifiait un pont futur `WorkTrustReportToGateResult` sans l’implémenter.

## Décision

1. **Gate work `trust`** sous `work.gates.trust`, **désactivée par défaut** (`enabled: false`, `mode: per-task`).
2. **Mapper** `worktrust.WorkTrustReportToGateResult` → `gates.Result` (seuils config, sans modifier le scorer).
3. **Hook** post-`verify_evidence` dans `VerifyFeature` : persiste `gates.history` + logs `.asagiri/logs/<task>/gates/trust.{json,log}`.
4. **Review guard** : FAIL ou WARN non-advisory bloque `review` ; la tâche reste `verified`.
5. **`internal/trust`** (spec-my-B) **inchangé**.

## Config (defaults)

```yaml
work:
  gates:
    trust:
      enabled: false
      mode: per-task
      min_score: 70
      block_verdicts: [blocked]
      warn_verdicts: [risky]
      warn_is_advisory: true
```

## Mapping verdict

| Condition | `gates.Result.status` |
|-----------|------------------------|
| verdict ∈ `block_verdicts` ou score < `min_score` | `fail` |
| verdict ∈ `warn_verdicts` | `warn` |
| sinon | `pass` |

## Conséquences

- Trust advisory CLI (`asa trust`) et enforcement gate sont distincts mais partagent la même synthèse live.
- Recommandations alignées intent restent au niveau CLI via `worktrustrecommend` (évite cycle import workflow ↔ worktrust).
