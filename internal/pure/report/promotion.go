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

	// TODO(issue-32): remove these transitional constant names after callers
	// finish migrating to challenger-explicit decisions.
	DecisionPromote = DecisionPromoteChallenger
	DecisionReject  = DecisionRejectChallenger
)

// Decision explains the round decision.
type Decision struct {
	Decision DecisionKind `json:"decision"`
	Reason   string       `json:"reason"`
}

// PromotionDecision is a transitional alias for code still migrating to
// round Decision vocabulary.
//
// TODO(issue-32): remove after app/report callers use Decision directly.
type PromotionDecision = Decision
