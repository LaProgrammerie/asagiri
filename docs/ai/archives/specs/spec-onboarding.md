Spec H — Project Onboarding & Readiness

1. Vision

Asagiri exige aujourd’hui trop de configuration manuelle avant le premier `asa work` utile :
copie de `config.yaml`, adaptation des commandes de validation, vérification des agents,
remplissage partiel de `docs/ai/`, création d’une première spec Kiro.

**Project Onboarding** comble ce gap : un parcours guidé qui prépare un dépôt (neuf ou
template forké) pour que le moteur Asagiri et l’Experience Platform puissent travailler
immédiatement sur les besoins métier.

Principe clé :

```
asa onboard  →  détecte  →  configure  →  valide  →  ready
                    ↓
              asa work / Mission Control
```

L’onboarding **ne remplace pas** la spec produit ni le handoff d’exécution : il garantit
que le **socle technique et documentaire minimal** existe et est cohérent avec la stack du dépôt.

⸻

2. Positionnement

2.1 Problème utilisateur

| Situation actuelle | Impact |
|--------------------|--------|
| `asa init` copie un template Go | Projets PHP/Node mal configurés |
| `asa doctor` vérifie peu de choses | Agents manquants découverts tard |
| UI Mission Control sans état « prêt » | Impression que l’UI devrait tout faire |
| `docs/ai/` placeholders vides | `asa work` / intent sans contexte |
| Pas de spec Kiro | Planner sans feature cible |

2.2 Expériences cibles

| Canal | Rôle |
|-------|------|
| **CLI interactive** | `asa onboard` — wizard questions/réponses |
| **CLI non-interactive** | `asa onboard --yes --stack auto` — CI / scripts |
| **TUI** | écran Onboarding dans Mission Control si non ready |
| **Readiness** | `asa ready` — score + checklist JSON/plain |

2.3 Non-objectifs V1

- Génération automatique du métier (features chatbot, auth, etc.)
- Remplacement de Kiro pour spécifier le produit
- Installation des binaires agents (cursor-agent, codex…) — détection + instructions seulement
- Migration de projets sans Git

⸻

3. Non-négociables

3.1 Parité CLI / TUI (ADR-027)

Toute action onboarding TUI a un équivalent CLI documenté.

3.2 Logique métier hors UI

Package `internal/onboarding/` ; UI consomme CommandBus/QueryBus uniquement.

3.3 Modes `--plain`, `--json`, `--ci`

Sorties structurées pour scripts et CI (ex. pipeline « repo prêt pour agents »).

3.4 Idempotent

Relancer `asa onboard` ne casse pas une config validée ; propose mise à jour ou skip.

3.5 Pas de secrets en clair

Config générée sans token ; références env (`NOTION_TOKEN`, etc.).

⸻

4. Flux utilisateur

4.1 Parcours nominal (dépôt template forké, ex. chatbot PHP)

```bash
cd mon-projet
git init   # si besoin
asa init   # conservé ; crée .asagiri/ minimal
asa onboard
```

Étapes wizard :

1. **Bienvenue** — explique ce que fait / ne fait pas l’onboarding
2. **Projet** — nom, branche default, description une ligne
3. **Stack** — auto-détectée + confirmation utilisateur
4. **Validation** — commandes QA proposées (Castor, go test, npm…)
5. **Agents** — choix default_agent ; test présence binaires
6. **Sources** — chemins specs (`.kiro/specs`, `docs/ai/active`)
7. **Canon docs** — remplir squelettes `01-product.md`, `current-spec.md`, handoff stub
8. **Première feature** — nom feature Kiro (ex. `chatbot-mvp`) + création `.kiro/specs/<feature>/`
9. **Récap** — diff config ; confirmation
10. **Validation finale** — doctor étendu + `ready` score

4.2 Parcours TUI

- Si `asa ready` ≠ OK et TTY : bannière Mission Control + entrée palette « Complete onboarding »
- **`asa onboard --ui`** : wizard TUI interactif full-screen (formulaire Bubble Tea, pas bilan read-only)
  - Étapes : Welcome → Project → Stack → Agents → Docs → Feature → Review → Apply → Ready
  - Champs préremplis (detect, config, dirname) ; navigation Prev/Next ; panneau Advanced repliable
  - Apply écrit config + bootstrap docs ; affiche score readiness
