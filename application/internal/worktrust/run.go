package worktrust

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

// RunStore loads run metadata and tasks for run-scoped synthesis.
type RunStore interface {
	GetRun(id string) (*sqlite.Run, error)
	ListTasksByRun(runID string) ([]sqlite.Task, error)
}

// BuildRunReport synthesizes trust reports for all tasks in a run plus the run plan gate.
func BuildRunReport(repoRoot string, cfg *config.Config, store RunStore, runID string) (RunTrustReport, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return RunTrustReport{}, fmt.Errorf("run id required")
	}
	if store == nil {
		return RunTrustReport{}, fmt.Errorf("store required")
	}

	run, err := store.GetRun(runID)
	if err != nil {
		return RunTrustReport{}, fmt.Errorf("run %q not found in local state: %w", runID, err)
	}

	tasks, err := store.ListTasksByRun(runID)
	if err != nil {
		return RunTrustReport{}, err
	}
	if len(tasks) == 0 {
		return RunTrustReport{}, fmt.Errorf("no tasks for run %q in local state", runID)
	}

	summaries := make([]FeatureTaskSummary, 0, len(tasks))
	var scoreSum float64
	worst := VerdictTrusted

	for _, task := range tasks {
		report, err := BuildTaskReport(repoRoot, cfg, task)
		if err != nil {
			return RunTrustReport{}, fmt.Errorf("task %s: %w", task.ID, err)
		}
		summaries = append(summaries, FeatureTaskSummary{
			TaskID:    task.ID,
			Status:    task.Status,
			Score:     report.Score.Overall,
			Verdict:   report.Score.Verdict,
			Command:   report.Recommendation.Command,
			Rationale: report.Recommendation.Rationale,
		})
		scoreSum += report.Score.Overall
		worst = worstVerdict(worst, report.Score.Verdict)
	}

	var planGate *RunPlanGateSummary
	if pg, ok := readRunPlanGate(repoRoot, runID); ok {
		planGate = &pg
		worst = worstVerdict(worst, planGateVerdict(pg.Status))
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Verdict != summaries[j].Verdict {
			return verdictRank(summaries[i].Verdict) > verdictRank(summaries[j].Verdict)
		}
		return summaries[i].Score < summaries[j].Score
	})

	avg := scoreSum / float64(len(summaries))

	return RunTrustReport{
		ReportVersion: ReportVersion,
		Scope: TrustScope{
			Kind:    "run",
			ID:      runID,
			Feature: run.Feature,
			Status:  run.Status,
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Score: WorkTrustScore{
			Overall: avg,
			Verdict: worst,
			Summary: verdictSummary(worst, avg),
		},
		TaskCount:   len(summaries),
		PlanGate:    planGate,
		Tasks:       summaries,
		NextActions: collectRunNextActions(repoRoot, cfg, run.Feature, tasks, summaries, planGate),
	}, nil
}

func readRunPlanGate(repoRoot, runID string) (RunPlanGateSummary, bool) {
	path := filepath.Join(repoRoot, ".asagiri", "logs", runID, "gates", "plan.json")
	body, err := os.ReadFile(path)
	if err != nil {
		return RunPlanGateSummary{}, false
	}
	var doc gates.LogDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return RunPlanGateSummary{}, false
	}
	notes := strings.Join(doc.Notes, "; ")
	if notes == "" && doc.ParseError != "" {
		notes = doc.ParseError
	}
	return RunPlanGateSummary{
		Status:     strings.TrimSpace(doc.Status),
		Confidence: doc.Confidence,
		Notes:      notes,
	}, true
}

func planGateVerdict(status string) Verdict {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "fail", "failed", "error":
		return VerdictBlocked
	case "warn", "warning":
		return VerdictRisky
	case "pass", "passed", "ok":
		return VerdictTrusted
	default:
		return VerdictAcceptable
	}
}

func collectRunNextActions(repoRoot string, cfg *config.Config, feature string, tasks []sqlite.Task, summaries []FeatureTaskSummary, planGate *RunPlanGateSummary) []WorkTrustRecommendation {
	actions := collectFeatureNextActions(feature, summaries)
	if planGate != nil && verdictRank(planGateVerdict(planGate.Status)) >= verdictRank(VerdictRisky) {
		rationale := fmt.Sprintf("plan gate %s — puis %s", humanGateStatus(planGate.Status), NextCommandForFeature(feature))
		priority := "required"
		if verdictRank(planGateVerdict(planGate.Status)) == verdictRank(VerdictRisky) {
			priority = "suggested"
		}
		return []WorkTrustRecommendation{{
			Action:    "next",
			Command:   NextCommandForFeature(feature),
			Rationale: rationale,
			Priority:  priority,
		}}
	}
	return actions
}
