# ADR-039 — Team Cloud (control plane metadata)

**Date :** 2026-06-10  
**Statut :** accepté (spec M3)  
**Spec :** `.kiro/specs/monetization-distribution-v1/m3-team-cloud.md`

## Contexte

Les équipes ont besoin de visibilité runs et registry partagé sans déplacer l’exécution agents vers le cloud (ADR-037 local-first).

## Décision

1. **Team Cloud** = service séparé (`api.asagiri.cloud`) + console web — **pas** dans `internal/workflow`.
2. **Data plane** reste local : `asa work`, ledger `.asagiri/logs/agents/ledger.jsonl`, providers utilisateur.
3. **Sync opt-in** : push/pull explicite via client **`asagiri-cloud`** (ou futur `asa cloud` wiring léger).
4. **Registry central** : AgentSpec versionné par org/projet ; conflits par content hash.
5. **Runs centralisés** : métadonnées ledger append-only ; **pas de prompts** par défaut.
6. **Analytics** : agrégats journaliers ; coût tokens = métadonnées indicatives — Asagiri ne proxy pas les LLM.
7. **RBAC Team** : owner, admin, developer, viewer — RBAC avancé en M4.
8. **Stack** : API + PostgreSQL + object store + queue ; détails implémentation spec `team-cloud-v1` future.

## Conséquences

- Aucun compte requis pour OSS ; cloud désactivé = comportement identique M1.
- Hook sync post-run = client ou CI — pas obligatoire dans moteur.
- Notion connecteur managé optionnel — distinct du connecteur local OSS.

## Références

- ADR-037, `m3-team-cloud.md`, agent ledger T24–T29
