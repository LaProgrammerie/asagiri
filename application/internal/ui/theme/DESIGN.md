# Asagiri Terminal Design System — Lot 1–2

Spécification visuelle designer → dev. Implémentation cible : `theme/palette.go`, `theme/theme.go`, lipgloss (cf. `internal/tui/rich_ui.go` : `RoundedBorder`, titres `Bold`).

Références : `spec-ui.md` §8, §11.2, §26, §28.

---

## 1. Palettes (tokens lipgloss)

Couleurs en `#RRGGBB` (truecolor / 256). Noms de tokens stables — ne pas renommer sans migration config.

| Token | Rôle | asagiri-dark | asagiri-light | high-contrast |
|-------|------|--------------|---------------|---------------|
| **primary** | Accent marque, focus, running, barres actives | `#9B6DFF` | `#6D28D9` | `#E8D4FF` |
| **success** | Terminé, OK, trust élevé | `#2DD4BF` | `#0D9488` | `#33FF99` |
| **warning** | Risque, dégradé, paused | `#F59E0B` | `#D97706` | `#FFCC00` |
| **error** | Échec, blocked, destructive | `#F87171` | `#DC2626` | `#FF5555` |
| **muted** | Labels, timestamps, états secondaires | `#6B7280` | `#64748B` | `#B0B0B0` |
| **border** | Contours panneaux, séparateurs, barres vides | `#3D3552` | `#CBD5E1` | `#FFFFFF` |

Tokens complémentaires (non demandés Lot 1 mais utilisés par les composants) :

| Token | asagiri-dark | asagiri-light | high-contrast |
|-------|--------------|---------------|---------------|
| `fg` (texte principal) | `#E5E7EB` | `#0F172A` | `#FFFFFF` |
| `bg` (fond app) | `#0D0F14` | `#F8F9FB` | `#000000` |
| `fgInverse` (texte sur primary) | `#0D0F14` | `#FFFFFF` | `#000000` |

**Règles lipgloss**

```go
// Exemple — palette.go
Primary:   lipgloss.Color("#9B6DFF")
Success:   lipgloss.Color("#2DD4BF")
BorderStyle: lipgloss.RoundedBorder() // héritage rich_ui.go
PanelTitle:  lipgloss.NewStyle().Bold(true).Foreground(Primary)
PanelBorder: lipgloss.NewStyle().Border(BorderStyle).BorderForeground(Border)
FocusRing:   lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(Primary)
```

Thèmes §26 hors Lot 1–2 : `minimal` (dark sans couleur sémantique, fg/muted/border seulement), `cyber` (primary → `#00F5FF`, success → `#39FF14`) — non spécifiés ici.

---

## 2. États visuels (§8.3)

Symbole **toujours** suivi d’un label texte (a11y §28). En `animations: false`, remplacer le spinner par le symbole statique `◐`.

| État | Couleur token | Symbole | Label plain (fallback) |
|------|---------------|---------|------------------------|
| **idle** | `muted` | `○` U+25CB | `[idle]` |
| **running** | `primary` | `⠋` U+280B (cycle ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏) | `[running]` |
| **success** | `success` | `✓` U+2713 | `[ok]` |
| **warning** | `warning` | `⚠` U+26A0 | `[warn]` |
| **error** | `error` | `✕` U+2715 | `[error]` |
| **blocked** | `error` | `⊘` U+2298 | `[blocked]` |
| **paused** | `warning` | `⏸` U+23F8 | `[paused]` |
| **waiting** | `muted` | `○` U+25CB (ou `⋯` U+22EF si file) | `[waiting]` |
| **unknown** | `muted` | `?` | `[unknown]` |

**Usage composants**

- `StatusBadge` : symbole + label court (`running`, `done`, …).
- Flow steps / Agent Theatre : symbole seul acceptable **si** la colonne voisine porte le nom (cf. wireframe §11.2).
- Barres trust : remplissage `success` si ≥ 80 %, `warning` si 50–79 %, `error` si < 50 %.

---

## 3. Typographie terminal

Polices (ordre de préférence si le terminal les expose) :

| Rôle | Police | Style lipgloss |
|------|--------|----------------|
| UI générale | Inter, IBM Plex Sans, system-ui | `DefaultStyle` |
| Données, métriques, tables, events | JetBrains Mono, monospace | `.Italic(false)` |
| Code / CLI equivalent | JetBrains Mono | `.Foreground(muted)` |

**Titres panneaux** (ex. `Runtime`, `Trust`, `Agent Theatre`)

- `Bold(true)`, couleur `fg` (dark/light) ou `primary` pour le bandeau header `ASAGIRI`.
- Casse : Title Case (pas ALL CAPS sauf marque `ASAGIRI`).
- Padding titre : 0–1 ligne au-dessus du contenu (aligné `rich_ui.go` `Padding(1, 2)`).

**Métriques** (`MetricCard`, lignes type `Agents  3`)

- Label : poids normal, `muted`, largeur colonne fixe ~14–18 runes.
- Valeur : `Bold`, `fg`, alignée à droite dans la colonne ou tabular nums.
- Unités / devise : `muted`, sans bold (`€0.42` → valeur bold, symbole muted).

**Barres de progression** (trust, coût budget)

- Rempli : `█` U+2588, couleur sémantique ou `primary`.
- Vide : `░` U+2591, couleur `border`.
- Pourcentage : `Bold`, 2 chiffres + `%`, après la barre.

