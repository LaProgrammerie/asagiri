# Requirements Document

> Feature : audit-coherence-consolidation
> Nature : **correction & simplification**, pas un nouveau sous-système.

## Introduction

Asagiri (binaire `asa`) est un orchestrateur CLI local-first, écrit en Go, pour
les workflows de développement agentique : `spec → tasks → worktrees → agents →
validation → review`. Le canon documentaire vit dans `docs/ai/` ; le site public
dans `docs-site/` (Fumadocs, locales `en/fr/de/es`). La dérive doc/code/spec est
suivie dans `problems.md`.

Un audit réel a déjà été mené sur le dépôt courant (voir
[`audit-report.md`](./audit-report.md)). Cet audit est une **activité ponctuelle
terminée**, pas un composant à construire. Il a relevé sept constats stables,
`AUD-001` à `AUD-007`, dont la dérive est concentrée et de cause unique : la
commande `runs` (ajoutée par cockpit-consolidation, ADR-029) n'a pas été suivie
d'une régénération de la doc CLI.

Cette feature **n'ajoute aucune surcouche** et **ne crée aucun moteur d'audit
runtime**. Elle a un objectif inverse : appliquer les **corrections concrètes**
des constats `AUD-001` à `AUD-007`, **simplifier** la couche d'orchestration
perçue comme opaque, et poser les **garde-fous** qui empêchent la régression. La
vision cible reste : l'utilisateur décrit un besoin → les outils transforment ce
besoin en artefacts → des Agent_Backend (Kiro, Ollama, Cursor CLI) collaborent
dans des workflows → l'humain valide aux jalons clés.

Les exigences couvrent : la parité CLI ↔ doc générée restaurée avec garde-fou
anti-régression (`AUD-001`/`AUD-002`), le gate lint réparé (`AUD-003`), le
registre de dérive à jour (`AUD-004`), un routing config-driven et explicable
(`AUD-005`), la cohérence policy ↔ canon (`AUD-006`), la restauration de la
simplicité du chemin guidé sans retirer l'unitaire (`AUD-007`), l'UX
d'onboarding, et la maturité production / open source sans dérive résiduelle.

Cette feature **ne réalise aucune modification de code pendant la phase
requirements** ; elle définit le contrat. Les corrections de code seront menées
en phases design/tasks, en respectant `docs/ai/active/handoff.md`,
`02-architecture.md` et `05-decisions.md`.

## Glossary

- **Asagiri / `asa`** : le produit et son binaire CLI Go.
- **Command_Tree** : l'ensemble des commandes Cobra atteignables sous la racine
  `asa` (y compris `runs`, ajoutée par ADR-029).
- **Unitary_Command** : commande CLI exécutant une étape de workflow de façon
  autonome (ex. `asa spec`, `asa plan`, `asa enrich`, `asa dev`, `asa verify`,
  `asa review`), utilisable seule sans dépendre du chemin guidé.
- **Docgen** : le générateur `asa docs generate-cli` qui produit les pages MDX de
  référence depuis le Command_Tree, vers
  `docs-site/content/docs/en/cli/generated/`.
- **Regeneration_Without_Diff** : propriété selon laquelle relancer Docgen sur le
  Command_Tree courant ne produit aucune différence (chemins et contenu) avec les
  pages MDX committées.
- **Regeneration_Check** : le test/contrôle CI (cible) qui échoue si Docgen
  produirait une quelconque différence avec les pages committées.
- **Quality_Gate** : l'ensemble des contrôles de readiness production
  `make build`, `go test ./...`, `go vet ./...` et `golangci-lint run`.
- **Remediation_Register** : le registre des corrections, matérialisé dans
  `problems.md`, reliant chaque constat corrigible à une action et un statut
  parmi `ouvert`, `en cours`, `clôturé`.
- **Finding / Constat** : une observation d'audit identifiée par un ID stable
  (`AUD-001` … `AUD-007`), portant zone, description, sévérité et statut.
- **Router** : la fonction de décision de routage `routing.Route`
  (`application/internal/routing/router.go`).
- **Routing_Decision** : la sortie du Router — Agent_Backend, modèle,
  `Local`, classe d'étape (`StepClass`) et raison (`Reason`).
- **Agent_Backend** : un backend d'agent nommé et déclaré dans `config.agents`
  (ex. `cursor`, `claude`, `ollama`, `kiro`).
- **Policy_Coherence_Check** : le contrôle (cible) qui vérifie que les listes de
  rôles autorisés/interdits d'Ollama (`internal/policy/ollama.go`) sont alignées
  sur le canon courant.
