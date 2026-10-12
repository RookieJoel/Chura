package exporter_test

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/adapter/exporter"
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func sampleSummary() *domain.SprintReviewSummary {
	start := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	return &domain.SprintReviewSummary{
		Title:   "Sprint Review Summary",
		Project: domain.ReviewProject{ID: "p1", Name: "Chura"},
		Sprint:  domain.Sprint{ID: "7", Name: "Sprint 1", Team: "Alpha", Status: domain.Completed, StartDate: &start, EndDate: &end},
		Quantitative: domain.QuantitativeSummary{
			Metrics: []domain.MetricValue{
				{Key: domain.MetricCompletedItems, Label: "Completed work items", Unit: domain.MetricUnitItems, Value: 4},
				{Key: domain.MetricCompletionRate, Label: "Completion rate", Unit: domain.MetricUnitPercent, Value: 66.7},
			},
			Progress: &domain.SprintProgress{TotalDays: 12, ElapsedDays: 12, TimeElapsedRate: 100},
		},
		Qualitative: domain.QualitativeSummary{
			Respondents: 1, TeamSize: 2, ResponseRate: 50,
			Topics: []domain.TopicFeedback{
				{Key: "went_well", Question: "What went well?", Responses: []domain.TopicResponse{{AuthorEmail: "u1@example.com", Answer: "=HYPERLINK(\"x\")"}}},
				{Key: "stop", Question: "What should we stop doing?", Responses: []domain.TopicResponse{}},
			},
		},
		GeneratedAt: time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC),
	}
}

func TestCSVExporter_WritesAllSections(t *testing.T) {
	data, err := exporter.NewCSVExporter().Export(sampleSummary())
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v\n%s", err, data)
	}
	text := string(data)
	for _, want := range []string{"report,sprint,Sprint 1", "report,start_date,2026-10-05", "quantitative,Completion rate,66.7,percent", "qualitative,respondents,1/2"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
	last := rows[len(rows)-1]
	if last[1] != "What went well?" || last[2] != "u1@example.com" {
		t.Fatalf("last row = %v", last)
	}
}

func TestCSVExporter_NeutralizesFormulas(t *testing.T) {
	data, _ := exporter.NewCSVExporter().Export(sampleSummary())
	if !strings.Contains(string(data), `'=HYPERLINK`) {
		t.Fatalf("formula was not neutralized:\n%s", data)
	}
}

func TestMarkdownExporter_RendersPresentation(t *testing.T) {
	data, err := exporter.NewMarkdownExporter().Export(sampleSummary())
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		"# Sprint Review Summary: Sprint 1",
		"- **Dates:** 2026-10-05 to 2026-10-16",
		"| Completed work items | 4 items |",
		"| Completion rate | 66.7% |",
		"12 of 12 sprint days elapsed (100%).",
		"1 of 2 members responded (50%).",
		"### What went well?",
		"### What should we stop doing?\n\n_No answers yet._",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
}

func TestExporters_DescribeTheirFormat(t *testing.T) {
	csvExp, md := exporter.NewCSVExporter(), exporter.NewMarkdownExporter()
	if csvExp.Format() != domain.ReportFormatCSV || csvExp.FileExtension() != "csv" || !strings.HasPrefix(csvExp.ContentType(), "text/csv") {
		t.Error("csv exporter metadata")
	}
	if md.Format() != domain.ReportFormatMarkdown || md.FileExtension() != "md" || !strings.HasPrefix(md.ContentType(), "text/markdown") {
		t.Error("markdown exporter metadata")
	}
}
