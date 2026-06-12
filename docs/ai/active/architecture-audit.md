# Architecture audit — hygiène technique

> **Date :** 2026-06-09 (màj extractions `agentscli` / `trustcli` / `doctorcli` / `workcli` / `knowledgecli` / `graphcli` / `replaycli` / `onboardingcli`)  
> **Périmètre :** formatage Go (`gofmt`), cartographie LOC, risques de couplage, extraction CLI par domaine.  
> **Hors scope :** changement fonctionnel, refactor workflow/gates/trust/agentspec.

## Synthèse

| Indicateur | Valeur |
|------------|--------|
| LOC Go (`application/`, hors vendor) | ~111 724 |
| Fichiers `.go` | 937 |
| Fichiers reformatés par `gofmt` | 132 |
| Delta net post-gofmt | −20 lignes (alignement uniquement) |
| `go build ./...` | ✅ |
| `go test ./... -count=1` | ✅ |
| `docgen TestNodiff` | ✅ |

`gofmt` n’altère pas le comportement : uniquement espaces, indentation et retours à la ligne. Aucune modification de logique dans workflow, gates, trust ou agentspec hors formatage.

---

## Top 20 packages par taille (LOC)

| Rang | Package | Fichiers | LOC | % codebase | Importeurs* |
|------|---------|----------|-----|------------|-------------|
| 1 | `internal/cli` | 60 | ~9 124 | ~8,2 % | 6 |
| — | `internal/workcli` | 11 | ~1 026 | ~0,9 % | 1 (`cli`) |
| — | `internal/graphcli` | 5 | ~702 | ~0,6 % | 1 (`cli`) |
| — | `internal/replaycli` | 4 | ~315 | ~0,3 % | 1 (`cli`) |
| — | `internal/knowledgecli` | 4 | ~355 | ~0,3 % | 1 (`cli`) |
| — | `internal/trustcli` | 5 | ~717 | ~0,6 % | 1 (`cli`) |
| — | `internal/agentscli` | 2 | ~476 | ~0,4 % | 1 (`cli`) |
| — | `internal/doctorcli` | 4 | ~240 | ~0,2 % | 1 (`cli`) |
| — | `internal/onboardingcli` | 2 | ~98 | ~0,1 % | 1 (`cli`) |
| 2 | `internal/executiongraph` | 48 | 7 610 | 6,8 % | 33 |
| 3 | `internal/ui/bus` | 20 | 6 380 | 5,7 % | 55 |
| 4 | `internal/workflow` | 38 | 6 032 | 5,4 % | 6 |
| 5 | `internal/knowledge` | 39 | 4 703 | 4,2 % | 67 |
| 6 | `internal/ui/app` | 24 | 4 231 | 3,8 % | 5 |
| 7 | `internal/coordination` | 40 | 3 590 | 3,2 % | 20 |
| 8 | `internal/config` | 24 | 3 507 | 3,1 % | 222 |
| 9 | `internal/onboarding` | 23 | 3 241 | 2,9 % | 26 |
| 10 | `internal/intent` | 22 | 3 125 | 2,8 % | 23 |
| 11 | `internal/knowledge/extractors` | 29 | 2 983 | 2,7 % | (sous `knowledge`) |
| 12 | `internal/trust/checks` | 42 | 2 764 | 2,5 % | 9 |
| 13 | `internal/worktrust` | 15 | 2 611 | 2,3 % | 12 |
| 14 | `internal/trust` | 27 | 2 603 | 2,3 % | 12 |
| 15 | `internal/replay` | 17 | 2 300 | 2,1 % | 6 |
| 16 | `internal/ui/screens/onboarding` | 11 | 2 290 | 2,0 % | — |
| 17 | `internal/product` | 16 | 2 100 | 1,9 % | 16 |
| 18 | `internal/investigation` | 26 | 2 092 | 1,9 % | 22 |
| 19 | `internal/doctor` | 10 | 1 903 | 1,7 % | 3 |
| 20 | `internal/agentledger` | 15 | 1 730 | 1,5 % | 11 |

