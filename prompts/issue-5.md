You are working in the searchbench-go repository.

The target GitHub issue is:

    Add retry-aware Eino evaluator execution and CLI pipeline validation

Before making any changes, use the GitHub CLI to find and read the full target issue.

Run:

    gh issue list --state open --search "Add retry-aware Eino evaluator execution and CLI pipeline validation"
    gh issue view <ISSUE_NUMBER> --comments --json title,body,comments,labels,state,url

If multiple issues match, choose the open issue whose title exactly matches:

    Add retry-aware Eino evaluator execution and CLI pipeline validation

If no issue matches, stop and report that the target issue could not be found. Do not implement from memory.

Treat the GitHub issue body as the source of truth for the implementation contract.

If this prompt conflicts with the GitHub issue, prefer the GitHub issue unless it clearly violates AGENTS.md/agents.md or the repository architecture docs. In that case, stop and explain the conflict before editing.

Before editing, read the project guidance and architecture docs:

1. AGENTS.md or agents.md, whichever exists
2. docs/engineering/issue-style-guide.md
3. docs/engineering/agentic-development-flow.md
4. docs/engineering/model-testing.md
5. integration-shape.md
6. todo.md
7. Existing package docs, especially:
   - internal/domain/doc.go
   - internal/run/doc.go
   - internal/compare/doc.go
   - internal/backend/doc.go
   - internal/score/doc.go
   - internal/executor/eino/doc.go
   - internal/prompts/evaluator/
   - internal/testing/doc.go
   - internal/testing/modeltest/doc.go
   - internal/surface/cli/doc.go
   - internal/surface/logging/doc.go

Also inspect the existing implementation from prior issues:

- internal/testing/modeltest/
- internal/domain/localization.go
- internal/domain/prediction.go
- internal/domain/task.go
- internal/domain/repo.go
- internal/run/phases.go
- internal/run/failure.go
- internal/run/record.go
- internal/executor/eino/evaluator.go
- internal/executor/eino/evaluator_test.go
- internal/executor/eino/finalizer.go
- internal/executor/eino/phases.go
- internal/executor/eino/errors.go
- internal/prompts/evaluator/

Do not start editing until you understand the existing package boundaries.

Task:
Implement the issue “Add retry-aware Eino evaluator execution and CLI pipeline validation”.

Goal:
Extend the existing minimal Eino evaluator loop with bounded retry machinery and a typed CLI validation pipeline.

This issue proves that SearchBench-Go can:

- retry evaluator execution on recoverable model/finalization/tool failures
- run local CLI validation steps before evaluator execution
- represent external command results as typed step results
- classify CLI failures into structured categories
- surface pipeline failures as typed evaluator failures
- test CLI failure behavior without real model calls
- preserve old Python SearchBench pipeline/repair semantics without porting the old Python state machine

Important: this is not the full writer/optimization loop.

This issue creates the reusable evaluator-side retry and CLI pipeline substrate that later writer/repair issues will consume.

Dependency context:
The deterministic model fixture testing issue is already done.

Use:

    internal/testing/modeltest.ScriptedModel

for evaluator retry tests.

Do not create a second model fixture system.

The LCA/domain schema issue is already done.

Use existing domain/run types where appropriate.

Do not reimplement:

- domain.LCATask
- domain.LCATaskIdentity
- domain.LCAContext
- domain.LCAGold
- domain.LCAHFRow
- domain.Prediction
- domain.CanonicalizePath
- domain.CanonicalizePaths

The minimal Eino evaluator loop is already done.

Build on the existing package:

    internal/executor/eino

Do not rewrite the minimal evaluator.

Do not replace Eino.

Do not introduce a custom agent framework.

Architecture rule:
SearchBench-Go owns:

- phase/result typing
- retry policy
- pipeline step definitions
- subprocess execution boundaries
- pipeline classification
- failure feedback construction
- evaluator result/failure envelopes

Eino owns:

- model/tool execution
- tool-call loop behavior
- model response handling

The new pipeline package should own:

- local CLI command execution
- command allowlisting
- step results
- classification of external command failures
- bounded feedback formatting

The modeltest package owns:

- scripted model responses
- deterministic model test behavior

Do not introduce:

- custom state machine library
- custom ReAct loop
- MCP-specific code
- repo materialization
- policy artifacts
- policy injection
- writer agent
- optimization loop
- graph-distance scoring
- LangSmith tracing
- live model calls in default tests
- broad public CLI product surface

Preserve from Python SearchBench:

