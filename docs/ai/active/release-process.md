# Release process — Asagiri (M1.1)

Procédure opérateur pour publier une version du binaire `asa` sur GitHub Releases et le tap Homebrew générique La Programmerie. Aucun impact sur le moteur (workflow, trust, agentledger, knowledge, runtime).

## Prérequis

| Élément | Détail |
|--------|--------|
| **Dépôt** | `LaProgrammerie/asagiri` |
| **Tap Homebrew** | `LaProgrammerie/homebrew-tap` (`brew tap LaProgrammerie/tap`) |
| **Secret CI** | `HOMEBREW_TAP_GITHUB_TOKEN` — PAT avec `contents:read` + `contents:write` sur `homebrew-tap` |
| **Outils locaux** | Go 1.25+, `goreleaser` (optionnel : `make release-snapshot`) |

Vérifier avant release :

```bash
make release-check
go test ./... -count=1
```

## Artefacts produits (automatique)

GoReleaser (`.goreleaser.yaml`) sur tag `v*` :

| Plateforme | Modèle de nom (tag `vX.Y.Z` → `VER=X.Y.Z`) |
|------------|---------|
| macOS arm64 | `asa_${VER}_darwin_arm64.tar.gz` |
| macOS amd64 | `asa_${VER}_darwin_amd64.tar.gz` |
| Linux amd64 | `asa_${VER}_linux_amd64.tar.gz` |
| Linux arm64 | `asa_${VER}_linux_arm64.tar.gz` |
| Windows | `asa_${VER}_windows_<arch>.zip` (hors `install.sh`) |

- **Checksums** : `checksums.txt` (SHA256)
- **Homebrew** : formule `asagiri` → binaire installé **`asagiri`** (`bin.install "asa" => "asagiri"`) dans `Formula/asagiri.rb` sur `homebrew-tap`
- **Version injectée** : ldflags `version.Version`, `version.Commit`, `version.Date`

## Créer une release

1. **Préparer le changelog** — mettre à jour `CHANGELOG.md`.
2. **Valider la branche** — CI verte ; `make release-check` ; tests complets.
3. **Tag annoté** (semver) :

   ```bash
   git tag -a v1.0.0 -m "v1.0.0"
   git push origin v1.0.0
   ```

4. **Workflow `Release`** (`.github/workflows/release.yml`) :
   - `go test ./...`
   - `goreleaser release --clean`
   - Publie les assets sur GitHub Releases
   - Commit `Formula/asagiri.rb` sur `LaProgrammerie/homebrew-tap`

5. **Vérifier** :

   ```bash
   gh release view v1.0.0 --repo LaProgrammerie/asagiri
   brew tap LaProgrammerie/tap
   brew install asagiri
   asagiri version
   ```

## Installation utilisateur

| Canal | Commande |
|-------|----------|
| Homebrew | `brew tap LaProgrammerie/tap && brew install asagiri` |
| Script | `curl -fsSL https://raw.githubusercontent.com/LaProgrammerie/asagiri/main/scripts/install.sh \| bash` |
| Archive manuelle | Voir `docs/ai/active/distribution-oss.md` |

Variable utile : `ASAGIRI_VERSION=v1.0.0` pour épingler une version dans `install.sh`.

## Publication Homebrew

- **Dépôt** : `github.com/LaProgrammerie/homebrew-tap` (tap générique, plusieurs formules futures)
- **Commande tap** : `brew tap LaProgrammerie/tap`
- **Formule Asagiri** : `asagiri` (fichier `Formula/asagiri.rb`, binaire `asa`)
- **Token** : secret `HOMEBREW_TAP_GITHUB_TOKEN` sur le dépôt `asagiri` (Settings → Secrets → Actions)

Le PAT doit avoir accès **write** au dépôt `LaProgrammerie/homebrew-tap` (fine-grained : Contents read/write ; classic : `repo` scope sur ce dépôt).

Si le push brew échoue en **403** : vérifier que le token cible `homebrew-tap` (pas un ancien dépôt `homebrew-asagiri`) et que le PAT n’a pas expiré.

### Formule legacy `asa`

L’ancienne formule `asa` sur le même tap peut coexister pendant la transition. Les nouvelles installs utilisent `brew install asagiri`.

## Relancer une release après échec Homebrew (ex. `v0.1.0-alpha.1`)

Cas typique : assets GitHub OK, push formule en 403 (mauvais dépôt tap ou token).

1. **Merger** le correctif `.goreleaser.yaml` (`repository.name: homebrew-tap`).
2. **Vérifier** le secret `HOMEBREW_TAP_GITHUB_TOKEN` (write sur `homebrew-tap`).
3. **Option A — Re-déclencher le workflow** (sans retagger) :
   - GitHub → Actions → Release → Run workflow n’existe pas sur tag ; utiliser **Re-run failed jobs** sur le run du tag si assets déjà publiés et seul le push brew a échoué.
4. **Option B — Retag propre** (si re-run insuffisant) :

   ```bash
   # Supprimer le tag distant et local (accord équipe)
   git push origin :refs/tags/v0.1.0-alpha.1
   git tag -d v0.1.0-alpha.1
   # Optionnel : supprimer la release GitHub si assets à republier
   gh release delete v0.1.0-alpha.1 --repo LaProgrammerie/asagiri --yes
   # Recréer sur le commit validé
   git tag -a v0.1.0-alpha.1 -m "v0.1.0-alpha.1"
   git push origin v0.1.0-alpha.1
   ```

5. **Contrôle** : `Formula/asagiri.rb` sur `homebrew-tap` pointe vers les archives `v0.1.0-alpha.1` ; `brew install asagiri` OK.

## Rollback

1. **GitHub Release** : pre-release ou suppression assets via l’UI si critique.
2. **Tag** : préférer un patch (`v0.1.0-alpha.2`) plutôt que réécrire un tag public.
3. **Homebrew** : commit manuel sur `homebrew-tap` (URL + sha256 version précédente) ou attendre le prochain tag GoReleaser.
4. **Communication** : note release + `CHANGELOG.md`.

## Snapshot local (sans publier)

```bash
make release-check
make release-snapshot
ls dist/homebrew/Formula/
```

## Hors scope

- Cosign / deb/rpm / cloud / licence / packs premium
