package report

// Decision is the final release-gate recommendation for a candidate system.
type Decision string

const (
	// DecisionPromote recommends replacing the baseline with the candidate.
	DecisionPromote Decision = "PROMOTE"
	// DecisionReview recommends human review before promotion.
	DecisionReview Decision = "REVIEW"
	// DecisionReject recommends not promoting the candidate.
	DecisionReject Decision = "REJECT"
)

// PromotionDecision explains whether a candidate should replace the baseline.
type PromotionDecision struct {
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason"`
}
