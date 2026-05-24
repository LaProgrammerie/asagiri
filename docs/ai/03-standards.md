# Standards et conventions (projet)

> **À remplir au bootstrap :** commandes, arborescence et barre de tests sont nécessaires pour des handoffs exécutables.

## Langage et style

- **Langages :** PHP (par défaut, sous `application/`) ; Node disponible dans le builder pour les assets ou une app dédiée si tu l’ajoutes.
- **PHP :** respecter la config du dépôt (PHP-CS-Fixer, PHPStan si activés) — voir `tools/`, `phpstan.neon`, `.php-cs-fixer.php`.
- **Nommage :** *(conventions d’équipe)*

## Arborescence (cible)

```
application/              # App web PHP par défaut
infrastructure/docker/    # Compose, images
castor.php                # Tâches CLI
docs/ai/                  # Canon + spec active
.kiro/specs/              # Specs Kiro
infra/yoimachi/           # Config infra générée / Yoimachi
```

## Commandes (référence rapide)

| Action | Commande |
|--------|----------|
| Démarrer la stack | `castor start` |
| Builder (shell outillé) | `castor builder` |
| Lister les tâches | `castor` ou `castor list` |
| Tests / qualité PHP | *(à définir : ex. `composer test`, `tools/bin/phpstan`)* |
| Yoimachi | `yoimachi validate` / `yoimachi generate` *(selon installation)* |

Adapte ce tableau quand le projet réel fixe Composer scripts et CI.

## Tests

- **Niveau attendu :** unitaires / intégration / e2e *(précise)*
- **Fichiers de tests :** *(emplacement et nommage)*

## Sécurité et données

- Pas de secrets dans le dépôt ; utiliser `.env` local (gitignored) et pipelines CI.
- *(PII, logs, rétention — lien avec `AGENTS.md` si besoin.)*
