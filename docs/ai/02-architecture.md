# Architecture

## Vue d’ensemble

- **Produit :** Asagiri — CLI Cobra (`application/cmd/asa`), commande **`asa`**.
- **Langage :** Go — module `github.com/LaProgrammerie/asagiri` (dépôt GitHub inchangé en phase 1 — ADR-016), code sous `application/internal/`.
- **État local :** répertoire **`.asagiri/`** dans chaque dépôt Git cible.
- **Runtime local :** Docker Compose sous `infrastructure/docker/` (dev Go + deps) ; optionnel pour le CLI pur.
- **Orchestration dev :** `Makefile` (`make build`, `make test`, `make dev`).

## Arborescence applicative

```
application/
  cmd/asa/main.go
  internal/
    cli/                 # Cobra : primitives V1 + work/continue/next/inbox/sync
    intent/              # resolver, planner, executor (specv2)
    source/              # LocalSource, Notion (specv2 §7–8)
    config/              # config.yaml typée (+ intent, work, sources)
    bootstrap/           # init, doctor, GitRoot
    env/                 # ASA_* + fallback AGENTFLOW_* (compat)
    agent/               # interface Agent
    agent/exec/          # subprocess agents (sans shell)
    worktree/            # git worktree par tâche
    workflow/            # orchestration runs / steps / verify / PR
    gates/               # modèle canonique work gates (Verdict, Result, parse/classify) — ADR-032
    spec/                # lecture .kiro/specs/<feature>/ + fallback current-spec.md
    plan/                # normalisation tâches markdown → JSON
    report/              # report.md + report.json
    store/sqlite/        # SQLite modernc, migrations embed
    version/
    product/               # spec-my-A : prototype, flows, contracts
    product/derivation/
    runtime/               # spec-my-A : daemon, sessions, API
    runtime/api/
    memory/
    memory/embedder/        # hash | ollama | cloud (PF-A-01, ADR-025)
    skills/
    analysis/              # spec-my-A : graphes structuraux
    investigation/         # spec-my-A : pipeline investigation
    embedutil/
    trust/                 # spec-my-B : moteur confiance & vérification
    trust/checks/
    trust/confidence/
    trust/replay/
    trust/safeid/
    executiongraph/        # spec-my-C : planner graphe d'exécution
    coordination/          # spec-my-D : rôles agents, handoffs, policies, agent.*
    knowledge/             # spec-my-E : graphe de connaissance structurelle
    knowledge/extractors/
    knowledge/renderers/
    knowledge/sqlite/
    replay/                # spec-my-F : replay packages ingénierie (≠ trust/replay)
    tui/                   # specv3 : façade rich/plain/json (pipeline work — pas Bubble Tea)
    ui/                    # spec-ui ADR-027 : Experience Platform Bubble Tea
      app/                 # tea.Model, router, bootstrap `asa`
      bus/                 # CommandBus / QueryBus → services métier
      layout/              # single, split-h, split-v (lot 1)
      theme/               # palettes Lip Gloss
      components/          # widgets composables
      screens/             # mission, dashboard, explorers…
      input/               # clavier, souris
      state/               # état navigation UI (pas SQLite métier)
.asagiri/                # créé par asa init
  config.yaml
  state.sqlite           # gitignored
  runs/ tasks/ logs/ worktrees/
  products/<product>/    # spec-my-A
  runtime/               # runtime.db, hooks.yaml, api.token
  analysis/<product>/
  investigations/
  skills/
  trust/<trust-id>/      # spec-my-B : report.md, report.json, replay.yaml
  graphs/<graph-id>/     # spec-my-C : execution-graph.yaml, plan.md, metrics.json, events.jsonl
  handoffs/<handoff-id>/ # spec-my-D : handoff.yaml (transferts structurés entre agents)
  knowledge/             # spec-my-E : graph.sqlite, graph.json, snapshots/
  replays/<replay-id>/   # spec-my-F : replay.yaml, graph/, trust/, investigations/, …
```

## Packages clés

