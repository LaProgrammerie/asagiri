# AgentFlow (hyper-fast-builder)

Orchestrateur CLI local en Go pour workflows de développement agentique (specs, worktrees, agents, validations).

Spécification produit : [`spec.md`](spec.md).  
Contexte agents et canon : [`AGENTS.md`](AGENTS.md), [`docs/ai/`](docs/ai/).

## Prérequis

- Go 1.25+ (voir `go.mod`)
- `make`, `git`
- Optionnel : Docker pour l’environnement conteneurisé de dev
- Agents externes (kiro, cursor, codex, ollama, claude) configurés dans `.agentflow/config.yaml` — **non requis** avec `--dry-run`

## Démarrage rapide

```bash
go mod download
make build
./bin/agentflow init
./bin/agentflow doctor
```

## Commandes V1

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
