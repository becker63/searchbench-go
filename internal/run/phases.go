package run

// Phase is a sealed marker interface for run lifecycle phases.
//
// Only this package should define phases. External packages should consume
// Run[Planned], PreparedRun, and ExecutedRun rather than inventing new phases.
type Phase interface {
	phase()
}

// Planned marks a run that has been requested but not prepared.
type Planned struct{}

// Prepared marks a run whose backend/session state has been initialized.
type Prepared struct{}

// Executed marks a run whose execution succeeded and produced a prediction.
type Executed struct{}

func (Planned) phase()  {}
func (Prepared) phase() {}
func (Executed) phase() {}

type PhaseName string

const (
	// PhasePlanned names the planned lifecycle state.
	PhasePlanned PhaseName = "planned"
	// PhasePrepared names the prepared lifecycle state.
	PhasePrepared PhaseName = "prepared"
	// PhaseExecuted names the executed lifecycle state.
	PhaseExecuted PhaseName = "executed"
)

// Name returns the stable phase name.
func (Planned) Name() PhaseName {
	return PhasePlanned
}

// Name returns the stable phase name.
func (Prepared) Name() PhaseName {
	return PhasePrepared
}

// Name returns the stable phase name.
func (Executed) Name() PhaseName {
	return PhaseExecuted
}
