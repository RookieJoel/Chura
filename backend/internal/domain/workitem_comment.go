package domain

import "time"

// WorkItemComment is one append-only comment on a Work Item.
type WorkItemComment struct {
	ID           string    `json:"id"`
	WorkItemID   string    `json:"work_item_id"`
	AuthorUserID string    `json:"author_user_id"`
	AuthorName   string    `json:"author_name"`
	AuthorEmail  string    `json:"author_email"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}
