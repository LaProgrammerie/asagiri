package trust

import (
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

// ComputeResidualRisk derives coarse risk from checks and confidence (lot 2).
func ComputeResidualRisk(checks []VerificationCheck, conf confidence.Report) ResidualRisk {
	if len(checks) == 0 {
		return ResidualRiskUnknown
	}
	for _, c := range checks {
		for _, f := range c.Findings {
			if f.Severity == SeverityCritical {
				return ResidualRiskHigh
			}
		}
		if c.Status == CheckStatusFailed {
			if conf.Overall < 0.7 {
				return ResidualRiskHigh
			}
		}
	}
	if conf.Overall < 0.5 {
		return ResidualRiskHigh
	}
	warnings := countWarnings(checks)
	if conf.Overall < 0.7 || warnings >= 2 || contractsCheckWarned(checks) {
		return ResidualRiskMedium
	}
	if conf.Overall >= 0.85 {
		failed := false
		for _, c := range checks {
			if c.Status == CheckStatusFailed {
				failed = true
				break
			}
		}
		if !failed {
			return ResidualRiskLow
		}
	}
	return ResidualRiskMedium
}

func countWarnings(checks []VerificationCheck) int {
	n := 0
	for _, c := range checks {
		for _, f := range c.Findings {
			if f.Severity == SeverityWarning {
				n++
			}
		}
		if c.Status == CheckStatusWarn {
			n++
		}
	}
	return n
}

func contractsCheckWarned(checks []VerificationCheck) bool {
	for _, c := range checks {
		if c.Type == CheckContracts && c.Status == CheckStatusWarn {
			return true
		}
	}
	return false
}
