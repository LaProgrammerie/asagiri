# Spec — Asagiri Executable Product Layer

Objectif : ajouter à **Asagiri** une couche de spécification produit exécutable inspirée du paradigme “prototype interactif comme spec vivante”, sans transformer Asagiri en clone de builder UI.

Cette évolution doit permettre à Asagiri de passer de :

```text
Intent → spec markdown → tasks → agents → code
```

à :

```text
Intent
→ prototype interactif
→ flows exécutables
→ contracts dérivés
→ specs générées
→ tasks agents
→ implémentation vérifiée
```

Le but est de faire d’Asagiri un système capable de transformer une intention produit en pipeline d’ingénierie industrialisable, mesurable et contrôlé.

---

## 1. Contexte

Asagiri est déjà un orchestrateur local de workflows de développement agentique.

Le système existant repose sur :

* une CLI Go ;
* une couche intentionnelle `work / continue / next / inbox / sync` ;
* des primitives bas niveau `spec / plan / enrich / dev / verify / review / pr` ;
* un moteur déterministe ;
* des agents interchangeables ;
* des worktrees Git isolés ;
* une validation externe obligatoire ;
* une stratégie local-first ;
* une estimation coût/tokens/temps ;
* des rapports vérifiables.

Cette spec ajoute une nouvelle couche située **avant** les specs et tasks classiques : la couche produit exécutable.

Elle doit permettre de matérialiser une idée sous forme de prototype interactif, puis d’en extraire des flows, contracts, specs et tâches.

---

## 2. Positionnement

Asagiri ne doit pas devenir :

* un clone de Lovable ;
* un clone de v0 ;
* un générateur Figma-like ;
* un outil de “vibe design” ;
* un générateur de landing pages décoratives.

Asagiri doit devenir :

> un orchestrateur d’ingénierie produit capable de transformer un prototype exécutable en système logiciel vérifiable.

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

---

## 3. Objectifs

### 3.1 Objectifs fonctionnels

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

### 3.2 Non-objectifs

Cette étape ne doit pas chercher à :

* construire un éditeur visuel complet ;
* gérer du drag-and-drop ;
* synchroniser avec Figma ;
* générer un design system complet ;
* produire une application production-ready directement depuis le prototype ;
* remplacer la review humaine ;
* contourner les primitives existantes ;
* exécuter du code non validé sans garde-fou.

---

## 4. Principe central

Le prototype devient une **spec exécutable**, mais pas la source de vérité unique.

La source de vérité reste le repository.

Pattern obligatoire :

```text
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
```

Tout artefact généré doit être versionnable, inspectable et modifiable.

---

## 5. Nouvelle arborescence

Ajouter la structure suivante :

```text
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
```

Notes :

* `prototype/` contient une mini app exécutable.
* `flows/` contient la représentation produit stable.
* `contracts/` contient les implications système.
* `generated-specs/` contient les specs dérivées.
* `extraction/` garde la trace de ce qui a été extrait automatiquement.

---

## 6. Nouvelles commandes CLI

Les commandes doivent utiliser le nouveau branding : `asa`.

### 6.1 `asa prototype create`

Créer un prototype interactif depuis une intention.

Usage :

```bash
asa prototype create "SaaS de gestion de workspaces avec invitation d’équipe"
```

Options :

```bash
asa prototype create "..." \
  --product workspace-saas \
  --stack react \
  --style minimal \
  --output .asagiri/products/workspace-saas/prototype
```

Responsabilités :

1. capturer l’intention ;
2. générer une structure prototype ;
3. créer les pages principales ;
4. créer une navigation fonctionnelle ;
5. simuler les états nécessaires ;
6. documenter les hypothèses ;
7. produire un premier `product.yaml`.

Critères :

* le prototype doit être lançable localement ;
* aucun backend réel n’est requis ;
* les données mockées doivent être explicitement identifiées ;
* les limites doivent être documentées.

---

### 6.2 `asa prototype run`

Lancer le prototype localement.

Usage :

```bash
asa prototype run workspace-saas
```

Responsabilités :

* détecter le prototype ;
* installer les dépendances si nécessaire avec confirmation ;
* lancer le serveur local ;
* afficher l’URL ;
* conserver le mode plain compatible CI.

---

### 6.3 `asa prototype patch`

Modifier un prototype existant depuis une instruction.

Usage :

```bash
asa prototype patch workspace-saas "Ajoute un onboarding en 3 étapes avant le dashboard"
```

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

---

### 6.4 `asa flows extract`

Extraire des flows structurés depuis un prototype.

Usage :

```bash
asa flows extract workspace-saas
```

Responsabilités :

* analyser pages, routes et composants ;
* détecter actions utilisateur ;
* détecter transitions ;
* détecter états loading/error/empty/success ;
* produire `.asagiri/products/<product>/flows/*.flow.yaml` ;
* produire un rapport d’extraction.

---

### 6.5 `asa flows inspect`

Inspecter les flows connus d’un produit.

Usage :

```bash
asa flows inspect workspace-saas
asa flows inspect workspace-saas --flow onboarding
```

Sortie attendue :

```text
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
```

---

### 6.6 `asa contracts extract`

Extraire les contracts techniques depuis les flows.

Usage :

```bash
asa contracts extract workspace-saas
```

Responsabilités :

* générer une première version OpenAPI ;
* extraire les permissions ;
* extraire les events métier ;
* extraire les besoins analytics ;
* extraire les besoins observabilité ;
* signaler les incertitudes.

---

### 6.7 `asa spec generate-from-product`

Générer des specs Asagiri depuis le modèle produit.

Usage :

```bash
asa spec generate-from-product workspace-saas
```

Sorties :