\*Nombre de fichiers Go important le package (approximation couplage entrant).

Les 20 premiers packages représentent **~62 %** du code Go.

---

## Responsabilités actuelles

### 1. `internal/cli` (~9 124 LOC, hors `cli/docgen`)
Point d’entrée Cobra : wiring root, UI screens. Délègue `asa onboard` / `asa ready` → `onboardingcli`, `asa work` → `workcli`, `asa graph` (+ `plan graph`/`plan explain`) → `graphcli`, `asa knowledge` → `knowledgecli`, `asa replay` → `replaycli`, `asa trust` → `trustcli`, `asa agents` → `agentscli` (+ `watch`), `asa doctor` → `doctorcli`.

### 1b. `internal/workcli` (~1 026 LOC)
Commande `asa work "<instruction>"` : classification intent, Product Layer (`HandleProductLayer`), strict-trust (`ResolveStrictTrust`), résolution agent (`ResolveWorkAgent`), affichage V3 (`PrintEstimateBoxed`, `PrintWorkSummary`). Injection `LoadWorkContext` depuis `cli`.

### 1c. `internal/trustcli` (~717 LOC)
Commandes `asa trust` (gates, replay, diff, task/feature/run work trust) et `asa verify trust`. Injection `LoadWorkContext` + `RunRootUI` depuis `cli`.

### 1d. `internal/agentscli` (~476 LOC)
Commandes `asa agents` (list, show, run, runs, diff, export, stats, sync, external).

### 1e. `internal/doctorcli` (~240 LOC)
Commandes `asa doctor` (--json, --full, --strict, --save), `asa doctor architecture` (--json, lecture seule) et `asa doctor diff`. Réutilise `trustcli.ResolveDiffPaths` et `trustcli.PrintReportSaved` pour les helpers snapshot.

### 1f. `internal/knowledgecli` (~355 LOC)
Commandes `asa knowledge` (build, query, explain, snapshot). Injection `LoadContext` + `RunRootUI` depuis `cli`. `asa impact` reste dans `cli` (hors périmètre knowledge).

### 1g. `internal/graphcli` (~702 LOC)
Commandes `asa graph` (run, status, resume, rollback, visualize) et `asa plan graph` / `asa plan explain`. Types JSON exportés (`PlanGraphResult`, `GraphRunJSONResult`, …), erreurs `ErrCIFailed` / `ErrNotEnabled` / `ErrFlowRequired`.

### 1h. `internal/replaycli` (~315 LOC)
Commandes `asa replay` (open, create, run, compare, explain, snapshot). Injection `LoadContext` + `OpenReplayScreen` (TUI Replay Explorer) depuis `cli`. `asa trust replay` reste dans `trustcli`.

### 1i. `internal/onboardingcli` (~98 LOC)
Commandes `asa onboard` (wizard, flags `--yes`, `--non-interactive`, `--stack`, `--check-only`, `--resume`, `--step`, `--plain`, `--json`, `--ci`, `--strict`, `--force-docs`, `--dry-run`, `--ui`) et `asa ready` (`--plain`, `--json`, `--ci`, `--strict`, `--autofix`). Délègue à `internal/onboarding` ; injection TUI `RunOnboardUI` depuis `cli`. Packages `internal/onboarding`, `internal/onboarding/detect`, `internal/doctor`, `internal/trust` inchangés.

### 2. `internal/executiongraph` (7 610 LOC)
Planification et exécution de graphes (spec-my-C) : runner, rollback, enrichissement trust, pont knowledge.

### 3. `internal/ui/bus` (6 380 LOC)
CommandBus / QueryBus de l’UI Bubble Tea : dispatch vers services métier, handlers read-only et mutations.

### 4. `internal/workflow` (6 032 LOC)
Orchestration runs/steps : dev, enrich, review, gates, trust gate, enregistrement ledger agent.

### 5. `internal/knowledge` (+ extractors 2 983 LOC)
Graphe de connaissance structurelle (spec-my-E), extraction, rendu, persistance SQLite.

