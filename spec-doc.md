Spec — Documentation publique AgentFlow

1. Objectif

Mettre en place une documentation publique complète, agréable, maintenable et prête pour l’open source.

La documentation doit être traitée comme une partie centrale du produit, pas comme un dossier Markdown annexe.

Objectifs :

* expliquer clairement la proposition de valeur ;
* permettre une prise en main rapide ;
* documenter exhaustivement les commandes ;
* rendre l’architecture compréhensible ;
* expliquer la logique cost/tokens/performance ;
* documenter la fiabilité et les garde-fous ;
* donner confiance aux contributeurs ;
* préparer une publication GitHub publique sous licence Apache 2.0.

Inspiration UX : documentation Kiro CLI, notamment clarté de navigation, lisibilité, structure, exemples terminal et confort de lecture.

⸻

2. Décision technique

2.1 Choix retenu

Utiliser :

Fumadocs + Next.js + MDX + GitHub Pages

Justification :

* rendu moderne ;
* bonne expérience développeur ;
* documentation structurée ;
* navigation agréable ;
* support MDX ;
* excellent rendu code/terminal ;
* dark mode ;
* extensible vers un site produit plus complet ;
* hébergement statique possible.

Fumadocs supporte plusieurs frameworks React, dont Next.js, et documente le déploiement ainsi que le static build. Next.js supporte l’export statique via output: 'export', adapté à GitHub Pages. Références : Fumadocs Deploying, Fumadocs Static Build, Next.js Static Exports, template officiel Next.js GitHub Pages.

2.2 Choix rejetés

Option	Décision	Raison
GitHub Wiki	Rejeté	Faible branding, versioning moins propre, mauvaise scalabilité docs
README seul	Rejeté	Insuffisant pour un outil technique ambitieux
Docusaurus	Acceptable mais non retenu	Plus classique, moins premium visuellement
Vercel	Possible plus tard	GitHub Pages suffit pour l’open source initial
Cloudflare Pages	Possible plus tard	Très bon choix si custom domain/perf devient prioritaire

⸻

3. Contraintes de déploiement

Le site doit être compatible avec GitHub Pages.

Implications :

* build statique obligatoire ;
* pas de SSR runtime ;
* pas de routes API Next.js ;
* pas de dépendance serveur ;
* configuration output: 'export' ;
* gestion correcte du basePath si le site est servi sous /agentflow ;
* images compatibles static export ;
* CI GitHub Actions pour build + deploy.

Exemple cible :

https://laprogrammerie.github.io/agentflow/

ou plus tard :

https://agentflow.dev/

⸻

4. Architecture repository

4.1 Structure recommandée

repo/
  application/
    cmd/agentflow/
    internal/
  docs-site/
    app/
    content/
      docs/
      blog/
    components/
    public/
    package.json
    next.config.mjs
    source.config.ts
    fumadocs.config.ts
  docs/
    architecture/
    decisions/
    specs/
    contributing/
  README.md
  LICENSE
  CONTRIBUTING.md
  CODE_OF_CONDUCT.md
  SECURITY.md
  ROADMAP.md

4.2 Séparation docs/ vs docs-site/

docs-site/ contient le site Fumadocs.

docs/ contient les documents projet plus bruts :

* ADR ;
* specs internes ;
* décisions ;
* documents de conception ;
* notes contributeur avancées.

Les pages publiques importantes doivent être disponibles dans docs-site/content/docs.

⸻

5. Arborescence documentation publique

