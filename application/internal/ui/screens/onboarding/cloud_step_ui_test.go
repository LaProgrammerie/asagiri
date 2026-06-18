package onboarding_test

import (
	"testing"

	onbdomain "github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/onboarding"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestCloudTestActionEnterDispatchesRequest(t *testing.T) {
	m := onboarding.NewModelFromForm(onbdomain.Form{
		Step: onbdomain.StepCloud,
		Answers: onbdomain.Answers{
			CloudMode:    "connect",
			CloudBaseURL: onbdomain.DefaultCloudOnboardBaseURL(),
			CloudToken:   "tok",
		},
	}, false)
	m.RefreshFieldRows()
	require.True(t, m.FocusFieldByKey("cloud_test"), "cloud_test row missing — vérifiez cloud_mode=connect")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	req, ok := msg.(onboarding.OnboardingTestCloudRequestMsg)
	require.True(t, ok)
	require.Equal(t, "connect", req.Form.Answers.CloudMode)
}
