package service_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
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
	projects   map[string]domain.Project
	auditors   map[string][]domain.ProjectAuditorAssignment // by project id
	feedback   map[string][]domain.ProjectFeedback          // by project id
	writes     int                                         // number of mutating calls, to prove read-only operations persist nothing
	log        *callLog
	createErr  error
	updateErr  error
	deleteErr  error
	assignErr  error
	feedbackErr error
	listQuery  out.ProjectListQuery
	listErr    error
	lockMu     sync.Mutex // real mutex so concurrency tests exercise serialisation
	locks      int        // times WithMembershipLock was taken
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{
		projects: map[string]domain.Project{},
		auditors: map[string][]domain.ProjectAuditorAssignment{},
		feedback: map[string][]domain.ProjectFeedback{},
		log:      &callLog{},
	}
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

func (r *fakeProjectRepo) ListSummaries(_ context.Context, query out.ProjectListQuery) ([]domain.ProjectSummary, int, error) {
	r.log.record("repo.ListSummaries")
	r.listQuery = query
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	allowed := make(map[string]struct{}, len(query.ProjectIDs))
	for _, id := range query.ProjectIDs {
		allowed[id] = struct{}{}
	}
	summaries := make([]domain.ProjectSummary, 0, len(r.projects))
	for id, project := range r.projects {
		if query.ProjectIDs != nil {
			if _, ok := allowed[id]; !ok {
				continue
			}
		}
		summaries = append(summaries, domain.ProjectSummary{
			ID: project.ID, Name: project.Name, Description: project.Description,
			TemplateID: project.TemplateID, Mode: project.Mode,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Name == summaries[j].Name {
			return summaries[i].ID < summaries[j].ID
		}
		return summaries[i].Name < summaries[j].Name
	})
	total := len(summaries)
	start := min(query.Offset, total)
	end := min(start+query.Limit, total)
	return summaries[start:end], total, nil
}

func (r *fakeProjectRepo) WithMembershipLock(ctx context.Context, projectID string, fn func(ctx context.Context) error) error {
	r.lockMu.Lock()
	defer r.lockMu.Unlock()
	r.locks++
	r.log.record("repo.WithMembershipLock(%s)", projectID)
	return fn(ctx)
}

func (r *fakeProjectRepo) Update(_ context.Context, p *domain.Project) error {
	r.writes++
	r.log.record("repo.Update(%s)", p.ID)
	if r.updateErr != nil {
		return r.updateErr
	}
	current, ok := r.projects[p.ID]
	if !ok {
		return fmt.Errorf("project %s: %w", p.ID, domain.ErrNotFound)
	}
	current.Name = p.Name
	current.Description = p.Description
	current.TemplateID = p.TemplateID
	current.Mode = p.Mode
	current.UpdatedAt = p.UpdatedAt
	r.projects[p.ID] = current
	return nil
}

func (r *fakeProjectRepo) Delete(_ context.Context, projectID string) error {
	r.writes++
	r.log.record("repo.Delete(%s)", projectID)
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if _, ok := r.projects[projectID]; !ok {
		return fmt.Errorf("project %s: %w", projectID, domain.ErrNotFound)
	}
	delete(r.projects, projectID)
	delete(r.auditors, projectID)
	delete(r.feedback, projectID)
	return nil
}

func (r *fakeProjectRepo) AssignAuditor(_ context.Context, assignment domain.ProjectAuditorAssignment) (domain.ProjectAuditorAssignment, error) {
	r.writes++
	r.log.record("repo.AssignAuditor(%s,%s)", assignment.ProjectID, assignment.AuditorUserID)
	if r.assignErr != nil {
		return domain.ProjectAuditorAssignment{}, r.assignErr
	}
	for _, existing := range r.auditors[assignment.ProjectID] {
		if existing.AuditorUserID == assignment.AuditorUserID {
			return existing, nil
		}
	}
	r.auditors[assignment.ProjectID] = append(r.auditors[assignment.ProjectID], assignment)
	return assignment, nil
}

func (r *fakeProjectRepo) ListAuditors(_ context.Context, projectID string) ([]domain.ProjectAuditorAssignment, error) {
	r.log.record("repo.ListAuditors(%s)", projectID)
	return append([]domain.ProjectAuditorAssignment(nil), r.auditors[projectID]...), nil
}

func (r *fakeProjectRepo) IsAuditorAssigned(_ context.Context, projectID, auditorUserID string) (bool, error) {
	r.log.record("repo.IsAuditorAssigned(%s,%s)", projectID, auditorUserID)
	for _, assignment := range r.auditors[projectID] {
		if assignment.AuditorUserID == auditorUserID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeProjectRepo) CreateFeedback(_ context.Context, feedback *domain.ProjectFeedback) error {
	r.writes++
	r.log.record("repo.CreateFeedback(%s)", feedback.ProjectID)
	if r.feedbackErr != nil {
		return r.feedbackErr
	}
	r.feedback[feedback.ProjectID] = append([]domain.ProjectFeedback{*feedback}, r.feedback[feedback.ProjectID]...)
	return nil
}

func (r *fakeProjectRepo) ListFeedback(_ context.Context, projectID string) ([]domain.ProjectFeedback, error) {
	r.log.record("repo.ListFeedback(%s)", projectID)
	return append([]domain.ProjectFeedback(nil), r.feedback[projectID]...), nil
}

// fakeSprintStore counts and hard-deletes sprints for project delete/update rules.
type fakeSprintStore struct {
	counts map[string]int
	log    *callLog
	err    error
}

func (s *fakeSprintStore) CountByProject(_ context.Context, projectID string) (int, error) {
	s.log.record("sprints.CountByProject(%s)", projectID)
	if s.err != nil {
		return 0, s.err
	}
	return s.counts[projectID], nil
}

func (s *fakeSprintStore) HardDeleteByProject(_ context.Context, projectID string) error {
	s.log.record("sprints.HardDeleteByProject(%s)", projectID)
	if s.err != nil {
		return s.err
	}
	s.counts[projectID] = 0
	return nil
}

type fakeWorkItemStore struct {
	items map[string][]domain.WorkItem
	log   *callLog
	err   error
}

func (g *fakeWorkItemStore) CountByProject(_ context.Context, projectID string) (int, error) {
	g.log.record("workItems.CountByProject(%s)", projectID)
	if g.err != nil {
		return 0, g.err
	}
	return len(g.items[projectID]), nil
}

func (g *fakeWorkItemStore) DeleteByProject(_ context.Context, projectID string) error {
	g.log.record("workItems.DeleteByProject(%s)", projectID)
	if g.err != nil {
		return g.err
	}
	delete(g.items, projectID)
	return nil
}

// fakeDirectory is an in-memory out.ProjectDirectory. errs injects a failure
// per method name; every call is recorded, with its arguments, in log.
type fakeDirectory struct {
	mu                 sync.Mutex
	users              map[string]domain.DirectoryUser // by email
	log                *callLog
	errs               map[string]error
	members            map[string][]domain.Member // by group id
	groups             map[string]string          // project id -> group id
	projectMemberships []string
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

func (d *fakeDirectory) FindUserByID(_ context.Context, userID string) (domain.DirectoryUser, error) {
	d.log.record("FindUserByID(%s)", userID)
	if err := d.errs["FindUserByID"]; err != nil {
		return domain.DirectoryUser{}, err
	}
	for _, user := range d.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return domain.DirectoryUser{}, fmt.Errorf("user %s: %w", userID, domain.ErrNotFound)
}

func (d *fakeDirectory) ListAuditors(_ context.Context) ([]domain.DirectoryUser, error) {
	d.log.record("ListAuditors")
	if err := d.errs["ListAuditors"]; err != nil {
		return nil, err
	}
	var auditors []domain.DirectoryUser
	for _, user := range d.users {
		if user.Role == domain.SystemRoleAuditor {
			auditors = append(auditors, user)
		}
	}
	return auditors, nil
}

func (d *fakeDirectory) ListProjectMemberships(_ context.Context, userID string) ([]string, error) {
	d.log.record("ListProjectMemberships(%s)", userID)
	if err := d.errs["ListProjectMemberships"]; err != nil {
		return nil, err
	}
	return append([]string(nil), d.projectMemberships...), nil
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
	d.mu.Lock()
	defer d.mu.Unlock()
	d.log.record("RemoveMember(%s,%s)", groupID, userID)
	if err := d.errs["RemoveMember"]; err != nil {
		return err
	}
	members := d.members[groupID]
	for i := range members {
		if members[i].UserID == userID {
			d.members[groupID] = append(members[:i], members[i+1:]...)
			break
		}
	}
	return nil
}