| Package | Rôle |
|---------|------|
| `internal/config` | Struct YAML ; `Load` + validation chemins relatifs |
| `internal/store/sqlite` | DB, migrations v1–v2, CRUD `runs` / `tasks` |
| `internal/bootstrap` | `Init`, `Doctor` |
| `internal/env` | Variables `ASA_*` ; compat `AGENTFLOW_*` |
| `internal/agent` + `agent/exec` | Interface agents ; exécution `exec.Command` (pas de shell) |
| `internal/worktree` | Création / suppression worktrees Git |
| `internal/workflow` | `PlanFeature`, `DevFeature`, `VerifyFeature`, etc. ; orchestration work gates (`gate_runtime.go`, `plan_gate.go`, `governance.go`) |
| `internal/gates` | Modèle canonique **work gates** agent read-only : `Verdict`, `Result`, `Finding`, `EvidenceRef`, parse/classify (ADR-032) |
| `internal/spec` | Lecture spec Kiro ou fallback |
| `internal/plan` | Parse `tasks.md`, export JSON tâches |
| `internal/report` | Rapports de run |
| `internal/cli` | Surface utilisateur + `--dry-run` global ; `asa` sans args → `ui/app.Run` si TTY |
| `internal/tui` | Façade sortie V3 (`UI` interface : rich/plain/json) — **pas** d’event loop ; utilisé par `pipeline` et commandes `work`/`estimate` |
| `internal/ui` | Experience Platform Bubble Tea (ADR-027) ; **client** du moteur via `ui/bus` uniquement |
| `internal/ui/bus` | Adapters Command/Query → `workflow`, `pipeline`, `runtime`, `trust`, `knowledge`, `replay`, `executiongraph`, `investigation` |
| `internal/intent` | Résolution d’intention, plan haut niveau, exécution via `workflow` |
| `internal/source` | Abstraction sources ; sync vers `.asagiri/specs/<feature>/` |
| `pkg/asagiri` | Types partagés (ex-`pkg/agentflow`) |

## Flux intention (specv2)

```
instruction → IntentResolver → HighLevelPlanner → primitives (spec/plan/dev/…) → rapport
```

Sources externes (Notion) : **sync obligatoire** avant exécution — jamais de spec distante directe.

## Flux critique V1

```
spec/plan → enrich → dev (worktree + agent) → verify → review → report / pr
```

### Work validation gates (ADR-031, ADR-032)

Trois familles de validation coexistent — **ne pas les fusionner** :

| Famille | Package | Rôle | Vocabulaire statut |
|---------|---------|------|-------------------|
| **Work gate** | `internal/gates` + `workflow/*_gate.go` | Validation **read-only par agent** (plan, governance, futures gates) ; config `work.gates.*` | `Verdict` : `pass` \| `warn` \| `fail` |
| **Check** | `internal/trust/checks` | Vérification **déterministe** (contracts, static-analysis, …) ; pipeline `verify trust` | `CheckStatus` : `passed` \| `failed` \| `skipped` \| `warn` |
| **Trust threshold gate** | `internal/trust/gates.go` | **Seuil** sur agrégat checks + confidence (`verification.gates` config) | `GateStatus` : `passed` \| `blocked` \| `not_configured` |

**Modèle canonique work gates** : `gates.Result` (enveloppe complète). Le type `gates.Verdict` est **uniquement le statut** (`pass|warn|fail`), pas l'enveloppe — **ne pas renommer** `Result` en `Verdict`.

Composants :

