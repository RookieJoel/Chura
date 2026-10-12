package domain

import "time"

// SprintReflection is one team member's answers to the Reflection Template of
// their Project, bound to a Sprint. Each member has at most one per Sprint;
// submitting again replaces the answers.
type SprintReflection struct {
	ID          string `json:"id"`
	SprintID    string `json:"sprint_id"`
	ProjectID   string `json:"project_id"`
	AuthorID    string `json:"author_id"`
	AuthorEmail string `json:"author_email"`
	// TemplateID is the Template whose topics Answers follow.
	TemplateID string `json:"template_id"`
	// Answers maps a ReflectionTopic key (e.g. ReflectionWentWell) to the answer.
	Answers   map[string]string `json:"answers"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
