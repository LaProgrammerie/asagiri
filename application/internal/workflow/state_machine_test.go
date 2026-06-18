package workflow

import (
	"errors"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

func TestTransitionTaskValid(t *testing.T) {
	if err := TransitionTask(asagiri.StatusPending, asagiri.StatusPlanned, false); err != nil {
		t.Fatal(err)
	}
}

func TestTransitionTaskInvalid(t *testing.T) {
	err := TransitionTask(asagiri.StatusPending, asagiri.StatusVerified, false)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("got %v", err)
	}
}

func TestTransitionTaskAlreadyDone(t *testing.T) {
	err := TransitionTask(asagiri.StatusVerified, asagiri.StatusVerified, false)
	if !errors.Is(err, ErrStepAlreadyDone) {
		t.Fatalf("got %v", err)
	}
}

func TestTransitionTaskImplementedToRunning(t *testing.T) {
	if err := TransitionTask(asagiri.StatusImplemented, asagiri.StatusRunning, true); err != nil {
		t.Fatalf("implemented -> running: %v", err)
	}
}

func TestTransitionTaskFailedToRunningRequiresForce(t *testing.T) {
	err := TransitionTask(asagiri.StatusFailed, asagiri.StatusRunning, false)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("without force: got %v", err)
	}
	if err := TransitionTask(asagiri.StatusFailed, asagiri.StatusRunning, true); err != nil {
		t.Fatalf("failed -> running with force: %v", err)
	}
}

func TestTransitionTaskFailedToVerifiedBlockedWithForce(t *testing.T) {
	err := TransitionTask(asagiri.StatusFailed, asagiri.StatusVerified, true)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("got %v", err)
	}
}

func TestNextWorkflowStep(t *testing.T) {
	if got := NextWorkflowStep([]string{asagiri.StatusPlanned}); got != "enrich" {
		t.Fatalf("got %q", got)
	}
	if got := NextWorkflowStep([]string{asagiri.StatusImplemented}); got != "verify" {
		t.Fatalf("got %q", got)
	}
}