```text
.asagiri/products/workspace-saas/generated-specs/requirements.md
.asagiri/products/workspace-saas/generated-specs/design.md
.asagiri/products/workspace-saas/generated-specs/tasks.md
.asagiri/specs/workspace-saas/
  spec.md
  tasks.yaml
  metadata.yaml
```

Responsabilités :

* transformer les flows en exigences ;
* transformer les contracts en design technique ;
* transformer les actions en tasks ;
* lier chaque task à une source produit ;
* générer les critères d’acceptation.

---

### 6.8 `asa product review`

Faire une review produit/architecture avant implémentation.

Usage :

```bash
asa product review workspace-saas
```

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

---

## 7. Format `product.yaml`

Créer un fichier canonique :

```yaml
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
```

---

## 8. Format `*.flow.yaml`

Chaque flow doit être structuré et exploitable.

Exemple :

```yaml
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
```

---

## 9. Format `*.screen.yaml`

Un écran doit être représenté indépendamment du framework UI.

```yaml
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
```

---

## 10. Contracts générés

### 10.1 API contract

Fichier :

```text
.asagiri/products/<product>/contracts/api.openapi.yaml
```

Exemple :

```yaml
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
```

### 10.2 Permissions

```yaml
roles:
  owner:
    permissions:
      - workspace.create
      - workspace.update
      - member.invite
  member:
    permissions:
      - workspace.read
```

### 10.3 Events

```yaml
events:
  - name: workspace.created
    producer: workspace.create
    payload:
      - workspace_id
      - owner_id
  - name: member.invited
    producer: member.invite
```

### 10.4 Analytics

```yaml
events:
  - name: onboarding_started
    source: onboarding.start_onboarding
  - name: workspace_created
    source: onboarding.submit_workspace
  - name: onboarding_completed
    source: onboarding.finish_onboarding
```

### 10.5 Observability

```yaml
metrics:
  - name: workspace_create_duration_ms
  - name: onboarding_completion_rate
logs:
  - name: workspace_create_failed
    level: warn
traces:
  - operation: createWorkspace
```

---

## 11. Extraction engine

Créer un module Go :

```text
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
```

Interfaces attendues :

```go
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
```

Le module ne doit pas dépendre directement de la TUI, des agents externes ou du CLI layer.

---

## 12. Génération prototype

### 12.1 Stack V1 recommandée

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

### 12.2 Contraintes

Le prototype doit :

* être lisible ;
* être facile à patcher ;
* avoir une structure stable ;
* contenir des noms explicites ;
* séparer pages, components et mock data ;
* annoter les actions importantes ;
* éviter les abstractions prématurées.

Structure minimale :

```text
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
```

---

## 13. Lien avec le moteur existant

La nouvelle couche produit ne doit pas remplacer le workflow Asagiri existant.

Elle doit alimenter les primitives existantes.

Mapping cible :

```text
product intent      → asa prototype create
prototype           → asa flows extract
flows               → asa contracts extract
contracts           → asa spec generate-from-product
spec/tasks          → asa work
implementation      → asa verify / review / report
```

Les tâches générées doivent utiliser le format canonique existant.

Exemple task dérivée :

```yaml
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
```

---

## 14. Intégration cost-aware

Toute génération ou extraction doit rester compatible avec la couche cost/performance.

### 14.1 Estimation

Ajouter :

```bash
asa prototype create "..." --estimate-only
asa flows extract workspace-saas --estimate-only
asa contracts extract workspace-saas --estimate-only
asa spec generate-from-product workspace-saas --estimate-only
```

Chaque estimation doit afficher :

* étapes locales ;
* étapes modèle ;
* tokens estimés ;
* coût estimé ;
* durée estimée ;
* niveau de risque ;
* justification du modèle choisi.

### 14.2 Local-first

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

---

## 15. Validation & tests

Ajouter des tests pour :

### 15.1 Unit tests

* parsing `product.yaml` ;
* parsing `*.flow.yaml` ;
* parsing `*.screen.yaml` ;
* validation contracts ;
* génération tasks depuis flows ;
* mapping actions → API operations ;
* détection flows incomplets.

### 15.2 Golden tests

Créer des fixtures :

```text
testdata/product-layer/
  simple-onboarding/
  workspace-management/
  billing-flow/
```

Chaque fixture doit vérifier :

* flows générés ;
* contracts générés ;
* specs générées ;
* tasks générées ;
* rapports générés.

### 15.3 Integration tests

Tester :

```bash
asa prototype create "workspace onboarding" --product test-product
asa flows extract test-product
asa contracts extract test-product
asa spec generate-from-product test-product
asa work "develop test-product" --plan-only
```

### 15.4 Non-régression

Les commandes existantes doivent continuer à fonctionner :

```bash
asa work "develop billing-v2" --plan-only
asa continue
asa next
asa estimate billing-v2
asa verify billing-v2
```

---

## 16. Sécurité

### 16.1 Prototype sandbox

Le prototype généré ne doit pas :

* lire des secrets ;
* accéder à `.env` ;
* faire des appels réseau externes sans confirmation ;
* écrire en dehors de son dossier ;
* installer des dépendances sans confirmation.

### 16.2 Dépendances

Toute dépendance ajoutée au prototype doit être :

* listée dans le rapport ;
* justifiée ;
* limitée ;
* compatible avec un usage open source.

### 16.3 Contracts sensibles

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

---

## 17. UX terminal attendue

Exemple :

```text
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
```

Après extraction :

```text
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
```

---

## 18. Documentation à mettre à jour

Mettre à jour ou créer :

```text
docs-site/content/docs/concepts/executable-product-layer.mdx
docs-site/content/docs/concepts/prototype-as-spec.mdx
docs-site/content/docs/cli/prototype.mdx
docs-site/content/docs/cli/flows.mdx
docs-site/content/docs/cli/contracts.mdx
docs-site/content/docs/workflows/product-to-code.mdx
docs-site/content/docs/reference/product-schema.mdx
docs-site/content/docs/reference/flow-schema.mdx
docs-site/content/docs/reference/screen-schema.mdx
```