- Route Mission Control : `ScreenOnboarding` (même Model, mode wizard si lancé via `--ui`)

4.3 Reprise

- `asa onboard --resume` reprend le questionnaire (état `.asagiri/onboarding/state.json`)
- `asa onboard --check-only` = alias readiness sans mutation

⸻

5. Détection de stack

5.1 Détecteurs (ordre de priorité, cumulables)

| Signal | Stack | Validation proposée |
|--------|-------|---------------------|
| `go.mod` à la racine ou `application/` | Go | `go test ./...`, `go vet ./...`, `golangci-lint run` |
| `castor.php` | PHP Castor | `castor qa:static-checks`, `castor qa:phpunit` |
| `composer.json` dans `application/` | PHP Composer | `composer test` si script défini |
| `package.json` | Node | `npm test` / script `qa:js` si présent |
| `Makefile` avec cibles test/lint | Generic | cibles extraites si sûres |
| `docker-compose.yml` / `infrastructure/docker/` | Docker | note prérequis `castor start` / compose |

5.2 Interface

```go
type StackDetector interface {
    Name() string
    Detect(repoRoot string) (StackMatch, error)
    ProposeValidation(match StackMatch) []ValidationCommand
    ProposeConfigPatch(match StackMatch, answers Answers) ConfigPatch
}
```

5.3 Registre extensible

Détecteurs enregistrés dans `internal/onboarding/detect/` ; V1 : Go, Castor/PHP, Node.

⸻

6. Génération & validation config

6.1 Entrées

- `config.yaml.example` (base)
- résultat détecteurs + réponses wizard
- config existante (merge, pas écrasement aveugle)

6.2 Champs minimum générés

```yaml
project:
  name: <from answers or dir name>
  default_branch: main

validation:
  commands: <stack-specific>

work:
  default_agent: cursor
  ...

specs:
  kiro_path: .kiro/specs
  active_spec_path: docs/ai/active/current-spec.md
  handoff_path: docs/ai/active/handoff.md

worktrees:
  branch_prefix: <project-slug>
```

6.3 Merge policy

| Clé | Comportement |
|-----|--------------|
| Absente | écriture |
| Valeur template default (`my-project`, `go test`) | remplacement proposé |
| Valeur custom utilisateur | conservée ; wizard signale « déjà configuré » |
| Conflit validation | liste côte à côte ; utilisateur choisit |

6.4 Validation post-écriture

Réutiliser `config.Load` + `config.Validate` ; erreurs affichées avec fix suggéré.

⸻

7. Doctor étendu & readiness

7.1 `asa doctor` (extension)

Conserver checks actuels + ajouter :

| Check | Sévérité |
|-------|----------|
| git propre (optionnel warn) | warn |
| `.gitignore` contient `.asagiri/state.sqlite`, `worktrees/` | error |
| `validation.commands` exécutable (dry probe ou `--skip-exec`) | warn |
| Agent `work.default_agent` présent dans PATH | warn |
| `docs/ai/01-product.md` non placeholder | warn |
| `.kiro/specs/` contient ≥1 feature | warn |
| Docker daemon (si stack docker) | warn |
| Conflit binaire `asa` système (macOS `/usr/bin/asa`) | info |

7.2 `asa ready`

Retourne un **ReadinessReport** :

```json
{
  "ready": false,
  "score": 72,
  "checks": [
    {"id": "config.valid", "status": "ok"},
    {"id": "agents.cursor", "status": "warn", "message": "cursor-agent not in PATH"},
    {"id": "spec.kiro", "status": "fail", "message": "no feature under .kiro/specs/"}
  ],
  "next_actions": [
    {"title": "Create first Kiro spec", "cli": "asa onboard --step feature"}
  ]
}
```

Seuil `ready: true` : tous checks **error** = ok ; warns acceptés sauf `--strict`.

⸻

8. Bootstrap canon documentaire

