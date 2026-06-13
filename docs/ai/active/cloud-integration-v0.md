# Cloud integration V0 — CLI Asagiri

> **Statut :** implémenté (login / logout / status / link / push)  
> **Cloud API :** `asagiri-cloud` Symfony `/api/v1`  
> **ADR :** [039-team-cloud](../../decisions/039-team-cloud.md)

## Décision — mapping repo local → projet cloud

### Options comparées

| Option | Description | Avantages | Inconvénients |
|--------|-------------|-----------|---------------|
| **A** | `cloud.project_id` dans `.asagiri/config.yaml` | Explicite, stable, versionnable par équipe (hors token), pas de magie | Saisie manuelle de l'UUID |
| **B** | Commande `asa cloud link` | Même persistance que A, validation API, DX onboarding | Une commande de plus |
| **C** | Détection auto (git remote, slug, etc.) | Zéro config si ça marche | **Impossible en V0** : entité `Project` cloud sans `git_remote` ; URLs SSH/HTTPS ambiguës ; forks |

### Choix V0

**A + B combinés** :

1. **Source de vérité** : `cloud.project_id` dans `.asagiri/config.yaml` (par dépôt).
2. **UX** : `asa cloud link <project-uuid>` ou `asa cloud link --slug <slug>` écrit ce champ après validation `GET /api/v1/projects`.
3. **Pas d'auto-détection** en V0 — évolution possible quand le cloud expose `git_remote` + endpoint de résolution.

Le **token** reste **global** (`~/.config/asagiri/cloud/token` par défaut), pas dans le dépôt.

## Configuration

```yaml
cloud:
  enabled: false                    # défaut — CLI 100 % local
  base_url: http://localhost        # dev : https://asagiri-cloud.test
  token_path: ~/.config/asagiri/cloud/token
  project_id: ""                    # requis pour push — via link ou manuel
```

## Commandes

```bash
asa cloud status [--json]
asa cloud login --token <token> [--base-url <url>]
asa cloud logout [--json]
asa cloud link <project-uuid> [--slug <slug>] [--enable] [--json]
asa cloud push [--dry-run] [--run <id>|--all] [--json]
```

## Contrat API

| Méthode | Route | CLI |
|---------|-------|-----|
| GET | `/api/v1/me` | `status` |
| GET | `/api/v1/projects` | `link` |
| POST | `/api/v1/runs` | `push` |
| POST | `/api/v1/ledger-entries` | `push` |

Détail : `asagiri-cloud/docs/api-contract.md`.

## Blocants V0 côté cloud

- `project` IRI obligatoire sur POST runs
- Dedup `(project_id, local_run_id)` — re-push non idempotent
- Pas de `git_remote` sur `Project` — auto-link reporté
