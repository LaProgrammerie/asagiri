---
fileMatch:
  - "**/*.php"
  - "**/*.{ts,tsx,js,jsx,mjs,cjs}"
---

# Stack applicative (template)

Base **PHP** + Docker/Castor ; **Node** possible (builder ou service dédié). L’architecture cible est dans `docs/ai/02-architecture.md`.

## Conventions

- **PHP :** typage strict là où le projet l’adopte ; pas de shortcuts sur les erreurs aux frontières publiques (HTTP, CLI).
- **TS/JS :** si présent, éviter `any` non justifié ; alignement avec `03-standards.md`.

## Pour Kiro / implémentation

- Ne pas introduire de dépendances entre couches hors de ce que `docs/ai/02-architecture.md` autorise.
- Les changements d’infra Docker ou Yoimachi doivent rester cohérents avec le canon architecture (`docs/ai/02-architecture.md`).
