package workflow

import (
	"errors"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
)

func TestTransitionTaskValid(t *testing.T) {
	if err := TransitionTask(agentflow.StatusPending, agentflow.StatusPlanned, false); err != nil {
		t.Fatal(err)
	}
}

func TestTransitionTaskInvalid(t *testing.T) {
	err := TransitionTask(agentflow.StatusPending, agentflow.StatusVerified, false)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("got %v", err)
	}
}

func TestTransitionTaskAlreadyDone(t *testing.T) {
	err := TransitionTask(agentflow.StatusVerified, agentflow.StatusVerified, false)
	if !errors.Is(err, ErrStepAlreadyDone) {
		t.Fatalf("got %v", err)
	}
}

func TestNextWorkflowStep(t *testing.T) {
	if got := NextWorkflowStep([]string{agentflow.StatusPlanned}); got != "enrich" {
		t.Fatalf("got %q", got)
	}
	if got := NextWorkflowStep([]string{agentflow.StatusImplemented}); got != "verify" {
		t.Fatalf("got %q", got)
	}
}
