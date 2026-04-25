package run

// Phase is a sealed marker interface for run lifecycle phases.
//
// Only this package should define phases. External packages should consume
// Run[Planned], PreparedRun, and ExecutedRun rather than inventing new phases.
type Phase interface {
	phase()
}

type Planned struct{}
type Prepared struct{}
type Executed struct{}

func (Planned) phase()  {}
func (Prepared) phase() {}
func (Executed) phase() {}

type PhaseName string

const (
	PhasePlanned  PhaseName = "planned"
	PhasePrepared PhaseName = "prepared"
	PhaseExecuted PhaseName = "executed"
)

func (Planned) Name() PhaseName {
	return PhasePlanned
}

func (Prepared) Name() PhaseName {
	return PhasePrepared
}

func (Executed) Name() PhaseName {
	return PhaseExecuted
}
