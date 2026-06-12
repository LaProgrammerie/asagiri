package intent

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

// BuildSnapshot loads feature/run state from disk and SQLite.
func BuildSnapshot(repoRoot string, cfg *config.Config, store *sqlite.Store) (StateSnapshot, error) {
	features := map[string]*FeatureState{}

	addFeature := func(name, specPath, status string) {
		name = slugFeature(name)
		if name == "" {
			return
		}
		fs, ok := features[name]
		if !ok {
			fs = &FeatureState{Name: name, Status: status}
			features[name] = fs
		}
		if specPath != "" {
			fs.SpecPath = specPath
			fs.HasLocalSpec = true
		}
		if status != "" {
			fs.Status = status
		}
	}

	if cfg != nil {
		for _, rel := range cfg.Sources.Local.Paths {
			scanFeatureDirs(repoRoot, rel, addFeature)
		}
		scanFeatureDirs(repoRoot, cfg.Specs.KiroPath, addFeature)
	} else {
		scanFeatureDirs(repoRoot, ".asagiri/specs", addFeature)
		scanFeatureDirs(repoRoot, ".kiro/specs", addFeature)
	}

	for name, fs := range features {
		tasksDir := filepath.Join(repoRoot, ".asagiri", "tasks", name)
		entries, _ := os.ReadDir(tasksDir)
		for _, e := range entries {
			if e.IsDir() || (!strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".json")) {
				continue
			}
			fs.HasTasks = true
			fs.TaskCount++
		}
		specDir := filepath.Join(repoRoot, ".asagiri", "specs", name)
		if meta := readMetadataStatus(specDir); meta != "" {
			fs.Status = meta
		}
		if src := readSourceJSON(specDir); src != "" {
			fs.SourceType = src
		}
		if fs.NextTaskID == "" {
			fs.NextTaskID, fs.NextTaskStatus = pickNextTask(repoRoot, name)
		}
		_ = fs
	}

	runs, _ := store.ListRuns(50)
	runStates := make([]RunState, 0, len(runs))
	var activeFeature string
	var latest time.Time
	for _, r := range runs {
		rs := RunState{
			ID:        r.ID,
			Feature:   r.Feature,
			Status:    r.Status,
			UpdatedAt: r.UpdatedAt.UTC().Format(time.RFC3339),
			Resumable: isResumableRun(r.Status),
		}
		runStates = append(runStates, rs)
		if r.UpdatedAt.After(latest) {
			latest = r.UpdatedAt
			activeFeature = r.Feature
		}
		tasks, err := store.ListTasksByFeature(r.Feature)
		if err == nil {
			mergeTasksIntoFeature(features, r.Feature, tasks, repoRoot, cfg)
		}
	}

	featList := make([]FeatureState, 0, len(features))
	for _, f := range features {
		featList = append(featList, *f)
	}
	sort.Slice(featList, func(i, j int) bool { return featList[i].Name < featList[j].Name })

	return StateSnapshot{
		Features:      featList,
		Runs:          runStates,
		ActiveFeature: activeFeature,
	}, nil
}

// RefreshFeatureTaskState reloads task-derived fields for one feature after a mutating executor step.
func RefreshFeatureTaskState(repoRoot string, cfg *config.Config, store *sqlite.Store, feature string, base FeatureState) FeatureState {
	if store == nil || strings.TrimSpace(feature) == "" {
		return base
	}
	tasks, err := store.ListTasksByFeature(feature)
	if err != nil || len(tasks) == 0 {
		return base
	}
	fs := base
	fs.Name = feature
	fs.HasTasks = true
	fs.TaskCount = len(tasks)
	applyTaskGateFields(&fs, tasks, repoRoot, cfg)
	return fs
}

func applyTaskGateFields(fs *FeatureState, tasks []sqlite.Task, repoRoot string, cfg *config.Config) {
	fs.NextTaskID, fs.NextTaskStatus = nextFromSQLiteTasks(tasks)
	fs.PendingGate = nil
	fs.EnrichGateBlocksDev = false
	fs.VerifyEvidenceGateBlocksReview = false
	fs.TrustGateBlocksReview = false
	if cfg == nil || fs.NextTaskID == "" {
		return
	}
	for _, t := range tasks {
		if t.ID != fs.NextTaskID {
			continue
		}
		fs.EnrichGateBlocksDev = gates.EnrichGateBlocksDev(cfg, t.Status, t.PayloadJSON)
		fs.VerifyEvidenceGateBlocksReview = gates.VerifyEvidenceGateBlocksReview(cfg, t.Status, t.PayloadJSON)
		fs.TrustGateBlocksReview = gates.TrustGateBlocksReview(cfg, t.Status, t.PayloadJSON)
		if pg, ok := gates.BlockingPendingForTask(repoRoot, cfg, t); ok {
			pgCopy := pg
			fs.PendingGate = &pgCopy
		}
		break
	}
}

