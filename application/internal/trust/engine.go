package trust

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/replay"
)

// TrustEngine runs the verification pipeline (spec §24).
type TrustEngine interface {
	Verify(ctx context.Context, req VerificationRequest) (VerificationResult, error)
}

// VerificationCheckRunner runs a single verification check (spec §24).
type VerificationCheckRunner interface {
	Run(ctx context.Context, scope VerificationScope) (VerificationCheck, error)
}

// ConfidenceAggregator combines check outputs into a confidence report (spec §24).
type ConfidenceAggregator interface {
	Aggregate(ctx context.Context, checks []VerificationCheck) (confidence.Report, error)
}

// EventEmitter optionally publishes runtime verification events (spec §18, lot 5).
type EventEmitter interface {
	Emit(ctx context.Context, name string, payload map[string]any) error
}

// VerificationRequest scopes a trust verification run.
type VerificationRequest struct {
	Flow       string
	Task       string
	Branch     string
	Strict     bool
	Product    string
	CheckTypes []string
}

// VerificationScope is the resolved execution scope for checks and reporting.
type VerificationScope struct {
	TrustID   string
	Flow      string
	Task      string
	Branch    string
	RepoRoot  string
	ProductID string
}

// VerificationResult is returned by Verify after reports are written.
type VerificationResult struct {
	TrustID  string
	Report   TrustReport
	MDPath   string
	JSONPath string
}

// Engine is the default TrustEngine implementation.
type Engine struct {
	RepoRoot   string
	Registry   *checks.Registry
	Aggregator ConfidenceAggregator
	Gates      GateEvaluator
	Emitter    EventEmitter
	Config     *config.Config
}

type realConfidenceAggregator struct {
	highCritFlow bool
}

func (a realConfidenceAggregator) Aggregate(ctx context.Context, checks []VerificationCheck) (confidence.Report, error) {
	if len(checks) == 0 {
		return confidence.StubAggregator{}.Aggregate(ctx, nil)
	}
	detailed := make([]confidence.DetailedCheck, len(checks))
	for i, c := range checks {
		findings := make([]confidence.FindingInput, len(c.Findings))
		for j, f := range c.Findings {
			findings[j] = confidence.FindingInput{
				Severity: string(f.Severity),
				Category: f.Category,
			}
		}
		detailed[i] = confidence.DetailedCheck{
			Type:       string(c.Type),
			Status:     string(c.Status),
			Confidence: c.Confidence,
			Findings:   findings,
		}
	}
	agg := confidence.NewRealAggregator()
	agg.HighCritFlow = a.highCritFlow
	return agg.AggregateDetailed(ctx, detailed)
}

// NewEngine returns an engine with default lot-2 checks and real confidence when checks run.
func NewEngine(repoRoot string) *Engine {
	deps := checks.DefaultDependencies()
	deps.KnowledgeBlastRadius = knowledgeBlastRadiusFromGraph
	return NewEngineWithChecks(repoRoot, checks.NewDefaultRegistry(deps))
}

func knowledgeBlastRadiusFromGraph(ctx context.Context, repoRoot, flowID string) (checks.BlastRadiusSummary, bool) {
	if strings.TrimSpace(repoRoot) == "" {
		return checks.BlastRadiusSummary{}, false
	}
	if _, err := os.Stat(knowledge.DBPath(repoRoot)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return checks.BlastRadiusSummary{}, false
		}
		return checks.BlastRadiusSummary{}, false
	}
	_, result, err := BlastRadiusFromGraph(ctx, repoRoot, knowledge.ImpactRequest{Flow: flowID})
	if err != nil {
		return checks.BlastRadiusSummary{}, false
	}
	return checks.BlastRadiusSummaryFromImpact(result, flowID), true
}

// NewEngineWithChecks wires a custom registry (empty registry keeps stub confidence).
func NewEngineWithChecks(repoRoot string, reg *checks.Registry) *Engine {
	eng := &Engine{
		RepoRoot:   repoRoot,
		Registry:   reg,
		Aggregator: realConfidenceAggregator{},
		Gates:      NewGateEvaluator(nil),
	}
	if reg == nil || len(reg.Runners()) == 0 {
		eng.Aggregator = stubConfidenceAggregator{}
	}
	return eng
}

// stubConfidenceAggregator implements ConfidenceAggregator via confidence.StubAggregator (lot 1).
type stubConfidenceAggregator struct{}

func (stubConfidenceAggregator) Aggregate(ctx context.Context, checks []VerificationCheck) (confidence.Report, error) {
	contributory := make([]confidence.ContributoryCheck, len(checks))
	for i := range checks {
		contributory[i] = checks[i]
	}
	return confidence.StubAggregator{}.Aggregate(ctx, contributory)
}

