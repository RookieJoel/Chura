package exporter

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type CSVExporter struct{}

func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

func (e *CSVExporter) Export(summary *domain.SprintReviewSummary) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	work := summary.Quantitative.Work
	rows := [][]string{
		{"sprint", summary.Sprint.Name},
		{"team", summary.Sprint.Team},
		{"status", string(summary.Sprint.Status)},
		{},
		{"quantitative", "count"},
		{"planned", strconv.Itoa(work.Planned)},
		{"completed", strconv.Itoa(work.Completed)},
		{"carry_over", strconv.Itoa(work.CarryOver)},
		{"blocked", strconv.Itoa(work.Blocked)},
		{},
		{"qualitative", "author", "attribute", "answer", "created_at"},
	}

	for _, r := range summary.Qualitative.Reflections {
		for _, key := range domain.DefaultReflectionTemplate() {
			rows = append(rows, []string{
				"reflection", r.Author, key, r.Answers[key], r.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if err := w.WriteAll(rows); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
