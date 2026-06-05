# Implementation Plan: audit-coherence-consolidation

> Feature : audit-coherence-consolidation
> Nature : **correction & simplification**, pas un nouveau sous-système.

## Overview

Ce plan transforme le design en une série d'étapes de code incrémentales et
ordonnées, suivant la structure **par constat** du design (`AUD-001` … `AUD-007`,
exigences `R1` … `R8`). On applique le **plus petit changement correct** à des
fichiers **existants** du dépôt, sans nouvelle couche ni moteur d'audit runtime.

Ordre d'exécution : (1) régénérer la doc CLI et poser le garde-fou
régénération-sans-diff ; (2) rendre le routing config-driven et explicable ;
(3) relier la policy Ollama au canon courant ; (4) réparer le gate lint et
ajouter la CI Go ; (5) mettre en avant le chemin guidé sans rien retirer ;
(6) borner et vérifier les garde-fous d'onboarding ; (7) faire de `problems.md`
le Remediation_Register et vérifier l'alignement des locales ; (8) prouver la
tranche par le Quality_Gate + régénération-sans-diff + clôture du registre.

Chaque garde-fou (test) atterrit **avec** sa correction afin de capter les
régressions tôt. La tâche finale **prouve** la tranche complète.

### Conventions de test

- Langage : **Go** (le design est entièrement en Go ; aucun pseudocode).
- Tests property-based : **une seule propriété par test**, **≥ 100 itérations**,
  générateurs déterministes. Chaque test porte le tag de commentaire
  **`Feature: audit-coherence-consolidation, Property {N}`** et référence la
  propriété `P1` … `P21` du design.
- Bibliothèque PBT : **`testing/quick`** de la librairie standard ; repli
  **`pgregory.net/rapid`** si des générateurs plus riches sont nécessaires.
  Aucune implémentation de PBT « maison ».
- `gofmt` / `go vet` propres ; **aucun `panic`** aux frontières CLI (erreurs
  retournées comme valeurs, `03-standards.md`).
- Les sous-tâches marquées `*` (tests) sont optionnelles pour un MVP et **ne sont
  pas implémentées** par défaut.

## Tasks

- [x] 1. Garde-fou docgen + régénération de la doc CLI (R1 — AUD-001/AUD-002)
  - [x] 1.1 Régénérer la référence CLI générée
    - Exécuter `go run ./application/cmd/asa docs generate-cli --output docs-site/content/docs/en/cli/generated`
    - Résultat attendu : ajout de `docs-site/content/docs/en/cli/generated/runs.mdx`
      (AUD-001) et rafraîchissement du lien fratrie `> - [Runs](./runs.mdx)` sur
      chaque page sœur (AUD-002) ; **aucune** modification de `docgen.go`
    - _Requirements: 1.1, 1.3, 1.6, 8.1_

  - [x] 1.2 Test de propriété : bijection commande ↔ page MDX
    - Nouveau fichier `application/internal/cli/docgen/regen_bijection_test.go` :
      construire l'arbre via `cli.RootCommand()`, régénérer en `t.TempDir()`,
      asserter une page `Slug(p).mdx` par commande de `CommandPathsWithoutRoot`
      (dont `runs`) et aucune page `.mdx` orpheline
    - **Property 1: Bijection commande ↔ page MDX**
    - **Validates: Requirements 1.1**

  - [x] 1.3 Test de propriété : déterminisme octet-à-octet de Docgen
    - Nouveau fichier `application/internal/cli/docgen/regen_determinism_test.go` :
      deux exécutions de `docgen.Generate` vers deux `t.TempDir()` distincts, diff
      binaire des `*.mdx` (mêmes chemins, contenus identiques octet pour octet)
    - **Property 2: Déterminisme octet-à-octet de Docgen**
    - **Validates: Requirements 1.2**

  - [x] 1.4 Test de propriété : Regeneration_Check (hors `meta.json`)
    - Nouveau fichier `application/internal/cli/docgen/regen_check_test.go` :
      injecter des divergences synthétiques (page manquante, octet modifié,
      `meta.json` présent) ; asserter exit ≠ 0 **ssi** fichier manquant/orphelin/
      divergent en listant les fichiers, et que `meta.json` ne déclenche jamais de
      divergence (`present = { e | ext(e)==".mdx" }`)
    - **Property 3: Regeneration_Check — exit ≠ 0 ssi différence, hors `meta.json`**
    - **Validates: Requirements 1.4, 1.5, 1.6**

  - [x] 1.5 Test exemple : Regeneration_Without_Diff + lien fratrie Runs
    - Nouveau fichier `application/internal/cli/docgen/regen_nodiff_test.go` :
      régénérer en tmp et differ contre `docs-site/content/docs/en/cli/generated/`
      (hors `meta.json`) → exit 0 attendu sur le dépôt corrigé ; vérifier qu'une
      page voisine de `runs` contient `> - [Runs](./runs.mdx)`
    - _Requirements: 1.3, 1.6, 8.1_

