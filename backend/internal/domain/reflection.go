package domain

import "time"

type SprintReflection struct {
	ID        string    `json:"id"`
	SprintID  string    `json:"sprint_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type SprintReviewSummary struct {
	Sprint          Sprint             `json:"sprint"`
	Reflections     []SprintReflection `json:"reflections"`
	ReflectionCount int                `json:"reflection_count"`
	GeneratedAt     time.Time          `json:"generated_at"`
}
