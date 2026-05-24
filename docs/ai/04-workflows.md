# Workflows projet

Doc opérationnelle pour l’équipe : plus de détail que le [`README.md`](../../README.md) racine.  
Les procédures **génériques** (review, release, debug, plan, refactor) restent dans **`~/.kiro/skills/`**.

---

## Quotidien : spec → handoff → implémentation

1. **Spécifier (Kiro)**  
   Travail dans `.kiro/specs/<feature>/` (requirements → design → tasks).

2. **Résumer**  
   Mettre à jour `docs/ai/active/current-spec.md` dès qu’un changement **matériel** de périmètre ou de critères d’acceptation.

3. **Cadrer l’exécution**  
   Mettre à jour `docs/ai/active/handoff.md` avant une session Cursor ou une PR (scope, fichiers, plan, tests, DoD).  
   Skill dépôt : `.kiro/skills/create-handoff/` si besoin.

4. **Implémenter**  
   **Handoff d’abord**, puis `03-standards.md` et les sections utiles de `02-architecture.md`.

5. **Boucler**  
   Si le code révèle un trou de spec : corriger **Kiro + projections** (`current-spec`, `handoff`) avant de seulement patcher le code.

---

## Quoi mettre à jour quand

| Type de changement | Action |
|--------------------|--------|
| Besoin métier / produit | `01-product.md` |
| Nouveau module, service Docker, binaire | `02-architecture.md` + `05-decisions.md` si durable |
| Nouvelle commande, linter, politique de tests | `03-standards.md` + `Makefile` |
| Convention d’équipe (branches, release) | Ce fichier + `05-decisions.md` si structurel |
| Tasks / design de la spec active | `.kiro/specs/...` → `current-spec.md` → `handoff.md` si besoin |
| Prochaine session de code seulement | `handoff.md` |

Règle simple : **si un humain ou un agent pourrait se tromper de périmètre demain, mets à jour le fichier canon aujourd’hui.**

---

## Développement local

1. Cloner le dépôt, installer **Go** (version `go.mod`).
2. `go mod download`
3. `make build && ./bin/agentflow init && ./bin/agentflow doctor`
4. Optionnel : `make dev` — stack Docker.
5. Boucle feature : `plan` → `enrich` → `dev` → `verify` → `review` → `report` / `pr` (ajouter `--dry-run` sans agents installés).
6. Avant PR : `go test -race ./...` ; `make lint` si golangci-lint disponible (Go ≥ version module).

Sans Docker : `./bin/agentflow` directement sur l’hôte.

---

## Review et merge

*(Branches, PRs, critères d’acceptation.)*  
Review structurée : skill `code-review` dans `~/.kiro/skills/`.

## Release

*(Versioning, changelog, déploiement.)*  
Checklist : skill `release-checklist` dans `~/.kiro/skills/`.

## Incidents / debug prod

*(Runbooks courts.)*  
Diagnostics : skill `debugging` dans `~/.kiro/skills/`.

---

## Maintenir le « système de contexte »

Changement d’outillage (nouvelle règle Cursor, steering Kiro, hooks) :

1. Éditer `.cursor/` ou `.kiro/` selon le cas.
2. Si règle **durable** d’équipe : ligne dans `05-decisions.md` + `context-map.md` ou ce fichier si le workflow humain change.

Pour « à quoi sert chaque dossier », voir [`context-map.md`](context-map.md) et le [README](../../README.md).
