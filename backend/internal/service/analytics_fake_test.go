package service_test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// fakeTemplates is an out.TemplateGateway backed by the real catalog.
type fakeTemplates struct {
	calls int
	err   error
}

func (f *fakeTemplates) RetrieveTemplateDefinition(_ context.Context, _ domain.Actor, id string) (domain.Template, error) {
	f.calls++
	if f.err != nil {
		return domain.Template{}, f.err
	}
	t, ok := domain.TemplateByID(id)
	if !ok {
		return domain.Template{}, fmt.Errorf("template %q: %w", id, domain.ErrNotFound)
	}
	return t, nil
}

// fakeSprintWorkItems is an out.SprintWorkItemReader returning canned items.
type fakeSprintWorkItems struct {
	items                 []domain.WorkItem
	err                   error
	calls                 int
	gotProject, gotSprint string
}

func (f *fakeSprintWorkItems) ListBySprint(_ context.Context, projectID, sprintID string) ([]domain.WorkItem, error) {
	f.calls++
	f.gotProject, f.gotSprint = projectID, sprintID
	return f.items, f.err
}

// fakeReflections is an in-memory out.ReflectionRepository keyed by (sprint, author).
type fakeReflections struct {
	rows  []domain.SprintReflection
	err   error
	calls int
}

func (f *fakeReflections) Upsert(_ context.Context, r *domain.SprintReflection) (*domain.SprintReflection, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.rows {
		if f.rows[i].SprintID == r.SprintID && f.rows[i].AuthorID == r.AuthorID {
			created := f.rows[i].CreatedAt
			f.rows[i] = *r
			f.rows[i].ID, f.rows[i].CreatedAt = strconv.Itoa(i+1), created
			c := f.rows[i]
			return &c, nil
		}
	}
	c := *r
	c.ID = strconv.Itoa(len(f.rows) + 1)
	f.rows = append(f.rows, c)
	return &c, nil
}

func (f *fakeReflections) ListBySprint(_ context.Context, sprintID string) ([]domain.SprintReflection, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	var out []domain.SprintReflection
	for _, r := range f.rows {
		if r.SprintID == sprintID {
			out = append(out, r)
		}
	}
	return out, nil
}

// fakeExporter renders the sprint name so tests can see what it received.
type fakeExporter struct{ got *domain.SprintReviewSummary }

func (*fakeExporter) Format() domain.ReportFormat { return domain.ReportFormatCSV }
func (*fakeExporter) ContentType() string         { return "text/csv" }
func (*fakeExporter) FileExtension() string       { return "csv" }
func (e *fakeExporter) Export(s *domain.SprintReviewSummary) ([]byte, error) {
	e.got = s
	return []byte("report:" + s.Sprint.Name), nil
}
