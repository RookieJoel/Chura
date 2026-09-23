package http

import (
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type workItemResponse struct {
	Operation string            `json:"operation"`
	WorkItem  *domain.WorkItem  `json:"work_item,omitempty"`
	WorkItems []domain.WorkItem `json:"work_items,omitempty"`
	Error     string            `json:"error,omitempty"`
}
