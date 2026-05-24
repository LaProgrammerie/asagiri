Spec d’évolution — AgentFlow Intent Layer & Sources externes

1. Objectif

Cette évolution ajoute une couche haut niveau à AgentFlow pour permettre un usage naturel du type :

agentflow work "développe la feature billing-v2"
agentflow work "reprends le dev de l'import CSV"
agentflow continue
agentflow sync notion

Les commandes existantes restent valides et deviennent les primitives bas niveau du moteur :

agentflow spec
agentflow plan
agentflow enrich
agentflow dev
agentflow verify
agentflow review
agentflow pr

La nouvelle couche ne remplace pas ces commandes. Elle les orchestre.

⸻

2. Positionnement

AgentFlow doit proposer deux niveaux d’usage.

2.1 Niveau bas niveau

Pour contrôle fin, debug, CI et agents :

agentflow plan billing-v2
agentflow dev billing-v2 --task task-003 --agent cursor
agentflow verify billing-v2 --task task-003
agentflow review billing-v2 --task task-003 --agent codex

2.2 Niveau intention

Pour usage quotidien :

agentflow work "reprends billing-v2"
agentflow continue
agentflow next

Le niveau intention doit :

* comprendre l’intention utilisateur ;
* retrouver la feature ou tâche concernée ;
* inspecter l’état courant ;
* déterminer la prochaine action ;
* enchaîner les primitives nécessaires ;
* produire un rapport clair.

⸻

3. Principes d’architecture

3.1 Le repo reste la source de vérité

Les sources externes comme Notion ne doivent pas être exécutées directement.

Pattern obligatoire :

Notion / GitHub Issue / autre source externe
  ↓ sync/import
.agentflow/specs/<feature>/
  spec.md
  tasks.yaml
  metadata.yaml
  source.json
  ↓
AgentFlow runtime

La spec exécutée par AgentFlow doit être locale, traçable et versionnable.

3.2 Les commandes haut niveau composent les commandes bas niveau

agentflow work ne doit pas réimplémenter le moteur.

Il doit produire un plan d’exécution basé sur les primitives existantes :

resolve intent
  ↓
load feature/task state
  ↓
decide next steps
  ↓
execute: spec/plan/enrich/dev/verify/review/pr

3.3 Toute décision doit être inspectable

Chaque commande intentionnelle doit produire :

* l’intention détectée ;
* la feature résolue ;
* la tâche résolue si applicable ;
* le plan d’actions ;
* les commandes primitives exécutées ;
* les raisons du choix.

⸻

4. Nouvelles commandes

4.1 agentflow work

Commande principale haut niveau.

Usage

agentflow work "développe billing-v2"
agentflow work "reprends le dev de billing-v2"
agentflow work "développe la spec Notion https://notion.so/..."
agentflow work "continue l'import CSV"
agentflow work "corrige les tests de la dernière tâche"

Responsabilités

work doit :

1. parser l’instruction ;
2. résoudre la source éventuelle ;
3. identifier feature/tâche/run ;
4. inspecter l’état ;
5. construire un plan d’exécution ;
6. afficher ce plan ;
7. exécuter sauf --plan-only ou --dry-run ;
8. produire un rapport.

Options

agentflow work "..." \
  --agent cursor \
  --reviewer codex \
  --source notion \
  --plan-only \
  --yes \
  --max-tasks 3 \
  --stop-after verify \
  --no-review

Comportement par défaut

Sans option :

resolve → plan if needed → enrich → dev next task → verify → report

La review indépendante peut être activée par config ou par flag.

⸻

4.2 agentflow continue

Reprend le travail le plus pertinent.

Usage

agentflow continue
agentflow continue --feature billing-v2
agentflow continue --run run-2026-05-17-001

Résolution par priorité

1. run interrompu ou failed mais reprenable ;
2. feature active avec tâche en cours ;
3. dernière feature modifiée ;
4. prochaine tâche pending d’une feature active ;
5. sinon afficher une liste d’options.

États reprenables

failed
verify_failed
review_failed
aborted
running_stale
implemented
verified

⸻

