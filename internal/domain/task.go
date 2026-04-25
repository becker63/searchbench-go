package domain

// BenchmarkName identifies the source benchmark/task family.
type BenchmarkName string

const (
	// BenchmarkLCA identifies the JetBrains LCA bug-localization benchmark.
	BenchmarkLCA BenchmarkName = "jetbrains-lca"
)

// TaskInput is the portion of a task that may be shown to the agent.
//
// Gold labels and scoring-only fields must not live here.
type TaskInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// TaskOracle is the scoring-only portion of a task.
//
// This must never be included in prompts.
type TaskOracle struct {
	GoldFiles []RepoRelPath `json:"gold_files"`
}

// TaskSpec is a single benchmark task over a specific repo snapshot.
//
// Conceptually:
//
//	TaskSpec = RepoSnapshot + agent-visible input + scorer-visible oracle.
type TaskSpec struct {
	ID        TaskID        `json:"id"`
	Benchmark BenchmarkName `json:"benchmark"`
	Repo      RepoSnapshot  `json:"repo"`
	Input     TaskInput     `json:"input"`
	Oracle    TaskOracle    `json:"oracle"`
}

// AgentInput returns the prompt-safe portion of the task.
func (t TaskSpec) AgentInput() TaskInput {
	return t.Input
}
