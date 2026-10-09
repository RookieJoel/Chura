package service_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

const validStory = "As a developer I want a board so that I can see work"

func ptr(v int32) *int32 { return &v }

// validateIn seeds a project from templateID owned by teamMember (u1) and runs
// validation as actor against it.
func validateIn(t *testing.T, templateID string, actor domain.Actor, attrs domain.WorkItemAttributes) (*fakeProjectRepo, domain.ValidationResult, error) {
	t.Helper()
	repo := newFakeProjectRepo()
	svc := newProjectService(repo)
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: templateID}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo.writes = 0
	got, err := svc.ValidateWorkItemAttributes(context.Background(), actor, seededProjectID, attrs)
	return repo, got, err
}

func assertResult(t *testing.T, got domain.ValidationResult, err error, want domain.ValidationResult) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

var noViolations = []domain.Violation{}

func TestValidateWorkItemAttributes_SEValidUserStory_IsValid(t *testing.T) {
	_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeUserStory, Title: "Board", Description: validStory,
		StoryPoints: ptr(5), AcceptanceCriteria: []string{"board is shown"},
	})

	assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
}

func TestValidateWorkItemAttributes_TypeNotInTemplate_IsRejected(t *testing.T) {
	_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: "epic", Title: "Big"})

	assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
		{Field: "type", Message: "type is not allowed in this project"},
	}})
}

func TestValidateWorkItemAttributes_BlankTitle_IsRejected(t *testing.T) {
	_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: " \t "})

	assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
		{Field: "title", Message: "title is required"},
	}})
}

func TestValidateWorkItemAttributes_NonPositiveStoryPoints_AreRejected(t *testing.T) {
	for _, points := range []int32{0, -1} {
		_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
			Type: domain.WorkItemTypeTask, Title: "Do it", StoryPoints: ptr(points),
		})

		assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
			{Field: "story_points", Message: "story_points must be greater than 0"},
		}})
	}
}

const storyFormatMessage = `description must read "As a … I want … so that …"`

func TestValidateWorkItemAttributes_SEUserStoryMalformedDescription_IsRejected(t *testing.T) {
	for name, description := range map[string]string{
		"missing so that":       "As a developer I want a board",
		"missing I want":        "As a developer so that I can see work",
		"missing As a":          "I want a board so that I can see work",
		"empty":                 "",
		"nothing after so that": "As a developer I want a board so that  ",
	} {
		t.Run(name, func(t *testing.T) {
			_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
				Type: domain.WorkItemTypeUserStory, Title: "Board", Description: description,
				AcceptanceCriteria: []string{"board is shown"},
			})

			assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
				{Field: "description", Message: storyFormatMessage},
			}})
		})
	}
}

func TestValidateWorkItemAttributes_SEUserStoryCaseAndNewlines_AreAccepted(t *testing.T) {
	_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeUserStory, Title: "Board", Description: "AS A dev\nI WANT x\nSO THAT y",
		AcceptanceCriteria: []string{"ok"},
	})

	assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
}

func TestValidateWorkItemAttributes_SEUserStoryWithoutAcceptanceCriteria_IsRejected(t *testing.T) {
	for name, criteria := range map[string][]string{
		"nil":        nil,
		"only blank": {"", "  \t"},
	} {
		t.Run(name, func(t *testing.T) {
			_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
				Type: domain.WorkItemTypeUserStory, Title: "Board", Description: validStory, AcceptanceCriteria: criteria,
			})

			assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
				{Field: "acceptance_criteria", Message: "at least one acceptance criterion is required"},
			}})
		})
	}
}

func TestValidateWorkItemAttributes_SETaskAndBug_DoNotRequireStoryStructureOrCriteria(t *testing.T) {
	for _, typ := range []domain.WorkItemType{domain.WorkItemTypeTask, domain.WorkItemTypeBug} {
		_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: typ, Title: "Fix it", Description: "just do it"})

		assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
	}
}

