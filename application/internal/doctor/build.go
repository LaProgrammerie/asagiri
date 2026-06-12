package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

// Build collects a read-only doctor report from startDir (typically cwd).
func Build(startDir string, opts Options) (Report, error) {
	report := Report{ReportVersion: ReportVersion}

	repoRoot, gitErr := bootstrap.GitRoot(startDir)
	if gitErr != nil {
		appendCheck(&report, Check{ID: "git", Status: StatusFail, Message: gitErr.Error()})
		report.NextActions = append(report.NextActions, Action{
			Title: "Initialiser un dépôt Git",
			CLI:   "git init",
		})
		Finalize(&report)
		return report, nil
	}
	report.Repository.GitRoot = repoRoot
	appendCheck(&report, Check{ID: "git", Status: StatusOK})

	cfgPath := config.ConfigPath(repoRoot)
	report.Repository.ConfigPath = cfgPath
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	if cfgErr != nil {
		report.Repository.ConfigError = cfgErr.Error()
		appendCheck(&report, Check{ID: "config", Status: StatusFail, Message: cfgErr.Error()})
		report.NextActions = append(report.NextActions, Action{
			Title: "Créer la configuration Asagiri",
			CLI:   "asa init",
		})
		Finalize(&report)
		return report, nil
	}
	report.Repository.ConfigLoaded = true
	appendCheck(&report, Check{ID: "config", Status: StatusOK})
	report.Gates = collectGates(cfg)
	report.Agents, report.MissingTools = collectAgents(cfg)
	specCollect := collectAgentSpecs(repoRoot, cfg)
	report.AgentRegistry = specCollect.registry
	report.AgentSpecs = specCollect.specs
	report.AgentDrift = specCollect.drifts
	report.LastOrchestrated = specCollect.last
	enrichAgentsWithSpecs(report.Agents, specCollect.specs)
	for _, c := range specCollect.checks {
		appendCheck(&report, c)
	}
	for _, a := range specCollect.actions {
		report.NextActions = appendUniqueAction(report.NextActions, a)
	}

	dbPath := cfg.StateDBPath(repoRoot)
	report.State.SQLitePath = dbPath
	if _, err := os.Stat(dbPath); err != nil {
		msg := "state.sqlite absent — lancez asa init"
		if !os.IsNotExist(err) {
			msg = err.Error()
		}
		appendCheck(&report, Check{ID: "sqlite", Status: StatusFail, Message: msg})
		report.NextActions = append(report.NextActions, Action{Title: msg, CLI: "asa init"})
	} else {
		report.State.SQLitePresent = true
		store, err := sqlite.Open(dbPath)
		if err != nil {
			appendCheck(&report, Check{ID: "sqlite", Status: StatusFail, Message: err.Error()})
		} else {
			defer func() { _ = store.Close() }()
			if err := store.Ping(); err != nil {
				appendCheck(&report, Check{ID: "sqlite", Status: StatusFail, Message: err.Error()})
			} else {
				appendCheck(&report, Check{ID: "sqlite", Status: StatusOK})
				if v, err := store.SchemaVersion(); err != nil {
					appendCheck(&report, Check{ID: "schema", Status: StatusFail, Message: err.Error()})
				} else {
					report.State.SchemaVersion = v
					if v < 1 {
						appendCheck(&report, Check{
							ID: "schema", Status: StatusFail,
							Message: "schéma non migré — lancez asa init",
						})
						report.NextActions = append(report.NextActions, Action{Title: "Migrer SQLite", CLI: "asa init"})
					} else {
						appendCheck(&report, Check{ID: "schema", Status: StatusOK})
						fillStateCounts(&report, store)
						snap, err := intent.BuildSnapshot(repoRoot, cfg, store)
						if err == nil {
							if report.State.ActiveFeature == "" {
								report.State.ActiveFeature = snap.ActiveFeature
							}
							report.Trust = collectTrust(repoRoot, cfg, store, snap)
							if report.Trust != nil && report.Trust.TasksAtRisk > 0 {
								report.NextActions = appendUniqueAction(report.NextActions, Action{
									Title: "Détail confiance par task",
									CLI:   fmt.Sprintf("asa trust feature %s", report.Trust.Feature),
								})
							}
							if rec, err := intent.RecommendNext(snap, snap.ActiveFeature, cfg); err == nil {
								if cli := strings.TrimSpace(rec.Primitive); cli != "" {
									title := strings.TrimSpace(rec.Reason)
									if title == "" {
										title = "Prochaine action workflow"
									}
									report.NextActions = appendUniqueAction(report.NextActions, Action{Title: title, CLI: cli})
								}
							}
						}
					}
				}
			}
		}
	}

	for _, sub := range []string{"runs", "tasks", "logs", "worktrees"} {
		p := filepath.Join(repoRoot, ".asagiri", sub)
		if _, err := os.Stat(p); err != nil {
			msg := filepath.Join(".asagiri", sub) + " manquant — lancez asa init"
			appendCheck(&report, Check{ID: "dir:" + sub, Status: StatusFail, Message: msg})
		} else {
			appendCheck(&report, Check{ID: "dir:" + sub, Status: StatusOK})
		}
	}

	// Onboarding checks (--full) : délégués à onboarding.RunDoctorChecks (pas de duplication).
	if opts.Full && cfg != nil {
		for _, c := range onboarding.RunDoctorChecks(repoRoot, cfg, onboarding.DoctorOpts{Full: true, SkipExec: true}) {
			appendCheck(&report, onboardingToCheck(c))
			if c.Status == onboarding.StatusWarn || c.Status == onboarding.StatusFail {
				if cli := strings.TrimSpace(c.FixCLI); cli != "" {
					report.NextActions = appendUniqueAction(report.NextActions, Action{
						Title: strings.TrimSpace(c.Message),
						CLI:   cli,
					})
				}
			}
		}
	}

	if len(report.NextActions) == 0 {
		if !hasFailChecks(report.Checks) {
			report.NextActions = append(report.NextActions, Action{
				Title: "Explorer l'état du projet",
				CLI:   "asa status",
			})
		} else if !report.State.SQLitePresent {
			report.NextActions = appendUniqueAction(report.NextActions, Action{Title: "Initialiser Asagiri", CLI: "asa init"})
		}
	}

	Finalize(&report)
	return report, nil
}

