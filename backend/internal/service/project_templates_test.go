package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var teamMember = domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}

func TestGetAvailableTemplates_ExposesGeneralAndSE(t *testing.T) {
	svc := service.NewProjectConfigurationService(newFakeProjectRepo())

	got, err := svc.GetAvailableTemplates(context.Background(), teamMember)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []domain.Template{
		{
			ID: "general", Name: "General", Mode: "general",
			WorkItemTypes: []domain.WorkItemType{"task"},
			Roles:         []domain.ProjectRole{"owner", "member"},
			DefaultRole:   "member", CreatorRole: "owner",
			Capabilities: []string{},
			ValidationRules: domain.ValidationRules{
				RequiredFields: []string{"title"},
				TypeRules:      []domain.WorkItemTypeRule{},
			},
		},
		{
			ID: "se", Name: "Software Engineering", Mode: "se",
			WorkItemTypes: []domain.WorkItemType{"user_story", "task", "bug"},
			Roles:         []domain.ProjectRole{"product_owner", "scrum_master", "developer"},
			DefaultRole:   "developer", CreatorRole: "product_owner",
			Capabilities: []string{"user_story_guidance", "acceptance_criteria", "story_points", "sprints", "se_templates"},
			ValidationRules: domain.ValidationRules{
				RequiredFields: []string{"title"},
				StoryPoints:    &domain.StoryPointRule{GreaterThan: 0},
				TypeRules: []domain.WorkItemTypeRule{{
					Type: "user_story", DescriptionFormat: domain.DescriptionFormatUserStory, RequireAcceptanceCriteria: true,
				}},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("templates mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestRetrieveTemplateDefinition_ReturnsSE(t *testing.T) {
	svc := service.NewProjectConfigurationService(newFakeProjectRepo())

	got, err := svc.RetrieveTemplateDefinition(context.Background(), teamMember, "se")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "se" || got.Mode != "se" || got.DefaultRole != "developer" || got.CreatorRole != "product_owner" {
		t.Fatalf("unexpected SE template: %+v", got)
	}
}

func TestRetrieveTemplateDefinition_UnknownIDIsNotFound(t *testing.T) {
	svc := service.NewProjectConfigurationService(newFakeProjectRepo())

	_, err := svc.RetrieveTemplateDefinition(context.Background(), teamMember, "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestTemplates_UnauthenticatedActorRejected(t *testing.T) {
	svc := service.NewProjectConfigurationService(newFakeProjectRepo())
	bad := []domain.Actor{
		{UserID: "", Role: domain.SystemRoleTeamMember},
		{UserID: "   ", Role: domain.SystemRoleAuditor},
		{UserID: "u1", Email: "u1@example.com", Role: "admin"},
	}
	for _, actor := range bad {
		if _, err := svc.GetAvailableTemplates(context.Background(), actor); !errors.Is(err, domain.ErrUnauthenticated) {
			t.Errorf("list %+v: want ErrUnauthenticated, got %v", actor, err)
		}
		if _, err := svc.RetrieveTemplateDefinition(context.Background(), actor, "se"); !errors.Is(err, domain.ErrUnauthenticated) {
			t.Errorf("get %+v: want ErrUnauthenticated, got %v", actor, err)
		}
	}
}

func TestTemplates_CallerMutationDoesNotAffectCatalog(t *testing.T) {
	svc := service.NewProjectConfigurationService(newFakeProjectRepo())

	first, _ := svc.RetrieveTemplateDefinition(context.Background(), teamMember, "se")
	first.Roles[0] = "hacked"
	first.Capabilities[0] = "hacked"

	second, _ := svc.RetrieveTemplateDefinition(context.Background(), teamMember, "se")
	if second.Roles[0] != "product_owner" || second.Capabilities[0] != "user_story_guidance" {
		t.Fatalf("catalog was mutated: %+v", second)
	}
}
