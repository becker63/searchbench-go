package report

// Decision is the final release-gate recommendation for a candidate system.
type Decision string

const (
	DecisionPromote Decision = "PROMOTE"
	DecisionReview  Decision = "REVIEW"
	DecisionReject  Decision = "REJECT"
)

// PromotionDecision explains whether a candidate should replace the baseline.
type PromotionDecision struct {
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason"`
}
