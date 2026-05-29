package confidence

import "context"

// Dimension is one of the six trust dimensions (spec §7).
type Dimension string

const (
	DimensionArchitecture   Dimension = "architecture"
	DimensionImplementation Dimension = "implementation"
	DimensionFlowIntegrity  Dimension = "flow_integrity"
	DimensionObservability  Dimension = "observability"
	DimensionSecurity       Dimension = "security"
	DimensionRegression     Dimension = "regression"
)

// AllDimensions lists spec §7 dimensions in stable order.
var AllDimensions = []Dimension{
	DimensionArchitecture,
	DimensionImplementation,
	DimensionFlowIntegrity,
	DimensionObservability,
	DimensionSecurity,
	DimensionRegression,
}

// ContributoryCheck is implemented by trust.VerificationCheck in the parent package.
type ContributoryCheck interface {
	CheckConfidence() float64
}

// Report holds aggregated confidence scores and explainability metadata.
type Report struct {
	Architecture   float64  `json:"architecture"`
	Implementation float64  `json:"implementation"`
	FlowIntegrity  float64  `json:"flow_integrity"`
	Observability  float64  `json:"observability"`
	Security       float64  `json:"security"`
	Regression     float64  `json:"regression"`
	Overall            float64  `json:"overall"`
	Limits             []string `json:"limits,omitempty"`
	UncoveredZones     []string `json:"uncovered_zones,omitempty"`
	InferredDimensions []string `json:"inferred_dimensions,omitempty"`
}

// Aggregator combines check outputs into a confidence report (spec §11, §24).
type Aggregator interface {
	Aggregate(ctx context.Context, checks []ContributoryCheck) (Report, error)
}

// RealAggregator runs Scorer → Weighter → Normalizer for lot 2+.
type RealAggregator struct {
	Scorer      Scorer
	Weighter    Weighter
	Normalizer  Normalizer
	HighCritFlow bool
}

// NewRealAggregator returns the default lot-2 confidence pipeline.
func NewRealAggregator() RealAggregator {
	return RealAggregator{
		Scorer:     DefaultScorer{},
		Weighter:   NewDefaultWeighter(),
		Normalizer: ClampNormalizer{},
	}
}

// AggregateDetailed scores from detailed checks (used when checks > 0).
func (a RealAggregator) AggregateDetailed(ctx context.Context, checks []DetailedCheck) (Report, error) {
	if len(checks) == 0 {
		return StubAggregator{}.Aggregate(ctx, nil)
	}
	scorer := a.Scorer
	if scorer == nil {
		scorer = DefaultScorer{}
	}
	weighter := a.Weighter
	if weighter == nil {
		weighter = NewDefaultWeighter()
	}
	normalizer := a.Normalizer
	if normalizer == nil {
		normalizer = ClampNormalizer{}
	}

	raw, err := scorer.Score(ctx, checks)
	if err != nil {
		return Report{}, err
	}

	limits := []string{
		"confidence aggregated from verification checks with per-dimension weighting (spec §7, §11)",
	}
	if a.HighCritFlow {
		limits = append(limits, "high criticality flow: thresholds tightened")
	}

	inferredSet := inferredDimensionsWithoutDedicatedChecks(checks)

	uncovered := make([]string, 0)
	inferred := make([]string, 0)
	var overallNum, overallDen float64
	rep := Report{Limits: limits}

	for _, d := range AllDimensions {
		r, ok := raw[d]
		if !ok {
			uncovered = append(uncovered, string(d)+": no evidence from lot-2 checks")
			continue
		}
		if inferredSet[d] {
			if r > InferredDimensionCap {
				r = InferredDimensionCap
			}
			inferred = append(inferred, string(d))
		}
		if a.HighCritFlow && (d == DimensionFlowIntegrity || d == DimensionRegression) && r < 0.8 {
			r *= 0.95
		}
		n := normalizer.Normalize(r)
		switch d {
		case DimensionArchitecture:
			rep.Architecture = n
		case DimensionImplementation:
			rep.Implementation = n
		case DimensionFlowIntegrity:
			rep.FlowIntegrity = n
		case DimensionObservability:
			rep.Observability = n
		case DimensionSecurity:
			rep.Security = n
		case DimensionRegression:
			rep.Regression = n
		}
		w := weighter.Weight(d)
		overallNum += w * n
		overallDen += w
	}
	if overallDen > 0 {
		rep.Overall = overallNum / overallDen
	}
	rep.UncoveredZones = uncovered
	rep.InferredDimensions = inferred
	return rep, nil
}

// StubAggregator returns zero scores with explicit limits (lot 1 skeleton).
type StubAggregator struct{}

// Aggregate implements Aggregator for lot 1.
func (StubAggregator) Aggregate(_ context.Context, _ []ContributoryCheck) (Report, error) {
	limits := []string{
		"lot 1 skeleton: confidence derived from zero executed checks",
		"no scorer, weighting, or normalization pipeline active yet",
	}
	uncovered := make([]string, len(AllDimensions))
	for i, d := range AllDimensions {
		uncovered[i] = string(d) + ": no evidence (checks not run)"
	}
	return Report{
		Limits:         limits,
		UncoveredZones: uncovered,
	}, nil
}

func inferredDimensionsWithoutDedicatedChecks(checks []DetailedCheck) map[Dimension]bool {
	out := make(map[Dimension]bool, len(InferredDimensions))
	for _, d := range InferredDimensions {
		out[d] = true
	}
	for _, c := range checks {
		if c.Status == checkStatusSkipped {
			continue
		}
		switch c.Type {
		case "observability":
			delete(out, DimensionObservability)
		case "security":
			delete(out, DimensionSecurity)
		}
	}
	return out
}

// ScoreFor returns the score for a dimension from a report.
func (r Report) ScoreFor(d Dimension) float64 {
	switch d {
	case DimensionArchitecture:
		return r.Architecture
	case DimensionImplementation:
		return r.Implementation
	case DimensionFlowIntegrity:
		return r.FlowIntegrity
	case DimensionObservability:
		return r.Observability
	case DimensionSecurity:
		return r.Security
	case DimensionRegression:
		return r.Regression
	default:
		return 0
	}
}
