package service_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// writeDelay is the pause before a directory write lands.
const writeDelay = 2 * time.Millisecond

// callLog records the order of calls across the fake repository and directory.
type callLog struct {
	mu    sync.Mutex
	calls []string
}

func (l *callLog) record(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calls = append(l.calls, fmt.Sprintf(format, args...))
}

// fakeProjectRepo is an in-memory out.ProjectRepository.
type fakeProjectRepo struct {
	projects  map[string]domain.Project
	writes    int // number of mutating calls, to prove read-only operations persist nothing
	log       *callLog
	createErr error
	lockMu    sync.Mutex // real mutex so concurrency tests exercise serialisation
	locks     int        // times WithMembershipLock was taken
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{projects: map[string]domain.Project{}, log: &callLog{}}
}

func (r *fakeProjectRepo) Create(_ context.Context, p *domain.Project) error {
	r.writes++
	r.log.record("repo.Create(%s)", p.ID)
	if r.createErr != nil {
		return r.createErr
	}
	c := *p
	c.Members = nil
	r.projects[p.ID] = c
	return nil
}

func (r *fakeProjectRepo) GetByID(_ context.Context, id string) (*domain.Project, error) {
	p, ok := r.projects[id]
	if !ok {
		return nil, fmt.Errorf("project %s: %w", id, domain.ErrNotFound)
	}
	return &p, nil
}

func (r *fakeProjectRepo) WithMembershipLock(ctx context.Context, projectID string, fn func(ctx context.Context) error) error {
	r.lockMu.Lock()
	defer r.lockMu.Unlock()
	r.locks++
	r.log.record("repo.WithMembershipLock(%s)", projectID)
	return fn(ctx)
}

// fakeDirectory is an in-memory out.ProjectDirectory. errs injects a failure
// per method name; every call is recorded, with its arguments, in log.
type fakeDirectory struct {
	mu      sync.Mutex
	users   map[string]domain.DirectoryUser // by email
	log     *callLog
	errs    map[string]error
	members map[string][]domain.Member // by group id
	groups  map[string]string          // project id -> group id
}

func newFakeDirectory(log *callLog) *fakeDirectory {
	teamMemberUser := func(id string) domain.DirectoryUser {
		return domain.DirectoryUser{ID: id, Email: id + "@example.com", Name: "User " + id, Role: domain.SystemRoleTeamMember}
	}
	users := map[string]domain.DirectoryUser{
		"u1@example.com": teamMemberUser("u1"),
		"u2@example.com": teamMemberUser("u2"),
	}
	return &fakeDirectory{users: users, log: log, errs: map[string]error{}, members: map[string][]domain.Member{}, groups: map[string]string{}}
}

func (d *fakeDirectory) CreateProjectGroup(_ context.Context, projectID string) (string, error) {
	d.log.record("CreateProjectGroup(%s)", projectID)
	if err := d.errs["CreateProjectGroup"]; err != nil {
		return "", err
	}
	groupID := "group-" + projectID[:1]
	d.groups[projectID] = groupID
	return groupID, nil
}

func (d *fakeDirectory) DeleteProjectGroup(_ context.Context, groupID string) error {
	d.log.record("DeleteProjectGroup(%s)", groupID)
	return d.errs["DeleteProjectGroup"]
}

func (d *fakeDirectory) FindUserByEmail(_ context.Context, email string) (domain.DirectoryUser, error) {
	d.log.record("FindUserByEmail(%s)", email)
	if err := d.errs["FindUserByEmail"]; err != nil {
		return domain.DirectoryUser{}, err
	}
	user, ok := d.users[email]
	if !ok {
		return domain.DirectoryUser{}, fmt.Errorf("user %s: %w", email, domain.ErrNotFound)
	}
	return user, nil
}

func (d *fakeDirectory) ListMembers(_ context.Context, groupID, _ string) ([]domain.Member, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log.record("ListMembers(%s)", groupID)
	if err := d.errs["ListMembers"]; err != nil {
		return nil, err
	}
	return append([]domain.Member(nil), d.members[groupID]...), nil
}

func (d *fakeDirectory) AddMember(_ context.Context, groupID, projectID, userID string, role domain.ProjectRole) error {
	time.Sleep(writeDelay) // widens the check-then-write window so a missing lock is caught
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log.record("AddMember(%s,%s,%s)", groupID, userID, role)
	if err := d.errs["AddMember"]; err != nil {
		return err
	}
	d.groups[projectID] = groupID
	d.members[groupID] = append(d.members[groupID], domain.Member{
		UserID: userID, Email: userID + "@example.com", Name: "User " + userID, Role: role,
	})
	return nil
}

func (d *fakeDirectory) SetMemberRole(_ context.Context, projectID, userID string, role domain.ProjectRole) error {
	time.Sleep(writeDelay) // widens the check-then-write window so a missing lock is caught
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log.record("SetMemberRole(%s,%s)", userID, role)
	if err := d.errs["SetMemberRole"]; err != nil {
		return err
	}
	members := d.members[d.groups[projectID]]
	for i := range members {
		if members[i].UserID == userID {
			members[i].Role = role
		}
	}
	return nil
}

func (d *fakeDirectory) RemoveMember(_ context.Context, groupID, _, userID string) error {
	d.log.record("RemoveMember(%s,%s)", groupID, userID)
	return d.errs["RemoveMember"]
}
