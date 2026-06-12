package gates

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const TrustGateName = "trust"

// TrustGateSatisfied reports whether the last trust gate entry clears review.
func TrustGateSatisfied(cfg *config.Config, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.Trust.IsActive() {
		return true
	}
	entry, ok := LastGateEntry(payloadJSON, TrustGateName)
	if !ok {
		return false
	}
	return GateEntrySatisfied(cfg.Work.Gates.Trust.WarnAdvisory(), entry)
}

// TrustGateBlocksReview is true when review must not run without a satisfied trust gate.
func TrustGateBlocksReview(cfg *config.Config, taskStatus, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.Trust.IsActive() {
		return false
	}
	if TrustGateSatisfied(cfg, payloadJSON) {
		return false
	}
	st := strings.ToLower(strings.TrimSpace(taskStatus))
	switch st {
	case asagiri.StatusVerified, "":
		return true
	default:
		return false
	}
}
