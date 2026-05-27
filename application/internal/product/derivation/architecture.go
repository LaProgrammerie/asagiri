package derivation

import "sort"

type FlowInput struct {
	ID                string
	BusinessObjective string
	Metrics           []string
	StepActions       []StepAction
}

type StepAction struct {
	StepID      string
	Action      string
	ContractRef string
	Sensitive   bool
	Errors      []string
}

type Projection struct {
	API             []string `yaml:"api"`
	Async           []string `yaml:"async"`
	Security        []string `yaml:"security"`
	Observability   []string `yaml:"observability"`
	Analytics       []string `yaml:"analytics"`
	Infrastructure  []string `yaml:"infrastructure"`
	Permissions     []string `yaml:"permissions"`
	MetricsCoverage []string `yaml:"metrics_coverage"`
}

func DeriveArchitecture(flows []FlowInput) Projection {
	p := Projection{
		API:             uniqueSorted(deriveAPIRequirements(flows)),
		Async:           uniqueSorted(deriveInfraRequirements(flows)),
		Security:        uniqueSorted(derivePermissionsRequirements(flows)),
		Observability:   uniqueSorted(deriveObservabilityRequirements(flows)),
		Analytics:       uniqueSorted(deriveAnalyticsRequirements(flows)),
		Infrastructure:  uniqueSorted(deriveInfrastructureRequirements(flows)),
		Permissions:     uniqueSorted(derivePermissionsMatrix(flows)),
		MetricsCoverage: uniqueSorted(deriveMetricsCoverage(flows)),
	}
	return p
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
