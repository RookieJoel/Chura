// Package keycloak implements out.ProjectDirectory on the Keycloak Admin API.
// A Project is a group named "project-<id>"; a member's Project Role is the
// single-valued user attribute "chura_project_role_<id>", writable by admins
// only (user profile unmanagedAttributePolicy ADMIN_EDIT).
package keycloak

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/RookieJoel/Chura/backend/internal/adapter/lockstripe"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

const (
	callTimeout      = 5 * time.Second
	groupNamePrefix  = "project-"
	roleAttrPrefix   = "chura_project_role_"
	memberPageSize   = 100
	maxGroupMembers  = 10_000 // hard ceiling; beyond it ListMembers fails rather than truncating
	realmRoleMember  = "member"
	realmRoleAuditor = "auditor"
)

// Config locates the realm and holds the confidential client used for the Admin API.
type Config struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

type Directory struct {
	client *gocloak.GoCloak
	realm  string
	tokens *tokenSource
	// userLocks serialises attribute read-modify-write per user within this
	// process: the Admin API replaces the whole attribute map on PUT.
	userLocks *lockstripe.Set
}

var _ out.ProjectDirectory = (*Directory)(nil)

func NewDirectory(cfg Config) *Directory {
	client := gocloak.NewClient(cfg.BaseURL)
	return &Directory{
		client:    client,
		realm:     cfg.Realm,
		userLocks: lockstripe.New(),
		tokens: &tokenSource{
			client: client, realm: cfg.Realm, clientID: cfg.ClientID, clientSecret: cfg.ClientSecret,
			now: time.Now,
		},
	}
}

var errTooManyMembers = errors.New("group has more members than the supported maximum")

// classify turns a Keycloak 404 into domain.ErrNotFound; other errors pass through unchanged.
func classify(err error) error {
	var apiErr *gocloak.APIError
	if errors.As(err, &apiErr) && apiErr.Code == 404 {
		return fmt.Errorf("keycloak resource: %w", domain.ErrNotFound)
	}
	return err
}

// collectPages gathers every item by calling fetch(first, max) page by page
// until a page is shorter than pageSize. Exceeding ceiling is an error, never
// a silently truncated result.
func collectPages[T any](fetch func(first, max int) ([]T, error), pageSize, ceiling int) ([]T, error) {
	var all []T
	for first := 0; ; first += pageSize {
		page, err := fetch(first, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(all) > ceiling {
			return nil, errTooManyMembers
		}
		if len(page) < pageSize {
			return all, nil
		}
	}
}

// call runs fn with a bounded context and a service-account token. Anything
// that is not a definite domain answer (domain.ErrNotFound from fn) becomes
// domain.ErrUnavailable; the cause is logged, never returned, so no upstream
// detail reaches callers.
func (d *Directory) call(ctx context.Context, op string, fn func(ctx context.Context, token string) error) error {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	token, err := d.tokens.get(ctx)
	if err == nil {
		err = fn(ctx, token)
	}
	err = classify(err)
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("keycloak %s: %w", op, err)
	}
	var apiErr *gocloak.APIError
	if errors.As(err, &apiErr) && apiErr.Code == 401 {
		d.tokens.invalidate()
	}
	slog.Error("keycloak call failed", "op", op, "error", err)
	return fmt.Errorf("keycloak %s: %w", op, domain.ErrUnavailable)
}

func (d *Directory) CreateProjectGroup(ctx context.Context, projectID string) (string, error) {
	var groupID string
	err := d.call(ctx, "create group", func(ctx context.Context, token string) error {
		id, err := d.client.CreateGroup(ctx, token, d.realm, gocloak.Group{Name: gocloak.StringP(groupNamePrefix + projectID)})
		if err != nil {
			return err
		}
		if id == "" {
			return errors.New("group created without a Location header")
		}
		groupID = id
		return nil
	})
	return groupID, err
}

// DeleteProjectGroup is idempotent: an already-missing group is success.
func (d *Directory) DeleteProjectGroup(ctx context.Context, groupID string) error {
	err := d.call(ctx, "delete group", func(ctx context.Context, token string) error {
		return d.client.DeleteGroup(ctx, token, d.realm, groupID)
	})
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	return err
}

func (d *Directory) FindUserByEmail(ctx context.Context, email string) (domain.DirectoryUser, error) {
	var found domain.DirectoryUser
	err := d.call(ctx, "find user by email", func(ctx context.Context, token string) error {
		users, err := d.client.GetUsers(ctx, token, d.realm, gocloak.GetUsersParams{
			Email: gocloak.StringP(strings.ToLower(strings.TrimSpace(email))),
			Exact: gocloak.BoolP(true),
		})
		if err != nil {
			return err
		}
		if len(users) == 0 || gocloak.PString(users[0].ID) == "" {
			return fmt.Errorf("user with this email: %w", domain.ErrNotFound)
		}
		user := users[0]
		roles, err := d.client.GetCompositeRealmRolesByUserID(ctx, token, d.realm, *user.ID)
		if err != nil {
			return err
		}
		found = domain.DirectoryUser{
			ID: *user.ID, Email: gocloak.PString(user.Email), Name: displayName(user), Role: systemRole(roles),
		}
		return nil
	})
	return found, err
}

