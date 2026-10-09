package domain

import (
	"fmt"
	"strings"
)

type SystemRole string

const (
	SystemRoleTeamMember SystemRole = "team_member"
	SystemRoleAuditor    SystemRole = "auditor"
)

type Actor struct {
	UserID string
	Role   SystemRole
}

// Validate reports ErrUnauthenticated for a blank user id or unknown role.
func (a Actor) Validate() error {
	if strings.TrimSpace(a.UserID) == "" {
		return fmt.Errorf("missing user id: %w", ErrUnauthenticated)
	}
	if a.Role != SystemRoleTeamMember && a.Role != SystemRoleAuditor {
		return fmt.Errorf("unknown role %q: %w", a.Role, ErrUnauthenticated)
	}
	return nil
}

// RequireTeamMember reports ErrUnauthenticated for an invalid actor and
// ErrForbidden for any valid actor who is not a Team Member (e.g. Auditors).
func (a Actor) RequireTeamMember() error {
	if err := a.Validate(); err != nil {
		return err
	}
	if a.Role != SystemRoleTeamMember {
		return fmt.Errorf("only team members may modify: %w", ErrForbidden)
	}
	return nil
}
