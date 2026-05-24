Spec — Release distribution & Homebrew install

1. Objectif

Mettre en place une chaîne de distribution propre pour AgentFlow afin de permettre :

* la génération automatique de releases GitHub ;
* la publication de binaires multi-plateformes ;
* la génération de checksums ;
* l’installation via Homebrew ;
* une documentation d’installation claire ;
* une base saine pour distribution open source.

La cible principale est une CLI Go distribuée sous forme de binaires autonomes.

⸻

2. Périmètre

Cette feature couvre :

* builds Linux/macOS/Windows ;
* architectures amd64/arm64 ;
* packaging en archives .tar.gz / .zip ;
* checksums SHA256 ;
* GitHub Release ;
* version injectée au build ;
* Homebrew tap ;
* formule Homebrew ;
* documentation installation ;
* vérification post-install.

Cette feature ne couvre pas encore :

* packages Debian/RPM ;
* Scoop ;
* Winget ;
* Docker image ;
* signature cosign ;
* SBOM ;
* notarization macOS ;
* release automatique sur chaque commit.

Ces sujets pourront être traités en V2.

⸻

3. Objectifs produit

L’utilisateur doit pouvoir installer AgentFlow simplement.

Installation macOS/Linux via Homebrew

brew tap LaProgrammerie/agentflow
brew install agentflow

ou, si le tap est nommé autrement :

brew tap LaProgrammerie/tap
brew install agentflow

Installation manuelle

curl -L https://github.com/LaProgrammerie/agentflow/releases/download/v0.1.0/agentflow_Darwin_arm64.tar.gz -o agentflow.tar.gz
tar -xzf agentflow.tar.gz
sudo mv agentflow /usr/local/bin/agentflow
agentflow version

⸻

4. Convention de version

AgentFlow doit utiliser des tags Git SemVer :

v0.1.0
v0.1.1
v0.2.0
v1.0.0

La version doit être injectée au build via -ldflags.

Exemple existant :

go build -ldflags "-X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Version=dev" \
  -o bin/agentflow ./application/cmd/agentflow

À généraliser :

go build \
  -ldflags "-s -w -X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Version=${VERSION}" \
  -o dist/agentflow ./application/cmd/agentflow

Le package application/internal/version doit exposer au minimum :

package version
var Version = "dev"
var Commit = "unknown"
var Date = "unknown"

Et la commande :

agentflow version

doit afficher :

AgentFlow v0.1.0
commit: abc1234
built: 2026-05-17T12:00:00Z

⸻

5. Artefacts attendus

Pour chaque release, produire au minimum :

agentflow_Darwin_arm64.tar.gz
agentflow_Darwin_x86_64.tar.gz
agentflow_Linux_arm64.tar.gz
agentflow_Linux_x86_64.tar.gz
agentflow_Windows_arm64.zip
agentflow_Windows_x86_64.zip
checksums.txt

Chaque archive doit contenir :

agentflow / agentflow.exe
LICENSE
README.md

Optionnel :

completion/
  agentflow.bash
  agentflow.zsh
  agentflow.fish

⸻

6. Plateformes cibles

OS	Arch	Archive
macOS	arm64	.tar.gz
macOS	amd64/x86_64	.tar.gz
Linux	arm64	.tar.gz
Linux	amd64/x86_64	.tar.gz
Windows	arm64	.zip
Windows	amd64/x86_64	.zip

Convention de nommage recommandée pour compatibilité Homebrew :

agentflow_${VERSION}_${OS}_${ARCH}.tar.gz

ou convention GoReleaser classique :

agentflow_${OS}_${ARCH}.tar.gz

Décision à prendre selon l’outil de release retenu.

⸻

7. Outil de release recommandé

Utiliser GoReleaser.

Justification :

* standard très répandu pour CLI Go ;
* support multi-OS / multi-arch ;
* génération checksums ;
* GitHub Release ;
* intégration Homebrew Tap ;
* injection version/commit/date ;
* configuration déclarative ;
* bonne maintenabilité.

