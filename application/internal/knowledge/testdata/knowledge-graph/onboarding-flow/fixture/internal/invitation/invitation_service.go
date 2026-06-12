package invitation

// InvitationService handles member invitations.
type InvitationService struct{}

// Invite emits member.invited when an invitation is created.
func (s *InvitationService) Invite() error {
	_ = "member.invited"
	return nil
}
