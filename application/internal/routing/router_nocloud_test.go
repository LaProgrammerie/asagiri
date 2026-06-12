package routing

// Feature: audit-coherence-consolidation, Property 7
//
// Property 7 (design P7) — Précédence no_cloud:
// For any input where noCloud is requested, Route (when it succeeds) produces a
// decision with Local == true and Reason == "no_cloud", regardless of allowCloud
// and preferLocal.
//
// Validates: Requirements 4.4

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// noCloudAgentSpec is a generated Agent_Backend declaration: a name and whether
// it is local (declared with a localhost endpoint) or cloud (declared with a
// command).
type noCloudAgentSpec struct {
	name  string
	local bool
}

// noCloudCase is the generated input space for Property 7. noCloud is always
// true (that is the property under test); allowCloud and preferLocal vary
// freely, as do the step class and the declared backends. The generator always
// declares at least one local backend so that no_cloud routing succeeds, which
// keeps the property non-vacuous.
type noCloudCase struct {
	agents             []noCloudAgentSpec
	defaultEnricher    string
	defaultAgent       string
	preferLocal        bool
	allowCloud         bool
	stepClass          string
	preferLocalForStep bool
}

// Generate implements quick.Generator with a smart generator constrained to the
// input space: it always declares one local backend, optionally adds cloud and
// extra local backends, and varies the orthogonal flags and step class.
func (noCloudCase) Generate(r *rand.Rand, _ int) reflect.Value {
	localPool := []string{"ollama", "localllm", "lmstudio"}
	cloudPool := []string{"cursor", "claude", "codex", "gemini"}
	classPool := []string{"spec", "plan", "enrich", "dev", "verify", "review", "", "custom"}

	tc := noCloudCase{}

	// Always declare at least one local backend so no_cloud routing succeeds.
	primaryLocal := localPool[r.Intn(len(localPool))]
	tc.agents = append(tc.agents, noCloudAgentSpec{name: primaryLocal, local: true})

	// Optionally add 0-2 cloud backends.
	for n := r.Intn(3); n > 0; n-- {
		tc.agents = append(tc.agents, noCloudAgentSpec{name: cloudPool[r.Intn(len(cloudPool))], local: false})
	}
	// Optionally add a second local backend.
	if r.Intn(2) == 0 {
		tc.agents = append(tc.agents, noCloudAgentSpec{name: localPool[r.Intn(len(localPool))], local: true})
	}

	// DefaultEnricher: empty, a declared backend, or an undeclared name (which
	// must still fall back to a declared local backend).
	switch r.Intn(3) {
	case 0:
		tc.defaultEnricher = ""
	case 1:
		tc.defaultEnricher = primaryLocal
	default:
		tc.defaultEnricher = "undeclared-enricher"
	}
	// DefaultAgent: empty or a cloud name (irrelevant to the no_cloud branch).
	if r.Intn(2) == 0 {
		tc.defaultAgent = cloudPool[r.Intn(len(cloudPool))]
	}

	tc.preferLocal = r.Intn(2) == 0
	tc.allowCloud = r.Intn(2) == 0
	tc.stepClass = classPool[r.Intn(len(classPool))]
	tc.preferLocalForStep = r.Intn(2) == 0

	return reflect.ValueOf(tc)
}

// build assembles a config.Config from the generated case. It always declares
// the generated backends and wires a strategy that may list the step class in
// prefer_local_for, to confirm no_cloud still prevails over prefer_local.
func (tc noCloudCase) build() *config.Config {
	cfg := config.NewTestConfig("t")
	cfg.Agents = map[string]config.Agent{}
	for _, a := range tc.agents {
		if a.local {
			cfg.Agents[a.name] = config.Agent{Endpoint: "http://localhost:11434"}
		} else {
			cfg.Agents[a.name] = config.Agent{Command: a.name}
		}
	}
	cfg.Work.DefaultEnricher = tc.defaultEnricher
	cfg.Work.DefaultAgent = tc.defaultAgent

	strat := config.RoutingStrategy{}
	if tc.preferLocalForStep {
		strat.PreferLocalFor = []string{tc.stepClass}
	}
	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{"cost_aware": strat}
	return cfg
}

// TestRouteNoCloudPrecedenceProperty asserts Property 7: whenever noCloud is
// requested and a local backend is declared, Route succeeds with Local == true
// and Reason == "no_cloud", whatever allowCloud and preferLocal are.
func TestRouteNoCloudPrecedenceProperty(t *testing.T) {
	property := func(tc noCloudCase) bool {
		cfg := tc.build()
		// noCloud is fixed true; allowCloud and preferLocal vary per case.
		d, err := Route(cfg, tc.stepClass, tc.preferLocal, true, tc.allowCloud)
		if err != nil {
			// A local backend is always declared, so no_cloud must succeed.
			return false
		}
		return d.Local && d.Reason == "no_cloud"
	}

	cfg := &quick.Config{
		MaxCount: 200, // >= 100 iterations
		Rand:     rand.New(rand.NewSource(7)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 7 (précédence no_cloud) violated: %v", err)
	}
}
