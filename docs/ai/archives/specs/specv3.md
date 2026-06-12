Spec V3 — AgentFlow Cost, Performance & Token Optimization

1. Objectif

La V3 d’AgentFlow ajoute une couche d’optimisation systématique des coûts, tokens, temps de traitement et performances d’exécution.

Objectif principal :

Faire localement tout ce qui peut être fait localement, mesurer le coût estimé avant exécution, limiter l’usage des modèles payants aux opérations à forte valeur ajoutée, et rendre chaque workflow observable en temps réel dans le terminal.

Cette V3 ne remplace pas les workflows existants. Elle ajoute :

* estimation tokens/coût/temps avant exécution ;
* budgets par tâche, run, feature et agent ;
* préparation locale du contexte ;
* extraction locale de fichiers ;
* investigation locale via commandes ou MCP ;
* compression intelligente du contexte ;
* routing local/cloud basé sur coût, risque et complexité ;
* UI terminal riche avec progression live.

⸻

2. Principes directeurs

2.1 Local-first

AgentFlow doit exécuter localement tout ce qui ne nécessite pas explicitement un modèle cloud :

* parsing de fichiers ;
* recherche dans le repo ;
* extraction de symboles ;
* diff analysis ;
* grep/ripgrep ;
* tree-sitter ;
* analyse dépendances ;
* estimation tokens ;
* résumés intermédiaires via modèle local ;
* enrichissement via Ollama local ;
* indexation RAG ;
* benchmark local ;
* génération de rapports.

2.2 Cloud only when justified

Les modèles cloud doivent être utilisés uniquement si :

* la tâche dépasse la capacité locale ;
* le risque est élevé ;
* le coût estimé est acceptable ;
* le budget configuré le permet ;
* le gain qualité attendu justifie l’usage.

2.3 Budget visible avant action

Avant toute commande coûteuse, AgentFlow doit produire une estimation :

Estimated execution
───────────────────
Context size:     ~42k tokens
Expected output:  ~6k tokens
Model:            gemini-3-flash-preview
Estimated cost:   ~€0.08
Estimated time:   ~2m30s
Risk:             medium
Budget status:    OK

2.4 Mesure réelle après action

Après exécution, AgentFlow doit stocker :

* tokens estimés ;
* tokens réels si disponibles ;
* coût estimé ;
* coût réel si disponible ;
* durée réelle ;
* modèle utilisé ;
* taille du contexte ;
* fichiers lus ;
* fichiers modifiés ;
* validation outcome.

2.5 Séparation investigation / raisonnement

Les agents ne doivent pas consommer des tokens pour faire des tâches triviales comme lire 200 fichiers ou chercher une occurrence.

Pattern obligatoire :

Local investigation
  ↓
context reduction
  ↓
model reasoning
  ↓
local verification

⸻

3. Nouveaux concepts

3.1 Cost profile

Un profil de coût décrit le comportement attendu d’un modèle ou agent.

models:
  ollama_local_qwen:
    provider: ollama
    class: local
    input_cost_per_1m_tokens: 0
    output_cost_per_1m_tokens: 0
    typical_latency_ms_per_1k_tokens: 120
    max_context_tokens: 32000
    usage:
      - summarize
      - classify
      - pre_review
      - context_selection
  gemini_flash:
    provider: ollama_cloud
    class: cloud_fast
    model: gemini-3-flash-preview
    input_cost_per_1m_tokens: configurable
    output_cost_per_1m_tokens: configurable
    typical_latency_ms_per_1k_tokens: configurable
    max_context_tokens: 1000000
    usage:
      - implementation
      - intermediate_review
  claude_opus:
    provider: anthropic
    class: cloud_heavy
    input_cost_per_1m_tokens: configurable
    output_cost_per_1m_tokens: configurable
    typical_latency_ms_per_1k_tokens: configurable
    usage:
      - architecture_review
      - security_review
      - complex_refactor

Les prix doivent être configurables, car ils changent dans le temps.

⸻

3.2 Budget

Budget configurable par :

* run ;
* feature ;
* task ;
* agent ;
* journée ;
* projet.

