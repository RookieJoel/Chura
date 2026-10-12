package domain

import "time"

type SprintReflection struct {
	ID        string            `json:"id"`
	SprintID  string            `json:"sprint_id"`
	Author    string            `json:"author"`
	Answers   map[string]string `json:"answers"`
	CreatedAt time.Time         `json:"created_at"`
}

type SprintWorkSummary struct {
	Planned   int `json:"planned"`
	Completed int `json:"completed"`
	CarryOver int `json:"carry_over"`
	Blocked   int `json:"blocked"`
}

type QuantitativeSummary struct {
	Work SprintWorkSummary `json:"work"`
}

type QualitativeSummary struct {
	Reflections []SprintReflection `json:"reflections"`
}

type SprintReviewSummary struct {
	Sprint       Sprint              `json:"sprint"`
	Quantitative QuantitativeSummary `json:"quantitative"`
	Qualitative  QualitativeSummary  `json:"qualitative"`
	GeneratedAt  time.Time           `json:"generated_at"`
}
