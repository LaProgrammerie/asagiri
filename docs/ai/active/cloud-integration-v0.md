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
asa cloud push [--dry-run] [--execute] [--run <id>|--all] [--json]
```

## Payload cloud (idempotence)

Le CLI envoie des identifiants métier **explicites** — le cloud ne doit plus déduire `taskId` depuis `rawPayload`.

### POST `/api/v1/runs`

```json
{
  "project": "/api/v1/projects/<uuid>",
  "localRunId": "run-2026-06-15-001",
  "taskId": "chatbot-001",
  "feature": "chatbot",
  "status": "completed",
  "startedAt": "2026-06-15T10:00:00Z",
  "endedAt": "2026-06-15T10:05:00Z",
  "durationMs": 300000
}
```

- **`taskId`** : premier `task_id` non vide parmi les entrées ledger du run.
- **`localRunId`** : clé d'idempotence run `(project, localRunId)`.

### POST `/api/v1/ledger-entries`

```json
{
  "run": "/api/v1/runs/<cloud-run-id>",
  "localEntryId": "a1b2c3…",
  "agentId": "dev",
  "phase": "dev",
  "startedAt": "2026-06-15T10:00:00Z",
  "endedAt": "2026-06-15T10:05:00Z",
  "durationMs": 300000,
  "exitCode": 0,
  "promptHash": "…",
  "outputHash": "…",
  "provider": "anthropic",
  "model": "claude-sonnet-4",
  "inputTokens": 12400,
  "outputTokens": 3200,
  "costUsd": 0.042,
  "rawPayload": { "task_id": "chatbot-001", "run_id": "…" }
}
```

- **`localEntryId`** : identifiant stable dérivé localement (voir ci-dessous).
- **`provider`**, **`model`**, **`inputTokens`**, **`outputTokens`**, **`costUsd`** : métadonnées LLM optionnelles (absentes si non renseignées localement).
- **`rawPayload`** : copie JSON de la ligne ledger (rétrocompat) ; sérialisation stable.

### Stratégie `localEntryId`

Pas de champ supplémentaire dans `ledger.jsonl`.

```
localEntryId = SHA256_hex(
  run_id + SEP + agent_id + SEP + phase + SEP +
  started_at + SEP + prompt_hash + SEP + output_hash
)
```

SEP = `\x1e`. Même ligne ledger → même `localEntryId` à chaque push. Re-pousser trois fois produit un payload JSON **identique**.

Implémentation : `agentledger.LocalEntryID`, `cloud.BuildRunCreateRequest`, `cloud.BuildLedgerCreateRequest`.

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
- Dedup `(project_id, local_run_id)` et `(run_id, local_entry_id)` — re-push idempotent avec `taskId` / `localEntryId`
- Pas de `git_remote` sur `Project` — auto-link reporté
