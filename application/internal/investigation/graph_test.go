package investigation_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/stretchr/testify/require"
)

func TestBuildRootCauseGraph(t *testing.T) {
	t.Parallel()
	rep := investigation.Report{
		Request: investigation.Request{Symptom: "invite fails"},
		Scope:   investigation.ResolvedScope{Flow: "onboarding"},
		Evidence: []investigation.Evidence{
			{ID: "e1", Kind: "test", Summary: "InvitationTest failed"},
		},
		Hypotheses: []investigation.Hypothesis{{ID: "h1", Statement: "email not sent", Score: 0.8}},
		LocalResult: investigation.InvestigationResult{CandidateFiles: []string{"application/internal/foo.go"}},
	}
	g := investigation.BuildRootCauseGraph(rep, investigation.ContextPack{
		APIs:   []string{"POST /invitations"},
		Events: []string{"member.invited"},
	})
	require.NotEmpty(t, g.Nodes)
	require.NotEmpty(t, g.Edges)
}
