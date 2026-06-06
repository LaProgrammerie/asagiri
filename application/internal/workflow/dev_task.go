package workflow

import (
	"context"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

type governanceOutcome int

const (
	governanceOK governanceOutcome = iota
	governanceRetryDev
)

// devOneTask runs dev agent for one task: running → implemented, reusing worktree when set.
func (s *Service) devOneTask(ctx context.Context, run *sqlite.Run, feature string, task sqlite.Task, a agent.Agent, force bool) (sqlite.Task, error) {
	if st := normalizeStatus(task.Status); st == asagiri.StatusPlanned || st == asagiri.StatusPending {
		if err := s.transitionTask(task, asagiri.StatusEnriched, true); err != nil {
			return task, err
		}
		if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
			task = *fresh
		}
	}
	if normalizeStatus(task.Status) != asagiri.StatusRunning {
		if err := s.transitionTask(task, asagiri.StatusRunning, force); err != nil {
			return task, err
		}
		if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
			task = *fresh
		}
	}

	worktreePath := strings.TrimSpace(task.WorktreePath)
	if worktreePath == "" {
		path, _, err := s.worktreeMngr.Create(ctx, feature, task.ID)
		if err != nil {
			_ = s.transitionTask(task, asagiri.StatusFailed, true)
			return task, err
		}
		worktreePath = path
		if err := s.store.UpdateTask(&sqlite.Task{ID: task.ID, WorktreePath: worktreePath}); err != nil {
			return task, err
		}
		task.WorktreePath = worktreePath
	}

	canonical, _ := payloadToCanonical(task.PayloadJSON)
	agentCtx := agent.BuildContext(run.ID, &canonical, s.contextFilesForTask(feature, canonical))
	res, runErr := a.Run(ctx, agent.RunRequest{
		Feature:    feature,
		TaskID:     task.ID,
		Prompt:     "Implémente la task " + task.ID,
		WorkingDir: worktreePath,
	})
	agentRes := agent.DryRunResult("implémentation simulée")
	if parsed, ok := agent.ParseResult(res.Stdout); ok {
		agentRes = parsed
	}
	_ = agent.WriteLogs(s.repoRoot, task.ID, agentCtx, agentRes)
	if runErr != nil {
		_ = s.transitionTask(task, asagiri.StatusFailed, true)
		return task, runErr
	}

	if err := s.writeTaskLog(task.ID, "dev.log", res.Stdout+"\n"+res.Stderr); err != nil {
		return task, err
	}
	if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
		task = *fresh
	}
	if err := s.transitionTask(task, asagiri.StatusImplemented, force); err != nil {
		return task, err
	}
	if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
		task = *fresh
	}
	return task, nil
}

func (s *Service) devTaskWithGovernanceRetries(ctx context.Context, run *sqlite.Run, feature string, task sqlite.Task, a agent.Agent, force bool) error {
	for {
		var err error
		task, err = s.devOneTask(ctx, run, feature, task, a, force)
		if err != nil {
			return err
		}
		outcome, next, err := s.processGovernanceAfterDev(ctx, feature, task)
		if err != nil {
			return err
		}
		task = next
		if outcome == governanceOK {
			return nil
		}
	}
}
