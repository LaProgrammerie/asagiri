package cloud

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const (
	runStatusCompleted = "completed"
	runStatusFailed    = "failed"
)

// PushOptions configures ledger push.
type PushOptions struct {
	RepoRoot string
	Config   *config.Config
	Token    string
	RunID    string
	All      bool
	DryRun   bool
}

// Push sends local ledger runs to the cloud API.
func Push(ctx context.Context, opts PushOptions) (PushReport, error) {
	repoRoot := strings.TrimSpace(opts.RepoRoot)
	if repoRoot == "" {
		return PushReport{}, fmt.Errorf("cloud push: repo_root requis")
	}
	if err := RequirePushReady(opts.Config, opts.Token); err != nil {
		return PushReport{}, err
	}
	if !opts.All && strings.TrimSpace(opts.RunID) == "" {
		return PushReport{}, fmt.Errorf("cloud push: --run <id> ou --all requis")
	}

	ledger, err := agentledger.List(repoRoot, agentledger.ListOptions{})
	if err != nil {
		return PushReport{}, err
	}
	groups := groupEntriesByRun(ledger.Entries, opts.RunID, opts.All)
	if len(groups) == 0 {
		return PushReport{}, fmt.Errorf("cloud push: aucune entrée ledger à pousser")
	}

	mode := "apply"
	if opts.DryRun {
		mode = "dry-run"
	}
	report := PushReport{
		ReportVersion: ReportVersionPush,
		Mode:          mode,
		ProjectID:     strings.TrimSpace(opts.Config.Cloud.ProjectID),
		BaseURL:       strings.TrimSpace(opts.Config.Cloud.BaseURL),
		Items:         make([]PushItem, 0, len(groups)),
	}
	if opts.DryRun {
		report.Hint = "Relancer avec --execute pour envoyer vers le cloud"
	}

	client := NewClient(ClientOptions{
		BaseURL: report.BaseURL,
		Token:   opts.Token,
	})
	projectIRI := ProjectIRI(report.BaseURL, report.ProjectID)

	for _, g := range groups {
		item := PushItem{
			LocalRunID: g.runID,
			Feature:    g.feature,
			EntryCount: len(g.entries),
			RunStatus:  g.status,
			DryRun:     opts.DryRun,
		}
		if opts.DryRun {
			report.Items = append(report.Items, item)
			continue
		}

		runReq := BuildRunCreateRequest(projectIRI, g)
		runRes, err := client.CreateRun(ctx, runReq)
		if err != nil {
			item.Error = RedactError(err)
			report.Items = append(report.Items, item)
			continue
		}
		item.CloudRunID = runRes.ID
		runIRI := RunIRI(report.BaseURL, runRes.ID)

		var firstErr error
		for _, entry := range g.entries {
			req := BuildLedgerCreateRequest(runIRI, entry)
			if err := client.CreateLedgerEntry(ctx, req); err != nil {
				firstErr = err
				break
			}
		}
		if firstErr != nil {
			item.Error = RedactError(firstErr)
		}
		report.Items = append(report.Items, item)
	}

	return report, nil
}

type runGroup struct {
	runID                string
	taskID               string
	feature              string
	status               string
	startedAt            string
	endedAt              string
	durationMs           int64
	configReleaseVersion int
	configContentHash    string
	entries              []agentledger.Entry
}

func groupEntriesByRun(entries []agentledger.Entry, runFilter string, all bool) []runGroup {
	runFilter = strings.TrimSpace(runFilter)
	byRun := map[string][]agentledger.Entry{}
	order := []string{}
	for _, e := range entries {
		rid := strings.TrimSpace(e.RunID)
		if rid == "" {
			continue
		}
		if !all && runFilter != "" && rid != runFilter {
			continue
		}
		if _, ok := byRun[rid]; !ok {
			order = append(order, rid)
		}
		byRun[rid] = append(byRun[rid], e)
	}
	sort.Strings(order)
	out := make([]runGroup, 0, len(order))
	for _, rid := range order {
		ents := byRun[rid]
		out = append(out, summarizeRunGroup(rid, ents))
	}
	return out
}

func summarizeRunGroup(runID string, entries []agentledger.Entry) runGroup {
	g := runGroup{
		runID:   runID,
		entries: entries,
		status:  runStatusCompleted,
	}
	for _, e := range entries {
		if g.taskID == "" && strings.TrimSpace(e.TaskID) != "" {
			g.taskID = strings.TrimSpace(e.TaskID)
		}
		if g.feature == "" && strings.TrimSpace(e.Feature) != "" {
			g.feature = e.Feature
		}
		if e.ExitCode != 0 {
			g.status = runStatusFailed
		}
		if g.startedAt == "" || e.StartedAt < g.startedAt {
			g.startedAt = e.StartedAt
		}
		if e.EndedAt > g.endedAt {
			g.endedAt = e.EndedAt
		}
		g.durationMs += e.DurationMS
		if g.configReleaseVersion == 0 && e.ConfigReleaseVersion > 0 {
			g.configReleaseVersion = e.ConfigReleaseVersion
		}
		if g.configContentHash == "" && strings.TrimSpace(e.ConfigContentHash) != "" {
			g.configContentHash = strings.TrimSpace(e.ConfigContentHash)
		}
	}
	return g
}
