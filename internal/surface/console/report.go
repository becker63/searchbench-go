package console

import (
	"fmt"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/score"
	"github.com/charmbracelet/lipgloss"
)

// RenderCandidateReport renders a candidate report for human terminal output.
func RenderCandidateReport(r report.CandidateReport, opts Options) string {
	if opts == (Options{}) {
		opts = DefaultOptions()
	}
	styles := NewStyles(opts)

	sections := []string{
		renderTitle(r, styles),
		renderDecision(r, styles),
		renderSystems(r, styles),
		renderRunSummary(r, styles),
		renderMetrics(r, styles),
		renderRegressions(r, styles),
		renderFailures(r, styles),
	}
	return strings.Join(sections, "\n\n")
}

// RenderCandidateReportDefault renders a candidate report using default options.
func RenderCandidateReportDefault(r report.CandidateReport) string {
	return RenderCandidateReport(r, DefaultOptions())
}

func renderTitle(r report.CandidateReport, styles Styles) string {
	title := styles.Title.Render("Searchbench Candidate Report")
	if r.ID == "" {
		return title
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", styles.Muted.Render(r.ID.String()))
}

func renderDecision(r report.CandidateReport, styles Styles) string {
	bodyStyle := stylesForDecision(r.Decision.Decision, styles)
	body := bodyStyle.Render(string(r.Decision.Decision))
	if r.Decision.Reason != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, styles.Muted.Render(r.Decision.Reason))
	}
	return renderSection("Decision", styles, styles.Box.Render(body))
}

func stylesForDecision(decision report.Decision, styles Styles) lipgloss.Style {
	switch decision {
	case report.DecisionPromote:
		return styles.Success
	case report.DecisionReview:
		return styles.Warning
	case report.DecisionReject:
		return styles.Danger
	default:
		return styles.Subtitle
	}
}

func renderSystems(r report.CandidateReport, styles Styles) string {
	lines := []string{
		renderSystemLine("Baseline", r.Spec.Systems.Baseline, styles),
		renderSystemLine("Candidate", r.Spec.Systems.Candidate, styles),
	}
	if r.Spec.Systems.Candidate.Policy != nil {
		policy := r.Spec.Systems.Candidate.Policy
		lines = append(lines, fmt.Sprintf(
			"Policy      %s:%s  %s",
			policy.Language,
			policy.Entrypoint,
			shortHash(string(policy.SHA256)),
		))
	}
	return renderSection("Systems", styles, strings.Join(lines, "\n"))
}

func renderRunSummary(r report.CandidateReport, styles Styles) string {
	lines := []string{
		fmt.Sprintf("Baseline   %d succeeded   %d failed", len(r.Runs.Baseline), len(r.Failures.Baseline)),
		fmt.Sprintf("Candidate  %d succeeded   %d failed", len(r.Runs.Candidate), len(r.Failures.Candidate)),
	}
	return renderSection("Run Summary", styles, strings.Join(lines, "\n"))
}

func renderMetrics(r report.CandidateReport, styles Styles) string {
	lines := make([]string, 0, len(r.Comparisons)+1)
	lines = append(lines, renderTableRow(
		styles.TableHead,
		columnWidths([]string{"Metric", "Baseline", "Candidate", "Delta", "Result"}),
		"Metric", "Baseline", "Candidate", "Delta", "Result",
	))
	for _, comparison := range r.Comparisons {
		result, style := metricResult(comparison, styles)
		lines = append(lines, renderTableRow(
			styles.TableCell,
			columnWidths([]string{"", "", "", "", ""}),
			string(comparison.Metric),
			fmt.Sprintf("%.2f", comparison.Baseline),
			fmt.Sprintf("%.2f", comparison.Candidate),
			fmt.Sprintf("%+.2f", comparison.Delta),
			style.Render(result),
		))
	}
	return renderSection("Metrics", styles, strings.Join(lines, "\n"))
}