- [x] 2. Routing config-driven et explicable (R4 — AUD-005)
  - [x] 2.1 Refactor de `routing.Route` (config-driven, précédence `no_cloud`)
    - Modifier `application/internal/routing/router.go` : changer la signature en
      `Route(cfg, stepClass, preferLocal, noCloud, allowCloud) (Decision, error)` ;
      exporter `var ErrNoDeclaredBackend` ; appliquer la priorité **`no_cloud`
      avant `prefer_local`** ; sélectionner uniquement un backend **déclaré** dans
      `cfg.Agents` (helpers `firstLocalDeclared` parcourant les clés triées,
      `cloudHeavyDeclared`) ; supprimer les littéraux `"cursor"`, `"claude"`,
      `"ollama"` du chemin de décision ; raisons dans
      `{prefer_local, no_cloud, cloud_heavy, cloud_fast, default}` ; si aucun
      backend déclaré ne correspond → retourner `Decision{}` zéro + erreur enveloppant
      `ErrNoDeclaredBackend`, sans `panic` ; fonction pure (aucun effet de bord)
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7_

  - [x] 2.2 Adapter l'appelant `cost/estimator.go` à la nouvelle signature
    - Modifier `application/internal/cost/estimator.go` : gérer la valeur `error`
      de `routing.Route` (propager/atténuer comme valeur, aucun `panic`)
    - _Requirements: 4.2, 4.7_

  - [x] 2.3 Étendre `asa explain` avec un mode non interactif `routing`
    - Modifier `application/internal/cli/root_ui.go` (`newExplainCmd`) : ajouter le
      sous-mode `asa explain routing --step-class <cls> [--prefer-local --no-cloud --json]`
      appelant `routing.Route` et rendant un DTO de présentation `RoutingExplanation`
      `{StepClass, Agent, Model, Local, Reason}` en parité plain/json, nommant le
      backend et la raison ; le rendu TUI existant reste inchangé (ADR-027)
    - _Requirements: 4.8_

  - [x] 2.4 Test de propriété : déterminisme et pureté du routing
    - Nouveau fichier `application/internal/routing/router_determinism_test.go` :
      configs / stepClass / flags générés ; double appel → même issue
      (succès/erreur) et `Decision` identique sur tous les champs, sans effet de bord
    - **Property 4: Déterminisme et pureté du routing**
    - **Validates: Requirements 4.1**

  - [x] 2.5 Test de propriété : backend déclaré et raison valide
    - Nouveau fichier `application/internal/routing/router_backend_test.go` :
      sur succès, asserter `Decision.Agent ∈ keys(cfg.Agents)` et
      `Decision.Reason ∈ {prefer_local, no_cloud, cloud_heavy, cloud_fast, default}`
    - **Property 5: Backend déclaré et raison valide**
    - **Validates: Requirements 4.2, 4.6**

  - [x] 2.6 Test de propriété : règle `prefer_local`
    - Nouveau fichier `application/internal/routing/router_preferlocal_test.go` :
      `preferLocal` (ou classe ∈ `prefer_local_for`) et `noCloud=false` → succès
      avec `Local=true` et raison `prefer_local`
    - **Property 6: Règle prefer_local**
    - **Validates: Requirements 4.3**

  - [x] 2.7 Test de propriété : précédence `no_cloud`
    - Nouveau fichier `application/internal/routing/router_nocloud_test.go` :
      `noCloud=true` avec `allowCloud`/`preferLocal` variés → `Local=true`,
      raison `no_cloud`
    - **Property 7: Précédence no_cloud**
    - **Validates: Requirements 4.4**

  - [x] 2.8 Test de propriété : stratégie par défaut
    - Nouveau fichier `application/internal/routing/router_default_test.go` :
      ni `noCloud` ni `preferLocal` → raison `cloud_heavy`/`cloud_fast`/`default`
      selon l'appartenance de la classe aux listes de stratégie
    - **Property 8: Stratégie par défaut**
    - **Validates: Requirements 4.5**

  - [x] 2.9 Test de propriété : erreur guidée sans panic
    - Nouveau fichier `application/internal/routing/router_guided_error_test.go` :
      configs sans backend adéquat (`cfg.Agents` vide ou local manquant) →
      `ErrNoDeclaredBackend`, `Decision` zéro, aucun backend non déclaré ;
      `recover` pour prouver l'absence de `panic`
    - **Property 9: Erreur guidée sans panic quand aucun backend ne correspond**
    - **Validates: Requirements 4.7**

  - [x] 2.10 Test exemple : `asa explain routing` nomme backend + raison
    - Nouveau fichier `application/internal/cli/root_ui_explain_test.go` : sur
      quelques `--step-class`, la sortie plain **et** json contient le nom du
      backend et la raison, cohérents avec la `Decision` de `Route`
    - _Requirements: 4.8_