| Concept | Type / artefact | Emplacement |
|---------|-----------------|-------------|
| **Gate** | Point de contrôle workflow (`plan`, `governance`, …) | Config `work.gates.<name>` ; runtime `workflow/gate_runtime.go` |
| **Result** (alias doc : GateEvaluation) | Enveloppe : status, confidence, findings, evidence, parse_error, … | `internal/gates/types.go` → `gates.Result` |
| **Verdict** | Statut classifié de la gate | `internal/gates/verdict.go` → `gates.Verdict` ; porté par `Result.Status` |
| **Finding** | Issue structurée (`code`, `severity`, `message`, `actions`) | `gates.Finding` (work) ; types parallèles dans `trust` / `pkg/asagiri` (dette acceptée) |
| **Evidence** | Référence vers matériel consulté (`kind`, path/source, note/summary) | `gates.EvidenceRef` (work) ; `trust.Evidence` / `checks.Evidence` (checks) |
| **Report** | Synthèse persistée | Logs gate : `.asagiri/logs/<scope-id>/gates/<gate>.{json,log}` (`gates.LogDocument`) ; run : section **`## Gates`** (`gatesMarkdown` dans `internal/report/gates_report.go`) ; trust : `.asagiri/trust/<id>/` |

**Gates livrées (Tranche A)** :

- **Plan** (`work.gates.plan`) — scope run ; après planification ; logs `.asagiri/logs/<run-id>/gates/plan.{json,log}`.
- **Governance** (`work.gates.governance`, legacy compat `work.governance`) — scope task ; après `implemented` dans `DevFeature` ; retry dev sur FAIL (`max_retries`) ; trace payload `governance.history` **et** `gates.history` (`GateHistoryEntry`) ; logs `.asagiri/logs/<task-id>/gates/governance.{json,log}`.
- **Human review** (`work.gates.human_review`) — scope task ; après governance OK (ou après `implemented` si governance inactive) ; verdict injecté (fichier ou `asa gates submit`) ; **pas d'agent** ; disabled par défaut ; logs `.asagiri/logs/<task-id>/gates/human_review.{json,log}`.

Spec détaillée : `.kiro/specs/work-gates-model/`. Pont futur **trust → gates.Result** documenté dans la spec, **non implémenté** (Phase 3).

**Gouvernance post-dev (résumé)** : si `work.gates.governance.enabled` et `mode: per-task`, chaque tâche passe une gate read-only **après** `implemented` (agent default `reviewer`, WARN advisory par défaut — voir ADR-031). Sinon le pipeline est inchangé.

État persistant : `runs.steps_json`, tâches `tasks.payload_json` + fichiers `.asagiri/tasks/<feature>/*.json`.

### Trust & Diagnostic Architecture (ADR-033)

Couche **read-only** distincte de `internal/trust` (spec-my-B) et des work gates d’enforcement :

| Package | Rôle | CLI |
|---------|------|-----|
| `internal/worktrust` | Synthèse advisory des work gates (6 dimensions, verdicts task / feature / run) | `asa trust task\|feature\|run` |
| `internal/doctor` | Hygiène dépôt, config, agents, onboarding | `asa doctor` |
| `internal/reportsink` | Snapshots JSON opt-in (`--save`, écriture atomique) | avec `--json` : stdout JSON pur ; confirmation `--save` sur **stderr** |

Artefacts : `.asagiri/reports/trust/{tasks,features,runs}/…`, `.asagiri/reports/doctor/latest.json`. Aucune relecture auto pour scorer ou planner (V1).

Référence : [`docs/decisions/033-trust-diagnostic-architecture.md`](../decisions/033-trust-diagnostic-architecture.md), spec `.kiro/specs/trust-engine-v1/`.

## Contrat Makefile

| Cible | Action |
|-------|--------|
| `make build` | `bin/asa` |
| `make test` | `go test ./...` |
| `make lint` | `golangci-lint run` (toolchain Go ≥ `go.mod`) |
| `make dev` | stack Docker dev |
| `make release-snapshot` | GoReleaser snapshot (artefacts `asa_*`) |

## Équivalence spec §11.1 ↔ dépôt (ADR-001, chemins ADR-016)

