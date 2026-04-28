package logging

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/charmbracelet/lipgloss"
)

func (l Logger) devInfo(event string, parts ...string) {
	l.writeDevLine("INFO", event, parts...)
}

func (l Logger) devWarn(event string, parts ...string) {
	l.writeDevLine("WARN", event, parts...)
}

func (l Logger) writeDevLine(level string, event string, parts ...string) {
	if l.mode != ModeDev {
		return
	}

	w := l.out
	if w == nil {
		return
	}

	parts = append(parts, l.contextParts()...)
	line := strings.Join(filterEmpty(parts), " · ")
	if line != "" {
		line = "  " + line
	}

	levelStyle := l.styles.Info
	switch level {
	case "WARN":
		levelStyle = l.styles.Warn
	case "ERROR":
		levelStyle = l.styles.Error
	}

	_, _ = fmt.Fprintf(
		w,
		"%s  %-5s  %s%-20s%s\n",
		l.styles.Time.Render(time.Now().Format("15:04:05")),
		levelStyle.Render(level),
		l.renderLoggerName(),
		l.eventStyle(event).Render(event),
		line,
	)
}

func (l Logger) contextParts() []string {
	if len(l.fields) == 0 {
		return nil
	}

	out := make([]string, 0, len(l.fields)/2)
	for i := 0; i+1 < len(l.fields); i += 2 {
		key, ok := l.fields[i].(string)
		if !ok {
			continue
		}
		out = append(out, fmt.Sprintf("%s=%v", key, l.fields[i+1]))
	}
	return out
}

func (l Logger) renderRole(role domain.Role) string {
	switch role {
	case domain.RoleBaseline:
		return l.styles.RoleBase.Render(roleLabel(role))
	case domain.RoleCandidate:
		return l.styles.RoleCand.Render(roleLabel(role))
	default:
		return roleLabel(role)
	}
}

func (l Logger) renderStatus(ok bool) string {
	if ok {
		return l.styles.Success.Render("ok")
	}
	return l.styles.Failure.Render("failed")
}

func (l Logger) renderMetric(name, value string) string {
	return l.styles.Metric.Render(name + "=" + value)
}

func (l Logger) renderMuted(value string) string {
	return l.styles.Muted.Render(value)
}

func (l Logger) renderLoggerName() string {
	if l.name == "" {
		return ""
	}
	return l.styles.Muted.Render("[" + l.name + "] ")
}

func (l Logger) eventStyle(event string) lipgloss.Style {
	switch event {
	case "comparison.started":
		return l.styles.ComparisonStarted
	case "comparison.completed":
		return l.styles.ComparisonCompleted
	case "task.started":
		return l.styles.TaskStarted
	case "task.completed":
		return l.styles.TaskCompleted
	case "run.started":
		return l.styles.RunStarted
	case "run.executed":
		return l.styles.RunExecuted
	case "run.scored":
		return l.styles.RunScored
	case "run.failed":
		return l.styles.RunFailed
	case "report.created":
		return l.styles.ReportCreated
	default:
		return l.styles.Event
	}
}

func filterEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func outputSyncer(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
