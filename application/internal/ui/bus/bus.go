package bus

import (
	"context"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// CommandBus dispatches UI commands to application handlers.
type CommandBus interface {
	Dispatch(ctx context.Context, cmd Command) (CommandResult, error)
}

// QueryBus dispatches read-only UI queries to application handlers.
type QueryBus interface {
	Query(ctx context.Context, query Query) (QueryResult, error)
}

// Command is one UI action request.
type Command interface {
	Name() string
	CLIEquivalent() string
}

// Query is one read-only UI request.
type Query interface {
	Name() string
}

// QueryResult is a marker interface for query responses.
type QueryResult interface {
	isQueryResult()
}

// CommandResult is a generic command response.
type CommandResult struct {
	Accepted      bool
	Message       string
	CLIEquivalent string
}

// RunSummary is a view model for recent workflow runs.
type RunSummary struct {
	ID        string
	Feature   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RunPipelineStep is one pipeline step (spec→plan→dev→verify→trust→report)
// projected for the Runs detail pane.
type RunPipelineStep struct {
	ID     string
	Label  string
	Status string
}

// RunDetail aggregates a single run's pipeline, worktree, agents, validation,
// trust gate, cost and recent events for the Runs screen. Built by ui/bus
// adapters from workflow/runtime/trust; no business logic lives in screens.
type RunDetail struct {
	ID         string
	Feature    string
	Status     string
	Worktree   string
	Pipeline   []RunPipelineStep
	Agents     []ActiveAgentSummary
	Validation string
	TrustGate  TrustSummaryResult
	CostEUR    float64
	Events     []EventSummary
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Warning    string
}

func (RunDetail) isQueryResult() {}

// GetRunDetailQuery loads the aggregated detail for one run.
type GetRunDetailQuery struct {
	RunID string
}

func (GetRunDetailQuery) Name() string { return "GetRunDetail" }

// EventSummary is a view model for runtime events.
type EventSummary struct {
	ID        string
	Type      string
	Source    string
	SessionID string
	FlowID    string
	CreatedAt time.Time
	Payload   map[string]any
}

// RuntimeStatusResult wraps daemon status and optional warning.
type RuntimeStatusResult struct {
	Status  runtime.DaemonStatus
	Warning string
}

// ListRunsResult contains recent runs.
type ListRunsResult struct {
	Runs    []RunSummary
	Warning string
}

// RecentEventsResult contains recent runtime events.
type RecentEventsResult struct {
	Events []EventSummary
}

// TrustDimensionScore is one trust confidence row.
type TrustDimensionScore struct {
	Label string
	Score float64
}

// TrustSummaryResult contains the latest trust snapshot.
type TrustSummaryResult struct {
	Overall     float64
	Dimensions  []TrustDimensionScore
	GeneratedAt time.Time
	Warning     string
}

// ActiveAgentSummary contains one active/known agent status line.
type ActiveAgentSummary struct {
	Role      string
	AgentRef  string
	Status    string
	FlowID    string
	UpdatedAt time.Time
}

// ActiveAgentsResult contains active agents from runtime events.
type ActiveAgentsResult struct {
	Agents  []ActiveAgentSummary
	Warning string
}

// FlowGraphStep is one flow step for Mission/Dashboard widgets.
type FlowGraphStep struct {
	ID     string
	Label  string
	Status string
}

// FlowGraphResult contains a compact flow projection.
type FlowGraphResult struct {
	FlowID  string
	Steps   []FlowGraphStep
	Warning string
}

// FlowStepDetail represents one enriched step for Flow Explorer.
type FlowStepDetail struct {
	ID         string
	Label      string
	Status     string
	API        string
	Service    string
	Event      string
	Tests      []string
	Metrics    []string
	TrustScore float64
	Risk       string
}

// FlowExplorerResult contains flow explorer data.
type FlowExplorerResult struct {
	FlowID   string
	Steps    []FlowStepDetail
	Selected string
	Warning  string
}

// GraphNodeSummary contains one graph node projection for Graph Explorer.
type GraphNodeSummary struct {
	ID            string
	Title         string
	Type          string
	Status        string
	Risk          string
	BlockedBy     []string
	CLIEquivalent string
}

// GraphExplorerResult contains graph explorer data.
type GraphExplorerResult struct {
	GraphID string
	Product string
	FlowID  string
	Status  string
	Nodes   []GraphNodeSummary
	Warning string
}

// KnowledgeMatch is one knowledge search result line.
type KnowledgeMatch struct {
	ID            string
	Type          string
	Name          string
	Path          string
	Score         float64
	CLIEquivalent string
}

// KnowledgeSearchResult contains lightweight knowledge search hits.
type KnowledgeSearchResult struct {
	Query   string
	Matches []KnowledgeMatch
	Warning string
}

// TrustEvidenceDimension contains score and evidence for one trust dimension.
type TrustEvidenceDimension struct {
	Label         string
	Score         float64
	Findings      []string
	Evidence      []string
	CLIEquivalent string
}

// TrustExplorerResult contains trust explorer details.
type TrustExplorerResult struct {
	Overall      float64
	ResidualRisk string
	GateStatus   string
	GateReason   string
	Dimensions   []TrustEvidenceDimension
	Warnings     []string
	Warning      string
}

// FocusContextKind identifies the UI focus subject for explainability.
type FocusContextKind string

const (
	FocusKindGraphNode      FocusContextKind = "graph-node"
	FocusKindFlowStep       FocusContextKind = "flow-step"
	FocusKindTrustDimension FocusContextKind = "trust-dimension"
	FocusKindAgent          FocusContextKind = "agent"
	FocusKindReplayEvent    FocusContextKind = "replay-event"
	FocusKindDecision       FocusContextKind = "decision"
)

// FocusContext captures the current drill-down target passed to QueryBus.
type FocusContext struct {
	Kind    FocusContextKind
	Subject string
	Detail  string
	Screen  string
}

// ExplainContext enriches explain queries with focus and typed questions.
type ExplainContext struct {
	Focus    FocusContext
	Question string
}

// ExplainResult contains read-only explainability content.
type ExplainResult struct {
	Subject              string
	Question             string
	SupportedQuestions   []string
	Reasons              []string
	Evidence             []string
	Source               string
	Alternatives         []string
	CLIEquivalent        string
	Warning              string
}

// RecommendedAction is one contextual next step for Mission Control.
type RecommendedAction struct {
	ID            string
	Title         string
	Description   string
	Priority      int
	CLIEquivalent string
	ActionID      string
}

// RecommendedActionsResult lists ranked actions for Mission Control.
type RecommendedActionsResult struct {
	Actions []RecommendedAction
}

func (RecommendedActionsResult) isQueryResult() {}

// AgentCard contains one live agent theatre card.
type AgentCard struct {
	Role            string
	AgentRef        string
	Status          string
	Task            string
	FilesActive     int
	Hypothesis      string
	TokensEstimated int
	CostEUR         float64
	Duration        time.Duration
	LastOutput      string
	Confidence      float64
	UpdatedAt       time.Time
}

// AgentTheatreResult contains live cards for agents screen.
type AgentTheatreResult struct {
	Agents  []AgentCard
	Warning string
}

// ReplayTimelineEvent contains one event row for replay timeline.
type ReplayTimelineEvent struct {
	Time      time.Time
	Type      string
	Source    string
	SessionID string
	FlowID    string
	Artifact  string
}

// ReplayPackageResult contains replay package details for replay screen.
type ReplayPackageResult struct {
	ReplayID   string
	CreatedAt  time.Time
	RepoBranch string
	RepoCommit string
	Mode       string
	Artifacts  []string
	Timeline   []ReplayTimelineEvent
	Warnings   []string
	Warning    string
}

// PrototypeFlowStep describes one extracted flow step.
type PrototypeFlowStep struct {
	FlowID    string
	StepID    string
	Action    string
	Screen    string
	Next      string
	Contract  string
	Trust     string
	Metric    string
	Sensitive bool
}

// PrototypePipelineResult contains prototype pipeline and split-view content.
type PrototypePipelineResult struct {
	Product          string
	WireframeTitle   string
	WireframePath    string
	PipelineStage    string
	StagesDone       []string
	Flow             string
	FlowExtraction   []PrototypeFlowStep
	SuggestedActions []string
	Warnings         []string
	Warning          string
}

// MissionControlSnapshotResult aggregates lot-2 Mission/Dashboard data.
type MissionControlSnapshotResult struct {
	Workspace     string
	Branch        string
	SessionStatus string
	Runtime       RuntimeStatusResult
	Trust         TrustSummaryResult
	Runs          []RunSummary
	Events        []EventSummary
	ActiveAgents  []ActiveAgentSummary
	Flow          FlowGraphResult
	FlowExplorer  FlowExplorerResult
	GraphExplorer GraphExplorerResult
	Knowledge     KnowledgeSearchResult
	TrustExplorer TrustExplorerResult
	Explain       ExplainResult
	AgentTheatre  AgentTheatreResult
	Replay        ReplayPackageResult
	Prototype     PrototypePipelineResult
	CostTodayEUR       float64
	CostMonthEUR       float64
	RecommendedActions []RecommendedAction
	Readiness          ReadinessResult
	UpdatedAt          time.Time
	Warnings           []string
}

func (RuntimeStatusResult) isQueryResult()          {}
func (ListRunsResult) isQueryResult()               {}
func (RecentEventsResult) isQueryResult()           {}
func (TrustSummaryResult) isQueryResult()           {}
func (ActiveAgentsResult) isQueryResult()           {}
func (FlowGraphResult) isQueryResult()              {}
func (FlowExplorerResult) isQueryResult()           {}
func (GraphExplorerResult) isQueryResult()          {}
func (KnowledgeSearchResult) isQueryResult()        {}
func (TrustExplorerResult) isQueryResult()          {}
func (ExplainResult) isQueryResult()                {}
func (AgentTheatreResult) isQueryResult()           {}
func (ReplayPackageResult) isQueryResult()          {}
func (PrototypePipelineResult) isQueryResult()      {}
func (MissionControlSnapshotResult) isQueryResult() {}
func (ReadinessResult) isQueryResult()              {}
func (OnboardingStateResult) isQueryResult()        {}
func (OnboardingWizardResult) isQueryResult()       {}

// ReadinessCheck is one readiness row for the UI.
type ReadinessCheck struct {
	ID      string
	Status  string
	Message string
	FixCLI  string
}

// ReadinessAction is a suggested CLI next step.
type ReadinessAction struct {
	Title string
	CLI   string
}

// AutofixOffer is a safe automatic correction offered after onboarding.
type AutofixOffer struct {
	ID          string
	Title       string
	Description string
	Lines       []string
}

// ReadinessResult mirrors onboarding readiness for QueryBus consumers.
type ReadinessResult struct {
	Ready         bool
	Score         int
	Checks        []ReadinessCheck
	NextActions   []ReadinessAction
	AutofixOffers []AutofixOffer
}

// OnboardingStateResult exposes wizard progress to the TUI.
type OnboardingStateResult struct {
	CurrentStep string
	Answers     map[string]string
	Completed   []string
}

// OnboardingWizardResult is the prefilled interactive wizard snapshot.
type OnboardingWizardResult struct {
	CurrentStep       string
	Steps             []string
	Fields            map[string]string
	Advanced          map[string]string
	ValidationPreview []string
	DetectedStacks    []string
	Errors            map[string]string
	SkippedFields     []string
}

// GetReadinessQuery loads project readiness report.
type GetReadinessQuery struct {
	Strict bool
}

func (GetReadinessQuery) Name() string { return "GetReadiness" }

// GetOnboardingStateQuery loads saved wizard state.
type GetOnboardingStateQuery struct{}

func (GetOnboardingStateQuery) Name() string { return "GetOnboardingState" }

// GetOnboardingWizardQuery loads prefilled wizard form for the TUI.
type GetOnboardingWizardQuery struct{}

func (GetOnboardingWizardQuery) Name() string { return "GetOnboardingWizard" }

// ValidateOnboardingStepQuery validates one wizard step from field maps.
type ValidateOnboardingStepQuery struct {
	Step     string
	Fields   map[string]string
	Advanced map[string]string
}

func (ValidateOnboardingStepQuery) Name() string { return "ValidateOnboardingStep" }

// ValidateOnboardingStepResult returns field errors for a step.
type ValidateOnboardingStepResult struct {
	Valid  bool
	Errors map[string]string
}

func (ValidateOnboardingStepResult) isQueryResult() {}

// RunOnboardingStepCommand advances or runs one wizard step.
type RunOnboardingStepCommand struct {
	Step string
	Yes  bool
}

func (RunOnboardingStepCommand) Name() string { return "RunOnboardingStep" }
func (c RunOnboardingStepCommand) CLIEquivalent() string {
	step := emptyCLIArg(c.Step, "<step>")
	return "asa onboard --step " + step + " --yes"
}

// AdvanceOnboardingStepCommand moves the wizard prev/next with draft fields.
type AdvanceOnboardingStepCommand struct {
	Direction string
	Fields    map[string]string
	Advanced  map[string]string
}

func (AdvanceOnboardingStepCommand) Name() string { return "AdvanceOnboardingStep" }
func (c AdvanceOnboardingStepCommand) CLIEquivalent() string {
	dir := emptyCLIArg(c.Direction, "next")
	return "asa onboard --step " + dir
}

// SetOnboardingFieldCommand updates one wizard field in persisted state.
type SetOnboardingFieldCommand struct {
	Field string
	Value string
}

func (SetOnboardingFieldCommand) Name() string { return "SetOnboardingField" }
func (c SetOnboardingFieldCommand) CLIEquivalent() string {
	return "asa onboard --resume"
}

// ApplyOnboardingConfigCommand runs onboard from wizard field maps.
type ApplyOnboardingConfigCommand struct {
	Yes      bool
	Stack    string
	Fields   map[string]string
	Advanced map[string]string
}

func (ApplyOnboardingConfigCommand) Name() string { return "ApplyOnboardingConfig" }
func (ApplyOnboardingConfigCommand) CLIEquivalent() string {
	return "asa onboard --yes --non-interactive"
}

// SkipOnboardingCheckCommand acknowledges a warn check (UI-only).
type SkipOnboardingCheckCommand struct {
	CheckID string
}

func (SkipOnboardingCheckCommand) Name() string { return "SkipOnboardingCheck" }
func (c SkipOnboardingCheckCommand) CLIEquivalent() string {
	return "asa ready --plain"
}

// ApplyReadinessAutofixCommand applies safe automatic readiness fixes (e.g. .gitignore).
type ApplyReadinessAutofixCommand struct{}

func (ApplyReadinessAutofixCommand) Name() string { return "ApplyReadinessAutofix" }
func (ApplyReadinessAutofixCommand) CLIEquivalent() string {
	return "asa ready --autofix"
}

// GetRuntimeStatusQuery loads current daemon counters.
type GetRuntimeStatusQuery struct{}

func (GetRuntimeStatusQuery) Name() string { return "GetRuntimeStatus" }

// ListRunsQuery lists recent persisted runs.
type ListRunsQuery struct {
	Limit int
}

func (ListRunsQuery) Name() string { return "ListRuns" }

// GetRecentEventsQuery lists recent runtime events.
type GetRecentEventsQuery struct {
	Limit int
}

func (GetRecentEventsQuery) Name() string { return "GetRecentEvents" }

// GetTrustSummaryQuery returns latest trust report summary.
type GetTrustSummaryQuery struct{}

func (GetTrustSummaryQuery) Name() string { return "GetTrustSummary" }

// ListActiveAgentsQuery lists active agents inferred from runtime events.
type ListActiveAgentsQuery struct {
	Limit int
}

func (ListActiveAgentsQuery) Name() string { return "ListActiveAgents" }

// GetFlowGraphQuery returns compact flow graph projection.
type GetFlowGraphQuery struct {
	FlowID string
	Limit  int
}

func (GetFlowGraphQuery) Name() string { return "GetFlowGraph" }

// GetFlowExplorerQuery returns enriched flow step details.
type GetFlowExplorerQuery struct {
	FlowID string
}

func (GetFlowExplorerQuery) Name() string { return "GetFlowExplorer" }

// GetGraphExplorerQuery returns graph explorer nodes and dependencies.
type GetGraphExplorerQuery struct {
	FlowID string
}

func (GetGraphExplorerQuery) Name() string { return "GetGraphExplorer" }

// SearchKnowledgeQuery searches knowledge nodes for a free-text query.
type SearchKnowledgeQuery struct {
	Query string
	Limit int
}

func (SearchKnowledgeQuery) Name() string { return "SearchKnowledge" }

// GetTrustExplorerQuery returns trust scores with findings and evidence.
type GetTrustExplorerQuery struct{}

func (GetTrustExplorerQuery) Name() string { return "GetTrustExplorer" }

// GetExplainQuery returns explainability content for a subject.
type GetExplainQuery struct {
	Subject string
	Context ExplainContext
}

func (GetExplainQuery) Name() string { return "GetExplain" }

// GetRecommendedActionsQuery returns contextual next steps for Mission Control.
type GetRecommendedActionsQuery struct {
	FlowID string
}

func (GetRecommendedActionsQuery) Name() string { return "GetRecommendedActions" }

// GetAgentTheatreQuery returns live, enriched agent cards.
type GetAgentTheatreQuery struct {
	Limit int
}

func (GetAgentTheatreQuery) Name() string { return "GetAgentTheatre" }

// GetReplayPackageQuery returns one replay package timeline and metadata.
type GetReplayPackageQuery struct {
	ReplayID string
	Limit    int
}

func (GetReplayPackageQuery) Name() string { return "GetReplayPackage" }

// GetPrototypePipelineQuery returns current prototype split-view data.
type GetPrototypePipelineQuery struct {
	Product string
	Limit   int
}

func (GetPrototypePipelineQuery) Name() string { return "GetPrototypePipeline" }

// GetMissionControlSnapshotQuery returns all data required by lot-2 screens.
type GetMissionControlSnapshotQuery struct {
	RunsLimit          int
	EventsLimit        int
	AgentsLimit        int
	Knowledge          string
	ExplainFor         string
	ExplainContext     ExplainContext
	FlowID             string
	ReplayID           string
	PrototypeProduct   string
	PrototypeFlowLimit int
}

func (GetMissionControlSnapshotQuery) Name() string { return "GetMissionControlSnapshot" }

// StartWorkCommand runs asa work from the UI.
type StartWorkCommand struct {
	Intent string
}

func (StartWorkCommand) Name() string { return "StartWork" }
func (c StartWorkCommand) CLIEquivalent() string {
	intent := "<intent>"
	if trimmed := strings.TrimSpace(c.Intent); trimmed != "" {
		intent = trimmed
	}
	return `asa work "` + intent + `"`
}

// RunInvestigationCommand runs asa investigate from the UI.
type RunInvestigationCommand struct {
	Symptom string
}

func (RunInvestigationCommand) Name() string { return "RunInvestigation" }
func (c RunInvestigationCommand) CLIEquivalent() string {
	symptom := "<symptom>"
	if trimmed := strings.TrimSpace(c.Symptom); trimmed != "" {
		symptom = trimmed
	}
	return `asa investigate "` + symptom + `"`
}

// VerifyTrustCommand runs asa verify trust from the UI.
type VerifyTrustCommand struct {
	Target string
}

func (VerifyTrustCommand) Name() string { return "VerifyTrust" }
func (c VerifyTrustCommand) CLIEquivalent() string {
	target := "<target>"
	if trimmed := strings.TrimSpace(c.Target); trimmed != "" {
		target = trimmed
	}
	return "asa verify trust " + target
}

// BuildKnowledgeGraphCommand rebuilds the knowledge graph.
type BuildKnowledgeGraphCommand struct {
	Incremental bool
}

func (BuildKnowledgeGraphCommand) Name() string { return "BuildKnowledgeGraph" }
func (BuildKnowledgeGraphCommand) CLIEquivalent() string {
	return "asa knowledge build"
}

// ReplayRunCommand replays a captured workflow package.
type ReplayRunCommand struct {
	RunID      string
	Offline    bool
	Simulation bool
}

func (ReplayRunCommand) Name() string { return "ReplayRun" }
func (c ReplayRunCommand) CLIEquivalent() string {
	id := "<replay-id>"
	if trimmed := strings.TrimSpace(c.RunID); trimmed != "" {
		id = trimmed
	}
	return "asa replay run " + id
}

// GraphRollbackCommand rolls back an execution graph by id.
type GraphRollbackCommand struct {
	GraphID string
}

func (GraphRollbackCommand) Name() string { return "GraphRollback" }
func (c GraphRollbackCommand) CLIEquivalent() string {
	id := "<graph-id>"
	if trimmed := strings.TrimSpace(c.GraphID); trimmed != "" {
		id = trimmed
	}
	return "asa graph rollback " + id
}

// ExportEventsCommand exports filtered runtime events to disk.
type ExportEventsCommand struct {
	TypeFilter string
	Search     string
	OutputPath string
}

func (ExportEventsCommand) Name() string { return "ExportEvents" }
func (c ExportEventsCommand) CLIEquivalent() string {
	filter := strings.TrimSpace(c.TypeFilter)
	if filter == "" {
		filter = "<filter>"
	}
	search := strings.TrimSpace(c.Search)
	if search == "" {
		search = "<query>"
	}
	return "asa runtime events --type " + filter + " --search " + search + " --export"
}

// GraphRollbackImpactResult is safety UX data for destructive graph rollback.
type GraphRollbackImpactResult struct {
	GraphID        string
	Title          string
	ImpactLines    []string
	RollbackPolicy string
	CLIEquivalent  string
	CanRollback    bool
}

func (GraphRollbackImpactResult) isQueryResult() {}

// GetGraphRollbackImpactQuery loads rollback impact for a graph id.
type GetGraphRollbackImpactQuery struct {
	GraphID string
}

func (GetGraphRollbackImpactQuery) Name() string { return "GetGraphRollbackImpact" }

// PaletteEntry is one command palette row.
type PaletteEntry struct {
	ID          string
	Title       string
	Description string
	CLI         string
	Keywords    []string
	ActionID    string
}

// PaletteEntriesResult lists palette rows for the current screen and query.
type PaletteEntriesResult struct {
	Entries []PaletteEntry
}

func (PaletteEntriesResult) isQueryResult() {}

// GetPaletteEntriesQuery returns static and dynamic palette entries.
type GetPaletteEntriesQuery struct {
	Screen string
	Query  string
	Limit  int
}

func (GetPaletteEntriesQuery) Name() string { return "GetPaletteEntries" }

// GraphViewMode selects one graph explorer projection.
type GraphViewMode string

const (
	GraphViewTimeline       GraphViewMode = "timeline"
	GraphViewDependency     GraphViewMode = "dependency"
	GraphViewCriticalPath   GraphViewMode = "critical-path"
	GraphViewParallelGroups GraphViewMode = "parallel-groups"
	GraphViewBlocked        GraphViewMode = "blocked"
)

// GraphViewModes lists switchable graph explorer views.
var GraphViewModes = []GraphViewMode{
	GraphViewTimeline,
	GraphViewDependency,
	GraphViewCriticalPath,
	GraphViewParallelGroups,
	GraphViewBlocked,
}

// GraphNodeDetail contains drill-down data for one graph node.
type GraphNodeDetail struct {
	GraphID       string
	NodeID        string
	Title         string
	Status        string
	Risk          string
	Type          string
	Dependencies  []string
	Dependents    []string
	BlockedBy     []string
	LogsHint      string
	CLIEquivalent string
}

// GraphViewResult contains nodes for one graph view mode.
type GraphViewResult struct {
	GraphID string
	FlowID  string
	View    GraphViewMode
	Nodes   []GraphNodeSummary
	Groups  []string
	Warning string
}

func (GraphViewResult) isQueryResult()      {}
func (GraphNodeDetail) isQueryResult()      {}
func (FlowStepDetailResult) isQueryResult() {}
func (KnowledgeMatchDetail) isQueryResult() {}
func (TrustDimensionDetail) isQueryResult() {}
func (ReplayEventDetail) isQueryResult()    {}
func (ReplayCompareResult) isQueryResult()  {}

// GetGraphViewQuery returns graph nodes filtered by view mode.
type GetGraphViewQuery struct {
	FlowID  string
	GraphID string
	View    GraphViewMode
}

func (GetGraphViewQuery) Name() string { return "GetGraphView" }

// GetGraphNodeDetailQuery returns dependency and status details for one node.
type GetGraphNodeDetailQuery struct {
	GraphID string
	NodeID  string
}

func (GetGraphNodeDetailQuery) Name() string { return "GetGraphNodeDetail" }

// FlowStepDetailResult wraps one selected flow step.
type FlowStepDetailResult struct {
	FlowID string
	Step   FlowStepDetail
}

// GetFlowStepDetailQuery returns one enriched flow step.
type GetFlowStepDetailQuery struct {
	FlowID string
	StepID string
}

func (GetFlowStepDetailQuery) Name() string { return "GetFlowStepDetail" }

// KnowledgeMatchDetail contains drill-down data for one knowledge match.
type KnowledgeMatchDetail struct {
	MatchID       string
	Name          string
	Type          string
	Path          string
	RelatedFlows  []string
	RelatedAPIs   []string
	RelatedTests  []string
	RelatedEvents []string
	CLIEquivalent string
}

// GetKnowledgeMatchDetailQuery returns related entities for one knowledge node.
type GetKnowledgeMatchDetailQuery struct {
	MatchID string
}

func (GetKnowledgeMatchDetailQuery) Name() string { return "GetKnowledgeMatchDetail" }

// TrustDimensionDetail contains drill-down trust evidence.
type TrustDimensionDetail struct {
	Label         string
	Score         float64
	Findings      []string
	Evidence      []string
	Checks        []string
	GateStatus    string
	GateReason    string
	ResidualRisk  string
	CLIEquivalent string
}

// GetTrustDimensionDetailQuery returns trust drill-down for one dimension label.
type GetTrustDimensionDetailQuery struct {
	Label string
}

func (GetTrustDimensionDetailQuery) Name() string { return "GetTrustDimensionDetail" }

// ReplayEventDetail contains one replay timeline event with artifact hints.
type ReplayEventDetail struct {
	ReplayID      string
	Index         int
	Type          string
	Time          time.Time
	Artifact      string
	ArtifactPath  string
	CLIEquivalent string
}

// GetReplayEventDetailQuery returns one replay timeline event.
type GetReplayEventDetailQuery struct {
	ReplayID string
	Index    int
}

func (GetReplayEventDetailQuery) Name() string { return "GetReplayEventDetail" }

// ReplayCompareResult contains replay comparison summary lines.
type ReplayCompareResult struct {
	ReplayA       string
	ReplayB       string
	Summary       []string
	Divergences   []string
	CLIEquivalent string
	Warning       string
}

// GetReplayCompareQuery compares two replay packages read-only.
type GetReplayCompareQuery struct {
	ReplayA string
	ReplayB string
}

func (GetReplayCompareQuery) Name() string { return "GetReplayCompare" }

// ExportGraphCommand exports a graph to disk (mermaid/json).
type ExportGraphCommand struct {
	GraphID string
	Format  string
}

func (ExportGraphCommand) Name() string { return "ExportGraph" }
func (c ExportGraphCommand) CLIEquivalent() string {
	id := "<graph-id>"
	if trimmed := strings.TrimSpace(c.GraphID); trimmed != "" {
		id = trimmed
	}
	format := strings.TrimSpace(c.Format)
	if format == "" {
		format = "mermaid"
	}
	return "asa graph visualize " + id + " --format " + format
}

// GraphResumeCommand resumes a paused execution graph.
type GraphResumeCommand struct {
	GraphID string
}

func (GraphResumeCommand) Name() string { return "GraphResume" }
func (c GraphResumeCommand) CLIEquivalent() string {
	id := "<graph-id>"
	if trimmed := strings.TrimSpace(c.GraphID); trimmed != "" {
		id = trimmed
	}
	return "asa graph resume " + id
}

// AnalyzeKnowledgeImpactCommand runs knowledge impact analysis.
type AnalyzeKnowledgeImpactCommand struct {
	Flow   string
	Action string
	File   string
}

func (AnalyzeKnowledgeImpactCommand) Name() string { return "AnalyzeKnowledgeImpact" }
func (c AnalyzeKnowledgeImpactCommand) CLIEquivalent() string {
	if strings.TrimSpace(c.File) != "" {
		return "asa impact analyze --file " + strings.TrimSpace(c.File)
	}
	flow := emptyCLIArg(c.Flow, "<flow>")
	action := emptyCLIArg(c.Action, "<action>")
	return "asa impact analyze --flow " + flow + " --action " + action
}

// BuildKnowledgeContextCommand builds a context pack around a knowledge node.
type BuildKnowledgeContextCommand struct {
	NodeID string
}

func (BuildKnowledgeContextCommand) Name() string { return "BuildKnowledgeContext" }
func (c BuildKnowledgeContextCommand) CLIEquivalent() string {
	node := emptyCLIArg(c.NodeID, "<node-id>")
	return `asa knowledge query --start ` + node + " --max-depth 2"
}

// CompareReplayCommand compares two replay packages.
type CompareReplayCommand struct {
	ReplayA string
	ReplayB string
}

func (CompareReplayCommand) Name() string { return "CompareReplay" }
func (c CompareReplayCommand) CLIEquivalent() string {
	a := emptyCLIArg(c.ReplayA, "<replay-a>")
	b := emptyCLIArg(c.ReplayB, "<replay-b>")
	return "asa replay compare " + a + " " + b
}

// ExplainReplayDivergenceCommand explains divergences between two replays.
type ExplainReplayDivergenceCommand struct {
	ReplayA string
	ReplayB string
}

func (ExplainReplayDivergenceCommand) Name() string { return "ExplainReplayDivergence" }
func (c ExplainReplayDivergenceCommand) CLIEquivalent() string {
	a := emptyCLIArg(c.ReplayA, "<replay-a>")
	b := emptyCLIArg(c.ReplayB, "<replay-b>")
	return "asa replay explain " + a + " " + b
}

func emptyCLIArg(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

// PrototypeCreateCommand runs asa prototype create from the UI.
type PrototypeCreateCommand struct {
	Intent  string
	Product string
	Stack   string
	Style   string
}

func (PrototypeCreateCommand) Name() string { return "PrototypeCreate" }
func (c PrototypeCreateCommand) CLIEquivalent() string {
	intent := emptyCLIArg(c.Intent, "<intent>")
	cmd := `asa prototype create "` + intent + `"`
	if product := strings.TrimSpace(c.Product); product != "" {
		cmd += " --product " + product
	}
	return cmd
}

// FlowsExtractCommand runs asa flows extract from the UI.
type FlowsExtractCommand struct {
	Product string
}

func (FlowsExtractCommand) Name() string { return "FlowsExtract" }
func (c FlowsExtractCommand) CLIEquivalent() string {
	return "asa flows extract " + emptyCLIArg(c.Product, "<product>")
}

// ContractsExtractCommand runs asa contracts extract from the UI.
type ContractsExtractCommand struct {
	Product string
}

func (ContractsExtractCommand) Name() string { return "ContractsExtract" }
func (c ContractsExtractCommand) CLIEquivalent() string {
	return "asa contracts extract " + emptyCLIArg(c.Product, "<product>")
}

// SpecGenerateFromProductCommand runs asa spec generate-from-product from the UI.
type SpecGenerateFromProductCommand struct {
	Product string
}

func (SpecGenerateFromProductCommand) Name() string { return "SpecGenerateFromProduct" }
func (c SpecGenerateFromProductCommand) CLIEquivalent() string {
	return "asa spec generate-from-product " + emptyCLIArg(c.Product, "<product>")
}

// FormatContractRef renders a product contract reference for TUI display.
func FormatContractRef(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "none"
	}
	upper := strings.ToUpper(ref)
	if strings.HasPrefix(upper, "TODO:") {
		name := strings.TrimSpace(ref[len("TODO:"):])
		if name == "" {
			return "pending contract"
		}
		return "pending: " + name
	}
	return ref
}
