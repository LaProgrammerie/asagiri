# Rapport d'audit — Asagiri (`asa`)

> **Nature :** audit réel de la plateforme (activité, **pas** un nouveau
> sous-système). Mené le `2026-06-03` sur le dépôt courant.
> **Méthode :** exécution réelle (`go build` / `vet` / `test` / `golangci-lint`),
> comparaison de l'arbre Cobra vivant ↔ doc générée ↔ locales, scan secrets,
> lecture ciblée du code d'orchestration. Aucune modification de code effectuée.
> **Suite :** ce rapport sert de base à la spec de correction/simplification.

## 1. Synthèse

L'outil est **globalement sain** : build, `vet` et toute la suite de tests
passent au vert ; aucun secret en clair ; les 4 locales de la doc sont alignées
hors référence CLI générée (volontairement `en`-only). La dérive réelle est
**concentrée et de cause unique** : la commande `runs` (ajoutée par
cockpit-consolidation, ADR-029) n'a pas été suivie d'une régénération de la
doc CLI. S'ajoute un **blocage outillage** (golangci-lint incompatible avec la
version Go cible) et quelques **incohérences de conception** dans la couche de
routing qui rejoignent directement ton ressenti de « surcouches ».

| Verdict | Détail |
|---------|--------|
| `make build` | ✅ exit 0 |
| `go vet ./...` | ✅ exit 0 |
| `go test ./...` | ✅ exit 0 (tous packages verts) |
| `golangci-lint run` | ❌ **ne s'exécute pas** (toolchain go1.24 < cible go1.25.0) |
| Secrets en clair | ✅ aucun |
| Alignement locales (hors CLI générée) | ✅ en = fr = de = es |
| Parité CLI ↔ doc générée | ❌ **dérive** (`runs` manquant + ~50 pages périmées) |

## 2. Constats (findings)

> Sévérité : `info` < `warn` < `error` < `blocking`. Statut initial : `ouvert`.

### AUD-001 — `runs.mdx` manquant dans la doc CLI générée
- **Zone :** docs-site / docgen. **Sévérité : `error`. Statut : ouvert.**
- **Preuve :** `asa runs` est enregistrée dans
  `application/internal/cli/root.go` (`newRunsCmd`), mais
  `docs-site/content/docs/en/cli/generated/runs.mdx` **n'existe pas**. La
  régénération (`asa docs generate-cli --output /tmp/...`) **produit** bien
  `runs.mdx` (113 fichiers vs 112 committés, hors `meta.json`).
- **Cause :** `asa docs generate-cli` non relancé après l'ajout de `runs`
  (ADR-029, `2026-05-31`).

### AUD-002 — ~50 pages CLI générées périmées (lien fratrie « Runs » manquant)
- **Zone :** docs-site / docgen. **Sévérité : `error`. Statut : ouvert.**
- **Preuve :** `diff -rq` entre les pages committées et une régénération propre
  signale ~50 fichiers divergents (`work.mdx`, `onboard.mdx`, `spec.mdx`, …).
  La seule différence par fichier est l'ajout de la ligne
  `> - [Runs](./runs.mdx)` dans la liste des commandes sœurs.
- **Cause :** identique à AUD-001 (régénération non effectuée). **Une seule
  action corrige AUD-001 et AUD-002.**

### AUD-003 — `golangci-lint` inopérant (version Go cible)
- **Zone :** outillage qualité / CI. **Sévérité : `error`. Statut : ouvert.**
- **Preuve :** `golangci-lint run` →
  `the Go language version (go1.24) used to build golangci-lint is lower than
  the targeted Go version (1.25.0)`. `go.mod` déclare `go 1.25.0` ; toolchain
  local `go1.25.0`.
- **Impact :** le garde-fou lint (Req 10.4 visé) ne peut pas s'exécuter →
  readiness production non prouvable. **Bloquant pour l'OSS si la CI dépend de
  golangci-lint.**

### AUD-004 — `problems.md` périmé vs réalité
- **Zone :** registre de dérive. **Sévérité : `warn`. Statut : ouvert.**
- **Preuve :** `problems.md` affirme (revue `2026-05-17`) « pas de flag
  documenté inexistant », « cohérents avec `work.go`/`root.go` », tous GAP
  clôturés. Or la dérive AUD-001/002 a été introduite le `2026-05-31` (après
  cette revue). Le registre ne reflète plus l'état réel.

### AUD-005 — Routing : noms d'agents codés en dur hors `config.agents`
- **Zone :** orchestration (`internal/routing/router.go`). **Sévérité : `warn`.
  Statut : ouvert.**
