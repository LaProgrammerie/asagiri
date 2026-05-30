Spec G — Asagiri Experience Platform

1. Vision

L’Asagiri Experience Platform transforme Asagiri d’une CLI technique en application terminal moderne pour piloter des workflows d’ingénierie IA complexes.

Le but n’est pas de remplacer la CLI existante.

Le but est de fournir une couche d’expérience interactive au-dessus des primitives existantes.

Core Engine
  ↓
Command API / primitives
  ↓
CLI commands + TUI application

Le point d’entrée interactif devient :

asa

Les commandes restent disponibles :

asa work "add invitations"
asa investigate "onboarding fails"
asa verify trust onboarding
asa plan graph workspace-saas
asa replay run replay-001

Principe clé :

La TUI est un client du moteur. Elle compose les mêmes commandes que la CLI. Elle ne crée pas de workflow opaque inaccessible en ligne de commande.

⸻

2. Positionnement

Asagiri doit proposer deux expériences complémentaires.

2.1 CLI stable

Pour :

* CI ;
* scripts ;
* power users ;
* agents ;
* automatisation ;
* documentation ;
* workflows reproductibles.

2.2 TUI interactive

Pour :

* exploration ;
* pilotage ;
* compréhension ;
* supervision live ;
* navigation dans les graphes ;
* review ;
* replay ;
* investigation ;
* prototypage.

La CLI reste l’API stable.

La TUI devient le cockpit.

⸻

3. Non-négociables UX

3.1 Toute action TUI doit avoir un équivalent CLI

Exemples :

Action TUI	Commande équivalente
Start work	asa work "..."
Run investigation	asa investigate "..."
Verify trust	asa verify trust ...
Open graph	asa graph visualize ...
Replay run	asa replay run ...
Build knowledge graph	asa knowledge build

3.2 Aucun moteur dans la couche UI

La TUI ne doit pas contenir de logique métier.

Elle doit appeler :

* services applicatifs ;
* runtime API ;
* command bus ;
* query bus ;
* read models.

3.3 Mode plain/json obligatoire

Tout workflow visible dans la TUI doit aussi fonctionner en :

--plain
--json
--ci

3.4 Progressive disclosure

L’interface doit montrer d’abord :

* état global ;
* prochaine action ;
* risques ;
* coût ;
* statut agents.

Puis permettre de descendre dans :

* logs ;
* graphes ;
* contexts ;
* reports ;
* evidence ;
* events.

⸻

4. Objectifs produit

L’Experience Platform doit permettre à l’utilisateur de :

1. comprendre l’état du workspace en moins de 10 secondes ;
2. voir les agents travailler en live ;
3. visualiser flows, graphs, trust et replay ;
4. naviguer au clavier et à la souris ;
5. lancer les actions importantes sans retenir 100 commandes ;
6. comprendre pourquoi Asagiri prend une décision ;
7. passer de l’intention au travail vérifié avec un cockpit clair ;
8. conserver une UX premium même dans un terminal.

⸻

5. Expérience cible

5.1 Entrée principale

asa

Ouvre Mission Control.

5.2 Commandes directes conservées

asa work "add workspace invitations"
asa dashboard
asa mission
asa agents watch
asa flow open onboarding
asa trust open
asa replay open replay-001

5.3 Règle de cohérence

asa et asa <command> doivent produire des états cohérents.

La TUI ne doit jamais masquer une action non reproductible.

⸻

6. Architecture technique

6.1 Stack recommandée

Utiliser l’écosystème Charmbracelet :

Bubble Tea  — event loop / TUI architecture
Lip Gloss   — styling
Bubbles     — widgets
Huh         — forms / prompts
Glamour     — markdown rendering

6.2 Architecture Go

Créer :

