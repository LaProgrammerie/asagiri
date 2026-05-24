# Vue d’ensemble du contexte

## Rôle du dépôt

Template **PHP** (par défaut) avec **Docker** et **Castor** ([docker-starter](https://github.com/jolicode/docker-starter)), complété par le workflow **spec → handoff → code** ([AI Engineering](https://github.com/LaProgrammerie/ai-engineering-framework)) et un emplacement pour l’infra générée par **Yoimachi**.

*(À adapter : mission produit, contraintes, public.)*

## Carte du système de contexte

Pour les **sources de vérité**, Kiro vs Cursor vs Copilot, et specs vs `docs/ai/active/`, lire **`context-map.md`** en premier.

## Chemins importants

| Chemin | Rôle |
|--------|------|
| `AGENTS.md` | Routeur court (toujours inclus par Kiro) |
| `docs/ai/context-map.md` | Carte des fichiers et anti-dérive |
| `docs/ai/` | Canon projet détaillé |
| `docs/ai/active/` | Résumé de spec + handoff d’exécution |
| `docs/ai/02-architecture.md` | Canon de stack (Castor, docker-starter, Yoimachi) |
| `.kiro/specs/` | Artefacts Kiro (requirements, design, tasks) |
| `.kiro/steering/` | Contexte et règles ciblés Kiro |
| `.kiro/skills/` | Skills **dépôt** (ex. create-handoff) |
| `.cursor/rules/` | Règles Cursor courtes |
| `castor.php`, `infrastructure/docker/` | Stack locale Docker |

## Liens rapides

- Produit : `01-product.md`
- Architecture : `02-architecture.md`
- Spec / handoff : `active/current-spec.md`, `active/handoff.md`
