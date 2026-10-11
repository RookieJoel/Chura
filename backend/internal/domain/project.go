package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxProjectNameRunes        = 100
	MaxProjectDescriptionRunes = 2000
	MaxUserIDRunes             = 100
	MaxEmailRunes              = 254

	// Caps on validateWorkItemAttributes input, applied in every Template.
	MaxWorkItemTitleRunes       = 200
	MaxWorkItemDescriptionRunes = 10000
	MaxAcceptanceCriteria       = 50
	MaxAcceptanceCriterionRunes = 1000
)

type ProjectRole string

type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TemplateID  string   `json:"template_id"`
	Mode        Mode     `json:"mode"`
	CreatedBy   string   `json:"created_by"`
	Members     []Member `json:"members"`
	// GroupID is the identity-provider group holding the members; internal only.
	GroupID   string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Member struct {
	UserID string      `json:"user_id"`
	Email  string      `json:"email"`
	Name   string      `json:"name"`
	Role   ProjectRole `json:"role"`
}

// DirectoryUser is a user known to the identity provider. Role is empty when
// the user holds no Chura role or both.
type DirectoryUser struct {
	ID    string
	Email string
	Name  string
	Role  SystemRole
}

// ValidateProjectInput returns the chosen Template when the input is valid.
func ValidateProjectInput(name, description, templateID string) (Template, error) {
	var violations []Violation
	switch trimmed := strings.TrimSpace(name); {
	case trimmed == "":
		violations = append(violations, Violation{Field: "name", Message: "name is required"})
	case utf8.RuneCountInString(trimmed) > MaxProjectNameRunes:
		violations = append(violations, Violation{Field: "name", Message: fmt.Sprintf("name must be at most %d characters", MaxProjectNameRunes)})
	}
	if utf8.RuneCountInString(strings.TrimSpace(description)) > MaxProjectDescriptionRunes {
		violations = append(violations, Violation{Field: "description", Message: fmt.Sprintf("description must be at most %d characters", MaxProjectDescriptionRunes)})
	}
	template, found := TemplateByID(templateID)
	if !found {
		violations = append(violations, Violation{Field: "template_id", Message: "unknown template"})
	}
	if len(violations) > 0 {
		return Template{}, &InvalidInputError{Violations: violations}
	}
	return template, nil
}

// MemberByID returns the member with the given user id.
func (p Project) MemberByID(userID string) (Member, bool) {
	for _, m := range p.Members {
		if m.UserID == userID {
			return m, true
		}
	}
	return Member{}, false
}

// CountRole returns how many members hold the given role.
func (p Project) CountRole(role ProjectRole) int {
	count := 0
	for _, m := range p.Members {
		if m.Role == role {
			count++
		}
	}
	return count
}
