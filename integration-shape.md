# Future integrations

Searchbench-Go should keep a strict separation between the pure Searchbench model, Searchbench-native implementation packages, narrow external adapters, and user/developer-facing surfaces.

The goal is not maximum theoretical pluggability. The goal is to keep the core model pure while making it obvious where real work happens.

```text
Core is pure.
Implementation does Searchbench work.
Adapters isolate backend/storage integrations.
Surface exposes the system.
App wiring composes everything.
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

## Four layers

Future work should be organized around four layers.

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

### 2. Searchbench-native implementation

Implementation packages do real Searchbench work, but still primarily speak Searchbench types.

Examples:

```text
internal/implementation/executor
internal/implementation/scoring
internal/implementation/graphing
internal/implementation/prompt
internal/implementation/finalizer
internal/implementation/promotion
internal/implementation/writerfeedback
internal/implementation/dataset
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
```

These packages are allowed to do real Searchbench work, but they should not become dumping grounds for backend SDKs, storage details, or external platform types.

### 3. Narrow adapters

Adapters isolate external runtime/storage boundaries that we intentionally want visually quarantined.

Adapters should live under:

```text
internal/adapters
```

Current intended adapters:

```text
internal/adapters/backend/iterativecontext
internal/adapters/backend/jcodemunch
internal/adapters/artifact/filesystem
```

Adapters may be messy. Their job is to translate external APIs, subprocesses, protocols, and persistence mechanisms into typed Searchbench values.

Adapters may import core packages.

Core packages must not import adapters.

Do not put every external library under adapters by default. Only use adapters when the package’s main job is to isolate an external system boundary.

### 4. Surface

Surface packages expose Searchbench to humans and developers.

Examples:

```text
internal/surface/logging
internal/surface/console
internal/surface/cli
```

Surface packages should not define the model. They should present, render, log, or command the system.

Examples:

```text
logging  structured/development event output
console  human terminal rendering for CandidateReport
cli      small command-line interface
```

Surface code may consume core/report values, but it should not become the owner of orchestration, scoring, backend execution, or artifact structure.

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
backend adapters
scoring implementation
promotion decider
dataset loader
artifact store
logging
console/CLI options
```

The CLI should call the app layer rather than manually constructing every concrete implementation.

The import direction should look like:

```text
cmd/searchbench -> internal/surface/cli
surface/cli     -> internal/app, internal/surface/console, internal/surface/logging
internal/app    -> core packages + implementation packages + adapters + surface dependencies
implementation  -> core packages
adapters        -> core packages
surface         -> core/report values
core packages   -> never adapters, never app
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

This structure should make it visually obvious which code is pure model code, which code performs Searchbench-native work, which code isolates external systems, and which code exposes the system.

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

Rules:

```text
Backend starts isolated sessions.
Session represents one isolated run/session.
Session is not assumed safe for concurrent tool calls.
Backend must be safe for concurrent StartSession if used with parallel Runner execution.
Concrete SDK/runtime/subprocess details stay inside adapters.
```

`compare` should never import `internal/adapters/backend/...`.

## Executor implementation

The executor layer should implement `compare.Executor`.

Suggested package:

```text
internal/implementation/executor
```

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
```

If Eino is used, keep Eino-specific details inside the executor implementation unless they become large enough to justify a subpackage.

Do not create a generic executor plugin system.

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

If OpenAI is only used through Eino, keep OpenAI configuration near the executor implementation.

The rule:

```text
Do not abstract over systems we are not using.
```

## LangSmith integration

LangSmith is a tracing/experiment integration, but do not create a generic telemetry framework unless there is a second real telemetry implementation.

If LangSmith integration becomes substantial, place it where it best reflects its role.

Preferred options:

```text
internal/implementation/observability/langsmith
```

or, if it is mostly artifact/report export:

```text
internal/implementation/dataset/langsmith
```

or, if you intentionally want to quarantine it as an external platform boundary:

```text
internal/adapters/observability/langsmith
```

Do not add it under adapters automatically just because it uses an SDK. Use adapters only if the package is mainly isolating external platform behavior.

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

## Dataset integration

Dataset loading should usually be Searchbench-native implementation work.

Use:

```text
internal/implementation/dataset
```

LCA-specific loading can live under:

```text
internal/implementation/dataset/lca
```

The output should be typed Searchbench tasks:

```text
[]domain.TaskSpec
```

or:

```text
domain.NonEmpty[domain.TaskSpec]
```

Do not let external dataset row structs flow into `compare`, `run`, `score`, or `report`.

If a dataset integration is mostly remote-platform synchronization rather than local Searchbench task construction, then consider an adapter package.

## Tree-sitter graphing

The core graph package remains:

```text
internal/codegraph
```

Tree-sitter-specific ingestion should not live directly in `codegraph`.

Preferred split:

```text
internal/codegraph
  graph abstraction and store

internal/implementation/graphing
  Searchbench-native ingestion orchestration

internal/implementation/graphing/treesitter
  tree-sitter parser integration
```

Tree-sitter is a library used by Searchbench-native graphing. It does not need to be under adapters unless the integration becomes a large external boundary.

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
internal/implementation/artifact
```

for Searchbench artifact concepts if needed, and:

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
internal/implementation/promotion
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
internal/implementation/writerfeedback
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

## Surface packages

Surface packages should move under:

```text
internal/surface
```

Recommended locations:

```text
internal/surface/logging
internal/surface/console
internal/surface/cli
```

Responsibilities:

```text
logging  structured/development event output
console  human CandidateReport rendering
cli      small command-line command surface
```

Surface packages may depend on core/report values.

Surface packages must not define core model types.

Surface packages must not own concrete backend execution, scoring algorithms, promotion policies, graph ingestion, or artifact storage.

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
implementation -> core
adapters -> core
app -> core + implementation + adapters + surface
surface -> core/report values
cli surface -> app
```

Forbidden:

```text
core -> adapters
core -> app
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

Do not create parallel report, score, task, or run models inside adapters or implementation packages.

Do not put every package that imports a third-party dependency under adapters.

## Guiding principle

The Python version taught that AI systems need replaceable external integrations.

The Go version should express that lesson without making the core vague.

```text
External systems change.
Searchbench vocabulary should not.
```

Keep the unstable parts isolated.

Keep the model pure.