La documentation doit expliquer :

* pourquoi le prototype n’est pas la production ;
* pourquoi les flows sont la couche stable ;
* comment les contracts sont dérivés ;
* comment les specs sont générées ;
* comment garder le contrôle ;
* quelles sont les limites.

Ton obligatoire : technique, transparent, sans promesse magique.

---

## 19. Critères d’acceptation

Cette évolution est acceptable si :

* `asa prototype create` crée un prototype local lançable ;
* `asa prototype run` lance le prototype ;
* `asa prototype patch` modifie un prototype existant sans tout réécrire ;
* `asa flows extract` génère au moins un flow YAML valide ;
* `asa flows inspect` affiche les flows et risques ;
* `asa contracts extract` génère contracts API/permissions/events/analytics/observability ;
* `asa spec generate-from-product` génère requirements/design/tasks ;
* les tasks générées restent compatibles avec `asa work` ;
* les artefacts sont versionnables ;
* les décisions sont inspectables ;
* les estimations coût/tokens sont disponibles ;
* les commandes ont un fallback plain text ;
* les tests unitaires/golden/integration critiques passent ;
* aucune dépendance lourde ou non justifiée n’est ajoutée au moteur ;
* aucune feature existante n’est cassée.

---

## 20. Découpage d’implémentation recommandé

### Phase 1 — Product model minimal

* ajouter `internal/product` ;
* définir `Product`, `Flow`, `Screen`, `Contract` ;
* ajouter validation YAML ;
* ajouter repository local `.asagiri/products` ;
* ajouter tests unitaires.

### Phase 2 — Prototype scaffold

* ajouter `asa prototype create` ;
* générer prototype Vite/React minimal ;
* stocker intent et product.yaml ;
* ajouter `asa prototype run` ;
* ajouter golden fixture.

### Phase 3 — Flow extraction V1

* scanner routes/pages/components ;
* générer flows simples ;
* ajouter `asa flows extract` ;
* ajouter `asa flows inspect` ;
* produire extraction-report.

### Phase 4 — Contracts extraction V1

* générer OpenAPI draft ;
* générer permissions/events/analytics/observability ;
* identifier incertitudes ;
* ajouter `asa contracts extract`.

### Phase 5 — Spec generation

* générer requirements/design/tasks ;
* mapper flows → tasks ;
* rendre compatible `asa work` ;
* ajouter `asa spec generate-from-product`.

### Phase 6 — Review & cost-aware

* ajouter `asa product review` ;
* brancher estimation tokens/coût ;
* ajouter rapports ;
* ajouter docs.

### Phase 7 — Hardening

* ajouter tests integration ;
* stabiliser schemas ;
* vérifier non-régression CLI ;
* documenter limites ;
* préparer exemples open source.

---

## 21. Risques et mitigations

### 21.1 Risque : devenir un builder UI générique

Mitigation :

* limiter la V1 au prototype exploitable ;
* privilégier flows/contracts/specs ;
* ne pas ajouter d’éditeur visuel ;
* documenter que le prototype n’est pas la production.

### 21.2 Risque : specs générées trop vagues

Mitigation :

* lier chaque exigence à un flow/action ;
* générer des critères d’acceptation concrets ;
* signaler les incertitudes ;
* demander review humaine avant tasks.

### 21.3 Risque : extraction incorrecte

Mitigation :

* conserver extraction-report ;
* permettre correction manuelle des YAML ;
* valider schema ;
* golden tests.

### 21.4 Risque : coût modèle excessif

Mitigation :

* estimation obligatoire ;
* local-first ;
* `--estimate-only` ;
* budgets ;
* contexte réduit.

### 21.5 Risque : complexité architecturelle

Mitigation :

* isoler `internal/product` ;
* ne pas coupler au moteur workflow ;
* alimenter les primitives existantes ;
* éviter les dépendances UI dans le core.

---

## 22. Mission Cursor

Implémenter l’**Executable Product Layer** dans Asagiri.

Contraintes :

* respecter le branding public `Asagiri` / CLI `asa` ;
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

```bash
asa prototype create "SaaS de gestion de workspaces" --product workspace-saas
asa prototype run workspace-saas
asa flows extract workspace-saas
asa flows inspect workspace-saas
asa contracts extract workspace-saas
asa spec generate-from-product workspace-saas
asa work "develop workspace-saas" --plan-only
```

Le dernier appel doit produire un plan d’exécution cohérent basé sur les tasks générées depuis les flows produit.

---

## 23. Business Intent Layer

### 23.1 Objectif

Ajouter à Asagiri une couche explicite de modélisation métier afin que le moteur ne raisonne plus uniquement en features et tâches techniques, mais également en :

* objectifs métier ;
* flows produit ;
* métriques de succès ;
* contraintes business ;
* impact utilisateur ;
* coût de delivery ;
* impact infrastructure.

Cette couche doit permettre :

```text
Business intent
  ↓
Product flows
  ↓
Architecture implications
  ↓
Implementation tasks
  ↓
Validation business-aware
```

Le but n’est pas d’ajouter une couche “agile” ou bureaucratique.

Le but est de rendre les décisions produit et métier exploitables par le moteur Asagiri.

---

### 23.2 Nouveau concept : Business Intent

Créer un nouveau fichier canonique :

```text
.asagiri/products/<product>/business.yaml
```

Exemple :

```yaml
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
```

---

### 23.3 Flows comme unité centrale

Le moteur Asagiri doit progressivement évoluer vers un modèle flow-centric.

