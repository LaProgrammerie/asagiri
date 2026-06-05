package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/version"
)

var graphArtifactNames = []string{
	"execution-graph.yaml",
	"execution-graph.json",
	"events.jsonl",
	"timeline.jsonl",
	"metrics.json",
	"plan.md",
}

// CaptureContext holds mutable state during package creation.
type CaptureContext struct {
	RepoRoot    string
	ReplayID    string
	ReplayDir   string
	Policies    CapturePolicies
	Manifest    Manifest
	Provenance  []ProvenanceRecord
	ArtifactRel []string
}

func replayDir(repoRoot, replayID string) string {
	return filepath.Join(repoRoot, RelDir, replayID)
}

func ensureReplayLayout(dir string) error {
	for _, sub := range []string{"context", "prompts", "outputs", "graph", "trust", "investigations", "runtime", "reports"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return fmt.Errorf("create replay subdir %q: %w", sub, err)
		}
	}
	return nil
}

// CapturePackage creates the replay directory and captures artefacts (spec §8).
func CapturePackage(ctx context.Context, req ReplayCreateRequest) (ReplayPackage, error) {
	_ = ctx
	if req.RepoRoot == "" {
		return ReplayPackage{}, fmt.Errorf("replay capture: repo root required")
	}
	if req.FromRun == "" && req.FromGraph == "" && req.FromInvestigation == "" {
		return ReplayPackage{}, ErrInvalidSource
	}

	policies := req.Config
	if policies == (CapturePolicies{}) {
		policies = DefaultCapturePolicies(nil)
	}

	replayID := NewReplayID()
	dir := replayDir(req.RepoRoot, replayID)
	if err := ensureReplayLayout(dir); err != nil {
		return ReplayPackage{}, err
	}

	commit, _ := bootstrap.GitHead(req.RepoRoot)
	branch, _ := gitBranch(req.RepoRoot)
	runtimeMode := "guided"
	if cfg, err := config.Load(config.ConfigPath(req.RepoRoot), req.RepoRoot); err == nil && cfg.Runtime.Mode != "" {
		runtimeMode = cfg.Runtime.Mode
	}

	cap := CaptureContext{
		RepoRoot:  req.RepoRoot,
		ReplayID:  replayID,
		ReplayDir: dir,
		Policies:  policies,
		Manifest: Manifest{
			ID:        replayID,
			CreatedAt: time.Now().UTC(),
			Source: SourceRef{
				Run:           strings.TrimSpace(req.FromRun),
				Graph:         strings.TrimSpace(req.FromGraph),
				Investigation: strings.TrimSpace(req.FromInvestigation),
			},
			Repo: RepoRef{Commit: commit, Branch: branch},
			Runtime: RuntimeRef{
				AsagiriVersion: version.Version,
				RuntimeMode:    runtimeMode,
			},
			Agents:   defaultAgents(cfgAgents(req.RepoRoot)),
			Policies: defaultPolicySnapshot(req.RepoRoot),
		},
	}

	if req.FromGraph != "" {
		if err := cap.captureGraph(req.FromGraph); err != nil {
			return ReplayPackage{}, err
		}
	}
	if req.FromRun != "" {
		if err := cap.captureRun(req.FromRun); err != nil {
			return ReplayPackage{}, err
		}
	}
	if req.FromInvestigation != "" {
		if err := cap.captureInvestigation(req.FromInvestigation); err != nil {
			return ReplayPackage{}, err
		}
	}
	if req.IncludeRuntime || policies.CaptureRuntimeEvents {
		if err := cap.captureRuntimeEvents(req.FromGraph); err != nil {
			return ReplayPackage{}, err
		}
	}
	if req.IncludePrompts || policies.CapturePrompts {
		if err := cap.capturePrompts(req.FromRun, req.FromGraph); err != nil {
			return ReplayPackage{}, err
		}
	}
	if req.IncludeEvents || policies.CaptureRuntimeEvents {
		if err := cap.captureEvents(req.FromGraph); err != nil {
			return ReplayPackage{}, err
		}
	}
	if err := cap.captureTrust(req.FromGraph); err != nil {
		return ReplayPackage{}, err
	}
	if err := cap.captureHandoffs(); err != nil {
		return ReplayPackage{}, err
	}
	if policies.CaptureAgentOutputs {
		if err := cap.captureOutputs(req.FromRun, req.FromGraph); err != nil {
			return ReplayPackage{}, err
		}
	}

	if policies.RedactSecrets {
		if err := redactReplayTree(dir); err != nil {
			return ReplayPackage{}, err
		}
	}

	if policies.CompressLargeFiles {
		if _, err := CompressLargeFiles(dir, []string{"prompts", "runtime"}, policies.CompressThresholdBytes); err != nil {
			return ReplayPackage{}, err
		}
	}

	cap.Manifest.Artifacts = cap.ArtifactRel
	if err := writeManifest(dir, cap.Manifest); err != nil {
		return ReplayPackage{}, err
	}
	if err := writeProvenanceIndex(dir, cap.ReplayID, cap.Provenance); err != nil {
		return ReplayPackage{}, err
	}
	if err := WriteBaselineHashes(dir, replayID); err != nil {
		return ReplayPackage{}, err
	}

	return ReplayPackage{
		ID:       replayID,
		Path:     dir,
		Manifest: cap.Manifest,
	}, nil
}

