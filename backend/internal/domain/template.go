package domain

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type Mode string

const (
	ModeGeneral Mode = "general"
	ModeSE      Mode = "se"
)

type Template struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Mode          Mode           `json:"mode"`
	WorkItemTypes []WorkItemType `json:"work_item_types"`
	Roles         []ProjectRole  `json:"roles"`
	DefaultRole   ProjectRole    `json:"default_role"`
	CreatorRole   ProjectRole    `json:"creator_role"`
	Capabilities  []string       `json:"capabilities"`
	// ValidationRules drive ValidateWorkItemAttributes for projects using this Template.
	ValidationRules ValidationRules `json:"validation_rules"`
	// Reflection lists the questions a Sprint Reflection must answer.
	Reflection ReflectionTemplate `json:"reflection"`
	// SprintReview lists the metrics a Sprint Review Summary may show.
	SprintReview SprintReviewTemplate `json:"sprint_review"`
}

const (
	TemplateIDGeneral = "general"
	TemplateIDSE      = "se"
)

func catalog() []Template {
	return []Template{
		{
			ID:            TemplateIDGeneral,
			Name:          "General",
			Mode:          ModeGeneral,
			WorkItemTypes: []WorkItemType{WorkItemTypeTask},
			Roles:         []ProjectRole{"owner", "member"},
			DefaultRole:   "member",
			CreatorRole:   "owner",
			Capabilities:  []string{},
			ValidationRules: ValidationRules{
				RequiredFields: []string{WorkItemFieldTitle},
				TypeRules:      []WorkItemTypeRule{},
			},
			Reflection:   StandardReflectionTemplate(),
			SprintReview: GeneralSprintReviewTemplate(),
		},
		{
			ID:            TemplateIDSE,
			Name:          "Software Engineering",
			Mode:          ModeSE,
			WorkItemTypes: []WorkItemType{WorkItemTypeUserStory, WorkItemTypeTask, WorkItemTypeBug},
			Roles:         []ProjectRole{"product_owner", "scrum_master", "developer"},
			DefaultRole:   "developer",
			CreatorRole:   "product_owner",
			Capabilities:  []string{"user_story_guidance", "acceptance_criteria", "story_points", "sprints", "se_templates"},
			ValidationRules: ValidationRules{
				RequiredFields: []string{WorkItemFieldTitle},
				StoryPoints:    &StoryPointRule{GreaterThan: 0},
				TypeRules: []WorkItemTypeRule{{
					Type:                      WorkItemTypeUserStory,
					DescriptionFormat:         DescriptionFormatUserStory,
					RequireAcceptanceCriteria: true,
				}},
			},
			Reflection:   StandardReflectionTemplate(),
			SprintReview: SESprintReviewTemplate(),
		},
	}
}

// AvailableTemplates returns a fresh copy of the catalog on every call.
func AvailableTemplates() []Template {
	return catalog()
}