Aujourd’hui :

```text
Feature
  ↓
Tasks
```

Évolution cible :

```text
Business intent
  ↓
Product flows
  ↓
Flow actions
  ↓
System contracts
  ↓
Implementation tasks
```

Chaque flow devient une primitive de haut niveau.

Chaque flow doit pouvoir être relié à :

* un objectif métier ;
* des métriques ;
* des contrats système ;
* des risques ;
* des événements analytics ;
* des contraintes observabilité ;
* des permissions.

---

### 23.4 Enrichissement du format `*.flow.yaml`

Ajouter les sections suivantes.

Exemple :

```yaml
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
```

---

### 23.5 Nouveau moteur de dérivation architecture

Créer un module :

```text
internal/product/derivation/
  architecture.go
  analytics.go
  observability.go
  permissions.go
  infra.go
```

Objectif : dériver automatiquement des implications système depuis les flows.

Exemple :

```text
Flow action:
invite_member
```

Doit pouvoir dériver :

* besoin email provider ;
* retry queue ;
* rate limiting ;
* anti-abuse ;
* audit log ;
* expiration token ;
* observabilité delivery ;
* métriques conversion.

---

### 23.6 Nouveau concept : Flow Review

Ajouter :

```bash
asa flows review workspace-saas
```

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

```text
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
```

---

### 23.7 Metrics-driven engineering

Chaque flow critique doit être associé à des métriques.

Le moteur doit pouvoir vérifier que :

* les analytics events existent ;
* les métriques existent ;
* les traces existent ;
* les dashboards nécessaires sont définis ;
* les alertes critiques sont prévues.

Exemple :

```yaml
metrics:
  onboarding_completion_rate:
    target: ">=70%"
    source: analytics
    dashboard_required: true

  invitation_delivery_success_rate:
    target: ">=99%"
    source: backend_events
    alert_threshold: "<95%"
```

---

### 23.8 Product-to-System Projection

Créer une étape explicite de projection système.

Commande :

```bash
asa architecture derive workspace-saas
```

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

```text
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
```

---

### 23.9 Flow-first task generation

Les tâches doivent être dérivées depuis les flows et non seulement depuis les specs markdown.

Pattern cible :

```text
Flow
  ↓
Actions
  ↓
Contracts
  ↓
Implementation tasks
```

Chaque task doit référencer explicitement :

```yaml
source:
  product: workspace-saas
  flow: onboarding
  step: invite_team
  action: invite_member
  business_objective: reduce onboarding friction
```

---

### 23.10 Nouveau mode de raisonnement

Asagiri doit progressivement évoluer d’un moteur principalement feature-centric vers un moteur flow-centric.

Avant :

```text
Feature → tasks → implementation
```

Après :

```text
Business intent
  ↓
Executable flows
  ↓
Architecture contracts
  ↓
Implementation tasks
  ↓
Validation business-aware
```

---

### 23.11 Critères d’acceptation

Cette évolution est acceptable si :

* un `business.yaml` valide peut être créé ;
* les flows peuvent référencer des objectifs métier ;
* les flows peuvent référencer des métriques ;
* `asa architecture derive` produit des implications cohérentes ;
* `asa flows review` détecte les gaps métier/observabilité/sécurité ;
* les tasks générées référencent explicitement les flows ;
* les métriques critiques peuvent être reliées aux analytics/contracts ;
* les projections système restent inspectables ;
* les primitives existantes continuent de fonctionner ;
* aucune dépendance à un framework agile externe n’est introduite.

---

### 23.12 Résumé

Cette évolution ajoute une couche business-aware au Product Layer.

Le moteur Asagiri ne doit plus uniquement raisonner en :

```text
feature → task → code
```

Mais en :

```text
business intent
  ↓
product flow
  ↓
system implications
  ↓
implementation
  ↓
measurable outcome
```

Le flow devient l’unité centrale reliant :

* produit ;
* architecture ;
* observabilité ;
* analytics ;
* sécurité ;
* implémentation ;
* validation.

---

## 24. Persistent Runtime & Session Graph Layer

### 24.1 Objectif

Transformer Asagiri d’une CLI principalement stateless vers un runtime d’ingénierie persistant, vivant et orchestrable.

Aujourd’hui :

```text
command
  ↓
execution
  ↓
report
```

Évolution cible :

```text
persistent runtime
  ↓
sessions
  ↓
branches
  ↓
flows
  ↓
agents
  ↓
events
  ↓
memory
```

Le runtime doit devenir la couche centrale pilotant :

* sessions ;
* branches ;
* flows ;
* orchestration ;
* mémoire ;
* events ;
* état d’exécution ;
* projections architecture ;
* coordination agents.

Le but n’est pas de transformer Asagiri en chatbot terminal.

Le but est de créer un environnement d’exploitation d’ingénierie produit piloté par flows.

---

### 24.2 Concepts principaux

#### Runtime

Processus persistant local maintenant :

* état actif ;
* sessions ;
* index ;
* mémoire ;
* orchestration ;
* event bus ;
* workers ;
* projections ;
* caches ;
* métriques runtime.

#### Session

Contexte d’ingénierie actif.

Une session peut représenter :

* une feature ;
* un flow ;
* une expérimentation ;
* une review ;
* une branche architecture ;
* un prototype.

#### Branch

Variation exploratoire d’un flow, design, architecture ou implémentation.

#### Runtime Event

Événement produit par le moteur.

Exemples :

```text
flow.started
flow.completed
task.failed
review.rejected
prototype.updated
architecture.risk_detected
cost.threshold_reached
```

#### Runtime State Graph

Graphe représentant :

* sessions ;
* branches ;
* flows ;
* tasks ;
* décisions ;
* events ;
* relations.

---

### 24.3 Nouvelles commandes

