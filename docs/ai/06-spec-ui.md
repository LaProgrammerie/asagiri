# Spec-ui — Asagiri Experience Platform (canon `docs/ai`)

**Statut :** livrée (`2026-05-29`)  
**Spec racine :** [`spec-ui.md`](../../spec-ui.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)

---

## 1. Résumé

Spec-ui introduit l’Experience Platform terminal au-dessus des primitives CLI existantes :

- `asa` (TTY) ouvre Mission Control ;
- les commandes directes restent stables (`asa work`, `asa investigate`, `asa verify trust`, etc.) ;
- la TUI est un client du moteur via CommandBus/QueryBus ;
- aucune logique métier n’est portée dans `internal/ui`.

---

## 2. Entrées UI livrées

| Entrée | Comportement |
|---|---|
| `asa` (TTY) | Mission Control |
| `asa mission` | Mission Control |
| `asa dashboard` | Dashboard live |
| `asa agents watch` | Agent Theatre |
| `asa graph` | Graph Explorer |
| `asa flow` / `asa flow open <name>` | Flow Explorer |
| `asa knowledge` | Knowledge Explorer |
| `asa trust` | Trust Explorer |
| `asa replay open <id>` | Replay Explorer |
| `asa prototype` | Prototype Mode |
| `asa explain` | Explain screen |

---

## 3. Architecture runtime UI

| Zone | Rôle |
|---|---|
| `application/internal/ui/app` | Bubble Tea model principal, router, palette, help, refresh loop |
| `application/internal/ui/bus` | Adaptateurs CommandBus/QueryBus vers workflow/runtime/trust/knowledge/replay/graph |
| `application/internal/ui/screens/*` | Rendu read-only orienté cockpit |
| `application/internal/ui/components` | Panels, cards, event feed |
| `application/internal/ui/layout` | Layout single/split + focus navigation |
| `application/internal/ui/theme` | Palettes + styles (dark/light/minimal/high-contrast/cyber) |
| `application/internal/ui/input` | Raccourcis clavier + support souris |

Invariant : **coexistence** avec `internal/tui` (specv3) sans régression.

---

## 4. Polish lot 6 (§25–28)

Livré :

- resize souris de panneaux (basique) via wheel et drag de divider ;
- mode no-animation (glyphes statiques pour états `running`) ;
- usage explicite high-contrast (thème + aide accessibilité) ;
- responsive raffiné (dashboard compact/wide/ultra-wide) ;
- écran d’aide accessibilité + raccourcis (`?`).

---

## 5. Configuration `ui:`

Champs supportés :

- `mode`, `live_logs`, `progress_bars`, `compact` ;
- `default_screen`, `theme`, `mouse`, `animations` ;
- `refresh_interval_ms`, `compact_threshold` ;
- `show_cli_equivalents`, `confirm_destructive_actions`.

Exemple :

```yaml
ui:
  default_screen: mission
  theme: asagiri-dark
  mouse: true
  animations: true
  refresh_interval_ms: 500
  compact_threshold: 100
  show_cli_equivalents: true
  confirm_destructive_actions: true
```

---

## 6. Documentation publique (§31)

Les sections `experience/` sont publiées en **4 locales** :

- `en/experience/*`
- `fr/experience/*`
- `de/experience/*`
- `es/experience/*`

Pages : `index`, `mission-control`, `dashboard`, `command-palette`, `keyboard-shortcuts`, `mouse-support`, `themes`, `accessibility`.

---

## 7. Validation et clôture §33

Validation exécutée :

```bash
go test ./... -count=1
make build
make build && ./bin/asa docs generate-cli
cd docs-site && pnpm docs:check
```

Matrice handoff : **24/24 (100 %)**.  
Critères d’acceptation §33 : **clôturés**.
