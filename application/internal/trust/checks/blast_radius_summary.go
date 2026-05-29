package checks

// BlastRadiusSummary is the structured blast radius block for trust reports (spec §12).
type BlastRadiusSummary struct {
	FlowsImpacted      int
	CriticalAPIs       int
	SharedModules      int
	MigrationRisk      string
	PublicContractRisk string
}
