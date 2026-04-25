# Future integrations

Searchbench-Go should keep a strict separation between the pure Searchbench model, Searchbench-native implementation packages, and external-system adapters.

The goal is not maximum theoretical pluggability. The goal is to keep the core model pure while making it obvious where real external work happens.

```text
Core is pure.
Implementations do Searchbench work.
Adapters touch the outside world.
App wiring composes them.
```

## Current core

The following packages form the pure Searchbench spine:

```text
internal/domain
internal/run
internal/score
internal/report
internal/compare
internal/backend
internal/codegraph
```

These packages define Searchbench vocabulary, lifecycle, scoring semantics, report artifacts, orchestration contracts, backend/session ports, and graph abstractions.

They must not import concrete external SDKs or runtime systems such as:

```text
OpenAI SDKs
Eino
LangSmith
tree-sitter
Iterative Context subprocess/runtime code
jCodeMunch runtime code
filesystem artifact stores
```

The core model should remain boring, typed, deterministic, and difficult to misuse.

## Three layers

Future work should be organized around three layers.

### 1. Pure core

Pure core packages define the stable model and ports.

Examples:

```text
internal/domain
internal/run
internal/score
internal/report
internal/compare
internal/backend
internal/codegraph
```

Core packages may define interfaces and value types. They should not know about concrete implementations.

For example:

```text
internal/backend
```

means the backend/session protocol, not all backend implementations.

### 2. Searchbench-native implementations

Implementation packages do real Searchbench work, but still primarily speak Searchbench types.

Examples:

```text
internal/executor
internal/scoring
internal/graphing
internal/prompt
internal/finalizer
internal/promotion
internal/writerfeedback
internal/dataset
internal/artifact
```

These packages may implement real algorithms or workflows:

```text
executor       run.Spec -> run.ExecutedRun
scoring        score.Input -> score.ScoreSet
graphing       repository snapshot -> codegraph.Graph
prompt         TaskSpec/SystemSpec -> prompt-safe model input
finalizer      model/tool observations -> domain.Prediction
promotion      comparisons/regressions -> report.PromotionDecision
writerfeedback CandidateReport -> writer steering input
dataset        external/internal rows -> domain.TaskSpec
artifact       reports/policies/graphs -> persisted artifacts
```

These packages are allowed to do real work, but should not become dumping grounds for SDK-specific details.

### 3. External adapters

Adapters touch external systems.

Adapters should live under:

```text
internal/adapters
```

Examples:

```text
internal/adapters/backend/iterativecontext
internal/adapters/backend/jcodemunch
internal/adapters/executor/eino
internal/adapters/telemetry/langsmith
internal/adapters/dataset/langsmith
internal/adapters/dataset/lca
internal/adapters/graphing/treesitter
internal/adapters/artifact/filesystem
```

Adapters may be messy. Their job is to translate external APIs, SDKs, protocols, subprocesses, and file formats into typed Searchbench values.

Adapters may import core packages.

Core packages must not import adapters.

## App wiring

Concrete composition should live in:

```text
internal/app
```

Example:

```text
internal/app/
  wiring.go
  compare.go
  demo.go
```

The app layer wires together:

```text
compare.Runner
executor implementation
backend implementations
scoring implementation
promotion decider
dataset loader
artifact store
logging
console/CLI options
LangSmith adapter
```

The CLI should call the app layer rather than manually constructing every concrete implementation.

The import direction should look like:

```text
cmd/searchbench -> internal/cli
internal/cli    -> internal/app, internal/console, internal/logging
internal/app    -> core packages + implementation packages + adapters
adapters        -> core packages
core packages   -> never adapters
```

## Recommended future tree

A future mature tree may look like:

```text
internal/
  domain/
  run/
  score/
  report/
  compare/
  backend/
  codegraph/

  surface/
    logging/
    console/
    cli/

  implementation/
    executor/
    scoring/
    graphing/
    prompt/
    finalizer/
    promotion/
    writerfeedback/
    dataset/

  adapters/
    backend/
      iterativecontext/
      jcodemunch/
    artifact/
      filesystem/

  app/
    wiring.go
    compare.go
    demo.go
```

This structure should make it visually obvious which code is pure model code and which code talks to the external world.

## Backend integrations

The core backend port remains:

```text
internal/backend
```

Concrete backend integrations should live under:

```text
internal/adapters/backend
```

Examples:

```text
internal/adapters/backend/iterativecontext
internal/adapters/backend/jcodemunch
```

Backend adapters implement:

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
Concrete SDK/runtime/subprocess details stay inside adapters.
```

`compare` should never import `internal/adapters/backend/...`.

## Executor integration

The executor layer should implement `compare.Executor`.

Suggested package:

```text
internal/executor
```

If Eino becomes the dominant execution runtime, Eino-specific code can live under:

```text
internal/adapters/executor/eino
```

or, if it is tightly bound to the executor implementation:

```text
internal/executor/eino
```

Prefer the adapter path when Eino-specific types, callbacks, graph compilation, or SDK details are prominent.

Executor responsibility:

```text
run.Spec
  -> backend session / Eino flow / model-tool loop
  -> domain.Prediction
  -> domain.UsageSummary
  -> run.ExecutedRun
