package routing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Feature: audit-coherence-consolidation, Property 4
//
// Design property P4 — Déterminisme et pureté du routing.
// *For any* configuration, classe d'étape et combinaison de flags, deux appels à
// routing.Route produisent la même issue (succès ou erreur) et, en cas de succès,
// une Decision identique sur tous les champs, sans effet de bord observable.
//
// Validates: Requirements 4.1

// routingInput bundles the generated arguments fed to routing.Route. It
// implements quick.Generator so testing/quick can produce constrained, valid
// inputs that span both the success and the ErrNoDeclaredBackend error paths.
type routingInput struct {
	cfg         *config.Config
	stepClass   string
	preferLocal bool
	noCloud     bool
	allowCloud  bool
}

// agentNamePool is a fixed set of candidate Agent_Backend names. Generation only
// ever picks from this pool so the input space stays bounded and deterministic.
var agentNamePool = []string{"cursor", "claude", "ollama", "kiro", "local-llm", "gemini"}

// stepClassPool covers declared classes, the empty class and an unknown class so
// every branch of Route (no_cloud / prefer_local / cloud_heavy / cloud_fast /
// default) and the guided-error path are all reachable.
var stepClassPool = []string{"spec", "plan", "enrich", "dev", "verify", "review", "implement", "", "unknown-class"}

// Generate builds a randomized but well-formed routing input. Because the only
// source of randomness is the *rand.Rand provided by testing/quick (seeded
// deterministically in the test below), regenerating with the same seed yields
// the same inputs — keeping the generator deterministic.
func (routingInput) Generate(r *rand.Rand, _ int) reflect.Value {
	cfg := config.NewTestConfig("t")
	cfg.Agents = map[string]config.Agent{}

	// Declare a random subset of agents; each is randomly local (localhost
	// endpoint) or cloud (command only), with a random model id.
	for _, name := range agentNamePool {
		if r.Intn(2) == 0 {
			continue
		}
		if r.Intn(2) == 0 {
			cfg.Agents[name] = config.Agent{Endpoint: "http://localhost:11434", Model: name + "-model"}
		} else {
			cfg.Agents[name] = config.Agent{Command: name, DefaultModel: name + "-model"}
		}
	}

	// Work defaults may reference a declared agent, an undeclared one, or be
	// empty — exercising both the success and the guided-error paths.
	cfg.Work.DefaultAgent = pickWorkRef(r)
	cfg.Work.DefaultEnricher = pickWorkRef(r)

	// Routing strategy: usually "cost_aware", sometimes empty or missing.
	switch r.Intn(3) {
	case 0:
		cfg.Routing.DefaultStrategy = "cost_aware"
	case 1:
		cfg.Routing.DefaultStrategy = ""
	default:
		cfg.Routing.DefaultStrategy = "missing-strategy"
	}
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {
			PreferLocalFor:   randClassSubset(r),
			UseCloudHeavyFor: randClassSubset(r),
			UseCloudFastFor:  randClassSubset(r),
		},
	}

	in := routingInput{
		cfg:         cfg,
		stepClass:   stepClassPool[r.Intn(len(stepClassPool))],
		preferLocal: r.Intn(2) == 0,
		noCloud:     r.Intn(2) == 0,
		allowCloud:  r.Intn(2) == 0,
	}
	return reflect.ValueOf(in)
}

// pickWorkRef returns a declared-ish agent name, an undeclared name, or "".
func pickWorkRef(r *rand.Rand) string {
	switch r.Intn(4) {
	case 0:
		return ""
	case 1:
		return "undeclared-agent"
	default:
		return agentNamePool[r.Intn(len(agentNamePool))]
	}
}

// randClassSubset returns a random subset of step classes for a strategy list.
func randClassSubset(r *rand.Rand) []string {
	var out []string
	for _, cls := range stepClassPool {
		if cls == "" {
			continue
		}
		if r.Intn(3) == 0 {
			out = append(out, cls)
		}
	}
	return out
}

func TestRouteDeterministicAndPure(t *testing.T) {
	property := func(in routingInput) bool {
		// Snapshot the config before any call. Route must not mutate it.
		before := fmt.Sprintf("%#v", in.cfg)

		d1, err1 := Route(in.cfg, in.stepClass, in.preferLocal, in.noCloud, in.allowCloud)
		d2, err2 := Route(in.cfg, in.stepClass, in.preferLocal, in.noCloud, in.allowCloud)

		// Same issue: both succeed or both fail identically.
		if (err1 == nil) != (err2 == nil) {
			return false
		}
		if err1 != nil {
			if err1.Error() != err2.Error() {
				return false
			}
			// On error, Route returns the zero Decision (no backend selected).
			if d1 != (Decision{}) || d2 != (Decision{}) {
				return false
			}
		} else if d1 != d2 {
			// On success, the Decision is identical on every field.
			return false
		}

		// No observable side effect: the config is unchanged after both calls.
		after := fmt.Sprintf("%#v", in.cfg)
		return before == after
	}

	// Deterministic generator: a fixed seed makes the whole run reproducible.
	// MaxCount = 300 satisfies the >= 100 iterations convention.
	cfg := &quick.Config{
		MaxCount: 300,
		Rand:     rand.New(rand.NewSource(20260531)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("routing is not deterministic/pure: %v", err)
	}
}
