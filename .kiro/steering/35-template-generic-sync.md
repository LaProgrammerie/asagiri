---
inclusion: always
---

# Synchronisation template (Generic project)

## Apres une modification

- Si tu touches `.githooks/`, `scripts/install-git-hooks.sh`, conventions `.cursor/` ou `.kiro/` partagees : demander si la modif est **specifique** ou **generique**.
- Si **generique** : planifier le port vers le depot template (Generic project), sans donnees sensibles ni logique metier.
- Pour le port : decider entre **copie exacte** et **adaptation du principe** ; l'alignement recherche est la logique, pas l'identite fichier-a-fichier.

## En debut de session / avant une grosse tache

- Si `GENERIC_TEMPLATE_ROOT` est configure (`.cursor/template-sync.env`) : considerer les derniers commits et fichiers recents du template (toutes zones).
- Si le template a avance : proposer quoi importer (import exact, adaptation, ou rejet justifie) via copie ciblee/cherry-pick/tache dediee.

## Documentation

- Carte : `docs/ai/context-map.md`.
- Hooks Cursor : `.cursor/hooks.json` + scripts sous `.cursor/hooks/`.
