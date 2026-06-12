package executiongraph

import (
	"fmt"
)

// graphAllowedTransitions maps graph status to permitted next statuses (spec §20).
var graphAllowedTransitions = map[GraphStatus][]GraphStatus{
	GraphStatusPlanned: {GraphStatusReady, GraphStatusAborted},
	GraphStatusReady:   {GraphStatusRunning, GraphStatusAborted},
	GraphStatusRunning: {
		GraphStatusBlocked, GraphStatusPaused, GraphStatusFailed,
		GraphStatusCompleted, GraphStatusAborted,
	},
	GraphStatusBlocked: {GraphStatusReady, GraphStatusRunning, GraphStatusAborted},
	GraphStatusPaused:  {GraphStatusRunning, GraphStatusAborted},
	GraphStatusFailed:  {GraphStatusReady, GraphStatusRolledBack, GraphStatusAborted},
}

// nodeAllowedTransitions maps node status to permitted next statuses (spec §20).
var nodeAllowedTransitions = map[NodeStatus][]NodeStatus{
	NodeStatusPending: {NodeStatusReady, NodeStatusSkipped, NodeStatusBlocked},
	NodeStatusReady:   {NodeStatusRunning, NodeStatusSkipped, NodeStatusBlocked},
	NodeStatusRunning: {NodeStatusSucceeded, NodeStatusFailed, NodeStatusBlocked},
	NodeStatusFailed:  {NodeStatusReady, NodeStatusRolledBack, NodeStatusBlocked},
	NodeStatusBlocked: {NodeStatusReady, NodeStatusSkipped},
}

// TransitionGraph moves a graph to toStatus unless already at a terminal success state.
func TransitionGraph(from, to GraphStatus, force bool) error {
	if from == to {
		if force {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrGraphAlreadyInState, from)
	}
	if !CanTransitionGraph(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

// CanTransitionGraph reports whether from may move to to.
func CanTransitionGraph(from, to GraphStatus) bool {
	if IsTerminalGraphStatus(from) {
		return false
	}
	next, ok := graphAllowedTransitions[from]
	if !ok {
		return false
	}
	for _, n := range next {
		if n == to {
			return true
		}
	}
	return false
}

// IsTerminalGraphStatus reports whether a graph status cannot transition further.
func IsTerminalGraphStatus(status GraphStatus) bool {
	switch status {
	case GraphStatusCompleted, GraphStatusAborted, GraphStatusRolledBack:
		return true
	default:
		return false
	}
}

// TransitionNode moves a node to toStatus unless already at a terminal success state.
func TransitionNode(from, to NodeStatus, force bool) error {
	if from == to {
		if force {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrNodeAlreadyInState, from)
	}
	if !CanTransitionNode(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidNodeTransition, from, to)
	}
	return nil
}

// CanTransitionNode reports whether from may move to to.
func CanTransitionNode(from, to NodeStatus) bool {
	if IsTerminalNodeStatus(from) {
		return false
	}
	next, ok := nodeAllowedTransitions[from]
	if !ok {
		return false
	}
	for _, n := range next {
		if n == to {
			return true
		}
	}
	return false
}

// IsTerminalNodeStatus reports whether a node status cannot transition further.
func IsTerminalNodeStatus(status NodeStatus) bool {
	switch status {
	case NodeStatusSucceeded, NodeStatusSkipped, NodeStatusRolledBack:
		return true
	default:
		return false
	}
}

// AllGraphStatuses returns every graph status in stable order.
func AllGraphStatuses() []GraphStatus {
	return []GraphStatus{
		GraphStatusPlanned, GraphStatusReady, GraphStatusRunning, GraphStatusBlocked,
		GraphStatusPaused, GraphStatusFailed, GraphStatusCompleted, GraphStatusAborted,
		GraphStatusRolledBack,
	}
}

// AllNodeStatuses returns every node status in stable order.
func AllNodeStatuses() []NodeStatus {
	return []NodeStatus{
		NodeStatusPending, NodeStatusReady, NodeStatusRunning, NodeStatusSucceeded,
		NodeStatusFailed, NodeStatusSkipped, NodeStatusBlocked, NodeStatusRolledBack,
	}
}
