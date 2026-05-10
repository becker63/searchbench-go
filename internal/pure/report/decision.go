package report

// DecisionKind is the final round-gate recommendation.
type DecisionKind string

const (
	// DecisionPromoteChallenger recommends advancing the challenger.
	DecisionPromoteChallenger DecisionKind = "PROMOTE_CHALLENGER"
	// DecisionReview recommends human review before advancing.
	DecisionReview DecisionKind = "REVIEW"
	// DecisionRejectChallenger recommends rejecting the challenger.
	DecisionRejectChallenger DecisionKind = "REJECT_CHALLENGER"
)

// Decision explains the round decision.
type Decision struct {
	Decision DecisionKind `json:"decision"`
	Reason   string       `json:"reason"`
}
