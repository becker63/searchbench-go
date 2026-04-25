# Searchbench-Go implementation roadmap

## Principle

Keep the existing core model stable.

Core packages:

```text
internal/domain
internal/run
internal/score
internal/report
internal/compare
internal/backend
internal/codegraph
```

Implementation packages should consume these core packages. They should not redefine their own task, run, score, report, backend, or prediction models.

Do not create a second Go module. Keep implementation code under `internal/`.

---

## Phase 1: Concrete executor boundary

Add:

```text
internal/executor/
  doc.go
  executor.go
  options.go
  result.go
```

Purpose:

```text
run.Spec
  -> backend session
  -> model/tool loop
  -> domain.Prediction
  -> domain.UsageSummary
  -> run.ExecutedRun
```

Rules:

- `executor` implements `compare.Executor`.
- `compare` must not know concrete backend/session/model details.
- Executor should own one run of one system on one task.
- Executor should not own scoring, reporting, or promotion decisions.

---

## Phase 2: Backend implementations

Keep the current abstract backend package:

```text
internal/backend/
  backend.go
  session.go
  errors.go
```

Add implementation subpackages:

```text
internal/backend/iterativecontext/
  backend.go
  session.go
  tools.go
  policy.go

internal/backend/jcodemunch/
  backend.go
  session.go
  tools.go
```

Purpose:

- Implement `backend.Backend`.
- Implement `backend.Session`.
- Hide concrete SDK/subprocess/Python/MCP details.
- Expose only `Tools`, `CallTool`, and `Close`.

Rules:

- No concrete backend types in `compare`.
- No SDK-specific types in `domain`, `run`, `score`, or `report`.
- Backend sessions must be isolated per run.
- Backend implementations must be safe for concurrent `StartSession` if used with `Runner.Parallelism`.

---

## Phase 3: Model/provider client

Add:

```text
internal/model/
  doc.go
  client.go
  message.go
  response.go
  usage.go

internal/model/openai/
  client.go
  dto.go

internal/model/cerebras/
  client.go
  dto.go
```

Purpose:

- Own LLM provider calls.
- Normalize provider responses.
- Normalize token usage into `domain.UsageSummary`.

Rules:

- Prefer typed DTOs.
- Avoid `map[string]any`.
- Provider weirdness stays in provider packages.
- Core packages receive typed values only.

---

## Phase 4: Prompt construction

Add:

```text
internal/prompt/
  doc.go
  builder.go
  templates.go
  task.go
```

Purpose:

- Convert `domain.TaskSpec` and `domain.SystemSpec` into model input.
- Keep scorer-only oracle data out of prompts.
- Load/render system prompts for IC and jCodeMunch-style runs.

Rules:

- Prompt builders may receive `TaskSpec`, but must only use prompt-safe fields.
- Do not leak `TaskOracle`.
- Do not leak policy source except where explicitly required by the backend/session setup.

---

## Phase 5: Finalization / prediction extraction

Add:

```text
internal/finalizer/
  doc.go
  prediction.go
  files.go
  errors.go
```

Purpose:

- Convert model responses/tool observations into `domain.Prediction`.
- Normalize repo-relative file paths.
- Reject malformed or unsafe predictions.

Rules:

- Finalizer returns typed `domain.Prediction`.
- Do not store raw model output on `run.ExecutedRun`.
- Raw/dynamic payload handling must remain adapter-local.

---

## Phase 6: Static graph ingestion

Current graph model:

```text
internal/codegraph/
```

Add ingestion package:

```text
internal/graphing/
  doc.go
  ingest.go
  treesitter.go
  normalize.go
  anchors.go
```

Alternative name:

```text
internal/codegraph/ingest/
```

Purpose:

- Build `codegraph.Graph` from repository snapshots.
- Use tree-sitter bindings.
- Normalize files/functions/symbols into graph nodes/edges.

Rules:

- Ingestion depends on `codegraph.Builder`.
- Scoring depends on `codegraph.Graph`.
- Keep `gograph` hidden behind `codegraph.Store`.
- Treat built graphs as read-only.

---

## Phase 7: Real scoring engine

Current scoring model:

```text
internal/score/
```

Add concrete scorer package:

```text
internal/scoring/
  doc.go
  engine.go
  gold_hop.go
  issue_hop.go
  token_efficiency.go
  cost.go
  composite.go
```

Purpose:

- Implement `compare.Scorer`.
- Compute required `score.ScoreSet`.
- Use `codegraph.Graph`, `run.ExecutedRun`, and task oracle.

Rules:

