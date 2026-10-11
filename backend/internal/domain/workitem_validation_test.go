package domain_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestValidateWorkItemAttributes_IsDrivenByTemplateRules(t *testing.T) {
	template := domain.Template{
		WorkItemTypes: []domain.WorkItemType{domain.WorkItemTypeTask},
		ValidationRules: domain.ValidationRules{
			RequiredFields: []string{"description"},
			StoryPoints:    &domain.StoryPointRule{GreaterThan: 2},
			TypeRules: []domain.WorkItemTypeRule{{
				Type: domain.WorkItemTypeTask, DescriptionFormat: domain.DescriptionFormatUserStory, RequireAcceptanceCriteria: true,
			}},
		},
	}
	points := int32(2)

	got := domain.ValidateWorkItemAttributes(template, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeTask, Title: "", StoryPoints: &points,
	})

	want := domain.ValidationResult{Violations: []domain.Violation{
		{Field: "description", Message: "description is required"},
		{Field: "description", Message: `description must read "As a … I want … so that …"`},
		{Field: "acceptance_criteria", Message: "at least one acceptance criterion is required"},
		{Field: "story_points", Message: "story_points must be greater than 2"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestValidateWorkItemAttributes_TemplateWithoutRulesOnlyChecksType(t *testing.T) {
	template := domain.Template{WorkItemTypes: []domain.WorkItemType{domain.WorkItemTypeBug}}
	points := int32(-5)

	got := domain.ValidateWorkItemAttributes(template, domain.WorkItemAttributes{Type: domain.WorkItemTypeBug, StoryPoints: &points})

	if !got.Valid || len(got.Violations) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestCatalogTemplates_CarryValidationRules(t *testing.T) {
	for _, tpl := range domain.AvailableTemplates() {
		if len(tpl.ValidationRules.RequiredFields) == 0 || tpl.ValidationRules.TypeRules == nil {
			t.Fatalf("template %s has no validation rules: %+v", tpl.ID, tpl.ValidationRules)
		}
	}
}

func TestValidateWorkItemAttributes_InputCapsAtLimit_AreAccepted(t *testing.T) {
	template := domain.Template{WorkItemTypes: []domain.WorkItemType{domain.WorkItemTypeTask}}
	criteria := make([]string, domain.MaxAcceptanceCriteria)
	for i := range criteria {
		criteria[i] = strings.Repeat("é", domain.MaxAcceptanceCriterionRunes)
	}

	got := domain.ValidateWorkItemAttributes(template, domain.WorkItemAttributes{
		Type:               domain.WorkItemTypeTask,
		Title:              strings.Repeat("é", domain.MaxWorkItemTitleRunes),
		Description:        strings.Repeat("é", domain.MaxWorkItemDescriptionRunes),
		AcceptanceCriteria: criteria,
	})

	if !got.Valid || len(got.Violations) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestValidateWorkItemAttributes_InputCapsExceeded_AreViolationsInAnyTemplate(t *testing.T) {
	template := domain.Template{WorkItemTypes: []domain.WorkItemType{domain.WorkItemTypeTask}}

	got := domain.ValidateWorkItemAttributes(template, domain.WorkItemAttributes{
		Type:               domain.WorkItemTypeTask,
		Title:              strings.Repeat("a", domain.MaxWorkItemTitleRunes+1),
		Description:        strings.Repeat("a", domain.MaxWorkItemDescriptionRunes+1),
		AcceptanceCriteria: make([]string, domain.MaxAcceptanceCriteria+1),
	})

	want := domain.ValidationResult{Violations: []domain.Violation{
		{Field: "title", Message: "title must be at most 200 characters"},
		{Field: "description", Message: "description must be at most 10000 characters"},
		{Field: "acceptance_criteria", Message: "acceptance_criteria must have at most 50 items"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestValidateWorkItemAttributes_TooLongCriterion_IsOneViolation(t *testing.T) {
	template := domain.Template{WorkItemTypes: []domain.WorkItemType{domain.WorkItemTypeTask}}
	long := strings.Repeat("a", domain.MaxAcceptanceCriterionRunes+1)

	got := domain.ValidateWorkItemAttributes(template, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeTask, AcceptanceCriteria: []string{"ok", long, long},
	})

	want := domain.ValidationResult{Violations: []domain.Violation{
		{Field: "acceptance_criteria", Message: "each acceptance criterion must be at most 1000 characters"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestGeneralTemplate_DoesNotConstrainStoryPoints(t *testing.T) {
	general, _ := domain.TemplateByID(domain.TemplateIDGeneral)
	points := int32(0)

	got := domain.ValidateWorkItemAttributes(general, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeTask, Title: "Do it", StoryPoints: &points,
	})

	if general.ValidationRules.StoryPoints != nil || !got.Valid {
		t.Fatalf("rules %+v, result %+v", general.ValidationRules, got)
	}
}