#### `asa daemon start`

Démarrer le runtime persistant.

```bash
asa daemon start
```

Responsabilités :

* initialiser runtime local ;
* ouvrir DB runtime ;
* charger mémoire ;
* démarrer event bus ;
* démarrer workers ;
* préparer cache ;
* exposer socket local/API.

---

#### `asa daemon status`

Afficher état runtime.

```bash
asa daemon status
```

Sortie attendue :

```text
Asagiri Runtime
────────────────
Status: running
Sessions: 4
Flows active: 2
Workers: 3
Queued events: 12
Memory size: 182 entries
DB size: 42 MB
```

---

#### `asa session create`

Créer une session runtime.

```bash
asa session create onboarding-redesign
```

---

#### `asa session attach`

Attacher terminal à une session.

```bash
asa session attach onboarding-redesign
```

---

#### `asa session branch`

Créer une branche exploratoire.

```bash
asa session branch onboarding-redesign --name onboarding-enterprise
```

---

#### `asa session graph`

Afficher graphe sessions/branches.

```bash
asa session graph
```

---

#### `asa runtime events`

Afficher événements runtime.

```bash
asa runtime events --follow
```

---

### 24.4 Nouvelle arborescence runtime

```text
.asagiri/runtime/
  runtime.db
  sessions/
  branches/
  memory/
  events/
  projections/
  metrics/
  cache/
```

---

### 24.5 Runtime database

Le runtime doit utiliser SQLite.

Tables minimales :

```text
sessions
branches
runtime_events
memory_entries
flow_states
runtime_metrics
workers
runtime_locks
```

---

### 24.6 Session model

```go
type Session struct {
    ID             string
    Name           string
    ProductID      string
    FlowID         string
    BranchID       string
    Status         SessionStatus
    ActiveTasks    []string
    RuntimeContext RuntimeContext
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

---

### 24.7 Branch model

```go
type Branch struct {
    ID             string
    ParentBranchID string
    SessionID      string
    Name           string
    Type           BranchType
    Description    string
    Divergence     DivergenceInfo
    CreatedAt      time.Time
}
```

Types possibles :

```text
product
flow
architecture
implementation
review
prototype
```

---

### 24.8 Runtime event bus

Créer un event bus interne.

Objectifs :

* découpler runtime ;
* déclencher automatisations ;
* produire observabilité ;
* permettre hooks ;
* permettre workflows réactifs.

Interface :

```go
type RuntimeEvent struct {
    ID          string
    Type        string
    Source      string
    SessionID   string
    FlowID      string
    Payload     map[string]any
    CreatedAt   time.Time
}
```

---

### 24.9 Runtime hooks

Ajouter hooks événementiels.

Exemple :

```yaml
hooks:
  on:
    flow.completed:
      - run: asa review auto
      - run: asa metrics update

    review.rejected:
      - run: asa task reopen
```

---

### 24.10 Knowledge & Memory Engine

Créer une mémoire persistante structurée.

Objectif :

* retenir patterns ;
* retenir décisions ;
* retenir erreurs ;
* retenir optimisations ;
* injecter contexte utile automatiquement.

Le système ne doit pas être un simple stockage markdown.

Créer :

```text
internal/memory/
  extraction.go
  consolidation.go
  retrieval.go
  scoring.go
  aging.go
  embedding.go
  graph.go
```

---

### 24.11 Mémoire scoped

Chaque mémoire doit avoir un scope.

Scopes :

```text
global
project
product
flow
feature
task
branch
agent
```

---

### 24.12 Memory entry model

```go
type MemoryEntry struct {
    ID          string
    Scope       MemoryScope
    Type        MemoryType
    Summary     string
    Source      string
    Relevance   float64
    Tags        []string
    LinkedFlows []string
    CreatedAt   time.Time
    LastUsedAt  time.Time
}
```

---

### 24.13 Memory extraction

Le runtime doit pouvoir extraire automatiquement :

* décisions architecture ;
* conventions projet ;
* erreurs récurrentes ;
* flows fréquents ;
* causes d’échec ;
* patterns de review ;
* optimisations performantes ;
* incidents.

---

### 24.14 Skills & Capabilities System

Créer un système de compétences réutilisables.

Objectif :

* partager des workflows ;
* spécialiser les agents ;
* mutualiser validations ;
* injecter expertise contextualisée.

Structure :

```text
.asagiri/skills/
  architecture/
  backend/
  frontend/
  infra/
  observability/
  review/
```

---

### 24.15 Skill model

```yaml
id: api-platform-review
name: API Platform Review
scope:
  - backend
  - architecture

capabilities:
  - contract_review
  - resource_validation
  - serialization_review

rules:
  - avoid_anemic_resources
  - enforce_validation_groups

checks:
  - phpstan
  - phpunit

metrics:
  - review_success_rate
```

---

### 24.16 Engineering Analysis Layer

Créer une couche d’intelligence structurelle locale.

Objectif :

Construire des graphes exploitables par les agents.

Créer :

```text
internal/analysis/
  ast/
  symbols/
  dependencies/
  events/
  flows/
  ownership/
  architecture/
```

Graphes cibles :

```text
symbol graph
dependency graph
flow graph
event graph
api graph
migration graph
ownership graph
```

---

### 24.17 Runtime modes

Le runtime doit supporter plusieurs modes.

```yaml
runtime:
  mode: guided
```

Modes :

```text
guided
interactive
headless
ci
review
exploration
```

---

### 24.18 Runtime API / SDK

Asagiri ne doit plus être uniquement une CLI.

Ajouter :

* runtime API locale ;
* SDK Go ;
* SDK TypeScript ;
* runtime embedding ;
* orchestration API.

Exemple :

```go
runtime := asagiri.Connect()