budgets:
  default_currency: EUR
  per_run:
    max_estimated_cost: 1.00
    max_estimated_tokens: 500000
    require_confirmation_above_cost: 0.20
  per_task:
    max_estimated_cost: 0.30
    max_estimated_tokens: 150000
  daily:
    max_estimated_cost: 10.00
  policies:
    block_when_over_budget: true
    allow_override_with_flag: true
    override_flag: "--allow-over-budget"

⸻

3.3 Execution estimate

Avant exécution, AgentFlow produit un ExecutionEstimate.

type ExecutionEstimate struct {
    RunID                string
    Feature              string
    TaskID               string
    PlannedSteps         []EstimatedStep
    TotalInputTokens     int
    TotalOutputTokens    int
    TotalTokens          int
    EstimatedCost        Money
    EstimatedDuration    time.Duration
    Confidence           float64
    BudgetStatus         BudgetStatus
    Warnings             []string
}
type EstimatedStep struct {
    Name              string
    Agent             string
    Model             string
    Local             bool
    InputTokens       int
    OutputTokens      int
    EstimatedCost     Money
    EstimatedDuration time.Duration
    Reason            string
}

⸻

4. Nouveaux modules Go

internal/
  cost/
    estimator.go
    pricing.go
    budget.go
    token_counter.go
    duration_model.go
  contextopt/
    collector.go
    reducer.go
    compressor.go
    relevance.go
    packer.go
  investigation/
    grep.go
    symbols.go
    ast.go
    dependencies.go
    filesystem.go
    shell.go
  telemetry/
    metrics.go
    run_metrics.go
    cost_metrics.go
  tui/
    progress.go
    spinner.go
    timeline.go
    live_logs.go
    dashboard.go

⸻

5. Token estimation locale

5.1 Objectif

AgentFlow doit estimer localement le volume de tokens avant d’envoyer du contexte à un modèle.

5.2 Sources à mesurer

* prompt système ;
* instruction utilisateur ;
* spec ;
* tâche ;
* fichiers de contexte ;
* diff ;
* logs ;
* résultats d’investigation ;
* format de sortie attendu ;
* historique éventuel.

5.3 Approche recommandée

V3.1

Utiliser une estimation approximative mais rapide :

characters / 4 ≈ tokens

Avec correction par type de contenu :

token_estimation:
  default_chars_per_token: 4.0
  code_chars_per_token: 3.2
  markdown_chars_per_token: 4.2
  json_chars_per_token: 3.6

V3.2

Ajouter des tokenizers spécifiques si disponibles :

* tokenizer OpenAI compatible ;
* SentencePiece pour modèles locaux ;
* tokenizer provider-specific si stable.

Le système doit rester fonctionnel sans tokenizer exact.

⸻

6. Cost estimation

6.1 Configuration pricing

pricing:
  currency: EUR
  models:
    gemini-3-flash-preview:
      input_per_1m_tokens: 0.00 # à configurer manuellement
      output_per_1m_tokens: 0.00
      source: manual
      updated_at: "2026-05-17"
    claude-opus:
      input_per_1m_tokens: 0.00
      output_per_1m_tokens: 0.00
      source: manual
      updated_at: "2026-05-17"

Les valeurs ne doivent pas être hardcodées dans le code.

6.2 Estimation

cost = input_tokens / 1_000_000 * input_price
     + output_tokens / 1_000_000 * output_price

6.3 Prise en compte des étapes locales

Les étapes locales ont un coût modèle nul, mais peuvent avoir :

* coût CPU estimé ;
* durée estimée ;
* mémoire estimée ;
* énergie non suivie en V3.

⸻

7. Time estimation

7.1 Objectif

Donner une estimation raisonnable du temps avant lancement.

7.2 Méthode V3.1

Combiner :

* moyenne historique par agent ;
* taille contexte ;
* complexité tâche ;
* type d’étape ;
* durée des runs précédents similaires.

type DurationModel interface {
    Estimate(ctx context.Context, step PlannedStep, history RunHistory) time.Duration
}

7.3 Historique local

Stocker dans SQLite :

agent
model
task_type
input_tokens
output_tokens
duration_ms
success
validation_status

Après quelques runs, AgentFlow doit apprendre localement des estimations plus fiables.

⸻

8. Context optimization

8.1 Problème

Les agents gaspillent des tokens quand on leur donne :

