package checks

import (
	"context"
	"fmt"
	"time"
)

// Check is the registry output shape (mapped to trust.VerificationCheck by the engine).
type Check struct {
	ID          string
	Name        string
	Type        string
	Status      string
	Confidence  float64
	Findings    []Finding
	Evidence    []Evidence
	Duration    time.Duration
	BlastRadius *BlastRadiusSummary
}

// Registry runs verification check runners in pipeline order (spec §4, lot 2).
type Registry struct {
	runners []Runner
	deps    Dependencies
}

// NewRegistry returns an empty check registry.
func NewRegistry() *Registry {
	return &Registry{deps: DefaultDependencies()}
}

// NewDefaultRegistry registers lot-2 and lot-3 runners in pipeline order.
func NewDefaultRegistry(deps Dependencies) *Registry {
	if deps.Investigate == nil {
		deps = DefaultDependencies()
	}
	runners := []Runner{
		StaticAnalysisRunner{},
		ContractsRunner{},
		FlowsRunner{},
		KnowledgeGraphRunner{},
		PermissionsRunner{},
		ObservabilityRunner{},
		AnalyticsRunner{},
		ArchitectureRunner{},
		SecurityRunner{},
		PerformanceRunner{},
		CostRunner{},
		BackwardCompatibilityRunner{},
		MigrationSafetyRunner{},
		BlastRadiusRunner{},
		TestsRunner{},
	}
	return &Registry{deps: deps, runners: runners}
}

// Runners returns registered runners (for engine wiring).
func (r *Registry) Runners() []Runner {
	return r.runners
}

// RunSelected runs only check types listed in types (empty → RunAll).
func (r *Registry) RunSelected(ctx context.Context, scope Scope, types []string) ([]Check, error) {
	if len(types) == 0 {
		return r.RunAll(ctx, scope)
	}
	want := make(map[string]struct{}, len(types))
	for _, t := range types {
		want[t] = struct{}{}
	}
	if len(r.runners) == 0 {
		return []Check{}, nil
	}
	out := make([]Check, 0, len(types))
	for _, runner := range r.runners {
		if _, ok := want[runner.Type()]; !ok {
			continue
		}
		result, err := runner.Run(ctx, scope, r.deps)
		if err != nil {
			out = append(out, ToRegistryCheck(CheckResult{
				ID:     checkID(runner.Type(), scope.TrustID),
				Name:   runner.Type(),
				Type:   runner.Type(),
				Status: statusFailed,
				Findings: []Finding{{
					Severity: "error",
					Category: "runner",
					Message:  fmt.Sprintf("runner error: %v", err),
				}},
			}))
			continue
		}
		out = append(out, ToRegistryCheck(result))
	}
	return out, nil
}

// RunAll executes registered checks sequentially; runner errors become failed checks.
func (r *Registry) RunAll(ctx context.Context, scope Scope) ([]Check, error) {
	if len(r.runners) == 0 {
		return []Check{}, nil
	}
	out := make([]Check, 0, len(r.runners))
	for _, runner := range r.runners {
		result, err := runner.Run(ctx, scope, r.deps)
		if err != nil {
			out = append(out, ToRegistryCheck(CheckResult{
				ID:     checkID(runner.Type(), scope.TrustID),
				Name:   runner.Type(),
				Type:   runner.Type(),
				Status: statusFailed,
				Findings: []Finding{{
					Severity: "error",
					Category: "runner",
					Message:  fmt.Sprintf("runner error: %v", err),
				}},
			}))
			continue
		}
		out = append(out, ToRegistryCheck(result))
	}
	return out, nil
}
