package trust

import "github.com/LaProgrammerie/asagiri/application/internal/trust/checks"

// BlastRadiusReport is the trust report blast radius section (spec §12).
type BlastRadiusReport struct {
	FlowsImpacted      int    `json:"flows_impacted"`
	CriticalAPIs       int    `json:"critical_apis"`
	SharedModules      int    `json:"shared_modules"`
	MigrationRisk      string `json:"migration_risk"`
	PublicContractRisk string `json:"public_contract_risk"`
}

func blastRadiusFromChecks(checks []checks.Check) *BlastRadiusReport {
	for _, c := range checks {
		if c.Type != "blast-radius" || c.BlastRadius == nil {
			continue
		}
		br := c.BlastRadius
		return &BlastRadiusReport{
			FlowsImpacted:      br.FlowsImpacted,
			CriticalAPIs:       br.CriticalAPIs,
			SharedModules:      br.SharedModules,
			MigrationRisk:      br.MigrationRisk,
			PublicContractRisk: br.PublicContractRisk,
		}
	}
	return nil
}
