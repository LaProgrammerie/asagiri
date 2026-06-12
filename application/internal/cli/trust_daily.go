package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

func printDailyTaskTrust(w io.Writer, actx *appContext, taskID string) {
	taskID = strings.TrimSpace(taskID)
	if actx == nil || taskID == "" {
		return
	}
	task, err := actx.Store.GetTask(taskID)
	if err != nil {
		return
	}
	report, err := worktrust.BuildTaskReport(actx.RepoRoot, actx.Config, *task)
	if err != nil {
		return
	}
	_, _ = fmt.Fprint(w, worktrust.FormatDailyNextBlock(report))
}

func printDailyFeatureTrust(w io.Writer, actx *appContext, feature string) {
	feature = strings.TrimSpace(feature)
	if actx == nil || feature == "" {
		return
	}
	tasks, err := actx.Store.ListTasksByFeature(feature)
	if err != nil || len(tasks) == 0 {
		return
	}
	report, err := worktrust.BuildFeatureReport(actx.RepoRoot, actx.Config, feature, tasks)
	if err != nil {
		return
	}
	_, _ = fmt.Fprint(w, worktrust.FormatDailyStatusBlock(report))
}
