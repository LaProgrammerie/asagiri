# Architecture

> **À compléter au bootstrap.** Mets à jour quand une ADR change la structure.

## Vue d’ensemble

- **Runtime local :** Docker Compose sous `infrastructure/docker/`, orchestré par les tâches **Castor** (`castor.php`).
- **Application web par défaut :** répertoire `application/` (PHP), servi par la stack documentée dans [README.docker-starter.md](../../README.docker-starter.md).
- **Tooling conteneurisé :** service `builder` pour Composer, Node, Yarn/npm selon docker-starter.
- **Infra déployable :** pilotée par **Yoimachi** (YAML → Terraform/OpenTofu) avec sources sous `infra/yoimachi/`.

## Composants runtime local (docker-starter)

- **`router` (Traefik) :** exposition locale des entrées HTTP/HTTPS sur `80` et `443` (dashboard `8080`) via `infrastructure/docker/docker-compose.dev.yml`.
- **`frontend` (PHP app) :** service applicatif principal, branché au routeur via labels Traefik et monté sur le code du dépôt.
- **`postgres` :** base locale par défaut (`postgres:16`) avec volume persistant `postgres-data`.
- **`builder` :** conteneur outillé (Composer + Node + Yarn/npm) pour installer les dépendances et lancer les commandes de build/qualité.
- **`worker_*` :** pattern prévu mais non activé par défaut ; activation à cadrer dans une ADR si un traitement asynchrone devient nécessaire.

## Contrat Castor (orchestration locale)

- **Point d’entrée standard :** `castor start` construit les images, installe l’app, monte la stack et lance la migration.
- **Install applicative :** `castor app:install`/`castor install` exécute Composer puis Yarn/npm selon les fichiers présents dans `application/`.
- **Shell outillé :** `castor builder` pour les opérations de dev dans l’environnement conteneurisé.
- **Règle équipe :** éviter les commandes Docker Compose ad hoc dans la doc projet ; documenter les workflows via Castor en priorité.

## Contrat Yoimachi (déploiement)

- **Source de vérité déploiement :** `infra/yoimachi/` (fichiers YAML décrivant la cible).
- **Pipeline logique :** `yoimachi.yaml -> parser -> catalog -> resolver -> planner -> generator -> terraform/tofu`.
- **Chaîne de génération :** Yoimachi produit des artefacts Terraform/OpenTofu à partir du YAML ; les sorties générées ne remplacent pas les sources YAML.
- **Référentiel d’exemples :** s’aligner sur les exemples officiels Yoimachi (`examples/`) plutôt que figer des conventions locales divergentes.
- **Frontière locale vs déploiement :** Docker Starter décrit l’exécution locale ; Yoimachi décrit l’infra cible (staging/prod ou environnements alternatifs).
- **Workflow nominal :** valider en local (`castor start`) -> décrire la cible dans `infra/yoimachi/` -> exécuter `yoimachi validate` puis `yoimachi generate`.

## Limites

- **Entrées / sorties :** trafic HTTP(S) entrant via `router` vers `frontend`; accès base via `postgres`; exécution outillée via `builder`.
- **Dépendances externes :** Postgres local par défaut ; cache/broker/workers hors scope tant qu’aucune ADR ne les introduit.
- **Interdit sans décision :** nouveau service d’infra durable (DB managée, broker, moteur de recherche, etc.) sans entrée `05-decisions.md`.

## Flux critiques

- **Bootstrap local :** `castor start` doit rester le chemin nominal pour démarrer un poste neuf.
- **Cohérence dépendances :** installation PHP/Node doit rester encapsulée par les tâches Castor pour limiter les écarts d’environnement.
- **Parité d’intention infra :** les capacités déclarées dans Yoimachi doivent rester compatibles avec les besoins identifiés côté runtime local.

## Node.js

Si l’app principale est **Node** : documente ici le service Docker, le root du code et comment il interagit avec le reverse-proxy — et mets `03-standards.md` à jour.

## Extension

- **Nouveaux services locaux :** extension sous `infrastructure/docker/` + exposition via Castor.
- **Nouvelles capacités de déploiement :** extension via `infra/yoimachi/` d’abord, puis mise à jour des sections Yoimachi dans `docs/ai/*` si procédure impactée.
- **Décisions durables :** consigner tout changement structurel dans `docs/ai/05-decisions.md`.