func renderRegressions(r report.CandidateReport, styles Styles) string {
	if len(r.Regressions) == 0 {
		return renderSection("Regressions", styles, styles.Muted.Render("none"))
	}

	lines := make([]string, 0, len(r.Regressions)+1)
	lines = append(lines, renderTableRow(
		styles.TableHead,
		columnWidths([]string{"Task", "Metric", "Baseline", "Candidate", "Delta", "Severity", "Reason"}),
		"Task", "Metric", "Baseline", "Candidate", "Delta", "Severity", "Reason",
	))
	for _, regression := range r.Regressions {
		lines = append(lines, renderTableRow(
			styles.TableCell,
			columnWidths([]string{"", "", "", "", "", "", ""}),
			regression.TaskID.String(),
			string(regression.Metric),
			fmt.Sprintf("%.2f", regression.Baseline),
			fmt.Sprintf("%.2f", regression.Candidate),
			fmt.Sprintf("%+.2f", regression.Delta),
			string(regression.Severity),
			regression.Reason,
		))
	}
	return renderSection("Regressions", styles, strings.Join(lines, "\n"))
}

func renderFailures(r report.CandidateReport, styles Styles) string {
	failures := make([]failureRow, 0, len(r.Failures.Baseline)+len(r.Failures.Candidate))
	for _, failure := range r.Failures.Baseline {
		failures = append(failures, failureRow{Role: domain.RoleBaseline, Failure: failure})
	}
	for _, failure := range r.Failures.Candidate {
		failures = append(failures, failureRow{Role: domain.RoleCandidate, Failure: failure})
	}
	if len(failures) == 0 {
		return renderSection("Failures", styles, styles.Muted.Render("none"))
	}

	lines := make([]string, 0, len(failures)+1)
	lines = append(lines, renderTableRow(
		styles.TableHead,
		columnWidths([]string{"Role", "Run ID", "Task ID", "System ID", "Stage", "Message"}),
		"Role", "Run ID", "Task ID", "System ID", "Stage", "Message",
	))
	for _, item := range failures {
		lines = append(lines, renderTableRow(
			styles.TableCell,
			columnWidths([]string{"", "", "", "", "", ""}),
			string(item.Role),
			item.Failure.RunID.String(),
			item.Failure.TaskID.String(),
			item.Failure.System.String(),
			string(item.Failure.Stage),
			item.Failure.Message,
		))
	}
	return renderSection("Failures", styles, strings.Join(lines, "\n"))
}

func renderSection(title string, styles Styles, body string) string {
	return lipgloss.JoinVertical(lipgloss.Left, styles.Section.Render(title), body)
}

func renderSystemLine(label string, system domain.SystemRef, styles Styles) string {
	main := fmt.Sprintf(
		"%-10s %-18s %s/%s",
		label,
		system.Backend,
		system.Model.Provider,
		system.Model.Name,
	)
	meta := []string{
		system.ID.String(),
		shortFingerprint(system.Fingerprint),
	}
	if system.PromptBundle.Name != "" {
		meta = append(meta, system.PromptBundle.Name)
		if system.PromptBundle.Version != "" {
			meta[len(meta)-1] = meta[len(meta)-1] + ":" + system.PromptBundle.Version
		}
	}
	return main + "  " + styles.Muted.Render(strings.Join(meta, "  "))
}

func renderTableRow(style lipgloss.Style, widths []int, cols ...string) string {
	padded := make([]string, 0, len(cols))
	for i, col := range cols {
		width := 12
		if i < len(widths) && widths[i] > 0 {
			width = widths[i]
		}
		padded = append(padded, lipgloss.NewStyle().Width(width).Render(col))
	}
	return style.Render(strings.Join(padded, "  "))
}

func columnWidths(cols []string) []int {
	// Stable widths keep the report readable without adding a table library.
	if len(cols) == 5 {
		return []int{18, 10, 10, 10, 10}
	}
	if len(cols) == 6 {
		return []int{10, 14, 12, 14, 10, 40}
	}
	if len(cols) == 7 {
		return []int{10, 18, 10, 10, 10, 10, 24}
	}
	widths := make([]int, len(cols))
	for i, col := range cols {
		if col == "" {
			widths[i] = 12
			continue
		}
		widths[i] = max(12, len(col))
	}
	return widths
}

func metricResult(comparison report.ScoreComparison, styles Styles) (string, lipgloss.Style) {
	direction, ok := score.DirectionForMetric(comparison.Metric)
	if !ok || comparison.Delta == 0 {
		return "same", styles.Muted
	}
	if score.Improved(direction, comparison.Baseline, comparison.Candidate) {
		return "improved", styles.Success
	}
	return "regressed", styles.Danger
}

func shortHash(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func shortFingerprint(value domain.SystemFingerprint) string {
	return shortHash(string(value))
}

type failureRow struct {
	Role    domain.Role
	Failure run.RunFailure
}
