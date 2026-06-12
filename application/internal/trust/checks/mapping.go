package checks

// ToRegistryCheck maps a CheckResult to the registry Check shape.
func ToRegistryCheck(r CheckResult) Check {
	return Check(r)
}
