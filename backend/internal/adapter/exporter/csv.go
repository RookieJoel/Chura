// Package exporter renders Sprint Review Summaries as downloadable files.
package exporter

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

// CSVExporter writes one section per block: report info, quantitative
// metrics, then qualitative feedback one answer per row.
type CSVExporter struct{}

var _ out.ReportExporter = CSVExporter{}

func NewCSVExporter() CSVExporter { return CSVExporter{} }

func (CSVExporter) Format() domain.ReportFormat { return domain.ReportFormatCSV }
func (CSVExporter) ContentType() string         { return "text/csv; charset=utf-8" }
func (CSVExporter) FileExtension() string       { return "csv" }

func (CSVExporter) Export(summary *domain.SprintReviewSummary) ([]byte, error) {
	rows := [][]string{
		{"section", "field", "value"},
		{"report", "title", summary.Title},
		{"report", "project", summary.Project.Name},
		{"report", "sprint", summary.Sprint.Name},
		{"report", "team", summary.Sprint.Team},
		{"report", "status", string(summary.Sprint.Status)},
		{"report", "start_date", formatDate(summary.Sprint.StartDate)},
		{"report", "end_date", formatDate(summary.Sprint.EndDate)},
		{"report", "generated_at", summary.GeneratedAt.Format(time.RFC3339)},
		{},
		{"quantitative", "metric", "value", "unit"},
	}
	for _, m := range summary.Quantitative.Metrics {
		rows = append(rows, []string{"quantitative", m.Label, formatValue(m.Value), string(m.Unit)})
	}
	q := summary.Qualitative
	rows = append(rows,
		[]string{},
		[]string{"qualitative", "respondents", strconv.Itoa(q.Respondents) + "/" + strconv.Itoa(q.TeamSize)},
		[]string{"qualitative", "question", "author", "answer"},
	)
	for _, topic := range q.Topics {
		for _, r := range topic.Responses {
			rows = append(rows, []string{"qualitative", topic.Question, r.AuthorEmail, r.Answer})
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, row := range rows {
		for i := range row {
			row[i] = neutralizeFormula(row[i])
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// neutralizeFormula stops spreadsheets from running user text as a formula.
func neutralizeFormula(cell string) string {
	if cell != "" && strings.ContainsRune("=+-@\t\r", rune(cell[0])) {
		return "'" + cell
	}
	return cell
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatValue(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
