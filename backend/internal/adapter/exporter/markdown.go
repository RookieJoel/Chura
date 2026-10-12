package exporter

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

// MarkdownExporter renders the summary as a document ready to present.
type MarkdownExporter struct{}

var _ out.ReportExporter = MarkdownExporter{}

func NewMarkdownExporter() MarkdownExporter { return MarkdownExporter{} }

func (MarkdownExporter) Format() domain.ReportFormat { return domain.ReportFormatMarkdown }
func (MarkdownExporter) ContentType() string         { return "text/markdown; charset=utf-8" }
func (MarkdownExporter) FileExtension() string       { return "md" }

func (MarkdownExporter) Export(summary *domain.SprintReviewSummary) ([]byte, error) {
	var b bytes.Buffer
	s := summary.Sprint
	fmt.Fprintf(&b, "# %s: %s\n\n", oneLine(summary.Title), oneLine(s.Name))
	fmt.Fprintf(&b, "- **Project:** %s\n", oneLine(summary.Project.Name))
	fmt.Fprintf(&b, "- **Team:** %s\n", oneLine(s.Team))
	fmt.Fprintf(&b, "- **Status:** %s\n", s.Status)
	if s.StartDate != nil && s.EndDate != nil {
		fmt.Fprintf(&b, "- **Dates:** %s to %s\n", formatDate(s.StartDate), formatDate(s.EndDate))
	}
	fmt.Fprintf(&b, "- **Generated:** %s\n\n", summary.GeneratedAt.Format(time.RFC3339))

	b.WriteString("## Sprint metrics\n\n| Metric | Value |\n| --- | --- |\n")
	for _, m := range summary.Quantitative.Metrics {
		fmt.Fprintf(&b, "| %s | %s |\n", m.Label, withUnit(m))
	}
	if p := summary.Quantitative.Progress; p != nil {
		fmt.Fprintf(&b, "\n%d of %d sprint days elapsed (%s%%).\n", p.ElapsedDays, p.TotalDays, formatValue(p.TimeElapsedRate))
	}

	q := summary.Qualitative
	fmt.Fprintf(&b, "\n## Team reflection\n\n%d of %d members responded (%s%%).\n", q.Respondents, q.TeamSize, formatValue(q.ResponseRate))
	for _, topic := range q.Topics {
		fmt.Fprintf(&b, "\n### %s\n\n", oneLine(topic.Question))
		if len(topic.Responses) == 0 {
			b.WriteString("_No answers yet._\n")
			continue
		}
		for _, r := range topic.Responses {
			fmt.Fprintf(&b, "- %s (%s)\n", oneLine(r.Answer), oneLine(r.AuthorEmail))
		}
	}
	return b.Bytes(), nil
}

func withUnit(m domain.MetricValue) string {
	switch m.Unit {
	case domain.MetricUnitPercent:
		return formatValue(m.Value) + "%"
	default:
		return formatValue(m.Value) + " " + string(m.Unit)
	}
}

// oneLine keeps user text inside its list item or table cell.
func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "|", "\\|")), " ")
}