- **Preuve :** `Route()` pose `Agent: "cursor"` en valeur initiale, force
  `d.Agent = "claude"` dans la branche `cloud_heavy`, et retombe sur `"ollama"`
  si `DefaultEnricher` est vide. Ces noms ne sont **pas vérifiés** contre
  `config.agents`. Si l'agent n'est pas déclaré, la décision pointe vers un
  backend inexistant (pas d'erreur guidée, pas de fallback contrôlé).
- **Lien produit :** c'est précisément la « surcouche IA » opaque que tu
  décris — le choix d'agent n'est ni piloté par la config ni expliqué.

### AUD-006 — Couplage policy/spec historique non vérifié
- **Zone :** orchestration (`internal/policy/ollama.go`). **Sévérité : `info`.
  Statut : ouvert.**
- **Preuve :** les rôles autorisés/interdits référencent « spec §10.1/§10.2 »
  (specv3 historique). Aucune vérification automatique ne garantit que cette
  liste reste alignée avec le canon ou la config. Risque de dérive silencieuse.

### AUD-007 — Surface CLI très large (~50 commandes racine)
- **Zone :** UX / orchestration. **Sévérité : `info`. Statut : ouvert.**
- **Preuve :** `asa --help` liste ~50 commandes de premier niveau
  (`spec`, `plan`, `dev`, `verify`, `review`, `work`, `graph`, `trust`,
  `knowledge`, `replay`, `coordination`, `runtime`, `mission`, `dashboard`,
  `runs`, `agents`, `prototype`, `flows`, `contracts`, …).
- **Lien produit :** confirme le ressenti « on a perdu la simplicité ». La
  valeur unitaire (lancer chaque étape) est là, mais le **chemin guidé** (« je
  décris un besoin → l'outil enchaîne ») est noyé dans la surface.

## 3. Ce qui est sain (non-findings, vérifiés)

- **Build/vet/test :** tout vert, suite complète exécutée.
- **Secrets :** scan committé propre (les occurrences `token`/`secret` sont du
  code légitime : comptage de tokens, `NotionToken()` lu depuis l'env).
- **Locales :** `en` = 197 pages ; `fr`/`de`/`es` = 95 chacune. Le différentiel
  (113) est exactement la référence CLI générée `en`-only (fallback assumé). Le
  contenu **non généré** est identique en chemins sur les 4 locales (`diff` exit
  0 pour fr, de, es). Aucune page orpheline.
- **Docgen déterministe :** deux régénérations successives produisent des
  fichiers identiques (le non-déterminisme n'est pas en cause ; c'est bien la
  non-régénération qui crée la dérive).

## 4. Lecture « simplicité perdue » (axe produit)

L'audit confirme objectivement ton intuition, sur deux plans :

1. **Abstraction d'orchestration opaque** (AUD-005/006) : le routage Kiro /
   Ollama / Cursor / Claude est fonctionnel mais codé en dur par endroits, non
   validé contre la config, et la « raison » du choix n'est pas exposée comme un
   contrat stable. → cible : une décision de routing **pure, pilotée par la
   config, explicable**, et un modèle mental unique « décrire → produire →
   valider ».
2. **Surface trop large vs chemin guidé** (AUD-007) : 50 commandes unitaires
   excellentes pour le script/CI, mais pas de « voie royale » mise en avant pour
   l'utilisateur qui veut juste décrire un besoin et être accompagné aux jalons.
   → cible : préserver l'unitaire **et** rendre le workflow guidé évident
   (onboarding → `work` → jalons de validation), avec remédiation guidée si un
   outil/config manque.

## 5. Recommandations de correction (pré-spec)

| ID | Action | Effort | Bloquant OSS ? |
|----|--------|--------|----------------|
| AUD-001/002 | Régénérer la doc CLI (`asa docs generate-cli`) + ajouter `runs.mdx` ; ajouter un garde-fou (test/CI « régénération sans diff ») | Faible | Oui (dérive doc) |
| AUD-003 | Aligner golangci-lint sur go1.25 (mise à jour binaire / pin version dans CI) ou ajuster la cible ; documenter la commande exacte | Faible | Oui (gate lint) |
| AUD-004 | Mettre `problems.md` à jour (tracer AUD-*), en faire le Remediation_Register de cette tranche | Faible | Non |
| AUD-005 | Router : valider l'agent contre `config.agents`, supprimer les noms codés en dur, erreur guidée si backend absent, exposer la raison | Moyen | Non (mais qualité) |
| AUD-006 | Documenter/relier la policy Ollama au canon courant ; test de cohérence | Faible | Non |
| AUD-007 | UX : mettre en avant le chemin guidé (onboarding → work → jalons) sans retirer l'unitaire ; clarifier le help racine | Moyen | Non (valeur produit) |

## 6. Mapping audit → exigences à (ré)écrire

La spec de correction couvrira : régénération + garde-fou doc CLI (AUD-001/002),
gate lint réparé (AUD-003), registre de dérive à jour (AUD-004), routing
config-driven et explicable (AUD-005/006), et restauration de la simplicité du
chemin guidé tout en préservant l'unitaire (AUD-007). La `requirements.md`
actuelle (qui décrivait un sous-système « Audit_Engine » à construire) sera
**réécrite** pour partir de ces findings réels, sans ajouter de surcouche.
