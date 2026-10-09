package domain_test

import (
	"errors"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestActor_RequireTeamMember(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		want  error
	}{
		{"team member", domain.Actor{UserID: "u1", Role: domain.SystemRoleTeamMember}, nil},
		{"auditor", domain.Actor{UserID: "a1", Role: domain.SystemRoleAuditor}, domain.ErrForbidden},
		{"blank user", domain.Actor{UserID: " ", Role: domain.SystemRoleTeamMember}, domain.ErrUnauthenticated},
		{"unknown role", domain.Actor{UserID: "u1", Role: "admin"}, domain.ErrUnauthenticated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.actor.RequireTeamMember()
			if tc.want == nil && err != nil || tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}
