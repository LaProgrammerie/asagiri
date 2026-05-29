package checks

// ToRegistryCheck maps a CheckResult to the registry Check shape.
func ToRegistryCheck(r CheckResult) Check {
	return Check{
		ID:          r.ID,
		Name:        r.Name,
		Type:        r.Type,
		Status:      r.Status,
		Confidence:  r.Confidence,
		Findings:    r.Findings,
		Evidence:    r.Evidence,
		Duration:    r.Duration,
		BlastRadius: r.BlastRadius,
	}
}
