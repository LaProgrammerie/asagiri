Spec — Asagiri Executable Product Layer

Objectif : ajouter à Asagiri une couche de spécification produit exécutable inspirée du paradigme “prototype interactif comme spec vivante”, sans transformer Asagiri en clone de builder UI.

Cette évolution doit permettre à Asagiri de passer de :

Intent → spec markdown → tasks → agents → code

à :

Intent
→ prototype interactif
→ flows exécutables
→ contracts dérivés
→ specs générées
→ tasks agents
→ implémentation vérifiée

Le but est de faire d’Asagiri un système capable de transformer une intention produit en pipeline d’ingénierie industrialisable, mesurable et contrôlé.

⸻

1. Contexte

Asagiri est déjà un orchestrateur local de workflows de développement agentique.

Le système existant repose sur :

* une CLI Go ;
* une couche intentionnelle work / continue / next / inbox / sync ;
* des primitives bas niveau spec / plan / enrich / dev / verify / review / pr ;
* un moteur déterministe ;
* des agents interchangeables ;
* des worktrees Git isolés ;
* une validation externe obligatoire ;
* une stratégie local-first ;
* une estimation coût/tokens/temps ;
* des rapports vérifiables.

Cette spec ajoute une nouvelle couche située avant les specs et tasks classiques : la couche produit exécutable.

Elle doit permettre de matérialiser une idée sous forme de prototype interactif, puis d’en extraire des flows, contracts, specs et tâches.

⸻

2. Positionnement

Asagiri ne doit pas devenir :

* un clone de Lovable ;
* un clone de v0 ;
* un générateur Figma-like ;
* un outil de “vibe design” ;
* un générateur de landing pages décoratives.

Asagiri doit devenir :

un orchestrateur d’ingénierie produit capable de transformer un prototype exécutable en système logiciel vérifiable.

Le différenciateur n’est pas la génération visuelle.

Le différenciateur est :

* l’extraction de flows ;
* la génération de contracts ;
* la dérivation de specs ;
* le découpage en tâches ;
* l’orchestration agents ;
* la mesure coût/performance ;
* la validation ;
* la traçabilité.

⸻

3. Objectifs

3.1 Objectifs fonctionnels

Cette évolution doit permettre de :

1. créer un prototype interactif depuis une intention produit ;
2. stocker ce prototype dans le repo ;
3. représenter les parcours utilisateur sous forme de flows structurés ;
4. extraire les actions, états, routes et besoins API depuis les flows ;
5. générer des specs markdown depuis le modèle produit ;
6. générer ou enrichir les tasks Asagiri existantes ;
7. produire des contracts exploitables par les agents ;
8. lier chaque tâche à un flow, une action ou un écran ;
9. générer des tests E2E depuis les flows ;
10. conserver la traçabilité intention → prototype → flow → spec → task → code.

3.2 Non-objectifs

Cette étape ne doit pas chercher à :

* construire un éditeur visuel complet ;
* gérer du drag-and-drop ;
* synchroniser avec Figma ;
* générer un design system complet ;
* produire une application production-ready directement depuis le prototype ;
* remplacer la review humaine ;
* contourner les primitives existantes ;
* exécuter du code non validé sans garde-fou.

⸻

4. Principe central

Le prototype devient une spec exécutable, mais pas la source de vérité unique.

La source de vérité reste le repository.

Pattern obligatoire :

User intent
  ↓
.asagiri/products/<product>/intent.md
  ↓
.asagiri/products/<product>/prototype/
  ↓
