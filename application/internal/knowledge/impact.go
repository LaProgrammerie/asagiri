package knowledge

import (
	"context"
	"fmt"
	"strings"
)

// ImpactRequest selects an impact analysis target.
type ImpactRequest struct {
	File   string
	Flow   string
	Action string
}

// ImpactResult summarizes blast radius (spec-my-E §14).
type ImpactResult struct {
	Input             string   `json:"input"`
	ImpactedFlows     []string `json:"impacted_flows,omitempty"`
	ImpactedAPIs      []string `json:"impacted_apis,omitempty"`
	ImpactedEvents    []string `json:"impacted_events,omitempty"`
	ImpactedTests     []string `json:"impacted_tests,omitempty"`
	Risk              string   `json:"risk,omitempty"`
	RecommendedChecks []string `json:"recommended_checks,omitempty"`
}

// ImpactAnalyzer analyzes change impact via the knowledge graph.
type ImpactAnalyzer interface {
	Analyze(ctx context.Context, req ImpactRequest) (ImpactResult, error)
}

// graphStoreImpactAnalyzer runs analysis on an already-open store (CLI/tests).
type graphStoreImpactAnalyzer struct {
	Store GraphStore
}

// NewImpactAnalyzer returns an analyzer backed by store.
func NewImpactAnalyzer(store GraphStore) ImpactAnalyzer {
	return &graphStoreImpactAnalyzer{Store: store}
}

// Analyze implements ImpactAnalyzer using flow/file scope resolution.
func (a *graphStoreImpactAnalyzer) Analyze(ctx context.Context, req ImpactRequest) (ImpactResult, error) {
	if a.Store == nil {
		return ImpactResult{}, fmt.Errorf("impact analyze: store required")
	}
	if err := validateImpactRequest(req); err != nil {
		return ImpactResult{}, err
	}

	var scope FlowScopeResult
	var input string
	var err error

	switch {
	case strings.TrimSpace(req.File) != "":
		input = normalizeImpactPath(req.File)
		scope, err = ResolveFileScope(ctx, a.Store, input)
	case strings.TrimSpace(req.Flow) != "":
		if err := validateFlowActionLink(ctx, a.Store, req.Flow, req.Action); err != nil {
			return ImpactResult{}, err
		}
		input = req.Flow + " / " + req.Action
		scope, err = ResolveFlowScope(ctx, a.Store, FlowScopeRequest{
			Flow:   req.Flow,
			Action: req.Action,
		})
	default:
		return ImpactResult{}, fmt.Errorf("impact analyze: --file or both --flow and --action are required")
	}
	if err != nil {
		return ImpactResult{}, err
	}

	return ImpactResult{
		Input:             input,
		ImpactedFlows:     scope.Flows,
		ImpactedAPIs:      scope.APIs,
		ImpactedEvents:    scope.Events,
		ImpactedTests:     scope.Tests,
		Risk:              impactRisk(scope),
		RecommendedChecks: recommendedChecks(scope, req.Flow),
	}, nil
}

func impactRisk(scope FlowScopeResult) string {
	score := len(scope.Flows) + len(scope.APIs) + len(scope.Events)
	if len(scope.Tests) == 0 && (len(scope.APIs) > 0 || len(scope.Events) > 0) {
		score += 2
	}
	switch {
	case score >= 6:
		return "high"
	case score >= 3:
		return "medium"
	default:
		return "low"
	}
}

func recommendedChecks(scope FlowScopeResult, flow string) []string {
	var checks []string
	for _, test := range scope.Tests {
		checks = append(checks, "go test -run "+test)
	}
	flowName := strings.TrimSpace(flow)
	if flowName == "" && len(scope.Flows) > 0 {
		if parts := strings.Split(scope.Flows[0], " / "); len(parts) > 0 {
			flowName = parts[0]
		}
	}
	if flowName != "" {
		checks = append(checks, "asa verify trust --flow "+flowName)
	}
	return sortedKeys(sliceToSet(checks))
}

func sliceToSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	return set
}

// StubImpactAnalyzer is a lot-1 placeholder.
type StubImpactAnalyzer struct{}

func (StubImpactAnalyzer) Analyze(_ context.Context, _ ImpactRequest) (ImpactResult, error) {
	return ImpactResult{}, ErrNotImplemented
}

func validateImpactRequest(req ImpactRequest) error {
	hasFile := strings.TrimSpace(req.File) != ""
	hasFlow := strings.TrimSpace(req.Flow) != ""
	hasAction := strings.TrimSpace(req.Action) != ""
	if hasFile && (hasFlow || hasAction) {
		return fmt.Errorf("impact analyze: use --file or --flow with --action, not both")
	}
	if !hasFile && (!hasFlow || !hasAction) {
		return fmt.Errorf("impact analyze: --file or both --flow and --action are required")
	}
	return nil
}

func validateFlowActionLink(ctx context.Context, store GraphStore, flow, action string) error {
	flowID := NodeID(NodeTypeFlow, strings.TrimSpace(flow))
	actionID := NodeID(NodeTypeAction, strings.TrimSpace(action))
	if _, err := store.GetNode(ctx, flowID); err != nil {
		return fmt.Errorf("impact analyze: flow %q: %w", flow, err)
	}
	if _, err := store.GetNode(ctx, actionID); err != nil {
		return fmt.Errorf("impact analyze: action %q: %w", action, err)
	}
	edges, err := store.ListEdges(ctx, EdgeFilter{
		FromNodeID: flowID,
		ToNodeID:   actionID,
		Type:       EdgeTypeRequires,
	})
	if err != nil {
		return err
	}
	if len(edges) == 0 {
		return fmt.Errorf("impact analyze: flow %q does not require action %q", flow, action)
	}
	return nil
}

func normalizeImpactPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	return filepathSlash(p)
}

// FormatImpactAnalysis renders terminal output (spec-my-E §14).
func FormatImpactAnalysis(r ImpactResult) string {
	var b strings.Builder
	b.WriteString("Impact Analysis\n")
	b.WriteString("───────────────\n")
	_, _ = fmt.Fprintf(&b, "Input: %s\n", r.Input)
	writeImpactList(&b, "Impacted flows:", r.ImpactedFlows)
	writeImpactList(&b, "Impacted APIs:", r.ImpactedAPIs)
	writeImpactList(&b, "Impacted events:", r.ImpactedEvents)
	writeImpactList(&b, "Impacted tests:", r.ImpactedTests)
	b.WriteString("Risk:\n")
	if r.Risk == "" {
		b.WriteString("low\n")
	} else {
		_, _ = fmt.Fprintf(&b, "%s\n", r.Risk)
	}
	b.WriteString("Recommended checks:\n")
	if len(r.RecommendedChecks) == 0 {
		b.WriteString("- review impacted graph nodes manually\n")
	} else {
		for _, check := range r.RecommendedChecks {
			_, _ = fmt.Fprintf(&b, "- %s\n", check)
		}
	}
	return b.String()
}

func writeImpactList(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString(title)
	b.WriteByte('\n')
	for _, item := range items {
		_, _ = fmt.Fprintf(b, "- %s\n", item)
	}
}
