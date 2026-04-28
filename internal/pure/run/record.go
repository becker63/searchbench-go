package run

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// Run is the phase-typed core run record.
//
// The phase type parameter prevents lifecycle mistakes such as scoring a
// planned run or executing something that has not been prepared.
type Run[P Phase] struct {
	id        domain.RunID
	spec      Spec
	phase     P
	createdAt time.Time
}

// NewPlanned creates the initial run state.
func NewPlanned(spec Spec) Run[Planned] {
	return Run[Planned]{
		id:        spec.ID,
		spec:      spec,
		phase:     Planned{},
		createdAt: time.Now().UTC(),
	}
}

func (r Run[P]) ID() domain.RunID {
	return r.id
}

// Spec returns the immutable planned request that this run is executing.
func (r Run[P]) Spec() Spec {
	return r.spec
}

// CreatedAt returns when the run record was first created.
func (r Run[P]) CreatedAt() time.Time {
	return r.createdAt
}

// Phase returns the current type-level phase marker.
func (r Run[P]) Phase() P {
	return r.phase
}

// PreparedRun means the backend session has been initialized successfully.
//
// For iterative-context, this means repo/session/policy setup has happened and
// tools are ready to call.
type PreparedRun struct {
	Run[Prepared]

	SessionID  domain.SessionID `json:"session_id"`
	PreparedAt time.Time        `json:"prepared_at"`
}

// NewPrepared advances a planned run into prepared state.
func NewPrepared(planned Run[Planned], sessionID domain.SessionID) PreparedRun {
	return PreparedRun{
		Run: Run[Prepared]{
			id:        planned.id,
			spec:      planned.spec,
			phase:     Prepared{},
			createdAt: planned.createdAt,
		},
		SessionID:  sessionID,
		PreparedAt: time.Now().UTC(),
	}
}

// ExecutedRun is the successful result of running a prepared backend session.
//
// Failed executions should return an error instead of constructing ExecutedRun.
// If this type exists, Searchbench has a prediction and usage summary.
type ExecutedRun struct {
	Run[Executed]

	SessionID  domain.SessionID    `json:"session_id"`
	Prediction domain.Prediction   `json:"prediction"`
	Usage      domain.UsageSummary `json:"usage"`
	TraceID    domain.TraceID      `json:"trace_id,omitempty"`
	StartedAt  time.Time           `json:"started_at"`
	FinishedAt time.Time           `json:"finished_at"`
}

// NewExecuted advances a prepared run into the successful executed state.
func NewExecuted(
	prepared PreparedRun,
	prediction domain.Prediction,
	usage domain.UsageSummary,
	traceID domain.TraceID,
	startedAt time.Time,
	finishedAt time.Time,
) ExecutedRun {
	return ExecutedRun{
		Run: Run[Executed]{
			id:        prepared.id,
			spec:      prepared.spec,
			phase:     Executed{},
			createdAt: prepared.createdAt,
		},
		SessionID:  prepared.SessionID,
		Prediction: prediction,
		Usage:      usage,
		TraceID:    traceID,
		StartedAt:  startedAt.UTC(),
		FinishedAt: finishedAt.UTC(),
	}
}
