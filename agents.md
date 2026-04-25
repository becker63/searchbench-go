# AGENTS.md

This repository is `searchbench-go`.

Searchbench-Go is a strongly typed evaluation harness for comparing agentic code-search systems. The core product object is a `report.CandidateReport`: a release-style artifact that records what was compared, what ran, what failed, what improved, what regressed, and whether a candidate should promote.

This project is intentionally type-first, boring, and explicit. Do not add clever shims, generic frameworks, dynamic maps, or parallel data models unless specifically asked.

## Core architecture

Current package responsibilities:

```text
internal/domain    stable vocabulary: IDs, tasks, systems, policies, repos, predictions, usage
internal/run       run lifecycle: specs, phases, executed runs, failures
internal/score     metric types, score sets, metric direction, scored runs
internal/report    candidate reports, comparison specs, regressions, promotion decisions
internal/compare   orchestration: plans, runner, task comparison, aggregation, parallelism
internal/backend   future backend/session/tool boundary
internal/codegraph graph model used by ingestion/scoring
internal/logging   safe structured/event logging
internal/console   human terminal rendering
internal/cli       thin command-line surface
```

The intended flow is:

```text
compare.Plan
  -> compare.Runner.Run
  -> compare.Runner.RunTasks
  -> compare.Runner.CompareTask
  -> compare.Runner.ExecuteAndScore
  -> compare.Results
  -> report.CandidateReport
```

Preserve this flow.

## Most important model distinctions

Do not blur these:

```text
compare.Plan              = executable comparison input
report.ComparisonSpec     = report-safe comparison identity

domain.SystemSpec         = executable system recipe
domain.SystemRef          = report-safe system identity

domain.PolicyArtifact     = may contain source code
domain.PolicyRef          = source-free policy identity

run.ExecutedRun           = execution succeeded
score.ScoredRun           = execution and scoring succeeded
run.RunFailure            = report-facing failure path
```

Reports, logs, CLI output, and console rendering must use report-safe views by default.

## What agents must not do

### Do not create parallel data models

Do not introduce new structs that duplicate existing model concepts.

Bad:

```go
type TaskResult struct {
    TaskID string
    Scores map[string]float64
}
```

Use the existing model instead:

```go
run.ExecutedRun
score.ScoreSet
score.ScoredRun
run.RunFailure
report.CandidateReport
```

If the existing model feels inconvenient, improve the existing model or add a small projection type in an appropriate package. Do not invent a second model.

### Avoid `map[string]any` even at adapter boundaries

Massively prefer not to use:

```go
map[string]any
interface{}
any
reflect
```

Even at adapter boundaries, these should be a last resort, not the default design.

Preferred order:

```text
1. Typed request/response structs
2. Small typed adapter DTOs
3. json.RawMessage at the narrow transport edge
4. map[string]any only when the external API truly has no stable shape
```

If an external API is dynamic, quarantine that dynamism immediately:

```text
provider SDK object
  -> adapter-local DTO / parser
  -> typed Searchbench value
```

Do not let dynamic payloads flow inward.

Acceptable narrow uses:

```text
backend tool arguments at the raw protocol boundary
provider SDK response parsing before typed normalization
telemetry metadata payload construction inside a telemetry adapter
test fixtures for malformed external payloads
```

Even then:

```text
keep it adapter-local
convert to typed values immediately
do not store it on core structs
do not pass it through compare/report/score/domain
do not use it as a convenience escape hatch
```

Bad:

```go
type ExecutedRun struct {
    Raw map[string]any
}

type ScoreInput struct {
    Metadata map[string]any
}

func Score(payload map[string]any) score.ScoreSet
```

Better:

```go
type ProviderUsageDTO struct {
    InputTokens  int64 `json:"input_tokens"`
    OutputTokens int64 `json:"output_tokens"`
    TotalTokens  int64 `json:"total_tokens"`
}

func NormalizeUsage(dto ProviderUsageDTO) domain.UsageSummary
```

If you think you need `map[string]any`, first ask:

```text
Can this be a small typed DTO?
Can this be json.RawMessage until parsed?
Can this live entirely inside the adapter?
Can this be converted before crossing a package boundary?
```

