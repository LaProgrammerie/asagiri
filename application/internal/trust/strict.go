package trust

import (
	"context"
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// StrictTrustError is returned when post-work strict trust verification fails.
type StrictTrustError struct {
	Report TrustReport
	Reason string
}

func (e *StrictTrustError) Error() string {
	if e.Reason != "" {
		return e.Reason
	}
	return "strict trust verification failed"
}

// RunStrictTrust runs verify trust + gates with strict CI semantics (spec §22).
func RunStrictTrust(ctx context.Context, eng *Engine, flow, branch, product string) (VerificationResult, error) {
	if flow == "" {
		return VerificationResult{}, fmt.Errorf("strict-trust: flow id required")
	}
	result, err := eng.Verify(ctx, VerificationRequest{
		Flow:    flow,
		Branch:  branch,
		Product: product,
		Strict:  true,
	})
	if err != nil {
		return VerificationResult{}, err
	}
	if reason, below := strictConfidenceBelowFloor(result.Report, eng.Gates); below {
		return result, &StrictTrustError{Report: result.Report, Reason: reason}
	}
	if CIShouldFail(result.Report, true) {
		reason := result.Report.Gate.Reason
		if reason == "" {
			reason = "trust checks or gates failed under --strict-trust"
		}
		return result, &StrictTrustError{Report: result.Report, Reason: reason}
	}
	return result, nil
}

// NewEngineForStrict wires engine with gates from config.
func NewEngineForStrict(repoRoot string, cfg *config.Config) *Engine {
	eng := NewEngine(repoRoot)
	if cfg != nil {
		eng.Gates = NewGateEvaluator(&cfg.Verification)
	}
	return eng
}

// strictConfidenceBelowFloor enforces LowConfidenceThreshold when gates are not configured (spec §22).
func strictConfidenceBelowFloor(report TrustReport, gates GateEvaluator) (reason string, below bool) {
	if gates.Configured() || report.Gate.Status != GateStatusNotConfigured {
		return "", false
	}
	overall := report.Confidence.Overall
	if overall >= LowConfidenceThreshold {
		return "", false
	}
	return fmt.Sprintf(
		"overall confidence %.0f%% below strict floor %.0f%% (verification gates not configured)",
		overall*100, LowConfidenceThreshold*100,
	), true
}