* trop de fichiers ;
* des logs complets ;
* des specs longues ;
* du contexte redondant ;
* des fichiers non pertinents.

8.2 Pipeline cible

Task
  ↓
Local investigation
  ↓
Candidate files
  ↓
Relevance scoring
  ↓
Context compression
  ↓
Context packing
  ↓
Cost estimate
  ↓
Agent execution

8.3 Stratégies

Recherche locale

Utiliser localement :

* ripgrep ;
* git grep ;
* tree-sitter ;
* langage-specific parsers ;
* composer/npm/go module graph ;
* test file mapping.

Context reducer

Réduire :

* logs complets → extraits pertinents ;
* fichiers longs → symboles pertinents ;
* diffs longs → hunks pertinents ;
* specs longues → sections liées à la tâche.

Context packer

Construire un contexte final trié :

1. objectif tâche
2. critères d’acceptation
3. contraintes fichiers
4. résultats investigation locale
5. extraits fichiers pertinents
6. commandes de validation
7. format attendu

⸻

9. Local investigation engine

9.1 Objectif

Permettre à AgentFlow de réaliser des investigations locales avant de solliciter un LLM.

Exemples :

agentflow investigate billing-v2 --task task-003
agentflow inspect symbol BillingCalculator
agentflow inspect tests src/Billing/Calculator.php
agentflow inspect diff --task task-003

9.2 Capacités V3

* rechercher symboles ;
* trouver tests liés ;
* extraire imports/dépendances ;
* lister fichiers modifiés ;
* extraire signatures de fonctions ;
* détecter routes/API endpoints ;
* détecter migrations DB ;
* détecter fichiers sensibles ;
* extraire erreurs de logs ;
* détecter fichiers volumineux à résumer.

9.3 Utilisation par les agents

Deux modes possibles.

Mode direct commands

AgentFlow appelle directement les commandes locales avant le prompt.

agentflow work
  ↓
investigation local automatique
  ↓
contexte réduit
  ↓
agent cloud

Mode MCP

AgentFlow expose un MCP local pour permettre à un agent de demander explicitement :

* recherche fichier ;
* extraction symbole ;
* estimation tokens ;
* résumé local ;
* diff local ;
* lecture sécurisée.

Le MCP ne doit pas remplacer les commandes internes. Il expose une partie des capacités aux agents compatibles.

⸻

10. MCP local AgentFlow

10.1 Objectif

Fournir aux agents une interface contrôlée vers les capacités locales d’AgentFlow.

agentflow mcp serve

10.2 Tools exposés

agentflow.search
agentflow.read_file_safe
agentflow.extract_symbols
agentflow.find_related_tests
agentflow.estimate_tokens
agentflow.estimate_cost
agentflow.get_task_context
agentflow.get_run_status
agentflow.get_diff_summary
agentflow.run_local_check

10.3 Sécurité MCP

Chaque tool MCP doit respecter :

* scope fichiers ;
* politiques secrets ;
* chemins interdits ;
* timeout ;
* taille de sortie maximale ;
* logging.

10.4 Exemple tool

{
  "tool": "agentflow.estimate_cost",
  "input": {
    "model": "gemini-3-flash-preview",
    "files": ["src/Billing/Calculator.php", "tests/Billing/CalculatorTest.php"],
    "expected_output_tokens": 4000
  }
}

⸻

11. Routing coût/qualité

11.1 Objectif

Choisir automatiquement le meilleur agent/modèle selon :

* complexité ;
* risque ;
* coût estimé ;
* budget ;
* disponibilité locale ;
* historique de réussite ;
* taille contexte.

11.2 Politique de routing

routing:
  default_strategy: cost_aware
  strategies:
    cost_aware:
      prefer_local_for:
        - summarize
        - classify
        - context_selection
        - pre_review
        - log_analysis
      use_cloud_fast_for:
        - implementation_medium
        - review_medium
        - planning_complex
      use_cloud_heavy_for:
        - architecture_critical
        - security_sensitive
        - large_refactor
        - repeated_failure
      escalation:
        local_failures_before_cloud: 1
        cloud_fast_failures_before_heavy: 1

11.3 Exemple décision

