package domain

import (
	"fmt"
	"strings"
	"time"
)

// SprintReviewSummary is the standardized report of generateSprintReviewSummary():
// quantitative metrics from the Sprint's work items next to the team's
// qualitative feedback, both laid out by the Project's Template.
type SprintReviewSummary struct {
	Title        string              `json:"title"`
	Project      ReviewProject       `json:"project"`
	Sprint       Sprint              `json:"sprint"`
	TemplateID   string              `json:"template_id"`
	Quantitative QuantitativeSummary `json:"quantitative"`
	Qualitative  QualitativeSummary  `json:"qualitative"`
	GeneratedAt  time.Time           `json:"generated_at"`
}

type ReviewProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// QuantitativeSummary holds the selected metrics plus the full work summary
// and progress they are computed from.
type QuantitativeSummary struct {
	Metrics     []MetricValue     `json:"metrics"`
	WorkSummary SprintWorkSummary `json:"work_summary"`
	Progress    *SprintProgress   `json:"progress"`
}

type MetricValue struct {
	Key   MetricKey  `json:"key"`
	Label string     `json:"label"`
	Unit  MetricUnit `json:"unit"`
	Value float64    `json:"value"`
}

// QualitativeSummary groups every reflection answer under its topic.
type QualitativeSummary struct {
	Topics []TopicFeedback `json:"topics"`
	// Respondents is how many members submitted a reflection, out of TeamSize.
	Respondents  int     `json:"respondents"`
	TeamSize     int     `json:"team_size"`
	ResponseRate float64 `json:"response_rate"`
}

type TopicFeedback struct {
	Key       string          `json:"key"`
	Question  string          `json:"question"`
	Responses []TopicResponse `json:"responses"`
}

type TopicResponse struct {
	AuthorID    string `json:"author_id"`
	AuthorEmail string `json:"author_email"`
	Answer      string `json:"answer"`
}

// ReviewInput is everything BuildSprintReviewSummary needs.
type ReviewInput struct {
	Project     Project
	Sprint      Sprint
	Template    Template
	Metrics     []MetricDefinition // already selected from Template.SprintReview
	WorkItems   []WorkItem
	Reflections []SprintReflection
	Now         time.Time
}

// BuildSprintReviewSummary consolidates quantitative metrics and qualitative
// feedback into the standardized Sprint Review Summary.
func BuildSprintReviewSummary(in ReviewInput) SprintReviewSummary {
	work := CalculateSprintWorkSummary(in.WorkItems)
	metrics := make([]MetricValue, 0, len(in.Metrics))
	for _, m := range in.Metrics {
		value, _ := work.MetricValue(m.Key)
		metrics = append(metrics, MetricValue{Key: m.Key, Label: m.Label, Unit: m.Unit, Value: value})
	}

	topics := make([]TopicFeedback, 0, len(in.Template.Reflection.Topics))
	for _, topic := range in.Template.Reflection.Topics {
		feedback := TopicFeedback{Key: topic.Key, Question: topic.Question, Responses: []TopicResponse{}}
		for _, r := range in.Reflections {
			if answer := r.Answers[topic.Key]; answer != "" {
				feedback.Responses = append(feedback.Responses, TopicResponse{AuthorID: r.AuthorID, AuthorEmail: r.AuthorEmail, Answer: answer})
			}
		}
		topics = append(topics, feedback)
	}
	teamSize := max(len(in.Project.Members), len(in.Reflections))

	return SprintReviewSummary{
		Title:      in.Template.SprintReview.Title,
		Project:    ReviewProject{ID: in.Project.ID, Name: in.Project.Name},
		Sprint:     in.Sprint,
		TemplateID: in.Template.ID,
		Quantitative: QuantitativeSummary{
			Metrics:     metrics,
			WorkSummary: work,
			Progress:    CalculateSprintProgress(in.Sprint, in.Now),
		},
		Qualitative: QualitativeSummary{
			Topics:       topics,
			Respondents:  len(in.Reflections),
			TeamSize:     teamSize,
			ResponseRate: percent(len(in.Reflections), teamSize),
		},
		GeneratedAt: in.Now,
	}
}

// ReportFormat is a file format exportSummaryReport() can produce.
type ReportFormat string

const (
	ReportFormatCSV      ReportFormat = "csv"
	ReportFormatMarkdown ReportFormat = "markdown"
)

// ParseReportFormat defaults an empty value to CSV; "md" means Markdown.
func ParseReportFormat(raw string) (ReportFormat, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(ReportFormatCSV):
		return ReportFormatCSV, nil
	case string(ReportFormatMarkdown), "md":
		return ReportFormatMarkdown, nil
	}
	return "", &InvalidInputError{Violations: []Violation{{
		Field: "format", Message: fmt.Sprintf("format must be one of %s, %s", ReportFormatCSV, ReportFormatMarkdown),
	}}}
}

// ExportedReport is a rendered Sprint Review Summary ready for download.
type ExportedReport struct {
	Filename    string
	ContentType string
	Data        []byte
}
