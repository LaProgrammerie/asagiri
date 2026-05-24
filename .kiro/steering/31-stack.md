---
fileMatch:
  - "**/*.go"
---

# Stack applicative (Go)

Module Go à la racine ; code sous `application/`. Runtime local optionnel via Docker Compose — voir `docs/ai/02-architecture.md`.

## Conventions

- `gofmt` / `go vet` ; golangci-lint si configuré.
- Pas de logique métier dans `cmd/` ; `internal/` pour le code privé.
- Erreurs retournées, pas de `panic` aux frontières publiques (HTTP, CLI).

## Pour Kiro / implémentation

- Ne pas introduire de dépendances ou services Docker hors de ce que `02-architecture.md` autorise.
- Déploiement / IaC : hors scope sauf ADR dans `05-decisions.md`.
