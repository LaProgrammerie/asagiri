# AGENTS.md — routeur d’entrée

Fichier **court** chargé par Kiro à la racine. Le détail est dans `docs/ai/` ; la carte des sources de vérité : `docs/ai/context-map.md`.  
**Humains :** guide d’usage → `README.md` à la racine.

## Rôle du dépôt

**Template** : base [docker-starter](https://github.com/jolicode/docker-starter) (Docker + [Castor](https://castor.jolicode.com/)) + couche [AI Engineering](https://github.com/LaProgrammerie/ai-engineering-framework) (`docs/ai/`, `.kiro/`, `.cursor/`) + point d’ancrage infra [Yoimachi](https://github.com/LaProgrammerie/yoimachi) (`docs/ai/02-architecture.md`, `infra/yoimachi/`).

**Après fork / copie :** remplace ce paragraphe par **une phrase** sur le produit ou service réel.

## Sources de vérité (résumé)

- **Canon projet :** `docs/ai/*.md`
- **Spec Kiro :** `.kiro/specs/<feature>/` (requirements, design, tasks)
- **Résumé inter-outils :** `docs/ai/active/current-spec.md`
- **Contrat d’exécution** (Cursor, Copilot) : `docs/ai/active/handoff.md`

Voir `docs/ai/context-map.md` pour la carte complète.

## Ordre de lecture

1. `docs/ai/context-map.md`
2. `docs/ai/00-overview.md`
3. `docs/ai/01-product.md` … `05-decisions.md` selon le besoin
4. **Pour implémenter :** `docs/ai/active/handoff.md`, puis `docs/ai/03-standards.md` et les sections utiles de `docs/ai/02-architecture.md`

## Règles d’exécution

- Lire `docs/ai/context-map.md` avant tout travail **non trivial**.
- Pour le code, traiter `docs/ai/active/handoff.md` comme **contrat principal** : ne pas élargir la portée sans mettre à jour la spec et le handoff.
- Tout changement structurel durable → `docs/ai/02-architecture.md` et si besoin `docs/ai/05-decisions.md`.

## Anti-divergence

- Ne pas dupliquer toute l’architecture ici : utiliser `docs/ai/` ou `.kiro/specs/`.
- Changement **matériel** sous `.kiro/specs/` → mettre à jour `current-spec.md` ; si le périmètre d’implémentation change → `handoff.md`.

## Où trouver quoi

| Besoin | Emplacement |
|--------|-------------|
| Carte de contexte | `docs/ai/context-map.md` |
| Règles Cursor | `.cursor/rules/*.mdc` |
| Steering Kiro projet | `.kiro/steering/` |
| Skills dépôt | `.kiro/skills/` |
| Skills globales | `~/.kiro/skills/` |
| Copilot | `.github/copilot-instructions.md` |
| Infra déployable | `docs/ai/02-architecture.md`, `infra/yoimachi/` |
| Docker / Castor (long) | `README.docker-starter.md` |

## Invariants

- Ne pas committer de secrets.
- Pas de dérive entre `.kiro/specs/`, `current-spec.md` et `handoff.md` : synchroniser dans le même flux quand c’est possible.
- Respecter les limites dans `docs/ai/02-architecture.md` et `01-product.md` une fois remplis.
- Décision durable sur l’architecture ou les standards → `05-decisions.md` + fichiers `docs/ai/` affectés.
