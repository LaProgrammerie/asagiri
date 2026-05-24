# Template projet — PHP / Node, Docker (Castor), IA, infra Yoimachi

Ce dépôt est une **base composite** pour démarrer un produit web :

| Couche | Rôle | Référence |
|--------|------|-----------|
| **Docker + Castor** | Stack locale (Traefik, PHP, Postgres, builder, tâches CLI) | [Castor](https://castor.jolicode.com/), [docker-starter](https://github.com/jolicode/docker-starter) — détail dans [`README.docker-starter.md`](README.docker-starter.md) |
| **AI Engineering** | Spec → handoff → code, Kiro, Cursor, `docs/ai/` | [ai-engineering-framework](https://github.com/LaProgrammerie/ai-engineering-framework) |
| **Infra / déploiement** | YAML → Terraform/OpenTofu (workspaces, destinations) | [Yoimachi](https://github.com/LaProgrammerie/yoimachi) — voir [`docs/ai/02-architecture.md`](docs/ai/02-architecture.md) et [`infra/yoimachi/`](infra/yoimachi/) |

## Démarrage rapide

1. **Configurer** `castor.php` → `create_default_variables()` (`project_name`, `root_domain`, `php_version`, etc.).
2. **Installer Castor** : [documentation officielle](https://castor.jolicode.com/).
3. **Lancer la stack** : `castor start` (voir aussi [`README.dist.md`](README.dist.md) après `castor init`).
4. **Couche IA** : lire [`AGENTS.md`](AGENTS.md) puis [`docs/ai/context-map.md`](docs/ai/context-map.md). Optionnel : cloner [ai-engineering-core](https://github.com/LaProgrammerie/ai-engineering-core) et exécuter `./sync-to-home.sh` pour les skills globales dans `~/.kiro`.
5. **Infra déployable** : décrire l’app dans un `yoimachi.yaml` (voir [`docs/ai/02-architecture.md`](docs/ai/02-architecture.md)).

## Structure utile

```
├── castor.php                 # Tâches Castor (docker-starter)
├── application/               # Point d’entrée web PHP par défaut
├── infrastructure/docker/     # Compose, Dockerfiles, Traefik
├── docs/ai/                   # Canon projet + spec active (framework IA)
├── .kiro/                     # Specs Kiro, steering, skill create-handoff
├── .cursor/rules/             # Règles Cursor
├── infra/yoimachi/            # Placeholder / exemple config Yoimachi
└── README.docker-starter.md   # Doc amont docker-starter (longue)
```

## Node.js à côté de PHP

La stack par défaut est **PHP** (docker-starter). Le **builder** inclut déjà Node/Yarn pour les assets. Pour une app **Node** principale, adapte `infrastructure/docker/` (service dédié, `Dockerfile`, nginx) et mets à jour [`docs/ai/03-standards.md`](docs/ai/03-standards.md) avec les commandes réelles.

## Tâches prêtes (Castor)

- QA : `castor qa:all`, `castor qa:static-checks`, `castor qa:qa-js`
- Frontend : `castor app:assets:dev`, `castor app:assets:prod`, `castor app:assets:watch`
- Optimisations : `castor optimize:all` (scripts npm si présents)
- Tests frontend optionnels : `castor test:all-optimizations` (scripts npm si présents)
- Déploiement infra : `castor deploy:validate`, `castor deploy:generate`, `castor deploy:deploy` (Yoimachi)

## Initialisation « produit » (docker-starter)

Quand le squelette te convient :

```bash
castor init
```

Cela remplace ce README par le flux documenté dans docker-starter ; **préserve** manuellement les sections « IA » et « Yoimachi » (ou fusionne depuis ce fichier / `README.dist.md`).

## Crédits licences

- **docker-starter** : MIT, [JoliCode](https://jolicode.com/).
- **Fichiers issus de ai-engineering-template** : MIT, [LaProgrammerie](https://github.com/LaProgrammerie).
- **Yoimachi** : voir le dépôt amont (fichiers d’exemple sous `infra/yoimachi/`).
