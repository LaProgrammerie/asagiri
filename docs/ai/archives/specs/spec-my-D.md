Spec — Asagiri Multi-Agent Coordination System

1. Vision

Le Multi-Agent Coordination System ajoute à Asagiri une orchestration explicite des agents.

Le but n’est pas de lancer plusieurs agents “en vrac”.

Le but est de permettre à Asagiri de :

* choisir le bon agent ;
* limiter son scope ;
* contrôler ses permissions ;
* gérer les dépendances ;
* coordonner les validations ;
* éviter les conflits ;
* mesurer coût/confiance/performance.

Pattern cible :

intent
  ↓
execution graph
  ↓
agent assignment
  ↓
isolated execution
  ↓
cross-agent validation
  ↓
trust gates
  ↓
merge / reject

⸻

2. Objectifs

Le système doit permettre :

* orchestration multi-agent déterministe ;
* rôles spécialisés ;
* isolation des agents ;
* coordination via runtime events ;
* handoff explicites ;
* validation croisée ;
* budgeting ;
* retry/escalation ;
* reprise runtime.

⸻

3. Types d’agents

Agents principaux

investigator
architect
planner
implementer
reviewer
validator
security_auditor
performance_auditor
observability_auditor
documenter

⸻

4. Rôles spécialisés

Chaque agent doit avoir :

* un rôle ;
* un scope ;
* des capacités ;
* des limites ;
* des permissions.

Exemple :

agent:
  id: claude-reviewer
  role: reviewer
capabilities:
  - architecture_review
  - flow_review
  - contract_review
restrictions:
  - no_direct_write
  - no_git_push
max_context_tokens: 32000

⸻

5. Isolation

Les agents ne doivent pas travailler directement dans le même workspace.

Modes possibles :

shared
isolated_worktree
readonly
sandbox

Par défaut :

execution:
  isolation: isolated_worktree

⸻

6. Agent Assignment

Le moteur doit choisir un agent selon :

* type de tâche ;
* risque ;
* coût ;
* taille contexte ;
* historique réussite ;
* spécialisation ;
* confiance.

Exemple :

assignment:
  investigation: local
  architecture_review: claude
  implementation.medium: cursor
  implementation.high: claude
  security_review: codex
  validation: local

⸻

7. Agent Pipeline

Exemple :

pipeline:
  - investigator
  - architect
  - implementer
  - reviewer
  - validator

Chaque étape doit produire :

* outputs ;
* confiance ;
* coûts ;
* événements ;
* contexte réduit.

⸻

8. Handoffs

Les handoffs doivent être explicites.

Créer :

.asagiri/handoffs/

Exemple :

from: investigator
to: implementer
summary:
  onboarding invite failure likely caused by missing retry handling
files:
  - src/Invitation/InvitationService.php
  - tests/Invitation/InvitationServiceTest.php
constraints:
  - preserve public API
  - maintain onboarding flow metrics
confidence: 0.78

⸻

9. Cross-agent Validation

Les agents critiques doivent être revus par un agent indépendant.

Exemple :

implementer
  ↓
reviewer
  ↓
validator

Règles :

* reviewer ≠ implementer ;
* validator ≠ implementer ;
* security review obligatoire sur flows sensibles.

⸻

10. Runtime Events

Ajouter événements :

agent.started
agent.completed
agent.failed
agent.blocked
agent.review_requested
agent.review_rejected
agent.handoff.created
agent.context_reduced

⸻

11. Coordination Policies

Créer :

coordination:
  max_parallel_agents: 2
  require_independent_review: true
  allow_self_review: false
  require_security_review_for:
    - auth
    - permissions
    - payments

⸻

12. Context Reduction

Chaque agent doit recevoir un contexte minimal.

Le coordinateur doit :

* réduire scope ;
* filtrer fichiers ;
* injecter flows liés ;
* injecter contrats ;
* injecter trust constraints ;
* injecter outputs investigation.

⸻

13. Budgeting

Le coordinateur doit suivre :

* coût agent ;
* tokens ;
* durée ;
* retries ;
* coût total pipeline.

Exemple :

Agent Budget
────────────
Investigation:   0.00€
Implementation:  0.14€
Review:          0.05€
Validation:      0.00€
Total:           0.19€

⸻

14. Retry & Escalation

Exemple :

retry:
  implementation:
    max_attempts: 2
escalation:
  after_failure:
    investigator
  after_second_failure:
    architecture_review

⸻

15. Conflict Detection

Le système doit détecter :

* modifications concurrentes ;
* conflits fichiers ;
* divergence contrats ;
* divergence flows ;
* trust downgrade.

⸻

16. Merge Policies

Créer politiques :

merge:
  require:
    - trust_passed
    - review_passed
    - validation_passed
  block_if:
    - unresolved_conflicts
    - low_security_confidence

⸻

17. Runtime Coordination Graph

Créer un graphe runtime représentant :

agent
↔ task
↔ flow
↔ branch
↔ review
↔ trust
↔ investigation

⸻

18. Architecture Go

Créer :

internal/coordination/
  coordinator.go
  assignment.go
  handoff.go
  policies.go
  runtime.go
  conflict.go
  budgeting.go
  escalation.go

Interfaces :

type AgentCoordinator interface {
    Coordinate(ctx context.Context, graph ExecutionGraph) error
}
type AgentAssigner interface {
    Assign(ctx context.Context, node GraphNode) (AgentAssignment, error)
}
type HandoffBuilder interface {
    Build(ctx context.Context, result AgentResult) (Handoff, error)
}

⸻

19. UX terminal cible

Asagiri Multi-Agent Runtime
═══════════════════════════
Pipeline
────────
✓ investigator
✓ architect
⠋ implementer
○ reviewer
○ validator
Agents
──────
implementer: cursor
reviewer: claude
validator: local
Costs
─────
Current: 0.11€
Budget:  1.00€
Warnings
────────
- reviewer pending
- trust validation required before merge

⸻

20. Critères d’acceptation

Cette évolution est acceptable si :

* les agents ont des rôles explicites ;
* les handoffs sont persistés ;
* les reviewers sont indépendants ;
* les conflits sont détectés ;
* les budgets sont suivis ;
* les retries/escalations fonctionnent ;
* les événements runtime sont émis ;
* les worktrees isolés fonctionnent ;
* les pipelines sont rejouables ;
* les validations trust peuvent bloquer ;
* les tests unitaires/intégration critiques passent.

⸻

21. Risques

Agent chaos

Mitigation :

* rôles stricts ;
* scopes ;
* isolation ;
* gates ;
* max_parallel_agents.

Explosion coûts

Mitigation :

* budgeting ;
* escalation ;
* retry caps ;
* context reduction.

Context drift

Mitigation :

* handoffs structurés ;
* replay ;
* runtime graph ;
* trust validation.

⸻

22. Résumé

Cette évolution transforme Asagiri d’un simple orchestrateur d’outils vers un coordinateur d’agents spécialisés.

Avant :

one agent tries to do everything

Après :

specialized agents
  ↓
controlled coordination
  ↓
isolated execution
  ↓
cross-validation
  ↓
verified delivery

Principe clé :

Les agents ne doivent pas être autonomes sans gouvernance. Asagiri doit coordonner, limiter, valider et expliquer leurs interactions.