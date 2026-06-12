package app

// Feature: audit-coherence-consolidation
//
// Example test for R7.6: when the repository is not onboarded, both the Mission
// Control and Runs screens must explicitly invite the user to run `asa onboard`.
//
// The invite content is purely presentational and lives in the UI screen
// packages (mission cockpit/banner and runs empty state). This test asserts the
// actual rendered screen output from the `app` package, which already composes
// both screens, rather than the domain `internal/onboarding` package (which
// holds no invite text and is imported *by* the UI, not the reverse).

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/mission"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/runs"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

const onboardInvite = "asa onboard"

// notOnboarded models a repository whose readiness check has not passed.
var notOnboarded = bus.ReadinessResult{Ready: false, Score: 20}

func TestMissionControlInvitesOnboardWhenNotOnboarded(t *testing.T) {
	// Flat Mission Control screen (plain/json parity path).
	flat := stripANSI(mission.Render(mission.ViewModel{
		RuntimeStatus: "stopped",
		Readiness:     notOnboarded,
		Theme:         theme.Default(),
	}))
	require.Contains(t, flat, onboardInvite,
		"flat Mission Control screen must invite `asa onboard` when not onboarded")

	// Panelised cockpit path also surfaces the readiness banner with the invite.
	cockpit := stripANSI(mission.RenderCockpit(mission.ViewModel{
		RuntimeStatus: "stopped",
		Readiness:     notOnboarded,
		Theme:         theme.Default(),
		Width:         120,
		Height:        40,
	}))
	require.Contains(t, cockpit, onboardInvite,
		"Mission Control cockpit must invite `asa onboard` when not onboarded")
}

func TestRunsInvitesOnboardWhenNotOnboarded(t *testing.T) {
	got := stripANSI(runs.Render(runs.ViewModel{
		Runs:      []bus.RunSummary{{ID: "run-1", Feature: "x", Status: "running"}},
		Readiness: notOnboarded,
		Theme:     theme.Default(),
		Width:     120,
		Height:    30,
	}))
	require.Contains(t, got, onboardInvite,
		"Runs screen must invite `asa onboard` when not onboarded")
}