- external validation steps return typed step results
- step results include name, command, exit status, stdout, stderr, duration, and pass/fail status
- command execution is scoped and allowlisted
- pipeline classification distinguishes type errors, lint/format errors, test failures, generation failures, and infrastructure failures
- warnings on successful commands are not failures
- feedback is structured and truncated to a safe size
- retry attempts receive classified failure context
- exhausted retries return a typed failure with the last useful classification

Do not preserve:

- Python-specific commands as production defaults
- Pydantic model implementation details
- Python state machine framework
- Langfuse-specific tracing shape
- worktree mutation as policy selection
- Python dict-shaped data passing
- ad-hoc subprocess calls spread across the codebase

Package placement:
Prefer existing project structure.

If no better location exists, create:

    internal/pipeline/
      doc.go
      step.go
      runner.go
      classify.go
      feedback.go
      pipeline_test.go

Modify or add under:

    internal/executor/eino/
      retry.go
      retry_test.go
      pipeline_integration_test.go

Use existing files where better.

If internal/run already owns a concept cleanly, reuse it rather than duplicating it.

Do not place subprocess execution in internal/domain.

Do not place Eino-specific retry logic in internal/pipeline.

Do not place testing-only fake commands in production packages.

Do not place pipeline code under internal/surface/cli. This is a CLI/subprocess validation pipeline, not the public SearchBench command-line surface.

Pipeline step model:
Define or reuse a typed step result for CLI checks.

A useful shape is:

    StepResult
      name
      command
      cwd
      passed
      exit_code
      stdout
      stderr
      stdout_tail
      stderr_tail
      duration
      skipped

Use exact field names that fit Go conventions and the existing repo.

Rules:

- command should be recorded for review/debugging
- cwd should be recorded if useful
- stdout/stderr may be stored fully if small
- include tail/truncation helpers where useful
- failed commands should not panic
- subprocess launch failures should become infrastructure failures
- context cancellation should be respected

Command execution boundary:
Implement a small command runner abstraction so tests do not need real subprocesses.

Suggested shape:

    CommandRunner
      Run(ctx, CommandSpec) StepResult

    CommandSpec
      name
      command
      cwd
      timeout

Production runner:

- uses os/exec
- captures stdout/stderr
- respects context cancellation
- returns typed StepResult

Test runner:

- returns scripted StepResult values
- records commands
- can simulate command failure
- can simulate subprocess launch/infrastructure failure

Do not use shell strings by default.

Prefer:

    []string{"go", "test", "./..."}

over:

    "go test ./..."

Command allowlist:
The pipeline runner should only execute known commands.

Initial allowed commands should be narrow and repository-local.

Reasonable candidates:

    templ generate
    gofmt check
    go test ./...
    go vet ./... if already standard

Use the repo’s actual conventions.

If templ generation is required because prompt.templ exists, include a templ_generate step only if it is available and consistent with the project.

If templ is not available in the default test environment, tests should use scripted command runners and should not require global templ execution.

Do not add arbitrary command execution.

Unknown/disallowed commands should fail closed as infrastructure failures.

Initial evaluator validation pipeline:
Create a minimal evaluator validation pipeline.

Suggested steps:

    templ_generate
    gofmt_check
    go_test

Optional only if already standard:

    go_vet

Do not add slow, flaky, networked, or paid checks.

Do not add staticcheck unless the repo already uses it and it is available in the dev environment.

Pipeline classification:
Classify failed step results into structured categories.

Suggested shape:

    PipelineClassification
      generation_failures
      format_errors
      type_errors
      lint_errors
      test_failures
      infrastructure_failures
      passed_steps

Suggested mapping:

    templ_generate failure:
      generation_failures

    gofmt_check failure:
      format_errors

    go_test failure caused by compile/type errors:
      type_errors if detectable, otherwise test_failures

    go_test ordinary test failure:
      test_failures

    go_vet failure:
      lint_errors or type_errors depending on message

    command launch failure / timeout / disallowed command:
      infrastructure_failures

Do not overfit classification to brittle exact stderr text.

Use simple, stable heuristics.

Warnings with exit code 0 should be recorded but not classified as failures.

Structured feedback:
Add a helper that converts PipelineClassification into bounded, readable feedback.

Suggested helper:

    FormatPipelineFeedback(classification, maxChars int) string

Feedback should include deterministic sections like:

    ## GENERATION FAILURES
    ## FORMAT ERRORS
    ## TYPE ERRORS
    ## LINT ERRORS
    ## TEST FAILURES
    ## INFRASTRUCTURE FAILURES
    ## PASSED STEPS