| Spec historique `agentflow/` | Réel (Asagiri) |
|------------------------------|----------------|
| `cmd/agentflow/` | `application/cmd/asa/` |
| `internal/cli/` | `application/internal/cli/` |
| `internal/workflow/` (+ `state_machine.go`) | `application/internal/workflow/` |
| `internal/agents/` | `application/internal/agent/` + `agent/exec/` |
| `internal/git/worktree` | `application/internal/worktree/` |
| `internal/validation/` | `application/internal/validation/` |
| `internal/state/sqlite` | `application/internal/store/sqlite/` |
| `internal/rag/` | `application/internal/rag/` |
| `internal/policy/` | `application/internal/policy/` |
| `internal/cost/` | Estimation tokens/coût/durée (specv3) |
| `internal/contextopt/` | Collecte, scoring, réduction, pack contexte |
| `internal/investigation/` | Grep, scan repo, symboles Go |
| `internal/telemetry/` | Métriques run/step → SQLite |
| `internal/tui/` | Affichage rich/plain/json (isolé du moteur) |
| `internal/pipeline/` | `RunV3Pipeline` (séquence work V3) |
| `internal/routing/` | Routing cost-aware local/cloud |
| `internal/mcp/` | Serveur MCP stdio |
| `pkg/agentflow/types` | `application/pkg/asagiri/` |
| `internal/product/` | Prototype, extraction flows/contracts, génération specs |
| `internal/runtime/` | Sessions, events, memory SQLite, worker, hooks |
| `internal/runtime/api/` | REST `127.0.0.1` + socket Unix |
| `internal/memory/` | Scoped memory, pluggable embedder (`hash`/`ollama`/`cloud`), cosine retrieval, `reindex` |
| `internal/memory/embedder/` | `Embedder` interface ; défaut `hash` ; Ollama local-first ; cloud opt-in |
| `internal/skills/` | `.asagiri/skills/*.yaml` |
| `internal/analysis/` | `asa analysis build` → graphs.json |
| `internal/investigation/` | Rapport, graph, impact, context-pack |
| `internal/trust/` | `TrustEngine`, rapports `.asagiri/trust/<id>/`, gates |
| `internal/trust/checks/` | Runners de vérification (contracts, flows, blast-radius, …) |
| `internal/trust/confidence/` | Agrégation 6 dimensions §7 spec-my-B |
| `internal/trust/replay/` | Manifest `replay.yaml` pour **`asa trust replay`** (spec-my-B §21) |
| `internal/replay/` | Replay packages ingénierie : capture, run, compare, divergence (**`asa replay`**, spec-my-F) |
| `internal/executiongraph/` | Planner graphe : model, dependency, scheduler, estimator, risk, checkpoints, rollback, runner |
| `internal/executiongraph/` (repository) | Persistance `.asagiri/graphs/<id>/` (YAML/JSON, plan.md, metrics) |
| `internal/runtime/` (GraphEmitter) | Événements `graph.*` §19 spec-my-C |
| `internal/coordination/` | Coordinator, assigner, handoffs, policies, `FormatMultiAgentRuntime` |
| `internal/coordination/` (handoff) | Persistance `.asagiri/handoffs/<id>/handoff.yaml` |
| `internal/runtime/` (agent events) | Événements `agent.*` §10 spec-my-D via `CoordinationEmitter` |
| `internal/knowledge/` | Graphe de connaissance : modèles, store SQLite, builder, query, impact, staleness, snapshot |
| `internal/knowledge/extractors/` | Extraction flows, contracts, code, tests, adr, infra, runtime |
| `internal/knowledge/renderers/` | Export JSON, DOT, Mermaid |

Détail spec-my-A : [`06-spec-my-a.md`](06-spec-my-a.md).  
Détail spec-my-B : [`06-spec-my-b.md`](06-spec-my-b.md).  
Détail spec-my-C : [`06-spec-my-c.md`](06-spec-my-c.md).  
Détail spec-my-D : [`06-spec-my-d.md`](06-spec-my-d.md).  
Détail spec-my-E : [`06-spec-my-e.md`](06-spec-my-e.md).  
Détail spec-my-F : [`06-spec-my-f.md`](06-spec-my-f.md).