- If any required metric cannot be computed, scoring fails.
- Do not add `Available bool`.
- Do not silently drop metrics.
- Keep metric direction centralized in `score`.

---

## Phase 8: Promotion rules

Current report output:

```text
report.PromotionDecision
```

Add:

```text
internal/promotion/
  doc.go
  policy.go
  decider.go
  rules.go
```

Purpose:

- Implement `compare.Decider`.
- Decide `PROMOTE`, `REVIEW`, or `REJECT`.
- Encode thresholds for regressions, failures, cost, and primary score changes.

Rules:

- `promotion` produces `report.PromotionDecision`.
- `report` keeps the report-facing decision type.
- Do not put a large rule engine into `compare`.

---

## Phase 9: Writer feedback projection

Add:

```text
internal/writerfeedback/
  doc.go
  feedback.go
  builder.go
  summary.go
```

Purpose:

- Convert `report.CandidateReport` into compact writer-agent input.
- Summarize metric deltas, regressions, failures, strong cases, weak cases, and guidance.

Rules:

- Writer feedback is a projection of `CandidateReport`.
- It should not re-run scoring.
- It should not depend on score history by default.
- Do not resurrect old score-over-time reducers unless proven useful.

---

## Phase 10: Writer agent

Add:

```text
internal/writer/
  doc.go
  writer.go
  prompt.go
  candidate.go
```

Purpose:

- Generate a new candidate policy/system from writer feedback.
- Produce a new `domain.PolicyArtifact` or candidate `domain.SystemSpec`.

Rules:

- Writer consumes `writerfeedback.Feedback`.
- Writer does not consume raw traces/logs as its primary interface.
- Writer output must validate as a `PolicyArtifact`.
- Candidate identity must flow through `SystemSpec -> SystemRef`.

---

## Phase 11: Dataset adapters

Add:

```text
internal/dataset/
  doc.go
  dataset.go
  loader.go

internal/dataset/lca/
  loader.go
  dto.go
  normalize.go
```

Purpose:

- Load benchmark tasks.
- Convert external dataset rows into `domain.TaskSpec`.

Rules:

- Dataset-specific DTOs stay in dataset packages.
- Core packages should only see `domain.TaskSpec`.
- Preserve oracle/prompt split.

---

## Phase 12: Artifact storage

Add:

```text
internal/artifact/
  doc.go
  store.go
  filesystem.go
```

Purpose:

- Write/read candidate reports, logs, policy artifacts, graph snapshots, and run outputs.
- Support CLI output paths later.

Rules:

- Artifact storage should not change report structure.
- Reports remain JSON-serializable `report.CandidateReport`.
- Do not put filesystem behavior into `report`.

---

## Phase 13: Telemetry integration

Add later, not first:

```text
internal/telemetry/
  doc.go
  observer.go
  events.go

internal/telemetry/langfuse/
  client.go
  dto.go
  observer.go
```

Purpose:

- Observe runs/reports.
- Emit spans/scores/events to Langfuse or another telemetry backend.

Rules:

- Telemetry observes the model.
- Telemetry must not define the model.
- No Langfuse SDK types in `compare`, `report`, `score`, `domain`, or `run`.
- No telemetry-driven control flow.

---

## Phase 14: Real CLI commands

Current CLI:

```text
searchbench demo-report
```

Add later:

```text
searchbench compare --config searchbench.yaml --output report.json
searchbench report <report.json>
searchbench report <report.json> --output json
```

Rules:

- CLI stays thin.
- CLI parses user intent and wires dependencies.
- CLI does not own execution, scoring, reporting, or backend logic.
- Keep command surface small and opinionated.

---

## Phase 15: Config

Add:

```text
internal/config/
  doc.go
  config.go
  load.go
```

Purpose:

- Load comparison configs.
- Build `compare.Plan`.
- Select backend/model/scoring/promotion implementations.

Rules:

- Config structs are typed.
- Avoid dynamic maps.
- Config should construct core values, not replace the core model.

---

## Do not port directly from Python

Translate old Searchbench concepts into the Go model.

Old Python concern -> Go destination:

```text
agents/localizer.py        -> executor + prompt + model + finalizer + backend
backends/ic.py             -> backend/iterativecontext
backends/jc.py             -> backend/jcodemunch
prompts/                   -> prompt
localization/static_graph  -> graphing + codegraph
scoring/                   -> scoring + score + promotion
telemetry/                 -> telemetry/langfuse
pipeline/orchestration     -> compare + writerfeedback + writer
datasets/HF/Langfuse       -> dataset adapters
```

Avoid recreating the Python mega-files.
