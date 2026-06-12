# Getting started — utilisateur externe (< 5 minutes)

> **Public cible :** développeur qui découvre Asagiri sans compte La Programmerie.  
> **Canon install / release :** [`distribution-oss.md`](distribution-oss.md), [`release-process.md`](release-process.md).

## What Asagiri is / is not

**Asagiri n’est pas un agent IA de plus.** C’est un **control plane local-first** pour le développement assisté par IA :

- il **orchestre** des runtimes externes (Cursor, Claude Code, Codex, Kiro, Ollama, …) ;
- il applique un **pipeline déterministe** (plan → dev → verify → review) ;
- il conserve **audit, ledger, replay et policies** sur votre machine ;
- il isole le travail dans des **git worktrees** et un état SQLite local (`.asagiri/`).

**Ce n’est pas :** un chat, un IDE, un SaaS obligatoire, ni un remplacement de vos agents — ils restent les moteurs d’exécution ; Asagiri coordonne et trace.

## No account required

| | |
|---|---|
| Compte La Programmerie | **Non** |
| Quota ou licence dans le binaire | **Non** |
| Cloud obligatoire | **Non** |
| Envoi automatique des runs vers un serveur | **Non** |

Tout fonctionne hors ligne une fois le binaire installé et les agents configurés localement. Notion / cloud sont **optionnels** plus tard.

---

## Parcours en 5 minutes

### 1. Installation (choisir un canal)

**Ordre recommandé :** Homebrew → script → archive manuelle.

#### Homebrew (macOS / Linux)

```bash
brew tap LaProgrammerie/tap
brew install asagiri
asagiri version
asagiri doctor
```

Sur macOS, ne pas utiliser `asa` seul si `which asa` pointe vers `/usr/bin/asa`. Alias optionnel : `alias asa='asagiri'`.

#### Script `install.sh`

```bash
curl -fsSL https://raw.githubusercontent.com/LaProgrammerie/asagiri/main/scripts/install.sh | bash
```

Version épinglée : `ASAGIRI_VERSION=v1.0.0` (remplacer par un [tag Releases](https://github.com/LaProgrammerie/asagiri/releases)).

#### Archive manuelle (fallback)

Convention GoReleaser v2 : `asa_<semver-sans-v>_<os>_<arch>.tar.gz`  
Exemple tag `v1.0.0` → archive `asa_1.0.0_darwin_arm64.tar.gz`.

```bash
VERSION=v1.0.0          # tag GitHub Releases
VER="${VERSION#v}"      # semver dans le nom d'archive
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch"; exit 1 ;;
esac
ARCHIVE="asa_${VER}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/LaProgrammerie/asagiri/releases/download/${VERSION}"

curl -fsSLO "${BASE}/checksums.txt"
curl -fsSLO "${BASE}/${ARCHIVE}"
grep " ${ARCHIVE}\$" checksums.txt | sha256sum -c -
tar -xzf "${ARCHIVE}"
install -m 755 asa "${HOME}/.local/bin/asa"
export PATH="${HOME}/.local/bin:$PATH"
asa version
```

Windows : archive `asa_${VER}_windows_<arch>.zip`, binaire `asa.exe`.

### 2. Vérifier l’environnement

```bash
asagiri version
asagiri doctor
```

`doctor` contrôle Git, la config, les binaires agents sur le `PATH`, et les erreurs courantes **avant** un `work`.

### 3. Initialiser un dépôt

Dans la racine d’un dépôt Git :

```bash
cd your-repo
asa init
```

Crée `.asagiri/` (config, SQLite, runs, worktrees). Ajoutez `.asagiri/state.sqlite` et `.asagiri/worktrees/` au `.gitignore`.

Alternative guidée (wizard) : `asa onboard` — équivalent plus verbeux pour la première fois.

### 4. Agents embarqués → disque

```bash
asa agents list
asa agents sync --write
```

Matérialise les AgentSpecs embarqués vers `.asagiri/agents/` pour que `doctor` et les workflows résolvent les profils sans surprise.

### 5. Premier usage minimal (sans cloud)

Sans appeler d’agents payants :

```bash
export ASA_DRY_RUN=1
asa work "explore my feature" --plan-only --yes
asa status
```

`ASA_DRY_RUN=1` : orchestration locale, pas de subprocess agent réel — idéal pour valider config et routing.

Quand les providers sont prêts :

```bash
unset ASA_DRY_RUN
asa work "my feature" --estimate-only
asa work "my feature" --yes
```

---

## Troubleshooting rapide

| Symptôme | Action |
|----------|--------|
| `asa: command not found` | Vérifier `PATH` (`~/.local/bin` ou `/usr/local/bin`) |
| `doctor` : agent manquant | Installer l’outil (cursor-agent, kiro, …) ou ajuster `.asagiri/config.yaml` |
| `doctor` : registry agents | `asa agents sync --write` |
| Archive introuvable | Vérifier tag `VERSION` et nom `asa_${VER}_${OS}_${ARCH}.tar.gz` sur [Releases](https://github.com/LaProgrammerie/asagiri/releases) |
| Checksum échoue | Retélécharger archive + `checksums.txt` pour la même version |
| Brew 404 | Première release pas encore publiée, ou tap non à jour — utiliser `install.sh` ou archive |
| Pas de dépôt Git | `asa init` / `work` nécessitent un repo Git |

JSON pour scripts : `asa doctor --json`, `asa agents list --json`.

---

## Suite

- Site : [Installation](https://asagiri.dev/docs/getting-started/installation) (multilingue)
- Détail OSS : [`distribution-oss.md`](distribution-oss.md)
- Mainteneur release : [`release-process.md`](release-process.md)
- README racine : parcours quotidien (`asa work`, `asa next`, `asa status`)
