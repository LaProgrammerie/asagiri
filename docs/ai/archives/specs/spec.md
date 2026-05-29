> **Historique** — spec rédigée sous le nom **AgentFlow** ; produit actuel : **Asagiri** / CLI **`asa`** (`spec-rename.md`).

Spec produit & technique — AgentFlow (legacy)

1. Vision

AgentFlow est un orchestrateur CLI local, écrit en Go, destiné à industrialiser les workflows de développement assistés par agents.

L’objectif n’est pas de créer un agent autonome généraliste, mais un moteur déterministe capable de :

* transformer une spec en tâches exécutables ;
* lancer des agents spécialisés comme Kiro CLI, Cursor Agent, Codex, Claude Code ou Ollama ;
* isoler les modifications dans des git worktree ;
* tracer chaque exécution ;
* vérifier automatiquement la qualité ;
* produire des handoffs exploitables ;
* permettre la reprise après échec.

Positionnement :

Un orchestrateur local, reproductible et traçable pour workflows de développement agentique : specs, worktrees, agents, validations, reviews.

⸻

2. Principes directeurs

2.1 Déterminisme avant autonomie

AgentFlow ne doit pas laisser un agent décider librement de tout le workflow.

Le workflow est défini par :

* une configuration versionnée ;
* des tâches atomiques ;
* des agents explicitement déclarés ;
* des commandes de validation ;
* des règles de scope fichiers ;
* un état persistant.

2.2 Repo comme source de vérité

Toutes les données critiques doivent rester dans le dépôt :

.agentflow/
  config.yaml
  runs/
  tasks/
  logs/
  state.sqlite

Les specs Kiro, règles Cursor, prompts, décisions et logs doivent être versionnables ou exportables.

2.3 Agents interchangeables

AgentFlow ne doit pas dépendre fortement d’un seul fournisseur.

Chaque agent est encapsulé derrière une interface stable :

type Agent interface {
    Name() string
    Capabilities() Capabilities
    Run(ctx context.Context, req RunRequest) (RunResult, error)
}

2.4 Validation externe obligatoire

Un agent ne valide jamais seul son propre travail.

La vérité finale vient de :

* tests ;
* analyse statique ;
* linters ;
* compilation ;
* diff review ;
* éventuellement CI distante.

2.5 Modèles locaux comme middleware

Ollama ne remplace pas les modèles premium sur les tâches critiques.

Il sert principalement à :

* indexer le repo ;
* router les tâches ;
* résumer specs, logs et diffs ;
* faire une pré-review low-cost ;
* préparer du contexte avant exécution ;
* générer des handoffs structurés.

⸻

3. Objectifs V1

3.1 Objectifs fonctionnels

La V1 doit permettre de :

1. initialiser AgentFlow dans un repo ;
2. lire une spec Kiro existante ;
3. générer ou normaliser une liste de tâches ;
4. enrichir les tâches avec Ollama ;
5. créer un git worktree par tâche ;
6. lancer Cursor Agent pour implémenter ;
7. lancer des validations locales ;
8. lancer une review indépendante via Codex ou Claude Code ;
9. produire un rapport de run ;
10. préparer une PR ou un patch.

3.2 Non-objectifs V1

Ne pas faire en V1 :

* UI desktop ;
* dashboard web ;
* orchestration distribuée ;
* queue distante ;
* multi-user ;
* marketplace de plugins ;
* support exhaustif de tous les agents ;
* autonomie complète sans validation humaine ;
* édition automatique de secrets ;
* déploiement production.

⸻

4. Cas d’usage principaux

4.1 Spécifier avec Kiro, développer avec Cursor

agentflow spec feature-x --agent kiro
agentflow plan feature-x
agentflow dev feature-x --agent cursor
agentflow verify feature-x
agentflow review feature-x --agent codex

4.2 Enrichir une tâche avec Ollama

agentflow enrich feature-x --task task-003 --agent ollama

Sortie attendue :

{
  "task_id": "task-003",
  "type": "backend_refactor",
  "risk": "medium",
  "recommended_agent": "cursor",
  "files_scope": ["src/Billing", "tests/Billing"],
  "validation_commands": ["composer test", "composer phpstan"]
}

4.3 Lancer une tâche isolée

agentflow dev feature-x --task task-003 --agent cursor

AgentFlow doit :

