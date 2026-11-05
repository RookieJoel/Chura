package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func TestListAccessibleProjectsFiltersBeforePagination(t *testing.T) {
	repo := newFakeProjectRepo()
	repo.projects["p2"] = domain.Project{ID: "p2", Name: "Beta", Description: "second", TemplateID: "se", Mode: domain.ModeSE}
	repo.projects["p1"] = domain.Project{ID: "p1", Name: "Alpha", Description: "first", TemplateID: "se", Mode: domain.ModeSE}
	repo.projects["p3"] = domain.Project{ID: "p3", Name: "Gamma", TemplateID: "se", Mode: domain.ModeSE}
	directory := newFakeDirectory(repo.log)
	directory.projectMemberships = []string{"p1", "p2", "stale-project"}
	svc := service.NewProjectConfigurationService(repo, directory)

	got, err := svc.ListAccessibleProjects(context.Background(), domain.Actor{
		UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember,
	}, in.ProjectListQuery{Page: 2, PageSize: 1})
	if err != nil {
		t.Fatalf("ListAccessibleProjects() error = %v", err)
	}
	want := in.ProjectListResult{
		Items: []domain.ProjectSummary{{
			ID: "p2", Name: "Beta", Description: "second", TemplateID: "se", Mode: domain.ModeSE,
		}},
		Page: 2, PageSize: 1, Total: 2,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(repo.listQuery.ProjectIDs, directory.projectMemberships) {
		t.Fatalf("repository project IDs = %v, want membership IDs %v", repo.listQuery.ProjectIDs, directory.projectMemberships)
	}
	if repo.listQuery.Offset != 1 || repo.listQuery.Limit != 1 {
		t.Fatalf("repository pagination = offset %d, limit %d; want offset 1, limit 1", repo.listQuery.Offset, repo.listQuery.Limit)
	}
}

func TestListAccessibleProjectsAuditorSeesAllWithoutDirectoryLookup(t *testing.T) {
	repo := newFakeProjectRepo()
	repo.projects["p2"] = domain.Project{ID: "p2", Name: "Beta", TemplateID: "se", Mode: domain.ModeSE}
	repo.projects["p1"] = domain.Project{ID: "p1", Name: "Alpha", TemplateID: "se", Mode: domain.ModeSE}
	directory := newFakeDirectory(repo.log)
	svc := service.NewProjectConfigurationService(repo, directory)

	got, err := svc.ListAccessibleProjects(context.Background(), domain.Actor{
		UserID: "auditor", Email: "auditor@example.com", Role: domain.SystemRoleAuditor,
	}, in.ProjectListQuery{Page: 1, PageSize: 15})
	if err != nil {
		t.Fatalf("ListAccessibleProjects() error = %v", err)
	}
	want := in.ProjectListResult{
		Items: []domain.ProjectSummary{
			{ID: "p1", Name: "Alpha", TemplateID: "se", Mode: domain.ModeSE},
			{ID: "p2", Name: "Beta", TemplateID: "se", Mode: domain.ModeSE},
		},
		Page: 1, PageSize: 15, Total: 2,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if repo.listQuery.ProjectIDs != nil {
		t.Fatalf("auditor repository scope = %v, want all Projects", repo.listQuery.ProjectIDs)
	}
	if len(directory.log.calls) != 1 || directory.log.calls[0] != "repo.ListSummaries" {
		t.Fatalf("calls = %v, want repository listing only and no directory lookup", directory.log.calls)
	}
}

func TestListAccessibleProjectsEmptyMembershipsReturnEmptyPage(t *testing.T) {
	repo := newFakeProjectRepo()
	directory := newFakeDirectory(repo.log)
	svc := service.NewProjectConfigurationService(repo, directory)

	got, err := svc.ListAccessibleProjects(context.Background(), domain.Actor{
		UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember,
	}, in.ProjectListQuery{Page: 1, PageSize: 15})
	if err != nil {
		t.Fatalf("ListAccessibleProjects() error = %v", err)
	}
	want := in.ProjectListResult{Items: []domain.ProjectSummary{}, Page: 1, PageSize: 15}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if len(repo.log.calls) != 1 || repo.log.calls[0] != "ListProjectMemberships(u1)" {
		t.Fatalf("calls = %v, want only membership lookup", repo.log.calls)
	}
}

func TestListAccessibleProjectsDoesNotReturnPartialResultsOnDependencyErrors(t *testing.T) {
	t.Run("directory", func(t *testing.T) {
		repo := newFakeProjectRepo()
		repo.projects["p1"] = domain.Project{ID: "p1", Name: "Alpha"}
		directory := newFakeDirectory(repo.log)
		directory.errs["ListProjectMemberships"] = domain.ErrUnavailable
		svc := service.NewProjectConfigurationService(repo, directory)

		got, err := svc.ListAccessibleProjects(context.Background(), domain.Actor{
			UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember,
		}, in.ProjectListQuery{Page: 1, PageSize: 15})
		if !errors.Is(err, domain.ErrUnavailable) {
			t.Fatalf("error = %v, want ErrUnavailable", err)
		}
		if got.Items != nil {
			t.Fatalf("result = %#v, want zero result on directory failure", got)
		}
	})

	t.Run("repository", func(t *testing.T) {
		repo := newFakeProjectRepo()
		repo.listErr = domain.ErrUnavailable
		directory := newFakeDirectory(repo.log)
		svc := service.NewProjectConfigurationService(repo, directory)

		got, err := svc.ListAccessibleProjects(context.Background(), domain.Actor{
			UserID: "auditor", Email: "auditor@example.com", Role: domain.SystemRoleAuditor,
		}, in.ProjectListQuery{Page: 1, PageSize: 15})
		if !errors.Is(err, domain.ErrUnavailable) {
			t.Fatalf("error = %v, want ErrUnavailable", err)
		}
		if got.Items != nil {
			t.Fatalf("result = %#v, want zero result on repository failure", got)
		}
	})
}

func TestListAccessibleProjectsValidatesActorAndPagination(t *testing.T) {
	repo := newFakeProjectRepo()
	svc := service.NewProjectConfigurationService(repo, newFakeDirectory(repo.log))
	actor := domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}
	for _, tc := range []struct {
		name  string
		actor domain.Actor
		query in.ProjectListQuery
	}{
		{"unauthenticated", domain.Actor{}, in.ProjectListQuery{Page: 1, PageSize: 15}},
		{"page zero", actor, in.ProjectListQuery{Page: 0, PageSize: 15}},
		{"page size zero", actor, in.ProjectListQuery{Page: 1, PageSize: 0}},
		{"negative page size", actor, in.ProjectListQuery{Page: 1, PageSize: -1}},
		{"page size over maximum", actor, in.ProjectListQuery{Page: 1, PageSize: 31}},
		{"offset overflow", actor, in.ProjectListQuery{Page: int(^uint(0) >> 1), PageSize: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.ListAccessibleProjects(context.Background(), tc.actor, tc.query)
			if err == nil {
				t.Fatal("ListAccessibleProjects() error = nil, want invalid input")
			}
			if tc.name == "unauthenticated" && !errors.Is(err, domain.ErrUnauthenticated) {
				t.Fatalf("error = %v, want ErrUnauthenticated", err)
			}
			if tc.name != "unauthenticated" && !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}
