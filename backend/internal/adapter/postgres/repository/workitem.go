package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/RookieJoel/Chura/backend/internal/adapter/postgres/sqlcgen"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ driven.WorkItemRepository = (*WorkItemRepository)(nil)
var _ driven.WorkItemRepository = (*SQLWorkItemRepository)(nil)

type WorkItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.WorkItem
}

func NewWorkItemRepository() *WorkItemRepository {
	return &WorkItemRepository{items: make(map[string]domain.WorkItem)}
}

type SQLWorkItemRepository struct {
	queries *sqlcgen.Queries
}

func NewSQLWorkItemRepository(db sqlcgen.DBTX) *SQLWorkItemRepository {
	return &SQLWorkItemRepository{queries: sqlcgen.New(db)}
}

func (repository *WorkItemRepository) Create(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.items[item.ID] = item
	return item, nil
}

func (repository *WorkItemRepository) Get(id string) (domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	item, ok := repository.items[id]
	if !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (repository *WorkItemRepository) List(projectID string) ([]domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	items := make([]domain.WorkItem, 0)
	for _, item := range repository.items {
		if item.ProjectID == projectID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (repository *WorkItemRepository) Update(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[item.ID]; !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	repository.items[item.ID] = item
	return item, nil
}

func (repository *WorkItemRepository) Delete(id string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(repository.items, id)
	return nil
}

func (repository *SQLWorkItemRepository) Create(item domain.WorkItem) (domain.WorkItem, error) {
	ctx := context.Background()
	reporterIDs := append([]string(nil), item.ReporterIDs...)
	params := sqlcgen.CreateWorkItemParams{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        string(item.Type),
		Status:      string(item.Status),
		Priority:    string(item.Priority),
		AssigneeID:  pgtypeText(item.AssigneeID),
		StoryPoints: item.StoryPoints,
		Features:    marshalFeatures(item.Features),
	}
	created, err := repository.queries.CreateWorkItem(ctx, params)
	if err != nil {
		return domain.WorkItem{}, err
	}
	item = workItemFromSQL(created)
	if err := repository.syncReporterIDs(ctx, item.ID, reporterIDs); err != nil {
		return domain.WorkItem{}, err
	}
	item.ReporterIDs = append([]string(nil), reporterIDs...)
	return item, nil
}

func (repository *SQLWorkItemRepository) Get(id string) (domain.WorkItem, error) {
	ctx := context.Background()
	row, err := repository.queries.GetWorkItem(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItem{}, domain.ErrNotFound
		}
		return domain.WorkItem{}, err
	}
	item := workItemFromSQL(row)
	reporters, err := repository.queries.ListWorkItemReporterIDs(ctx, id)
	if err != nil {
		return domain.WorkItem{}, err
	}
	item.ReporterIDs = reporterIDsFromSQL(reporters)
	return item, nil
}

func (repository *SQLWorkItemRepository) List(projectID string) ([]domain.WorkItem, error) {
	ctx := context.Background()
	rows, err := repository.queries.ListWorkItems(ctx, projectID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.WorkItem, 0, len(rows))
	for _, row := range rows {
		item := workItemFromSQL(row)
		reporters, err := repository.queries.ListWorkItemReporterIDs(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		item.ReporterIDs = reporterIDsFromSQL(reporters)
		items = append(items, item)
	}
	return items, nil
}

func (repository *SQLWorkItemRepository) Update(item domain.WorkItem) (domain.WorkItem, error) {
	ctx := context.Background()
	reporterIDs := append([]string(nil), item.ReporterIDs...)
	updated, err := repository.queries.UpdateWorkItem(ctx, sqlcgen.UpdateWorkItemParams{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        string(item.Type),
		Status:      string(item.Status),
		Priority:    string(item.Priority),
		AssigneeID:  pgtypeText(item.AssigneeID),
		StoryPoints: item.StoryPoints,
		Features:    marshalFeatures(item.Features),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItem{}, domain.ErrNotFound
		}
		return domain.WorkItem{}, err
	}
	item = workItemFromSQL(updated)
	if err := repository.syncReporterIDs(ctx, item.ID, reporterIDs); err != nil {
		return domain.WorkItem{}, err
	}
	item.ReporterIDs = append([]string(nil), reporterIDs...)
	return item, nil
}

func (repository *SQLWorkItemRepository) Delete(id string) error {
	ctx := context.Background()
	if err := repository.queries.RemoveAllWorkItemReporters(ctx, id); err != nil {
		return err
	}
	return repository.queries.DeleteWorkItem(ctx, id)
}

func (repository *SQLWorkItemRepository) syncReporterIDs(ctx context.Context, workItemID string, reporterIDs []string) error {
	if err := repository.queries.RemoveAllWorkItemReporters(ctx, workItemID); err != nil {
		return err
	}
	for _, reporterID := range reporterIDs {
		reporterID = strings.TrimSpace(reporterID)
		if reporterID == "" {
			continue
		}
		if err := repository.queries.AddWorkItemReporter(ctx, sqlcgen.AddWorkItemReporterParams{
			ID:         uuid.NewString(),
			WorkItemID: workItemID,
			ReporterID: pgtype.Text{String: reporterID, Valid: true},
		}); err != nil {
			return err
		}
	}
	return nil
}

func workItemFromSQL(item sqlcgen.WorkItem) domain.WorkItem {
	features := map[string]any(nil)
	if len(item.Features) > 0 {
		if err := json.Unmarshal(item.Features, &features); err != nil {
			features = map[string]any{}
		}
	}
	if features == nil {
		features = map[string]any{}
	}
	return domain.WorkItem{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        domain.WorkItemType(item.Type),
		Status:      domain.WorkItemStatus(item.Status),
		Priority:    domain.WorkItemPriority(item.Priority),
		AssigneeID:  textValue(item.AssigneeID),
		StoryPoints: item.StoryPoints,
		Features:    features,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}
}

func marshalFeatures(features map[string]any) []byte {
	if features == nil {
		return []byte("{}")
	}
	payload, err := json.Marshal(features)
	if err != nil {
		return []byte("{}")
	}
	return payload
}

func reporterIDsFromSQL(reporters []pgtype.Text) []string {
	if len(reporters) == 0 {
		return nil
	}
	result := make([]string, 0, len(reporters))
	for _, reporter := range reporters {
		if reporter.Valid {
			result = append(result, reporter.String)
		}
	}
	return result
}

func pgtypeText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
