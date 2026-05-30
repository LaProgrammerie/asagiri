# Walkthrough — Project Onboarding

Parcours minimal pour préparer un dépôt forké avant le premier `asa work`.

## Prérequis

- Git initialisé (`git init` si dépôt neuf)
- Binaire `asa` compilé (`make build` à la racine du repo Asagiri)

## Étapes

```bash
cd mon-projet
asa init
asa onboard --yes --non-interactive
asa ready --json
asa doctor --full
```

## Stack PHP Castor

Avec un `castor.php` à la racine :

```bash
asa onboard --yes --stack php --non-interactive
asa ready --json | jq '.checks[] | select(.id | startswith("validation"))'
```

Les commandes `castor qa:static-checks` et `castor qa:phpunit` sont proposées dans `validation.commands`.

## Reprise et dry-run

```bash
asa onboard --dry-run          # preview sans écriture
asa onboard --resume           # reprend .asagiri/onboarding/state.json
asa onboard --check-only       # alias readiness (identique à asa ready)
```

## TUI

```bash
asa onboard --ui               # wizard Mission Control
asa mission                    # bannière readiness si score < 100 %
```

## CI

```bash
asa ready --json --strict
# exit 0 si ready: true ; exit 1 sinon
```