docs-site/content/docs/
  index.mdx
  getting-started/
    index.mdx
    installation.mdx
    quickstart.mdx
    first-workflow.mdx
    requirements.mdx
  concepts/
    index.mdx
    philosophy.mdx
    deterministic-orchestration.mdx
    local-first.mdx
    worktrees.mdx
    agents.mdx
    runs-tasks-features.mdx
    cost-aware-workflows.mdx
    trust-and-explainability.mdx
  cli/
    index.mdx
    init.mdx
    doctor.mdx
    spec.mdx
    plan.mdx
    enrich.mdx
    dev.mdx
    verify.mdx
    review.mdx
    pr.mdx
    work.mdx
    continue.mdx
    next.mdx
    inbox.mdx
    sync.mdx
    estimate.mdx
    investigate.mdx
    context.mdx
    cost.mdx
    mcp.mdx
  workflows/
    index.mdx
    kiro-to-cursor.mdx
    notion-to-code.mdx
    local-first-workflow.mdx
    cost-aware-workflow.mdx
    full-feature-workflow.mdx
    failure-recovery.mdx
    ci-workflow.mdx
  configuration/
    index.mdx
    config-file.mdx
    agents.mdx
    models.mdx
    pricing.mdx
    budgets.mdx
    routing.mdx
    sources.mdx
    notion.mdx
    validation.mdx
    ui.mdx
  agents/
    index.mdx
    kiro.mdx
    cursor-agent.mdx
    codex.mdx
    claude-code.mdx
    ollama.mdx
    local-models.mdx
    cloud-models.mdx
    provider-interface.mdx
  cost-performance/
    index.mdx
    token-estimation.mdx
    cost-estimation.mdx
    context-optimization.mdx
    local-investigation.mdx
    routing-strategy.mdx
    budgets.mdx
    reports.mdx
    benchmarks.mdx
  reliability/
    index.mdx
    state-machine.mdx
    worktree-isolation.mdx
    validation.mdx
    reviews.mdx
    retries-resume.mdx
    deterministic-prompts.mdx
    confidence-scoring.mdx
    failure-analysis.mdx
  mcp/
    index.mdx
    overview.mdx
    tools.mdx
    security.mdx
    examples.mdx
  security/
    index.mdx
    secrets.mdx
    filesystem-scope.mdx
    sandboxing.mdx
    network-policy.mdx
    sensitive-logs.mdx
  architecture/
    index.mdx
    overview.mdx
    modules.mdx
    execution-pipeline.mdx
    state-storage.mdx
    context-pipeline.mdx
    agent-contracts.mdx
    extension-points.mdx
  contributing/
    index.mdx
    development-setup.mdx
    testing.mdx
    documentation.mdx
    architecture-decisions.mdx
    release-process.mdx
  reference/
    index.mdx
    file-structure.mdx
    task-schema.mdx
    run-schema.mdx
    agent-contract.mdx
    config-schema.mdx
    environment-variables.mdx
    exit-codes.mdx
    glossary.mdx
  roadmap/
    index.mdx
    v1.mdx
    v2.mdx
    v3.mdx

⸻

6. Pages prioritaires MVP docs

Pour la première publication, produire au minimum :

1. Home docs ;
2. Installation ;
3. Quickstart ;
4. First workflow ;
5. Philosophy ;
6. CLI overview ;
7. agentflow work ;
8. Config file ;
9. Agents overview ;
10. Ollama/local models ;
11. Cost & token estimation ;
12. Context optimization ;
13. Worktree isolation ;
14. Architecture overview ;
15. Contributing ;
16. Security ;
17. Roadmap.

⸻

7. Ton éditorial

La documentation doit être :

* directe ;
* technique ;
* fiable ;
* transparente ;
* sans promesses marketing absurdes ;
* orientée usage réel ;
* claire sur les limites.

Éviter :

* “autonomous AI magic” ;
* bullshit marketing ;
* promesses de remplacement développeur ;
* jargon non expliqué ;
* pages trop abstraites ;
* exemples incomplets.

Positionnement éditorial :

AgentFlow helps engineers orchestrate AI coding workflows with reproducible plans, local-first investigation, cost-aware routing, and explicit validation.

⸻

8. Structure type d’une page CLI

Chaque commande doit suivre le même format.

---
title: agentflow work
description: Run an intent-based workflow from a natural instruction.
---
# agentflow work
Short explanation.
## When to use it
Concrete situations.
## Usage
```bash
agentflow work "develop billing-v2"

Options

Option	Description
–estimate-only	Show estimate without running
–plan-only	Build the execution plan only

Examples

Develop a feature

agentflow work "develop billing-v2"

Resume interrupted work

agentflow work "resume billing-v2"

What happens internally

1. resolve intent
2. inspect state
3. optimize context
4. estimate cost
5. execute plan

Output

Example terminal output.

Failure modes

Common errors and fixes.

Related commands

* agentflow continue
* agentflow estimate
* agentflow status

---
## 9. Structure type d’une page concept
```mdx
---
title: Local-first workflows
description: Why AgentFlow does local investigation before calling models.
---
# Local-first workflows
## Problem
Explain the real engineering problem.
## AgentFlow approach
Explain the solution.
## Example
Before / after.
## Trade-offs
What this improves and what it does not solve.
## Configuration
Relevant config.
## Related pages

⸻

10. Home docs

La page d’accueil docs doit expliquer en moins de 30 secondes :

