package artifact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func marshalScoreEvidencePkl(doc score.ScoreEvidenceDocument) ([]byte, error) {
	if err := doc.Validate(); err != nil {
		return nil, err
	}

	var w pklWriter
	w.linef("schemaVersion = %s", pklString(doc.SchemaVersion))
	w.linef("reportId = %s", pklString(doc.ReportID.String()))
	w.blank()

	w.object("systems", func() {
		writeSystemRef(&w, "baseline", doc.Systems.Baseline)
		writeSystemRef(&w, "candidate", doc.Systems.Candidate)
	})
	w.blank()

	writeRoleCounts(&w, "runCounts", doc.RunCounts)
	w.blank()
	writeRoleCounts(&w, "failureCounts", doc.FailureCounts)
	w.blank()

	w.object("localizationDistance", func() {
		writeOptionalMetricEvidence(&w, "goldHop", doc.LocalizationDistance.GoldHop)
		writeOptionalMetricEvidence(&w, "issueHop", doc.LocalizationDistance.IssueHop)
	})
	w.blank()

	writeUsageEvidence(&w, "usage", doc.Usage)
	w.blank()
	writeUsageEvidence(&w, "baselineUsage", doc.BaselineUsage)
	w.blank()

	w.object("regressions", func() {
		w.linef("count = %d", doc.Regressions.Count)
		w.linef("minorCount = %d", doc.Regressions.MinorCount)
		w.linef("severeCount = %d", doc.Regressions.SevereCount)
	})
	w.blank()

	w.list("regressionDetails", func() {
		for _, detail := range doc.RegressionDetails {
			w.listObject(func() {
				w.linef("taskId = %s", pklString(detail.TaskID.String()))
				w.linef("metric = %s", pklString(string(detail.Metric)))
				w.linef("baseline = %s", pklFloat(detail.Baseline))
				w.linef("candidate = %s", pklFloat(detail.Candidate))
				w.linef("delta = %s", pklFloat(detail.Delta))
				w.linef("severity = %s", pklString(detail.Severity))
				w.linef("reason = %s", pklString(detail.Reason))
			})
		}
	})
	w.blank()

	w.object("invalidPredictions", func() {
		w.linef("known = %s", pklBool(doc.InvalidPredictions.Known))
		w.linef("count = %d", doc.InvalidPredictions.Count)
	})
	w.blank()

	w.list("metrics", func() {
		for _, metric := range doc.Metrics {
			metric := metric
			w.listObject(func() {
				writeMetricEvidenceBody(&w, metric)
			})
		}
	})
	w.blank()

	w.object("promotionDecision", func() {
		w.linef("decision = %s", pklString(doc.PromotionDecision.Decision))
		w.linef("reason = %s", pklString(doc.PromotionDecision.Reason))
	})

	return []byte(w.String()), nil
}

// MarshalScoreEvidencePKL deterministically serializes score evidence into the
// Pkl-native score artifact format used by local scoring and bundles.
func MarshalScoreEvidencePKL(doc score.ScoreEvidenceDocument) ([]byte, error) {
	return marshalScoreEvidencePkl(doc)
}

func writeSystemRef(w *pklWriter, name string, ref domain.SystemRef) {
	w.object(name, func() {
		w.linef("id = %s", pklString(ref.ID.String()))
		w.linef("name = %s", pklString(ref.Name))
		w.linef("backend = %s", pklString(string(ref.Backend)))
		w.object("model", func() {
			w.linef("provider = %s", pklString(ref.Model.Provider))
			w.linef("name = %s", pklString(ref.Model.Name))
		})
		w.object("promptBundle", func() {
			w.linef("name = %s", pklString(ref.PromptBundle.Name))
			w.linef("version = %s", pklString(ref.PromptBundle.Version))
		})
		if ref.Policy == nil {
			w.line("policy = null")
		} else {
			w.object("policy", func() {
				w.linef("id = %s", pklString(ref.Policy.ID.String()))
				w.linef("language = %s", pklString(string(ref.Policy.Language)))
				w.linef("sha256 = %s", pklString(string(ref.Policy.SHA256)))
				w.linef("entrypoint = %s", pklString(ref.Policy.Entrypoint))
			})
		}
		w.object("runtime", func() {
			w.linef("maxSteps = %d", ref.Runtime.MaxSteps)
			w.linef("maxToolCalls = %d", ref.Runtime.MaxToolCalls)
			w.linef("maxInputTokens = %d", ref.Runtime.MaxInputTokens)
			w.linef("maxOutputTokens = %d", ref.Runtime.MaxOutputTokens)
			w.linef("maxContextTokens = %d", ref.Runtime.MaxContextTokens)
			w.linef("toolResultMaxBytes = %d", ref.Runtime.ToolResultMaxBytes)
		})
		w.linef("fingerprint = %s", pklString(string(ref.Fingerprint)))
	})
}

func writeRoleCounts(w *pklWriter, name string, counts score.RoleCounts) {
	w.object(name, func() {
		w.linef("baseline = %d", counts.Baseline)
		w.linef("candidate = %d", counts.Candidate)
	})
}

func writeOptionalMetricEvidence(w *pklWriter, name string, metric *score.MetricEvidence) {
	if metric == nil {
		w.linef("%s = null", name)
		return
	}
	w.object(name, func() {
		writeMetricEvidenceBody(w, *metric)
	})
}

func writeMetricEvidenceBody(w *pklWriter, metric score.MetricEvidence) {
	w.linef("metric = %s", pklString(string(metric.Metric)))
	w.linef("direction = %s", pklString(string(metric.Direction)))
	w.linef("baseline = %s", pklFloat(metric.Baseline))
	w.linef("candidate = %s", pklFloat(metric.Candidate))
	w.linef("delta = %s", pklFloat(metric.Delta))
	w.linef("improved = %s", pklBool(metric.Improved))
	w.linef("regressed = %s", pklBool(metric.Regressed))
}

func writeUsageEvidence(w *pklWriter, name string, usage score.UsageEvidence) {
	w.object(name, func() {
		w.linef("available = %s", pklBool(usage.Available))
		w.linef("measuredRuns = %d", usage.MeasuredRuns)
		w.linef("inputTokens = %d", usage.InputTokens)
		w.linef("outputTokens = %d", usage.OutputTokens)
		w.linef("totalTokens = %d", usage.TotalTokens)
		w.linef("costUsd = %s", pklFloat(usage.CostUSD))
	})
}

type pklWriter struct {
	lines  []string
	indent int
}

func (w *pklWriter) String() string {
	return strings.Join(w.lines, "\n") + "\n"
}

func (w *pklWriter) blank() {
	w.lines = append(w.lines, "")
}

func (w *pklWriter) line(text string) {
	w.lines = append(w.lines, strings.Repeat("  ", w.indent)+text)
}

func (w *pklWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}

func (w *pklWriter) object(name string, body func()) {
	w.linef("%s {", name)
	w.indent++
	body()
	w.indent--
	w.line("}")
}

func (w *pklWriter) list(name string, body func()) {
	w.linef("%s = new {", name)
	w.indent++
	body()
	w.indent--
	w.line("}")
}

func (w *pklWriter) listObject(body func()) {
	w.line("new {")
	w.indent++
	body()
	w.indent--
	w.line("}")
}

func pklString(value string) string {
	return strconv.Quote(value)
}

func pklBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func pklFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
