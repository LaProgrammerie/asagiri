package executiongraph

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraphTransitionsValid(t *testing.T) {
	valid := [][2]GraphStatus{
		{GraphStatusPlanned, GraphStatusReady},
		{GraphStatusPlanned, GraphStatusAborted},
		{GraphStatusReady, GraphStatusRunning},
		{GraphStatusRunning, GraphStatusBlocked},
		{GraphStatusRunning, GraphStatusPaused},
		{GraphStatusRunning, GraphStatusFailed},
		{GraphStatusRunning, GraphStatusCompleted},
		{GraphStatusBlocked, GraphStatusReady},
		{GraphStatusPaused, GraphStatusRunning},
		{GraphStatusFailed, GraphStatusRolledBack},
		{GraphStatusFailed, GraphStatusReady},
	}
	for _, tc := range valid {
		require.True(t, CanTransitionGraph(tc[0], tc[1]), "%s -> %s", tc[0], tc[1])
		require.NoError(t, TransitionGraph(tc[0], tc[1], false))
	}
}

func TestGraphTransitionsInvalid(t *testing.T) {
	for _, from := range AllGraphStatuses() {
		for _, to := range AllGraphStatuses() {
			if from == to {
				continue
			}
			if CanTransitionGraph(from, to) {
				continue
			}
			err := TransitionGraph(from, to, false)
			require.Error(t, err, "%s -> %s should be invalid", from, to)
			require.True(t, errors.Is(err, ErrInvalidTransition) || IsTerminalGraphStatus(from))
		}
	}
}

func TestGraphTransitionSameState(t *testing.T) {
	err := TransitionGraph(GraphStatusPlanned, GraphStatusPlanned, false)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrGraphAlreadyInState)

	require.NoError(t, TransitionGraph(GraphStatusPlanned, GraphStatusPlanned, true))
}

func TestGraphTerminalStatuses(t *testing.T) {
	for _, status := range []GraphStatus{GraphStatusCompleted, GraphStatusAborted, GraphStatusRolledBack} {
		require.True(t, IsTerminalGraphStatus(status))
		require.False(t, CanTransitionGraph(status, GraphStatusReady))
	}
	require.False(t, IsTerminalGraphStatus(GraphStatusRunning))
}

func TestNodeTransitionsValid(t *testing.T) {
	valid := [][2]NodeStatus{
		{NodeStatusPending, NodeStatusReady},
		{NodeStatusPending, NodeStatusSkipped},
		{NodeStatusReady, NodeStatusRunning},
		{NodeStatusRunning, NodeStatusSucceeded},
		{NodeStatusRunning, NodeStatusFailed},
		{NodeStatusFailed, NodeStatusReady},
		{NodeStatusFailed, NodeStatusRolledBack},
		{NodeStatusBlocked, NodeStatusReady},
	}
	for _, tc := range valid {
		require.True(t, CanTransitionNode(tc[0], tc[1]), "%s -> %s", tc[0], tc[1])
		require.NoError(t, TransitionNode(tc[0], tc[1], false))
	}
}

func TestNodeTransitionsInvalid(t *testing.T) {
	for _, from := range AllNodeStatuses() {
		for _, to := range AllNodeStatuses() {
			if from == to {
				continue
			}
			if CanTransitionNode(from, to) {
				continue
			}
			err := TransitionNode(from, to, false)
			require.Error(t, err, "%s -> %s should be invalid", from, to)
			require.True(t, errors.Is(err, ErrInvalidNodeTransition) || IsTerminalNodeStatus(from))
		}
	}
}

func TestNodeTransitionSameState(t *testing.T) {
	err := TransitionNode(NodeStatusPending, NodeStatusPending, false)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNodeAlreadyInState)

	require.NoError(t, TransitionNode(NodeStatusPending, NodeStatusPending, true))
}

func TestNodeTerminalStatuses(t *testing.T) {
	for _, status := range []NodeStatus{NodeStatusSucceeded, NodeStatusSkipped, NodeStatusRolledBack} {
		require.True(t, IsTerminalNodeStatus(status))
		require.False(t, CanTransitionNode(status, NodeStatusReady))
	}
	require.False(t, IsTerminalNodeStatus(NodeStatusRunning))
}