* ce qu’est AgentFlow ;
* à qui ça sert ;
* le workflow principal ;
* pourquoi c’est différent ;
* comment démarrer.

Contenu attendu :

# AgentFlow
Deterministic orchestration for AI coding workflows.
AgentFlow turns specs into auditable, cost-aware development runs using local investigation, git worktrees, AI coding agents, validation commands, and reproducible reports.
```bash
agentflow work "develop billing-v2"

Core ideas

* Local-first investigation
* Cost and token estimation
* Git worktree isolation
* Agent routing
* Validation before trust
* Reproducible reports

---
## 11. Guides obligatoires
### 11.1 First workflow
Montrer un workflow complet :
```bash
agentflow init
agentflow doctor
agentflow work "develop billing-v2" --estimate-only
agentflow work "develop billing-v2"
agentflow status
agentflow report <run-id>

11.2 Kiro → Cursor → Verify

Montrer :

agentflow spec billing-v2 --agent kiro
agentflow plan billing-v2
agentflow enrich billing-v2 --agent ollama
agentflow dev billing-v2 --agent cursor
agentflow verify billing-v2
agentflow review billing-v2 --agent codex

11.3 Notion → local spec → work

Montrer :

agentflow sync notion --page <url>
agentflow inbox
agentflow work "develop billing-v2"

Préciser que Notion n’est jamais la source d’exécution directe : snapshot local obligatoire.

11.4 Cost-aware workflow

Montrer :

agentflow estimate billing-v2 --task task-003
agentflow context billing-v2 --task task-003 --optimize
agentflow work "develop billing-v2" --budget 0.50

⸻

12. Design system documentation

12.1 Style visuel

Le site doit être :

* dark mode first ;
* clair en light mode ;
* sobre ;
* très lisible ;
* proche d’une documentation développeur premium ;
* sans surcharge marketing.

12.2 Composants docs souhaités

* Terminal blocks ;
* callouts ;
* cards ;
* stepper ;
* tabs ;
* badges ;
* architecture diagrams ;
* command reference tables ;
* config examples ;
* warning blocks ;
* comparison tables.

12.3 Conventions callouts

Types :

Note
Warning
Cost
Security
Local-first
Experimental

Exemple :

<Callout type="warning">
AgentFlow does not prove correctness. Validation and human review remain mandatory for critical changes.
</Callout>

⸻

13. Diagrammes

Utiliser Mermaid ou équivalent si compatible avec le stack choisi.

Diagrammes minimaux :

1. execution pipeline ;
2. local-first context pipeline ;
3. agent routing ;
4. state machine ;
5. worktree isolation ;
6. cost estimation flow ;
7. MCP interaction.

Exemple :

flowchart TD
  A[User intent] --> B[Intent Resolver]
  B --> C[High-level Planner]
  C --> D[Local Investigation]
  D --> E[Context Optimization]
  E --> F[Cost Estimate]
  F --> G[Agent Execution]
  G --> H[Validation]
  H --> I[Review]
  I --> J[Report]

⸻

14. Documentation générée automatiquement

14.1 CLI reference

Prévoir une génération automatique ou semi-automatique depuis Cobra.

Objectif : éviter la divergence entre CLI et docs.

Commande cible :

agentflow docs generate-cli --output docs-site/content/docs/cli/generated

ou script interne :

go run ./application/cmd/agentflow docs generate-cli

14.2 Config schema

Prévoir une génération depuis structures Go ou JSON Schema.

Sorties :

docs-site/content/docs/reference/config-schema.mdx
docs-site/public/schemas/config.schema.json

14.3 Golden tests docs

Ajouter des tests pour vérifier que :

* toutes les commandes documentées existent ;
* les options documentées correspondent à Cobra ;
* les exemples critiques sont syntaxiquement valides ;
* les fichiers de docs buildent.

⸻

15. CI/CD documentation

15.1 Scripts npm attendus

Dans docs-site/package.json :

{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "lint": "next lint",
    "typecheck": "tsc --noEmit",
    "docs:check": "next build"
  }
}

15.2 GitHub Action build docs

Créer :

.github/workflows/docs.yml

Responsabilités :

* checkout ;
* setup Node ;
* install dependencies ;
* typecheck ;
* build static ;
* upload artifact ;
* deploy GitHub Pages sur main.

15.3 Static export

Configurer Next.js :

const nextConfig = {
  output: 'export',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
}
export default nextConfig

Si déploiement sous sous-chemin GitHub Pages :

