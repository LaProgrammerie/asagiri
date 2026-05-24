# Quickstart AgentFlow (dry-run)

Ce dossier documente un parcours **sans agents cloud** pour valider l’installation.

## Prérequis

- Go 1.25+, git, make
- Depuis la racine du dépôt cloné

## Étapes

```bash
make build
export AGENTFLOW_DRY_RUN=1
./bin/agentflow init
./bin/agentflow doctor
./bin/agentflow work "développe agentflow-test" --dry-run --plan-only --yes
./bin/agentflow estimate agentflow-test
```

## Config minimale

Copier `.agentflow/config.yaml.example` vers `.agentflow/config.yaml` à la racine du projet cible.

`mcp.enabled` reste `false` par défaut (sécurité).

## Benchmark

```bash
make benchmark
# ou
FEATURE=mon-feature ./scripts/benchmark-workflow.sh
```