internal/ui/
  app/
    app.go
    router.go
    commands.go
  layout/
    engine.go
    panes.go
    split.go
    focus.go
  components/
    panel.go
    card.go
    table.go
    tree.go
    graph.go
    timeline.go
    progress.go
    logview.go
    palette.go
  screens/
    mission/
    dashboard/
    agents/
    flows/
    trust/
    graph/
    knowledge/
    replay/
    prototype/
    settings/
  theme/
    theme.go
    palette.go
  input/
    keyboard.go
    mouse.go
  state/
    store.go
    subscriptions.go

6.3 Séparation stricte

internal/ui
  depends on application services
  does not own business logic

Interdit :

* accès direct SQLite métier depuis composants UI ;
* logique trust dans les composants ;
* logique investigation dans les écrans ;
* appels agents depuis widgets.

⸻

7. Command Bus / Query Bus

La TUI doit passer par une API interne stable.

type CommandBus interface {
    Dispatch(ctx context.Context, cmd Command) (CommandResult, error)
}
type QueryBus interface {
    Query(ctx context.Context, query Query) (QueryResult, error)
}

Exemples :

StartWorkCommand
RunInvestigationCommand
VerifyTrustCommand
BuildKnowledgeGraphCommand
ReplayRunCommand
GetRuntimeStatusQuery
ListActiveAgentsQuery
GetTrustSummaryQuery
GetFlowGraphQuery
GetRecentEventsQuery

⸻

8. Terminal Design System

Créer un vrai design system terminal.

8.1 Composants de base

Panel
Card
MetricCard
ProgressBar
StatusBadge
Timeline
TreeView
TableView
GraphView
LogView
EventFeed
CommandPalette
Tabs
Breadcrumb
Modal
Drawer
SplitPane
Toast

8.2 Composants métier

RuntimeCard
AgentCard
TrustCard
FlowCard
RiskCard
CostCard
GraphNodeCard
InvestigationCard
ReplayCard
KnowledgeCard
PrototypeCard

8.3 États visuels

idle
running
success
warning
error
blocked
paused
waiting
unknown

⸻

9. Layout Engine

Supporter plusieurs layouts.

single
split-horizontal
split-vertical
grid
dashboard
focus
fullscreen

Fonctionnalités :

* focus par panneau ;
* redimensionnement ;
* collapse/expand ;
* tabs ;
* panels persistants ;
* layout adapté largeur terminal.

⸻

10. Navigation

10.1 Raccourcis globaux

Ctrl+P  Command Palette
Ctrl+D  Dashboard
Ctrl+M  Mission Control
Ctrl+L  Logs
Ctrl+E  Explain
Ctrl+R  Replay
Ctrl+K  Knowledge
?       Help
q       Quit / Back
Esc     Close modal
Tab     Next panel
Shift+Tab Previous panel

10.2 Souris

Supporter :

* clic ;
* double clic ;
* scroll ;
* hover si possible ;
* menus contextuels ;
* sélection ;
* resize panels.

Le clavier reste obligatoire.

La souris améliore l’expérience mais ne devient pas indispensable.

⸻

11. Mission Control

Commande :

asa mission

Ou écran par défaut avec :

asa

Objectif : comprendre l’état global en moins de 10 secondes.

11.1 Contenu

* workspace actif ;
* runtime status ;
* sessions ;
* flows critiques ;
* agents actifs ;
* trust global ;
* coût du jour/mois ;
* derniers événements ;
* actions recommandées.

11.2 Exemple

╭────────────────────────────── ASAGIRI ──────────────────────────────╮
│ Workspace  workspace-saas              Runtime   running             │
│ Branch     onboarding-v2                Session   active              │
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Runtime ────────────╮ ╭──────────── Trust ───────────────╮
│ Agents      3                   │ │ Architecture  █████████░░ 82%    │
│ Sessions    4                   │ │ Security      ███████░░░░ 71%    │
│ Queue       2                   │ │ Observability ██████░░░░░ 63%    │
│ Cost today  €0.42               │ │ Regression    ████████░░ 78%     │
╰────────────────────────────────╯ ╰──────────────────────────────────╯
╭──────────── Active Flow ─────────────────────────────────────────────╮
│ Onboarding                                                           │
│ ✓ create_workspace   ⠋ invite_member   ○ accept_invite   ○ complete  │
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Agent Theatre ───────────────────────────────────────────╮
│ Investigator ✓ done     Architect ✓ done     Implementer ⠋ running   │
│ Reviewer     ○ waiting  Validator ○ waiting                          │
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Events ──────────────────────────────────────────────────╮
│ 08:14 investigation.completed     confidence=0.78                    │
│ 08:15 graph.generated             nodes=12 edges=16                  │
│ 08:16 implementation.started      agent=cursor                       │
╰──────────────────────────────────────────────────────────────────────╯