- [x] 3. Cohérence policy ↔ canon (R5 — AUD-006)
  - [x] 3.1 Relier les rôles Ollama au canon courant (source canonique unique)
    - Modifier `application/internal/policy/ollama.go` : remplacer les renvois
      historiques `§10.1/§10.2` par une **source canonique unique**
      (`ollamaRoleCanon` référençant le canon courant) ; préserver à l'identique
      le comportement de `CheckOllamaRole` (refus nommant le rôle, sans `panic`) et
      `IsOllamaAgent` ; garantir `Allowed ∩ Forbidden = ∅`
    - _Requirements: 5.1, 5.4_

  - [x] 3.2 Test de propriété : cohérence policy ↔ canon (Policy_Coherence_Check)
    - Nouveau fichier `application/internal/policy/coherence_test.go` : pour chaque
      rôle des listes Allowed/Forbidden, le check réussit **ssi** le rôle existe
      dans le canon courant ; sinon il échoue en **nommant** le rôle divergent
      (injecter un rôle divergent)
    - **Property 10: Cohérence policy ↔ canon**
    - **Validates: Requirements 5.2, 5.3**

  - [x] 3.3 Test de propriété : refus des rôles interdits sans panic
    - Nouveau fichier `application/internal/policy/ollama_role_test.go` :
      `CheckOllamaRole(role)` retourne une erreur nommant le rôle **ssi**
      `role ∈ OllamaForbiddenRoles`, `nil` sinon, et ne `panic` jamais
    - **Property 11: Refus des rôles interdits sans panic**
    - **Validates: Requirements 5.4**