runtime.StartSession("workspace-redesign")
runtime.RunFlow("onboarding")
```

---

### 24.19 Runtime observability

Le runtime lui-même doit être observable.

Metrics :

```text
active_sessions
memory_hits
context_reduction_ratio
runtime_event_rate
agent_failure_rate
flow_completion_rate
review_rejection_rate
```

---

### 24.20 UX terminal cible

```text
Asagiri Runtime
════════════════
Session: onboarding-redesign
Branch:  onboarding-enterprise
Flow:    onboarding

Runtime
───────
Workers active:        3
Queued events:         12
Memory hits:           84%
Context reduction:     71%

Flows
─────
✓ onboarding.start
⠋ invite_team
○ billing_setup

Recent events
─────────────
flow.started
prototype.updated
architecture.risk_detected
review.pending
```

---

### 24.21 Critères d’acceptation

Cette évolution est acceptable si :

* `asa daemon start` démarre un runtime persistant ;
* les sessions peuvent être créées/reprises ;
* les branches peuvent être créées ;
* les runtime events sont persistés ;
* les hooks fonctionnent ;
* la mémoire scoped fonctionne ;
* les skills peuvent être chargés ;
* les graphes locaux peuvent être construits ;
* le runtime reste compatible CLI plain ;
* les primitives existantes fonctionnent toujours ;
* le runtime peut fonctionner sans cloud ;
* aucune dépendance UI n’est imposée au core.

---

### 24.22 Risques

#### Runtime trop complexe

Mitigation :

* architecture modulaire ;
* runtime optional ;
* fallback stateless ;
* isolation runtime/core.

#### Mémoire polluée

Mitigation :

* scoring ;
* aging ;
* scopes ;
* consolidation ;
* review mémoire.

#### Event storm

Mitigation :

* throttling ;
* batching ;
* retention ;
* priority queues.

#### Branch chaos

Mitigation :

* graph visualization ;
* branch lineage ;
* merge policies ;
* archival.

---

### 24.23 Résumé

Cette évolution transforme Asagiri en runtime d’ingénierie persistant.

Avant :

```text
CLI → execution → report
```

Après :

```text
persistent runtime
  ↓
sessions
  ↓
flows
  ↓
memory
  ↓
analysis graphs
  ↓
agents
  ↓
events
  ↓
reviews
```

Le runtime devient le système vivant pilotant :

* flows ;
* architecture ;
* mémoire ;
* orchestration ;
* analyse ;
* agents ;
* validations.

---

## 25. Investigation & Root Cause Engine

### 25.1 Objectif

Ajouter à Asagiri une couche d’investigation logicielle structurée, locale-first et agent-aware.

Le but est de permettre à Asagiri de comprendre un problème avant de demander à un agent de modifier le code.

Aujourd’hui, beaucoup d’agents suivent implicitement ce pattern fragile :

```text
prompt
  ↓
lecture partielle du repo
  ↓
patch approximatif
  ↓
validation tardive
```

Évolution cible :

```text
symptom / intent
  ↓
investigation locale
  ↓
hypothèses scorées
  ↓
root cause candidates
  ↓
context pack minimal
  ↓
agent spécialisé
  ↓
validation
  ↓
rapport capitalisable
```

L’investigation devient une étape de première classe dans Asagiri.

Elle doit être :

* explicite ;
* traçable ;
* mesurable ;
* rejouable ;
* locale-first ;
* connectée aux flows, tasks, contracts, tests et événements.

---

### 25.2 Positionnement

Cette couche ne remplace pas :

* `asa work` ;
* `asa verify` ;
* `asa review` ;
* les agents d’implémentation ;
* les tests.

Elle s’insère avant l’exécution agentique.

Pattern cible :

```text
asa work
  ↓
resolve intent
  ↓
if uncertainty or failure: investigate
  ↓
produce investigation report
  ↓
build reduced context
  ↓
execute implementation/review
```

---

### 25.3 Cas d’usage

#### Debugging

```bash
asa investigate "l’onboarding échoue après invitation d’un membre"
```

#### Échec de tests

```bash
asa investigate --from-failed-tests
```

#### Échec de flow

```bash
asa investigate --flow onboarding
```

#### Échec de run

```bash
asa investigate --run run-2026-05-27-001
```

#### Analyse avant implémentation

```bash
asa investigate billing-v2 --task task-003
```

#### Analyse d’impact

```bash
asa investigate impact --flow onboarding --change "make invitations async"
```

---

### 25.4 Nouvelle commande principale : `asa investigate`

Usage :

```bash
asa investigate "why onboarding fails after invite"
```

Options :

```bash
asa investigate "..." \
  --flow onboarding \
  --task task-003 \
  --run run-2026-05-27-001 \
  --from-failed-tests \
  --depth standard \
  --max-files 80 \
  --max-duration 5m \
  --no-cloud \
  --estimate-only \
  --output markdown
```

Modes :

```text
quick
standard
deep
ci
```

---

### 25.5 Pipeline d’investigation

Pipeline cible :

```text
Input symptom / intent
  ↓
Scope resolution
  ↓
Local discovery
  ↓
Static analysis
  ↓
Runtime evidence collection
  ↓
Flow / contract correlation
  ↓
Hypothesis generation
  ↓
Hypothesis scoring
  ↓
Root cause candidates
  ↓
Context pack generation
  ↓
