package routing

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Feature: audit-coherence-consolidation, Property 5
// Design property P5: Backend déclaré et raison valide.
//
// Pour toute entrée où Route réussit, Decision.Agent est une clé de cfg.Agents
// et Decision.Reason appartient à l'ensemble fermé des cinq raisons.
//
// Validates: Requirements 4.2, 4.6

// p5ValidReasons is the closed set of reasons Route may expose (Requirement 4.6).
var p5ValidReasons = map[string]bool{
	"prefer_local": true,
	"no_cloud":     true,
	"cloud_heavy":  true,
	"cloud_fast":   true,
	"default":      true,
}

// p5Input bundles a generated Route call: a config plus the step class and flag
// combination. It implements quick.Generator so testing/quick can drive it with
// a deterministic, seeded source.
type p5Input struct {
	cfg         *config.Config
	stepClass   string
	preferLocal bool
	noCloud     bool
	allowCloud  bool
}

// p5AgentPool is the pool of candidate Agent_Backend names used by the
// generator. Some are declared as local (localhost endpoint), some as cloud.
var p5AgentPool = []string{"cursor", "claude", "ollama", "kiro", "codex", "local-llm"}

// p5ClassPool is the pool of step classes used both for strategy lists and for
// the queried step class.
var p5ClassPool = []string{"spec", "plan", "enrich", "dev", "verify", "review", "implement", "analyze"}

// Generate builds a deterministic p5Input from the provided rand source so the
// property explores varied declared-agent sets, strategies, step classes and
// flag combinations without observable side effects (Requirement 4.1 holds for
// Route; here we only need reproducible inputs).
func (p5Input) Generate(r *rand.Rand, _ int) reflect.Value {
	cfg := config.NewTestConfig("p5")
	cfg.Agents = map[string]config.Agent{}

	for _, name := range p5AgentPool {
		if r.Intn(3) == 0 {
			continue // sometimes leave the agent undeclared
		}
		if r.Intn(2) == 0 {
			cfg.Agents[name] = config.Agent{Endpoint: "http://localhost:11434"} // local
		} else {
			cfg.Agents[name] = config.Agent{Command: name} // cloud
		}
	}

	declared := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		declared = append(declared, name)
	}
	sort.Strings(declared)

	// pick returns a declared agent, an undeclared name, or "" so the generator
	// also exercises the guided-error path (which yields err != nil; the
	// property is then vacuously satisfied).
	pick := func() string {
		switch r.Intn(4) {
		case 0:
			return ""
		case 1:
			return "undeclared-agent"
		default:
			if len(declared) == 0 {
				return ""
			}
			return declared[r.Intn(len(declared))]
		}
	}

	cfg.Work.DefaultAgent = pick()
	cfg.Work.DefaultEnricher = pick()

	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {
			PreferLocalFor:   p5RandSubset(r, p5ClassPool),
			UseCloudHeavyFor: p5RandSubset(r, p5ClassPool),
			UseCloudFastFor:  p5RandSubset(r, p5ClassPool),
		},
	}

	stepClass := ""
	if r.Intn(8) != 0 { // mostly a known class, occasionally empty/unknown
		stepClass = p5ClassPool[r.Intn(len(p5ClassPool))]
	}

	in := p5Input{
		cfg:         cfg,
		stepClass:   stepClass,
		preferLocal: r.Intn(2) == 0,
		noCloud:     r.Intn(2) == 0,
		allowCloud:  r.Intn(2) == 0,
	}
	return reflect.ValueOf(in)
}

// p5RandSubset returns a deterministic random subset of items.
func p5RandSubset(r *rand.Rand, items []string) []string {
	var out []string
	for _, item := range items {
		if r.Intn(2) == 0 {
			out = append(out, item)
		}
	}
	return out
}

func TestRouterBackendDeclaredAndReasonValid(t *testing.T) {
	// Feature: audit-coherence-consolidation, Property 5 (design P5).
	// Validates: Requirements 4.2, 4.6
	prop := func(in p5Input) bool {
		d, err := Route(in.cfg, in.stepClass, in.preferLocal, in.noCloud, in.allowCloud)
		if err != nil {
			// The property only constrains successful decisions; the guided
			// error path is covered by Property 9.
			return true
		}
		if _, ok := in.cfg.Agents[d.Agent]; !ok {
			t.Errorf("Decision.Agent %q is not a declared key of cfg.Agents %v", d.Agent, p5SortedKeys(in.cfg))
			return false
		}
		if !p5ValidReasons[d.Reason] {
			t.Errorf("Decision.Reason %q is not in the closed reason set", d.Reason)
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200, // >= 100 iterations
		Rand:     rand.New(rand.NewSource(20240605)),
	}
	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("Property 5 (backend déclaré et raison valide) failed: %v", err)
	}
}

// p5SortedKeys returns the declared agent names sorted, for stable diagnostics.
func p5SortedKeys(cfg *config.Config) []string {
	keys := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
