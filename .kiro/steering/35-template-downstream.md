---
inclusion: always
---

# Ce dépôt = template amont

- Les projets produits partent souvent de cette base : `.githooks/`, `scripts/install-git-hooks.sh`, `.cursor/`, `.kiro/`, `Makefile`, squelette Go (`application/`, `go.mod`).
- Lors d’une évolution **réutilisable**, assume qu’un dérivé devra **merger ou copier** ; documente les ruptures (breaking) dans le changelog si pertinent.
- Les dépôts clients peuvent définir `GENERIC_TEMPLATE_ROOT` vers ce clone pour les **rappels de drift** (voir leur `.cursor/template-sync.env.example`).
