package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontract"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/agentobservability"
	"github.com/LaProgrammerie/asagiri/application/internal/devresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
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
	if gates.EnrichGateBlocksDev(s.cfg, task.Status, task.PayloadJSON) {
		return task, fmt.Errorf("enrich gate required before dev: run asa enrich %s --task %s --force", feature, task.ID)
	}
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
	contextFiles := s.contextFilesForTask(feature, canonical)
	agentCtx := agent.BuildContext(run.ID, &canonical, contextFiles)
	resolved, resolveErr := devresolve.Resolve(devresolve.Params{
		RepoRoot:     s.repoRoot,
		Config:       s.cfg,
		AgentKey:     a.Name(),
		RunID:        run.ID,
		Feature:      feature,
		TaskID:       task.ID,
		ContextFiles: contextFiles,
	})
	if resolveErr != nil {
		_ = s.transitionTask(task, asagiri.StatusFailed, true)
		return task, resolveErr
	}
	if resolved.Warning != "" {
		_ = s.writeTaskLog(task.ID, "dev-orchestration.warn", resolved.Warning+"\n")
	}
	obs := agentobservability.New(s.repoRoot, task.ID, resolved.AgentID, s.cfg)
	res, runErr := a.Run(ctx, agent.RunRequest{
		Feature:    feature,
		TaskID:     task.ID,
		Prompt:     resolved.Prompt,
		WorkingDir: worktreePath,
	})
	agentRes := agent.DryRunResult("implémentation simulée")
	if parsed, ok := agent.ParseResult(res.Stdout); ok {
		agentRes = parsed
	}
	if err := obs.WriteLegacyLogs(agentCtx, agentRes); err != nil {
		_ = s.transitionTask(task, asagiri.StatusFailed, true)
		return task, err
	}
	var contractValid *bool
	if resolved.Orchestrated {
		contract := agentcontract.ValidateOutput(resolved.Spec, res.Stdout)
		if err := obs.WriteContract(contract); err != nil {
			_ = s.transitionTask(task, asagiri.StatusFailed, true)
			return task, err
		}
		valid := contract.Valid
		contractValid = &valid
	}
	if err := obs.RecordLedger(agentledger.Params{
		TaskID:        task.ID,
		RunID:         run.ID,
		Feature:       feature,
		AgentKey:      a.Name(),
		AgentID:       resolved.AgentID,
		Role:          resolved.Role,
		Provider:      agentledger.ProviderFromConfig(s.cfg, a.Name()),
		Phase:         "dev",
		Prompt:        resolved.Prompt,
		ContextHash:   resolved.ContextHash,
		ContractValid: contractValid,
		LogDir:        resolved.LogDir,
		Result:        res,
	}); err != nil {
		_ = s.transitionTask(task, asagiri.StatusFailed, true)
		return task, err
	}
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
		if outcome == governanceRetryDev {
			continue
		}
		return s.processHumanReviewAfterDev(ctx, feature, task)
	}
}