Task: task-003
Type: backend_refactor
Risk: medium
Context estimate: 38k tokens
Local model expected quality: acceptable for enrichment, not for implementation
Decision:
- context selection: ollama_local_qwen
- implementation: gemini_flash
- review: codex

⸻

12. Extension de agentflow work

12.1 Nouvelle séquence par défaut

resolve intent
  ↓
sync source if needed
  ↓
local investigation
  ↓
context optimization
  ↓
estimate cost/time/tokens
  ↓
show execution plan
  ↓
execute local preprocessing
  ↓
execute agent tasks
  ↓
verify locally
  ↓
measure actuals
  ↓
report

12.2 Options ajoutées

agentflow work "développe billing-v2" \
  --estimate-only \
  --budget 0.50 \
  --prefer-local \
  --max-input-tokens 100000 \
  --max-output-tokens 10000 \
  --max-duration 20m \
  --show-context-plan \
  --no-cloud \
  --allow-cloud \
  --allow-over-budget

12.3 --estimate-only

Ne lance rien. Produit seulement :

* étapes prévues ;
* taille contexte ;
* tokens estimés ;
* coût estimé ;
* durée estimée ;
* modèle recommandé ;
* risques ;
* recommandations de réduction.

⸻

13. UI terminal live

13.1 Objectif

Rendre l’exécution lisible, utile et agréable sans masquer les informations importantes.

L’interface doit rester compatible avec :

* mode interactif ;
* mode CI ;
* logs texte simples ;
* sortie JSON.

13.2 Modes UI

ui:
  mode: auto # auto | rich | plain | json
  live_logs: true
  progress_bars: true
  compact: false

13.3 Affichage work

Exemple :

