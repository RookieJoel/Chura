package domain

import (
	"fmt"
	"strings"
)

const (
	ReflectionWentWell = "went_well"
	ReflectionStop     = "stop"
	ReflectionStart    = "start"
)

// ReflectionTemplate defines which attributes a sprint reflection must answer.
type ReflectionTemplate struct {
	Name       string   `json:"name"`
	Attributes []string `json:"attributes"`
}

// DefaultReflectionTemplate is used until template definitions are provided
// by the Project Configuration Service.
func DefaultReflectionTemplate() ReflectionTemplate {
	return ReflectionTemplate{
		Name: "default",
		Attributes: []string{
			ReflectionWentWell,
			ReflectionStop,
			ReflectionStart,
		},
	}
}

func (t ReflectionTemplate) HasAttribute(key string) bool {
	for _, attribute := range t.Attributes {
		if attribute == key {
			return true
		}
	}
	return false
}

// ValidateAnswers checks that answers contain every template attribute
// and nothing outside of it.
func (t ReflectionTemplate) ValidateAnswers(answers map[string]string) error {
	for key := range answers {
		if !t.HasAttribute(key) {
			return fmt.Errorf("unknown reflection attribute: %s", key)
		}
	}

	for _, attribute := range t.Attributes {
		if strings.TrimSpace(answers[attribute]) == "" {
			return fmt.Errorf("%s is required", attribute)
		}
	}

	return nil
}