func cfgAgents(repoRoot string) map[string]string {
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil || cfg == nil {
		return nil
	}
	out := map[string]string{}
	if cfg.Work.DefaultAgent != "" {
		out["implementer"] = cfg.Work.DefaultAgent
	}
	if cfg.Work.DefaultReviewer != "" {
		out["reviewer"] = cfg.Work.DefaultReviewer
	}
	out["validator"] = "local"
	return out
}

func defaultAgents(from map[string]string) map[string]string {
	if len(from) == 0 {
		return map[string]string{
			"implementer": "cursor",
			"reviewer":    "codex",
			"validator":   "local",
		}
	}
	return from
}

func defaultPolicySnapshot(repoRoot string) map[string]any {
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil || cfg == nil {
		return map[string]any{"strict_trust": true, "max_parallel": 2}
	}
	return map[string]any{
		"strict_trust": cfg.ExecutionGraph.Gates.TrustRequiredForHighRisk,
		"max_parallel": cfg.ExecutionGraph.MaxParallel,
	}
}

func (c *CaptureContext) captureGraph(graphID string) error {
	src := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID)
	for _, name := range graphArtifactNames {
		if err := c.copyArtifact(src, filepath.Join("graph", name), "execution-graph", name, graphID); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	cpSrc := filepath.Join(src, "checkpoints")
	cpDst := filepath.Join(c.ReplayDir, "graph", "checkpoints")
	if err := copyDirIfExists(cpSrc, cpDst, func(rel string) {
		c.ArtifactRel = append(c.ArtifactRel, filepath.ToSlash(filepath.Join("graph", "checkpoints", rel)))
		c.Provenance = append(c.Provenance, NewProvenanceRecord(
			c.ReplayID,
			filepath.ToSlash(filepath.Join("graph", "checkpoints", rel)),
			"execution-graph",
			filepath.ToSlash(filepath.Join(cpSrc, rel)),
			c.Manifest.Repo.Commit,
			graphID,
		))
	}); err != nil {
		return err
	}
	return nil
}

func (c *CaptureContext) captureRun(runID string) error {
	src := filepath.Join(c.RepoRoot, ".asagiri", "runs", runID)
	for _, name := range []string{"report.md", "report.json", "context-pack.md"} {
		if err := c.copyArtifact(src, filepath.Join("reports", name), "workflow", name, runID); err != nil && !os.IsNotExist(err) {
			return err
		}
		if name == "context-pack.md" {
			_ = c.copyArtifact(src, filepath.Join("context", name), "workflow", name, runID)
		}
	}
	return nil
}

func (c *CaptureContext) captureInvestigation(invID string) error {
	src := filepath.Join(c.RepoRoot, ".asagiri", "investigations", invID)
	for _, name := range []string{"report.md", "report.json", "context-pack.md", "replay-pack.json", "graph.json"} {
		dstSub := "investigations"
		if name == "context-pack.md" {
			_ = c.copyArtifact(src, filepath.Join("context", name), "investigation", name, invID)
		}
		if err := c.copyArtifact(src, filepath.Join(dstSub, name), "investigation", name, invID); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (c *CaptureContext) captureRuntimeEvents(graphID string) error {
	if graphID == "" {
		return nil
	}
	src := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID)
	for _, name := range []string{"events.jsonl", "timeline.jsonl", "metrics.json"} {
		_ = c.copyArtifact(src, filepath.Join("runtime", name), "runtime", name, graphID)
	}
	return nil
}

func (c *CaptureContext) capturePrompts(runID, graphID string) error {
	if runID != "" {
		src := filepath.Join(c.RepoRoot, ".asagiri", "runs", runID)
		_ = c.copyArtifact(src, filepath.Join("prompts", "agent-prompts.jsonl"), "workflow", "agent-prompts.jsonl", runID)
	}
	if graphID != "" {
		src := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID)
		_ = c.copyArtifact(src, filepath.Join("prompts", "plan.md"), "execution-graph", "plan.md", graphID)
	}
	return nil
}

func (c *CaptureContext) captureEvents(graphID string) error {
	if graphID == "" {
		return nil
	}
	src := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID)
	return c.copyArtifact(src, filepath.Join("runtime", "events.jsonl"), "runtime", "events.jsonl", graphID)
}

func (c *CaptureContext) captureTrust(graphID string) error {
	trustRoot := filepath.Join(c.RepoRoot, ".asagiri", "trust")
	entries, err := os.ReadDir(trustRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	limit := 3
	copied := 0
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "trust-") {
			continue
		}
		trustID := e.Name()
		src := filepath.Join(trustRoot, trustID)
		for _, name := range []string{"report.md", "report.json", "replay.yaml"} {
			dst := filepath.Join("trust", trustID, name)
			if err := c.copyArtifact(src, dst, "trust-engine", name, trustID); err == nil {
				copied++
			}
		}
		if copied >= limit {
			break
		}
	}
	return nil
}

