package investigation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/google/uuid"
)

// RunInvestigation executes the full local-first pipeline (spec-my-A §25.5).
func RunInvestigation(ctx context.Context, req Request, cfg *config.Config) (Report, error) {
	if req.RepoRoot == "" {
		return Report{}, fmt.Errorf("investigation: repo root required")
	}
	if req.Depth == "" {
		req.Depth = DepthStandard
	}
	if req.Symptom == "" && req.Feature != "" {
		req.Symptom = req.Feature
	}
	scope := ResolveScope(req)

	pattern := req.Symptom
	if pattern == "" {
		pattern = req.Feature
	}
	if len(scope.SearchPatterns) > 0 && pattern == "" {
		pattern = scope.SearchPatterns[0]
	}

	feature := req.Feature
	if feature == "" {
		feature = req.Symptom
	}
	local, err := Run(ctx, req.RepoRoot, feature, req.TaskID, cfg)
	if err != nil {
		return Report{}, err
	}

	var graphPack ContextPack
	usedKnowledgeGraph := false
	if gp, ok := EnrichFromKnowledgeGraph(ctx, req.RepoRoot, &scope, &local); ok {
		graphPack = gp
		usedKnowledgeGraph = true
	}

	hyps, evidence := GenerateHypotheses(scope, local, graphPack)
	sort.Slice(hyps, func(i, j int) bool { return hyps[i].Score > hyps[j].Score })

	var candidates []Hypothesis
	for _, h := range hyps {
		if h.Score >= 0.5 {
			candidates = append(candidates, h)
		}
	}
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}

	maxFiles := req.MaxFiles
	if maxFiles <= 0 {
		switch req.Depth {
		case DepthQuick:
			maxFiles = 30
		case DepthDeep:
			maxFiles = 120
		default:
			maxFiles = 80
		}
	}

	rep := Report{
		ID:                  uuid.NewString(),
		CreatedAt:           time.Now().UTC(),
		Request:             req,
		Scope:               scope,
		Evidence:            evidence,
		Hypotheses:          hyps,
		RootCauseCandidates: candidates,
		LocalResult:         local,
		SuggestedActions:    suggestActions(scope, candidates),
		Limits:              []string{"V1 deterministic hypothesis scoring", "No cloud analysis when --no-cloud"},
		Risks:               []string{"Hypotheses are candidates, not confirmed root causes"},
	}
	if req.NoCloud {
		rep.Limits = append(rep.Limits, "Cloud model calls disabled")
	}

	tokens := estimateTokens(local, len(evidence))
	rep.EstimateTokens = tokens
	rep.EstimateCostEUR = float64(tokens) / 1_000_000 * 0.5

	if req.EstimateOnly {
		return rep, nil
	}

	pack, err := BuildContextPack(req.RepoRoot, local, scope, maxFiles)
	if err != nil {
		return rep, err
	}
	if usedKnowledgeGraph {
		pack = MergeGraphScope(pack, graphPack)
	}
	dir := filepath.Join(req.RepoRoot, ".asagiri", "investigations", rep.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return rep, err
	}
	ctxPath, mdPath, err := WriteContextPackArtifacts(dir, rep, pack)
	if err != nil {
		return rep, err
	}
	rep.ContextPackPath = mdPath
	_ = ctxPath
	replayPath, err := WriteReplayPack(dir, rep)
	if err != nil {
		return rep, err
	}
	rep.ReplayPackPath = replayPath
	_, err = WriteReport(req.RepoRoot, rep)
	if err != nil {
		return rep, err
	}
	_, _ = WriteGraph(req.RepoRoot, rep, pack)
	return rep, nil
}

func suggestActions(scope ResolvedScope, candidates []Hypothesis) []string {
	var actions []string
	if scope.Flow != "" {
		actions = append(actions, "Review flow YAML under .asagiri/products")
	}
	if len(candidates) > 0 {
		actions = append(actions, "Validate top hypothesis: "+candidates[0].Statement)
	}
	actions = append(actions, "Run targeted tests from context pack")
	if scope.TaskID != "" {
		actions = append(actions, "Inspect task artefact for "+scope.TaskID)
	}
	return actions
}

func estimateTokens(local InvestigationResult, evidenceCount int) int {
	n := len(local.CandidateFiles)*200 + len(local.GrepHits)*50 + evidenceCount*100
	if n < 2000 {
		return 2000
	}
	return n
}

// FeedMemory stores investigation summary in runtime scoped memory (spec-my-A §25.24).
func FeedMemory(repoRoot string, rep Report) error {
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return err
	}
	defer store.Close()
	summary := rep.Request.Symptom
	if len(rep.RootCauseCandidates) > 0 {
		summary = rep.RootCauseCandidates[0].Statement
	}
	if err := store.UpsertMemory(runtime.MemoryEntry{
		Scope:       runtime.ScopeFeature,
		Type:        "investigation",
		Summary:     summary,
		Source:      "investigation:" + rep.ID,
		Relevance:   0.75,
		Tags:        []string{"investigation", string(rep.Request.Depth)},
		LinkedFlows: []string{rep.Scope.Flow},
	}); err != nil {
		return err
	}
	for _, h := range rep.Hypotheses {
		if h.Score < 0.4 {
			continue
		}
		_ = store.UpsertMemory(runtime.MemoryEntry{
			Scope:       runtime.ScopeFlow,
			Type:        "hypothesis",
			Summary:     h.Statement,
			Source:      "investigation:" + rep.ID,
			Relevance:   h.Score,
			Tags:        []string{"hypothesis"},
			LinkedFlows: []string{rep.Scope.Flow},
		})
	}
	return nil
}
