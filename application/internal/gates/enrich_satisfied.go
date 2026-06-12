package gates

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const EnrichGateName = "enrich"

// EnrichGateSatisfied reports whether the last enrich gate entry in payload clears dev.
// Gate inactive → always satisfied (no enrich gate requirement).
func EnrichGateSatisfied(cfg *config.Config, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.Enrich.IsActive() {
		return true
	}
	entry, ok := LastGateEntry(payloadJSON, EnrichGateName)
	if !ok {
		return false
	}
	return GateEntrySatisfied(cfg.Work.Gates.Enrich.WarnAdvisory(), entry)
}

// EnrichGateBlocksDev is true when dev must not run without a satisfied enrich gate (planned, pending, enriched).
func EnrichGateBlocksDev(cfg *config.Config, taskStatus, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.Enrich.IsActive() {
		return false
	}
	if EnrichGateSatisfied(cfg, payloadJSON) {
		return false
	}
	st := strings.ToLower(strings.TrimSpace(taskStatus))
	switch st {
	case asagiri.StatusPlanned, asagiri.StatusPending, asagiri.StatusEnriched, "":
		return true
	default:
		return false
	}
}