### 6. `internal/ui/app` (4 231 LOC)
Modèle Tea, router écrans, bootstrap UI `asa` (mission, dashboard, agents theatre…).

### 7. `internal/coordination` (3 590 LOC)
Rôles agents, handoffs, policies, pipeline multi-agents (spec-my-D).

### 8. `internal/config` (3 507 LOC)
Schéma typé `config.yaml` : work, gates, trust, agents, sources, intent.

### 9. `internal/onboarding` (3 241 LOC)
Wizard init, doctor onboarding, autofix registry agents.

### 10. `internal/intent` (3 125 LOC)
Resolver, planner, executor specv2, snapshots, recommandations alignement.

### 11–20
Packages domaine spécialisés : trust checks, worktrust (CLI trust work), replay ingénierie, product layer, investigation, doctor santé repo, agentledger T24–T29.

---

## Risques de couplage

| Zone | Risque | Sévérité |
|------|--------|----------|
| **`cli` → tout** | God-package en réduction (−~3 088 LOC vs audit initial) : extractions `workcli` / `graphcli` / `knowledgecli` / `replaycli` / `agentscli` / `trustcli` / `doctorcli` / `onboardingcli` ; reste workflow, investigate, UI… | **Élevé** (en baisse) |
| **`config` hub** | 222 importeurs : toute évolution de struct casse de nombreux packages. Attendu mais fragile. | **Moyen** |
| **`ui/bus` ↔ domaine** | 55 importeurs + handlers volumineux : logique métier parfois dupliée avec `cli`. | **Élevé** |
| **`executiongraph` ↔ trust/knowledge** | Graphe d’exécution enrichi trust + pont knowledge : cycle conceptuel possible. | **Moyen** |
| **`workflow` central** | Peu d’importeurs (6) mais gros package : point unique pour dev/gates/ledger. | **Moyen** |
| **`knowledge` + extractors** | ~7,7k LOC cumulés, 67 importeurs : surface d’API large. | **Moyen** |
| **Doublon replay** | `internal/replay` vs `internal/trust/replay` : confusion nommage. | **Faible** |
| **Agent platform T13–T29** | `agentledger` 1,7k LOC en croissance rapide ; reste bien isolé (11 importeurs, lecture seule). | **Faible** |

---

## Recommandations de découpage

### Court terme (hygiène, sans feature)

1. **`cli` par domaine** — ✅ `workcli`, ✅ `graphcli`, ✅ `knowledgecli`, ✅ `replaycli`, ✅ `agentscli`, ✅ `trustcli`, ✅ `doctorcli`, ✅ `onboardingcli`. Prochaines cibles : `impact`, `investigate`, `analysis` groupes.
2. **Geler les imports `cli` → `workflow`** — Nouvelles features agent/ledger passent par packages `internal/agent*` ; `cli` reste façade mince.
3. **Documenter frontières `ui/bus` vs `cli`** — Une seule voie d’appel par capacité (éviter double maintenance handlers).

### Moyen terme

4. **`executiongraph`** — Séparer `runner` / `rollback` / `trust_enrichment` en fichiers packages ou sous-dossiers avec interfaces explicites.
5. **`knowledge`** — Isoler `extractors` derrière une interface `Extractor` registry ; réduire imports croisés depuis `executiongraph`.
6. **`config`** — Découper par domaine (`config/work.go`, déjà partiel) ; envisager validation par sous-struct lazy-load pour limiter recompilations.

### Long terme

7. **Facade `internal/app`** — Couche application unique appelée par `cli` et `ui/bus` (CQRS léger) pour éliminer la duplication dispatch.
8. **Budget LOC `cli`** — Objectif < 8 000 LOC dans `internal/cli` (extraire docgen déjà à part, continuer extraction command groups).

### Non recommandé maintenant

- Refactor workflow/gates/trust pour le découpage seul.
- Fusion `replay` / `trust/replay` sans spec dédiée.
- Split physique du module Go (monorepo unique reste cohérent à ~112k LOC).

