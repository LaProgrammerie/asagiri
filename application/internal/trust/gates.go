package trust

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

// GateStatus is the outcome of gate evaluation (spec §19).
type GateStatus string

const (
	GateStatusPassed        GateStatus = "passed"
	GateStatusBlocked       GateStatus = "blocked"
	GateStatusNotConfigured GateStatus = "not_configured"
)

// GateEvaluation summarizes whether workflow gates allow progress.
type GateEvaluation struct {
	Status  GateStatus `json:"status"`
	Reason  string     `json:"reason,omitempty"`
	Profile string     `json:"profile,omitempty"`
}

// GateEvaluator applies configured verification gates.
type GateEvaluator struct {
	Profiles       map[string]config.GateProfile
	DefaultProfile string
	ActiveProfile  string
}

// Configured reports whether at least one gate profile is defined.
func (g GateEvaluator) Configured() bool {
	return len(g.Profiles) > 0
}

// NewGateEvaluator builds an evaluator from verification config (nil or empty gates → not configured).
func NewGateEvaluator(v *config.VerificationConfig) GateEvaluator {
	if v == nil || len(v.Gates) == 0 {
		return GateEvaluator{}
	}
	profiles := make(map[string]config.GateProfile, len(v.Gates))
	for name, p := range v.Gates {
		profiles[name] = p
	}
	return GateEvaluator{
		Profiles:       profiles,
		DefaultProfile: v.DefaultProfile,
	}
}

// Evaluate applies the active gate profile against confidence and check results.
func (g GateEvaluator) Evaluate(_ context.Context, conf confidence.Report, checks []VerificationCheck) GateEvaluation {
	profileName, profile, ok := g.resolveProfile()
	if !ok {
		return GateEvaluation{
			Status: GateStatusNotConfigured,
			Reason: "verification gates not configured",
		}
	}

	var reasons []string
	for dimKey, min := range profile.MinConfidence {
		score, known := dimensionScore(conf, dimKey)
		if !known {
			reasons = append(reasons, fmt.Sprintf("unknown dimension %q in gate config", dimKey))
			continue
		}
		if score < min {
			reasons = append(reasons, fmt.Sprintf(
				"%s confidence %.0f%% below required %.0f%%",
				dimKey, score*100, min*100,
			))
		}
	}

	byType := make(map[CheckType]VerificationCheck, len(checks))
	for _, c := range checks {
		byType[c.Type] = c
	}
	for _, req := range profile.RequiredChecks {
		ct := CheckType(req)
		c, ran := byType[ct]
		if !ran {
			reasons = append(reasons, fmt.Sprintf("required check %q was not executed", req))
			continue
		}
		if c.Status == CheckStatusFailed {
			reasons = append(reasons, fmt.Sprintf("required check %q failed", req))
		}
	}

	if len(reasons) > 0 {
		sort.Strings(reasons)
		return GateEvaluation{
			Status:  GateStatusBlocked,
			Reason:  strings.Join(reasons, "; "),
			Profile: profileName,
		}
	}
	return GateEvaluation{
		Status:  GateStatusPassed,
		Reason:  fmt.Sprintf("profile %q satisfied", profileName),
		Profile: profileName,
	}
}

func (g GateEvaluator) resolveProfile() (string, config.GateProfile, bool) {
	if len(g.Profiles) == 0 {
		return "", config.GateProfile{}, false
	}
	candidates := []string{}
	if g.ActiveProfile != "" {
		candidates = append(candidates, g.ActiveProfile)
	}
	if g.DefaultProfile != "" {
		candidates = append(candidates, g.DefaultProfile)
	}
	candidates = append(candidates, "production")
	for _, name := range candidates {
		if p, ok := g.Profiles[name]; ok {
			return name, p, true
		}
	}
	names := make([]string, 0, len(g.Profiles))
	for name := range g.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names[0], g.Profiles[names[0]], true
}

// ProfileNames returns configured gate profile names in stable order.
func (g GateEvaluator) ProfileNames() []string {
	if len(g.Profiles) == 0 {
		return nil
	}
	names := make([]string, 0, len(g.Profiles))
	for name := range g.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func dimensionScore(conf confidence.Report, key string) (float64, bool) {
	norm := strings.ToLower(strings.TrimSpace(key))
	norm = strings.ReplaceAll(norm, "-", "_")
	switch norm {
	case "architecture":
		return conf.Architecture, true
	case "implementation":
		return conf.Implementation, true
	case "flow_integrity", "flowintegrity":
		return conf.FlowIntegrity, true
	case "observability":
		return conf.Observability, true
	case "security":
		return conf.Security, true
	case "regression":
		return conf.Regression, true
	case "overall":
		return conf.Overall, true
	default:
		return 0, false
	}
}

// CIShouldFail reports whether CI mode should exit non-zero (spec §23).
func CIShouldFail(report TrustReport, strict bool) bool {
	if report.Gate.Status == GateStatusBlocked {
		return true
	}
	for _, c := range report.Checks {
		if c.Status == CheckStatusFailed {
			return true
		}
		if strict && c.Status == CheckStatusWarn {
			return true
		}
	}
	return false
}