- [x] 4. Gate de lint réparé + Go CI (R2/R8 — AUD-003)
  - [x] 4.1 Corriger `.golangci.yml` en schéma v2 valide
    - Modifier `.golangci.yml` : ajouter `version: "2"`, `run.go: "1.25"`
      (aligné sur `go.mod`), `run.timeout: 5m`, `run.tests: true` ; activer
      `errcheck, govet, ineffassign, staticcheck, unused` ; **retirer** les clés v1
      interdites (ex. `issues.exclude-use-default`)
    - _Requirements: 2.1, 2.2_

  - [x] 4.2 Ajouter le workflow Go CI (Quality_Gate + Regeneration_Check)
    - Créer `.github/workflows/go-ci.yml` déclenché sur `push`/`pull_request` :
      `actions/setup-go` avec `go-version-file: go.mod` ; étapes `make build`,
      `go vet ./...`, `go test ./...` ; installer golangci-lint **pinné v2**
      (binaire officiel bâti go ≥ 1.25) puis `golangci-lint run` ; étape
      Regeneration_Check (`asa docs generate-cli --output /tmp/cli-regen` puis
      `diff -ruq /tmp/cli-regen docs-site/content/docs/en/cli/generated --exclude=meta.json`) ;
      scan secrets en clair (exit ≠ 0 si détection, sans ré-émettre de valeur)
    - _Requirements: 2.3, 2.4, 8.1, 8.2, 8.4_

  - [x] 4.3 Documenter la commande d'installation golangci-lint pinné
    - Mettre à jour `docs/ai/03-standards.md` : commande reproductible
      (`curl … install.sh | sh -s -- -b "$(go env GOPATH)/bin" v2.12.2`), note
      « éviter `go install` » et définition du Quality_Gate (`build` ∧ `test` ∧
      `vet` ∧ `lint`, tous exit 0)
    - _Requirements: 2.2, 2.3_