8.1 Fichiers touchés (avec consentement)

| Fichier | Action |
|---------|--------|
| `docs/ai/01-product.md` | injecte nom, problème, utilisateurs (depuis Q&R) |
| `docs/ai/03-standards.md` | tableau commandes QA aligné validation |
| `docs/ai/active/current-spec.md` | pointe vers feature créée |
| `docs/ai/active/handoff.md` | stub objectif + plan placeholder |
| `AGENTS.md` | remplace paragraphe template par une phrase produit |
| `.kiro/specs/<feature>/requirements.md` | squelette minimal |
| `.kiro/specs/<feature>/design.md` | squelette minimal |
| `.kiro/specs/<feature>/tasks.md` | tâche 1 « bootstrap env » |

8.2 Garde-fous

- Jamais écraser un fichier déjà substantiel (> N lignes non-placeholder)
- `--force-docs` pour override explicite
- Diff preview avant écriture

⸻

9. Intégration UI (lot 2)

9.1 QueryBus

- `GetReadinessQuery` → ReadinessReport
- `GetOnboardingStateQuery` → step, answers partielles

9.2 CommandBus

- `RunOnboardingStepCommand`
- `ApplyOnboardingConfigCommand`
- `SkipOnboardingCheckCommand` (warn → acknowledged)

9.3 Mission Control

- Bandeau « Projet non prêt (72%) » + CTA
- Palette : `Complete onboarding`, `Run doctor`, `Show readiness`

⸻

10. CLI

10.1 Commande principale

```
asa onboard [flags]

Flags:
  --yes                 Accept defaults
  --non-interactive     No prompts (requires --yes)
  --stack auto|go|php|node  Override detection
  --check-only          Readiness only, no writes
  --resume              Resume saved wizard state
  --step <name>         Jump to step (project|stack|agents|docs|feature|validate)
  --ui                  Open TUI wizard
  --plain --json --ci   Output modes
  --strict              Treat warns as failures in ready
  --dry-run             Show planned changes only
```

10.2 Alias

- `asa ready` → `asa onboard --check-only`
- `asa doctor --full` → doctor + onboarding checks

⸻

11. Artefacts

```
.asagiri/
  onboarding/
    state.json          # wizard progress (gitignored)
    report.json         # last readiness report
    backups/
      config.yaml.<ts>  # before overwrite
```

⸻

12. Critères d’acceptation

12.1 CLI

- [ ] Dépôt PHP Castor vierge : `asa onboard --yes` produit config avec `castor qa:*`
- [ ] Dépôt Go : validation Go par défaut
- [ ] Relance idempotente sans duplication commandes validation
- [ ] `asa ready --json` parseable en CI
- [ ] `asa onboard --dry-run` liste fichiers modifiés sans écrire

12.2 Doctor / ready

- [ ] Détecte absence cursor-agent avec message actionable
- [ ] Détecte conflit `/usr/bin/asa` sur macOS

12.3 Docs

- [ ] Crée `.kiro/specs/<feature>/` minimal
- [ ] Ne écrase pas handoff rempli manuellement

12.4 UI (lot 2)

- [ ] Mission Control affiche readiness si score < 100
- [ ] Parité CLI pour chaque action TUI

12.5 Tests

- [ ] Tests unitaires détecteurs (fixtures go/php/node)
- [ ] Test intégration temp dir : init → onboard → ready true
- [ ] Golden `asa ready --plain`

⸻

13. Phasage livraison

| Lot | Contenu |
|-----|---------|
| **1 — Core CLI** | `internal/onboarding/`, detecteurs, writer, `asa onboard`, `asa ready`, doctor étendu |
| **2 — TUI** | écran onboarding, Mission Control bandeau, bus |
| **3 — Polish** | `--resume`, backups config, docs-site 4 locales, `asa docs generate-cli` |

⸻

14. Références

- ADR-005 — validation externe configurable
- ADR-027 — Experience Platform (parité CLI/TUI)
- `spec-ui.md` §3.1 — équivalent CLI obligatoire
- Template downstream : `.kiro/steering/35-template-downstream.md`
