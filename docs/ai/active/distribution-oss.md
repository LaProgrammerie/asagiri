# Distribution OSS — Asagiri (`asa`)

> **Phase :** monetization-distribution-v1 **M1**  
> **Statut :** documentation canon (M1.0 + M1.1 automation release)  
> **ADR :** [037-monetization-distribution.md](../../decisions/037-monetization-distribution.md)  
> **Spec :** [`.kiro/specs/monetization-distribution-v1/`](../../../.kiro/specs/monetization-distribution-v1/)

## Positionnement

**Asagiri** est un **local-first AI development control plane** — le « Terraform de l’orchestration IA de dev ».

La distribution OSS vise un développeur solo ou une équipe qui :

- installe `asa` **sans compte** La Programmerie ;
- exécute **100 % en local** (`asa work`, gates, trust, ledger) ;
- n’a **aucun quota** ni dépendance cloud ;
- pourra plus tard ajouter **Pro Local** (packs) ou **Team Cloud** (opt-in) sans changer de binaire.

---

## Stratégie de licence OSS recommandée

| Élément | Décision |
|---------|----------|
| **Licence moteur** | **Apache License 2.0** — fichier `LICENSE` à la racine (ADR-011) |
| **Binaire `asa`** | Même licence ; distribué avec `LICENSE` + `README.md` dans chaque archive GoReleaser |
| **Docs publiques** | Contenu docs-site / `docs/ai/` — licence projet sauf mention contraire |
| **SDK npm** | `@laprogrammerie/asagiri` — semver indépendant, tag `sdk-v*` (ADR-026) |
| **Packs Pro (M2)** | Licence **contenu** distincte — **pas** Apache sur les packs premium |
| **Enforcement** | **Aucun** dans `cmd/asa` — pas de clé licence, pas de phone-home obligatoire |

**Principe ADR-037 :** le cœur OSS reste complet ; la monétisation est **additive** (packs, cloud, gouvernance).

---

## Versioning

