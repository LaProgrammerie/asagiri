# Publication npm

Package : `@laprogrammerie/asagiri` (source : `sdk/typescript/`).

## Validation locale (sans publish)

```bash
cd sdk/typescript
./scripts/publish-dry-run.sh
```

Équivalent manuel : `npm ci && npm run build && npm test && npm pack --dry-run`.

## Publication réelle

Réservée à la CI (tag `sdk-v*`) ou à un mainteneur avec token :

```bash
cd sdk/typescript
npm ci
npm run build
npm test
npm publish --access public
```

Ne pas publier depuis une branche de dev non taguée sans validation explicite.

## Secret `NPM_TOKEN` (GitHub)

1. Créer un **Access Token** npm (type **Automation** recommandé pour CI) avec droits **publish** sur le scope `@laprogrammerie`.
2. Dans le dépôt GitHub : **Settings → Secrets and variables → Actions → New repository secret**.
3. Nom : `NPM_TOKEN`, valeur : le token npm (sans préfixe `npm_` supplémentaire dans le secret — coller le token tel quel).
4. Le workflow `.github/workflows/sdk-npm-publish.yml` injecte ce secret via `NODE_AUTH_TOKEN` après `actions/setup-node` avec `registry-url: https://registry.npmjs.org`.

Déclenchement CI :

- Push d’un tag `sdk-v*` (ex. `sdk-v0.1.1`)
- Ou **workflow_dispatch** manuel (même secret requis)

## Versioning

- Incrémenter `version` dans `package.json` avant le tag.
- Tag Git `sdk-v<semver>` aligné sur la version npm (ex. package `0.1.1` → tag `sdk-v0.1.1`).
- Semver npm **indépendant** du binaire Go `asa`.

## Prérequis compte

Compte npm membre de l’organisation / scope `@laprogrammerie` avec droit de publication sur le package.
