package bus

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestQueryBusReadOnlyHandlers(t *testing.T) {
	repoRoot := t.TempDir()
	statePath := filepath.Join(repoRoot, ".asagiri", "state.sqlite")

	rt, err := runtime.Open(repoRoot)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rt.Close() })
	_, err = rt.CreateSession("mission", "workspace", "onboarding")
	require.NoError(t, err)
	_, err = rt.EmitEvent("runtime.started", "tests", "", "onboarding", map[string]any{"ok": true})
	require.NoError(t, err)

	st, err := sqlite.Open(statePath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.Migrate())
	require.NoError(t, st.CreateRun(&sqlite.Run{
		ID:      "run-1",
		Feature: "spec-ui",
		Status:  sqlite.StatusRunning,
	}))

	qb := NewQueryBus(Deps{
		RepoRoot:    repoRoot,
		StateDBPath: statePath,
	})

	rsAny, err := qb.Query(context.Background(), GetRuntimeStatusQuery{})
	require.NoError(t, err)
	rs, ok := rsAny.(RuntimeStatusResult)
	require.True(t, ok)
	require.GreaterOrEqual(t, rs.Status.Sessions, 1)

	runsAny, err := qb.Query(context.Background(), ListRunsQuery{Limit: 10})
	require.NoError(t, err)
	runs, ok := runsAny.(ListRunsResult)
	require.True(t, ok)
	require.Len(t, runs.Runs, 1)
	require.Equal(t, "run-1", runs.Runs[0].ID)

	eventsAny, err := qb.Query(context.Background(), GetRecentEventsQuery{Limit: 10})
	require.NoError(t, err)
	events, ok := eventsAny.(RecentEventsResult)
	require.True(t, ok)
	require.NotEmpty(t, events.Events)
}

func TestCommandBusStubsExposeCLIEquivalent(t *testing.T) {
	called := false
	cb := NewCommandBus(Deps{
		StartWork: func(_ context.Context, _ Deps, cmd StartWorkCommand) (CommandResult, error) {
			called = true
			return CommandResult{
				Accepted:      true,
				Message:       "ok",
				CLIEquivalent: cmd.CLIEquivalent(),
			}, nil
		},
	})

	res, err := cb.Dispatch(context.Background(), StartWorkCommand{Intent: "add invitations"})
	require.NoError(t, err)
	require.True(t, called)
	require.True(t, res.Accepted)
	require.Equal(t, `asa work "add invitations"`, res.CLIEquivalent)
}

func TestCommandBusDispatchesInvestigationAndTrustHandlers(t *testing.T) {
	t.Parallel()

	investigateCalled := false
	verifyCalled := false
	cb := NewCommandBus(Deps{
		Investigate: func(_ context.Context, _ Deps, cmd RunInvestigationCommand) (CommandResult, error) {
			investigateCalled = true
			return CommandResult{Accepted: true, CLIEquivalent: cmd.CLIEquivalent()}, nil
		},
		VerifyTrust: func(_ context.Context, _ Deps, cmd VerifyTrustCommand) (CommandResult, error) {
			verifyCalled = true
			return CommandResult{Accepted: true, CLIEquivalent: cmd.CLIEquivalent()}, nil
		},
	})

	resInv, err := cb.Dispatch(context.Background(), RunInvestigationCommand{Symptom: "onboarding fails"})
	require.NoError(t, err)
	require.True(t, investigateCalled)
	require.Equal(t, `asa investigate "onboarding fails"`, resInv.CLIEquivalent)

	resTrust, err := cb.Dispatch(context.Background(), VerifyTrustCommand{Target: "onboarding"})
	require.NoError(t, err)
	require.True(t, verifyCalled)
	require.Equal(t, "asa verify trust onboarding", resTrust.CLIEquivalent)
}
