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
    skills/
    analysis/              # spec-my-A : graphes structuraux
    investigation/         # spec-my-A : pipeline investigation
    embedutil/
    trust/                 # spec-my-B : moteur confiance & vérification
    trust/checks/
    trust/confidence/
    trust/replay/
    trust/safeid/
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
| `internal/workflow` | `PlanFeature`, `DevFeature`, `VerifyFeature`, etc. |
| `internal/spec` | Lecture spec Kiro ou fallback |
| `internal/plan` | Parse `tasks.md`, export JSON tâches |
| `internal/report` | Rapports de run |
| `internal/cli` | Surface utilisateur + `--dry-run` global |
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

État persistant : `runs.steps_json`, tâches `tasks.payload_json` + fichiers `.asagiri/tasks/<feature>/*.json`.

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
| `internal/memory/` | Scoped memory, embeddings hash, retrieval |
| `internal/skills/` | `.asagiri/skills/*.yaml` |
| `internal/analysis/` | `asa analysis build` → graphs.json |
| `internal/investigation/` | Rapport, graph, impact, context-pack |
| `internal/trust/` | `TrustEngine`, rapports `.asagiri/trust/<id>/`, gates, replay |
| `internal/trust/checks/` | Runners de vérification (contracts, flows, blast-radius, …) |
| `internal/trust/confidence/` | Agrégation 6 dimensions §7 spec-my-B |

Détail spec-my-A : [`06-spec-my-a.md`](06-spec-my-a.md).  
Détail spec-my-B : [`06-spec-my-b.md`](06-spec-my-b.md).

Interfaces §11.2 : `WorkflowEngine`, `TaskStore`, `WorktreeManager`, `Validator` déclarées dans `internal/workflow/interfaces.go` ; implémentations = `Service`, `sqlite.Store`, `worktree.Manager`, `validation.Runner`.

## Distribution & docs (ADR-015, ADR-016)

- Releases GitHub : repo **`LaProgrammerie/asagiri`**, binaire **`asa`**, archives `asa_{OS}_{ARCH}`.
- Homebrew : `brew install LaProgrammerie/tap/asa`.
- Docs publiques : `docs-site/` → Cloudflare Pages projet **`asagiri-docs`** ; `basePath` legacy **`/asagiri`** si `GITHUB_PAGES=true`.

## Limites connues

- Commandes §6.2 restantes : `bench`, `search`, `graph`, `export`.
- RAG : recherche textuelle sur chunks (pas d’embeddings Ollama en V1).
- `resume --execute` : dry-run uniquement.
- Agents externes requis hors mode dry-run.

## Extension

- Nouvelle commande : `internal/cli/` + `03-standards.md`.
- Migration DB : `internal/store/sqlite/migrations/` + ADR si breaking.
- Décisions : `05-decisions.md`.
