# Quickstart Asagiri (dry-run)

Ce dossier documente un parcours **sans agents cloud** pour valider l’installation.

## Prérequis

- Go 1.25+, git, make
- Depuis la racine du dépôt cloné

## Étapes

```bash
make build
export ASA_DRY_RUN=1
./bin/asa init
./bin/asa doctor
./bin/asa work "développe asa-test" --dry-run --plan-only --yes
./bin/asa estimate asa-test
```

## Config minimale

Copier `.asagiri/config.yaml.example` vers `.asagiri/config.yaml` à la racine du projet cible.

`mcp.enabled` reste `false` par défaut (sécurité).

## Benchmark

```bash
make benchmark
# ou
FEATURE=mon-feature ./scripts/benchmark-workflow.sh
```
