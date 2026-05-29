# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche :** `spec-my-B` (Asagiri Trust & Verification Engine) — **livrée** (`2026-05-29`).

## Objectif

Livrer intégralement [`spec-my-B.md`](../../../spec-my-B.md) (§1–29, critères d'acceptation §26). **Terminé.**

---

## Périmètre autorisé

- `application/internal/trust/**` (nouveau)
- `application/internal/cli/**` (commandes trust/verify)
- `application/internal/config/**` (bloc `verification:` gates)
- `application/internal/runtime/**` (émission événements §18 uniquement)
- `application/internal/workflow/**` ou `work.go` (flag `--strict-trust` §22)
- `application/pkg/asagiri/**` si types exportés
- `docs/ai/**`, `docs-site/content/docs/**` (EN/FR/DE/ES)
- `.asagiri/config.yaml.example`
- `spec-my-B.md`

---

## Hors scope

- [`spec-phase-finale.md`](../../../spec-phase-finale.md)
- Commit / push par l'agent

---

## Lots — Definition of Done

### Lot 1 — Modèles, squelette engine, rapports (§6, §9–11, §24, §25)

> **Lot 1 livré** — 2026-05-29

**DoD testable :**

- [x] `go test ./internal/trust/...` — modèles, sérialisation report JSON/MD
- [x] Génération report vers `.asagiri/trust/<id>/` (test unitaire ou integration avec fixture)
- [x] `TrustEngine` interface + `engine.go` squelette exécutable sans checks métier

---

### Lot 2 — Pipeline Verify, checks de base, confidence (§4, §7, §8, §11)

> **Lot 2 livré** — 2026-05-29

**DoD testable :**

- [x] `go test ./internal/trust/checks/...` — 4 checks de base
- [x] Score confidence agrégé sur 6 dimensions §7
- [x] Pipeline Verify produit findings + score depuis fixtures produit

---

### Lot 3 — Checks avancés, dimensions restantes (§12–17, §8)

> **Lot 3 livré** — 2026-05-29

**DoD testable :**

- [x] Tous les types de check §8 enregistrés et testés unitairement
- [x] Blast radius calcule dépendances / flows sensibles (fixture)
- [x] Confidence final couvre les 6 dimensions avec pondération documentée

---

### Lot 4 — CLI, gates, acceptation (§5, §19, §23, §25, §26)

> **Lot 4 livré** — 2026-05-29

**DoD testable :**

- [x] `go test ./internal/cli/...` — commandes trust/verify
- [x] `./bin/asa verify trust <flow> --json` retourne JSON valide
- [x] `./bin/asa trust gates` lit config et affiche état gates
- [x] `--ci` exit code non-zéro si gate bloquant échoue

---

### Lot 5 — Runtime, replay, review, work strict (§18, §20, §21, §22)

> **Lot 5 livré** — 2026-05-29 — `internal/runtime/verification.go`, `internal/trust/replay/`, `work --strict-trust`, review gates §20

**DoD testable :**

- [x] Événements émis et capturés en test runtime (`runtime/verification_test.go`)
- [x] `replay.yaml` généré et rejouable via `asa trust replay` (fixture `trust-2026-05-29-aefcf60d`)
- [x] `asa work --strict-trust` bloque si confiance sous seuil
- [x] Review orchestration enchaîne gates + review selon spec §20

---

### Lot 6 — Documentation canon + site (Phase 2 doc)

> **Lot 6 livré** — 2026-05-29

**DoD testable :**

- [x] Canon `06-spec-my-b.md` trace §1–29 vers code
- [x] Site 4 locales sans stubs sur pages trust
- [x] `context-map.md` référence spec-my-B + handoff
- [x] ADR-020/021 dans `05-decisions.md`
- [x] CLI `generated/` à jour (`make build && ./bin/asa docs generate-cli`)

---

## Matrice traçabilité

| ID | Livrable | Lot | Statut |
|----|----------|-----|--------|
| B-CLI-1 | `asa verify trust` + flags | 4 | [x] |
| B-CLI-2 | `asa trust gates` | 4 | [x] |
| B-CLI-3 | `asa trust replay` | 5 | [x] |
| B-CLI-4 | `asa work --strict-trust` | 5 | [x] |
| B-CLI-5 | `--ci --json` | 4 | [x] |
| B-REP-1 | `.asagiri/trust/<id>/report.md` + `.json` | 1 | [x] |
| B-REP-2 | `.asagiri/trust/<id>/replay.yaml` | 5 | [x] |
| B-MOD-1 | VerificationCheck, Finding models | 1 | [x] |
| B-MOD-2 | Confidence engine | 2 | [x] |
| B-MOD-3 | TrustEngine interface + engine.go | 1–2 | [x] |
| B-CHK-1 | checks/ (tous types §8) | 2–3 | [x] |
| B-DIM-1 | 6 dimensions confidence §7 | 2–3 | [x] |
| B-BR-1 | Blast radius §12 | 3 | [x] |
| B-GAT-1 | Gates blocking §19 | 4 | [x] |
| B-EVT-1 | Runtime events §18 | 5 | [x] |
| B-REV-1 | Review orchestration §20 | 5 | [x] |
| B-INT-1 | Intégration analysis/contracts/flows | 2–3 | [x] |
| B-ACC-1 | Critères acceptation §26 | 4–5 | [x] |
| B-DOC-1 | Canon + site 4 locales | 6 | [x] |

---

## DoD global spec-my-B (§26)

- [x] `asa verify trust` fonctionne
- [x] Plusieurs checks spécialisés enregistrés (`internal/trust/checks/`, types §8)
- [x] Trust report généré sous `.asagiri/trust/<id>/`
- [x] Scores explicables (6 dimensions + overall dans report)
- [x] Findings avec évidences (catégorie, message, liens check)
- [x] Gates peuvent bloquer un workflow (`verification:` + `--ci`)
- [x] Vérifications rejouables (`asa trust replay`, `replay.yaml`)
- [x] Flows / contracts validables (checks `flows`, `contracts`)
- [x] Risques visibles (blast radius, residual risk, warnings)
- [x] Événements runtime émis (§18, tests runtime)
- [x] Résultats exportables JSON (`--json`, `report.json`)
- [x] Runtime local-first (pas de service trust distant)

---

## Validation globale

```bash
cd application && go test ./...
make build && ./bin/asa docs generate-cli
./bin/asa verify trust workspace-onboarding --json
./bin/asa trust gates
./bin/asa trust replay <trust-id> --json
```

---

## Audit handoff (clôture — 2026-05-29)

| Vérification | Résultat |
|--------------|----------|
| Lots 1–6 + DoD global §26 | OK — tous [x], note 2026-05-29 |
| `go test ./...` (`application/`) | OK — exit 0, 38 packages `ok`, 0 `FAIL` |
| `go test ./internal/trust/... ./internal/cli/...` | OK — 7 packages (trust×5 + cli×2) |
| `06-spec-my-b.md` + ADR-020/021 + `02-architecture.md` | OK — canon + `internal/trust/` documenté |
| Site EN/FR/DE/ES | OK — `concepts/trust-engine` + `cli/verify-trust`, `trust-gates`, `trust-replay` (12 MDX) ; EN `cli/generated/trust*.mdx` (4 fichiers) |
| `current-spec.md` | OK — phase spec-my-B livrée, lots 1–6, CLI/docs listés |
| `context-map.md` | OK — entrée « Trust & Verification Engine (spec-my-B, livré) » |
| `make build` | OK — `bin/asa` (Version=dev, Commit=bef75b8, Date=2026-05-29T08:20:06Z) |
| `./bin/asa docs generate-cli` | OK — `docs-site/content/docs/en/cli/generated` |
| `./bin/asa verify trust workspace-onboarding --json` | OK — exit 0, JSON valide (`trust-2026-05-29-f045cd36`) |
| `./bin/asa trust gates` | OK — exit 0 (`Verification gates: not configured` sur template) |
| `./bin/asa trust replay trust-2026-05-29-aefcf60d --json` | OK — exit 0, replay produit un nouveau `trust_id` |
| Artefact exemple `.asagiri/trust/trust-2026-05-29-aefcf60d/` | OK — `report.md`, `report.json`, `replay.yaml` (14 checks, overall 25 %) |

---

## Références

- [`spec-my-B.md`](../../../spec-my-B.md)
- [`06-spec-my-b.md`](../06-spec-my-b.md)
- [`06-spec-my-a.md`](../06-spec-my-a.md) (prérequis livré)
- [`spec-my-A.md`](../../../spec-my-A.md) — livrée 2026-05-27