4.3 agentflow next

Affiche ou exécute la prochaine action recommandée.

Usage

agentflow next
agentflow next --feature billing-v2
agentflow next --execute

Sortie attendue

Feature: billing-v2
Next action: verify task-003
Reason: implementation completed but validation missing
Command: agentflow verify billing-v2 --task task-003

⸻

4.4 agentflow inbox

Liste les specs/tâches candidates depuis les sources configurées.

Usage

agentflow inbox
agentflow inbox --source notion
agentflow inbox --source local

Sortie attendue

Notion:
- billing-v2              updated 2026-05-17  status ready
- import-csv              updated 2026-05-16  status draft
Local:
- payment-routing         .kiro/specs/payment-routing
- api-contract-cleanup    docs/ai/specs/api-contract-cleanup

⸻

4.5 agentflow sync

Synchronise une source externe vers le repo local.

Usage

agentflow sync notion
agentflow sync notion --page https://notion.so/...
agentflow sync notion --feature billing-v2
agentflow sync all

Règles

* ne jamais exécuter directement une spec distante ;
* écrire dans .agentflow/specs/<feature>/ ;
* conserver les métadonnées de source ;
* détecter les conflits ;
* ne pas écraser une spec locale modifiée sans confirmation ou --force.

⸻

5. Intent Resolver

5.1 Rôle

Le IntentResolver transforme une instruction libre en intention structurée.

Interface Go

type IntentResolver interface {
    Resolve(ctx context.Context, input IntentInput) (ResolvedIntent, error)
}

Input

type IntentInput struct {
    RawInstruction string
    WorkingDir     string
    Config         Config
    StateSnapshot  StateSnapshot
}

Output

type ResolvedIntent struct {
    Action        IntentAction
    Feature       string
    TaskID        string
    RunID         string
    Source        string
    SourceRef     string
    Confidence    float64
    RequiresSync  bool
    RequiresPlan  bool
    RequiresHuman bool
    Constraints   IntentConstraints
}

Actions possibles

type IntentAction string
const (
    IntentDevelop        IntentAction = "develop"
    IntentResume         IntentAction = "resume"
    IntentContinue       IntentAction = "continue"
    IntentVerify         IntentAction = "verify"
    IntentReview         IntentAction = "review"
    IntentFix            IntentAction = "fix"
    IntentImport         IntentAction = "import"
    IntentSync           IntentAction = "sync"
    IntentStatus         IntentAction = "status"
    IntentUnknown        IntentAction = "unknown"
)

5.2 Résolution hybride

Le resolver doit être robuste sans LLM.

Ordre recommandé :

1. règles déterministes ;
2. matching sur features/tasks/runs existants ;
3. fuzzy matching local ;
4. LLM local via Ollama si ambigu ;
5. fallback interactif si nécessaire.

Exemple

Instruction :

reprends le dev de l'import CSV

Résultat :

{
  "action": "resume",
  "feature": "import-csv",
  "confidence": 0.91,
  "requires_sync": false,
  "requires_plan": false
}

⸻

6. Planner haut niveau

6.1 Rôle

Le HighLevelPlanner convertit une intention résolue en étapes primitives.

Interface Go

type HighLevelPlanner interface {
    BuildPlan(ctx context.Context, intent ResolvedIntent) (ExecutionPlan, error)
}

Exemple de plan

intent: develop
feature: billing-v2
steps:
  - command: sync
    args: ["notion", "--feature", "billing-v2"]
    condition: source_requires_sync
  - command: plan
    args: ["billing-v2"]
    condition: no_tasks
  - command: enrich
    args: ["billing-v2", "--task", "task-003", "--agent", "ollama"]
    condition: task_not_enriched
  - command: dev
    args: ["billing-v2", "--task", "task-003", "--agent", "cursor"]
    condition: task_pending_or_enriched
  - command: verify
    args: ["billing-v2", "--task", "task-003"]
    condition: implementation_done
  - command: review
    args: ["billing-v2", "--task", "task-003", "--agent", "codex"]
    condition: review_enabled

6.2 Conditions supportées

