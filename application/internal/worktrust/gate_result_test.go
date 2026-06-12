package worktrust

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/stretchr/testify/require"
)

func trustGateCfg() config.WorkTrustGateConfig {
	min := 70.0
	return config.WorkTrustGateConfig{
		Enabled:       true,
		Mode:          config.GovernanceModePerTask,
		MinScore:      &min,
		BlockVerdicts: config.DefaultTrustGateBlockVerdicts(),
		WarnVerdicts:  config.DefaultTrustGateWarnVerdicts(),
	}
}

func TestWorkTrustReportToGateResult_Pass(t *testing.T) {
	cfg := trustGateCfg()
	report := WorkTrustReport{
		Scope: TrustScope{Kind: "task", ID: "task-1", TaskID: "task-1"},
		Score: WorkTrustScore{Overall: 85, Verdict: VerdictTrusted, Summary: "trusted"},
	}
	result := WorkTrustReportToGateResult(report, cfg)
	require.Equal(t, gates.VerdictPass, result.Status)
	require.Equal(t, "trust_gate", result.GateID)
	require.Equal(t, "task-1", result.Scope)
}

func TestWorkTrustReportToGateResult_BlockedVerdict(t *testing.T) {
	cfg := trustGateCfg()
	report := WorkTrustReport{
		Scope: TrustScope{TaskID: "task-2"},
		Score: WorkTrustScore{Overall: 90, Verdict: VerdictBlocked},
	}
	result := WorkTrustReportToGateResult(report, cfg)
	require.Equal(t, gates.VerdictFail, result.Status)
}

func TestWorkTrustReportToGateResult_RiskyWarn(t *testing.T) {
	cfg := trustGateCfg()
	report := WorkTrustReport{
		Scope: TrustScope{TaskID: "task-3"},
		Score: WorkTrustScore{Overall: 75, Verdict: VerdictRisky},
	}
	result := WorkTrustReportToGateResult(report, cfg)
	require.Equal(t, gates.VerdictWarn, result.Status)
}

func TestWorkTrustReportToGateResult_ScoreBelowMin(t *testing.T) {
	cfg := trustGateCfg()
	report := WorkTrustReport{
		Scope: TrustScope{TaskID: "task-4"},
		Score: WorkTrustScore{Overall: 65, Verdict: VerdictAcceptable},
	}
	result := WorkTrustReportToGateResult(report, cfg)
	require.Equal(t, gates.VerdictFail, result.Status)
	require.NotEmpty(t, result.Findings)
	require.Equal(t, "trust_score_below_min", result.Findings[len(result.Findings)-1].Code)
}

func TestWorkTrustReportToGateResult_MapsFindingsAndEvidence(t *testing.T) {
	cfg := trustGateCfg()
	report := WorkTrustReport{
		Scope: TrustScope{TaskID: "task-5"},
		Score: WorkTrustScore{Overall: 80, Verdict: VerdictTrusted},
		Findings: []WorkTrustFinding{{
			Code: "hr_pending", Severity: "high", Message: "human review pending",
		}},
		Evidences: []WorkTrustEvidence{{
			Kind: "gate", Ref: ".asagiri/logs/task-5/gates/human_review.json", Summary: "pending",
		}},
	}
	result := WorkTrustReportToGateResult(report, cfg)
	require.Len(t, result.Findings, 1)
	require.Equal(t, "fail", result.Findings[0].Severity)
	require.Len(t, result.Evidence, 1)
	require.Equal(t, "gate", result.Evidence[0].Kind)
}