const repo = 'agentflow'
const isGithubPages = process.env.GITHUB_PAGES === 'true'
const nextConfig = {
  output: 'export',
  basePath: isGithubPages ? `/${repo}` : '',
  assetPrefix: isGithubPages ? `/${repo}/` : '',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
}

⸻

16. README principal

Le README doit rester court et orienté conversion.

Structure :

# AgentFlow
Deterministic orchestration for AI coding workflows.
## Why
## Install
## Quickstart
## Core workflow
## Documentation
## Status
## Contributing
## License

Le README doit pointer vers la documentation complète.

⸻

17. Contenu “trust” obligatoire

Créer des pages explicites :

17.1 Limits

Expliquer que AgentFlow :

* ne garantit pas la correction ;
* ne remplace pas les tests ;
* ne remplace pas la review humaine ;
* ne protège pas contre tous les mauvais prompts ;
* réduit les risques par orchestration, isolation et validation.

17.2 Security model

Expliquer :

* scope fichiers ;
* secrets ;
* MCP ;
* logs ;
* agents externes ;
* cloud models ;
* données envoyées aux providers.

17.3 Cost model

Expliquer :

* estimation approximative ;
* tokens estimés vs réels ;
* pricing configurable ;
* budgets ;
* limites.

⸻

18. Documentation des limites expérimentales

Toute feature non stabilisée doit être marquée :

Experimental

Exemples :

* MCP local ;
* cloud model routing ;
* confidence scoring ;
* automatic context compression ;
* Notion sync ;
* benchmark comparisons.

⸻

19. Critères d’acceptation

La documentation est acceptable si :

* le site docs build en static export ;
* GitHub Pages déploie automatiquement ;
* la navigation principale est complète ;
* le quickstart permet d’utiliser AgentFlow en moins de 10 minutes ;
* toutes les commandes principales sont documentées ;
* la philosophie local-first/cost-aware est claire ;
* la sécurité et les limites sont documentées ;
* les exemples CLI sont cohérents avec la CLI réelle ;
* la doc contient au moins un workflow complet Kiro → Cursor → Verify ;
* la doc contient au moins un workflow cost-aware ;
* la doc contient au moins un diagramme d’architecture ;
* le README pointe vers la doc ;
* aucune information sensible n’est présente ;
* les pages expérimentales sont clairement marquées.

⸻

20. Découpage d’implémentation recommandé

Phase 1 — Setup docs-site

* initialiser Fumadocs + Next.js ;
* config static export ;
* config GitHub Pages ;
* theme minimal ;
* home docs ;
* navigation.

Phase 2 — Docs MVP

Créer les pages prioritaires :

* installation ;
* quickstart ;
* first workflow ;
* philosophy ;
* CLI overview ;
* work ;
* config ;
* architecture overview ;
* cost/tokens ;
* security ;
* contributing.

Phase 3 — CLI reference complète

* documenter toutes les commandes ;
* ajouter exemples ;
* ajouter failure modes ;
* générer ou vérifier la référence depuis Cobra.

Phase 4 — Workflows avancés

* Kiro → Cursor ;
* Notion → AgentFlow ;
* Ollama local ;
* cost-aware ;
* failure recovery ;
* CI.

Phase 5 — Trust & OSS readiness

* limits ;
* security model ;
* roadmap ;
* contribution guide ;
* release process ;
* benchmarks.

Phase 6 — Polish UX

* callouts ;
* terminal blocks ;
* diagrams ;
* cards ;
* dark/light QA ;
* mobile QA ;
* search QA.

⸻

21. Mission à donner aux agents

Mission :

Implement a complete Fumadocs-based documentation site for AgentFlow, deployable as a static site on GitHub Pages, with a documentation structure suitable for an open-source AI coding workflow orchestrator focused on deterministic execution, local-first investigation, cost/token optimization, and reliability.

Contraintes :

* do not remove existing CLI commands ;
* do not invent unsupported features as stable ;
* mark future or incomplete capabilities as experimental ;
* keep examples aligned with actual CLI output ;
* prefer generated docs where possible ;
* build must pass locally and in CI ;
* docs must be readable without marketing hype ;
* GitHub Pages must be functional.

⸻

22. Résumé exécutable

Choix :

Fumadocs + Next.js + MDX + GitHub Pages

But :

Créer une documentation produit complète, belle, fiable, versionnée et prête open source.

Principe clé :

La documentation doit démontrer la fiabilité d’AgentFlow autant qu’elle l’explique.