source_requires_sync
no_local_spec
no_tasks
task_not_enriched
task_pending_or_enriched
implementation_done
verification_failed
review_enabled
requires_human_approval

⸻

7. Sources externes

7.1 Interface Source

type Source interface {
    Name() string
    List(ctx context.Context) ([]SourceItem, error)
    Fetch(ctx context.Context, ref SourceRef) (SourceDocument, error)
    Sync(ctx context.Context, ref SourceRef, dest LocalSpecPath) (SyncResult, error)
}

7.2 Sources V1

Local source

* .kiro/specs
* docs/ai/active
* .agentflow/specs

Notion source

* pages Notion ;
* databases Notion ;
* propriétés title/status/updated_at ;
* contenu converti en Markdown.

7.3 Sources futures

* GitHub Issues ;
* Linear ;
* Jira ;
* Markdown distant ;
* Google Docs ;
* Slack export.

⸻

8. Intégration Notion

8.1 Configuration

sources:
  notion:
    enabled: true
    token_env: NOTION_TOKEN
    default_database_id: ""
    specs_database_id: ""
    tasks_database_id: ""
    status_property: Status
    title_property: Name
    updated_time_property: Last edited time
    import_path: .agentflow/specs

8.2 Format local généré

.agentflow/specs/billing-v2/
  spec.md
  tasks.yaml
  source.json
  metadata.yaml

source.json

{
  "type": "notion",
  "page_id": "xxx",
  "url": "https://notion.so/...",
  "last_synced_at": "2026-05-17T12:00:00+02:00",
  "remote_updated_at": "2026-05-17T11:54:00+02:00"
}

metadata.yaml

feature: billing-v2
source: notion
status: ready
owner: matt
synced_at: "2026-05-17T12:00:00+02:00"

8.3 Règles de sync

* convertir Notion en Markdown propre ;
* préserver les titres ;
* extraire les checklists comme tâches si possible ;
* conserver les blocs non supportés sous forme de note ;
* refuser les specs vides ;
* marquer les specs ambiguës comme draft.

⸻

9. Configuration mise à jour

project:
  name: my-project
  default_branch: main
intent:
  enabled: true
  default_mode: guided
  resolver:
    use_ollama_fallback: true
    min_confidence: 0.75
    ask_when_below_confidence: true
work:
  default_agent: cursor
  default_reviewer: codex
  default_enricher: ollama
  stop_after: report
  auto_verify: true
  auto_review: false
  max_tasks_per_run: 1
  require_plan_confirmation: true
sources:
  local:
    enabled: true
    paths:
      - .agentflow/specs
      - .kiro/specs
      - docs/ai/active
  notion:
    enabled: false
    token_env: NOTION_TOKEN
    specs_database_id: ""
    import_path: .agentflow/specs

⸻

10. Modes d’exécution

10.1 Mode guided

Mode par défaut.

AgentFlow affiche le plan et demande confirmation avant exécution si l’action est risquée.

Intent: develop
Feature: billing-v2
Next task: task-003
Plan:
  1. enrich task-003 with ollama
  2. dev task-003 with cursor
  3. verify task-003
Proceed? [y/N]

10.2 Mode auto

Pour usage avancé ou CI.

agentflow work "développe billing-v2" --yes

10.3 Mode plan-only

agentflow work "reprends billing-v2" --plan-only

Produit uniquement le plan d’exécution.

⸻

11. Interaction avec les agents

Les agents doivent pouvoir utiliser les commandes haut niveau, mais les commandes bas niveau restent préférées pour les plans explicites.

Exemple pour un agent orchestrateur

Entrée utilisateur :

reprends le dev de billing-v2

L’agent peut lancer :

agentflow work "reprends le dev de billing-v2" --plan-only

Puis exécuter :

agentflow work "reprends le dev de billing-v2" --yes

Ou dérouler manuellement les primitives retournées.

⸻

12. UX terminal attendue

agentflow work