func (c *CaptureContext) captureHandoffs() error {
	handoffRoot := filepath.Join(c.RepoRoot, ".asagiri", "handoffs")
	entries, err := os.ReadDir(handoffRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		src := filepath.Join(handoffRoot, e.Name(), "handoff.yaml")
		dst := filepath.Join("context", "handoffs", e.Name(), "handoff.yaml")
		if err := c.copyArtifact(filepath.Dir(src), dst, "coordination", "handoff.yaml", e.Name()); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (c *CaptureContext) captureOutputs(runID, graphID string) error {
	if runID != "" {
		src := filepath.Join(c.RepoRoot, ".asagiri", "runs", runID)
		for _, name := range []string{"report.md", "report.json", "context-pack.md", "agent-prompts.jsonl"} {
			if err := c.copyArtifact(src, filepath.Join("outputs", name), "workflow", name, runID); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	if graphID != "" {
		cpSrc := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID, "checkpoints")
		cpDst := filepath.Join(c.ReplayDir, "outputs", "checkpoints")
		if err := copyDirIfExists(cpSrc, cpDst, func(rel string) {
			relOut := filepath.ToSlash(filepath.Join("outputs", "checkpoints", rel))
			c.ArtifactRel = append(c.ArtifactRel, relOut)
		}); err != nil {
			return err
		}
		if err := c.captureGraphNodeOutputs(graphID); err != nil {
			return err
		}
	}
	return nil
}

func (c *CaptureContext) captureGraphNodeOutputs(graphID string) error {
	graphPath := filepath.Join(c.RepoRoot, ".asagiri", "graphs", graphID, "execution-graph.json")
	body, err := os.ReadFile(graphPath)
	if err != nil {
		return nil
	}
	var graph struct {
		Nodes []struct {
			ID      string   `json:"id"`
			Outputs []string `json:"outputs"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(body, &graph); err != nil {
		return nil
	}
	for _, n := range graph.Nodes {
		for _, outPath := range n.Outputs {
			outPath = strings.TrimSpace(outPath)
			if outPath == "" {
				continue
			}
			abs := outPath
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(c.RepoRoot, outPath)
			}
			relName := filepath.Base(abs)
			dst := filepath.Join("outputs", "nodes", n.ID, relName)
			if err := c.copyArtifact(filepath.Dir(abs), dst, "execution-graph", relName, graphID); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func (c *CaptureContext) copyArtifact(srcDir, relDst, producedBy, fileName, sourceID string) error {
	src := filepath.Join(srcDir, fileName)
	body, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dst := filepath.Join(c.ReplayDir, relDst)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, body, 0o644); err != nil {
		return err
	}
	rel := filepath.ToSlash(relDst)
	c.ArtifactRel = append(c.ArtifactRel, rel)
	c.Provenance = append(c.Provenance, NewProvenanceRecord(
		c.ReplayID,
		rel,
		producedBy,
		filepath.ToSlash(filepath.Join(srcDir, fileName)),
		c.Manifest.Repo.Commit,
		sourceID,
	))
	return nil
}

func writeManifest(dir string, m Manifest) error {
	body, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal replay manifest: %w", err)
	}
	path := filepath.Join(dir, ManifestName)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write replay manifest: %w", err)
	}
	return nil
}

func writeProvenanceIndex(dir, replayID string, records []ProvenanceRecord) error {
	idx := ProvenanceIndex{ReplayID: replayID, Records: records}
	body, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "context", "provenance.json")
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func redactReplayTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(path, gzipSuffix) || path == filepath.Join(dir, ManifestName) {
			return nil
		}
		if !isTextArtifact(path) {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		redacted := RedactSecrets(string(body))
		if ShouldRedactFile(filepath.Base(path)) {
			redacted = redactedPlaceholder + "\n"
		}
		return os.WriteFile(path, []byte(redacted), 0o644)
	})
}

func copyDirIfExists(src, dst string, onFile func(rel string)) error {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return err
		}
		if onFile != nil {
			onFile(rel)
		}
		return nil
	})
}

func isTextArtifact(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".json", ".jsonl", ".yaml", ".yml", ".txt", ".env":
		return true
	default:
		return strings.Contains(filepath.Base(path), ".")
	}
}

// LoadPackage reads a replay package from disk.
func LoadPackage(repoRoot, replayID string) (ReplayPackage, error) {
	if err := ValidateReplayID(replayID); err != nil {
		return ReplayPackage{}, err
	}
	dir := replayDir(repoRoot, replayID)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return ReplayPackage{}, fmt.Errorf("%w: %s", ErrReplayNotFound, replayID)
		}
		return ReplayPackage{}, err
	}
	body, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		return ReplayPackage{}, fmt.Errorf("read replay manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(body, &m); err != nil {
		return ReplayPackage{}, fmt.Errorf("parse replay manifest: %w", err)
	}
	if m.ID == "" {
		m.ID = replayID
	}
	return ReplayPackage{ID: replayID, Path: dir, Manifest: m}, nil
}
