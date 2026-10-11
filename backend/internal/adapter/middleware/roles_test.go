package middleware

import (
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestSystemRoleFrom(t *testing.T) {
	cases := []struct {
		name  string
		roles []string
		want  domain.SystemRole
		ok    bool
	}{
		{"member", []string{"member"}, domain.SystemRoleTeamMember, true},
		{"auditor", []string{"auditor"}, domain.SystemRoleAuditor, true},
		{"with unrelated", []string{"offline_access", "auditor"}, domain.SystemRoleAuditor, true},
		{"none", nil, "", false},
		{"unrelated only", []string{"admin"}, "", false},
		{"both", []string{"member", "auditor"}, "", false},
		{"case sensitive", []string{"Member"}, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := systemRoleFrom(tc.roles)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}