AgentFlow Work
══════════════
Feature  billing-v2
Task     task-003
Mode     cost-aware
Budget   €0.50
Plan
────
✓ Resolve intent              0.2s
✓ Sync local spec             0.1s
⠋ Investigate repo            src/Billing/**
  ├─ files scanned            142
  ├─ candidate files          8
  └─ related tests            3
Estimate
────────
Input tokens                  41.8k
Output tokens                 6.0k
Estimated cost                €0.08
Estimated time                2m30s
Budget status                 OK
Execution
─────────
[████████░░░░░░░░░░░░] 40%  Context optimization
Current step: summarizing large files locally

13.4 Live metrics

Afficher pendant le run :

* étape courante ;
* durée écoulée ;
* fichiers scannés ;
* tokens estimés ;
* budget consommé ;
* agent actif ;
* modèle actif ;
* commandes locales en cours ;
* validations passées/échouées.

13.5 Implémentation Go recommandée

Bibliothèques possibles :

* charmbracelet/bubbletea pour TUI avancée ;
* charmbracelet/lipgloss pour styling ;
* charmbracelet/bubbles pour progress/spinners ;
* fallback plain text automatique si terminal non interactif.

La dépendance TUI doit rester isolée dans internal/tui.

⸻

14. Stockage métriques

14.1 SQLite tables

run_metrics

CREATE TABLE run_metrics (
  run_id TEXT PRIMARY KEY,
  feature TEXT,
  task_id TEXT,
  started_at TEXT,
  finished_at TEXT,
  estimated_input_tokens INTEGER,
  estimated_output_tokens INTEGER,
  actual_input_tokens INTEGER,
  actual_output_tokens INTEGER,
  estimated_cost_cents INTEGER,
  actual_cost_cents INTEGER,
  estimated_duration_ms INTEGER,
  actual_duration_ms INTEGER,
  status TEXT
);

step_metrics

CREATE TABLE step_metrics (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  step_name TEXT,
  agent TEXT,
  model TEXT,
  local BOOLEAN,
  estimated_input_tokens INTEGER,
  estimated_output_tokens INTEGER,
  actual_input_tokens INTEGER,
  actual_output_tokens INTEGER,
  estimated_cost_cents INTEGER,
  actual_cost_cents INTEGER,
  estimated_duration_ms INTEGER,
  actual_duration_ms INTEGER,
  status TEXT
);

⸻

15. Reports V3

Le rapport doit inclure une section coût/performance.

## Cost & Performance
| Metric | Estimated | Actual |
|---|---:|---:|
| Input tokens | 41,800 | 39,950 |
| Output tokens | 6,000 | 5,420 |
| Cost | €0.08 | €0.07 |
| Duration | 2m30s | 2m12s |
## Local Work Saved
- 142 files scanned locally
- 8 candidate files selected
- 3 large files summarized locally
- estimated cloud context reduced from 210k to 41.8k tokens
- estimated token savings: 80.1%

⸻

16. Nouvelles commandes V3

agentflow estimate <feature> [--task <id>]
agentflow investigate <feature> [--task <id>]
agentflow context <feature> [--task <id>] --show
agentflow context <feature> [--task <id>] --optimize
agentflow cost report [--since 7d]
agentflow cost models
agentflow mcp serve

16.1 agentflow estimate

Produit l’estimation sans exécuter.

agentflow estimate billing-v2 --task task-003

16.2 agentflow cost report

Affiche les coûts historiques.

Last 7 days
───────────
Runs:              18
Estimated cost:    €4.20
Actual cost:       €3.76
Local steps:       74%
Cloud steps:       26%
Avg token savings: 68%

16.3 agentflow context --optimize

Affiche le contexte prévu et les réductions possibles.

Original context: 212k tokens
Optimized:        43k tokens
Savings:          79.7%

⸻

17. Critères d’acceptation V3

La V3 est acceptable si :

* agentflow work affiche une estimation tokens/coût/temps avant exécution ;
* agentflow estimate fonctionne sans lancer d’agent cloud ;
* les tâches locales sont exécutées avant les appels modèle ;
* le contexte envoyé aux modèles est mesuré ;
* le contexte peut être réduit automatiquement ;
* les budgets bloquent ou demandent confirmation ;
* les métriques réelles sont stockées ;
* un rapport coût/performance est généré ;
* la TUI rich fonctionne en terminal interactif ;
* un fallback plain text fonctionne en CI ;
* le MCP local expose au moins search/read/estimate/context ;
* aucune donnée sensible n’est exposée via MCP ou contexte optimisé.

⸻

18. Découpage d’implémentation recommandé

Phase 1 — Estimation locale

* token counter approximatif ;
* pricing config ;
* ExecutionEstimate ;
* agentflow estimate ;
* affichage dans work --estimate-only.

Phase 2 — Metrics réelles

* tables SQLite ;
* collecte durée ;
* collecte tokens estimés/réels si disponibles ;
* rapport coût/performance.

Phase 3 — Investigation locale

* ripgrep/git grep ;
* fichiers candidats ;
* tests liés ;
* fichiers sensibles ;
* logs locaux.

Phase 4 — Context optimization

* reducer ;
* packer ;
* summaries locaux ;
* mesure savings.

Phase 5 — Routing coût/qualité

* stratégies de routing ;
* budgets ;
* escalade local → cloud fast → cloud heavy.

Phase 6 — UI terminal rich

* progress live ;
* timeline ;
* live metrics ;
* fallback plain/json.

Phase 7 — MCP local

* serveur MCP ;
* tools sécurisés ;
* intégration agents compatibles.

⸻

19. Risques

19.1 Estimation imprécise

Mitigation :

* afficher une confiance ;
* stocker estimé vs réel ;
* améliorer avec historique ;
* ne pas prétendre à une exactitude parfaite.

19.2 Optimisation de contexte trop agressive

Mitigation :

* conserver le contexte pack généré ;
* permettre --no-context-reduction ;
* logguer les fichiers exclus ;
* permettre review du contexte avant exécution.

19.3 TUI trop lourde

Mitigation :

* isoler internal/tui ;
* fallback plain obligatoire ;
* aucune dépendance UI dans le moteur.

19.4 MCP trop permissif

Mitigation :

* scope strict ;
* sorties bornées ;
* secrets scanner ;
* audit logs ;
* désactivé par défaut en V3.1.

⸻

20. Résumé exécutable

La V3 ajoute une couche cost/performance-aware :

Intent
  ↓
Local investigation
  ↓
Context optimization
  ↓
Token/cost/time estimate
  ↓
Budget validation
  ↓
Agent execution
  ↓
Metrics actuals
  ↓
Cost/performance report

Principe clé :

Les tokens doivent être dépensés uniquement pour le raisonnement à valeur ajoutée. Tout le reste doit être traité localement, mesuré et compressé avant appel modèle.