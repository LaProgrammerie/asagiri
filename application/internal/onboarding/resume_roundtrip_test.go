package onboarding

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// Feature: audit-coherence-consolidation, Property 19
//
// Design property P19 — Round-trip save / resume de l'onboarding.
// *For any* état du wizard enregistré, la reprise via `--resume` (SaveState puis
// LoadState) restaure l'étape courante et l'ensemble des réponses déjà
// collectées à l'identique de l'état enregistré.
//
// Validates: Requirements 7.4

// p19StringPool is a small, bounded set of answer values. It deliberately
// includes the empty string so the generator exercises both populated and
// absent (omitempty) answer fields without exploding the input space.
var p19StringPool = []string{"", "asagiri", "main", "go", "kiro", "cursor", "codex", "ollama", "feature-x", "one liner", "devs"}

// p19WizardState wraps onboarding.State so it can implement quick.Generator.
// Helpers are prefixed p19 to avoid clashing with sibling onboarding tests
// (tasks 6.3 / 6.4) that share this package.
type p19WizardState struct {
	st State
}

// Generate builds a randomized but well-formed wizard state. The only source of
// randomness is the *rand.Rand supplied by testing/quick (seeded deterministically
// in the test below), so the generator stays reproducible.
//
// Constraints that keep the round-trip an exact identity:
//   - CurrentStep is always drawn from stepOrder, so it is never the empty value
//     that LoadState would rewrite to StepWelcome.
//   - Completed is either nil or a non-empty slice, never an empty non-nil slice,
//     because the `omitempty` JSON tag would otherwise turn `[]string{}` into nil
//     after a round-trip.
func (p19WizardState) Generate(r *rand.Rand, _ int) reflect.Value {
	st := State{
		CurrentStep: stepOrder[r.Intn(len(stepOrder))],
		Answers: Answers{
			ProjectName:      p19PickString(r),
			DefaultBranch:    p19PickString(r),
			Tagline:          p19PickString(r),
			Stack:            p19PickString(r),
			DefaultSpecAgent: p19PickString(r),
			DefaultAgent:     p19PickString(r),
			DefaultReviewer:  p19PickString(r),
			DefaultEnricher:  p19PickString(r),
			FeatureSlug:      p19PickString(r),
			ProductOneLiner:  p19PickString(r),
			ProductUsers:     p19PickString(r),
		},
		Completed: p19GenCompleted(r),
	}
	return reflect.ValueOf(p19WizardState{st: st})
}

// p19PickString returns a random value from the bounded answer pool.
func p19PickString(r *rand.Rand) string {
	return p19StringPool[r.Intn(len(p19StringPool))]
}

// p19GenCompleted returns nil or a non-empty slice of completed step names.
func p19GenCompleted(r *rand.Rand) []string {
	n := r.Intn(len(stepOrder) + 1) // 0..len → 0 yields nil
	if n == 0 {
		return nil
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, string(stepOrder[r.Intn(len(stepOrder))]))
	}
	return out
}

func TestSaveResumeRoundTrip(t *testing.T) {
	// Single temp repo root reused across iterations; SaveState overwrites the
	// state file each time, so iterations remain independent.
	repoRoot := t.TempDir()

	property := func(in p19WizardState) bool {
		if err := SaveState(repoRoot, in.st); err != nil {
			return false
		}
		loaded, err := LoadState(repoRoot)
		if err != nil {
			return false
		}
		// The resumed current step matches the saved one.
		if loaded.CurrentStep != in.st.CurrentStep {
			return false
		}
		// The full set of collected answers is restored identically.
		if !reflect.DeepEqual(loaded.Answers, in.st.Answers) {
			return false
		}
		// Whole-state identity (step + answers + completed) holds too.
		return reflect.DeepEqual(loaded, in.st)
	}

	// Deterministic generator (fixed seed) with MaxCount = 200, satisfying the
	// >= 100 iterations convention.
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(20260531)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("save/resume round-trip is not identity: %v", err)
	}
}