**Trust replay vs engineering replay :** `internal/trust/replay/` écrit un manifeste sous `.asagiri/trust/<id>/` pour rejouer une **vérification** (`asa trust replay`). `internal/replay/` gère des **packages** sous `.asagiri/replays/<replay-id>/` pour capturer et rejouer un workflow complet (`asa replay create|run|compare|…`). Les rapports trust peuvent être copiés dans un replay package ; les deux CLI restent distinctes.

Interfaces §11.2 : `WorkflowEngine`, `TaskStore`, `WorktreeManager`, `Validator` déclarées dans `internal/workflow/interfaces.go` ; implémentations = `Service`, `sqlite.Store`, `worktree.Manager`, `validation.Runner`.

## Distribution & docs (ADR-015, ADR-016)

- Releases GitHub : repo **`LaProgrammerie/asagiri`**, binaire **`asa`**, archives `asa_{OS}_{ARCH}`.
- Homebrew : `brew install LaProgrammerie/tap/asa`.
- Docs publiques : `docs-site/` → Cloudflare Pages projet **`asagiri-docs`** ; `basePath` legacy **`/asagiri`** si `GITHUB_PAGES=true`.

## Execution Graph (spec-my-C)

Flux :

```text
product + flow → planner.Build → scheduler.Schedule → repository.SaveAll
  → graph run (dry-run ou agents) → checkpoints → trust gates → report
```

Artefacts sous `.asagiri/graphs/<graph-id>/` :

| Fichier | Rôle |
|---------|------|
| `execution-graph.yaml` / `.json` | Graphe canonique (nœuds, arêtes, checkpoints, rollback) |
| `plan.md` | Résumé humain du plan |
| `metrics.json` | Estimation coût, durée, risque, groupes parallèles |
| `events.jsonl` / `timeline.jsonl` | Journal runtime |
| `report.md` | Rapport post-exécution |

CLI : `asa plan graph`, `asa plan explain`, `asa graph run|status|resume|visualize`.  
Config : bloc `execution_graph:` (ADR-022). Durcissement phase finale : `enabled: false` bloque les commandes graph ; `gates.*` injecté dans enrichment/runner ; `--checkpoint-every node|group` persisté ; `graph resume` exige au moins un checkpoint (`ErrNoCheckpoint`) ; gates trust via `trust.Engine` (PF-C-01…05).

## Multi-Agent Coordination (spec-my-D)

Flux :

```text
execution graph → DefaultCoordinator.Coordinate → assignments (rôle, isolation, agent ref)
  → handoffs (.asagiri/handoffs/) → policies (review indépendant, security flows)
  → agent.* events → trust gates avant merge
```

Artefacts sous `.asagiri/handoffs/<handoff-id>/` :

| Fichier | Rôle |
|---------|------|
| `handoff.yaml` | Transfert structuré from/to, summary, files, constraints, confidence |

Config : bloc `coordination:` (ADR-023). **D-FULL :** `EnsureWorktree` (git worktree par assignation isolée), `NodeExecutor` branché sur `executiongraph.DefaultRunner`, `AssignmentHistory` pour `ScoringAssigner`. UX terminal : `coordination.FormatMultiAgentRuntime` (§19).

## Engineering Knowledge Graph (spec-my-E)

Flux :

```text
.asagiri/products (flows, contracts) + code + tests
  → knowledge build (extractors, incremental, staleness)
  → graph.sqlite + graph.json
  → query / impact / explain / snapshot
  → bridges investigation, trust, executiongraph, coordination
```

Artefacts sous `.asagiri/knowledge/` :

| Fichier / répertoire | Rôle |
|----------------------|------|
| `graph.sqlite` | Store SQLite (nodes, edges, index_metadata) |
| `graph.json` | Export JSON du graphe |
| `snapshots/<name>/` | Snapshot nommé (metadata + graph.json) |

CLI : `asa knowledge build|query|explain|snapshot`, `asa impact analyze`, `asa context build --from-graph`.  
Config : bloc `knowledge:` (ADR-024). Extractors dont **analytics** (`contracts/analytics.yaml` → nœuds events/metrics, liens observability). UX terminal : `knowledge.FormatKnowledgeBuild` (§23).