func mergeTasksIntoFeature(features map[string]*FeatureState, feature string, tasks []sqlite.Task, repoRoot string, cfg *config.Config) {
	fs, ok := features[feature]
	if !ok {
		fs = &FeatureState{Name: feature}
		features[feature] = fs
	}
	if len(tasks) > 0 {
		fs.HasTasks = true
		fs.TaskCount = len(tasks)
	}
	applyTaskGateFields(fs, tasks, repoRoot, cfg)
}

func nextFromSQLiteTasks(tasks []sqlite.Task) (string, string) {
	priority := []string{
		asagiri.StatusRunning,
		asagiri.StatusImplemented,
		asagiri.StatusVerifyFailed,
		asagiri.StatusVerified,
		asagiri.StatusReviewFailed,
		asagiri.StatusEnriched,
		asagiri.StatusPending,
		asagiri.StatusPlanned,
	}
	for _, want := range priority {
		for _, t := range tasks {
			if t.Status == want || (want == asagiri.StatusImplemented && t.Status == sqlite.StatusDone) {
				return t.ID, t.Status
			}
		}
	}
	if len(tasks) > 0 {
		return tasks[0].ID, tasks[0].Status
	}
	return "", ""
}

func pickNextTask(repoRoot, feature string) (string, string) {
	dir := filepath.Join(repoRoot, ".asagiri", "tasks", feature)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}
	type pair struct {
		id, status string
	}
	var tasks []pair
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var t asagiri.Task
		if err := yaml.Unmarshal(data, &t); err != nil {
			continue
		}
		tasks = append(tasks, pair{t.ID, t.Status})
	}
	priority := []string{
		asagiri.StatusImplemented,
		asagiri.StatusEnriched,
		asagiri.StatusPending,
		asagiri.StatusPlanned,
	}
	for _, want := range priority {
		for _, p := range tasks {
			if p.status == want {
				return p.id, p.status
			}
		}
	}
	if len(tasks) > 0 {
		return tasks[0].id, tasks[0].status
	}
	return "", ""
}

func scanFeatureDirs(repoRoot, rel string, add func(name, specPath, status string)) {
	root := filepath.Join(repoRoot, filepath.Clean(rel))
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		dir := filepath.Join(root, name)
		status := "ready"
		if _, err := os.Stat(filepath.Join(dir, "spec.md")); err == nil {
			add(name, dir, status)
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "requirements.md")); err == nil {
			add(name, dir, status)
		}
	}
}

func readMetadataStatus(specDir string) string {
	path := filepath.Join(specDir, "metadata.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var meta struct {
		Status string `yaml:"status"`
	}
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return ""
	}
	return meta.Status
}

func readSourceJSON(specDir string) string {
	path := filepath.Join(specDir, "source.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var src struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &src); err != nil {
		return ""
	}
	return src.Type
}

func isResumableRun(status string) bool {
	switch status {
	case sqlite.StatusFailed, sqlite.StatusRunning, sqlite.StatusPending:
		return true
	default:
		return false
	}
}

func slugFeature(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return strings.Trim(s, "-")
}

// FindFeature fuzzy-matches name against snapshot features.
func FindFeature(snapshot StateSnapshot, name string) (FeatureState, float64, bool) {
	name = slugFeature(name)
	if name == "" {
		return FeatureState{}, 0, false
	}
	var best FeatureState
	var bestScore float64
	for _, f := range snapshot.Features {
		score := fuzzyScore(name, f.Name)
		if score > bestScore {
			bestScore = score
			best = f
		}
		if f.Name == name {
			return f, 1.0, true
		}
	}
	if bestScore >= 0.6 {
		return best, bestScore, true
	}
	return FeatureState{}, 0, false
}

func fuzzyScore(a, b string) float64 {
	if a == b {
		return 1
	}
	if strings.Contains(b, a) || strings.Contains(a, b) {
		return 0.85
	}
	return 0
}

// ListFeatureNames returns all known feature slugs.
func ListFeatureNames(snapshot StateSnapshot) []string {
	out := make([]string, 0, len(snapshot.Features))
	for _, f := range snapshot.Features {
		out = append(out, f.Name)
	}
	return out
}

// FindResumableRun picks highest-priority resumable run (spec §4.2).
func FindResumableRun(snapshot StateSnapshot, feature, runID string) (*RunState, error) {
	if runID != "" {
		for i := range snapshot.Runs {
			if snapshot.Runs[i].ID == runID {
				return &snapshot.Runs[i], nil
			}
		}
		return nil, sql.ErrNoRows
	}
	if feature != "" {
		for i := range snapshot.Runs {
			if snapshot.Runs[i].Feature == feature && snapshot.Runs[i].Resumable {
				return &snapshot.Runs[i], nil
			}
		}
	}
	for i := range snapshot.Runs {
		if snapshot.Runs[i].Resumable {
			return &snapshot.Runs[i], nil
		}
	}
	return nil, sql.ErrNoRows
}