Only use dynamic maps when the answer to all of those is no.

### Do not log or render unsafe values

Never log, render, or JSON-dump these directly:

```text
domain.SystemSpec
domain.PolicyArtifact
domain.TaskSpec with oracle fields
compare.Plan
raw report.CandidateReport in logs
raw model/provider responses
```

Use safe views:

```go
system.Ref()
domain.PolicyRef
logging.SystemSpecKV(system)
logging.TaskKV(task)
logging.ReportSummaryKV(report)
console.RenderCandidateReport(report, opts)
```

Policy source and scorer-only oracle data must not leak into reports, logs, or terminal output.

### Do not put rendering on report types

Do not add:

```go
func (r CandidateReport) String() string
func (r CandidateReport) Pretty() string
```

Rendering belongs in:

```text
internal/console
```

Structured report data belongs in:

```text
internal/report
```

### Do not make `compare` know concrete backends

`compare.Runner` coordinates. It should not know how Iterative Context, jCodeMunch, Eino, Langfuse, OpenAI, Cerebras, tree-sitter, or filesystem materialization works.

Concrete integrations must enter through interfaces:

```go
compare.Executor
compare.GraphProvider
compare.Scorer
compare.Decider
backend.Backend
backend.Session
codegraph.Graph
codegraph.Builder
```

### Do not recreate the old Python mega-localizer

Do not create one large file that owns all of these at once:

```text
prompt construction
model calls
tool dispatch
context budget management
backend sessions
usage normalization
prediction finalization
retry logic
telemetry
scoring
report generation
```

Split these into explicit packages or interfaces. Keep `compare` as orchestration only.

### Do not add generic pipeline frameworks

Avoid abstract pipeline machinery like:

```go
type Pipeline[A, B, C any] struct { ... }
type Step[T any] interface { ... }
type Reducer[A, B any] interface { ... }
```

Prefer Searchbench nouns:

```text
Plan
Runner
TaskComparisonResult
Results
CandidateReport
ScoreSet
RunFailure
```

Generics are welcome for simple structural helpers like `Pair[T]`, `NonEmpty[T]`, or task-aligned containers. They should not hide the domain.

### Do not add unnecessary interfaces

Interfaces should live at consumer boundaries.

Good:

```go
type Executor interface {
    Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error)
}
```

Bad:

```go
type ExecutorProvider interface {
    Executor() Executor
}

type ReportProvider interface {
    Report() report.CandidateReport
}
```

Do not add an interface unless a package consumes behavior through that interface.

### Do not spread concurrency everywhere

Parallelism currently belongs in:

```go
compare.Runner.RunTasks
```

Do not add goroutines, channels, mutexes, worker pools, or concurrent mutation in domain/report/score packages.

Current rule:

```text
parallelize over tasks only
collect TaskWorkResult values
restore deterministic plan order
mutate Results only on the main goroutine
```

Do not parallelize baseline/candidate inside a task unless explicitly requested.

### Do not use package globals for runtime services

Avoid global loggers, global clients, global registries, global backend sessions, or global mutable state.

Prefer explicit construction and injection.

## Error handling rules

Use existing failure concepts.

Internal control/classification error:

```go
compare.StageError
```

Report-facing failure:

```go
run.RunFailure
```

Execution/scoring failures should usually become `run.RunFailure`, not top-level process errors.

Top-level errors are for:

```text
invalid configuration
invalid plan
context cancellation
orchestration bugs
I/O failures at CLI/adapters
```

Wrap errors with context, but do not create huge parallel error taxonomies unless needed.

## Scoring rules

`score.ScoreSet` means all required metrics were computed.

Do not represent missing required metrics with booleans like:

```go
Available bool
```

If a required metric cannot be computed, scoring should fail and become a `RunFailure`.

Metric direction matters. Positive delta is not always good.

Examples:

```text
gold_hop          lower is better
issue_hop         lower is better
token_efficiency  higher is better
cost              lower is better
composite         higher is better
```

Use existing score comparison helpers instead of hand-rolling delta logic.

## Report rules