- **Guided_Path** : le chemin guidé unique mis en avant
  (`onboarding → work → jalons de validation`), distinct des Unitary_Command.
- **Guided_Remediation** : message actionnable retourné lorsqu'un outil, une
  config ou une dépendance manque, nommant l'élément manquant et l'action de
  résolution.
- **Sensitive_Action** : action exigeant une validation humaine renforcée
  (paiement, données personnelles, authentification, permissions, suppression,
  export, migration de données).
- **Onboarding_Wizard** : le parcours `asa onboard` (CLI et TUI `--ui`) de
  préparation du dépôt vers un état prêt.
- **Readiness** : l'état de préparation du dépôt, mesuré par `asa ready` /
  `asa doctor`, persisté dans `.asagiri/onboarding/report.json` (score 0–100 et
  contrôles `ok` / `warn` / `fail`).
- **Plain_Output** / **JSON_Output** : modes de sortie texte plat (compatible CI)
  et JSON structuré, devant rester en parité d'information.

## Requirements

### Requirement 1: Parité CLI ↔ documentation générée restaurée et garde-fou anti-régression

**User Story:** En tant qu'utilisateur de la documentation, je veux que chaque
commande CLI atteignable ait exactement une page de référence à jour et qu'un
garde-fou empêche toute régression, afin de ne jamais rencontrer une commande non
documentée ni une page périmée. (AUD-001, AUD-002)

#### Acceptance Criteria

1. WHEN Docgen est exécuté sur le Command_Tree courant, THE Docgen SHALL produire
   exactement une page MDX sous `docs-site/content/docs/en/cli/generated/` pour
   chaque commande atteignable du Command_Tree, y compris la commande `runs`.
2. WHEN Docgen est exécuté deux fois sur un Command_Tree identique, THE Docgen
   SHALL produire des fichiers MDX identiques octet pour octet, mêmes chemins et
   même contenu.
3. WHEN Docgen est régénéré après la correction, THE référence CLI générée SHALL
   satisfaire Regeneration_Without_Diff vis-à-vis des pages committées.
4. THE Regeneration_Check SHALL retourner un code de sortie non nul s'il existe
   au moins une différence entre les pages MDX committées et celles régénérées
   par Docgen, et un code de sortie nul en l'absence de toute différence.
5. WHEN la Regeneration_Check s'exécute après ajout d'une commande au
   Command_Tree sans régénération de Docgen, THE Regeneration_Check SHALL échouer
   en listant les fichiers MDX manquants ou divergents.
6. WHEN la correction `AUD-002` est appliquée, THE page MDX de chaque commande
   sœur SHALL inclure le lien fratrie `> - [Runs](./runs.mdx)`.

### Requirement 2: Gate de lint réparé et opérationnel

**User Story:** En tant que mainteneur préparant l'open source, je veux que
`golangci-lint` s'exécute réellement sur la cible Go du dépôt, afin de prouver la
readiness production via un gate qualité opérationnel. (AUD-003)

#### Acceptance Criteria

1. WHEN `golangci-lint run` est exécuté sur le code Go sous `application/`, THE
   Quality_Gate SHALL exécuter l'analyse statique jusqu'à son terme sans erreur de
   version de toolchain Go.
2. THE Quality_Gate SHALL être documenté par une commande exacte et reproductible
   permettant d'exécuter `golangci-lint` aligné sur la cible `go 1.25.0` déclarée
   dans `go.mod`.
3. WHEN `make build`, `go test ./...`, `go vet ./...` et `golangci-lint run` sont
   exécutés, THE readiness production SHALL être considérée atteinte si et
   seulement si les quatre commandes terminent toutes avec un code de sortie égal
   à 0.
4. IF la version de toolchain utilisée pour `golangci-lint` est inférieure à la
   cible déclarée dans `go.mod`, THEN THE Quality_Gate SHALL signaler une erreur
   explicite nommant la version attendue et la version détectée.

### Requirement 3: Registre de dérive à jour (problems.md)

**User Story:** En tant que mainteneur, je veux que `problems.md` reflète l'état
réel et trace chaque constat bloquant, afin que rien ne reste en suspens avant la
livraison. (AUD-004)

#### Acceptance Criteria

1. THE Remediation_Register SHALL contenir exactement une entrée pour chacun des
   constats de sévérité `error` ou `blocking` (`AUD-001`, `AUD-002`, `AUD-003`),
   reliant l'identifiant stable du constat à une action de correction et à un
   statut.
2. THE Remediation_Register SHALL porter, pour chaque entrée, la zone, la
   description, la sévérité et exactement une valeur de statut parmi `ouvert`,
   `en cours` ou `clôturé`, les transitions autorisées étant
   `ouvert → en cours → clôturé` et la réouverture `clôturé → ouvert`.
