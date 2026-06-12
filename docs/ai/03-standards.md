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
docs/ai/archives/specs/
.asagiri/config.yaml.example
application/cmd/asa/
application/internal/{cli,config,bootstrap,env,agent,worktree,workflow,validation,policy,rag,spec,plan,report,store/sqlite,version}
application/pkg/asagiri/
bin/asa
```

## CLI Asagiri (`asa`)

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
| `asa index` | Index RAG local (§10.3) |
| `work`, `continue`, `next`, `estimate`, `inbox`, `sync` | Intent layer (specv2) |
| `bench`, `search`, `graph`, `export` | Hors scope |

**Flags communs :** `--force` sur `enrich`, `dev`, `verify`, `review`, `resume` ; `resume --execute` (dry-run).

**Dry-run :** `--dry-run` ou **`ASA_DRY_RUN=1`** (fallback déprécié `AGENTFLOW_DRY_RUN=1` avec warning).

**Config (§7) :** `validation.commands`, `policies`, agents (`timeout`, `default_model`, `endpoint`/`model` pour Ollama). Défauts Go injectés si `go.mod` absent de surcharge.

**Tâches (§8) :** statuts canoniques + fichiers `.asagiri/tasks/<feature>/<id>.yaml`.

**Logs agent (§9) :** `.asagiri/logs/<task-id>/context.json`, `result.json`.

## Chemins (relatifs à la racine Git)

| Artefact | Défaut |
|----------|--------|
| Config | `.asagiri/config.yaml` |
| État SQLite | `.asagiri/state.sqlite` |
| Worktrees | `.asagiri/worktrees/` |
| Runs / tasks / logs | `.asagiri/runs/`, `tasks/`, `logs/` |
| Specs Kiro | `.kiro/specs` |
| Spec active / handoff | `docs/ai/active/current-spec.md`, `handoff.md` |

## Commandes Makefile / Go

| Action | Commande |
|--------|----------|
| Build CLI | `make build` |
| Init / doctor | `./bin/asa init` puis `doctor` |
| Tests | `make test` ou `go test -race ./...` |
| Lint | `make lint` (nécessite golangci-lint pinné — voir ci-dessous) |
| Dev Docker | `make dev` |
| Release check | `make release-check` |

## Quality_Gate (readiness production)

Le **Quality_Gate** est l'ensemble des quatre contrôles suivants ; la readiness
production est atteinte **si et seulement si** les quatre terminent avec un code
de sortie égal à `0` :

```text
Quality_Gate = make build  ∧  go test ./...  ∧  go vet ./...  ∧  golangci-lint run
               (tous exit 0)
```

### Installation de golangci-lint (binaire pinné, reproductible)

`golangci-lint` doit être bâti avec une toolchain Go **≥ la cible `go.mod`**
(`go 1.25.0`), sinon il refuse de tourner. On installe donc le **binaire officiel
pinné** (indépendant du Go local), version **`v2.12.2`** (ligne v2 courante) :

```bash
# binaire dans $(go env GOPATH)/bin
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" v2.12.2
golangci-lint --version   # doit afficher "built with go1.25" ou supérieur
```

> **Éviter `go install …/golangci-lint`** : cette voie compile l'outil avec le Go
> **local**, ce qui ne garantit pas la reproductibilité ni l'alignement sur la
> cible `go.mod`. On privilégie le binaire officiel pinné ci-dessus.

Assure-toi que `$(go env GOPATH)/bin` est dans le `PATH`. Une fois installé,
`make lint` (ou `golangci-lint run`) s'exécute sur le code Go sous
`application/` via `.golangci.yml` (schéma v2, `run.go: "1.25"`).

## Tests

- **Unitaires :** `*_test.go` à côté du code.
- **Intégration CLI :** `internal/cli/integration_test.go` (repo git temporaire, `--dry-run`).
- **Git requis** pour init/doctor/worktree.

## Sécurité et données

- Pas de secrets dans le dépôt ; `state.sqlite` et `worktrees/` gitignored.
- Pas d’injection shell sur les agents ni les validations.

## Branding public

- Produit : **Asagiri** ; CLI : **`asa`** ; pas de `agentflow` / `AgentFlow` dans help, docs-site, README, exemples release (voir `spec-rename.md`).
