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
		{"team member", domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}, nil},
		{"auditor", domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}, domain.ErrForbidden},
		{"blank user", domain.Actor{UserID: " ", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}, domain.ErrUnauthenticated},
		{"blank email", domain.Actor{UserID: "u1", Email: "  ", Role: domain.SystemRoleTeamMember}, domain.ErrUnauthenticated},
		{"missing email", domain.Actor{UserID: "u1", Role: domain.SystemRoleAuditor}, domain.ErrUnauthenticated},
		{"unknown role", domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, domain.ErrUnauthenticated},
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

func TestActor_Validate(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		want  error
	}{
		{"valid member", domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}, nil},
		{"valid auditor", domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}, nil},
		{"blank email", domain.Actor{UserID: "u1", Email: " ", Role: domain.SystemRoleAuditor}, domain.ErrUnauthenticated},
		{"empty email", domain.Actor{UserID: "u1", Role: domain.SystemRoleTeamMember}, domain.ErrUnauthenticated},
		{"blank user", domain.Actor{Email: "u1@example.com", Role: domain.SystemRoleTeamMember}, domain.ErrUnauthenticated},
		{"no role", domain.Actor{UserID: "u1", Email: "u1@example.com"}, domain.ErrUnauthenticated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.actor.Validate()
			if tc.want == nil && err != nil || tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}
