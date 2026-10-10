package keycloak

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func roles(names ...*string) []*gocloak.Role {
	out := make([]*gocloak.Role, 0, len(names))
	for _, n := range names {
		out = append(out, &gocloak.Role{Name: n})
	}
	return out
}

func TestSystemRole(t *testing.T) {
	cases := []struct {
		name  string
		roles []*gocloak.Role
		want  domain.SystemRole
	}{
		{"member only", roles(gocloak.StringP("member")), domain.SystemRoleTeamMember},
		{"auditor only", roles(gocloak.StringP("auditor")), domain.SystemRoleAuditor},
		{"both", roles(gocloak.StringP("member"), gocloak.StringP("auditor")), ""},
		{"neither", roles(gocloak.StringP("offline_access")), ""},
		{"no roles", nil, ""},
		{"nil name entries ignored", roles(nil, gocloak.StringP("member")), domain.SystemRoleTeamMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := systemRole(tc.roles); got != tc.want {
				t.Errorf("systemRole = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDisplayName(t *testing.T) {
	cases := []struct {
		name string
		user *gocloak.User
		want string
	}{
		{"first and last", &gocloak.User{FirstName: gocloak.StringP("Ada"), LastName: gocloak.StringP("Lovelace")}, "Ada Lovelace"},
		{"first only", &gocloak.User{FirstName: gocloak.StringP("Ada")}, "Ada"},
		{"falls back to username", &gocloak.User{Username: gocloak.StringP("ada")}, "ada"},
		{"blank names fall back to username", &gocloak.User{FirstName: gocloak.StringP(" "), LastName: gocloak.StringP(""), Username: gocloak.StringP("ada")}, "ada"},
		{"all nil", &gocloak.User{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := displayName(tc.user); got != tc.want {
				t.Errorf("displayName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToMember(t *testing.T) {
	attrs := map[string][]string{
		"chura_project_role_p1": {"product_owner"},
		"chura_project_role_p2": {},
	}
	cases := []struct {
		name      string
		projectID string
		user      *gocloak.User
		want      domain.Member
	}{
		{"nil attributes", "p1", &gocloak.User{ID: gocloak.StringP("u1"), Email: gocloak.StringP("a@x.io"), Username: gocloak.StringP("ada")},
			domain.Member{UserID: "u1", Email: "a@x.io", Name: "ada"}},
		{"role attribute present", "p1", &gocloak.User{ID: gocloak.StringP("u1"), FirstName: gocloak.StringP("Ada"), Attributes: &attrs},
			domain.Member{UserID: "u1", Name: "Ada", Role: "product_owner"}},
		{"attribute for another project", "p3", &gocloak.User{ID: gocloak.StringP("u1"), Attributes: &attrs},
			domain.Member{UserID: "u1"}},
		{"empty attribute values", "p2", &gocloak.User{ID: gocloak.StringP("u1"), Attributes: &attrs},
			domain.Member{UserID: "u1"}},
		{"all nil pointers", "p1", &gocloak.User{}, domain.Member{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := toMember(tc.user, tc.projectID); got != tc.want {
				t.Errorf("toMember = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	other := errors.New("boom")
	cases := []struct {
		name         string
		err          error
		wantNotFound bool
		wantSame     bool
	}{
		{"gocloak 404", &gocloak.APIError{Code: 404, Message: "gone"}, true, false},
		{"wrapped gocloak 404", fmt.Errorf("ctx: %w", &gocloak.APIError{Code: 404}), true, false},
		{"gocloak 500", &gocloak.APIError{Code: 500}, false, true},
		{"gocloak 401", &gocloak.APIError{Code: 401}, false, true},
		{"plain error", other, false, true},
		{"nil", nil, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classify(tc.err)
			if errors.Is(got, domain.ErrNotFound) != tc.wantNotFound {
				t.Errorf("classify(%v) = %v, ErrNotFound = %v, want %v", tc.err, got, !tc.wantNotFound, tc.wantNotFound)
			}
			if tc.wantSame && got != tc.err {
				t.Errorf("classify(%v) = %v, want error unchanged", tc.err, got)
			}
		})
	}
}

func TestCollectPages(t *testing.T) {
	pageOf := func(total int) func(first, max int) ([]int, error) {
		return func(first, max int) ([]int, error) {
			var page []int
			for i := first; i < first+max && i < total; i++ {
				page = append(page, i)
			}
			return page, nil
		}
	}
	errFetch := errors.New("page 2 failed")
	cases := []struct {
		name     string
		fetch    func(first, max int) ([]int, error)
		pageSize int
		ceiling  int
		wantLen  int
		wantErr  error
	}{
		{"empty", pageOf(0), 3, 100, 0, nil},
		{"less than one page", pageOf(2), 3, 100, 2, nil},
		{"exactly one page then empty", pageOf(3), 3, 100, 3, nil},
		{"several pages", pageOf(8), 3, 100, 8, nil},
		{"more than old 500 cap", pageOf(1234), 100, 10_000, 1234, nil},
		{"at ceiling", pageOf(9), 3, 9, 9, nil},
		{"over ceiling", pageOf(10), 3, 9, 0, errTooManyMembers},
		{"fetch error", func(first, max int) ([]int, error) {
			if first >= 3 {
				return nil, errFetch
			}
			return pageOf(10)(first, max)
		}, 3, 100, 0, errFetch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := collectPages(tc.fetch, tc.pageSize, tc.ceiling)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if len(got) != tc.wantLen {
				t.Errorf("len = %d, want %d", len(got), tc.wantLen)
			}
			for i, v := range got {
				if v != i {
					t.Fatalf("got[%d] = %d, want in-order unique items", i, v)
				}
			}
		})
	}
}

func TestTokenSource_NeedsRefresh(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		token string
		now   time.Time
		want  bool
	}{
		{"fresh", "cached", base, false},
		{"just before margin", "cached", base.Add(time.Minute - tokenRefreshMargin - time.Second), false},
		{"inside margin", "cached", base.Add(time.Minute - tokenRefreshMargin), true},
		{"expired", "cached", base.Add(2 * time.Minute), true},
		{"no token yet", "", base, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &tokenSource{now: func() time.Time { return tc.now }, token: tc.token, expires: base.Add(time.Minute)}
			if got := s.needsRefresh(); got != tc.want {
				t.Errorf("needsRefresh = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTokenSource_InvalidateForcesRefresh(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s := &tokenSource{now: func() time.Time { return base }, token: "cached", expires: base.Add(time.Hour)}
	if s.needsRefresh() {
		t.Fatal("needsRefresh = true for fresh token")
	}
	s.invalidate()
	if !s.needsRefresh() {
		t.Fatal("needsRefresh = false after invalidate")
	}
}
