# AgentFlow (hyper-fast-builder)

Orchestrateur CLI **local-first** et **cost-aware** pour workflows de développement agentique : specs, worktrees, agents, validations — de façon déterministe et observable.

**Philosophie :** les scans et l’investigation restent locaux ; les modèles cloud sont réservés aux étapes à forte valeur (spec, dev, review). AgentFlow estime tokens/coût **avant** exécution et explique le choix des tiers dans le terminal.

| Spec | Rôle |
|------|------|
| [`spec-postv123.md`](spec-postv123.md) | Consolidation & OSS (mission courante) |
| [`specv3.md`](specv3.md) | Cost / perf / MCP |
| [`specv2.md`](specv2.md) | Intent layer |
| [`spec.md`](spec.md) | Primitives V1 |

Docs agents : [`AGENTS.md`](AGENTS.md), [`docs/ai/`](docs/ai/), audits : [`docs/consolidation/`](docs/consolidation/).  
Roadmap : [`ROADMAP.md`](ROADMAP.md) · Contribuer : [`CONTRIBUTING.md`](CONTRIBUTING.md)

## Prérequis

- Go 1.25+ (voir `go.mod`)
- `make`, `git`
- Optionnel : Docker pour l’environnement conteneurisé de dev
- Agents externes (kiro, cursor, codex, ollama, claude) configurés dans `.agentflow/config.yaml` — **non requis** avec `--dry-run`

## Démarrage rapide

```bash
go mod download
make build
export AGENTFLOW_DRY_RUN=1   # optionnel : sans binaires kiro/cursor
./bin/agentflow init
./bin/agentflow doctor
./bin/agentflow work "développe ma-feature" --dry-run --plan-only --yes
```

Benchmark dry-run : `make benchmark` · Exemple détaillé : [`examples/quickstart/`](examples/quickstart/).

## Commandes intention (specv2)

| Commande | Description |
|----------|-------------|
| `agentflow work "<instruction>"` | Résout l’intention, affiche/exécute un plan de primitives |
| `agentflow continue [--feature] [--run]` | Reprend le travail le plus pertinent |
| `agentflow next [--feature] [--execute]` | Prochaine action recommandée |
| `agentflow inbox [--source notion\|local]` | Liste les specs candidates |
| `agentflow sync notion\|all [--page URL] [--feature] [--force]` | Import Notion → `.agentflow/specs/` |

Options `work` : `--agent`, `--reviewer`, `--plan-only`, `--yes`, `--max-tasks`, `--stop-after`, `--no-review`, `--source`, et (V3) `--estimate-only`, `--budget`, `--prefer-local`, `--max-input-tokens`, `--max-output-tokens`, `--allow-over-budget`, `--no-context-reduction`, etc.

## Commandes cost/perf (specv3)

| Commande | Description |
|----------|-------------|
| `agentflow estimate <feature> [--task]` | Estimation tokens/coût/temps sans exécuter |
| `agentflow investigate <feature> [--task]` | Investigation locale (grep, candidats, tests) |
| `agentflow context <feature> --show` / `--optimize` | Contexte pack + économies |
| `agentflow cost report [--since 7d]` | Historique coûts (SQLite) |
| `agentflow cost models` | Pricing et profils modèles configurés |
| `agentflow inspect symbol\|tests\|diff` | Inspection ciblée |
| `agentflow mcp serve` | MCP stdio (nécessite `mcp.enabled: true`) |

Configurer `models`, `budgets`, `pricing`, `token_estimation`, `routing`, `ui` dans `.agentflow/config.yaml` (voir `config.yaml.example`).

```bash
./bin/agentflow estimate billing-v2 --task task-003
./bin/agentflow work "développe billing-v2" --estimate-only --budget 0.50
./bin/agentflow investigate billing-v2
```

```bash
./bin/agentflow work "développe agentflow-test" --dry-run --plan-only
./bin/agentflow continue
./bin/agentflow inbox --source local
```

### Notion

1. Activer dans `.agentflow/config.yaml` : `sources.notion.enabled: true`
2. Exporter le token : `export NOTION_TOKEN=secret_…` (jamais loggé par AgentFlow)
3. Optionnel : `sources.notion.specs_database_id` pour l’inbox database
4. Sync : `./bin/agentflow sync notion --page 'https://notion.so/…'`

Test d’intégration Notion (opt-in) : `NOTION_TOKEN` + `NOTION_TEST_PAGE_ID` → `go test -tags=integration ./application/internal/source/notion/...`

## Commandes V1 (primitives)

| Commande | Description |
|----------|-------------|
| `agentflow init` | Bootstrap `.agentflow/` + SQLite |
| `agentflow doctor` | Contrôles Git, config, schéma |
| `agentflow spec <feature> --agent kiro` | Phase spec via agent |
| `agentflow plan <feature>` | Normalise tâches (Kiro ou `current-spec.md`) |
| `agentflow enrich <feature> [--task id] --agent ollama` | Enrichit le payload tâche |
| `agentflow dev <feature> [--task id] --agent cursor` | Worktree + implémentation |
| `agentflow verify <feature> [--task id]` | Validations (`go test`, `go vet`, `make lint`) |
| `agentflow review <feature> --agent codex` | Review indépendante |
| `agentflow status` | Liste des runs |
| `agentflow index` | Index RAG local (`application/`, `docs/`, `.kiro/`, `spec.md`, `go.mod`) |
| `agentflow resume <run-id> [--execute]` | Prochain step ; `--execute` en dry-run |
| `agentflow report <run-id>` | Rapport markdown + JSON |
| `agentflow clean [--merged] [--failed]` | Nettoie worktrees |
| `agentflow pr <feature>` | Diff + checklist PR |

**Dry-run (tests / CI sans agents) :**

```bash
./bin/agentflow plan agentflow-test --dry-run
./bin/agentflow dev agentflow-test --dry-run
```

## Makefile

| Cible | Description |
|-------|-------------|
| `make build` | Compile `bin/agentflow` |
| `make test` | `go test ./...` |
| `make lint` | `golangci-lint run` |
| `make dev` | Conteneur dev Docker |

## Configuration

Copier `.agentflow/config.yaml.example` → `.agentflow/config.yaml` (fait automatiquement par `init` si l’example existe).

Sections **validation** (commandes nommées), **policies** (git propre, secrets, plafond fichiers) et agents étendus : voir `spec.md` §7.1. Le template Go préremplit `go test`, `go vet`, `golangci-lint run` lorsque `go.mod` est présent.
