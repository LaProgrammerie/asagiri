package checks

import (
	"fmt"
	"time"
)

func failedLot3(scope Scope, start time.Time, typ, name, category string, err error) CheckResult {
	return CheckResult{
		ID:         checkID(typ, scope.TrustID),
		Name:       name,
		Type:       typ,
		Status:     statusFailed,
		Confidence: 0,
		Findings: []Finding{{
			Severity: "error",
			Category: category,
			Message:  err.Error(),
		}},
		Duration: time.Since(start),
	}
}

func finishLot3(scope Scope, start time.Time, typ, name string, findings []Finding, evidence []Evidence) CheckResult {
	status := statusFromFindings(findings)
	return CheckResult{
		ID:         checkID(typ, scope.TrustID),
		Name:       name,
		Type:       typ,
		Status:     status,
		Confidence: roundConfidence(confidenceFromStatus(status)),
		Findings:   findings,
		Evidence:   evidence,
		Duration:   time.Since(start),
	}
}

func finishLot3WithBlast(scope Scope, start time.Time, typ, name string, findings []Finding, evidence []Evidence, br *BlastRadiusSummary) CheckResult {
	r := finishLot3(scope, start, typ, name, findings, evidence)
	r.BlastRadius = br
	if br != nil {
		evidence = append(evidence, Evidence{
			Kind:    "blast-radius",
			Source:  scope.Flow,
			Summary: formatBlastEvidence(*br),
		})
		r.Evidence = evidence
	}
	return r
}

func formatBlastEvidence(br BlastRadiusSummary) string {
	return fmt.Sprintf("flows=%d apis=%d modules=%d migration=%s contract=%s",
		br.FlowsImpacted, br.CriticalAPIs, br.SharedModules, br.MigrationRisk, br.PublicContractRisk)
}