// Verify runs the pipeline: scope → checks → confidence → gates → reports.
func (e *Engine) Verify(ctx context.Context, req VerificationRequest) (VerificationResult, error) {
	if e.RepoRoot == "" {
		return VerificationResult{}, fmt.Errorf("trust engine: repo root required")
	}
	trustID := NewTrustID()
	scope, err := ResolveScope(e.RepoRoot, trustID, req)
	if err != nil {
		return VerificationResult{}, err
	}

	if e.Emitter != nil {
		if re, ok := e.Emitter.(*RuntimeEmitter); ok {
			re.flow = scope.Flow
		}
		_ = e.Emitter.Emit(ctx, runtime.EventVerificationStarted, map[string]any{
			"trust_id": trustID,
			"flow":     scope.Flow,
			"branch":   scope.Branch,
		})
	}

	checkScope := checks.Scope{
		TrustID:   scope.TrustID,
		Flow:      scope.Flow,
		Task:      scope.Task,
		Branch:    scope.Branch,
		RepoRoot:  scope.RepoRoot,
		ProductID: scope.ProductID,
	}
	var rawChecks []checks.Check
	if len(req.CheckTypes) > 0 {
		rawChecks, err = e.Registry.RunSelected(ctx, checkScope, req.CheckTypes)
	} else {
		rawChecks, err = e.Registry.RunAll(ctx, checkScope)
	}
	if err != nil {
		return VerificationResult{}, fmt.Errorf("run checks: %w", err)
	}
	checkList := mapRegistryChecks(rawChecks)

	highCrit := flowCriticalityHigh(e.RepoRoot, scope.ProductID, scope.Flow)
	if agg, ok := e.Aggregator.(realConfidenceAggregator); ok {
		agg.highCritFlow = highCrit
		e.Aggregator = agg
	}

	confReport, err := e.Aggregator.Aggregate(ctx, checkList)
	if err != nil {
		return VerificationResult{}, fmt.Errorf("aggregate confidence: %w", err)
	}

	gate := e.Gates.Evaluate(ctx, confReport, checkList)
	blast := mapBlastRadiusReport(rawChecks)
	suggested := SuggestReviews(TrustReport{
		Checks:       checkList,
		Confidence:   confReport,
		BlastRadius:  blast,
		ResidualRisk: ComputeResidualRisk(checkList, confReport),
	})
	report := NewTrustReport(scope, checkList, confReport, gate, blast, suggested)

	mdPath, jsonPath, err := WriteReport(e.RepoRoot, trustID, report)
	if err != nil {
		return VerificationResult{}, err
	}

	manifest := buildReplayManifest(e.RepoRoot, scope, checkList, e.Config)
	if err := replay.WriteReplay(e.RepoRoot, trustID, manifest); err != nil {
		return VerificationResult{}, fmt.Errorf("write replay manifest: %w", err)
	}

	e.emitLifecycleEvents(ctx, scope, checkList, confReport, gate)

	return VerificationResult{
		TrustID:  trustID,
		Report:   report,
		MDPath:   mdPath,
		JSONPath: jsonPath,
	}, nil
}

// ResolveScope maps a request and trust id to a VerificationScope.
func ResolveScope(repoRoot, trustID string, req VerificationRequest) (VerificationScope, error) {
	productID := req.Product
	if productID == "" {
		var err error
		productID, err = ResolveProductID(repoRoot, req.Flow)
		if err != nil {
			return VerificationScope{}, err
		}
	}
	return VerificationScope{
		TrustID:   trustID,
		Flow:      req.Flow,
		Task:      req.Task,
		Branch:    req.Branch,
		RepoRoot:  repoRoot,
		ProductID: productID,
	}, nil
}

// NewTrustID returns a unique trust run identifier (spec §21 example shape).
func NewTrustID() string {
	now := time.Now().UTC()
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("trust-%s-%s", now.Format("2006-01-02"), suffix)
}

func mapBlastRadiusReport(raw []checks.Check) *BlastRadiusReport {
	br := blastRadiusFromChecks(raw)
	if br == nil {
		return nil
	}
	return br
}

func mapRegistryChecks(raw []checks.Check) []VerificationCheck {
	if len(raw) == 0 {
		return []VerificationCheck{}
	}
	out := make([]VerificationCheck, len(raw))
	for i, c := range raw {
		out[i] = VerificationCheck{
			ID:         c.ID,
			Name:       c.Name,
			Type:       CheckType(c.Type),
			Status:     CheckStatus(c.Status),
			Confidence: c.Confidence,
			Findings:   mapCheckFindings(c.Findings),
			Evidence:   mapCheckEvidence(c.Evidence),
			Duration:   c.Duration,
		}
	}
	return out
}

func mapCheckFindings(in []checks.Finding) []Finding {
	if len(in) == 0 {
		return nil
	}
	out := make([]Finding, len(in))
	for i, f := range in {
		out[i] = Finding{
			Severity:     Severity(f.Severity),
			Category:     f.Category,
			Message:      f.Message,
			SuggestedFix: f.SuggestedFix,
		}
	}
	return out
}

func mapCheckEvidence(in []checks.Evidence) []Evidence {
	if len(in) == 0 {
		return nil
	}
	out := make([]Evidence, len(in))
	for i, e := range in {
		out[i] = Evidence{
			Kind:    e.Kind,
			Source:  e.Source,
			Summary: e.Summary,
		}
	}
	return out
}

func flowCriticalityHigh(repoRoot, productID, flowID string) bool {
	if productID == "" || flowID == "" {
		return false
	}
	flowPath := filepath.Join(checks.ProductDir(repoRoot, productID), "flows", flowID+".flow.yaml")
	raw, err := os.ReadFile(flowPath)
	if err != nil {
		return false
	}
	var meta struct {
		Business struct {
			Criticality string `yaml:"criticality"`
		} `yaml:"business"`
	}
	if yaml.Unmarshal(raw, &meta) != nil {
		return false
	}
	return meta.Business.Criticality == "high"
}
