# Hooks Cursor (projet)

## Fichier actif

- **`../hooks.json`** (a la racine `.cursor/hooks.json`) enregistre les hooks versionnes pour ce depot.

## Template Generic project

| Hook | Evenement | Script |
|------|-----------|--------|
| Debut de session | `sessionStart` | `template-sync-session-start.sh` |
| Apres ecriture agent | `postToolUse` (`Write`) | `template-sync-post-write.sh` |

**Configuration :** copier `.cursor/template-sync.env.example` vers `.cursor/template-sync.env` et definir `GENERIC_TEMPLATE_ROOT` (chemin absolu du clone **Generic project**). Sans ce fichier, le hook session reste silencieux ; le hook post-ecriture continue de rappeler le port pour les chemins "template".

## Reference format

Voir aussi la documentation officielle Cursor **Hooks** et `hooks.json.example` (variante illustrative).