| Artefact | Convention | Exemple |
|----------|------------|---------|
| **Binaire Go `asa`** | SemVer Git tag `vMAJOR.MINOR.PATCH` | `v1.0.0` |
| **Module Go** | `github.com/LaProgrammerie/asagiri` | `go install …@v1.0.0` |
| **CHANGELOG** | [Keep a Changelog](https://keepachangelog.com/) — fichier racine `CHANGELOG.md` | Sections Added / Changed / Fixed |
| **GoReleaser changelog** | Généré à la release ; exclut `docs:`, `test:`, `chore:` | `.goreleaser.yaml` |
| **SDK TypeScript** | SemVer indépendant, tag **`sdk-v*`** | `sdk-v0.2.0` |
| **Dev local** | `version.Version = "dev"` si build sans ldflags | `asa version` |

**Règle release :** une tag `v*` sur `main` déclenche `.github/workflows/release.yml` → GoReleaser.

**Pré-1.0 :** `v0.x.y` acceptable tant que l’API CLI publique peut encore évoluer ; documenter breaking changes dans CHANGELOG.

---

## Canaux d’installation

### 1. Homebrew (recommandé macOS / Linux)

Tap générique La Programmerie (plusieurs outils) :

```bash
brew tap LaProgrammerie/tap
brew install asagiri
asagiri version
asagiri doctor
```

- **Binaire Homebrew :** `asagiri` (évite le conflit macOS avec `/usr/bin/asa`). Alias optionnel : `alias asa='asagiri'`.
- **Archives GitHub / `install.sh` :** binaire `asa` dans les tarballs `asa_*` (inchangé).
- **Formule Homebrew :** `asagiri` → `LaProgrammerie/homebrew-tap/Formula/asagiri.rb`.
- **Procédure opérateur :** `docs/ai/active/release-process.md`.

**Legacy (formule `asa` sur le même tap) :**

```bash
brew install asa
```

Préférer `brew install asagiri` pour les nouvelles installs.

### 2. `go install` (depuis source module)

```bash
go install github.com/LaProgrammerie/asagiri/application/cmd/asa@v1.0.0
asa version
```

- Nécessite Go 1.25+ et toolchain Git.
- Idéal pour contributeurs et CI éphémère.

### 3. Archive release (fallback manuel)

**Convention GoReleaser v2 :** `asa_<semver-sans-v>_<os>_<arch>.tar.gz`  
Tag `vX.Y.Z` → `VER=X.Y.Z`, `os` ∈ `darwin|linux`, `arch` ∈ `amd64|arm64`.

```bash
VERSION=v1.0.0
VER="${VERSION#v}"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$(uname -m)" in x86_64|amd64) ARCH=amd64 ;; arm64|aarch64) ARCH=arm64 ;; esac
ARCHIVE="asa_${VER}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/LaProgrammerie/asagiri/releases/download/${VERSION}"
curl -fsSLO "${BASE}/checksums.txt" "${BASE}/${ARCHIVE}"
grep " ${ARCHIVE}\$" checksums.txt | sha256sum -c -
tar -xzf "${ARCHIVE}"
install -m 755 asa "${HOME}/.local/bin/asa"
asa doctor --json
```

Windows : `asa_${VER}_windows_<arch>.zip`, binaire `asa.exe`.

### 4. Script `install.sh` (recommandé curl)

```bash
curl -fsSL https://raw.githubusercontent.com/LaProgrammerie/asagiri/main/scripts/install.sh | bash
# ou version épinglée :
ASAGIRI_VERSION=v1.0.0 bash scripts/install.sh
```

Détecte OS/arch, télécharge l’archive release, vérifie SHA256 (`checksums.txt`), installe dans `~/.local/bin` par défaut.

### 5. Build depuis les sources

```bash
git clone https://github.com/LaProgrammerie/asagiri.git
cd asagiri
make build
./bin/asa version
```

Voir `Makefile` : `make build`, `make test`, `make release-check`.

---

## Release GitHub

### Infrastructure existante (ne pas réimplémenter en M1.0)

| Fichier | Rôle |
|---------|------|
| `.goreleaser.yaml` | Builds `asa` linux/darwin/windows × amd64/arm64 ; archives ; checksums ; brew |
| `.github/workflows/release.yml` | Trigger sur tag `v*` ; tests + GoReleaser |
| `.github/workflows/release-check.yml` | Validation config sur PR |
| `Makefile` | `release-snapshot`, `release-check` |

### Dépôt releases

- **GitHub :** `LaProgrammerie/asagiri`
- **Page :** https://github.com/LaProgrammerie/asagiri/releases

### Déclenchement manuel (mainteneur)

```bash
# 1. Mettre à jour CHANGELOG.md (section [Unreleased] → vX.Y.Z)
# 2. Commit + tag annoté
git tag -a v1.0.0 -m "Asagiri v1.0.0"
git push origin v1.0.0
# 3. CI release.yml → GoReleaser publie assets + brew (si HOMEBREW_TAP_GITHUB_TOKEN)
```

**Secrets CI :** `GITHUB_TOKEN` (assets release) ; `HOMEBREW_TAP_GITHUB_TOKEN` (push formule Homebrew).

### Snapshot local (pré-release)

```bash
make release-check      # goreleaser check
make release-snapshot   # dist/ sans publier
```

---

## Artefacts publiés par release (`v*`)

| Artefact | Nom | Contenu |
|----------|-----|---------|
| Archives macOS arm64 | `asa_{version}_darwin_arm64.tar.gz` | `asa`, LICENSE, README |
| Archives macOS amd64 | `asa_{version}_darwin_amd64.tar.gz` | idem |
| Archives Linux arm64 | `asa_{version}_linux_arm64.tar.gz` | idem |
| Archives Linux amd64 | `asa_{version}_linux_amd64.tar.gz` | idem |
| Archives Windows arm64 | `asa_{version}_windows_arm64.zip` | `asa.exe`, LICENSE, README |
| Archives Windows amd64 | `asa_{version}_windows_amd64.zip` | idem |
| Checksums | `checksums.txt` | SHA256 de toutes les archives |
| Notes release | GitHub Release body | Changelog généré GoReleaser |
| Formule Homebrew | `LaProgrammerie/homebrew-tap/Formula/asagiri.rb` | Poussée automatiquement par GoReleaser |
| Script install | `scripts/install.sh` | Curl archive + vérif SHA256 |
| Procédure release | `docs/ai/active/release-process.md` | Tag, brew, rollback |

**Hors scope release binaire :** packs Pro (M2), CLI cloud (M3), `asagiri-packs` (M2).

---

## Changelog

- **Fichier canon :** `CHANGELOG.md` à la racine (maintenu manuellement avant tag).
- **Format :** [Keep a Changelog](https://keepachangelog.com/) — `Unreleased`, puis `## [v1.0.0] - YYYY-MM-DD`.
- **GoReleaser** complète les release notes GitHub à partir des commits (filtres dans `.goreleaser.yaml`).
- **SDK :** `sdk/typescript/CHANGELOG.md` — cycle séparé tag `sdk-v*`.

**Première release publique OSS :** remplir `CHANGELOG.md` avec le périmètre T29 + trust + gates (référence handoff agent-orchestration-platform-v1).

---

## Checklist release OSS

Exécuter **avant** tout tag `v*` public :

### Qualité code (moteur inchangé fonctionnellement)

- [ ] `cd application && go build ./...`
- [ ] `cd application && go test ./... -count=1`
- [ ] `go test ./internal/cli/docgen/... -run TestNodiff -count=1`
- [ ] `make release-check`

### Smoke CLI (binaire candidat ou snapshot)

- [ ] `asa version` — version, commit, date
- [ ] `asa doctor --json` — exit 0, JSON stdout, stderr vide
- [ ] `asa doctor architecture --json` — exit 0
- [ ] `asa agents list --json` — exit 0
- [ ] `asa onboard --check-only` (dans dépôt Git de test)

### Documentation

- [ ] `CHANGELOG.md` section version complète
- [ ] README install aligné (Homebrew → install.sh → archive manuelle)
- [ ] Parcours externe : `docs/ai/active/getting-started-public.md`
- [ ] docs-site déployée si changement CLI docgen
- [ ] `current-spec.md` / `handoff.md` à jour

### Publication

- [ ] Tag `vX.Y.Z` sur commit validé
- [ ] Workflow `release.yml` vert
- [ ] Assets GitHub présents (6 archives + checksums)
- [ ] Formule Homebrew mise à jour (tap)
- [ ] Annonce (README badge / release notes) — optionnel M1.4

### Post-release

- [ ] Vérifier install Homebrew : `brew upgrade asa` ou cible `asagiri`
- [ ] Vérifier curl archive sur macOS + Linux
- [ ] Issue milestone fermée si applicable

---

## Compatibilité éditions futures

| Édition | Impact sur distribution OSS |
|---------|----------------------------|
| **OSS** | Binaire unique `asa` — complet, sans compte |
| **Pro Local (M2)** | Packs importés dans `.asagiri/` ; **même binaire** ; outil `asagiri-packs` séparé |
| **Team Cloud (M3)** | Sync opt-in ; CLI locale reste source d’exécution |
| **Enterprise (M4)** | Appliance / policies — pas de fork OSS |

**Invariant :** aucune release OSS ne doit exiger cloud, compte ou clé Pro pour `asa work`.

---

## Hors scope M1 (explicitement)

| Hors scope | Phase |
|------------|-------|
| Licence enforcement / télémétrie obligatoire | Jamais dans moteur (ADR-037) |
| Packs payants Pro Local | M2 |
| Console cloud / workspace | M3 |
| SSO / audit Enterprise | M4 |
| cosign / SBOM / provenance SLSA | M1.1+ optionnel |
| deb/rpm/scoop/chocolatey | Post-M1 |
| Renommage module Go ou binaire | Hors scope — `asa` stable |
| Modification workflow/gates/trust/agentspec | Interdit |
| Implémentation GoReleaser from scratch | Déjà présent — M1.1 = migration tap + install.sh |
| Première tag `v1.0.0` publique | M1.4 (après checklist verte) |

---

## Références

- `.goreleaser.yaml` — config release actuelle
- `README.md` § Install
- `docs/ai/03-standards.md` — `make release-check`
- `docs/migration/github-rename-asagiri.md` — rename dépôt
- `docs/consolidation/08-oss-readiness.md` — audit readiness
- `.kiro/specs/monetization-distribution-v1/tasks.md` — découpage M1.0–M1.4
