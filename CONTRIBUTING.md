# Contribuer à AgentFlow

Merci de votre intérêt pour AgentFlow (dépôt `hyper-fast-builder`).

## Avant de coder

1. Lire [`AGENTS.md`](AGENTS.md) et [`docs/ai/context-map.md`](docs/ai/context-map.md).
2. Pour une feature : aligner avec [`spec-postv123.md`](spec-postv123.md) ou la spec active (`docs/ai/active/handoff.md`).
3. Ne pas committer de secrets (`.env`, tokens Notion, clés API).

## Environnement

```bash
go mod download
make build
make test
```

Tests avec détecteur de courses : `go test -race ./...`

## Conventions

- Code Go sous `application/internal/` ; point d’entrée `application/cmd/agentflow`.
- Erreurs explicites ; pas de `panic` aux frontières CLI.
- Agents externes : toujours subprocess sans shell (`exec.Command`).
- Nouvelle décision d’architecture → entrée dans `docs/ai/05-decisions.md`.

## Pull requests

- Une PR = un sujet cohérent (feature, fix, docs).
- Inclure ou mettre à jour les tests pour le comportement modifié.
- Vérifier `go test -race ./...` avant ouverture.
- Mettre à jour `docs/ai/active/handoff.md` / `current-spec.md` si le périmètre change.

## Rapports de bugs

Inclure : version (`agentflow --version`), OS, commande exacte, sortie (secrets masqués), config redacted (`.agentflow/config.yaml` sans tokens).
