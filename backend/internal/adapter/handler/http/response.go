package http

import (
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type workItemResponse struct {
	RequestID  string             `json:"request_id"`
	Operation  string             `json:"operation"`
	WorkItem   *domain.WorkItem   `json:"work_item,omitempty"`
	WorkItems  []domain.WorkItem  `json:"work_items,omitempty"`
	Valid      *bool              `json:"valid,omitempty"`
	Violations []domain.Violation `json:"violations,omitempty"`
	Error      string             `json:"error,omitempty"`
}
