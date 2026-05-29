package bus

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
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
}

type startWorkHandler func(ctx context.Context, deps Deps, cmd StartWorkCommand) (CommandResult, error)
type runInvestigationHandler func(ctx context.Context, deps Deps, cmd RunInvestigationCommand) (CommandResult, error)
type verifyTrustHandler func(ctx context.Context, deps Deps, cmd VerifyTrustCommand) (CommandResult, error)

// Deps centralizes infrastructure dependencies used by ui/bus.
type Deps struct {
	RepoRoot    string
	StateDBPath string
	Config      *config.Config
	DryRun      bool
	RuntimeOpen func(repoRoot string) (runtimeStore, error)
	StateOpen   func(path string) (stateStore, error)
	StartWork   startWorkHandler
	Investigate runInvestigationHandler
	VerifyTrust verifyTrustHandler
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
	return d
}
