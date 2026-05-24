#!/usr/bin/env bash
# Dry-run benchmark: mesure temps CLI estimate/work plan-only sans agents réels.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
BIN="${BIN:-$ROOT/bin/agentflow}"
if [[ ! -x "$BIN" ]]; then
  make build
  BIN="$ROOT/bin/agentflow"
fi
export AGENTFLOW_DRY_RUN=1
FEATURE="${FEATURE:-agentflow-test}"
echo "=== AgentFlow benchmark (dry-run) ==="
echo "feature=$FEATURE bin=$BIN"
start=$(date +%s%N)
"$BIN" estimate "$FEATURE" 2>&1 | tail -20
mid=$(date +%s%N)
"$BIN" work "développe $FEATURE" --dry-run --plan-only --yes 2>&1 | tail -15
end=$(date +%s%N)
est_ms=$(( (mid - start) / 1000000 ))
work_ms=$(( (end - mid) / 1000000 ))
total_ms=$(( (end - start) / 1000000 ))
echo "---"
echo "estimate_ms=$est_ms work_plan_only_ms=$work_ms total_ms=$total_ms"
echo "Note: métriques tokens/coût affichées par estimate ; pas d'appel cloud en dry-run."
