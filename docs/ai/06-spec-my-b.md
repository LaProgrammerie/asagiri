# Spec-my-B — Trust & Verification Engine (canon `docs/ai`)

**Statut :** livré (`2026-05-29`)  
**Spec racine :** [`spec-my-B.md`](../archives/specs/spec-my-B.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Prérequis :** [`06-spec-my-a.md`](06-spec-my-a.md) (product, runtime, analysis, investigation)

---

## 1. Résumé

Spec-my-B ajoute une couche **local-first** de validation structurée entre l’implémentation agentique et la livraison :

```text
implementation → trust checks → confidence scoring → verification gates → review → delivery
```

Le moteur répond à : *pourquoi ce changement est-il considéré comme sûr ?* et *quels risques restent non couverts ?*  
**Confidence ≠ correctness** : un score élevé n’est jamais une garantie (ADR-021).

---

## 2. Arborescence `.asagiri/trust/`

```text
.asagiri/trust/<trust-id>/
  report.md
  report.json
  replay.yaml
```

- `<trust-id>` : identifiant stable (ex. `trust-2026-05-29-aefcf60d`)
- `replay.yaml` : manifeste rejouable (`asa trust replay <id>`)

Config gates : bloc `verification:` dans `.asagiri/config.yaml` (§19).

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/trust/` | `TrustEngine`, `Engine`, rapports, gates, strict mode |
| `internal/trust/checks/` | Runners : static-analysis, contracts, flows, tests, blast-radius, observability, security, … |
| `internal/trust/confidence/` | scoring, weighting, aggregation, normalization (6 dimensions §7) |
| `internal/trust/replay/` | Chargement / manifest `replay.yaml` |
| `internal/trust/safeid/` | IDs trust déterministes |
| `internal/config/` | `VerificationConfig`, `GateProfile` |
| `internal/runtime/` | Émission événements `verification.*`, `trust.*` (§18) |
| `internal/cli/` | `verify trust`, `trust gates`, `trust replay`, `work --strict-trust` |

---

## 4. Pipeline et dimensions

**Pipeline (§4)** : checks enregistrés → agrégation confidence → évaluation gates → rapport.

**Six dimensions (§7)** : architecture, implementation, flow integrity, observability, security, regression.

**Checks (§8)** — ordre pipeline par défaut :

`static-analysis`, `contracts`, `flows`, `permissions`, `observability`, `analytics`, `architecture`, `security`, `performance`, `cost`, `backward-compatibility`, `migration-safety`, `blast-radius`, `tests`

Intégrations : graphes `internal/analysis/`, produit `.asagiri/products/<id>/` (flows, contracts).

---

## 5. Commandes CLI

| Commande | Rôle |
|----------|------|
| `asa verify trust <flow>` | Pipeline complet ; `--json`, `--ci`, `--strict`, `--product`, `--task`, `--branch` |
| `asa trust gates` | Affiche le profil actif et l’état des gates |
| `asa trust replay <id>` | Rejoue depuis `replay.yaml` |
| `asa work … --strict-trust` | Enchaîne implémentation + verify trust + gates (§22) |

CI : `asa verify trust <flow> --ci --json` → exit non-zéro si gate bloquant ou checks en échec.

---

## 6. Gates (§19)

```yaml
verification:
  default_profile: production
  gates:
    production:
      min_confidence:
        architecture: 0.8
        implementation: 0.75
        security: 0.85
      required_checks:
        - contracts
        - flows
        - observability
        - security
```

---

## 7. Runtime et replay (§18, §21)

Événements typiques : `verification.started`, `verification.completed`, `trust.low_confidence`, `security.issue_detected`, `flow.integrity_failed`, `contract.breaking_change_detected`.

`replay.yaml` capture : checks, commit, flow, branche, commandes associées.

---

## 8. Documentation publique (site)

Pages par locale `docs-site/content/docs/{en,fr,de,es}/` :

| Sujet | Chemins |
|-------|---------|
| Concept | `concepts/trust-engine` |
| CLI | `cli/verify-trust`, `cli/trust-gates`, `cli/trust-replay` |
| Config | section `verification` dans `configuration/config-file` |
| Généré | `en/cli/generated/verify-trust`, `trust`, `trust-gates`, `trust-replay` |

---

## 9. Décisions

- **ADR-020** — Trust engine local-first (`.asagiri/trust/`, pas de service distant)
- **ADR-021** — Confidence scoring explicite ; score ≠ garantie de correction

---

## 10. Validation

```bash
cd application && go test ./...
make build && ./bin/asa docs generate-cli
./bin/asa verify trust onboarding-flow --json
./bin/asa trust gates
./bin/asa trust replay trust-2026-05-29-<id>
```

Traçabilité handoff : matrice B-* dans [`active/handoff.md`](active/handoff.md).
