Voici une spec prête à donner à tes agents.

> **Implémenté (2026-05-17)** : workflows `docs-cloudflare-pages.yml` + `docs-check.yml` ; `docs.yml` (GitHub Pages) supprimé ; `docs-site/` migré **pnpm** (`pnpm-lock.yaml`). Le job CI exécute toujours `go run … docs generate-cli` avant le build Node (non montré dans le YAML minimal ci-dessous).

# Spec — Déploiement Fumadocs sur Cloudflare Pages via GitHub Actions
## 1. Objectif
Mettre en place le déploiement automatique de la documentation `docs-site/` sur Cloudflare Pages via GitHub Actions.
Le workflow doit :
- builder la documentation Fumadocs/Next.js ;
- produire un export statique ;
- déployer sur Cloudflare Pages ;
- ne se déclencher que si `docs-site/**` ou le workflow docs change ;
- créer des previews sur Pull Requests ;
- déployer en production sur `main` ;
- rester compatible avec une documentation open source.
Cloudflare recommande le déploiement Pages par upload direct avec Wrangler pour les builds custom en CI.  [oai_citation:0‡Cloudflare Docs](https://developers.cloudflare.com/pages/how-to/use-direct-upload-with-continuous-integration/?utm_source=chatgpt.com)
---
## 2. Contraintes techniques
Le site est une app Next.js/Fumadocs située dans :
```txt
docs-site/

Le build doit produire :

docs-site/out/

Next.js doit être configuré en export statique :

// docs-site/next.config.mjs
const nextConfig = {
  output: 'export',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
}
export default nextConfig

Cloudflare documente le déploiement d’un site Next.js statique via Pages.  ￼

⸻

3. Secrets GitHub requis

Créer les secrets GitHub suivants :

CLOUDFLARE_API_TOKEN
CLOUDFLARE_ACCOUNT_ID
CLOUDFLARE_PAGES_PROJECT

Le token Cloudflare doit avoir les droits minimum nécessaires pour déployer sur Cloudflare Pages.

Cloudflare/Wrangler recommande d’utiliser les secrets GitHub pour l’authentification CI.  ￼

⸻

4. Workflow GitHub Actions

Créer :

.github/workflows/docs-cloudflare-pages.yml

Contenu cible :

name: Deploy documentation
on:
  push:
    branches:
      - main
    paths:
      - "docs-site/**"
      - ".github/workflows/docs-cloudflare-pages.yml"
  pull_request:
    paths:
      - "docs-site/**"
      - ".github/workflows/docs-cloudflare-pages.yml"
  workflow_dispatch:
concurrency:
  group: docs-cloudflare-pages-${{ github.ref }}
  cancel-in-progress: true
jobs:
  build-and-deploy-docs:
    name: Build and deploy docs
    runs-on: ubuntu-latest
    timeout-minutes: 15
    defaults:
      run:
        working-directory: docs-site
    permissions:
      contents: read
      deployments: write
      pull-requests: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Enable Corepack
        run: corepack enable
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: docs-site/pnpm-lock.yaml
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
      - name: Typecheck
        run: pnpm typecheck
      - name: Lint
        run: pnpm lint
      - name: Build static docs
        run: pnpm build
      - name: Verify static output
        run: test -d out
      - name: Deploy preview to Cloudflare Pages
        if: github.event_name == 'pull_request'
        uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          workingDirectory: docs-site
          command: pages deploy out --project-name=${{ secrets.CLOUDFLARE_PAGES_PROJECT }} --branch=${{ github.head_ref }}
      - name: Deploy production to Cloudflare Pages
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          workingDirectory: docs-site
          command: pages deploy out --project-name=${{ secrets.CLOUDFLARE_PAGES_PROJECT }} --branch=main

⸻

5. Scripts attendus dans docs-site/package.json

Vérifier ou ajouter :

{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "typecheck": "tsc --noEmit",
    "lint": "next lint"
  }
}

Si next lint n’est pas disponible selon la version Next utilisée, remplacer par ESLint explicite :

{
  "lint": "eslint ."
}

⸻

6. Critères d’acceptation

Le travail est terminé si :

* le workflow se déclenche uniquement sur modifications docs-site/** ou du workflow ;
* pnpm install --frozen-lockfile fonctionne ;
* pnpm typecheck passe ;
* pnpm lint passe ;
* pnpm build génère docs-site/out/ ;
* une PR crée un déploiement preview Cloudflare Pages ;
* un push sur main déploie en production ;
* aucun secret n’est exposé dans les logs ;
* le workflow est relançable manuellement via workflow_dispatch.

⸻

7. Points de vigilance

* Ne pas utiliser SSR, API routes ou middleware Next.js.
* Ne pas dépendre de variables runtime serveur.
* Les images doivent être compatibles static export.
* Ne pas hardcoder le nom du projet Cloudflare si on peut passer par secret.
* Garder Cloudflare Pages comme cible de déploiement, pas Workers.
* Le build doit rester reproductible localement :

cd docs-site
pnpm install
pnpm typecheck
pnpm lint
pnpm build
test -d out
Option robuste : ajoute aussi un workflow séparé `docs-check.yml` sur toutes les PR pour valider la doc même si tu ne veux pas déployer de preview.