⸻

12. Dashboard

Commande :

asa dashboard

Dashboard live avec refresh temps réel.

Widgets :

* runtime ;
* agents ;
* costs ;
* trust ;
* flows ;
* events ;
* knowledge ;
* sessions ;
* queue ;
* performance.

12.1 Contraintes

* refresh non bloquant ;
* fallback plain si terminal non compatible ;
* mode compact pour petit écran ;
* mode wide pour grands écrans.

⸻

13. Agent Theatre

Commande :

asa agents watch

Objectif : visualiser les agents en temps réel.

13.1 Données affichées

Pour chaque agent :

* rôle ;
* statut ;
* tâche ;
* fichiers actifs ;
* hypothèse courante ;
* tokens estimés ;
* coût ;
* durée ;
* dernier output ;
* confiance.

13.2 Exemple

╭──────── Investigator ────────╮ ╭──────── Implementer ─────────╮
│ Status      done             │ │ Status      running           │
│ Files       421 scanned      │ │ Editing     InvitationService │
│ Hypothesis  missing retry    │ │ Progress    ███████░░░ 67%    │
│ Confidence  78%              │ │ Cost        €0.09             │
╰──────────────────────────────╯ ╰──────────────────────────────╯
╭──────── Reviewer ────────────╮ ╭──────── Validator ───────────╮
│ Status      waiting          │ │ Status      waiting           │
│ Blocked by  implementer      │ │ Next        phpunit           │
╰──────────────────────────────╯ ╰──────────────────────────────╯

⸻

14. Execution Graph Explorer

Commande :

asa graph

Objectif : comprendre le plan, les dépendances et les blocages.

Vues :

timeline
dependency
critical-path
parallel-groups
blocked

Actions :

* ouvrir nœud ;
* voir logs ;
* voir dépendances ;
* expliquer décision ;
* lancer/reprendre ;
* exporter Mermaid/JSON.

⸻

15. Flow Explorer

Commande :

asa flow
asa flow open onboarding

Objectif : naviguer produit → système.

Vue cible :

Flow: onboarding
create_workspace
  ↓
invite_member
  ↓
accept_invite
  ↓
complete
Selected: invite_member
───────────────────────
API        POST /invitations
Service    InvitationService
Event      member.invited
Tests      InvitationServiceTest
Metrics    invitation_success_rate
Trust      71%
Risk       medium

⸻

16. Knowledge Explorer

Commande :

asa knowledge

Recherche interactive :

invite_member

Affiche :

* flows ;
* API ;
* services ;
* tests ;
* events ;
* metrics ;
* incidents ;
* reviews ;
* costs.

Actions :

* impact analyze ;
* build context ;
* open graph ;
* explain relationship.

⸻

17. Trust Explorer

Commande :

asa trust

Objectif : explorer confiance et risques.

Vue :

Trust Summary
─────────────
Architecture     82%
Implementation   67%
Security         71%
Observability    63%
Regression       78%
Warnings
────────
- no retry validation for invite_member
- onboarding funnel missing dashboard
- rate limiting not verified

Chaque score doit être cliquable/ouvrable pour voir :

* evidence ;
* findings ;
* checks ;
* gates ;
* residual risks.

⸻

18. Replay Explorer / Time Travel

Commande :

asa replay open replay-001