AgentFlow resolved intent
─────────────────────────
Instruction: reprends le dev de billing-v2
Action:      resume
Feature:     billing-v2
Task:        task-003
Confidence:  0.94
Execution plan
──────────────
1. enrich billing-v2 --task task-003 --agent ollama
2. dev billing-v2 --task task-003 --agent cursor
3. verify billing-v2 --task task-003
4. report <run-id>

agentflow continue

Continuing last active feature
──────────────────────────────
Feature: billing-v2
Task:    task-003
State:   implemented
Next:    verify
Running:
agentflow verify billing-v2 --task task-003

agentflow inbox

Inbox
─────
Source  Status  Updated              Feature
notion  ready   2026-05-17 12:00     billing-v2
notion  draft   2026-05-16 18:20     import-csv
local   ready   2026-05-15 09:42     api-cleanup

⸻

13. Sécurité & garde-fous

13.1 Confirmation obligatoire

Confirmation requise pour :

* suppression worktree ;
* sync qui écrase une spec locale modifiée ;
* changement de dépendance ;
* migration DB ;
* modification fichier sensible ;
* plus de N fichiers modifiés ;
* exécution de plusieurs tâches d’un coup.

13.2 Ambiguïté

Si confiance inférieure à intent.resolver.min_confidence, AgentFlow doit afficher les candidats :

Instruction ambiguous. Did you mean:
1. billing-v2
2. billing-export
3. billing-cleanup

En mode non interactif, retourner une erreur structurée.

⸻

14. Critères d’acceptation

Cette évolution est acceptable si :

* les commandes existantes continuent de fonctionner ;
* agentflow work "développe <feature>" produit un plan cohérent ;
* agentflow continue reprend le dernier run ou la prochaine action logique ;
* agentflow next affiche la prochaine action sans effets de bord ;
* agentflow inbox liste les specs locales ;
* agentflow sync notion importe une page Notion en local ;
* une spec Notion n’est jamais exécutée directement sans copie locale ;
* le plan généré référence les primitives bas niveau ;
* les décisions du resolver sont visibles ;
* les cas ambigus sont gérés explicitement.

⸻

15. Découpage d’implémentation recommandé

Étape 1 — Façade locale sans Notion

* ajouter work ;
* ajouter continue ;
* ajouter next ;
* resolver déterministe ;
* planner haut niveau ;
* aucune dépendance externe.

Étape 2 — Inbox locale

* scanner .agentflow/specs ;
* scanner .kiro/specs ;
* afficher status local ;
* fuzzy matching feature.

Étape 3 — Source abstraction

* interface Source ;
* LocalSource ;
* persistance source.json.

Étape 4 — Notion import minimal

* config token ;
* fetch page ;
* conversion Markdown ;
* écriture .agentflow/specs/<feature>.

Étape 5 — Notion database

* liste inbox depuis database ;
* sync par feature ;
* détection statut ready/draft.

Étape 6 — LLM fallback resolver

* utiliser Ollama pour instructions ambiguës ;
* sortie JSON strict ;
* fallback déterministe si LLM indisponible.

⸻

16. Risques

16.1 Risque : commande work trop magique

Mitigation :

* afficher toujours l’intention résolue ;
* --plan-only ;
* mode guided par défaut ;
* plan composé de commandes primitives.

16.2 Risque : Notion devient source de vérité floue

Mitigation :

* sync locale obligatoire ;
* snapshot versionnable ;
* metadata source ;
* conflits explicites.

16.3 Risque : resolver LLM imprévisible

Mitigation :

* règles déterministes d’abord ;
* LLM seulement en fallback ;
* JSON schema strict ;
* score de confiance ;
* confirmation si ambigu.

⸻

17. Résumé exécutable

Ajouter une couche intentionnelle au-dessus du moteur existant :

work / continue / next / inbox / sync

Ces commandes doivent simplifier l’usage quotidien sans affaiblir le contrôle fin.

Architecture cible :

User intent
  ↓
IntentResolver
  ↓
HighLevelPlanner
  ↓
Existing primitives
  spec / plan / enrich / dev / verify / review / pr
  ↓
Report

Principe clé :

Les commandes haut niveau pilotent AgentFlow. Les commandes bas niveau restent l’API stable du moteur.