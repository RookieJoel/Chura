package domain_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestSprintValidate_NilSprintIsInvalidInput(t *testing.T) {
	var s *domain.Sprint

	if err := s.Validate(); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestSprintValidate_DefaultsStatusToPlanned(t *testing.T) {
	s := &domain.Sprint{Name: "S", Team: "T"}

	if err := s.Validate(); err != nil || s.Status != domain.Planned {
		t.Fatalf("err=%v status=%q", err, s.Status)
	}
}

func TestSprintValidate_TeamTooLong(t *testing.T) {
	s := &domain.Sprint{Name: "S", Team: strings.Repeat("t", domain.MaxSprintTeamRunes+1)}

	var invalid *domain.InvalidInputError
	err := s.Validate()
	want := []domain.Violation{{Field: "team", Message: "team must be at most 120 characters"}}
	if !errors.As(err, &invalid) || !reflect.DeepEqual(invalid.Violations, want) {
		t.Fatalf("got %v", err)
	}
}