* créer un worktree dédié ;
* préparer le prompt ;
* injecter les contraintes ;
* lancer l’agent ;
* capturer stdout/stderr ;
* enregistrer le diff ;
* lancer les validations ;
* produire un résumé.

4.4 Reprendre après échec

agentflow status
agentflow resume run-2026-05-17-001

AgentFlow doit reprendre à partir de l’état persistant, sans relancer inutilement les étapes déjà validées.

⸻

5. Workflow cible

┌────────────┐
│ Kiro CLI   │
│ spec       │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ AgentFlow  │
│ normalize  │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ Ollama     │
│ enrich/RAG │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ Worktree   │
│ per task   │
└─────┬──────┘
      │
      ▼
┌──────────────┐
│ Cursor Agent │
│ implement    │
└─────┬────────┘
      │
      ▼
┌────────────┐
│ Verify     │
│ tests      │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ Codex /    │
│ Claude     │
│ review     │
└─────┬──────┘
      │
      ▼
┌────────────┐
│ Report / PR│
└────────────┘

⸻

6. CLI attendue

6.1 Commandes principales

agentflow init
agentflow doctor
agentflow spec <feature> --agent kiro
agentflow plan <feature>
agentflow enrich <feature> [--task <id>] --agent ollama
agentflow dev <feature> [--task <id>] --agent cursor
agentflow verify <feature> [--task <id>]
agentflow review <feature> [--task <id>] --agent codex
agentflow status
agentflow resume <run-id>
agentflow report <run-id>
agentflow clean [--merged] [--failed]

6.2 Commandes futures

agentflow pr <feature>
agentflow bench <feature>
agentflow index
agentflow search <query>
agentflow graph <feature>
agentflow export <run-id> --format markdown|json

⸻

7. Configuration

7.1 Fichier .agentflow/config.yaml

project:
  name: my-project
  default_branch: main
specs:
  kiro_path: .kiro/specs
  active_spec_path: docs/ai/active/current-spec.md
  handoff_path: docs/ai/active/handoff.md
state:
  backend: sqlite
  path: .agentflow/state.sqlite
worktrees:
  base_path: .agentflow/worktrees
  branch_prefix: agentflow
  cleanup_policy: keep_failed
agents:
  kiro:
    command: kiro
    args: ["--cli"]
  cursor:
    command: cursor-agent
    default_model: auto
    timeout: 3600
  codex:
    command: codex
    timeout: 3600
  claude:
    command: claude
    timeout: 3600
  ollama:
    endpoint: http://localhost:11434
    model: qwen2.5-coder:14b
    embedding_model: nomic-embed-text
    timeout: 300
validation:
  commands:
    - name: tests
      command: composer test
      required: true
    - name: static-analysis
      command: composer phpstan
      required: true
    - name: lint
      command: composer lint
      required: false
policies:
  require_clean_git: true
  forbid_untracked_secret_files: true
  max_files_changed_per_task: 20
  allow_network: false
  require_human_approval_for:
    - database_migration
    - security_sensitive_change
    - dependency_upgrade

⸻

8. Modèle de tâche

8.1 Structure canonique

id: task-003
title: Refactor billing calculation service
feature: feature-x
status: pending
risk: medium
type: backend_refactor
source:
  spec: .kiro/specs/feature-x/tasks.md
  section: "3"