```

Executor should not own:

```text
score computation
promotion decisions
report rendering
dataset loading
artifact storage
telemetry platform semantics
```

## OpenAI usage

Do not add a generic AI model provider abstraction unless a second provider is actually required.

Searchbench currently assumes OpenAI usage through the concrete execution path.

Avoid premature packages like:

```text
internal/model/openai
internal/model/cerebras
internal/model/ollama
```

unless the code genuinely needs an independently reusable model client layer.

If OpenAI is only used through Eino, keep OpenAI configuration near the Eino executor adapter.

The rule:

```text
Do not abstract over systems we are not using.
```

## LangSmith integration

LangSmith is an adapter concern.

Recommended location:

```text
internal/adapters/telemetry/langsmith
```

LangSmith should consume Searchbench artifacts. It should not define them.

LangSmith may receive:

```text
report.CandidateReport
run.ExecutedRun
run.RunFailure
score.ScoreSet
report.ScoreComparison
report.Regression
domain.TaskSpec
domain.SystemRef
```

LangSmith should be used for:

```text
trace storage
experiment visibility
dataset/example linkage
feedback/score display
human review surfaces
```

Searchbench remains the authoritative evaluator.

Do not make LangSmith evaluators the source of truth for:

```text
gold_hop
issue_hop
token_efficiency
cost
composite
regressions
promotion decisions
CandidateReport
writer feedback
```

Those remain Searchbench-owned.

## LangSmith datasets

Dataset mirroring can live under:

```text
internal/adapters/dataset/langsmith
```

It should map between:

```text
domain.TaskSpec
```

and LangSmith dataset examples.

Rules:

```text
LangSmith dataset/example IDs are external refs.
They are not core Searchbench identity.
Dataset DTOs stay inside the adapter.
Core packages should only see domain.TaskSpec.
```

## LCA dataset integration

LCA-specific loading can live under:

```text
internal/adapters/dataset/lca
```

or if it becomes Searchbench-native enough:

```text
internal/dataset/lca
```

Use the adapter path if the package mostly talks to Hugging Face, JSONL, remote APIs, or external file formats.

Use the native dataset path if it mainly transforms already-local data into `domain.TaskSpec`.

In either case, the output should be typed Searchbench tasks:

```text
[]domain.TaskSpec
```

or:

```text
domain.NonEmpty[domain.TaskSpec]
```

Do not let external dataset row structs flow into `compare`, `run`, `score`, or `report`.

## Tree-sitter graphing

The core graph package remains:

```text
internal/codegraph
```

Tree-sitter-specific ingestion should not live directly in `codegraph`.

Use:

```text
internal/adapters/graphing/treesitter
```

or:

```text
internal/graphing/treesitter
```

Preferred split:

```text
internal/codegraph
  graph abstraction and store

internal/graphing
  Searchbench-native ingestion orchestration

internal/adapters/graphing/treesitter
  tree-sitter-specific parser integration
```

Rules:

```text
tree-sitter code writes into codegraph.Builder
scoring consumes codegraph.Graph
gograph remains hidden behind codegraph.Store
built graphs should be treated as read-only
```

## Artifact storage

Artifact storage should not live in `report`.

Use:

```text
internal/artifact
```

for Searchbench artifact concepts, and:

```text
internal/adapters/artifact/filesystem
```

for filesystem persistence.

Artifacts may include:

```text
CandidateReport JSON
policy artifacts
graph snapshots
trace links
writer feedback snapshots
run outputs
```

Reports should remain plain structured data.

## Promotion rules

Promotion logic should be Searchbench-native, not an adapter.

Use:

```text
internal/promotion
```

It should implement `compare.Decider` and return `report.PromotionDecision`.

Promotion rules may consider:

```text
primary metric improvement
regression count
protected task regressions
candidate failures
cost increases
token efficiency
composite score
```

Do not put promotion rule complexity directly in `compare`.

## Writer feedback

Writer feedback should be a projection of `report.CandidateReport`.

Use:

```text
internal/writerfeedback
```

It should produce a compact steering object for the writer agent:

```text
decision
metric deltas
regressions
failures
strong cases
weak cases
guidance
```

Writer feedback should not re-run scoring.

Writer feedback should not depend on score history by default.

## Dynamic data rules

Avoid dynamic data structures even at adapter boundaries.

Preferred order:

```text
1. Typed request/response structs
2. Small typed adapter DTOs
3. json.RawMessage at the narrow transport edge
4. map[string]any only when the external API truly has no stable shape
```

If an external API is dynamic, quarantine that dynamism immediately:

```text
external SDK object
  -> adapter-local DTO/parser
  -> typed Searchbench value
```

Do not let dynamic payloads flow inward.

## Import rules

Allowed:

```text
adapters -> core
implementations -> core
app -> adapters + implementations + core
cli -> app + console + logging
```

Forbidden:

```text
core -> adapters
domain -> high-level packages
report -> console/logging/adapters
score -> adapters
compare -> concrete backend implementations
backend -> concrete backend implementations
```

## What not to do

Do not create a giant generic plugin system.

Do not create a generic model-provider abstraction just because the AI ecosystem changes quickly.

Do not create a generic telemetry framework unless there is a second real telemetry implementation.

Do not move SDK-specific types into core packages.

Do not let `compare` become the new mega-localizer.

Do not let adapter DTOs become Searchbench model types.

Do not create parallel report, score, task, or run models inside adapters.

## Guiding principle

The Python version taught that AI systems need replaceable external integrations.

The Go version should express that lesson without making the core vague.

```text
External systems change.
Searchbench vocabulary should not.
```

Keep the unstable parts isolated.

Keep the model pure.
