package cli

import (
	"bytes"
	"testing"

	uiapp "github.com/LaProgrammerie/asagiri/application/internal/ui/app"
	"github.com/stretchr/testify/require"
)

func TestRootNoArgsNonTTYShowsHelp(t *testing.T) {
	root := newRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{})

	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Usage:")
	require.Contains(t, out.String(), "Asagiri")
}

func TestUICommandsRegisteredAndNonTTYSafe(t *testing.T) {
	root := newRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetErr(out)

	missionCmd, _, err := root.Find([]string{"mission"})
	require.NoError(t, err)
	require.NotNil(t, missionCmd)
	dashboardCmd, _, err := root.Find([]string{"dashboard"})
	require.NoError(t, err)
	require.NotNil(t, dashboardCmd)
	explainCmd, _, err := root.Find([]string{"explain"})
	require.NoError(t, err)
	require.NotNil(t, explainCmd)
	agentsCmd, _, err := root.Find([]string{"agents"})
	require.NoError(t, err)
	require.NotNil(t, agentsCmd)
	flowCmd, _, err := root.Find([]string{"flow"})
	require.NoError(t, err)
	require.NotNil(t, flowCmd)
	graphCmd, _, err := root.Find([]string{"graph"})
	require.NoError(t, err)
	require.NotNil(t, graphCmd)
	knowledgeCmd, _, err := root.Find([]string{"knowledge"})
	require.NoError(t, err)
	require.NotNil(t, knowledgeCmd)
	trustCmd, _, err := root.Find([]string{"trust"})
	require.NoError(t, err)
	require.NotNil(t, trustCmd)

	root.SetArgs([]string{"mission"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"dashboard"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"explain"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"agents", "watch"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"flow"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"graph"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"knowledge"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"trust"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"prototype"})
	require.NoError(t, root.Execute())

	out.Reset()
	root.SetArgs([]string{"replay", "open", "replay-001"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Usage:")

	flowOpenCmd, _, err := root.Find([]string{"flow", "open"})
	require.NoError(t, err)
	require.NotNil(t, flowOpenCmd)

	out.Reset()
	root.SetArgs([]string{"flow", "open", "onboarding"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Usage:")
}

func TestFlowOpenOptionsSetupSetsFlowID(t *testing.T) {
	opts := uiapp.Options{InitialScreen: uiapp.ScreenFlow}
	flowOpenOptionsSetup([]string{"  onboarding  "}, &opts)
	require.Equal(t, "onboarding", opts.FlowID)
	require.Equal(t, uiapp.ScreenFlow, opts.InitialScreen)
}