- [x] 5. Restauration de la simplicité du chemin guidé (R6 — AUD-007)
  - [x] 5.1 Bloc « Pour commencer » curé dans le help racine
    - Modifier `application/internal/cli/help.go` (`rootLong`) : insérer un
      Guided_Path unique et découvrable (`onboarding → work → jalons de
      validation`) et rappeler que les Unitary_Command restent disponibles ;
      **aucune** commande retirée de l'arbre Cobra
    - _Requirements: 6.1, 6.2_

  - [x] 5.2 Page docs d'entrée du Guided_Path (4 locales)
    - Créer `docs-site/content/docs/{en,fr,de,es}/<guided-path>.mdx` décrivant
      l'enchaînement `onboarding → work → jalons de validation` ; hors
      `cli/generated/`, donc présente dans les 4 locales (alignement R8.3)
    - _Requirements: 6.1_

  - [x] 5.3 Test exemple : Guided_Path présent + Unitary_Command préservées
    - Nouveau fichier `application/internal/cli/help_guided_test.go` : `rootLong`
      contient le bloc « Pour commencer » ; l'ensemble des Unitary_Command de
      référence (`spec, plan, enrich, dev, verify, review`) reste présent dans
      `RootCommand()` (garde anti-régression)
    - _Requirements: 6.1, 6.2_

  - [x] 5.4 Test de propriété : remédiation guidée sans panic
    - Nouveau fichier `application/internal/cli/guided_remediation_test.go` :
      prérequis manquants variés (outil/config/dépendance) → arrêt sans `panic`,
      exit ≠ 0, message nommant l'élément manquant **et** au moins une action de
      résolution
    - **Property 12: Remédiation guidée sans panic**
    - **Validates: Requirements 6.3, 7.7**

  - [x] 5.5 Test de propriété : dépôt inchangé sur abandon
    - Nouveau fichier `application/internal/cli/abort_unchanged_test.go` :
      snapshot du dépôt avant/après un arrêt sur prérequis manquant → identique
      (aucun artefact partiel)
    - **Property 13: Dépôt inchangé sur abandon**
    - **Validates: Requirements 6.4**

  - [x] 5.6 Test de propriété : dry-run sans effet persistant
    - Nouveau fichier `application/internal/cli/dryrun_test.go` : commandes
      supportées avec `--dry-run` (mocks d'agents/commandes) → aucune invocation
      d'Agent_Backend réel, aucune commande externe, état persistant inchangé
    - **Property 14: Dry-run sans effet persistant**
    - **Validates: Requirements 6.6**

  - [x] 5.7 Test de propriété : non-interactif sans `--yes`
    - Nouveau fichier `application/internal/cli/noninteractive_test.go` : commande
      à confirmation en mode non-TTY / `--non-interactive` sans `--yes` → arrêt
      sans saisie, exit ≠ 0, message nommant le flag à fournir
    - **Property 15: Non-interactif sans `--yes`**
    - **Validates: Requirements 6.7**

  - [x] 5.8 Test de propriété : parité Plain_Output / JSON_Output
    - Nouveau fichier `application/internal/cli/output_parity_test.go` :
      comparaison ensembliste des champs d'information plain vs json pour les
      Unitary_Command, le Guided_Path et les sorties onboarding/ready → ensembles
      égaux, indépendamment du mode de rendu
    - **Property 16: Parité Plain_Output / JSON_Output**
    - **Validates: Requirements 6.8, 7.5**

- [x] 6. Garde-fous d'onboarding (R7)
  - [x] 6.1 Borne haute du score + innocuité `--check-only`
    - Modifier `application/internal/onboarding/doctor.go` : clamp explicite du
      score `AssessReadiness` dans `[0, 100]` (borne haute) ; garantir que
      `--check-only` n'écrit/modifie/supprime aucun fichier ; tout check non `ok`
      fournit une Guided_Remediation nommant l'outil/clé manquant (sans `panic`)
    - _Requirements: 7.1, 7.2, 7.3, 7.7_

  - [x] 6.2 Invite `asa onboard` sur Mission Control et Runs non onboardés
    - Vérifier/ajouter, dans les écrans Mission Control et Runs, le contenu
      conditionnel invitant explicitement à lancer `asa onboard` quand le dépôt
      n'est pas onboardé (présentation uniquement ; l'UI reste cliente du bus,
      ADR-027)
    - _Requirements: 7.6_

  - [x] 6.3 Test de propriété : non-dégradation de la Readiness
    - Nouveau fichier `application/internal/onboarding/readiness_monotonic_test.go` :
      états initiaux variés ; score après application de l'Onboarding_Wizard
      ≥ score avant
    - **Property 17: Non-dégradation de la Readiness**
    - **Validates: Requirements 7.1, 7.2**

  - [x] 6.4 Test de propriété : `--check-only` n'écrit rien
    - Nouveau fichier `application/internal/onboarding/checkonly_test.go` :
      snapshot du système de fichiers avant/après `--check-only` → aucune
      création/modification/suppression
    - **Property 18: `--check-only` n'écrit rien**
    - **Validates: Requirements 7.3**

  - [x] 6.5 Test de propriété : round-trip save / resume
    - Nouveau fichier `application/internal/onboarding/resume_roundtrip_test.go` :
      états wizard variés ; enregistrer puis `--resume` restaure l'étape courante
      et l'ensemble des réponses à l'identique
    - **Property 19: Round-trip save / resume de l'onboarding**
    - **Validates: Requirements 7.4**

  - [x] 6.6 Test exemple : invites onboard présentes
    - Nouveau fichier `application/internal/onboarding/onboard_invite_test.go` :
      en état non onboardé, les écrans Mission Control et Runs contiennent l'invite
      `asa onboard`
    - _Requirements: 7.6_

