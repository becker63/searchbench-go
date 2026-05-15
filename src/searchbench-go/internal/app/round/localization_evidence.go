package round

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func enrichLocalizationEvidence(evidence *score.RoundEvidenceDocument, records []report.MatchExecutionRecord) {
	if evidence == nil || len(records) == 0 {
		return
	}
	n := 0
	for _, rec := range records {
		gold := rec.Task.Oracle.GoldFiles
		n += countPredictionMisses(rec.Incumbent, gold)
		n += countPredictionMisses(rec.Challenger, gold)
	}
	evidence.InvalidPredictions = score.InvalidPredictionEvidence{
		Known: true,
		Count: n,
	}
}

func countPredictionMisses(out report.RoleExecutionOutcome, gold []domain.RepoRelPath) int {
	if out.Failed != nil || out.Scored == nil {
		return 0
	}
	if predictionMissesGold(out.Scored.Execution.Prediction, gold) {
		return 1
	}
	return 0
}

func predictionMissesGold(pred domain.Prediction, gold []domain.RepoRelPath) bool {
	if len(gold) == 0 {
		return len(pred.Files) == 0
	}
	if len(pred.Files) == 0 {
		return true
	}
	want := make(map[domain.RepoRelPath]struct{}, len(gold))
	for _, g := range gold {
		want[g] = struct{}{}
	}
	for _, f := range pred.Files {
		if _, ok := want[f]; ok {
			return false
		}
	}
	return true
}
