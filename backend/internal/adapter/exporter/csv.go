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

	rows := [][]string{
		{"sprint", summary.Sprint.Name},
		{"team", summary.Sprint.Team},
		{"status", string(summary.Sprint.Status)},
		{"reflection_count", strconv.Itoa(summary.ReflectionCount)},
		{},
		{"author", "content", "created_at"},
	}
	for _, r := range summary.Reflections {
		rows = append(rows, []string{r.Author, r.Content, r.CreatedAt.Format(time.RFC3339)})
	}

	if err := w.WriteAll(rows); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
