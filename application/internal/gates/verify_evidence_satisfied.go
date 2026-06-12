package gates

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const VerifyEvidenceGateName = "verify_evidence"

// VerifyEvidenceGateSatisfied reports whether the last verify_evidence gate entry clears review.
// Gate inactive → always satisfied (no verify evidence gate requirement).
func VerifyEvidenceGateSatisfied(cfg *config.Config, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.VerifyEvidence.IsActive() {
		return true
	}
	entry, ok := LastGateEntry(payloadJSON, VerifyEvidenceGateName)
	if !ok {
		return false
	}
	return GateEntrySatisfied(cfg.Work.Gates.VerifyEvidence.WarnAdvisory(), entry)
}

// VerifyEvidenceGateBlocksReview is true when review must not run without a satisfied verify_evidence gate.
func VerifyEvidenceGateBlocksReview(cfg *config.Config, taskStatus, payloadJSON string) bool {
	if cfg == nil || !cfg.Work.Gates.VerifyEvidence.IsActive() {
		return false
	}
	if VerifyEvidenceGateSatisfied(cfg, payloadJSON) {
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
