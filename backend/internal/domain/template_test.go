package domain

import "testing"

func TestReflectionTemplateValidateAnswers(t *testing.T) {
	template := DefaultReflectionTemplate()

	valid := map[string]string{
		ReflectionWentWell: "good teamwork",
		ReflectionStop:     "late merges",
		ReflectionStart:    "writing tests",
	}
	if err := template.ValidateAnswers(valid); err != nil {
		t.Fatalf("want valid, got %v", err)
	}

	missing := map[string]string{ReflectionWentWell: "good teamwork"}
	if err := template.ValidateAnswers(missing); err == nil {
		t.Fatal("want error for missing attributes")
	}

	unknown := map[string]string{
		ReflectionWentWell: "a",
		ReflectionStop:     "b",
		ReflectionStart:    "c",
		"other":            "d",
	}
	if err := template.ValidateAnswers(unknown); err == nil {
		t.Fatal("want error for unknown attribute")
	}
}