## Replay & Deterministic Execution (spec-my-F)

Flux :

```text
run / graph / investigation
  → replay create (capture policies, redaction)
  → .asagiri/replays/<replay-id>/
  → replay run (full | simulation | offline | strict)
  → compare / explain (divergences)
  → snapshot
```

Artefacts sous `.asagiri/replays/<replay-id>/` :

| Fichier / répertoire | Rôle |
|----------------------|------|
| `replay.yaml` | Manifest (source, commit, agents, policies, artefact list) |
| `graph/` | Copie execution-graph, metrics, events |
| `trust/` | Rapports / manifests trust capturés |
| `investigations/` | replay-pack, rapports investigation |
| `runtime/`, `prompts/`, `outputs/` | Selon policies |
| `snapshots/<name>/` | Snapshot nommé du package |

CLI : `asa replay create|run|compare|explain|snapshot`.  
Config : bloc `replay:` (`capture_*`, `redact_secrets`, `offline_mode_default`, `compress_threshold_bytes`). UX terminal : `replay.WriteReplay*` (§26).

## Experience Platform (spec-ui, ADR-027)

### Principe

```text
CLI Cobra (asa <cmd>)          internal/ui (asa → Mission Control)
        │                                │
        └──────── ui/bus ────────────────┘
                      │
    workflow · pipeline · runtime · trust · knowledge
    replay · executiongraph · investigation · coordination
```

- **`internal/ui`** : navigation, layout, thème, widgets — **zéro** logique trust/graph/workflow dans les écrans.
- **`internal/tui`** : rendu ligne de commande du pipeline V3 ; reste la cible de `--rich` / `--plain` / `--json` hors application interactive.
- Toute action TUI doit exposer `CLIEquivalent()` (spec §3.1).

### Entrée

| Invocation | Comportement |
|------------|--------------|
| `asa` (TTY) | Lance `ui/app.Run` → écran `ui.default_screen` (défaut `mission`) |
| `asa` (non-TTY) | Aide Cobra |
| `asa work "…"` | Inchangé — CLI directe |

### Config `ui:` (specv3 + spec-ui §29)

Champs specv3 : `mode`, `live_logs`, `progress_bars`, `compact`.  
Champs Experience Platform : `default_screen`, `theme`, `mouse`, `animations`, `refresh_interval_ms`, `compact_threshold`, `show_cli_equivalents`, `confirm_destructive_actions`.

### Layout Experience Platform

`single` · `split-horizontal` · `split-vertical` — focus par panneau, seuil `compact_threshold` colonnes, resize souris basique (wheel/drag divider).

Polish livré : mode no-animation, aide accessibilité (`?`), responsive compact/wide/ultra-wide sur dashboard.

### Dépendances Charm

`bubbletea`, `lipgloss` (déjà), `bubbles`, `huh`, `glamour` — versions figées dans ADR-027.

Détail : [`docs/decisions/027-experience-platform-ui.md`](../decisions/027-experience-platform-ui.md).

## Limites connues

- Commandes §6.2 restantes : `bench`, `search`, `export`.
- **Experience Platform** : tranche livrée lots 1–6 ; conserver l’invariant « CLI-first + équivalent CLI visible ».
- **Mémoire runtime** : embedder pluggable (ADR-025) ; **`asa memory doctor`** vérifie Ollama, dimensions, orphelins.
- **Index RAG** : `asa index` persiste des embeddings (`runtime.memory.embedder`) ; `index search` cosinus par défaut, `--keyword` pour LIKE.
- **`asa resume --execute`** : enchaîne les steps restants (agents réels hors `--dry-run` global) ; sans `--execute`, diagnostic du prochain step.
- **Execution graph** : inférence dépendances V2 (events, projection architecture, mémoire historique) — `dependency_v2.go`.
- Agents externes requis hors mode dry-run.

## Extension

- Nouvelle commande : `internal/cli/` + `03-standards.md`.
- Migration DB : `internal/store/sqlite/migrations/` + ADR si breaking.
- Décisions : `05-decisions.md`.