Investigation report
```

---

### 25.6 Scope resolution

Le moteur doit d’abord réduire le périmètre.

Sources possibles :

* flow ;
* task ;
* feature ;
* run ;
* failed tests ;
* changed files ;
* error logs ;
* user instruction ;
* contracts ;
* architecture projection.

Exemple :

```text
Instruction: onboarding fails after invite
Resolved scope:
- flow: onboarding
- step: invite_team
- action: invite_member
- likely modules: src/Invitation, src/Onboarding, tests/Onboarding
- contracts: member.invite, invitation_email_sent
```

---

### 25.7 Local discovery

Asagiri doit effectuer localement les recherches simples avant tout appel modèle.

Capacités minimales :

* `ripgrep` / grep ;
* recherche fichiers ;
* recherche symboles ;
* mapping tests associés ;
* lecture ciblée ;
* extraction imports ;
* détection routes/endpoints ;
* détection handlers/events ;
* détection migrations ;
* détection config ;
* détection fichiers sensibles.

---

### 25.8 Static analysis

Créer un module :

```text
internal/investigation/static/
  symbols.go
  dependencies.go
  routes.go
  tests.go
  events.go
  config.go
  migrations.go
```

Objectifs :

* comprendre les relations code ;
* identifier les dépendances ;
* détecter blast radius ;
* identifier tests pertinents ;
* construire un graphe minimal de cause possible.

---

### 25.9 Runtime evidence collection

Quand disponible, Asagiri doit exploiter :

* logs de run ;
* stdout/stderr agents ;
* résultats tests ;
* traces locales ;
* rapports précédents ;
* diff courant ;
* historique métriques ;
* états SQLite.

Le moteur ne doit pas inventer d’évidence.

Toute conclusion doit référencer une source.

---

### 25.10 Flow / contract correlation

L’investigation doit relier le problème aux artefacts produit.

Exemple :

```text
Failure: invite_member returns 500
Flow: onboarding
Step: invite_team
Action: invite_member
Contract: POST /invitations
Event expected: member.invited
Metric expected: invitation_delivery_success_rate
```

Cela permet de distinguer :

* bug UI ;
* bug backend ;
* contrat incomplet ;
* état erreur oublié ;
* observabilité insuffisante ;
* mauvaise hypothèse produit.

---

### 25.11 Hypothesis Engine

Créer un moteur d’hypothèses.

Interface :

```go
type Hypothesis struct {
    ID          string
    Title       string
    Description string
    Evidence    []Evidence
    Confidence  float64
    Risk        RiskLevel
    SuggestedChecks []Check
    LinkedFiles []string
    LinkedFlows []string
}
```

Exemple :

```yaml
id: hyp-001
title: invitation email failure is not handled
confidence: 0.78
evidence:
  - failed test: InvitationServiceTest::testInviteMember
  - missing error state in onboarding.flow.yaml
  - no retry policy in architecture projection
suggested_checks:
  - run tests/Invitation
  - inspect InvitationService
  - verify email provider config
```

---

### 25.12 Hypothesis scoring

Le scoring doit prendre en compte :

* proximité avec l’erreur ;
* nombre de sources convergentes ;
* changements récents ;
* historique d’échecs ;
* criticité flow ;
* couverture tests ;
* blast radius ;
* confiance du parsing.

Score exemple :

```text
confidence = evidence_strength * locality * recency * test_correlation * flow_relevance
```

Le score doit rester indicatif, pas présenté comme vérité.

---

### 25.13 Root Cause Graph

Créer un graphe de cause probable.

Nœuds possibles :

```text
symptom
flow
action
contract
file
symbol
test
log
event
metric
hypothesis
```

Relations possibles :

```text
fails_at
implemented_by
tested_by
emits
requires
depends_on
changed_in
supports
contradicts
```

---

### 25.14 Context Pack Generation

L’investigation doit produire un contexte minimal pour agent.

Sortie :

```text
.asagiri/investigations/<id>/context-pack.md
```

Le context pack doit contenir :

* problème résumé ;
* scope ;
* flows liés ;
* contrats liés ;
* fichiers clés ;
* extraits pertinents ;
* hypothèses scorées ;
* commandes à lancer ;
* critères de succès ;
* restrictions.

Il ne doit pas contenir :

* secrets ;
* logs complets inutiles ;
* fichiers hors scope ;
* contexte redondant.

---

### 25.15 Investigation Report

Créer :

```text
.asagiri/investigations/<id>/report.md
.asagiri/investigations/<id>/report.json
```

Structure Markdown :

```md
# Investigation Report

## Input

## Resolved Scope

## Evidence Collected

## Related Flows

## Related Contracts

## Hypotheses

## Root Cause Candidates

## Suggested Next Actions

## Context Pack

## Risks

## Limits
```

---

### 25.16 Intégration avec `asa work`

Ajouter options :

```bash
asa work "fix onboarding invite" --investigate-first
asa work "fix onboarding invite" --investigation-depth deep
asa work "fix onboarding invite" --use-investigation <id>
```

Comportement :

1. résoudre intention ;
2. lancer investigation ;
3. afficher hypothèses ;
4. demander confirmation si risque élevé ;
5. générer context pack ;
6. appeler agent avec contexte réduit ;
7. valider ;
8. enrichir mémoire.

---

### 25.17 Intégration avec échecs

Quand une étape échoue, Asagiri doit proposer :

```text
Verification failed.
Run investigation? [Y/n]
```

Ou en mode CI :

```bash
asa investigate --run <run-id> --from-failed-tests --output json
```

---

### 25.18 Mémoire d’investigation

Les investigations doivent alimenter la mémoire runtime.

À mémoriser :

* cause validée ;
* hypothèses fausses ;
* fichiers liés ;
* commandes utiles ;
* patterns d’échec ;
* fixes efficaces ;
* temps/coût ;
* agent utilisé.

Scopes :

```text
project
flow
feature
task
agent
```

---

### 25.19 Reproductibility Pack

Chaque investigation doit pouvoir être rejouée.

Créer :

```text
.asagiri/investigations/<id>/replay.yaml
```

Contenu :

```yaml
input: "onboarding fails after invite"
repo_commit: abc123
scope:
  flow: onboarding
  task: task-003
