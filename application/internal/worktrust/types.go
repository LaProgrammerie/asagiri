package worktrust

// ReportVersion is the schema version for WorkTrustReport JSON.
const ReportVersion = "1"

// Verdict is the UX trust level for a task, feature, or run.
type Verdict string

const (
	VerdictTrusted    Verdict = "trusted"
	VerdictAcceptable Verdict = "acceptable"
	VerdictRisky      Verdict = "risky"
	VerdictBlocked    Verdict = "blocked"
)

// DimensionID identifies a fixed trust dimension in V1.
type DimensionID string

const (
	DimSpecificationAlignment DimensionID = "specification_alignment"
	DimImplementationQuality  DimensionID = "implementation_quality"
	DimValidationStrength     DimensionID = "validation_strength"
	DimGateConfidence         DimensionID = "gate_confidence"
	DimHumanConfidence        DimensionID = "human_confidence"
	DimResidualRisk           DimensionID = "residual_risk"
)

// DimensionStatus describes how strong a dimension score is.
type DimensionStatus string

const (
	DimStatusStrong      DimensionStatus = "strong"
	DimStatusModerate    DimensionStatus = "moderate"
	DimStatusWeak        DimensionStatus = "weak"
	DimStatusUnevaluated DimensionStatus = "unevaluated"
	DimStatusFailed      DimensionStatus = "failed"
)

// UnevaluatedScore marks a dimension with no applicable signal.
const UnevaluatedScore = -1

// WorkTrustReport is the V1 synthesis report for work gates (read-only).
type WorkTrustReport struct {
	ReportVersion  string                  `json:"report_version"`
	Scope          TrustScope              `json:"scope"`
	GeneratedAt    string                  `json:"generated_at"`
	Score          WorkTrustScore          `json:"score"`
	Dimensions     []WorkTrustDimension    `json:"dimensions"`
	Findings       []WorkTrustFinding      `json:"findings"`
	Evidences      []WorkTrustEvidence     `json:"evidences"`
	Recommendation WorkTrustRecommendation `json:"recommendation"`
}

// TrustScope identifies what was synthesized.
type TrustScope struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Feature string `json:"feature,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
	Status  string `json:"status,omitempty"`
}

// WorkTrustScore is the aggregate score and verdict.
type WorkTrustScore struct {
	Overall float64 `json:"overall"`
	Verdict Verdict `json:"verdict"`
	Summary string  `json:"summary"`
}

// WorkTrustDimension is one scored axis.
type WorkTrustDimension struct {
	ID          DimensionID     `json:"id"`
	Label       string          `json:"label"`
	Score       float64         `json:"score"`
	Status      DimensionStatus `json:"status"`
	Summary     string          `json:"summary"`
	SourceGates []string        `json:"source_gates,omitempty"`
}

// WorkTrustFinding is a structured risk or issue.
type WorkTrustFinding struct {
	Code     string   `json:"code"`
	Severity string   `json:"severity"`
	Message  string   `json:"message"`
	Source   string   `json:"source"`
	Actions  []string `json:"actions,omitempty"`
}

// WorkTrustEvidence points to supporting material.
type WorkTrustEvidence struct {
	Kind    string `json:"kind"`
	Ref     string `json:"ref"`
	Summary string `json:"summary,omitempty"`
	At      string `json:"at,omitempty"`
}

// WorkTrustRecommendation is the suggested next CLI action.
type WorkTrustRecommendation struct {
	Action    string `json:"action"`
	Command   string `json:"command"`
	Rationale string `json:"rationale"`
	Priority  string `json:"priority"`
}

// FeatureTrustReport aggregates task-level work trust for one feature.
type FeatureTrustReport struct {
	ReportVersion string                    `json:"report_version"`
	Scope         TrustScope                `json:"scope"`
	GeneratedAt   string                    `json:"generated_at"`
	Score         WorkTrustScore            `json:"score"`
	TaskCount     int                       `json:"task_count"`
	Tasks         []FeatureTaskSummary      `json:"tasks"`
	NextActions   []WorkTrustRecommendation `json:"next_actions,omitempty"`
}

// FeatureTaskSummary is the per-task rollup inside a feature or run report.
type FeatureTaskSummary struct {
	TaskID    string  `json:"task_id"`
	Status    string  `json:"status"`
	Score     float64 `json:"score"`
	Verdict   Verdict `json:"verdict"`
	Command   string  `json:"command,omitempty"`
	Rationale string  `json:"rationale,omitempty"`
}

// RunTrustReport aggregates task-level work trust for one workflow run.
type RunTrustReport struct {
	ReportVersion string                    `json:"report_version"`
	Scope         TrustScope                `json:"scope"`
	GeneratedAt   string                    `json:"generated_at"`
	Score         WorkTrustScore            `json:"score"`
	TaskCount     int                       `json:"task_count"`
	PlanGate      *RunPlanGateSummary       `json:"plan_gate,omitempty"`
	Tasks         []FeatureTaskSummary      `json:"tasks"`
	NextActions   []WorkTrustRecommendation `json:"next_actions,omitempty"`
}

// RunPlanGateSummary is the run-scoped plan gate signal (from gate logs).
type RunPlanGateSummary struct {
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes,omitempty"`
}