Objectif : naviguer dans une exécution passée.

Vue timeline :

08:12 investigation.started
08:14 investigation.completed
08:15 graph.generated
08:16 implementation.started
08:21 trust.started
08:23 review.inserted
08:25 completed

Actions :

* jump to event ;
* inspect artifact ;
* compare run ;
* replay offline ;
* explain divergence.

⸻

19. Prototype Mode

Commande :

asa prototype

Objectif : rendre visible le pipeline :

wireframe
  ↓
journey
  ↓
flow
  ↓
contracts
  ↓
tasks
  ↓
execution graph

Vue split :

╭──────────── Wireframe ────────────╮ ╭────── Flow Extraction ───────╮
│                                   │ │ Flow: login                  │
│  Email     [____________]         │ │ Action: submit_login         │
│  Password  [____________]         │ │ API: POST /login             │
│                                   │ │ Metric: auth_success_rate    │
│  [ Sign in ]                      │ │ Trust: pending               │
╰───────────────────────────────────╯ ╰──────────────────────────────╯

La TUI peut lancer :

* asa prototype create ;
* asa flows extract ;
* asa contracts extract ;
* asa spec generate-from-product.

⸻

20. Command Palette

Accessible partout via :

Ctrl+P

Fonctions :

* recherche commandes ;
* recherche flows ;
* recherche agents ;
* recherche reports ;
* actions contextuelles.

Exemples :

> investigate onboarding
> verify trust
> open replay last
> show graph
> explain current decision

Chaque entrée doit afficher la commande CLI équivalente.

Exemple :

Run investigation
CLI: asa investigate "onboarding"

⸻

21. Explainability Layer

Commande :

asa explain

Accessible depuis toute décision.

Questions supportées :

Why was review required?
Why is this node blocked?
Why is security confidence low?
Why was this agent selected?
Why is this flow high risk?
Why did Asagiri insert investigation?

Chaque réponse doit montrer :

* reasons ;
* evidence ;
* source ;
* alternatives ;
* CLI equivalent.

⸻

22. Event Feed

Afficher événements live.

Types :

runtime
agent
trust
investigation
graph
knowledge
replay
prototype

Fonctions :

* filter ;
* search ;
* pause ;
* export ;
* open artifact.

⸻

23. Widget System

Créer widgets composables.

23.1 Widget interface

type Widget interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Widget, tea.Cmd)
    View() string
    Title() string
    MinSize() Size
}

23.2 Widgets V1

RuntimeWidget
AgentWidget
TrustWidget
CostWidget
FlowWidget
RiskWidget
EventWidget
KnowledgeWidget
ReplayWidget
ProgressWidget

⸻

24. Live Updates

Utiliser subscriptions Bubble Tea.

Sources :

* runtime events ;
* graph state ;
* agent status ;
* trust checks ;
* cost metrics ;
* logs.

Contraintes :

* pas de flicker ;
* pas de blocage ;
* throttling ;
* degradation si terminal lent.

⸻

25. Animation System

Animations sobres :

* spinners ;
* progress bars ;
* loading shimmer minimal ;
* collapsible panels ;
* transitions de focus ;
* live counters ;
* sparklines.

Ne pas faire :

* animations décoratives inutiles ;
* effets qui masquent les logs ;
* surconsommation CPU.

⸻

26. Theme System

Thèmes :

asagiri-dark
asagiri-light
minimal
high-contrast
cyber

Config :

ui:
  theme: asagiri-dark
  mouse: true
  animations: true
  compact: false

⸻

27. Responsive Terminal

Support :

* narrow terminal ;
* standard terminal ;
* wide terminal ;
* ultra-wide.

Règles :

* narrow → single column ;
* standard → split ;
* wide → dashboard grid ;
* CI → plain output.

⸻

28. Accessibility

Support :

* navigation clavier complète ;
* contraste élevé ;
* no-animation mode ;
* mode plain ;
* screen-reader compatible where possible ;
* raccourcis listables.

