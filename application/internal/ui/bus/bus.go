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

// ExplainResult contains read-only explainability content.
type ExplainResult struct {
	Subject       string
	Reasons       []string
	Evidence      []string
	Source        string
	Alternatives  []string
	CLIEquivalent string
	Warning       string
}

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
	CostTodayEUR  float64
	CostMonthEUR  float64
	UpdatedAt     time.Time
	Warnings      []string
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
}

func (GetExplainQuery) Name() string { return "GetExplain" }

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

// BuildKnowledgeGraphCommand is a command stub for lot 1.
type BuildKnowledgeGraphCommand struct{}

func (BuildKnowledgeGraphCommand) Name() string          { return "BuildKnowledgeGraph" }
func (BuildKnowledgeGraphCommand) CLIEquivalent() string { return "asa knowledge build" }

// ReplayRunCommand is a command stub for lot 1.
type ReplayRunCommand struct {
	RunID string
}

func (ReplayRunCommand) Name() string          { return "ReplayRun" }
func (ReplayRunCommand) CLIEquivalent() string { return "asa replay run <replay-id>" }
