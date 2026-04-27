You are implementing GitHub issue Decouple evaluator execution from CLI validation pipeline.

Before doing anything else, you MUST read the issue directly using the GitHub CLI:

gh issue view #7 --comments

Do not rely only on this prompt, local notes, or memory. The GitHub issue body is the implementation contract. If anything in this prompt conflicts with the issue, stop and report the conflict.

After reading the issue, also read:

- AGENTS.md
- docs/engineering/agentic-development-flow.md
- docs/engineering/issue-style-guide.md
- docs/engineering/model-testing.md
- docs/engineering/pure-center.md

Treat the issue body as the source of truth.

---

GOAL

Decouple evaluator execution from the CLI validation pipeline.

The evaluator must NOT run any subprocesses such as:

- templ generate
- gofmt
- go test
- ruff
- pyright
- pytest

The evaluator must not execute pipeline steps at all.

Pipeline code must remain in internal/pipeline for future writer/candidate validation flows.

---

CORE ARCHITECTURAL RULE

Evaluator owns ONLY:

- prompt rendering
- Eino model/tool execution
- strict prediction finalization
- evaluator-local retries
- returning typed evaluator results

Evaluator must NOT own:

- command execution
- validation pipeline
- pipeline classification
- bounded pipeline feedback
- writer repair logic

Preserve the pipeline package, but completely remove its use from internal/executor/eino.

---

EXPECTED IMPLEMENTATION SHAPE

Inspect and update:

- internal/executor/eino/evaluator.go
- internal/executor/eino/retry.go
- internal/executor/eino/errors.go
- internal/executor/eino/phases.go
- internal/executor/eino/evaluator_test.go
- internal/executor/eino/retry_test.go
- internal/executor/eino/pipeline_integration_test.go
- internal/pipeline/doc.go
- internal/pipeline/pipeline_test.go

REMOVE or DECOUPLE:

- Config.CommandRunner
- Config.PipelineSteps
- evaluator pipeline preflight
- evaluator phase "run_pipeline"
- evaluator phase "classify_pipeline"
- pipeline failure handling inside evaluator
- retry logic tied to pipeline
- tests that assume pipeline failure blocks model execution
- evaluator result fields that only exist for pipeline behavior (unless strictly required temporarily)

Prefer deletion over deprecation when reasonable.

---

REQUIRED EVALUATOR FLOW

The evaluator lifecycle must look like:

Run(ctx, task)
  create evaluator result

  for attempt := 0; attempt < maxAttempts; attempt++:
      phase: render_prompt
      phase: run_evaluator
      phase: finalize_prediction
      phase: complete

  phase: exhausted

There must be NO:

- run_pipeline phase
- classify_pipeline phase

---

RETRY REQUIREMENTS

Retries must ONLY cover evaluator-owned failures:

- recoverable model errors
- recoverable tool failures (if enabled)
- finalization failures
- empty or invalid predictions

Do NOT retry anything related to pipeline behavior.

RetryPolicy must not reference pipeline logic.

---

PIPELINE PACKAGE REQUIREMENTS

Keep internal/pipeline intact.

Preserve:

- CommandSpec
- StepResult
- command allowlisting
- ExecCommandRunner
- Classification
- FormatPipelineFeedback
- pipeline tests

Update package documentation so that pipeline is clearly NOT evaluator infrastructure.

Use wording like:

Package pipeline owns small typed local validation pipelines.

It provides allowlisted CLI step execution, step result capture, failure classification,
and bounded feedback formatting for writer, repair, and candidate validation flows.

It does NOT own evaluator execution, model calls, prompt rendering, backend runtimes,
or the public CLI surface.

---

TEST REQUIREMENTS

Evaluator tests must prove:

1. evaluator does NOT require a command runner
2. evaluator Run does NOT execute pipeline steps
3. evaluator phases do NOT include "run_pipeline"
4. evaluator phases do NOT include "classify_pipeline"
5. retry still works for malformed outputs
6. retry still works for empty predictions
7. tool failure behavior still works
8. prompt render failure is NOT retried
9. default tests remain offline and deterministic

If any command-runner field temporarily remains:

- add a regression test with a panic/failing runner
- prove it is NEVER called

Pipeline tests must remain in internal/pipeline and continue to verify:

- StepResult recording
- disallowed commands fail closed
- classification works (generation / format / test / type / infrastructure)
- feedback formatting is deterministic and bounded
- ExecCommandRunner records duration

---

NON-GOALS

Do NOT implement:

- writer agent
- repair loop
- candidate generation
- policy generation or installation
- MCP
- Iterative Context
- jCodeMunch
- graph scoring
- LangSmith
- repo materialization
- new CLI product surface
- pipeline scheduling systems
- state machine frameworks
- real model calls

This is strictly a decoupling task.

---

VALIDATION

Run:

go test ./...

Also run:

go test ./internal/executor/eino ./internal/pipeline

All tests must pass.

All tests must remain offline and deterministic.

---

FINAL RESPONSE FORMAT

Summarize:

- files changed
- how evaluator/pipeline coupling was removed
- tests added or updated
- commands run
- any deviations from the issue contract
