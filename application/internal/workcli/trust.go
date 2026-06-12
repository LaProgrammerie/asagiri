package workcli

import (
	"fmt"
	"io"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

// ResolveStrictTrust resolves product flow for --strict-trust on work.
func ResolveStrictTrust(repoRoot, trustFlowFlag, featureHint string) (flowID, productID string, err error) {
	if flow := strings.TrimSpace(trustFlowFlag); flow != "" {
		flowID, productID, err = trust.ResolveProductFlow(repoRoot, flow)
		if err != nil {
			return "", "", fmt.Errorf("strict-trust: %w (pass --trust-flow with a flow id from .asagiri/products/*/flows/)", err)
		}
		return flowID, productID, nil
	}
	feature := strings.TrimSpace(featureHint)
	if feature == "" {
		return "", "", fmt.Errorf("strict-trust: pass --trust-flow with a product flow id")
	}
	flowID, productID, err = trust.ResolveProductFlow(repoRoot, feature)
	if err != nil {
		return "", "", fmt.Errorf(
			"strict-trust: could not resolve product flow from feature %q: %w; pass --trust-flow",
			feature, err,
		)
	}
	return flowID, productID, nil
}

func printPostWorkTrustLine(w io.Writer, actx *WorkContext, feature, taskID, runID string) {
	if actx == nil {
		return
	}
	feature = strings.TrimSpace(feature)
	taskID = strings.TrimSpace(taskID)
	runID = strings.TrimSpace(runID)

	if taskID != "" {
		task, err := actx.Store.GetTask(taskID)
		if err != nil {
			return
		}
		report, err := worktrust.BuildTaskReport(actx.RepoRoot, actx.Config, *task)
		if err != nil {
			return
		}
		scope := taskID
		if feature != "" {
			scope = feature + " / " + taskID
		}
		_, _ = fmt.Fprintln(w, worktrust.FormatDailyPostWorkLine(scope, report))
		return
	}
	if runID != "" {
		report, err := worktrust.BuildRunReport(actx.RepoRoot, actx.Config, actx.Store, runID)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintln(w, worktrust.FormatDailyPostWorkFromRun(report))
		return
	}
	if feature != "" {
		tasks, err := actx.Store.ListTasksByFeature(feature)
		if err != nil || len(tasks) == 0 {
			return
		}
		report, err := worktrust.BuildFeatureReport(actx.RepoRoot, actx.Config, feature, tasks)
		if err != nil {
			return
		}
		lineReport := worktrust.WorkTrustReport{
			Scope: worktrust.TrustScope{Kind: "feature", ID: feature, Feature: feature},
			Score: report.Score,
		}
		_, _ = fmt.Fprintln(w, worktrust.FormatDailyPostWorkLine(feature, lineReport))
	}
}
