package domain

// BenchmarkName identifies the source benchmark family for a match.
type BenchmarkName string

const (
	// BenchmarkLCA identifies the JetBrains LCA bug-localization benchmark.
	BenchmarkLCA BenchmarkName = "jetbrains-lca"
)

// MatchInput is the portion of a match that may be shown to the agent.
//
// Gold labels and scoring-only fields must not live here.
type MatchInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// MatchOracle is the scoring-only portion of a match.
//
// This must never be included in prompts for an evaluator agent in order to not skew judgement.
type MatchOracle struct {
	GoldFiles []RepoRelPath `json:"gold_files"`
}

// MatchSpec is a single benchmark match over a specific repo snapshot.
//
// Conceptually:
//
//	MatchSpec = RepoSnapshot + agent-visible input + scorer-visible oracle.
type MatchSpec struct {
	ID        MatchID       `json:"id"`
	Benchmark BenchmarkName `json:"benchmark"`
	Repo      RepoSnapshot  `json:"repo"`
	Input     MatchInput    `json:"input"`
	Oracle    MatchOracle   `json:"oracle"`
}

// AgentInput returns the prompt-safe portion of the match.
func (m MatchSpec) AgentInput() MatchInput {
	return m.Input
}
