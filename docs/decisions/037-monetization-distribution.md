# ADR-037 — Monetization & Distribution V1

**Date :** 2026-06-10  
**Statut :** accepté (spec M0)  
**Spec :** `.kiro/specs/monetization-distribution-v1/`

## Contexte

Asagiri dispose d’un moteur local complet (workflow, gates, trust, agentspec, ledger T24–T29, doctor, knowledge, execution graph, UI). Avant toute monétisation, il faut fixer **éditions**, **périmètre OSS** et **modèle commercial** sans bridage artificiel ni licence enforcement dans le CLI.

Positionnement retenu :

- **« Terraform pour l’orchestration IA de dev »** — déclaratif, reproductible, état versionné.
- **« Local-first AI development control plane »** — exécution sur la machine du dev ; cloud optionnel.

## Décision

### 1. Quatre éditions

| Édition | Rôle |
|---------|------|
| **OSS** | Moteur complet open source (Apache 2.0) — une codebase |
| **Pro Local** | Packs premium offline (agents, gates, knowledge, UI) — contenu additif |
| **Team Cloud** | Collaboration opt-in (workspace, sync registry/ledger, dashboard) |
| **Enterprise** | Team Cloud + SSO, RBAC, audit, on-prem, registre packs privé |

### 2. Monétisation additive uniquement

- **OSS** : aucune limite artificielle sur tâches, runs, agents ou dépôts.
- **Pro** : vente de **catalogues de contenu** et support — pas de feature flags dans `internal/workflow`, `gates`, `trust`, `agentspec`.
- **Cloud / Enterprise** : **services managés** et gouvernance — jamais requis pour `asa work` local.

### 3. Inviolables (jamais paywallés)

Boucle orchestration (`work`, gates, trust), formats ouverts (AgentSpec, JSON reports), ledger agent local (T24–T29), doctor/trust diagnostic, export run bundle, providers agents locaux, état `.asagiri/`, SDK npm OSS, contrats `--json` stdout/stderr.

### 4. Pas d’enforcement dans le moteur (cette phase)

Validation de packs premium = outil ou service **séparé** (`asagiri-packs` futur, cloud) — **pas** dans `cmd/asa`.

### 5. Ordre de mise en marché

1. **M1 — Distribution OSS** (releases, install, docs)
2. **M2 — Packs Pro Local**
3. **M3 — Team Cloud**
4. **M4 — Enterprise**

## Conséquences

- Le code livré (T1–T29, trust, graphs, knowledge) reste **OSS** ; toute PR qui paywall le moteur est refusée sans révision ADR.
- Les specs futures (`team-cloud-v1`, `asagiri-packs`) sont **hors** du module Go principal ou en binaires distincts.
- `current-spec.md` et `handoff.md` pointent vers `monetization-distribution-v1` pour la phase produit active.
- Notion et embeddings cloud tiers restent des **connecteurs/config utilisateur** — distincts du SaaS Team Cloud Asagiri.

## Références

- `.kiro/specs/monetization-distribution-v1/requirements.md` — exigences et inviolables
- `.kiro/specs/monetization-distribution-v1/design.md` — matrice capacités × éditions
- `docs/ai/active/distribution-oss.md` — M1 distribution OSS
- `m2-pro-local.md` / ADR-038 — packs Pro Local
- `m3-team-cloud.md` / ADR-039 — Team Cloud
- `m4-enterprise.md` / ADR-040 — Enterprise
- ADR-011 (Apache 2.0), ADR-033 (JSON streams), ADR-036 (agent platform)
