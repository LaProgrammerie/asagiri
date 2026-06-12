package invitation

import "testing"

func TestInvitationService(t *testing.T) {
	var svc InvitationService
	if err := svc.Invite(); err != nil {
		t.Fatal(err)
	}
}
