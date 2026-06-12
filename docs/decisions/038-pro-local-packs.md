# ADR-038 — Pro Local packs (`asagiri-pack-v1`)

**Date :** 2026-06-10  
**Statut :** accepté (spec M2)  
**Spec :** `.kiro/specs/monetization-distribution-v1/m2-pro-local.md`

## Contexte

La monétisation Pro Local doit rester **additive** (ADR-037) : contenu premium sans feature flags dans `cmd/asa`.

## Décision

1. Format archive **`.asagiri-pack`** avec manifest **`pack.yaml`** (`apiVersion: asagiri-pack/v1`).
2. **Signature Ed25519** détachée (`manifest.sig`) ; vérification par **`asagiri-packs`** — pas dans le moteur.
3. **SemVer pack** indépendant du binaire `asa` ; champs `asagiri_min` / `asagiri_max`.
4. **Dépendances DAG** entre packs ; lock file `.asagiri/packs/lock.yaml`.
5. **Fusion** vers `.asagiri/agents/`, `config.d/`, graphs, knowledge, doctor rules, UI assets.
6. **Licence contenu** distincte d’Apache 2.0 (`LaProgrammerie-Pro-1.0`).
7. **Install / update / rollback** via `asagiri-packs` avec snapshots `history/`.

## Conséquences

- Repo moteur inchangé ; CLI packs = binaire ou repo séparé.
- Marketplace partenaire (M2.5) réutilise le même format.
- Enterprise (M4) étend avec registre privé et clés org.

## Références

- ADR-037, `m2-pro-local.md`, `agentspec.SemanticHash` (analogie content hash)