Add an action-hint helper if useful:

    InferPipelineActionHint(classification) string

Suggested priority:

    infrastructure failures first
    generation failures
    format errors
    type errors
    lint errors
    test failures
    all passed

Keep this small. It is future-facing for writer/repair loops, but must be testable now.

Retry policy:
Add bounded retry support to the Eino evaluator runner.

A useful shape:

    RetryPolicy
      max_attempts
      retry_on_model_error
      retry_on_tool_failure
      retry_on_finalization_failure
      retry_on_invalid_prediction
      retry_on_pipeline_failure

Default behavior for this issue:

- retry finalization failure
- retry invalid/empty predictions
- retry recoverable model errors
- retry recoverable tool failures only if explicitly marked recoverable
- do not retry pipeline failure by default
- do not retry prompt render failure
- do not retry infrastructure failure by default

Default max attempts should be small:

    2 or 3

Use existing project conventions if already defined.

Retry feedback:
When retrying evaluator execution, the next model attempt should receive concise structured feedback about why the previous attempt failed.

Examples:

    previous attempt returned malformed JSON
    previous attempt returned empty predicted files
    previous tool call failed: fake_resolve
    previous model call failed: context length exceeded

For this issue, retry feedback may be appended to evaluator prompt input or passed through a typed retry context.

Do not mutate global prompt state.

Do not expose gold/oracle data in retry feedback.

Do not include full unbounded stderr/tool output.

Pipeline integration:
Run the evaluator validation pipeline as part of evaluator execution lifecycle, but keep it separate from model execution.

For this issue, run the validation pipeline before the model call.

Reason:

- if local evaluator code cannot compile or prompt generation is broken, do not spend model calls
- tests can prove pipeline failure prevents model execution
- later writer issues can reuse the same pipeline after candidate generation

This issue may expose a lower-level function so future writer/repair loops can run the pipeline after candidate artifacts are written.

Phase flow:
Use ordinary Go control flow, not a state machine library.

Expected flow:

    Run(ctx, task)
      create execution result accumulator

      phase: run_pipeline
          run evaluator validation pipeline
          collect StepResult[]
          if command launch failed:
              classify as infrastructure failure
          if any step failed:
              phase: classify_pipeline
                  classify step results
                  format pipeline feedback
              return pipeline_failed without calling model

      for attempt := 0; attempt < maxAttempts; attempt++ {
          phase: render_prompt
              build typed evaluator prompt input
              include retry feedback if attempt > 0
              render .templ XML-style prompt
              if render failed:
                  return prompt_render_failed

          phase: run_evaluator
              run Eino evaluator agent
              expose allowed fake/local tools
              if evaluator/model failed:
                  if recoverable and attempts remain:
                      phase: prepare_retry
                          record evaluator_failed feedback
                          continue
                  return evaluator_failed

              if tool failed:
                  if recoverable and attempts remain:
                      phase: prepare_retry
                          record tool_call_failed feedback
                          continue
                  return tool_call_failed

          phase: finalize_prediction
              parse final structured JSON response
              normalize predicted files
              if malformed:
                  if attempts remain:
                      phase: prepare_retry
                          record finalization_failed feedback
                          continue
                  return finalization_failed

              if predicted files empty:
                  if attempts remain:
                      phase: prepare_retry
                          record invalid_prediction feedback
                          continue
                  return invalid_prediction

          phase: complete
              return successful evaluator result
      }

      phase: exhausted
          return failed result with last failure and attempt history

Additional phase names introduced by this issue:

    run_pipeline
    classify_pipeline
    prepare_retry
    exhausted

Add or reuse typed failure kinds for:

    pipeline_failed
    pipeline_infrastructure_failed
    prompt_render_failed
    evaluator_failed
    tool_call_failed
    finalization_failed
    invalid_prediction
    retries_exhausted

If existing eino/run failure types already express these, reuse them.

Evaluator result behavior:
Extend the existing evaluator result type as needed.

It should include enough information for review and later tracing:

    success
    prediction/executed run
    failure
    phases
    attempts
    pipeline_results
    pipeline_classification
    metadata

Use existing result/run models if they fit.

Do not move this into pure domain unless the existing architecture clearly requires it.

Testing policy:
Default tests must be offline, deterministic, and safe.

They must not require:

- API keys
- network access
- real OpenAI/OpenRouter/Anthropic calls
- paid model usage
- real external command execution unless covered by a tiny optional hermetic local command test

Use scripted command runners for most pipeline tests.