// TemplateByID returns a fresh copy of the template with the given id.
func TemplateByID(id string) (Template, bool) {
	for _, t := range catalog() {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}

// ---- Reflection template: the qualitative part of a Sprint Review ----

// Reflection topic keys: the three retrospective questions.
const (
	ReflectionWentWell = "went_well" // What went well?
	ReflectionStop     = "stop"      // What should we stop doing?
	ReflectionStart    = "start"     // What should we start doing?
)

const MaxReflectionAnswerRunes = 2000

// ReflectionTopic is one question a team member answers in a Sprint Reflection.
type ReflectionTopic struct {
	Key      string `json:"key"`
	Question string `json:"question"`
	Hint     string `json:"hint"`
	Required bool   `json:"required"`
	MaxRunes int    `json:"max_runes"`
}

// ReflectionTemplate lists the topics of a Sprint Reflection, in display order.
type ReflectionTemplate struct {
	Topics []ReflectionTopic `json:"topics"`
}

// StandardReflectionTemplate is the went well / stop / start retrospective.
func StandardReflectionTemplate() ReflectionTemplate {
	return ReflectionTemplate{Topics: []ReflectionTopic{
		{Key: ReflectionWentWell, Question: "What went well?", Hint: "Practices, tools or habits that helped the team this sprint.", Required: true, MaxRunes: MaxReflectionAnswerRunes},
		{Key: ReflectionStop, Question: "What should we stop doing?", Hint: "Things that slowed the team down or caused problems.", Required: true, MaxRunes: MaxReflectionAnswerRunes},
		{Key: ReflectionStart, Question: "What should we start doing?", Hint: "New ideas to try in the next sprint.", Required: true, MaxRunes: MaxReflectionAnswerRunes},
	}}
}

// Topic returns the topic with the given key.
func (t ReflectionTemplate) Topic(key string) (ReflectionTopic, bool) {
	for _, topic := range t.Topics {
		if topic.Key == key {
			return topic, true
		}
	}
	return ReflectionTopic{}, false
}

// ValidateAnswers returns the trimmed answers when they fit the template:
// every required topic answered, no unknown topic, no answer too long.
// Optional topics left blank are dropped. Every problem is reported as an
// *InvalidInputError violation on "answers.<key>".
func (t ReflectionTemplate) ValidateAnswers(answers map[string]string) (map[string]string, error) {
	var unknown []string
	for key := range answers {
		if _, ok := t.Topic(key); !ok {
			unknown = append(unknown, key)
		}
	}
	slices.Sort(unknown)
	var violations []Violation
	for _, key := range unknown {
		violations = append(violations, Violation{Field: "answers." + key, Message: "unknown reflection topic"})
	}
	clean := make(map[string]string, len(t.Topics))
	for _, topic := range t.Topics {
		answer := strings.TrimSpace(answers[topic.Key])
		field := "answers." + topic.Key
		switch {
		case answer == "" && topic.Required:
			violations = append(violations, Violation{Field: field, Message: fmt.Sprintf("%q is required", topic.Question)})
		case topic.MaxRunes > 0 && utf8.RuneCountInString(answer) > topic.MaxRunes:
			violations = append(violations, Violation{Field: field, Message: fmt.Sprintf("answer must be at most %d characters", topic.MaxRunes)})
		case answer != "":
			clean[topic.Key] = answer
		}
	}
	if len(violations) > 0 {
		return nil, &InvalidInputError{Violations: violations}
	}
	return clean, nil
}

// ---- Sprint Review template: the quantitative part of a Sprint Review ----

type MetricKey string

const (
	MetricPlannedItems   MetricKey = "planned_items"
	MetricCompletedItems MetricKey = "completed_items"
	MetricCarryOverItems MetricKey = "carry_over_items"
	MetricBlockedItems   MetricKey = "blocked_items"
	MetricCompletionRate MetricKey = "completion_rate"

	MetricPlannedPoints        MetricKey = "planned_points"
	MetricCompletedPoints      MetricKey = "completed_points"
	MetricCarryOverPoints      MetricKey = "carry_over_points"
	MetricBlockedPoints        MetricKey = "blocked_points"
	MetricPointsCompletionRate MetricKey = "points_completion_rate"
)

type MetricUnit string

const (
	MetricUnitItems   MetricUnit = "items"
	MetricUnitPoints  MetricUnit = "points"
	MetricUnitPercent MetricUnit = "percent"
)

// MetricDefinition describes one metric a Sprint Review Summary can show.
type MetricDefinition struct {
	Key   MetricKey  `json:"key"`
	Label string     `json:"label"`
	Unit  MetricUnit `json:"unit"`
}

// SprintReviewTemplate is the standard layout of a Sprint Review Summary:
// the metrics a team may pick from (in report order) and the ones shown when
// the team picks none.
type SprintReviewTemplate struct {
	Title          string             `json:"title"`
	Metrics        []MetricDefinition `json:"metrics"`
	DefaultMetrics []MetricKey        `json:"default_metrics"`
}

var metricLabels = map[MetricKey]MetricDefinition{
	MetricPlannedItems:         {Label: "Planned work items", Unit: MetricUnitItems},
	MetricCompletedItems:       {Label: "Completed work items", Unit: MetricUnitItems},
	MetricCarryOverItems:       {Label: "Carry-over work items", Unit: MetricUnitItems},
	MetricBlockedItems:         {Label: "Blocked work items", Unit: MetricUnitItems},
	MetricCompletionRate:       {Label: "Completion rate", Unit: MetricUnitPercent},
	MetricPlannedPoints:        {Label: "Planned story points", Unit: MetricUnitPoints},
	MetricCompletedPoints:      {Label: "Completed story points (velocity)", Unit: MetricUnitPoints},
	MetricCarryOverPoints:      {Label: "Carry-over story points", Unit: MetricUnitPoints},
	MetricBlockedPoints:        {Label: "Blocked story points", Unit: MetricUnitPoints},
	MetricPointsCompletionRate: {Label: "Story point completion rate", Unit: MetricUnitPercent},
}

func metricDefinitions(keys ...MetricKey) []MetricDefinition {
	definitions := make([]MetricDefinition, 0, len(keys))
	for _, key := range keys {
		d := metricLabels[key]
		d.Key = key
		definitions = append(definitions, d)
	}
	return definitions
}

// GeneralSprintReviewTemplate counts work items only: General projects have
// no story points.
func GeneralSprintReviewTemplate() SprintReviewTemplate {
	return SprintReviewTemplate{
		Title:          "Sprint Review Summary",
		Metrics:        metricDefinitions(MetricPlannedItems, MetricCompletedItems, MetricCarryOverItems, MetricBlockedItems, MetricCompletionRate),
		DefaultMetrics: []MetricKey{MetricPlannedItems, MetricCompletedItems, MetricCarryOverItems, MetricBlockedItems, MetricCompletionRate},
	}
}

// SESprintReviewTemplate adds story-point metrics to the work item counts.
func SESprintReviewTemplate() SprintReviewTemplate {
	return SprintReviewTemplate{
		Title: "Sprint Review Summary",
		Metrics: metricDefinitions(
			MetricPlannedItems, MetricCompletedItems, MetricCarryOverItems, MetricBlockedItems, MetricCompletionRate,
			MetricPlannedPoints, MetricCompletedPoints, MetricCarryOverPoints, MetricBlockedPoints, MetricPointsCompletionRate,
		),
		DefaultMetrics: []MetricKey{
			MetricPlannedItems, MetricCompletedItems, MetricCarryOverItems, MetricBlockedItems,
			MetricPlannedPoints, MetricCompletedPoints, MetricPointsCompletionRate,
		},
	}
}

// Metric returns the definition of the metric with the given key.
func (t SprintReviewTemplate) Metric(key MetricKey) (MetricDefinition, bool) {
	for _, m := range t.Metrics {
		if m.Key == key {
			return m, true
		}
	}
	return MetricDefinition{}, false
}

// SelectMetrics returns the definitions of the requested metrics in template
// order, without duplicates. No request means the template defaults. A metric
// the template does not offer is an *InvalidInputError on "metrics".
func (t SprintReviewTemplate) SelectMetrics(requested []MetricKey) ([]MetricDefinition, error) {
	if len(requested) == 0 {
		requested = t.DefaultMetrics
	}
	wanted := make(map[MetricKey]bool, len(requested))
	var violations []Violation
	for _, key := range requested {
		if _, ok := t.Metric(key); !ok {
			violations = append(violations, Violation{Field: "metrics", Message: fmt.Sprintf("metric %q is not available in this template", key)})
			continue
		}
		wanted[key] = true
	}
	if len(violations) > 0 {
		return nil, &InvalidInputError{Violations: violations}
	}
	selected := make([]MetricDefinition, 0, len(wanted))
	for _, m := range t.Metrics {
		if wanted[m.Key] {
			selected = append(selected, m)
		}
	}
	return selected, nil
}