⸻

29. Configuration UI

ui:
  default_screen: mission
  theme: asagiri-dark
  mouse: true
  animations: true
  refresh_interval_ms: 500
  compact_threshold: 100
  show_cli_equivalents: true
  confirm_destructive_actions: true

⸻

30. Safety UX

Toute action destructive doit afficher :

* impact ;
* commande CLI équivalente ;
* artefacts concernés ;
* rollback possible ou non ;
* confirmation explicite.

Exemple :

You are about to rollback graph-001.
Impacted:
- 2 worktrees
- 4 generated reports
- 1 active session
CLI equivalent:
asa graph rollback graph-001
Proceed? [y/N]

⸻

31. Documentation

Créer :

docs-site/content/docs/experience/
  index.mdx
  mission-control.mdx
  dashboard.mdx
  command-palette.mdx
  keyboard-shortcuts.mdx
  mouse-support.mdx
  themes.mdx
  accessibility.mdx

⸻

32. Tests

Unit tests

* widgets ;
* layout engine ;
* command routing ;
* theme ;
* keybindings.

Golden tests

* snapshots textuels ;
* compact layout ;
* wide layout ;
* error states.

Integration tests

* start asa ;
* open dashboard ;
* command palette ;
* trigger command ;
* verify CLI equivalent.

⸻

33. Critères d’acceptation

Cette évolution est acceptable si :

* asa ouvre Mission Control ;
* asa dashboard affiche un dashboard live ;
* la navigation clavier fonctionne ;
* la souris fonctionne quand supportée ;
* la Command Palette fonctionne ;
* chaque action affiche ou expose son équivalent CLI ;
* les widgets runtime/trust/agents/events fonctionnent ;
* les graphes peuvent être explorés ;
* les flows peuvent être explorés ;
* les reports trust/replay/investigation sont visibles ;
* le mode plain/json reste disponible ;
* aucune logique métier n’est dans internal/ui ;
* l’UI reste utilisable en terminal standard ;
* les tests unitaires/golden/intégration passent.

⸻

34. Découpage d’implémentation recommandé

Phase 1 — UI foundation

* Bubble Tea app ;
* router ;
* layout engine minimal ;
* theme ;
* basic panels.

Phase 2 — Mission Control

* runtime status ;
* active flows ;
* recent events ;
* trust summary.

Phase 3 — Command Palette

* actions ;
* CLI equivalents ;
* search ;
* execution.

Phase 4 — Dashboard widgets

* runtime ;
* agents ;
* trust ;
* cost ;
* events.

Phase 5 — Explorers

* Flow Explorer ;
* Graph Explorer ;
* Knowledge Explorer ;
* Trust Explorer.

Phase 6 — Replay & Prototype Mode

* timeline ;
* time travel ;
* prototype split view.

Phase 7 — Polish

* mouse ;
* animations ;
* themes ;
* responsive ;
* accessibility ;
* docs.

⸻

35. Risques

UI qui cache la CLI

Mitigation :

* CLI equivalents visibles ;
* aucune action TUI-only ;
* docs CLI-first conservées.

TUI trop lourde

Mitigation :

* phase progressive ;
* fallback plain ;
* internal/ui isolé.

Trop d’informations

Mitigation :

* progressive disclosure ;
* dashboard synthétique ;
* drill-down.

Performance terminal

Mitigation :

* throttling ;
* virtualized lists ;
* lazy loading ;
* no-animation mode.

⸻

36. Résumé

Cette évolution transforme Asagiri en application terminal moderne sans sacrifier la CLI.

Avant :

commands
  ↓
text output

Après :

CLI stable
  +
interactive terminal cockpit
  +
live dashboards
  +
explorable graphs
  +
agent theatre
  +
explainability everywhere

Principe clé :

Asagiri doit rester scriptable comme une CLI et devenir pilotable comme une application. La TUI rend la complexité visible, navigable et compréhensible sans enfermer l’utilisateur dans une interface opaque.