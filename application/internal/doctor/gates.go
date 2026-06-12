package doctor

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func collectGates(cfg *config.Config) []GateInfo {
	if cfg == nil {
		return nil
	}
	g := cfg.Work.Gates
	return []GateInfo{
		summarizePlanGate(cfg, g.Plan),
		summarizeModeGate("governance", g.Governance.Enabled, g.Governance.Mode, g.Governance.IsActive(), g.Governance.EnabledButInactive()),
		summarizeModeGate("enrich", g.Enrich.Enabled, g.Enrich.Mode, g.Enrich.IsActive(), g.Enrich.EnabledButInactive()),
		summarizeModeGate("human_review", g.HumanReview.Enabled, g.HumanReview.Mode, g.HumanReview.IsActive(), humanReviewInactive(g.HumanReview)),
		summarizeModeGate("verify_evidence", g.VerifyEvidence.Enabled, g.VerifyEvidence.Mode, g.VerifyEvidence.IsActive(), g.VerifyEvidence.EnabledButInactive()),
		summarizeTrustGate(g.Trust),
	}
}

func summarizePlanGate(cfg *config.Config, g config.WorkPlanGateConfig) GateInfo {
	status := "disabled"
	if g.Enabled {
		status = "active"
	}
	return GateInfo{
		Name:    "plan",
		Enabled: g.Enabled,
		Status:  status,
		Detail:  planGateDetail(cfg, g),
	}
}

func planGateDetail(cfg *config.Config, g config.WorkPlanGateConfig) string {
	if !g.Enabled {
		return "gate désactivée"
	}
	if cfg != nil {
		if a := strings.TrimSpace(cfg.PlanGateAgent()); a != "" {
			return "agent " + a
		}
	}
	return "active après plan"
}

func summarizeModeGate(name string, enabled bool, mode string, active, inactive bool) GateInfo {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = config.GovernanceModeOff
	}
	info := GateInfo{Name: name, Enabled: enabled, Mode: mode}
	switch {
	case !enabled:
		info.Status = "disabled"
		info.Detail = "gate désactivée"
	case active:
		info.Status = "active"
		info.Detail = "mode " + mode
	case inactive:
		info.Status = "invalid_mode"
		info.Detail = fmt.Sprintf("enabled mais mode %q invalide — utiliser off ou per-task", mode)
	default:
		info.Status = "inactive"
		info.Detail = "enabled, mode off — gate inactive"
	}
	return info
}

func summarizeTrustGate(g config.WorkTrustGateConfig) GateInfo {
	info := summarizeModeGate("trust", g.Enabled, g.Mode, g.IsActive(), g.EnabledButInactive())
	if g.IsActive() {
		info.Detail = fmt.Sprintf("mode %s, min_score %.0f", strings.TrimSpace(g.Mode), g.MinScoreValue())
	}
	return info
}

func humanReviewInactive(g config.WorkHumanReviewGateConfig) bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, config.GovernanceModeOff)
}