⸻

8. Configuration GoReleaser

Créer :

.goreleaser.yaml

Configuration cible :

version: 2
project_name: agentflow
before:
  hooks:
    - go mod tidy
    - go test ./...
builds:
  - id: agentflow
    main: ./application/cmd/agentflow
    binary: agentflow
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Version={{.Version}}
      - -X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Commit={{.Commit}}
      - -X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Date={{.Date}}
archives:
  - id: default
    builds:
      - agentflow
    format_overrides:
      - goos: windows
        formats:
          - zip
    files:
      - LICENSE
      - README.md
checksum:
  name_template: checksums.txt
snapshot:
  version_template: "{{ incpatch .Version }}-next"
changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
release:
  github:
    owner: LaProgrammerie
    name: agentflow
  draft: false
  prerelease: auto

Adapter le module path si le repo final n’est plus github.com/LaProgrammerie/hyper-fast-builder.

⸻

9. GitHub Actions release

Créer ou adapter :

.github/workflows/release.yml

Objectif : lancer une release uniquement quand un tag v* est poussé.

name: Release
on:
  push:
    tags:
      - "v*"
permissions:
  contents: write
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    timeout-minutes: 30
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - name: Run tests
        run: go test ./...
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

Si le go.mod n’est pas à la racine, adapter :

working-directory: application

ou utiliser go-version explicite.

⸻

10. Snapshot release locale

Ajouter une commande Makefile :

release-snapshot:
	goreleaser release --snapshot --clean

Ajouter aussi :

release-check:
	goreleaser check

Objectif : pouvoir tester la release localement avant de créer un tag.

⸻

11. Homebrew Tap

11.1 Stratégie recommandée

Créer un repository dédié :

github.com/LaProgrammerie/homebrew-tap

Homebrew utilise la convention :

brew tap LaProgrammerie/tap
brew install agentflow

Le repo doit contenir :

Formula/agentflow.rb
README.md

11.2 Formule Homebrew attendue

Exemple cible :

class Agentflow < Formula
  desc "Deterministic orchestration for AI coding workflows"
  homepage "https://github.com/LaProgrammerie/agentflow"
  version "0.1.0"
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/LaProgrammerie/agentflow/releases/download/v0.1.0/agentflow_Darwin_arm64.tar.gz"
    sha256 "..."
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/LaProgrammerie/agentflow/releases/download/v0.1.0/agentflow_Darwin_x86_64.tar.gz"
    sha256 "..."
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/LaProgrammerie/agentflow/releases/download/v0.1.0/agentflow_Linux_arm64.tar.gz"
    sha256 "..."
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/LaProgrammerie/agentflow/releases/download/v0.1.0/agentflow_Linux_x86_64.tar.gz"
    sha256 "..."
  end
  def install
    bin.install "agentflow"
  end
  test do
    assert_match version.to_s, shell_output("#{bin}/agentflow version")
  end
end

Mais la formule ne doit idéalement pas être maintenue à la main.

⸻

12. Homebrew via GoReleaser

GoReleaser doit publier automatiquement la formule dans LaProgrammerie/homebrew-tap.

Ajouter dans .goreleaser.yaml :

brews:
  - name: agentflow
    repository:
      owner: LaProgrammerie
      name: homebrew-tap
    directory: Formula
    homepage: "https://github.com/LaProgrammerie/agentflow"
    description: "Deterministic orchestration for AI coding workflows"
    license: "Apache-2.0"
    test: |
      system "#{bin}/agentflow", "version"
    install: |
      bin.install "agentflow"

Selon les permissions GitHub, il peut être nécessaire d’utiliser un token personnel :

HOMEBREW_TAP_GITHUB_TOKEN

Puis dans GitHub Actions :

env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  GORELEASER_CURRENT_TAG: ${{ github.ref_name }}
  HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}

Et configurer GoReleaser pour utiliser ce token si nécessaire.

