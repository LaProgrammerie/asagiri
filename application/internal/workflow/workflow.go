package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	agentexec "github.com/LaProgrammerie/asagiri/application/internal/agent/exec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/plan"
	"github.com/LaProgrammerie/asagiri/application/internal/policy"
	"github.com/LaProgrammerie/asagiri/application/internal/rag"
	"github.com/LaProgrammerie/asagiri/application/internal/report"
	"github.com/LaProgrammerie/asagiri/application/internal/spec"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktree"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// StepState tracks one step execution in run.steps_json.
type StepState struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type Service struct {
	repoRoot       string
	cfg            *config.Config
	store          *sqlite.Store
	specReader     *spec.Reader
	reportWriter   *report.Writer
	worktreeMngr   *worktree.Manager
	dryRun         bool
	agentFactories map[string]func(bool) (agent.Agent, error)
}

func NewService(repoRoot string, cfg *config.Config, store *sqlite.Store, dryRun bool) *Service {
	worktreesPath := cfg.Resolve(repoRoot, cfg.Worktrees.BasePath)
	return &Service{
		repoRoot:     repoRoot,
		cfg:          cfg,
		store:        store,
		specReader:   spec.NewReader(repoRoot, cfg),
		reportWriter: report.NewWriter(repoRoot),
		worktreeMngr: worktree.New(
			repoRoot,
			worktreesPath,
			cfg.Worktrees.BranchPrefix,
			cfg.Project.DefaultBranch,
			dryRun,
		),
		dryRun: dryRun,
	}
}

func (s *Service) ensureAgent(name string) (agent.Agent, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("agent requis")
	}
	cfgAgent, ok := s.cfg.Agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q non défini dans config", name)
	}
	return agentexec.New(name, cfgAgent, s.dryRun)
}

func newRunID(feature string) string {
	return fmt.Sprintf("run-%s-%s", sanitize(feature), time.Now().UTC().Format("20060102-150405.000"))
}

func sanitize(in string) string {
	x := strings.ToLower(strings.TrimSpace(in))
	x = strings.ReplaceAll(x, " ", "-")
	x = strings.ReplaceAll(x, "/", "-")
	x = strings.Trim(x, "-")
	if x == "" {
		return "feature"
	}
	return x
}

func serializeSteps(steps []StepState) string {
	body, _ := json.Marshal(steps)
	return string(body)
}

func deserializeSteps(raw string) []StepState {
	if strings.TrimSpace(raw) == "" {
		return []StepState{}
	}
	var out []StepState
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []StepState{}
	}
	return out
}