// systemRole maps effective realm roles to a Chura system role; neither or both yield "".
func systemRole(roles []*gocloak.Role) domain.SystemRole {
	var isMember, isAuditor bool
	for _, r := range roles {
		switch gocloak.PString(r.Name) {
		case realmRoleMember:
			isMember = true
		case realmRoleAuditor:
			isAuditor = true
		}
	}
	switch {
	case isMember && !isAuditor:
		return domain.SystemRoleTeamMember
	case isAuditor && !isMember:
		return domain.SystemRoleAuditor
	default:
		return ""
	}
}

func displayName(u *gocloak.User) string {
	name := strings.TrimSpace(gocloak.PString(u.FirstName) + " " + gocloak.PString(u.LastName))
	if name == "" {
		return gocloak.PString(u.Username)
	}
	return name
}

func (d *Directory) ListMembers(ctx context.Context, groupID, projectID string) ([]domain.Member, error) {
	var members []domain.Member
	err := d.call(ctx, "list group members", func(ctx context.Context, token string) error {
		users, err := collectPages(func(first, max int) ([]*gocloak.User, error) {
			return d.client.GetGroupMembers(ctx, token, d.realm, groupID, gocloak.GetGroupsParams{
				BriefRepresentation: gocloak.BoolP(false),
				First:               gocloak.IntP(first),
				Max:                 gocloak.IntP(max),
			})
		}, memberPageSize, maxGroupMembers)
		if err != nil {
			return err
		}
		members = make([]domain.Member, 0, len(users))
		for _, u := range users {
			members = append(members, toMember(u, projectID))
		}
		return nil
	})
	return members, err
}

func toMember(u *gocloak.User, projectID string) domain.Member {
	member := domain.Member{
		UserID: gocloak.PString(u.ID), Email: gocloak.PString(u.Email), Name: displayName(u),
	}
	if u.Attributes != nil {
		if values := (*u.Attributes)[roleAttrPrefix+projectID]; len(values) > 0 {
			member.Role = domain.ProjectRole(values[0])
		}
	}
	if member.Role == "" {
		slog.Warn("project member has no project role attribute", "project_id", projectID, "user_id", member.UserID)
	}
	return member
}

func (d *Directory) AddMember(ctx context.Context, groupID, projectID, userID string, role domain.ProjectRole) error {
	return d.call(ctx, "add member", func(ctx context.Context, token string) error {
		if err := d.client.AddUserToGroup(ctx, token, d.realm, userID, groupID); err != nil {
			return err
		}
		if err := d.writeRoleAttribute(ctx, token, projectID, userID, string(role)); err != nil {
			// Best effort: do not leave a member without a role behind.
			leaveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), callTimeout)
			defer cancel()
			if leaveErr := d.client.DeleteUserFromGroup(leaveCtx, token, d.realm, userID, groupID); leaveErr != nil {
				slog.Error("keycloak undo add member failed", "group_id", groupID, "user_id", userID, "error", leaveErr)
			}
			return err
		}
		return nil
	})
}

func (d *Directory) SetMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error {
	return d.call(ctx, "set member role", func(ctx context.Context, token string) error {
		return d.writeRoleAttribute(ctx, token, projectID, userID, string(role))
	})
}

// RemoveMember is idempotent: a missing user or group is success.
func (d *Directory) RemoveMember(ctx context.Context, groupID, projectID, userID string) error {
	err := d.call(ctx, "remove member", func(ctx context.Context, token string) error {
		if err := d.writeRoleAttribute(ctx, token, projectID, userID, ""); err != nil {
			return err
		}
		return d.client.DeleteUserFromGroup(ctx, token, d.realm, userID, groupID)
	})
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	return err
}

// writeRoleAttribute sets (or, for an empty value, deletes) the user's Project
// Role attribute. The Admin API replaces the whole user on PUT, so the user is
// read first and written back with every other field and attribute intact,
// under a per-user mutex so concurrent writes for one user do not overwrite
// each other.
func (d *Directory) writeRoleAttribute(ctx context.Context, token, projectID, userID, value string) error {
	unlock, err := d.userLocks.Lock(ctx, userID)
	if err != nil {
		return err
	}
	defer unlock()
	user, err := d.client.GetUserByID(ctx, token, d.realm, userID)
	if err != nil {
		return err
	}
	attrs := map[string][]string{}
	if user.Attributes != nil {
		for k, v := range *user.Attributes {
			attrs[k] = v
		}
	}
	key := roleAttrPrefix + projectID
	if value == "" {
		delete(attrs, key)
	} else {
		attrs[key] = []string{value}
	}
	user.Attributes = &attrs
	return d.client.UpdateUser(ctx, token, d.realm, *user)
}
