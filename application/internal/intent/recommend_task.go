package intent

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// RecommendForTask computes the next primitive for a specific task (same rules as RecommendNext).
func RecommendForTask(repoRoot string, cfg *config.Config, task sqlite.Task) (NextRecommendation, error) {
	feature := strings.TrimSpace(task.Feature)
	if feature == "" {
		return NextRecommendation{}, fmt.Errorf("task sans feature")
	}
	fs := FeatureState{
		Name:           feature,
		HasTasks:       true,
		TaskCount:      1,
		NextTaskID:     task.ID,
		NextTaskStatus: task.Status,
	}
	if cfg != nil {
		fs.EnrichGateBlocksDev = gates.EnrichGateBlocksDev(cfg, task.Status, task.PayloadJSON)
		fs.VerifyEvidenceGateBlocksReview = gates.VerifyEvidenceGateBlocksReview(cfg, task.Status, task.PayloadJSON)
		fs.TrustGateBlocksReview = gates.TrustGateBlocksReview(cfg, task.Status, task.PayloadJSON)
		if pg, ok := gates.BlockingPendingForTask(repoRoot, cfg, task); ok {
			pgCopy := pg
			fs.PendingGate = &pgCopy
		}
	}
	return recommendNextFromFeatureState(fs, feature, cfg)
}

// RecommendNextFromTasks computes the next primitive for a feature from its SQLite tasks.
func RecommendNextFromTasks(repoRoot string, cfg *config.Config, feature string, tasks []sqlite.Task) (NextRecommendation, error) {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return NextRecommendation{}, fmt.Errorf("feature required")
	}
	fs := FeatureState{Name: feature, HasTasks: len(tasks) > 0, TaskCount: len(tasks)}
	if len(tasks) > 0 {
		applyTaskGateFields(&fs, tasks, repoRoot, cfg)
	}
	return recommendNextFromFeatureState(fs, feature, cfg)
}

func recommendNextFromFeatureState(fs FeatureState, feature string, cfg *config.Config) (NextRecommendation, error) {
	devAgent := config.DefaultAgentDev
	if cfg != nil && cfg.Work.DefaultAgent != "" {
		devAgent = cfg.Work.DefaultAgent
	}
	reviewAgent := config.DefaultAgentReviewer
	if cfg != nil && cfg.Work.DefaultReviewer != "" {
		reviewAgent = cfg.Work.DefaultReviewer
	}
	taskID := fs.NextTaskID
	if rec, ok := recommendForPendingGate(fs, feature, taskID); ok {
		return rec, nil
	}
	switch fs.NextTaskStatus {
	case asagiri.StatusImplemented, asagiri.StatusRunning:
		cmd := fmt.Sprintf("asa verify %s --task %s", feature, taskID)
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "verify",
			Reason: "implementation completed but validation missing", Primitive: cmd,
		}, nil
	case asagiri.StatusVerified:
		if fs.VerifyEvidenceGateBlocksReview {
			cmd := fmt.Sprintf("asa verify %s --task %s --force", feature, taskID)
			return NextRecommendation{
				Feature: feature, TaskID: taskID, Action: "verify",
				Reason: "verify evidence gate not satisfied", Primitive: cmd,
			}, nil
		}
		if fs.TrustGateBlocksReview {
			cmd := fmt.Sprintf("asa verify %s --task %s --force", feature, taskID)
			return NextRecommendation{
				Feature: feature, TaskID: taskID, Action: "verify",
				Reason: "trust gate not satisfied", Primitive: cmd,
			}, nil
		}
		cmd := fmt.Sprintf("asa review %s --task %s --agent %s", feature, taskID, reviewAgent)
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "review",
			Reason: "verified but review missing", Primitive: cmd,
		}, nil
	case asagiri.StatusEnriched, asagiri.StatusPending, asagiri.StatusPlanned, "":
		if taskID == "" {
			cmd := fmt.Sprintf("asa plan %s", feature)
			return NextRecommendation{Feature: feature, Action: "plan", Reason: "no tasks planned", Primitive: cmd}, nil
		}
		if fs.EnrichGateBlocksDev {
			enricher := config.DefaultAgentEnrich
			if cfg != nil && cfg.Work.DefaultEnricher != "" {
				enricher = cfg.Work.DefaultEnricher
			}
			cmd := fmt.Sprintf("asa enrich %s --task %s --agent %s --force", feature, taskID, enricher)
			return NextRecommendation{
				Feature: feature, TaskID: taskID, Action: "enrich",
				Reason: "enrich gate not satisfied", Primitive: cmd,
			}, nil
		}
		cmd := fmt.Sprintf("asa dev %s --task %s --agent %s", feature, taskID, devAgent)
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "dev",
			Reason: "task ready for implementation", Primitive: cmd,
		}, nil
	case asagiri.StatusVerifyFailed:
		cmd := fmt.Sprintf("asa verify %s --task %s --force", feature, taskID)
		return NextRecommendation{Feature: feature, TaskID: taskID, Action: "verify", Reason: "verification failed", Primitive: cmd}, nil
	default:
		cmd := "asa status"
		return NextRecommendation{Feature: feature, Action: "status", Reason: "inspect current state", Primitive: cmd}, nil
	}
}
