package domain

import (
	"math"
	"time"
)

// WorkCategory is where a Sprint work item ends up in the Sprint Work Summary.
type WorkCategory string

const (
	WorkCategoryCompleted WorkCategory = "completed"
	WorkCategoryCarryOver WorkCategory = "carry_over"
	WorkCategoryBlocked   WorkCategory = "blocked"
)

// CategorizeWorkItem puts done items in Completed, blocked items in Blocked
// and every other unfinished item in Carry-over.
func CategorizeWorkItem(status WorkItemStatus) WorkCategory {
	switch status {
	case WorkItemStatusDone:
		return WorkCategoryCompleted
	case WorkItemStatusBlocked:
		return WorkCategoryBlocked
	default:
		return WorkCategoryCarryOver
	}
}

// WorkBucket counts work items and their story points.
type WorkBucket struct {
	Items  int `json:"items"`
	Points int `json:"points"`
}

func (b *WorkBucket) add(item WorkItem) {
	b.Items++
	b.Points += int(item.StoryPoints)
}

// SprintWorkSummary is the result of calculateSprintWorkSummary(). Planned is
// every work item in the Sprint and equals Completed + CarryOver + Blocked.
// While a Sprint is still running, CarryOver is the work that would carry over
// if it ended now.
type SprintWorkSummary struct {
	Planned   WorkBucket `json:"planned"`
	Completed WorkBucket `json:"completed"`
	CarryOver WorkBucket `json:"carry_over"`
	Blocked   WorkBucket `json:"blocked"`
}

// CalculateSprintWorkSummary aggregates the Sprint's work items into planned,
// completed, carry-over and blocked work.
func CalculateSprintWorkSummary(items []WorkItem) SprintWorkSummary {
	var s SprintWorkSummary
	for _, item := range items {
		s.Planned.add(item)
		switch CategorizeWorkItem(item.Status) {
		case WorkCategoryCompleted:
			s.Completed.add(item)
		case WorkCategoryBlocked:
			s.Blocked.add(item)
		default:
			s.CarryOver.add(item)
		}
	}
	return s
}

// CompletionRate is the percentage of planned work items that are done.
func (s SprintWorkSummary) CompletionRate() float64 {
	return percent(s.Completed.Items, s.Planned.Items)
}

// PointsCompletionRate is the percentage of planned story points that are done.
func (s SprintWorkSummary) PointsCompletionRate() float64 {
	return percent(s.Completed.Points, s.Planned.Points)
}

// MetricValue returns the value of a Sprint Review metric.
func (s SprintWorkSummary) MetricValue(key MetricKey) (float64, bool) {
	switch key {
	case MetricPlannedItems:
		return float64(s.Planned.Items), true
	case MetricCompletedItems:
		return float64(s.Completed.Items), true
	case MetricCarryOverItems:
		return float64(s.CarryOver.Items), true
	case MetricBlockedItems:
		return float64(s.Blocked.Items), true
	case MetricCompletionRate:
		return s.CompletionRate(), true
	case MetricPlannedPoints:
		return float64(s.Planned.Points), true
	case MetricCompletedPoints:
		return float64(s.Completed.Points), true
	case MetricCarryOverPoints:
		return float64(s.CarryOver.Points), true
	case MetricBlockedPoints:
		return float64(s.Blocked.Points), true
	case MetricPointsCompletionRate:
		return s.PointsCompletionRate(), true
	}
	return 0, false
}

// percent returns part/whole as a percentage rounded to one decimal; 0 when
// there is nothing to divide by.
func percent(part, whole int) float64 {
	if whole <= 0 {
		return 0
	}
	return math.Round(float64(part)*1000/float64(whole)) / 10
}

// SprintProgress compares the Sprint calendar with today. Days are counted
// inclusively, so a Sprint from Monday to Friday has 5 TotalDays.
type SprintProgress struct {
	TotalDays       int     `json:"total_days"`
	ElapsedDays     int     `json:"elapsed_days"`
	RemainingDays   int     `json:"remaining_days"`
	TimeElapsedRate float64 `json:"time_elapsed_rate"`
}

// CalculateSprintProgress returns nil when the Sprint has no start or end date.
func CalculateSprintProgress(sprint Sprint, now time.Time) *SprintProgress {
	if sprint.StartDate == nil || sprint.EndDate == nil {
		return nil
	}
	start, end, today := dateOnly(*sprint.StartDate), dateOnly(*sprint.EndDate), dateOnly(now)
	total := daysBetween(start, end) + 1
	elapsed := min(max(daysBetween(start, today)+1, 0), total)
	return &SprintProgress{
		TotalDays:       total,
		ElapsedDays:     elapsed,
		RemainingDays:   total - elapsed,
		TimeElapsedRate: percent(elapsed, total),
	}
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(math.Round(to.Sub(from).Hours() / 24))
}

// SprintWorkItem is the view of one work item inside the Sprint Metrics.
type SprintWorkItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Type        WorkItemType     `json:"type"`
	Status      WorkItemStatus   `json:"status"`
	Priority    WorkItemPriority `json:"priority"`
	AssigneeID  string           `json:"assignee_id,omitempty"`
	StoryPoints int32            `json:"story_points"`
	Category    WorkCategory     `json:"category"`
}

// SprintMetrics is the result of retrieveSprintMetrics(): performance data,
// work items and progress stats of one Sprint.
type SprintMetrics struct {
	Sprint               Sprint                 `json:"sprint"`
	WorkSummary          SprintWorkSummary      `json:"work_summary"`
	CompletionRate       float64                `json:"completion_rate"`
	PointsCompletionRate float64                `json:"points_completion_rate"`
	StatusBreakdown      map[WorkItemStatus]int `json:"status_breakdown"`
	TypeBreakdown        map[WorkItemType]int   `json:"type_breakdown"`
	Progress             *SprintProgress        `json:"progress"`
	WorkItems            []SprintWorkItem       `json:"work_items"`
	CalculatedAt         time.Time              `json:"calculated_at"`
}

// CalculateSprintMetrics derives every Sprint metric from the Sprint and its work items.
func CalculateSprintMetrics(sprint Sprint, items []WorkItem, now time.Time) SprintMetrics {
	summary := CalculateSprintWorkSummary(items)
	metrics := SprintMetrics{
		Sprint:               sprint,
		WorkSummary:          summary,
		CompletionRate:       summary.CompletionRate(),
		PointsCompletionRate: summary.PointsCompletionRate(),
		StatusBreakdown:      map[WorkItemStatus]int{},
		TypeBreakdown:        map[WorkItemType]int{},
		Progress:             CalculateSprintProgress(sprint, now),
		WorkItems:            make([]SprintWorkItem, 0, len(items)),
		CalculatedAt:         now,
	}
	for _, item := range items {
		metrics.StatusBreakdown[item.Status]++
		metrics.TypeBreakdown[item.Type]++
		metrics.WorkItems = append(metrics.WorkItems, SprintWorkItem{
			ID: item.ID, Title: item.Title, Type: item.Type, Status: item.Status, Priority: item.Priority,
			AssigneeID: item.AssigneeID, StoryPoints: item.StoryPoints, Category: CategorizeWorkItem(item.Status),
		})
	}
	return metrics
}
