package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExecuteReplay runs a replay session according to mode flags (spec §11-13).
func ExecuteReplay(ctx context.Context, req ReplayRunRequest) (ReplayResult, error) {
	if err := ValidateReplayID(req.ReplayID); err != nil {
		return ReplayResult{}, err
	}
	pkg, err := LoadPackage(req.RepoRoot, req.ReplayID)
	if err != nil {
		return ReplayResult{}, err
	}

	mode := resolveMode(req)
	result := ReplayResult{
		ReplayID:  req.ReplayID,
		Mode:      mode,
		Offline:   req.Offline || mode == ModeOffline,
		Artifacts: auditArtifacts(pkg.Path),
	}

	if req.DryRun {
		result.Warnings = append(result.Warnings, "dry-run: no agents re-executed")
		return result, nil
	}

	switch mode {
	case ModeSimulation:
		if err := replaySimulation(pkg); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "simulation: replayed graph/events/outputs without agent execution")
		result.Warnings = append(result.Warnings, simulationReplayNotes(pkg)...)
	case ModeOffline, ModeAudit:
		if err := replayOffline(pkg); err != nil {
			return result, err
		}
		result.Warnings = append(result.Warnings, "offline: no cloud or external API calls")
	case ModeCompare:
		result.Warnings = append(result.Warnings, "compare mode: use `asa replay compare` for full diff")
	case ModeFull:
		if req.Offline {
			if err := replayOffline(pkg); err != nil {
				return result, err
			}
		} else if err := replayFull(ctx, pkg); err != nil {
			return result, err
		}
	default:
		return ReplayResult{}, fmt.Errorf("%w: %q", ErrInvalidMode, mode)
	}

	if req.Compare {
		sessionPath := filepath.Join(pkg.Path, "reports", "last-session.json")
		if _, err := os.Stat(sessionPath); err == nil {
			result.Warnings = append(result.Warnings, "prior session available for compare")
		}
	}

	baselineDivs, err := CompareBaseline(pkg.Path)
	if err != nil {
		return result, err
	}
	result.Divergences = append(result.Divergences, baselineDivs...)

	if req.Strict {
		for _, d := range result.Divergences {
			if d.Severity == "error" || d.Severity == "warning" {
				return result, ErrStrictDivergence
			}
		}
	}

	if err := writeSessionReport(pkg.Path, result); err != nil {
		return result, err
	}
	return result, nil
}

func resolveMode(req ReplayRunRequest) Mode {
	if req.Simulation {
		return ModeSimulation
	}
	if req.Compare {
		return ModeCompare
	}
	if req.Offline {
		return ModeOffline
	}
	if req.Mode != "" {
		return req.Mode
	}
	return ModeFull
}

func auditArtifacts(replayDir string) map[string]bool {
	checks := map[string]string{
		"execution_graph":      "graph/execution-graph.json",
		"trust_report":         "trust",
		"investigation_report": "investigations/report.json",
		"runtime_events":       "runtime/events.jsonl",
		"handoffs":             "context/handoffs",
	}
	out := map[string]bool{}
	for name, rel := range checks {
		path := filepath.Join(replayDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err == nil {
			out[name] = true
			continue
		}
		if strings.HasSuffix(rel, ".json") {
			if _, err := os.Stat(path + gzipSuffix); err == nil {
				out[name] = true
			}
		}
	}
	return out
}

func replaySimulation(pkg ReplayPackage) error {
	required := filepath.Join(pkg.Path, "graph", "execution-graph.json")
	if _, err := os.Stat(required); os.IsNotExist(err) {
		if _, gzErr := os.Stat(required + gzipSuffix); gzErr != nil {
			return fmt.Errorf("simulation replay: missing execution graph")
		}
	}
	for _, rel := range []string{
		"runtime/events.jsonl",
		"trust",
		"outputs",
	} {
		path := filepath.Join(pkg.Path, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := replayEventsTimeline(pkg.Path); err != nil {
		return err
	}
	return nil
}

func replayEventsTimeline(replayDir string) error {
	eventsPath := filepath.Join(replayDir, "runtime", "events.jsonl")
	body, err := ReadMaybeCompressed(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	replayed := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var evt struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			return fmt.Errorf("simulation replay: invalid event line: %w", err)
		}
		replayed++
	}
	_ = replayed
	return nil
}

func simulationReplayNotes(pkg ReplayPackage) []string {
	var notes []string
	if _, err := os.Stat(filepath.Join(pkg.Path, "outputs")); err == nil {
		notes = append(notes, "simulation: agent outputs present (not re-executed)")
	}
	return notes
}

func replayOffline(pkg ReplayPackage) error {
	if _, err := os.Stat(filepath.Join(pkg.Path, ManifestName)); err != nil {
		return ErrOfflineViolation
	}
	return replaySimulation(pkg)
}

func replayFull(ctx context.Context, pkg ReplayPackage) error {
	_ = ctx
	return replaySimulation(pkg)
}

func writeSessionReport(replayDir string, result ReplayResult) error {
	dir := filepath.Join(replayDir, "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "last-session.json"), append(body, '\n'), 0o644)
}
