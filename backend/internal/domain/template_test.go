package domain_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestCatalog_EveryTemplateAsksWentWellStopStart(t *testing.T) {
	for _, tmpl := range domain.AvailableTemplates() {
		var keys []string
		for _, topic := range tmpl.Reflection.Topics {
			keys = append(keys, topic.Key)
		}
		want := []string{domain.ReflectionWentWell, domain.ReflectionStop, domain.ReflectionStart}
		if !reflect.DeepEqual(keys, want) {
			t.Errorf("%s reflection topics = %v, want %v", tmpl.ID, keys, want)
		}
	}
}

func TestCatalog_OnlySEOffersStoryPointMetrics(t *testing.T) {
	general, _ := domain.TemplateByID(domain.TemplateIDGeneral)
	se, _ := domain.TemplateByID(domain.TemplateIDSE)

	if _, ok := general.SprintReview.Metric(domain.MetricPlannedPoints); ok {
		t.Error("general template must not offer story point metrics")
	}
	if _, ok := se.SprintReview.Metric(domain.MetricPlannedPoints); !ok {
		t.Error("SE template must offer story point metrics")
	}
	for _, tmpl := range []domain.Template{general, se} {
		for _, key := range tmpl.SprintReview.DefaultMetrics {
			if _, ok := tmpl.SprintReview.Metric(key); !ok {
				t.Errorf("%s default metric %q is not offered", tmpl.ID, key)
			}
		}
		for _, m := range tmpl.SprintReview.Metrics {
			if m.Label == "" || m.Unit == "" {
				t.Errorf("%s metric %q has no label or unit", tmpl.ID, m.Key)
			}
		}
	}
}

func TestValidateAnswers_TrimsValidAnswers(t *testing.T) {
	got, err := domain.StandardReflectionTemplate().ValidateAnswers(map[string]string{
		domain.ReflectionWentWell: "  pair programming ",
		domain.ReflectionStop:     "late merges",
		domain.ReflectionStart:    "writing tests first",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		domain.ReflectionWentWell: "pair programming",
		domain.ReflectionStop:     "late merges",
		domain.ReflectionStart:    "writing tests first",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestValidateAnswers_ReportsEveryProblem(t *testing.T) {
	_, err := domain.StandardReflectionTemplate().ValidateAnswers(map[string]string{
		domain.ReflectionWentWell: "   ",
		domain.ReflectionStop:     strings.Repeat("x", domain.MaxReflectionAnswerRunes+1),
		"zzz":                     "b",
		"aaa":                     "a",
	})

	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want InvalidInputError, got %v", err)
	}
	var fields []string
	for _, v := range invalid.Violations {
		fields = append(fields, v.Field)
	}
	want := []string{"answers.aaa", "answers.zzz", "answers.went_well", "answers.stop", "answers.start"}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("violation fields = %v, want %v", fields, want)
	}
}

func TestValidateAnswers_OptionalTopicMayBeBlank(t *testing.T) {
	tmpl := domain.ReflectionTemplate{Topics: []domain.ReflectionTopic{
		{Key: "went_well", Question: "What went well?", Required: true},
		{Key: "notes", Question: "Anything else?"},
	}}
	got, err := tmpl.ValidateAnswers(map[string]string{"went_well": "demo", "notes": " "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, map[string]string{"went_well": "demo"}) {
		t.Fatalf("got %v", got)
	}
}

func TestSelectMetrics_EmptyRequestUsesDefaults(t *testing.T) {
	tmpl := domain.GeneralSprintReviewTemplate()
	got, err := tmpl.SelectMetrics(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(tmpl.DefaultMetrics) {
		t.Fatalf("got %d metrics, want %d", len(got), len(tmpl.DefaultMetrics))
	}
}

func TestSelectMetrics_KeepsTemplateOrderAndDropsDuplicates(t *testing.T) {
	got, err := domain.SESprintReviewTemplate().SelectMetrics([]domain.MetricKey{
		domain.MetricCompletedPoints, domain.MetricPlannedItems, domain.MetricCompletedPoints,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var keys []domain.MetricKey
	for _, m := range got {
		keys = append(keys, m.Key)
	}
	want := []domain.MetricKey{domain.MetricPlannedItems, domain.MetricCompletedPoints}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("got %v, want %v", keys, want)
	}
}

func TestSelectMetrics_RejectsMetricOutsideTemplate(t *testing.T) {
	_, err := domain.GeneralSprintReviewTemplate().SelectMetrics([]domain.MetricKey{domain.MetricPlannedPoints, "bogus"})

	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 2 || invalid.Violations[0].Field != "metrics" {
		t.Fatalf("want 2 metrics violations, got %v", err)
	}
}
