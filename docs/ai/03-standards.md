# Standards et conventions (projet)

## Langage et style

- **Langage :** Go (version dans `go.mod`).
- **Formatage :** `gofmt` / `go fmt`.
- **Analyse statique :** `go vet`, **golangci-lint** (`.golangci.yml`) — toolchain Go alignée sur `go.mod`.
- **Erreurs :** retours explicites ; pas de `panic` aux frontières CLI.
- **Sécurité agents :** `exec.Command` avec binaire + args (jamais `sh -c`).

## Arborescence

```
go.mod
Makefile
spec.md
.agentflow/config.yaml.example
application/cmd/agentflow/
application/internal/{cli,config,bootstrap,agent,worktree,workflow,validation,policy,rag,spec,plan,report,store/sqlite,version}
application/pkg/agentflow/
bin/agentflow
```

## CLI AgentFlow (V1)

| Commande | Statut |
|----------|--------|
| `init`, `doctor`, `version` | Implémenté |
| `spec <feature> --agent <name>` | Implémenté |
| `plan <feature>` | Implémenté |
| `enrich <feature> [--task id] --agent <name>` | Implémenté |
| `dev <feature> [--task id] --agent <name>` | Implémenté |
| `verify <feature> [--task id]` | Implémenté |
| `review <feature> [--task id] --agent <name>` | Implémenté |
| `status`, `resume <run-id>`, `report <run-id>` | Implémenté |
| `clean [--merged] [--failed]`, `pr <feature>` | Implémenté |
| `agentflow index` | Index RAG local (§10.3) |
| `bench`, `search`, `graph`, `export` | Hors scope |

**Flags communs :** `--force` sur `enrich`, `dev`, `verify`, `review`, `resume` ; `resume --execute` (dry-run).

**Dry-run :** `--dry-run` ou `AGENTFLOW_DRY_RUN=1` — simule agents et validations (CI/tests).

**Config (§7) :** `validation.commands`, `policies`, agents (`timeout`, `default_model`, `endpoint`/`model` pour Ollama). Défauts Go injectés si `go.mod` absent de surcharge.

**Tâches (§8) :** statuts canoniques + fichiers `.agentflow/tasks/<feature>/<id>.yaml`.

**Logs agent (§9) :** `.agentflow/logs/<task-id>/context.json`, `result.json`.

## Chemins (relatifs à la racine Git)

| Artefact | Défaut |
|----------|--------|
| Config | `.agentflow/config.yaml` |
| État SQLite | `.agentflow/state.sqlite` |
| Worktrees | `.agentflow/worktrees/` |
| Runs / tasks / logs | `.agentflow/runs/`, `tasks/`, `logs/` |
| Specs Kiro | `.kiro/specs` |
| Spec active / handoff | `docs/ai/active/current-spec.md`, `handoff.md` |

## Commandes Makefile / Go

| Action | Commande |
|--------|----------|
| Build CLI | `make build` |
| Init / doctor | `./bin/agentflow init` puis `doctor` |
| Tests | `make test` ou `go test -race ./...` |
| Lint | `make lint` (nécessite golangci-lint + Go ≥ version module) |
| Dev Docker | `make dev` |

## Tests

- **Unitaires :** `*_test.go` à côté du code.
- **Intégration CLI :** `internal/cli/integration_test.go` (repo git temporaire, `--dry-run`).
- **Git requis** pour init/doctor/worktree.

## Sécurité et données

- Pas de secrets dans le dépôt ; `state.sqlite` et `worktrees/` gitignored.
- Pas d’injection shell sur les agents ni les validations.
