package checks

const (
	typeBlastRadius           = "blast-radius"
	typeObservability         = "observability"
	typeSecurity              = "security"
	typeCost                  = "cost"
	typeAnalytics             = "analytics"
	typeArchitecture          = "architecture"
	typePerformance           = "performance"
	typeBackwardCompatibility = "backward-compatibility"
	typeMigrationSafety       = "migration-safety"
	typePermissions           = "permissions"
)

// Lot3RunnerTypes is the stable pipeline order after lot-2 base checks (spec §8, handoff lot 3).
var Lot3RunnerTypes = []string{
	typePermissions,
	typeObservability,
	typeAnalytics,
	typeArchitecture,
	typeSecurity,
	typePerformance,
	typeCost,
	typeBackwardCompatibility,
	typeMigrationSafety,
	typeBlastRadius,
}

// AllRegisteredCheckTypes returns lot-2 + lot-3 check type ids in pipeline order.
func AllRegisteredCheckTypes() []string {
	out := make([]string, 0, 4+len(Lot3RunnerTypes))
	out = append(out, typeStaticAnalysis, typeContracts, typeFlows, typeKnowledgeGraph)
	out = append(out, Lot3RunnerTypes...)
	out = append(out, typeTests)
	return out
}
