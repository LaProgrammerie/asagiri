# Runbook — renommage dépôt GitHub (action manuelle)

> **État code (fait)** : module `github.com/LaProgrammerie/asagiri`, URLs docs/release pointent vers `LaProgrammerie/asagiri`.  
> **Reste à faire sur GitHub** : renommer le dépôt `hyper-fast-builder` → `asagiri` (GitHub redirige les anciennes URLs).

## Objectif

Aligner le nom du dépôt GitHub sur le module Go et la doc déjà migrés.

## Checklist

1. **Annoncer** la fenêtre (releases, Homebrew, clones).
2. GitHub : **Settings → General → Repository name** → `asagiri`.
3. Localement :
   ```bash
   git remote set-url origin git@github.com:LaProgrammerie/asagiri.git
   ```
4. Vérifier CI, GoReleaser release, Homebrew tap (URLs déjà `LaProgrammerie/asagiri` dans le code).
5. **Releases existantes** sur l’ancien nom restent accessibles via redirection GitHub.

## Rollback

Renommer le repo vers `hyper-fast-builder` si besoin ; revert `go.mod` uniquement si la migration module a été poussée.

## Références

- ADR-016 / ADR-017 dans `docs/ai/05-decisions.md`
- `spec-rename.md`