3. WHEN un constat est corrigé et vérifié, THE Remediation_Register SHALL marquer
   l'entrée correspondante `clôturé` et SHALL refléter cet état sans divergence
   avec la réalité du dépôt.
4. IF un constat précédemment marqué `clôturé` réapparaît lors d'une vérification
   ultérieure, THEN THE Remediation_Register SHALL rouvrir l'entrée correspondante
   au statut `ouvert` et signaler la divergence.
5. THE Remediation_Register SHALL ne laisser aucun constat de sévérité `blocking`
   au statut `ouvert` absent du registre à la clôture de la tranche.

### Requirement 4: Routing config-driven et explicable

**User Story:** En tant qu'utilisateur, je veux que le choix d'agent soit une
décision pure, pilotée par la configuration et explicable, afin de retrouver la
simplicité « décrire → produire → valider » sans subir une surcouche IA opaque.
(AUD-005)

#### Acceptance Criteria

1. WHEN le Router est appelé deux fois avec la même configuration, la même classe
   d'étape et la même combinaison de flags, THE Router SHALL produire des
   Routing_Decision identiques sur tous les champs (Agent_Backend, modèle,
   `Local`, classe d'étape, raison), sans effet de bord observable.
2. THE Routing_Decision SHALL toujours désigner un Agent_Backend déclaré dans
   `config.agents`, sans nom d'agent codé en dur hors configuration.
3. IF `preferLocal` est demandé ou que la classe d'étape figure dans
   `prefer_local_for`, et que `noCloud` n'est pas demandé, THEN THE Router SHALL
   produire une Routing_Decision avec `Local = true` et la raison `prefer_local`.
4. IF `noCloud` est demandé, THEN THE Router SHALL produire une Routing_Decision
   avec `Local = true` et la raison `no_cloud`, cette règle prévalant sur
   `allowCloud` et sur `preferLocal`.
5. WHEN ni `noCloud` ni `preferLocal` ne sont demandés, THE Router SHALL appliquer
   la stratégie par défaut et exposer la raison `cloud_heavy`, `cloud_fast` ou
   `default` selon la classe d'étape.
6. THE Routing_Decision SHALL exposer une raison lisible appartenant à l'ensemble
   `{prefer_local, no_cloud, cloud_heavy, cloud_fast, default}` expliquant le
   choix.
7. IF aucun Agent_Backend déclaré dans `config.agents` ne correspond à la
   Routing_Decision calculée, THEN THE Router SHALL retourner une erreur guidée,
   ne sélectionner aucun backend non déclaré, et terminer sans `panic`.
8. WHEN l'utilisateur demande une explication d'exécution (sortie de type
   `asa explain`), THE Asagiri SHALL nommer l'Agent_Backend retenu et la raison,
   sans exiger de l'utilisateur la connaissance des backends sous-jacents.

### Requirement 5: Cohérence policy ↔ canon

**User Story:** En tant que mainteneur, je veux que les rôles Ollama
autorisés/interdits soient reliés au canon courant et vérifiés, afin d'éliminer
toute référence périmée et toute dérive silencieuse. (AUD-006)

#### Acceptance Criteria

1. THE listes de rôles Ollama autorisés et interdits (`internal/policy/ollama.go`)
   SHALL référencer le canon courant et non une référence de spec historique
   périmée (`§10.1`/`§10.2`).
2. WHEN le Policy_Coherence_Check est exécuté, THE Policy_Coherence_Check SHALL
   vérifier que chaque rôle autorisé et interdit correspond à une entrée du canon
   courant.
3. IF un rôle de policy ne correspond plus au canon courant, THEN THE
   Policy_Coherence_Check SHALL retourner un code de sortie non nul en nommant le
   rôle divergent.
4. WHEN un rôle interdit est demandé pour Ollama, THE Asagiri SHALL refuser
   l'opération en retournant une erreur nommant le rôle concerné, sans `panic`.

### Requirement 6: Restauration de la simplicité du chemin guidé

**User Story:** En tant qu'utilisateur, je veux un chemin guidé unique et
découvrable, sans perdre aucune commande unitaire, afin de décrire un besoin et
d'être accompagné aux jalons sans me noyer dans la surface CLI. (AUD-007)

#### Acceptance Criteria

1. THE Asagiri SHALL exposer dans l'aide racine et la documentation d'entrée un
   Guided_Path unique et découvrable suivant l'enchaînement
   `onboarding → work → jalons de validation`.
2. THE Asagiri SHALL conserver chaque Unitary_Command existante exécutable de
   façon autonome, sans qu'aucune commande unitaire ne soit retirée par la mise en
   avant du Guided_Path.
3. IF un outil, une configuration ou une dépendance requis est absent lors d'une
   commande, THEN THE Asagiri SHALL s'arrêter sans `panic`, retourner un code de
   sortie non nul, et émettre une Guided_Remediation nommant l'élément manquant et
   au moins une action de résolution.
4. IF une commande s'arrête à cause d'un prérequis manquant, THEN THE Asagiri
   SHALL préserver l'état du dépôt inchangé, sans écrire d'artefact partiel.
5. WHEN un jalon de validation humaine est atteint (confirmation de plan,
   dépassement de budget, Sensitive_Action), THE Asagiri SHALL exiger une réponse
   affirmative explicite avant d'exécuter l'étape concernée.
6. WHILE le mode `--dry-run` est actif, THE Asagiri SHALL simuler chaque étape
   sans invoquer d'Agent_Backend réel, sans exécuter de commande externe et sans
   modifier l'état persistant du dépôt.
7. IF la sortie n'est pas interactive (CI ou `--non-interactive`) et qu'une
   confirmation est requise sans `--yes`, THEN THE Asagiri SHALL s'arrêter sans
   attendre de saisie, retourner un code de sortie non nul et nommer le flag à
   fournir.
8. THE Guided_Path et les Unitary_Command SHALL conserver la parité Plain_Output /
   JSON_Output sans conditionner les champs d'information au mode de rendu.

### Requirement 7: UX d'onboarding

**User Story:** En tant que nouvel utilisateur, je veux un onboarding simple et
guidé (CLI et TUI) qui mène le dépôt à un état prêt, afin de démarrer sans
comprendre toute la mécanique interne.

#### Acceptance Criteria

1. WHEN `asa onboard` s'exécute (CLI ou TUI `--ui`) sur un dépôt dont la stack est
   détectable, THE Onboarding_Wizard SHALL mener le dépôt à un état prêt sans
   dégrader la Readiness existante.
2. WHEN l'Onboarding_Wizard applique sa configuration, THE score de Readiness
   SHALL être supérieur ou égal au score précédant l'application.
3. THE Onboarding_Wizard SHALL fournir un mode `--check-only` qui évalue la
   Readiness sans créer, modifier ni supprimer aucun fichier.
4. WHEN l'utilisateur enregistre l'état du wizard puis reprend avec `--resume`,
   THE Onboarding_Wizard SHALL restaurer l'étape courante et l'ensemble des
   réponses déjà collectées identiques à l'état enregistré.
5. THE Onboarding_Wizard SHALL conserver un Plain_Output et un JSON_Output en
   parité d'information.
6. WHERE le dépôt n'est pas onboardé, THE écran Mission Control et l'écran Runs
   SHALL inviter explicitement à lancer `asa onboard`.
7. IF la stack du dépôt n'est pas détectable, THEN THE Onboarding_Wizard SHALL
   retourner une Guided_Remediation indiquant comment renseigner la stack
   manuellement, sans écrire de configuration partielle et sans `panic`.

### Requirement 8: Maturité production et open source sans dérive résiduelle

**User Story:** En tant que mainteneur, je veux une barre de qualité production
sans dérive documentaire résiduelle, afin de publier l'outil en open source sans
réserve.

#### Acceptance Criteria

1. WHEN la tranche de correction est livrée, THE référence CLI générée SHALL
   satisfaire Regeneration_Without_Diff vis-à-vis du Command_Tree courant.
2. WHEN `make build`, `go test ./...`, `go vet ./...` et `golangci-lint run` sont
   exécutés sur le dépôt de livraison, THE Quality_Gate SHALL terminer chaque
   commande avec un code de sortie égal à 0.
3. THE site de documentation `docs-site/content/docs/` SHALL conserver, pour
   chaque locale non `en` (`fr`, `de`, `es`), un ensemble de chemins de pages
   égal à celui de `en` privé de la référence CLI générée `en`-only.
4. THE Asagiri SHALL ne committer aucun secret en clair (clés d'API, jetons,
   identifiants, clés privées).
5. THE Asagiri SHALL préserver les invariants d'architecture existants : l'UI
   reste cliente du bus (ADR-027), aucune logique métier n'est introduite dans
   `internal/ui`, et le moteur reste local-first et déterministe (sorties
   identiques pour des entrées identiques).
6. THE Remediation_Register SHALL ne laisser aucun constat de sévérité `blocking`
   au statut `ouvert` à la livraison.
