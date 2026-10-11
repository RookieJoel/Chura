package domain_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestValidateProjectInput_LengthMessagesUseLimits(t *testing.T) {
	_, err := domain.ValidateProjectInput(strings.Repeat("a", domain.MaxProjectNameRunes+1),
		"  "+strings.Repeat("d", domain.MaxProjectDescriptionRunes)+"  ", "se")

	invalid, ok := err.(*domain.InvalidInputError)
	want := []domain.Violation{{Field: "name", Message: "name must be at most 100 characters"}}
	if !ok || !reflect.DeepEqual(invalid.Violations, want) {
		t.Fatalf("got %v (description surrounding whitespace must not count)", err)
	}
}

func TestErrNotFound_HasGenericMessage(t *testing.T) {
	if domain.ErrNotFound.Error() != "not found" {
		t.Fatalf("ErrNotFound = %q", domain.ErrNotFound)
	}
}