func appendCheck(r *Report, c Check) {
	r.Checks = append(r.Checks, c)
}

func hasFailChecks(checks []Check) bool {
	for _, c := range checks {
		if c.Status == StatusFail {
			return true
		}
	}
	return false
}

func fillStateCounts(report *Report, store *sqlite.Store) {
	runs, err := store.ListRuns(500)
	if err != nil {
		return
	}
	report.State.RunCount = len(runs)
	seen := map[string]struct{}{}
	var latest time.Time
	for _, r := range runs {
		if r.UpdatedAt.After(latest) {
			latest = r.UpdatedAt
			report.State.ActiveFeature = r.Feature
		}
		tasks, err := store.ListTasksByFeature(r.Feature)
		if err != nil {
			continue
		}
		for _, t := range tasks {
			if _, ok := seen[t.ID]; ok {
				continue
			}
			seen[t.ID] = struct{}{}
			report.State.TaskCount++
		}
	}
}

func collectTrust(repoRoot string, cfg *config.Config, store *sqlite.Store, snap intent.StateSnapshot) *TrustInfo {
	feature := strings.TrimSpace(snap.ActiveFeature)
	if feature == "" {
		return nil
	}
	tasks, err := store.ListTasksByFeature(feature)
	if err != nil || len(tasks) == 0 {
		return nil
	}
	fr, err := worktrust.BuildFeatureReport(repoRoot, cfg, feature, tasks)
	if err != nil {
		return nil
	}
	atRisk := 0
	for _, t := range fr.Tasks {
		if trustVerdictRank(t.Verdict) >= trustVerdictRank(worktrust.VerdictRisky) {
			atRisk++
		}
	}
	info := &TrustInfo{
		Feature:     feature,
		Verdict:     trustVerdictLabel(fr.Score.Verdict),
		TasksAtRisk: atRisk,
	}
	if atRisk > 0 {
		info.Summary = fmtTasksAtRisk(atRisk)
	} else if len(fr.NextActions) > 0 {
		info.Summary = strings.TrimSpace(fr.NextActions[0].Rationale)
	}
	return info
}

func onboardingToCheck(c onboarding.Check) Check {
	ch := Check{ID: c.ID, Message: c.Message}
	switch c.Status {
	case onboarding.StatusOK:
		ch.Status = StatusOK
	case onboarding.StatusWarn:
		ch.Status = StatusWarn
	default:
		ch.Status = StatusFail
	}
	return ch
}

func appendUniqueAction(actions []Action, a Action) []Action {
	a.CLI = strings.TrimSpace(a.CLI)
	if a.CLI == "" {
		return actions
	}
	for _, existing := range actions {
		if existing.CLI == a.CLI {
			return actions
		}
	}
	return append(actions, a)
}

func fmtTasksAtRisk(n int) string {
	if n == 1 {
		return "1 task à risque"
	}
	return fmt.Sprintf("%d tasks à risque", n)
}
