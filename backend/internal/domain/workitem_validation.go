package domain

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type WorkItemAttributes struct {
	Type               WorkItemType `json:"type"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	StoryPoints        *int32       `json:"story_points,omitempty"`
	AcceptanceCriteria []string     `json:"acceptance_criteria,omitempty"`
}

type ValidationResult struct {
	Valid      bool        `json:"valid"`
	Violations []Violation `json:"violations"`
}

// Work Item text fields that ValidationRules.RequiredFields may name.
const (
	WorkItemFieldTitle       = "title"
	WorkItemFieldDescription = "description"
)

// DescriptionFormatUserStory requires "As a … I want … so that …".
const DescriptionFormatUserStory = "user_story"

// ValidationRules is the data-driven part of a Template that decides which
// Work Item attributes are acceptable.
type ValidationRules struct {
	// RequiredFields lists text fields that must not be blank (WorkItemField* values).
	RequiredFields []string `json:"required_fields"`
	// StoryPoints is nil when story_points are unconstrained.
	StoryPoints *StoryPointRule    `json:"story_points,omitempty"`
	TypeRules   []WorkItemTypeRule `json:"type_rules"`
}

// StoryPointRule constrains story_points when present: value > GreaterThan.
type StoryPointRule struct {
	GreaterThan int32 `json:"greater_than"`
}

// WorkItemTypeRule adds requirements for one Work Item type.
type WorkItemTypeRule struct {
	Type                      WorkItemType `json:"type"`
	DescriptionFormat         string       `json:"description_format,omitempty"`
	RequireAcceptanceCriteria bool         `json:"require_acceptance_criteria"`
}

var descriptionFormats = map[string]struct {
	pattern *regexp.Regexp
	message string
}{
	DescriptionFormatUserStory: {
		pattern: regexp.MustCompile(`(?is)^\s*as an?\s+\S.*\bi want\b.*\bso that\b\s*\S.*$`),
		message: `description must read "As a … I want … so that …"`,
	},
}

// ValidateWorkItemAttributes checks draft Work Item attributes against a
// Template's types and ValidationRules. It is pure: nothing is persisted.
// Violations are reported in field order: type, required fields, input caps,
// per-type rules, story_points. Input caps apply in every Template.
func ValidateWorkItemAttributes(t Template, a WorkItemAttributes) ValidationResult {
	violations := []Violation{}
	if !slices.Contains(t.WorkItemTypes, a.Type) {
		violations = append(violations, Violation{Field: "type", Message: "type is not allowed in this project"})
	}
	rules := t.ValidationRules
	for _, field := range rules.RequiredFields {
		if value, known := textField(a, field); known && strings.TrimSpace(value) == "" {
			violations = append(violations, Violation{Field: field, Message: field + " is required"})
		}
	}
	violations = append(violations, capViolations(a)...)
	for _, rule := range rules.TypeRules {
		if rule.Type == a.Type {
			violations = append(violations, typeRuleViolations(rule, a)...)
		}
	}
	if a.StoryPoints != nil && rules.StoryPoints != nil && *a.StoryPoints <= rules.StoryPoints.GreaterThan {
		violations = append(violations, Violation{Field: "story_points", Message: fmt.Sprintf("story_points must be greater than %d", rules.StoryPoints.GreaterThan)})
	}
	return ValidationResult{Valid: len(violations) == 0, Violations: violations}
}

// capViolations enforces the Max* input caps on free-text attributes.
func capViolations(a WorkItemAttributes) []Violation {
	var violations []Violation
	if utf8.RuneCountInString(a.Title) > MaxWorkItemTitleRunes {
		violations = append(violations, Violation{Field: "title", Message: fmt.Sprintf("title must be at most %d characters", MaxWorkItemTitleRunes)})
	}
	if utf8.RuneCountInString(a.Description) > MaxWorkItemDescriptionRunes {
		violations = append(violations, Violation{Field: "description", Message: fmt.Sprintf("description must be at most %d characters", MaxWorkItemDescriptionRunes)})
	}
	if len(a.AcceptanceCriteria) > MaxAcceptanceCriteria {
		violations = append(violations, Violation{Field: "acceptance_criteria", Message: fmt.Sprintf("acceptance_criteria must have at most %d items", MaxAcceptanceCriteria)})
	}
	if slices.ContainsFunc(a.AcceptanceCriteria, func(c string) bool { return utf8.RuneCountInString(c) > MaxAcceptanceCriterionRunes }) {
		violations = append(violations, Violation{Field: "acceptance_criteria", Message: fmt.Sprintf("each acceptance criterion must be at most %d characters", MaxAcceptanceCriterionRunes)})
	}
	return violations
}

func typeRuleViolations(rule WorkItemTypeRule, a WorkItemAttributes) []Violation {
	var violations []Violation
	if format, ok := descriptionFormats[rule.DescriptionFormat]; ok && !format.pattern.MatchString(a.Description) {
		violations = append(violations, Violation{Field: "description", Message: format.message})
	}
	if rule.RequireAcceptanceCriteria && !hasNonBlank(a.AcceptanceCriteria) {
		violations = append(violations, Violation{Field: "acceptance_criteria", Message: "at least one acceptance criterion is required"})
	}
	return violations
}

func textField(a WorkItemAttributes, field string) (string, bool) {
	switch field {
	case WorkItemFieldTitle:
		return a.Title, true
	case WorkItemFieldDescription:
		return a.Description, true
	}
	return "", false
}

func hasNonBlank(items []string) bool {
	return slices.ContainsFunc(items, func(s string) bool { return strings.TrimSpace(s) != "" })
}