**Tables** (`TableView`, events)

- En-tête : `Bold`, séparateur `border`.
- Corps : monospace, lignes alternées optionnelles (`bg` légèrement plus clair + `#161822` dark / `#EEF2F7` light).
- Timestamp events : `muted`, format `HH:MM`, largeur 5.
- Troncature : ellipsis `…`, jamais couper le symbole d’état.

**Échelle**

- Terminal ≥ 100 cols (standard) : typo ci-dessus.
- `< compact_threshold` (100) : labels abrégés, métriques une colonne, titres sans sous-ligne.

---

## 4. Mission Control — wireframe texte (§11.2)

Layout cible : **standard** (≥ 100 cols, split horizontal). Focus initial : panneau Runtime.

```
╭────────────────────────────── ASAGIRI ──────────────────────────────╮  ← header: primary bold, border
│ Workspace  workspace-saas              Runtime   ● running          │  ← kv pairs: label muted, value bold
│ Branch     onboarding-v2                Session   active            │  ← runtime status: symbole §2
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Runtime ────────────╮ ╭──────────── Trust ───────────────╮  ← split 50/50, min 36 cols/pane
│ Agents      3                   │ │ Architecture  █████████░░ 82%    │
│ Sessions    4                   │ │ Security      ███████░░░░ 71%    │
│ Queue       2                   │ │ Observability ██████░░░░░ 63%    │
│ Cost today  €0.42               │ │ Regression    ████████░░░ 78%    │
╰────────────────────────────────╯ ╰──────────────────────────────────╯
╭──────────── Active Flow ─────────────────────────────────────────────╮  ← full width
│ Onboarding                                                           │  ← flow name: Bold
│ ✓ create_workspace   ⠋ invite_member   ○ accept_invite   ○ complete  │  ← pipeline horizontal, wrap si narrow
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Agent Theatre ───────────────────────────────────────────╮
│ Investigator ✓ done     Architect ✓ done     Implementer ⠋ running   │  ← agent + état §2
│ Reviewer     ○ waiting  Validator ○ waiting                          │
╰──────────────────────────────────────────────────────────────────────╯
╭──────────── Events ──────────────────────────────────────────────────╮  ← scroll interne, max 5 lignes visibles
│ 08:14 investigation.completed     confidence=0.78                    │
│ 08:15 graph.generated             nodes=12 edges=16                  │
│ 08:16 implementation.started      agent=cursor                       │
╰──────────────────────────────────────────────────────────────────────╯
```

**Grille & comportement**

| Zone | Composant | Query / source |
|------|-----------|----------------|
| Header | `Panel` + `Breadcrumb` | workspace, branch, runtime, session |
| Runtime / Trust | `MetricCard` ×2 | runtime status, trust summary |
| Active Flow | `FlowCard` | flow critique en cours |
| Agent Theatre | `AgentCard` grid | agents actifs |
| Events | `EventFeed` | derniers N événements |

**Responsive (§27, Lot 2 minimal)**

- `< 100 cols` : colonne unique — Header → Runtime → Trust → Flow → Agents → Events (stack vertical).
- `Ctrl+M` depuis n’importe quel écran ; `q` / `Esc` retour selon router.

**Footer optionnel Lot 2** (non bloquant) : ligne `muted` — `? help · Ctrl+P palette · Ctrl+D dashboard`.

---

## 5. Accessibilité — contraintes designer → dev (§28)

| Exigence | Implémentation dev |
|----------|-------------------|
| Navigation clavier complète | Tous panneaux focusables (`Tab` / `Shift+Tab`) ; ordre logique header → panes → events ; raccourcis §10.1 sans conflit OS. |
| Contraste élevé | Thème `high-contrast` selectable via `ui.theme` ; ratios ≥ 4.5:1 texte/fond (tokens §1). |
| `no-animation` / `animations: false` | Spinner `⠋` → symbole fixe `◐` ; pas de shimmer ; refresh compteurs sans clignotement. |
| Mode plain | `ui.mode: plain` ou non-TTY : pas de bordures Unicode, états = labels `[running]` du tableau §2. |
| Screen reader (dans la mesure du possible) | Annoncer changements via ligne statut plain ; jamais couleur seule pour un état ; symbole + mot (`✓ done` pas `✓` seul hors pipeline). |
| Raccourcis listables | Écran `?` / Help : tableau complet ; Command Palette expose équivalent CLI (§3.1). |
| Souris optionnelle | Clic focus pane ; scroll events ; clavier reste suffisant pour toute action Lot 2. |
| Couleur + forme | États error/warning : symbole distinct + couleur (pas rouge vs orange seuls). |
| Destructive (§30) | `error` border + confirmation ; afficher `CLI equivalent:` en monospace muted. |

**Checklist revue design avant merge Lot 2**

- [ ] Mission Control lisible en 80×24 (compact).
- [ ] `high-contrast` : tous les états §2 distinguables sans couleur (symboles différents).
- [ ] Plain output reprend les mêmes infos que le wireframe §4.
- [ ] Focus visible (`FocusRing` primary) sur panneau actif.

---

## 6. Mapping config

```yaml
ui:
  theme: asagiri-dark   # asagiri-light | high-contrast
  animations: true
  compact: false
```

Variable d’env / flag `--plain` force le rendu sans styles (tokens ignorés, labels §2).