Décision : si le repo tap est différent du repo principal, prévoir un secret dédié.

⸻

13. Installation documentation

Mettre à jour la documentation :

docs-site/content/docs/getting-started/installation.mdx
README.md

Inclure :

Homebrew

brew tap LaProgrammerie/tap
brew install agentflow
agentflow version

Manual install

curl -LO <release-url>
sha256sum -c checksums.txt

From source

git clone https://github.com/LaProgrammerie/agentflow.git
cd agentflow
make build
./bin/agentflow version

Verify install

agentflow doctor

⸻

14. Release process documentation

Créer :

docs/release-process.md

Contenu attendu :

# Release process
1. Ensure main is green.
2. Update changelog if needed.
3. Create tag:
```bash
git tag v0.1.0
git push origin v0.1.0

4. GitHub Actions runs GoReleaser.
5. GitHub Release is published.
6. Homebrew tap is updated.
7. Verify install:

brew update
brew install LaProgrammerie/tap/agentflow
agentflow version
---
## 15. Vérifications CI attendues
Avant release :
- `go test ./...` ;
- `go vet ./...` si disponible ;
- build GoReleaser check ;
- release snapshot test ;
- version command test.
Ajouter un workflow non-release pour vérifier GoReleaser sur PR :
```yaml
name: Release check
on:
  pull_request:
    paths:
      - ".goreleaser.yaml"
      - ".github/workflows/release.yml"
      - "application/**"
      - "go.mod"
      - "go.sum"
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: check

⸻

16. Critères d’acceptation

La feature est terminée si :

* un tag v* déclenche une release GitHub ;
* les artefacts Linux/macOS/Windows amd64/arm64 sont générés ;
* chaque archive contient le binaire, README et LICENSE ;
* checksums.txt est généré ;
* agentflow version affiche version/commit/date corrects ;
* une release snapshot peut être générée localement ;
* une formule Homebrew est générée ou mise à jour automatiquement ;
* brew tap LaProgrammerie/tap && brew install agentflow fonctionne ;
* brew test agentflow fonctionne ;
* la documentation d’installation est à jour ;
* le release process est documenté ;
* aucun secret n’est exposé dans les logs.

⸻

17. Risques et garde-fous

17.1 Mauvais module path

Le module Go actuel peut encore porter un nom temporaire.

Action : vérifier le module path avant d’écrire les ldflags définitifs.

17.2 Permissions Homebrew tap

Le GITHUB_TOKEN du repo principal ne peut pas toujours pousser dans un autre repo.

Action : prévoir HOMEBREW_TAP_GITHUB_TOKEN avec accès limité au tap.

17.3 Incompatibilité archives Homebrew

Homebrew doit pouvoir télécharger l’archive et trouver le binaire à la racine.

Action : vérifier la structure des archives en snapshot.

17.4 Windows

Windows doit produire agentflow.exe et une archive .zip.

Action : tester l’archive générée.

⸻

18. Roadmap distribution V2

Après cette feature, envisager :

* signatures cosign ;
* SBOM ;
* attestations SLSA ;
* packages Debian/RPM ;
* Scoop ;
* Winget ;
* Docker image ;
* completions shell installées via Homebrew ;
* manpages ;
* changelog automatisé plus propre ;
* notarization macOS si nécessaire.

⸻

19. Mission agent

Implémenter une chaîne de release complète pour AgentFlow avec GoReleaser, GitHub Actions et Homebrew Tap.

Contraintes :

* ne pas casser le build existant ;
* préserver la commande agentflow version ;
* adapter les chemins au repo réel ;
* tester avec une release snapshot ;
* documenter l’installation ;
* préparer Homebrew pour usage open source ;
* éviter tout secret hardcodé.

Résultat attendu :

git tag v0.1.0
git push origin v0.1.0

produit automatiquement :

* GitHub Release ;
* binaires multi-plateformes ;
* checksums ;
* formule Homebrew ;
* installation via brew.