commands:
  - rg "invite_member" src tests
  - composer test -- tests/Onboarding
artifacts:
  - report.md
  - context-pack.md
  - root-cause-graph.json
```

---

### 25.20 Architecture Go

Créer :

```text
internal/investigation/
  engine.go
  scope.go
  evidence.go
  hypothesis.go
  scoring.go
  graph.go
  context_pack.go
  report.go
  replay.go
  static/
  runtime/
  collectors/
```

Interfaces :

```go
type InvestigationEngine interface {
    Investigate(ctx context.Context, req InvestigationRequest) (InvestigationResult, error)
}

type EvidenceCollector interface {
    Collect(ctx context.Context, scope InvestigationScope) ([]Evidence, error)
}

type HypothesisGenerator interface {
    Generate(ctx context.Context, evidence []Evidence) ([]Hypothesis, error)
}

type ContextPackBuilder interface {
    Build(ctx context.Context, result InvestigationResult) (ContextPack, error)
}
```

---

### 25.21 UX terminal cible

```text
Asagiri Investigation
─────────────────────
Input: onboarding fails after invite

Scope
─────
Flow:   onboarding
Step:   invite_team
Action: invite_member
Files:  14 candidates
Tests:  3 related

Evidence
────────
✓ Found route POST /invitations
✓ Found InvitationService
✓ Found related tests
✓ Found failed test output
⚠ No retry policy found
⚠ No error state defined for invite failure

Hypotheses
──────────
1. Invitation failure is not handled correctly       confidence: 0.78
2. Email provider config missing in test env        confidence: 0.63
3. Permission check rejects invited member flow      confidence: 0.51

Suggested next action
─────────────────────
Run focused fix with context pack:
asa work "fix invite failure handling" --use-investigation inv-2026-05-27-001
```

---

### 25.22 Sécurité

L’investigation doit respecter :

* scope fichiers ;
* chemins interdits ;
* secret scanning ;
* taille max de sortie ;
* timeout ;
* politique réseau ;
* logs sensibles masqués.

Par défaut :

* aucune donnée sensible dans context pack ;
* pas de lecture `.env` ;
* pas de dump logs complet ;
* pas d’appel cloud si `--no-cloud`.

---

### 25.23 Cost-aware investigation

Ajouter estimation :

```bash
asa investigate "..." --estimate-only
```

Sortie :

```text
Investigation Estimate
──────────────────────
Local steps:       8
Model steps:       1
Files to scan:     ~120
Estimated tokens:  18k
Estimated cost:    €0.02
Estimated time:    1m40s
```

Par défaut, l’investigation doit être principalement locale.

---

### 25.24 Critères d’acceptation

Cette évolution est acceptable si :

* `asa investigate "..."` produit un rapport ;
* le scope est résolu automatiquement quand possible ;
* les évidences sont référencées ;
* les hypothèses sont scorées ;
* un context pack minimal est généré ;
* un replay pack est généré ;
* `asa work --investigate-first` fonctionne ;
* les échecs de tests peuvent déclencher investigation ;
* les investigations alimentent la mémoire ;
* le mode `--no-cloud` fonctionne ;
* les fichiers sensibles sont exclus ;
* les tests unitaires/golden/integration critiques passent.

---

### 25.25 Découpage d’implémentation recommandé

#### Phase 1 — Investigation model

* types Go ;
* investigation request/result ;
* evidence model ;
* report writer ;
* tests unitaires.

#### Phase 2 — Local collectors

* grep collector ;
* file collector ;
* test collector ;
* route collector ;
* config collector.

#### Phase 3 — Scope resolver

* flow/task/run scope ;
* failed tests scope ;
* changed files scope.

#### Phase 4 — Hypothesis engine V1

* règles déterministes ;
* scoring simple ;
* hypotheses report.

#### Phase 5 — Context pack

* contexte réduit ;
* exclusions sensibles ;
* pack markdown/json.

#### Phase 6 — Integration work/verify

* `--investigate-first` ;
* investigation sur échec ;
* mémoire.

#### Phase 7 — Root cause graph

* graph JSON ;
* relations ;
* visualisation CLI minimale.

---

### 25.26 Risques

#### Hypothèses fausses

Mitigation :

* présenter comme candidates ;
* afficher evidence ;
* scorer ;
* demander validation.

#### Trop de contexte

Mitigation :

* context pack borné ;
* max files ;
* reducers ;
* déduplication.

#### Investigation lente

Mitigation :

* modes quick/standard/deep ;
* timeouts ;
* cache ;
* collectors parallélisables.

#### Fuite de secrets

Mitigation :

* secret scanner ;
* forbidden paths ;
* redaction ;
* audit logs.

---

### 25.27 Résumé

Cette évolution rend Asagiri capable d’investiguer avant d’agir.

Avant :

```text
agent receives broad context
  ↓
tries to fix
```

Après :

```text
symptom
  ↓
local investigation
  ↓
evidence
  ↓
hypotheses
  ↓
minimal context pack
  ↓
targeted agent execution
  ↓
validated fix
```

Principe clé :

> Les agents ne doivent pas explorer le repo à l’aveugle. Asagiri doit enquêter localement, structurer les causes probables, puis fournir aux agents un contexte réduit et vérifiable.

---

## 26. Résumé exécutable

Cette évolution ajoute à Asagiri une couche produit exécutable :

```text
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
```

Principe clé :

> Le prototype sert à clarifier le produit. Les flows servent à structurer l’ingénierie. Les contracts servent à cadrer le système. Les specs servent à piloter les agents.

Asagiri ne doit pas promettre de la magie IA. Il doit fournir une chaîne de transformation contrôlée, inspectable et industrialisable.