.asagiri/products/<product>/flows/*.yaml
  ↓
.asagiri/products/<product>/contracts/*
  ↓
.asagiri/specs/<feature>/
  ↓
Asagiri runtime

Tout artefact généré doit être versionnable, inspectable et modifiable.

⸻

5. Nouvelle arborescence

Ajouter la structure suivante :

.asagiri/
  products/
    <product>/
      intent.md
      product.yaml
      prototype/
        package.json
        index.html
        src/
          App.tsx
          components/
          pages/
          flows/
        README.md
      flows/
        onboarding.flow.yaml
        workspace-management.flow.yaml
      screens/
        dashboard.screen.yaml
        settings.screen.yaml
      contracts/
        api.openapi.yaml
        permissions.yaml
        events.yaml
        analytics.yaml
        observability.yaml
      extraction/
        extracted-model.yaml
        extraction-report.md
      generated-specs/
        requirements.md
        design.md
        tasks.md
      reviews/
        product-review.md
        architecture-review.md

Notes :

* prototype/ contient une mini app exécutable.
* flows/ contient la représentation produit stable.
* contracts/ contient les implications système.
* generated-specs/ contient les specs dérivées.
* extraction/ garde la trace de ce qui a été extrait automatiquement.

⸻

6. Nouvelles commandes CLI

Les commandes doivent utiliser le nouveau branding : asa.

6.1 asa prototype create

Créer un prototype interactif depuis une intention.

Usage :

asa prototype create "SaaS de gestion de workspaces avec invitation d’équipe"

Options :

asa prototype create "..." \
  --product workspace-saas \
  --stack react \
  --style minimal \
  --output .asagiri/products/workspace-saas/prototype

Responsabilités :

1. capturer l’intention ;
2. générer une structure prototype ;
3. créer les pages principales ;
4. créer une navigation fonctionnelle ;
5. simuler les états nécessaires ;
6. documenter les hypothèses ;
7. produire un premier product.yaml.

Critères :

* le prototype doit être lançable localement ;
* aucun backend réel n’est requis ;
* les données mockées doivent être explicitement identifiées ;
* les limites doivent être documentées.

⸻

6.2 asa prototype run

Lancer le prototype localement.

Usage :

asa prototype run workspace-saas

Responsabilités :

* détecter le prototype ;
* installer les dépendances si nécessaire avec confirmation ;
* lancer le serveur local ;
* afficher l’URL ;
* conserver le mode plain compatible CI.

⸻

6.3 asa prototype patch

Modifier un prototype existant depuis une instruction.

Usage :

asa prototype patch workspace-saas "Ajoute un onboarding en 3 étapes avant le dashboard"

Responsabilités :

1. lire l’état actuel du prototype ;
2. identifier les fichiers impactés ;
3. proposer un plan ;
4. appliquer le patch ;
5. mettre à jour les flows si nécessaire ;
6. produire un rapport.

Garde-fou :

* ne jamais réécrire tout le prototype si un patch ciblé suffit ;
* conserver les composants existants ;
* signaler les changements de structure majeurs.

⸻

6.4 asa flows extract

Extraire des flows structurés depuis un prototype.

Usage :

asa flows extract workspace-saas

Responsabilités :

* analyser pages, routes et composants ;
* détecter actions utilisateur ;
* détecter transitions ;
* détecter états loading/error/empty/success ;
* produire .asagiri/products/<product>/flows/*.flow.yaml ;
* produire un rapport d’extraction.

⸻

6.5 asa flows inspect

Inspecter les flows connus d’un produit.

Usage :

asa flows inspect workspace-saas
asa flows inspect workspace-saas --flow onboarding

Sortie attendue :

Product: workspace-saas
Flows: 3
onboarding
  steps: 4
  actions: 5
  unresolved contracts: 2
  risk: medium
workspace-management
  steps: 6
  actions: 9
  unresolved contracts: 4
  risk: high

⸻

6.6 asa contracts extract

Extraire les contracts techniques depuis les flows.

Usage :

asa contracts extract workspace-saas

Responsabilités :

* générer une première version OpenAPI ;
* extraire les permissions ;
* extraire les events métier ;
* extraire les besoins analytics ;
* extraire les besoins observabilité ;
* signaler les incertitudes.

⸻

6.7 asa spec generate-from-product

Générer des specs Asagiri depuis le modèle produit.

Usage :

asa spec generate-from-product workspace-saas

Sorties :

.asagiri/products/workspace-saas/generated-specs/requirements.md
.asagiri/products/workspace-saas/generated-specs/design.md
.asagiri/products/workspace-saas/generated-specs/tasks.md
.asagiri/specs/workspace-saas/
  spec.md
  tasks.yaml
  metadata.yaml

Responsabilités :

* transformer les flows en exigences ;
* transformer les contracts en design technique ;
* transformer les actions en tasks ;
* lier chaque task à une source produit ;
* générer les critères d’acceptation.

⸻

6.8 asa product review

Faire une review produit/architecture avant implémentation.

Usage :

asa product review workspace-saas

La review doit couvrir :

* cohérence UX ;
* flows incomplets ;
* edge cases oubliés ;
* complexité inutile ;
* risques backend ;
* sécurité ;
* permissions ;
* observabilité ;
* coût d’implémentation ;
* dette potentielle.

⸻

7. Format product.yaml

Créer un fichier canonique :

id: workspace-saas
name: Workspace SaaS
status: draft
intent: >
  Permettre à un utilisateur de créer un workspace,
  inviter une équipe et configurer un premier projet.
prototype:
  stack: react
  path: prototype
  status: generated
  mock_data: true
business:
  target_user: technical founder
  value_proposition: reduce setup time
  monetization: subscription
quality:
  requires_auth: true
  requires_a11y_review: true
  requires_security_review: true
sources:
  created_from: user_intent
  created_at: "2026-05-27T00:00:00+02:00"

⸻

8. Format *.flow.yaml

Chaque flow doit être structuré et exploitable.

Exemple :

id: onboarding
name: Onboarding
status: draft
source:
  prototype_path: prototype/src/flows/onboarding.tsx
  extracted_at: "2026-05-27T00:00:00+02:00"
actor:
  type: authenticated_user
  role: owner
entrypoint:
  route: /onboarding
  screen: onboarding_welcome
steps:
  - id: welcome
    screen: onboarding_welcome
    purpose: introduce product value
    actions:
      - id: start_onboarding
        label: Start
        type: navigate
        target: create_workspace
  - id: create_workspace
    screen: create_workspace
    purpose: collect workspace name
    states:
      - idle
      - submitting
      - validation_error
      - success
    actions:
      - id: submit_workspace
        label: Create workspace
        type: command
        command: workspace.create
        requires_api: true
        validation:
          - workspace_name_required
  - id: invite_team
    screen: invite_team
    purpose: invite members
    actions:
      - id: invite_member
        type: command
        command: member.invite
        requires_api: true
      - id: skip_invite
        type: navigate
        target: review
  - id: review
    screen: onboarding_review
    purpose: confirm setup
    actions:
      - id: finish_onboarding
        type: command
        command: onboarding.complete
        requires_api: true
        target: dashboard
exitpoints:
  - route: /dashboard
    condition: onboarding_completed
risks:
  - missing_invitation_error_state
  - unclear_permission_model

⸻

9. Format *.screen.yaml

Un écran doit être représenté indépendamment du framework UI.

id: create_workspace
route: /onboarding/workspace
title: Create workspace
layout: centered_form
components:
  - id: workspace_name_input
    type: text_input
    required: true
    validation:
      - required
      - max_length: 80
  - id: submit_button
    type: button
    action: submit_workspace
states:
  empty: false
  loading: true
  error: true
  success: true
accessibility:
  keyboard_navigation_required: true
  error_messages_required: true

⸻

10. Contracts générés

10.1 API contract

Fichier :

.asagiri/products/<product>/contracts/api.openapi.yaml

Exemple :

openapi: 3.1.0
info:
  title: Workspace SaaS API
  version: 0.1.0
paths:
  /workspaces:
    post:
      operationId: createWorkspace
      summary: Create workspace
      requestBody:
        required: true
      responses:
        "201":
          description: Workspace created
        "400":
          description: Validation error

10.2 Permissions

roles:
  owner:
    permissions:
      - workspace.create
      - workspace.update
      - member.invite
  member:
    permissions:
      - workspace.read

10.3 Events

events:
  - name: workspace.created
    producer: workspace.create
    payload:
      - workspace_id
      - owner_id
  - name: member.invited
    producer: member.invite

10.4 Analytics

events:
  - name: onboarding_started
    source: onboarding.start_onboarding
  - name: workspace_created
    source: onboarding.submit_workspace
  - name: onboarding_completed
    source: onboarding.finish_onboarding

10.5 Observability

metrics:
  - name: workspace_create_duration_ms
  - name: onboarding_completion_rate
logs:
  - name: workspace_create_failed
    level: warn
traces:
  - operation: createWorkspace

⸻

11. Extraction engine

Créer un module Go :

internal/product/
  model.go
  repository.go
  prototype.go
  extractor.go
  flow.go
  screen.go
  contracts.go
  spec_generator.go
  reviewer.go

Interfaces attendues :

type PrototypeGenerator interface {
    Generate(ctx context.Context, req PrototypeRequest) (PrototypeResult, error)
    Patch(ctx context.Context, req PrototypePatchRequest) (PrototypePatchResult, error)
}
type FlowExtractor interface {
    Extract(ctx context.Context, req FlowExtractionRequest) (FlowExtractionResult, error)
}
type ContractExtractor interface {
    Extract(ctx context.Context, product ProductModel) (ContractExtractionResult, error)
}
type ProductSpecGenerator interface {
    Generate(ctx context.Context, req ProductSpecRequest) (ProductSpecResult, error)
}

Le module ne doit pas dépendre directement de la TUI, des agents externes ou du CLI layer.

⸻

12. Génération prototype

12.1 Stack V1 recommandée

Pour la V1, utiliser un prototype React simple généré localement :

* Vite ;
* React ;
* TypeScript ;
* CSS simple ou Tailwind si déjà disponible ;
* données mockées ;
* pas de backend ;
* pas de SSR ;
* pas de dépendance lourde.

Objectif : prototype cliquable, pas application finale.

12.2 Contraintes

Le prototype doit :

* être lisible ;
* être facile à patcher ;
* avoir une structure stable ;
* contenir des noms explicites ;
* séparer pages, components et mock data ;
* annoter les actions importantes ;
* éviter les abstractions prématurées.

Structure minimale :

prototype/
  package.json
  index.html
  src/
    App.tsx
    main.tsx
    data/mock.ts
    pages/
    components/
    flows/
    styles.css

⸻

13. Lien avec le moteur existant

La nouvelle couche produit ne doit pas remplacer le workflow Asagiri existant.

Elle doit alimenter les primitives existantes.

Mapping cible :

product intent      → asa prototype create
prototype           → asa flows extract
flows               → asa contracts extract
contracts           → asa spec generate-from-product
spec/tasks          → asa work
implementation      → asa verify / review / report

Les tâches générées doivent utiliser le format canonique existant.

Exemple task dérivée :

id: task-004
title: Implement workspace creation endpoint
feature: workspace-saas
status: pending
risk: medium
type: backend_feature
source:
  product: workspace-saas
  flow: onboarding
  step: create_workspace
  action: submit_workspace
scope:
  allowed_paths:
    - src/Workspace/**
    - tests/Workspace/**
    - config/routes/**
acceptance:
  - POST /workspaces creates a workspace
  - validation error returned when name is missing
  - owner permission is required
  - workspace.created event is emitted
validation:
  commands:
    - composer test -- tests/Workspace
    - composer phpstan
agents:
  implementer: cursor
  reviewer: codex
  enricher: ollama

⸻

14. Intégration cost-aware

Toute génération ou extraction doit rester compatible avec la couche cost/performance.

14.1 Estimation

Ajouter :

asa prototype create "..." --estimate-only
asa flows extract workspace-saas --estimate-only
asa contracts extract workspace-saas --estimate-only
asa spec generate-from-product workspace-saas --estimate-only

Chaque estimation doit afficher :

* étapes locales ;
* étapes modèle ;
* tokens estimés ;
* coût estimé ;
* durée estimée ;
* niveau de risque ;
* justification du modèle choisi.

14.2 Local-first

Les étapes suivantes doivent être locales par défaut :

* scan des fichiers prototype ;
* extraction routes simples ;
* extraction composants ;
* parsing YAML ;
* génération de rapports ;
* validation schema ;
* comparaison diff.

Les modèles doivent être utilisés surtout pour :

* génération initiale du prototype ;
* interprétation produit ;
* extraction ambiguë ;
* review UX/architecture ;
* transformation en specs lisibles.

⸻

15. Validation & tests

Ajouter des tests pour :

15.1 Unit tests

* parsing product.yaml ;
* parsing *.flow.yaml ;
* parsing *.screen.yaml ;
* validation contracts ;
* génération tasks depuis flows ;
* mapping actions → API operations ;
* détection flows incomplets.

15.2 Golden tests

Créer des fixtures :

testdata/product-layer/
  simple-onboarding/
  workspace-management/
  billing-flow/

Chaque fixture doit vérifier :

* flows générés ;
* contracts générés ;
* specs générées ;
* tasks générées ;
* rapports générés.

15.3 Integration tests

Tester :

asa prototype create "workspace onboarding" --product test-product
asa flows extract test-product
asa contracts extract test-product
asa spec generate-from-product test-product
asa work "develop test-product" --plan-only

15.4 Non-régression

Les commandes existantes doivent continuer à fonctionner :

asa work "develop billing-v2" --plan-only
asa continue
asa next
asa estimate billing-v2
asa verify billing-v2

⸻

16. Sécurité

16.1 Prototype sandbox

Le prototype généré ne doit pas :

* lire des secrets ;
* accéder à .env ;
* faire des appels réseau externes sans confirmation ;
* écrire en dehors de son dossier ;
* installer des dépendances sans confirmation.

16.2 Dépendances

Toute dépendance ajoutée au prototype doit être :

* listée dans le rapport ;
* justifiée ;
* limitée ;
* compatible avec un usage open source.

16.3 Contracts sensibles

Les flows doivent signaler les actions sensibles :

* paiement ;
* données personnelles ;
* authentification ;
* permissions ;
* suppression ;
* export ;
* changement de rôle ;
* migration de données.

Ces actions doivent forcer :

* review humaine ;
* review sécurité ;
* validation explicite.

⸻

17. UX terminal attendue

Exemple :

Asagiri Product Layer
─────────────────────
Product: workspace-saas
Intent:  SaaS de gestion de workspaces avec invitation d’équipe
Plan
────
1. Generate interactive prototype       model: cloud_fast   risk: low
2. Extract user flows                   local               risk: low
3. Extract system contracts             local+model         risk: medium
4. Generate Asagiri specs               model               risk: medium
5. Create implementation tasks          local               risk: low
Estimated cost: €0.12
Estimated time: 3m40s
Budget status: OK
Proceed? [y/N]

Après extraction :

Flow extraction complete
────────────────────────
Product: workspace-saas
Flows:   3
Screens: 9
Actions: 18
API candidates: 7
Permission candidates: 5
Analytics events: 6
Warnings:
- invite_member has no explicit error state
- delete_workspace is destructive and requires confirmation policy
- billing flow references plan selection but no payment state was found

⸻

18. Documentation à mettre à jour

Mettre à jour ou créer :

docs-site/content/docs/concepts/executable-product-layer.mdx
docs-site/content/docs/concepts/prototype-as-spec.mdx
docs-site/content/docs/cli/prototype.mdx
docs-site/content/docs/cli/flows.mdx
docs-site/content/docs/cli/contracts.mdx
docs-site/content/docs/workflows/product-to-code.mdx
docs-site/content/docs/reference/product-schema.mdx
docs-site/content/docs/reference/flow-schema.mdx
docs-site/content/docs/reference/screen-schema.mdx

La documentation doit expliquer :

* pourquoi le prototype n’est pas la production ;
* pourquoi les flows sont la couche stable ;
* comment les contracts sont dérivés ;
* comment les specs sont générées ;
* comment garder le contrôle ;
* quelles sont les limites.

Ton obligatoire : technique, transparent, sans promesse magique.

⸻

19. Critères d’acceptation

Cette évolution est acceptable si :

* asa prototype create crée un prototype local lançable ;
* asa prototype run lance le prototype ;
* asa prototype patch modifie un prototype existant sans tout réécrire ;
* asa flows extract génère au moins un flow YAML valide ;
* asa flows inspect affiche les flows et risques ;
* asa contracts extract génère contracts API/permissions/events/analytics/observability ;
* asa spec generate-from-product génère requirements/design/tasks ;
* les tasks générées restent compatibles avec asa work ;
* les artefacts sont versionnables ;
* les décisions sont inspectables ;
* les estimations coût/tokens sont disponibles ;
* les commandes ont un fallback plain text ;
* les tests unitaires/golden/integration critiques passent ;
* aucune dépendance lourde ou non justifiée n’est ajoutée au moteur ;
* aucune feature existante n’est cassée.

⸻

20. Découpage d’implémentation recommandé

Phase 1 — Product model minimal

* ajouter internal/product ;
* définir Product, Flow, Screen, Contract ;
* ajouter validation YAML ;
* ajouter repository local .asagiri/products ;
* ajouter tests unitaires.

Phase 2 — Prototype scaffold

* ajouter asa prototype create ;
* générer prototype Vite/React minimal ;
* stocker intent et product.yaml ;
* ajouter asa prototype run ;
* ajouter golden fixture.

Phase 3 — Flow extraction V1

* scanner routes/pages/components ;
* générer flows simples ;
* ajouter asa flows extract ;
* ajouter asa flows inspect ;
* produire extraction-report.

Phase 4 — Contracts extraction V1

* générer OpenAPI draft ;
* générer permissions/events/analytics/observability ;
* identifier incertitudes ;
* ajouter asa contracts extract.

Phase 5 — Spec generation

* générer requirements/design/tasks ;
* mapper flows → tasks ;
* rendre compatible asa work ;
* ajouter asa spec generate-from-product.

Phase 6 — Review & cost-aware

* ajouter asa product review ;
* brancher estimation tokens/coût ;
* ajouter rapports ;
* ajouter docs.

Phase 7 — Hardening

* ajouter tests integration ;
* stabiliser schemas ;
* vérifier non-régression CLI ;
* documenter limites ;
* préparer exemples open source.

⸻

21. Risques et mitigations

21.1 Risque : devenir un builder UI générique

Mitigation :

* limiter la V1 au prototype exploitable ;
* privilégier flows/contracts/specs ;
* ne pas ajouter d’éditeur visuel ;
* documenter que le prototype n’est pas la production.

21.2 Risque : specs générées trop vagues

Mitigation :

* lier chaque exigence à un flow/action ;
* générer des critères d’acceptation concrets ;
* signaler les incertitudes ;
* demander review humaine avant tasks.

21.3 Risque : extraction incorrecte

Mitigation :

* conserver extraction-report ;
* permettre correction manuelle des YAML ;
* valider schema ;
* golden tests.

21.4 Risque : coût modèle excessif

Mitigation :

* estimation obligatoire ;
* local-first ;
* --estimate-only ;
* budgets ;
* contexte réduit.

21.5 Risque : complexité architecturelle

Mitigation :

* isoler internal/product ;
* ne pas coupler au moteur workflow ;
* alimenter les primitives existantes ;
* éviter les dépendances UI dans le core.

⸻

22. Mission Cursor

Implémenter l’Executable Product Layer dans Asagiri.

Contraintes :

* respecter le branding public Asagiri / CLI asa ;
* ne pas réintroduire les anciens noms publics ;
* ne pas casser les commandes existantes ;
* ne pas transformer Asagiri en clone de builder UI ;
* privilégier artefacts locaux, versionnables et inspectables ;
* garder le moteur déterministe ;
* garder le local-first ;
* garder les estimations coût/tokens ;
* ajouter les tests avant stabilisation finale ;
* documenter les limites.

Résultat attendu :

asa prototype create "SaaS de gestion de workspaces" --product workspace-saas
asa prototype run workspace-saas
asa flows extract workspace-saas
asa flows inspect workspace-saas
asa contracts extract workspace-saas
asa spec generate-from-product workspace-saas
asa work "develop workspace-saas" --plan-only

Le dernier appel doit produire un plan d’exécution cohérent basé sur les tasks générées depuis les flows produit.

⸻

23. Business Intent Layer

23.1 Objectif

Ajouter à Asagiri une couche explicite de modélisation métier afin que le moteur ne raisonne plus uniquement en features et tâches techniques, mais également en :

* objectifs métier ;
* flows produit ;
* métriques de succès ;
* contraintes business ;
* impact utilisateur ;
* coût de delivery ;
* impact infrastructure.

Cette couche doit permettre :

Business intent
  ↓
Product flows
  ↓
Architecture implications
  ↓
Implementation tasks
  ↓
Validation business-aware

Le but n’est pas d’ajouter une couche “agile” ou bureaucratique.

Le but est de rendre les décisions produit et métier exploitables par le moteur Asagiri.

⸻

23.2 Nouveau concept : Business Intent

Créer un nouveau fichier canonique :

.asagiri/products/<product>/business.yaml

Exemple :

objective:
  primary: reduce onboarding friction
  secondary:
    - reduce support requests
    - improve activation rate
target_users:
  - technical_founder
  - small_team
success_metrics:
  - id: onboarding_completion_rate
    target: ">=70%"
  - id: avg_setup_duration_minutes
    target: "<=5"
  - id: workspace_creation_success_rate
    target: ">=99%"
constraints:
  - low_operational_cost
  - mobile_friendly
  - email_verification_required
business_risk:
  level: medium
  reasons:
    - onboarding critical for conversion
monetization:
  model: subscription
  activation_event: onboarding_completed
expected_scale:
  users:
    first_year: 10000
  workspaces:
    first_year: 25000
compliance:
  - gdpr
observability_requirements:
  - onboarding funnel
  - invite delivery monitoring
  - setup duration tracking

⸻

23.3 Flows comme unité centrale

Le moteur Asagiri doit progressivement évoluer vers un modèle flow-centric.

Aujourd’hui :

Feature
  ↓
Tasks

Évolution cible :

Business intent
  ↓
Product flows
  ↓
Flow actions
  ↓
System contracts
  ↓
Implementation tasks

Chaque flow devient une primitive de haut niveau.

Chaque flow doit pouvoir être relié à :

* un objectif métier ;
* des métriques ;
* des contrats système ;
* des risques ;
* des événements analytics ;
* des contraintes observabilité ;
* des permissions.

⸻

23.4 Enrichissement du format *.flow.yaml

Ajouter les sections suivantes.

Exemple :

business:
  objective: reduce onboarding friction
  criticality: high
  monetization_impact: high
metrics:
  - onboarding_completion_rate
  - avg_setup_duration_minutes
architecture_implications:
  - async_email_delivery
  - audit_logs
  - rate_limiting
  - retry_policy
  - invitation_expiration
observability:
  traces:
    - onboarding.start
    - onboarding.complete
  metrics:
    - onboarding_step_duration
    - onboarding_dropoff_rate
  logs:
    - invitation_failed
security:
  requires_authentication: true
  sensitive_actions:
    - invite_member
cost_profile:
  expected_complexity: medium
  infrastructure_cost_risk: low

⸻

23.5 Nouveau moteur de dérivation architecture

Créer un module :

internal/product/derivation/
  architecture.go
  analytics.go
  observability.go
  permissions.go
  infra.go

Objectif : dériver automatiquement des implications système depuis les flows.

Exemple :

Flow action:
invite_member

Doit pouvoir dériver :

* besoin email provider ;
* retry queue ;
* rate limiting ;
* anti-abuse ;
* audit log ;
* expiration token ;
* observabilité delivery ;
* métriques conversion.

⸻

23.6 Nouveau concept : Flow Review

Ajouter :

asa flows review workspace-saas

La review ne doit pas uniquement porter sur le code ou la structure UI.

Elle doit également analyser :

* cohérence métier ;
* friction utilisateur ;
* flows incomplets ;
* absence d’états erreur ;
* manque de métriques ;
* observabilité insuffisante ;
* sécurité insuffisante ;
* permissions incohérentes ;
* complexité excessive ;
* coût d’implémentation disproportionné.

Sortie attendue :

Flow Review
───────────
Flow: onboarding
Business alignment:
✓ onboarding objective clearly mapped
✓ activation metrics defined
Warnings:
- no analytics event for invite abandonment
- no retry strategy for invitation delivery
- no rate limit defined for invite_member
- onboarding success metric exists but no dashboard contract found
Risk:
medium

⸻

23.7 Metrics-driven engineering

Chaque flow critique doit être associé à des métriques.

Le moteur doit pouvoir vérifier que :

* les analytics events existent ;
* les métriques existent ;
* les traces existent ;
* les dashboards nécessaires sont définis ;
* les alertes critiques sont prévues.

Exemple :

metrics:
  onboarding_completion_rate:
    target: ">=70%"
    source: analytics
    dashboard_required: true
  invitation_delivery_success_rate:
    target: ">=99%"
    source: backend_events
    alert_threshold: "<95%"

⸻

23.8 Product-to-System Projection

Créer une étape explicite de projection système.

Commande :

asa architecture derive workspace-saas

Responsabilités :

* dériver besoins API ;
* dériver besoins async ;
* dériver besoins auth ;
* dériver besoins observabilité ;
* dériver besoins analytics ;
* dériver besoins infra ;
* dériver besoins sécurité ;
* produire un rapport architecture.

Sortie :

Architecture Projection
───────────────────────
Detected requirements:
API:
- workspace creation endpoint
- invitation endpoint
- onboarding status endpoint
Async:
- invitation email queue
- retry worker
Security:
- invitation token expiration
- owner-only invite permission
- rate limiting required
Observability:
- onboarding funnel metrics
- invitation delivery traces
- workspace creation latency
Infrastructure:
- queue system required
- email provider required
- metrics backend required

⸻

23.9 Flow-first task generation

Les tâches doivent être dérivées depuis les flows et non seulement depuis les specs markdown.

Pattern cible :

Flow
  ↓
Actions
  ↓
Contracts
  ↓
Implementation tasks

Chaque task doit référencer explicitement :

source:
  product: workspace-saas
  flow: onboarding
  step: invite_team
  action: invite_member
  business_objective: reduce onboarding friction

⸻

23.10 Nouveau mode de raisonnement

Asagiri doit progressivement évoluer d’un moteur principalement feature-centric vers un moteur flow-centric.

Avant :

Feature → tasks → implementation

Après :

Business intent
  ↓
Executable flows
  ↓
Architecture contracts
  ↓
Implementation tasks
  ↓
Validation business-aware

⸻

23.11 Critères d’acceptation

Cette évolution est acceptable si :

* un business.yaml valide peut être créé ;
* les flows peuvent référencer des objectifs métier ;
* les flows peuvent référencer des métriques ;
* asa architecture derive produit des implications cohérentes ;
* asa flows review détecte les gaps métier/observabilité/sécurité ;
* les tasks générées référencent explicitement les flows ;
* les métriques critiques peuvent être reliées aux analytics/contracts ;
* les projections système restent inspectables ;
* les primitives existantes continuent de fonctionner ;
* aucune dépendance à un framework agile externe n’est introduite.

⸻

23.12 Résumé

Cette évolution ajoute une couche business-aware au Product Layer.

Le moteur Asagiri ne doit plus uniquement raisonner en :

feature → task → code

Mais en :

business intent
  ↓
product flow
  ↓
system implications
  ↓
implementation
  ↓
measurable outcome

Le flow devient l’unité centrale reliant :

* produit ;
* architecture ;
* observabilité ;
* analytics ;
* sécurité ;
* implémentation ;
* validation.

⸻

24. Résumé exécutable

Cette évolution ajoute à Asagiri une couche produit exécutable :

Intent
  ↓
Prototype interactif
  ↓
Flows structurés
  ↓
Contracts système
  ↓
Specs générées
  ↓
Tasks Asagiri
  ↓
Agents
  ↓
Validation

Principe clé :

Le prototype sert à clarifier le produit. Les flows servent à structurer l’ingénierie. Les contracts servent à cadrer le système. Les specs servent à piloter les agents.

Asagiri ne doit pas promettre de la magie IA. Il doit fournir une chaîne de transformation contrôlée, inspectable et industrialisable.