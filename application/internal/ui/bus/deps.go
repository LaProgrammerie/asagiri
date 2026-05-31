package bus

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

type runtimeStore interface {
	Close() error
	Status() (runtime.DaemonStatus, error)
	ListEvents(limit int) ([]runtime.RuntimeEvent, error)
	ListSessions() ([]runtime.Session, error)
	CollectMetrics() (runtime.MetricsSnapshot, error)
}

type stateStore interface {
	Close() error
	Migrate() error
	ListRuns(limit int) ([]sqlite.Run, error)
	GetRun(id string) (*sqlite.Run, error)
	ListTasksByRun(runID string) ([]sqlite.Task, error)
	GetRunMetric(runID string) (*telemetry.RunMetric, error)
}

type startWorkHandler func(ctx context.Context, deps Deps, cmd StartWorkCommand) (CommandResult, error)
type runInvestigationHandler func(ctx context.Context, deps Deps, cmd RunInvestigationCommand) (CommandResult, error)
type verifyTrustHandler func(ctx context.Context, deps Deps, cmd VerifyTrustCommand) (CommandResult, error)
type buildKnowledgeGraphHandler func(ctx context.Context, deps Deps, cmd BuildKnowledgeGraphCommand) (CommandResult, error)
type replayRunHandler func(ctx context.Context, deps Deps, cmd ReplayRunCommand) (CommandResult, error)
type graphRollbackHandler func(ctx context.Context, deps Deps, cmd GraphRollbackCommand) (CommandResult, error)
type exportEventsHandler func(ctx context.Context, deps Deps, cmd ExportEventsCommand) (CommandResult, error)
type exportGraphHandler func(ctx context.Context, deps Deps, cmd ExportGraphCommand) (CommandResult, error)
type graphResumeHandler func(ctx context.Context, deps Deps, cmd GraphResumeCommand) (CommandResult, error)
type analyzeKnowledgeImpactHandler func(ctx context.Context, deps Deps, cmd AnalyzeKnowledgeImpactCommand) (CommandResult, error)
type buildKnowledgeContextHandler func(ctx context.Context, deps Deps, cmd BuildKnowledgeContextCommand) (CommandResult, error)
type compareReplayHandler func(ctx context.Context, deps Deps, cmd CompareReplayCommand) (CommandResult, error)
type explainReplayDivergenceHandler func(ctx context.Context, deps Deps, cmd ExplainReplayDivergenceCommand) (CommandResult, error)
type prototypeCreateHandler func(ctx context.Context, deps Deps, cmd PrototypeCreateCommand) (CommandResult, error)
type flowsExtractHandler func(ctx context.Context, deps Deps, cmd FlowsExtractCommand) (CommandResult, error)
type contractsExtractHandler func(ctx context.Context, deps Deps, cmd ContractsExtractCommand) (CommandResult, error)
type specGenerateFromProductHandler func(ctx context.Context, deps Deps, cmd SpecGenerateFromProductCommand) (CommandResult, error)

// Deps centralizes infrastructure dependencies used by ui/bus.
type Deps struct {
	RepoRoot    string
	StateDBPath string
	Config      *config.Config
	DryRun      bool
	RuntimeOpen func(repoRoot string) (runtimeStore, error)
	StateOpen   func(path string) (stateStore, error)
	StartWork            startWorkHandler
	Investigate          runInvestigationHandler
	VerifyTrust          verifyTrustHandler
	BuildKnowledgeGraph  buildKnowledgeGraphHandler
	ReplayRun            replayRunHandler
	GraphRollback        graphRollbackHandler
	ExportEvents         exportEventsHandler
	ExportGraph          exportGraphHandler
	GraphResume          graphResumeHandler
	AnalyzeKnowledgeImpact analyzeKnowledgeImpactHandler
	BuildKnowledgeContext  buildKnowledgeContextHandler
	CompareReplay          compareReplayHandler
	ExplainReplayDivergence explainReplayDivergenceHandler
	PrototypeCreate         prototypeCreateHandler
	FlowsExtract            flowsExtractHandler
	ContractsExtract        contractsExtractHandler
	SpecGenerateFromProduct specGenerateFromProductHandler
}

func (d Deps) withDefaults() Deps {
	if d.RuntimeOpen == nil {
		d.RuntimeOpen = func(repoRoot string) (runtimeStore, error) {
			return runtime.Open(repoRoot)
		}
	}
	if d.StateOpen == nil {
		d.StateOpen = func(path string) (stateStore, error) {
			return sqlite.Open(path)
		}
	}
	if !d.DryRun {
		d.DryRun = env.DryRunEnabled()
	}
	if d.StartWork == nil {
		d.StartWork = dispatchStartWork
	}
	if d.Investigate == nil {
		d.Investigate = dispatchRunInvestigation
	}
	if d.VerifyTrust == nil {
		d.VerifyTrust = dispatchVerifyTrust
	}
	if d.BuildKnowledgeGraph == nil {
		d.BuildKnowledgeGraph = dispatchBuildKnowledgeGraph
	}
	if d.ReplayRun == nil {
		d.ReplayRun = dispatchReplayRun
	}
	if d.GraphRollback == nil {
		d.GraphRollback = dispatchGraphRollback
	}
	if d.ExportEvents == nil {
		d.ExportEvents = dispatchExportEvents
	}
	if d.ExportGraph == nil {
		d.ExportGraph = dispatchExportGraph
	}
	if d.GraphResume == nil {
		d.GraphResume = dispatchGraphResume
	}
	if d.AnalyzeKnowledgeImpact == nil {
		d.AnalyzeKnowledgeImpact = dispatchAnalyzeKnowledgeImpact
	}
	if d.BuildKnowledgeContext == nil {
		d.BuildKnowledgeContext = dispatchBuildKnowledgeContext
	}
	if d.CompareReplay == nil {
		d.CompareReplay = dispatchCompareReplay
	}
	if d.ExplainReplayDivergence == nil {
		d.ExplainReplayDivergence = dispatchExplainReplayDivergence
	}
	if d.PrototypeCreate == nil {
		d.PrototypeCreate = dispatchPrototypeCreate
	}
	if d.FlowsExtract == nil {
		d.FlowsExtract = dispatchFlowsExtract
	}
	if d.ContractsExtract == nil {
		d.ContractsExtract = dispatchContractsExtract
	}
	if d.SpecGenerateFromProduct == nil {
		d.SpecGenerateFromProduct = dispatchSpecGenerateFromProduct
	}
	return d
}