func (s *Service) createRun(feature string, stepNames ...string) (*sqlite.Run, error) {
	steps := make([]StepState, 0, len(stepNames))
	for _, step := range stepNames {
		steps = append(steps, StepState{Name: step, Status: sqlite.StatusPending})
	}
	run := &sqlite.Run{
		ID:        newRunID(feature),
		Feature:   feature,
		Status:    sqlite.StatusPending,
		StepsJSON: serializeSteps(steps),
	}
	if err := s.store.CreateRun(run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) updateStep(run *sqlite.Run, stepName, status, message string) error {
	steps := deserializeSteps(run.StepsJSON)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for idx := range steps {
		if steps[idx].Name != stepName {
			continue
		}
		steps[idx].Status = status
		steps[idx].Message = message
		if status == sqlite.StatusRunning {
			steps[idx].StartedAt = now
		}
		if status == sqlite.StatusDone || status == sqlite.StatusFailed || status == sqlite.StatusVerified || status == sqlite.StatusReviewed {
			if steps[idx].StartedAt == "" {
				steps[idx].StartedAt = now
			}
			steps[idx].EndedAt = now
		}
	}
	run.StepsJSON = serializeSteps(steps)
	run.Status = status
	return s.store.UpdateRun(run)
}

func (s *Service) PlanFeature(feature string) (string, []plan.Task, error) {
	run, err := s.createRun(feature, "plan")
	if err != nil {
		return "", nil, err
	}
	if err := s.updateStep(run, "plan", sqlite.StatusRunning, "normalisation des tasks"); err != nil {
		return "", nil, err
	}

	doc, err := s.specReader.ReadFeature(feature)
	if err != nil {
		_ = s.updateStep(run, "plan", sqlite.StatusFailed, err.Error())
		return run.ID, nil, err
	}
	tasks, err := plan.Normalize(feature, doc)
	if err != nil {
		_ = s.updateStep(run, "plan", sqlite.StatusFailed, err.Error())
		return run.ID, nil, err
	}
	canonicalTasks := make([]asagiri.Task, 0, len(tasks))
	for _, task := range tasks {
		canonical := planToCanonical(feature, task)
		canonical.Status = asagiri.StatusPlanned
		canonical.Validation.Commands = s.cfg.ValidationCommandLines()
		payload, marshalErr := canonicalToPayload(canonical)
		if marshalErr != nil {
			_ = s.updateStep(run, "plan", sqlite.StatusFailed, marshalErr.Error())
			return run.ID, nil, marshalErr
		}
		if err := s.store.CreateTask(&sqlite.Task{
			ID:          task.ID,
			RunID:       run.ID,
			Feature:     feature,
			Status:      asagiri.StatusPlanned,
			PayloadJSON: payload,
		}); err != nil {
			_ = s.updateStep(run, "plan", sqlite.StatusFailed, err.Error())
			return run.ID, nil, err
		}
		canonicalTasks = append(canonicalTasks, canonical)
	}
	if err := s.persistCanonicalTaskFiles(feature, canonicalTasks); err != nil {
		_ = s.updateStep(run, "plan", sqlite.StatusFailed, err.Error())
		return run.ID, tasks, err
	}
	if err := s.updateStep(run, "plan", sqlite.StatusDone, "plan généré"); err != nil {
		return run.ID, tasks, err
	}
	return run.ID, tasks, nil
}

func (s *Service) SpecFeature(ctx context.Context, feature, agentName string) (string, error) {
	run, err := s.createRun(feature, "spec")
	if err != nil {
		return "", err
	}
	if err := s.updateStep(run, "spec", sqlite.StatusRunning, "génération/lecture spec"); err != nil {
		return run.ID, err
	}
	a, err := s.ensureAgent(agentName)
	if err != nil {
		_ = s.updateStep(run, "spec", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	_, err = a.Run(ctx, agent.RunRequest{
		Feature:    feature,
		Prompt:     "Lire ou produire la spec de la feature " + feature,
		WorkingDir: s.repoRoot,
	})
	if err != nil {
		_ = s.updateStep(run, "spec", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	if err := s.updateStep(run, "spec", sqlite.StatusDone, "spec traitée"); err != nil {
		return run.ID, err
	}
	return run.ID, nil
}

func (s *Service) pickTasks(feature, taskID string) ([]sqlite.Task, error) {
	all, err := s.store.ListTasksByFeature(feature)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("aucune tâche pour la feature %q — lancez asa plan %s", feature, feature)
	}
	if taskID == "" {
		return all, nil
	}
	for _, t := range all {
		if t.ID == taskID {
			return []sqlite.Task{t}, nil
		}
	}
	return nil, fmt.Errorf("task %q introuvable pour la feature %q", taskID, feature)
}

func (s *Service) EnrichFeature(ctx context.Context, feature, taskID, agentName string, force bool) (string, error) {
	run, err := s.createRun(feature, "enrich")
	if err != nil {
		return "", err
	}
	if err := s.updateStep(run, "enrich", sqlite.StatusRunning, "enrichissement des tasks"); err != nil {
		return run.ID, err
	}

	tasks, err := s.pickTasks(feature, taskID)
	if err != nil {
		_ = s.updateStep(run, "enrich", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}

	_ = policy.IsOllamaAgent(agentName)
	a, err := s.ensureAgent(agentName)
	if err != nil {
		_ = s.updateStep(run, "enrich", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	for _, task := range tasks {
		if st := normalizeStatus(task.Status); st == asagiri.StatusPending {
			if err := s.transitionTask(task, asagiri.StatusPlanned, true); err != nil {
				_ = s.updateStep(run, "enrich", sqlite.StatusFailed, err.Error())
				return run.ID, err
			}
			if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
				task = *fresh
			}
		}
		if err := s.transitionTask(task, asagiri.StatusEnriched, force); err != nil {
			_ = s.updateStep(run, "enrich", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
		canonical, _ := payloadToCanonical(task.PayloadJSON)
		contextFiles := s.contextFilesForTask(feature, canonical)
		agentCtx := agent.BuildContext(run.ID, &canonical, contextFiles)
		res, runErr := a.Run(ctx, agent.RunRequest{
			Feature:    feature,
			TaskID:     task.ID,
			Prompt:     "Enrich task " + task.ID + " pour feature " + feature,
			WorkingDir: s.repoRoot,
		})
		agentRes := agent.DryRunResult("enrichissement simulé")
		if parsed, ok := agent.ParseResult(res.Stdout); ok {
			agentRes = parsed
		}
		_ = agent.WriteLogs(s.repoRoot, task.ID, agentCtx, agentRes)
		enriched := defaultEnrichment(s.repoRoot, task, agentName, res, s.dryRun)
		enriched["context_files"] = contextFiles
		body, _ := json.Marshal(enriched)
		if updateErr := s.store.UpdateTask(&sqlite.Task{
			ID:          task.ID,
			PayloadJSON: string(body),
			Status:      asagiri.StatusEnriched,
		}); updateErr != nil {
			_ = s.updateStep(run, "enrich", sqlite.StatusFailed, updateErr.Error())
			return run.ID, updateErr
		}
		if runErr != nil {
			_ = s.updateStep(run, "enrich", sqlite.StatusFailed, runErr.Error())
			return run.ID, runErr
		}
	}

	if err := s.updateStep(run, "enrich", sqlite.StatusDone, "tasks enrichies"); err != nil {
		return run.ID, err
	}
	return run.ID, nil
}

func reportTitle(payload string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return ""
	}
	title, _ := m["title"].(string)
	return title
}

func defaultEnrichment(repoRoot string, task sqlite.Task, agentName string, res agent.RunResult, dryRun bool) map[string]any {
	title := reportTitle(task.PayloadJSON)
	if title == "" {
		title = task.ID
	}
	out := map[string]any{
		"task_id":             task.ID,
		"type":                "implementation",
		"risk":                "medium",
		"recommended_agent":   "cursor",
		"files_scope":         []string{"application/"},
		"validation_commands": validationLinesForRepo(repoRoot),
		"title":               title,
		"source_run_id":       task.RunID,
		"enrichment_command":  res.Command,
	}
	if dryRun {
		out["agent_output"] = res.Stdout
		out["recommended_agent"] = agentName
	} else if res.Stdout != "" {
		out["agent_output"] = res.Stdout
	}
	return out
}

func (s *Service) contextFilesForTask(feature string, task asagiri.Task) []string {
	if s.dryRun {
		return rag.HeuristicContextFiles(s.repoRoot, feature)
	}
	db, err := rag.OpenDB(s.repoRoot)
	if err != nil {
		return rag.HeuristicContextFiles(s.repoRoot, feature)
	}
	defer db.Close()
	paths, err := rag.NewRetriever(db).Search(task.Title, 8)
	if err != nil || len(paths) == 0 {
		return rag.HeuristicContextFiles(s.repoRoot, feature)
	}
	return paths
}

func (s *Service) DevFeature(ctx context.Context, feature, taskID, agentName string, force bool) (string, error) {
	run, err := s.createRun(feature, "dev")
	if err != nil {
		return "", err
	}
	if err := s.updateStep(run, "dev", sqlite.StatusRunning, "implémentation en cours"); err != nil {
		return run.ID, err
	}

	tasks, err := s.pickTasks(feature, taskID)
	if err != nil {
		_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	a, err := s.ensureAgent(agentName)
	if err != nil {
		_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}

	for _, task := range tasks {
		if st := normalizeStatus(task.Status); st == asagiri.StatusPlanned || st == asagiri.StatusPending {
			if err := s.transitionTask(task, asagiri.StatusEnriched, true); err != nil {
				_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
				return run.ID, err
			}
			if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
				task = *fresh
			}
		}
		if err := s.transitionTask(task, asagiri.StatusRunning, force); err != nil {
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
		worktreePath, _, err := s.worktreeMngr.Create(ctx, feature, task.ID)
		if err != nil {
			_ = s.transitionTask(task, asagiri.StatusFailed, true)
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
		if err := s.store.UpdateTask(&sqlite.Task{ID: task.ID, WorktreePath: worktreePath}); err != nil {
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
			return run.ID, err
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
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, runErr.Error())
			return run.ID, runErr
		}

		if err := s.writeTaskLog(task.ID, "dev.log", res.Stdout+"\n"+res.Stderr); err != nil {
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
		if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
			task = *fresh
		}
		if err := s.transitionTask(task, asagiri.StatusImplemented, force); err != nil {
			_ = s.updateStep(run, "dev", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
	}

	if err := s.updateStep(run, "dev", sqlite.StatusDone, "développement terminé"); err != nil {
		return run.ID, err
	}
	return run.ID, nil
}

func (s *Service) writeTaskLog(taskID, fileName, body string) error {
	logDir := filepath.Join(s.repoRoot, ".asagiri", "logs", taskID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create task log dir: %w", err)
	}
	logPath := filepath.Join(logDir, fileName)
	if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write task log: %w", err)
	}
	return nil
}

func (s *Service) VerifyFeature(ctx context.Context, feature, taskID string, force bool) (string, error) {
	run, err := s.createRun(feature, "verify")
	if err != nil {
		return "", err
	}
	if err := s.updateStep(run, "verify", sqlite.StatusRunning, "validation locale"); err != nil {
		return run.ID, err
	}

	tasks, err := s.pickTasks(feature, taskID)
	if err != nil {
		_ = s.updateStep(run, "verify", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	for _, task := range tasks {
		targetDir := s.repoRoot
		if task.WorktreePath != "" {
			targetDir = task.WorktreePath
		}
		if err := s.runVerification(ctx, targetDir, task.PayloadJSON); err != nil {
			_ = s.transitionTask(task, asagiri.StatusVerifyFailed, true)
			_ = s.updateStep(run, "verify", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
		if err := s.transitionTask(task, asagiri.StatusVerified, force); err != nil {
			_ = s.updateStep(run, "verify", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
	}

	if err := s.updateStep(run, "verify", sqlite.StatusDone, "validations passées"); err != nil {
		return run.ID, err
	}
	return run.ID, nil
}

func (s *Service) ReviewFeature(ctx context.Context, feature, taskID, agentName string, force bool) (string, error) {
	run, err := s.createRun(feature, "review")
	if err != nil {
		return "", err
	}
	if err := s.updateStep(run, "review", sqlite.StatusRunning, "review indépendante"); err != nil {
		return run.ID, err
	}

	tasks, err := s.pickTasks(feature, taskID)
	if err != nil {
		_ = s.updateStep(run, "review", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	a, err := s.ensureAgent(agentName)
	if err != nil {
		_ = s.updateStep(run, "review", sqlite.StatusFailed, err.Error())
		return run.ID, err
	}
	for _, task := range tasks {
		_, runErr := a.Run(ctx, agent.RunRequest{
			Feature:    feature,
			TaskID:     task.ID,
			Prompt:     "Review indépendante pour task " + task.ID,
			WorkingDir: s.repoRoot,
		})
		if runErr != nil {
			_ = s.transitionTask(task, asagiri.StatusReviewFailed, true)
			_ = s.updateStep(run, "review", sqlite.StatusFailed, runErr.Error())
			return run.ID, runErr
		}
		if err := s.transitionTask(task, asagiri.StatusReviewed, force); err != nil {
			_ = s.updateStep(run, "review", sqlite.StatusFailed, err.Error())
			return run.ID, err
		}
	}

	if err := s.updateStep(run, "review", sqlite.StatusDone, "review terminée"); err != nil {
		return run.ID, err
	}
	return run.ID, nil
}

func (s *Service) Status(limit int) ([]sqlite.Run, error) {
	return s.store.ListRuns(limit)
}

func (s *Service) ResumeRun(runID string, force bool) (string, error) {
	run, err := s.store.GetRun(runID)
	if err != nil {
		return "", err
	}
	tasks, err := s.store.ListTasksByFeature(run.Feature)
	if err != nil {
		return "", err
	}
	statuses := make([]string, 0, len(tasks))
	for _, t := range tasks {
		statuses = append(statuses, t.Status)
	}
	next := NextWorkflowStep(statuses)
	if next == "" || next == "report" {
		steps := deserializeSteps(run.StepsJSON)
		for _, step := range steps {
			if step.Status == sqlite.StatusPending || step.Status == sqlite.StatusRunning || step.Status == sqlite.StatusFailed {
				if !force && step.Status == sqlite.StatusDone {
					continue
				}
				return step.Name, nil
			}
		}
		return "", nil
	}
	_ = force
	return next, nil
}

// ResumeRunDryExecute simulates continuing from the next workflow step (dry-run only).
func (s *Service) ResumeRunDryExecute(ctx context.Context, runID string, force bool) (string, error) {
	if !s.dryRun {
		return "", fmt.Errorf("reprise automatique réservée au mode --dry-run pour l'instant")
	}
	run, err := s.store.GetRun(runID)
	if err != nil {
		return "", err
	}
	step, err := s.ResumeRun(runID, force)
	if err != nil || step == "" {
		return step, err
	}
	switch step {
	case "plan":
		_, _, err = s.PlanFeature(run.Feature)
	case "enrich":
		_, err = s.EnrichFeature(ctx, run.Feature, "", "ollama", force)
	case "dev":
		_, err = s.DevFeature(ctx, run.Feature, "", "cursor", force)
	case "verify":
		_, err = s.VerifyFeature(ctx, run.Feature, "", force)
	case "review":
		_, err = s.ReviewFeature(ctx, run.Feature, "", "codex", force)
	case "report":
		_, _, err = s.GenerateReport(runID)
	default:
		return step, nil
	}
	return step, err
}

func (s *Service) GenerateReport(runID string) (string, string, error) {
	run, err := s.store.GetRun(runID)
	if err != nil {
		return "", "", err
	}
	tasks, err := s.store.ListTasksByFeature(run.Feature)
	if err != nil {
		return "", "", err
	}
	rawSteps := deserializeSteps(run.StepsJSON)
	steps := make([]report.Step, 0, len(rawSteps))
	for _, step := range rawSteps {
		steps = append(steps, report.Step{
			Name:      step.Name,
			Status:    step.Status,
			Message:   step.Message,
			StartedAt: step.StartedAt,
			EndedAt:   step.EndedAt,
		})
	}
	return s.reportWriter.Write(*run, tasks, steps)
}

func (s *Service) Clean(ctx context.Context, onlyMerged bool, onlyFailed bool) (int, error) {
	runs, err := s.store.ListRuns(200)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, run := range runs {
		tasks, listErr := s.store.ListTasksByRun(run.ID)
		if listErr != nil {
			return removed, listErr
		}
		for _, task := range tasks {
			if task.WorktreePath == "" {
				continue
			}
			if onlyFailed && task.Status != sqlite.StatusFailed {
				continue
			}
			if onlyMerged && task.Status != sqlite.StatusDone && task.Status != sqlite.StatusReviewed {
				continue
			}
			if err := s.worktreeMngr.Remove(ctx, task.WorktreePath); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}

func (s *Service) PreparePR(ctx context.Context, feature string) (string, error) {
	tasks, err := s.store.ListTasksByFeature(feature)
	if err != nil {
		return "", err
	}
	if len(tasks) == 0 {
		return "", fmt.Errorf("aucune tâche pour la feature %q", feature)
	}

	baseDir := filepath.Join(s.repoRoot, ".asagiri", "runs", "pr-"+sanitize(feature))
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", err
	}
	diffPath := filepath.Join(baseDir, "pr.diff")
	checklistPath := filepath.Join(baseDir, "pr-checklist.md")

	var combinedDiff strings.Builder
	for _, task := range tasks {
		if task.WorktreePath == "" || s.dryRun {
			continue
		}
		cmd := osExec.CommandContext(ctx, "git", "-C", task.WorktreePath, "diff")
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			return "", fmt.Errorf("git diff %s: %w", task.ID, cmdErr)
		}
		if combinedDiff.Len() > 0 {
			combinedDiff.WriteString("\n")
		}
		combinedDiff.WriteString("# task: " + task.ID + "\n")
		combinedDiff.Write(out)
	}
	if err := os.WriteFile(diffPath, []byte(combinedDiff.String()), 0o644); err != nil {
		return "", err
	}

	var checklist strings.Builder
	checklist.WriteString("# PR Checklist\n\n")
	checklist.WriteString("- [ ] Scope relu\n")
	checklist.WriteString("- [ ] Tests locaux passés\n")
	checklist.WriteString("- [ ] Lint/format passés\n")
	checklist.WriteString("- [ ] Diff revu\n")
	checklist.WriteString("\n## Tasks\n\n")
	for _, task := range tasks {
		checklist.WriteString(fmt.Sprintf("- [ ] `%s` (%s)\n", task.ID, task.Status))
	}
	if err := os.WriteFile(checklistPath, []byte(checklist.String()), 0o644); err != nil {
		return "", err
	}

	return checklistPath, nil
}