func TestValidateWorkItemAttributes_General_AcceptsTaskOnly(t *testing.T) {
	_, got, err := validateIn(t, "general", teamMember, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it"})
	assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})

	for _, typ := range []domain.WorkItemType{domain.WorkItemTypeUserStory, domain.WorkItemTypeBug} {
		_, got, err := validateIn(t, "general", teamMember, domain.WorkItemAttributes{Type: typ, Title: "Do it"})

		assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
			{Field: "type", Message: "type is not allowed in this project"},
		}})
	}
}

func TestValidateWorkItemAttributes_General_DoesNotConstrainStoryPoints(t *testing.T) {
	for _, points := range []*int32{nil, ptr(0), ptr(-1), ptr(5)} {
		_, got, err := validateIn(t, "general", teamMember, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it", StoryPoints: points})

		assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
	}
}

func TestValidateWorkItemAttributes_InputCapsExceeded_AreViolations(t *testing.T) {
	for _, templateID := range []string{"general", "se"} {
		_, got, err := validateIn(t, templateID, teamMember, domain.WorkItemAttributes{
			Type: domain.WorkItemTypeTask, Title: strings.Repeat("a", domain.MaxWorkItemTitleRunes+1),
		})

		assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
			{Field: "title", Message: "title must be at most 200 characters"},
		}})
	}
}

func TestValidateWorkItemAttributes_StoryPointsNilOrPositive_AreAccepted(t *testing.T) {
	for _, points := range []*int32{nil, ptr(1), ptr(5)} {
		_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it", StoryPoints: points})

		assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
	}
}

func TestValidateWorkItemAttributes_MultipleViolations_AreInStableFieldOrder(t *testing.T) {
	_, got, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{
		Type: domain.WorkItemTypeUserStory, Title: " ", Description: "make it nice", StoryPoints: ptr(0),
	})

	assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
		{Field: "title", Message: "title is required"},
		{Field: "description", Message: storyFormatMessage},
		{Field: "acceptance_criteria", Message: "at least one acceptance criterion is required"},
		{Field: "story_points", Message: "story_points must be greater than 0"},
	}})

	_, got, err = validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: "epic", Title: "", StoryPoints: ptr(-3)})

	assertResult(t, got, err, domain.ValidationResult{Violations: []domain.Violation{
		{Field: "type", Message: "type is not allowed in this project"},
		{Field: "title", Message: "title is required"},
		{Field: "story_points", Message: "story_points must be greater than 0"},
	}})
}

func TestValidateWorkItemAttributes_AsAuditor_IsAllowed(t *testing.T) {
	_, got, err := validateIn(t, "se", domain.Actor{UserID: "a1", Role: domain.SystemRoleAuditor}, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it"})

	assertResult(t, got, err, domain.ValidationResult{Valid: true, Violations: noViolations})
}

func TestValidateWorkItemAttributes_AsNonMember_IsNotFound(t *testing.T) {
	_, _, err := validateIn(t, "se", domain.Actor{UserID: "u2", Role: domain.SystemRoleTeamMember}, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it"})

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestValidateWorkItemAttributes_UnauthenticatedActor_IsRejected(t *testing.T) {
	_, _, err := validateIn(t, "se", domain.Actor{UserID: "u1", Role: "admin"}, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it"})

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

func TestValidateWorkItemAttributes_MalformedProjectID_IsInvalidID(t *testing.T) {
	svc := newProjectService(newFakeProjectRepo())

	_, err := svc.ValidateWorkItemAttributes(context.Background(), teamMember, "not-a-uuid", domain.WorkItemAttributes{})

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
}

func TestValidateWorkItemAttributes_PersistsNothing(t *testing.T) {
	repo, _, err := validateIn(t, "se", teamMember, domain.WorkItemAttributes{Type: domain.WorkItemTypeTask, Title: "Do it"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.writes != 0 {
		t.Fatalf("want zero repository writes, got %d", repo.writes)
	}
}