---

## Actions réalisées (cette passe)

- [x] `gofmt -w` sur `application/` (132 fichiers)
- [x] Validation build + tests + docgen nodiff
- [x] Rapport présent (ce fichier)
- [x] Extraction **`internal/agentscli`** (~476 LOC) — `asa agents` hors TUI `watch`
- [x] Extraction **`internal/trustcli`** (~717 LOC) — `asa trust` + `asa verify trust`
- [x] Extraction **`internal/doctorcli`** (~240 LOC) — `asa doctor` + `asa doctor architecture` + `asa doctor diff`
- [x] Extraction **`internal/workcli`** (~1 026 LOC) — `asa work` (+ product layer, strict-trust, agent)
- [x] Extraction **`internal/knowledgecli`** (~355 LOC) — `asa knowledge` (build, query, explain, snapshot)
- [x] Extraction **`internal/graphcli`** (~702 LOC) — `asa graph` + `asa plan graph` / `plan explain`
- [x] Extraction **`internal/replaycli`** (~315 LOC) — `asa replay` (open, create, run, compare, explain, snapshot)
- [x] Extraction **`internal/onboardingcli`** (~98 LOC) — `asa onboard` + `asa ready`

### Delta LOC `internal/cli` (extractions CLI)

| Extraction | LOC extraite | Fichiers retirés de `cli` | Wiring restant dans `cli` |
|------------|--------------|---------------------------|---------------------------|
| `agentscli` | ~433 (`cmd.go`) + test | `agents_cmd.go` | `root_ui.go` (`newAgentsCmd` + `watch`) |
| `trustcli` | ~610 (`trust_*_cmd.go`) + helpers | `trust_cmd.go`, `trust_work_cmd.go`, `trust_diff_cmd.go` | `trust_wiring.go` (~22 LOC), `verify` dans `root.go` |
| `doctorcli` | ~239 (`cmd.go`, `architecture_cmd.go`, `diff_cmd.go`) + test | `doctor.go`, `doctor_diff.go` | `doctor_wiring.go` (~10 LOC) |
| `workcli` | ~996 (`cmd.go` + product/trust/agent/display) + tests | `work.go`, `work_product.go`, `work_trust.go`, `work_agent.go` + tests | `work_wiring.go` (~30 LOC) ; `estimate.go` / `root.go` appellent `PrintEstimateBoxed` / `ResolveWorkAgent` |
| `knowledgecli` | ~334 (`cmd.go` + helpers) + test | `knowledge_cmd.go` | `knowledge_wiring.go` (~22 LOC) |
| `graphcli` | ~680 (`cmd.go` + engine/types/errors) + test | `graph_cmd.go`, `graph_cmd_test.go` | `graph_wiring.go` (~22 LOC) ; `plan` dans `root.go` via `newPlanGraphCmd` / `newPlanExplainCmd` |
| `replaycli` | ~294 (`cmd.go` + context/helpers) + test | `replay_cmd.go` | `replay_wiring.go` (~26 LOC) ; tests intégration dans `replay_cmd_test.go` |
| `onboardingcli` | ~75 (`cmd.go`) + test | `onboard_cmd.go` | `onboarding_wiring.go` (~16 LOC) |
| **Total** | **~3 883 LOC** dans packages dédiés | **−18 fichiers** | **`cli` : 12 212 → ~9 124 LOC (−~3 088)** |

Comportement utilisateur inchangé : mêmes commandes Cobra, flags, stdout/stderr, codes d’erreur (`graphcli.ErrCIFailed`, `doctorcli.ErrFailed`, `trustcli.ErrCIFailed`, etc.). `asa impact` non déplacé. `asa trust replay` reste dans `trustcli` (distinct de `asa replay`).

## Prochaine étape suggérée

Prioriser **extraction `impact` / `investigate` / `analysis` CLI** comme prochain découpage à faible risque ; objectif budget `< 8 000 LOC` pour `internal/cli` hors docgen.