`report.CandidateReport` is the product artifact.

It should answer:

```text
What was compared?
What ran successfully?
What failed?
Which metrics improved?
Which metrics regressed?
Should the candidate promote, be reviewed, or be rejected?
```

Do not put runtime-only, executable-only, or source-bearing fields into reports.

## Logging rules

Searchbench uses `internal/logging` with Zap sugar and event-shaped helpers.

Use event helpers:

```go
logger.RunStarted(role, spec)
logger.RunExecuted(role, executed)
logger.RunScored(role, executed, scores)
logger.RunFailed(role, failure)
logger.ReportCreated(report)
```

Avoid raw calls like:

```go
logger.Sugar().Infow("thing", "system", systemSpec)
```

Development logs should be pretty and human-readable. JSON logs should remain structured and machine-readable.

## Console and CLI rules

`internal/console` renders structured artifacts for humans.

`internal/cli` parses user intent and wires commands.

The CLI should stay small and opinionated. Do not expose every internal concept as a command or flag.

Prefer commands like:

```text
searchbench demo-report
searchbench report <path>
searchbench compare --config searchbench.yaml
```

Avoid large command trees unless requested.

## Backend implementation rules

The backend boundary is:

```go
type Backend interface {
    StartSession(ctx context.Context, spec SessionSpec) (Session, error)
}

type Session interface {
    Tools(ctx context.Context) ([]ToolSpec, error)
    CallTool(ctx context.Context, name string, args json.RawMessage) (ToolResult, error)
    Close(ctx context.Context) error
}
```

Rules:

```text
Backend starts isolated sessions.
Session represents one isolated run/session.
Session is not assumed safe for concurrent tool calls.
Backend must be safe for concurrent StartSession if used with parallel Runner execution.
Concrete SDK weirdness stays inside backend adapters.
```

Do not leak SDK-specific types into `compare`, `report`, `score`, or `domain`.

## Dynamic boundary rules

External APIs are allowed to be ugly. Core packages are not.

When integrating external systems:

```text
provider SDK object -> adapter -> typed Searchbench object
```

Do not push raw SDK objects inward.

Examples:

```text
OpenAI/Cerebras usage -> typed adapter DTO -> domain.UsageSummary
model output          -> typed adapter DTO/finalizer -> domain.Prediction
Langfuse span         -> telemetry adapter interface
IC tool result        -> backend.ToolResult -> typed finalizer
```

## Testing rules

Place tests near the package that owns the invariant.

Examples:

```text
score invariants       -> internal/score
report safety          -> internal/report
console output safety  -> internal/console
logging safety         -> internal/logging
runner orchestration   -> internal/compare
CLI smoke tests        -> internal/cli
```

Do not test everything through the CLI.

Avoid brittle golden tests for ANSI output. Prefer substring and invariant checks:

```text
contains high-level section names
does not contain policy source
does not contain "source"
contains decision
contains expected metric keys
```

## Agent review checklist

Before finishing a change, verify:

```text
1. Did I use existing Searchbench nouns instead of creating a parallel model?
2. Did I avoid map[string]any entirely unless it is truly unavoidable and adapter-local?
3. Did I avoid raw SystemSpec/PolicyArtifact/TaskOracle/report dumps?
4. Did I keep compare free of concrete backend/provider logic?
5. Did I keep report as data and console as rendering?
6. Did I avoid package globals?
7. Did I avoid unnecessary interfaces?
8. Did I avoid generic pipeline abstractions?
9. Did I preserve deterministic report ordering?
10. Did I add tests near the invariant I changed?
```

## Required commands

Before handing off, run:

```bash
gofmt -w .
go test ./...
go mod tidy
```

If CLI behavior changed, also run the relevant command manually, for example:

```bash
go run ./cmd/searchbench --quiet --no-color demo-report
go run ./cmd/searchbench --quiet demo-report --output json
```

## Guiding principle

This codebase exists because the Python version became too flexible.

When in doubt, choose:

```text
explicit types
small packages
boring constructors
clear interfaces
safe report views
deterministic orchestration
```

Do not be clever. Make the model hard to misuse.