scope:
  allowed_paths:
    - src/Billing/**
    - tests/Billing/**
  forbidden_paths:
    - config/secrets/**
    - migrations/**
acceptance:
  - Existing billing tests pass
  - New edge case tests added
  - Public API unchanged
validation:
  commands:
    - composer test -- tests/Billing
    - composer phpstan
agents:
  implementer: cursor
  reviewer: codex
  enricher: ollama
metadata:
  created_at: "2026-05-17T12:00:00+02:00"
  updated_at: "2026-05-17T12:00:00+02:00"

8.2 États possibles

pending
planned
enriched
running
implemented
verify_failed
verified
review_failed
reviewed
ready_for_pr
merged
aborted

⸻

9. Contrat agent

9.1 Input standardisé

Chaque agent reçoit un contexte structuré :

{
  "run_id": "run-2026-05-17-001",
  "task_id": "task-003",
  "objective": "Refactor billing calculation service",
  "allowed_paths": ["src/Billing/**", "tests/Billing/**"],
  "forbidden_paths": ["config/secrets/**"],
  "acceptance_criteria": [
    "Existing billing tests pass",
    "New edge case tests added",
    "Public API unchanged"
  ],
  "validation_commands": [
    "composer test -- tests/Billing",
    "composer phpstan"
  ],
  "context_files": [
    "src/Billing/Calculator.php",
    "tests/Billing/CalculatorTest.php"
  ],
  "output_format": "agentflow-v1"
}

9.2 Output standardisé

{
  "status": "completed",
  "summary": "Refactored billing calculator and added edge case tests.",
  "changed_files": [
    "src/Billing/Calculator.php",
    "tests/Billing/CalculatorTest.php"
  ],
  "commands_run": [
    {
      "command": "composer test -- tests/Billing",
      "exit_code": 0
    }
  ],
  "risks": [
    "Tax rounding behavior should be checked with product owner."
  ],
  "requires_human_review": true
}

⸻

10. Ollama dans l’architecture

10.1 Rôles autorisés

Ollama peut être utilisé pour :

* classifier les tâches ;
* détecter les risques ;
* sélectionner les fichiers de contexte ;
* générer un résumé de diff ;
* produire un handoff ;
* faire une pré-review ;
* interroger un index local du repo.

10.2 Rôles interdits par défaut

Ollama ne doit pas, par défaut :

* modifier directement le code critique ;
* valider une tâche seul ;
* décider d’une migration DB ;
* changer des dépendances ;
* modifier des secrets ;
* publier une PR sans validation externe.

10.3 RAG local

Index local recommandé :

.agentflow/index/
  chunks.sqlite
  embeddings.sqlite

Sources indexables :

* src/
* tests/
* docs/
* .kiro/
* .cursor/
* README.md
* composer.json, go.mod, package.json

Sources exclues :

* .git/
* vendor/
* node_modules/
* fichiers secrets ;
* dumps ;
* fichiers binaires lourds.

⸻

11. Architecture Go

11.1 Structure projet

agentflow/
  cmd/agentflow/
    main.go
  internal/
    app/
      app.go
      config.go
    cli/
      init.go
      doctor.go
      spec.go
      plan.go
      enrich.go
      dev.go
      verify.go
      review.go
      status.go
      resume.go
      report.go
      clean.go
    workflow/
      engine.go
      run.go
      task.go
      state_machine.go
    agents/
      agent.go
      registry.go
      process_agent.go
      kiro/
      cursor/
      codex/
      claude/
      ollama/
    git/
      repository.go
      worktree.go
      diff.go
    validation/
      runner.go
      command.go
      result.go
    state/
      store.go
      sqlite_store.go
      migrations/
    rag/
      chunker.go
      indexer.go
      retriever.go
      embeddings.go
    policy/
      policy.go
      scope.go
      secrets.go
    logs/
      writer.go
      tail.go
    report/
      markdown.go
      json.go
  pkg/
    agentflow/
      types.go
  .agentflow.example/
    config.yaml

11.2 Interfaces principales

type WorkflowEngine interface {
    Run(ctx context.Context, req WorkflowRequest) (WorkflowResult, error)
    Resume(ctx context.Context, runID string) (WorkflowResult, error)
}
type TaskStore interface {
    SaveRun(ctx context.Context, run Run) error
    GetRun(ctx context.Context, id string) (Run, error)
    SaveTask(ctx context.Context, task Task) error
    ListTasks(ctx context.Context, feature string) ([]Task, error)
}
type WorktreeManager interface {
    Create(ctx context.Context, task Task) (Worktree, error)
    Diff(ctx context.Context, wt Worktree) (Diff, error)
    Cleanup(ctx context.Context, wt Worktree) error
}
type Validator interface {
    Run(ctx context.Context, wt Worktree, commands []ValidationCommand) ([]ValidationResult, error)
}

⸻

12. State machine

pending
  ↓ plan
enriched
  ↓ create_worktree
running
  ↓ agent_success
implemented
  ↓ verify_success
verified
  ↓ review_success
reviewed
  ↓ report
ready_for_pr

Échecs :

running -> failed
implemented -> verify_failed
verified -> review_failed
any -> aborted

Règles :

* une étape réussie ne doit pas être relancée sans --force ;
* une étape échouée doit conserver ses logs ;
* un run doit être inspectable après crash ;
* le retry doit être explicite.

⸻