Use Tier 1 scripted model fixtures for evaluator retry tests.

Do not use real model APIs.

Required tests:
Add tests proving:

1. pipeline step result records command, exit code, stdout/stderr, duration, and pass/fail
2. successful pipeline classification records passed steps
3. templ_generate failure becomes generation_failures
4. gofmt_check failure becomes format_errors
5. go_test failure becomes test_failures or type_errors according to the implemented heuristic
6. command launch failure becomes infrastructure_failures
7. disallowed command does not execute and becomes infrastructure failure
8. formatted pipeline feedback is deterministic and bounded
9. pipeline failure prevents evaluator model execution
10. successful pipeline allows evaluator model execution
11. malformed final output retries once and then succeeds when the second scripted response is valid
12. empty predicted files retries according to policy
13. retries exhausted returns retries_exhausted or the final typed failure with attempt history
14. prompt render failure is not retried
15. pipeline failure is not retried by default
16. retry feedback does not include gold/oracle data
17. default tests do not require API keys or real network access

Required CLI failure test:
Add at least one test specifically proving CLI failure behavior.

Example:

    scripted command runner returns:
      StepResult{
        Name: "go_test",
        Command: []string{"go", "test", "./..."},
        Passed: false,
        ExitCode: 1,
        Stderr: "FAIL ./internal/executor/eino",
      }

Assert:

- evaluator returns a typed failure
- phase is run_pipeline or classify_pipeline
- kind is pipeline_failed
- classification includes test_failures
- feedback includes TEST FAILURES
- model call count is zero

This test is important. It preserves old SearchBench behavior where external checks were first-class loop information.

Optional tiny real command test:
A tiny real subprocess test is optional.

If added, it must be hermetic and cheap.

Example:

    command runner can execute go version

or a tiny temp package test.

But this is not required. Prefer scripted command runners.

Do not depend on global tools unless the repo already requires them for normal go test ./....

Important non-goals:
Do not implement:

- writer agent
- policy generation
- policy artifact installation
- optimization loop
- MCP
- Iterative Context
- jCodeMunch
- repo materialization
- graph-distance scoring
- LangSmith tracing
- GitHub issue automation
- broad public CLI product surface
- live model calls
- broad arbitrary command execution
- retrying forever
- automatic code modification by evaluator
- candidate policy repair
- post-candidate writer feedback loop

Acceptance criteria:

- A typed CLI pipeline step result exists or an existing equivalent is reused.
- Step results record command, exit code, stdout, stderr, duration, and pass/fail.
- CLI command execution is behind a small runner abstraction.
- Tests can script command success/failure without spawning real commands.
- Command execution is allowlisted.
- Disallowed commands fail closed.
- A minimal evaluator validation pipeline exists.
- Pipeline failures are classified into structured categories.
- Pipeline feedback is deterministic and bounded.
- Pipeline failure prevents model execution by default.
- The evaluator runner supports bounded retry attempts.
- Retry attempts record attempt number and failure reason.
- Recoverable finalization failure can retry and succeed.
- Empty predicted files can retry according to policy.
- Prompt render failures are not retried.
- Pipeline failures are not retried by default.
- Retries exhausted returns a typed failure.
- Tests cover a failing CLI command.
- Tests cover retry success after an initial bad model response.
- Tests cover retry exhaustion.
- Default tests do not require API keys.
- Default tests do not call real model APIs.
- Default tests do not require external network access.
- The implementation uses ordinary Go control flow, not a state machine library.
- The implementation does not introduce writer/optimization/MCP/policy machinery.

Implementation guidance:
Keep this boring and small.

Do not overbuild.

Do not create a general pipeline framework.

Do not create a public CLI command unless the GitHub issue explicitly asks for it.

Do not create a generic provider abstraction.

Do not attempt MCP/tool runtime integration.

The goal is to make evaluator retry and CLI validation failures typed, testable, and reusable.

Before handing off, run:

    gofmt -w .
    templ generate
    go test ./...
    go mod tidy
    go test ./...

If templ is not available or not required by the current change, explain that in the final summary and run the applicable commands.

Then summarize:

- issue number and title implemented
- files changed
- new pipeline package or files added
- retry changes made to internal/executor/eino
- tests added
- how scripted command runners were used
- how Tier 1 modeltest helpers were used
- whether any real subprocess test was added
- any deliberate deviations from the issue
- confirmation that no real model/provider/network calls were added
- confirmation that pipeline failures prevent model execution
- confirmation that no MCP/backend/policy/scoring/writer work was added
