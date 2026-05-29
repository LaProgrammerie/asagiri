package memory

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// DoctorCheck is one memory health diagnostic (spec-phase-finale PF-A-01).
type DoctorCheck struct {
	Name string
	Err  error
}

// Doctor runs memory embedder and store consistency checks.
func (e *Engine) Doctor(ctx context.Context) ([]DoctorCheck, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("memory: store required")
	}
	emb := e.embedder
	if emb == nil {
		emb = embedder.Current()
	}
	var checks []DoctorCheck
	checks = append(checks, checkOllama(ctx, emb))
	dimCheck, err := checkDimensions(ctx, e.store, emb)
	if err != nil {
		return checks, err
	}
	checks = append(checks, dimCheck)
	orphanCheck, err := checkOrphanEntries(e.store)
	if err != nil {
		return checks, err
	}
	checks = append(checks, orphanCheck)
	return checks, nil
}

func checkOllama(ctx context.Context, emb embedder.Embedder) DoctorCheck {
	if emb == nil || emb.Name() != "ollama" {
		return DoctorCheck{Name: "ollama", Err: nil}
	}
	o, ok := emb.(*embedder.OllamaEmbedder)
	if !ok {
		return DoctorCheck{Name: "ollama", Err: fmt.Errorf("configured embedder is ollama but implementation is %T", emb)}
	}
	return DoctorCheck{Name: "ollama", Err: o.Reachable(ctx)}
}

func checkDimensions(ctx context.Context, store *runtime.Store, emb embedder.Embedder) (DoctorCheck, error) {
	expected, err := expectedDimensions(ctx, emb)
	if err != nil {
		return DoctorCheck{Name: "dimensions", Err: err}, nil
	}
	entries, err := store.ListMemory("", 0)
	if err != nil {
		return DoctorCheck{}, err
	}
	var mismatched, missing int
	for _, ent := range entries {
		if strings.TrimSpace(ent.Summary) == "" {
			continue
		}
		vec := UnmarshalEmbedding(ent.EmbeddingJSON)
		if len(vec) == 0 {
			missing++
			continue
		}
		if expected > 0 && len(vec) != expected {
			mismatched++
		}
	}
	if missing == 0 && mismatched == 0 {
		return DoctorCheck{Name: "dimensions", Err: nil}, nil
	}
	var parts []string
	if mismatched > 0 {
		parts = append(parts, fmt.Sprintf("%d entrée(s) avec dimension %d attendue par l'embedder %q", mismatched, expected, emb.Name()))
	}
	if missing > 0 {
		parts = append(parts, fmt.Sprintf("%d entrée(s) sans embedding (lancer: asa memory reindex)", missing))
	}
	return DoctorCheck{
		Name: "dimensions",
		Err:  fmt.Errorf("%s", strings.Join(parts, "; ")),
	}, nil
}

func expectedDimensions(ctx context.Context, emb embedder.Embedder) (int, error) {
	if emb == nil {
		return 0, fmt.Errorf("embedder not configured")
	}
	if d := emb.Dimensions(); d > 0 {
		return d, nil
	}
	vec, err := emb.Embed(ctx, ".")
	if err != nil {
		return 0, fmt.Errorf("probe embed: %w", err)
	}
	if len(vec) == 0 {
		return 0, fmt.Errorf("probe embed returned empty vector")
	}
	return len(vec), nil
}

func checkOrphanEntries(store *runtime.Store) (DoctorCheck, error) {
	known, err := store.KnownFlowIDs()
	if err != nil {
		return DoctorCheck{}, err
	}
	entries, err := store.ListMemory("", 0)
	if err != nil {
		return DoctorCheck{}, err
	}
	var orphans []string
	seen := map[string]struct{}{}
	for _, ent := range entries {
		for _, flowID := range ent.LinkedFlows {
			flowID = strings.TrimSpace(flowID)
			if flowID == "" {
				continue
			}
			if _, ok := known[flowID]; ok {
				continue
			}
			key := ent.ID + "→" + flowID
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			orphans = append(orphans, fmt.Sprintf("%s (flow %q)", ent.ID, flowID))
		}
	}
	if len(orphans) == 0 {
		return DoctorCheck{Name: "orphans", Err: nil}, nil
	}
	const maxList = 5
	msg := fmt.Sprintf("%d entrée(s) liée(s) à un flow inconnu", len(orphans))
	if len(orphans) <= maxList {
		return DoctorCheck{Name: "orphans", Err: fmt.Errorf("%s: %s", msg, strings.Join(orphans, ", "))}, nil
	}
	return DoctorCheck{
		Name: "orphans",
		Err:  fmt.Errorf("%s (ex. %s, …)", msg, strings.Join(orphans[:maxList], ", ")),
	}, nil
}

// FormatDoctor prints check results; returns an error when any check failed.
func FormatDoctor(w io.Writer, checks []DoctorCheck) error {
	ok := true
	for _, c := range checks {
		if c.Err != nil {
			ok = false
			if _, err := fmt.Fprintf(w, "✗ %s: %v\n", c.Name, c.Err); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "✓ %s\n", c.Name); err != nil {
				return err
			}
		}
	}
	if ok {
		_, err := fmt.Fprintln(w, "Mémoire runtime saine.")
		return err
	}
	return fmt.Errorf("memory doctor: au moins un contrôle a échoué")
}