- [x] 7. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Remediation_Register + alignement des locales (R3/R8 — AUD-004)
  - [x] 8.1 Faire de `problems.md` le Remediation_Register
    - Modifier `problems.md` : une entrée par constat dans la table
      `| ID | Zone | Problème | Sévérité | Statut |` pour `AUD-001` … `AUD-007`
      (statuts initiaux selon le design) ; statuts dans `{ouvert, en cours,
      clôturé}` ; exactement une entrée pour chaque constat `error`/`blocking`
      (`AUD-001`, `AUD-002`, `AUD-003`)
    - _Requirements: 3.1, 3.2, 3.5_

  - [x] 8.2 Test exemple : cohérence du registre
    - Nouveau fichier `application/internal/policy/register_coherence_test.go` :
      parser `problems.md` → exactement une entrée par constat `error`/`blocking`,
      colonnes non vides, statut unique dans le domaine, aucun `blocking` au
      statut `ouvert` (filet de sécurité de clôture)
    - _Requirements: 3.1, 3.5, 8.6_

  - [x] 8.3 Test de propriété : automate de statut du Remediation_Register
    - Nouveau fichier `application/internal/policy/register_status_test.go` :
      séquences de transitions générées ; seules `ouvert → en cours`,
      `en cours → clôturé`, `clôturé → ouvert` acceptées ; statut toujours dans
      `{ouvert, en cours, clôturé}` (fonction de transition pure définie dans le test)
    - **Property 20: Automate de statut du Remediation_Register**
    - **Validates: Requirements 3.2, 3.4**

  - [x] 8.4 Test de propriété : alignement des ensembles de chemins de locales
    - Nouveau fichier `application/internal/cli/locales_alignment_test.go` :
      pour chaque locale non `en` (`fr`, `de`, `es`), l'ensemble des chemins de
      pages sous `docs-site/content/docs/<loc>/` est égal à celui de `en` **privé
      de** `en/cli/generated/`
    - **Property 21: Alignement des ensembles de chemins de locales**
    - **Validates: Requirements 8.3**

- [x] 9. Clôture : Quality_Gate + régénération-sans-diff + registre (R8)
  - [x] 9.1 Prouver la tranche et clôturer le registre
    - Exécuter le Quality_Gate : `make build`, `go test ./...`, `go vet ./...`,
      `golangci-lint run` (binaire pinné v2) — tous exit 0 ; exécuter la
      Regeneration_Without_Diff (`asa docs generate-cli` vers tmp + `diff` hors
      `meta.json`) → aucune divergence ; mettre à jour `problems.md` : passer les
      `AUD-*` corrigés/vérifiés à `clôturé`, rouvrir toute réapparition en
      signalant la divergence ; vérifier **zéro** constat `blocking` au statut
      `ouvert`
    - _Requirements: 8.1, 8.2, 8.6, 3.3_

  - [x] 9.2 Test exemple : invariants d'architecture (ADR-027)
    - Nouveau fichier `application/internal/cli/arch_invariants_test.go` :
      contrôle statique que `internal/ui` n'importe aucune logique métier
      interdite (routing/policy restent hors UI) et reste cliente du bus
    - _Requirements: 8.5_

- [x] 10. Checkpoint final - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Les sous-tâches marquées `*` sont optionnelles et peuvent être sautées pour un
  MVP plus rapide ; elles ne sont pas implémentées par défaut.
- Chaque tâche référence des exigences granulaires (`R{n}.{m}`) pour la
  traçabilité, et chaque tâche de test PBT référence explicitement une propriété
  `P1` … `P21` du design.
- Les garde-fous (tests) atterrissent **avec** leur correction ; la tâche `9.1`
  **prouve** la tranche complète (Quality_Gate + régénération-sans-diff + clôture
  du registre).
- Aucune nouvelle couche, aucun package `internal/audit`, aucune commande
  `asa audit` : on corrige des constats sur des fichiers existants.
- Critères couverts hors PBT (par tests exemple / CI) : `1.3`, `2.1`–`2.4`,
  `3.1`, `3.5`, `4.8`, `5.1`, `6.1`, `6.2`, `7.6`, `8.1`, `8.2`, `8.4`, `8.5`,
  `8.6`.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "2.1", "3.1", "4.1", "4.2", "4.3", "5.1", "5.2", "6.1", "6.2", "8.1"] },
    { "id": 1, "tasks": ["1.2", "1.3", "1.4", "1.5", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8", "2.9", "3.2", "3.3", "5.3", "5.4", "5.5", "5.6", "5.7", "5.8", "6.3", "6.4", "6.5", "6.6", "8.2", "8.3", "8.4", "9.2"] },
    { "id": 2, "tasks": ["2.10"] },
    { "id": 3, "tasks": ["9.1"] }
  ]
}
```
