# Robustesse workflows agents

**Date :** 2026-05-17 · **Référence :** spec-postv123 §5

## Matrice de tests manuels / dry-run

| Scénario | Commande / mode | Attendu |
|----------|-----------------|--------|
| Kiro seul | `spec --agent kiro --dry-run` | Log dry-run, pas subprocess réel |
| Cursor seul | `dev --agent cursor --dry-run` | Worktree dry ou skip selon config |
| Ollama seul | `enrich --agent ollama` | Local endpoint |
| Hybride | `work "…" --prefer-local` | Steps enrich local ; dev cloud si config |
| Fallback | intent resolver + `use_ollama_fallback` | Résolution ou ambiguïté message |
| Gros repo | `investigate` + `estimate` | Temps borné ; candidats listés |
| Ambiguïté | `work "fix it"` non-interactif | Erreur ou feature explicite requise |
| Budget dépassé | `work --budget 0.01` | BLOCK ou CONFIRM |
| MCP off | `mcp serve` sans enable | Erreur config claire |
| Reprise | `continue --dry-run` | Plan repris depuis SQLite |

## Fixtures dry-run

- Package `cli/integration_test.go` — régression commandes racine.
- Feature test : `.kiro/specs/agentflow-test` ou équivalent local.
- **Gap :** pas de fixture multi-agent hybride automatisée → P1 tests golden plan JSON.

## Risques comportementaux

| Risque | Mitigation |
|--------|------------|
| Hallucination feature | Resolver seuil confiance + inbox |
| Commandes agent arbitraires | Allowlist primitives intent |
| Fuite secret dans rapport | redact + ne pas persister env |

## Non-déterminisme accepté

- IDs run horodatés ; estimation tokens heuristique (±20 % typique).
